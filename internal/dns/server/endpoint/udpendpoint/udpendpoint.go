package udpendpoint

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bluguard/dnshield/internal/dns/dto"
	"github.com/bluguard/dnshield/internal/dns/resolver"
	"github.com/bluguard/dnshield/internal/dns/server/endpoint"
)

const (
	udpTimeout = 200 * time.Millisecond
	workers    = 10
	maxPending = 1000
)

var _ endpoint.Endpoint = &UDPEndpoint{}

type question struct {
	message     []byte
	destination net.UDPAddr
	arrival     time.Time
}

// NewUDPEndpoint create a new udp enpoint with the given chain
func NewUDPEndpoint(address string, chain *resolver.ResolverChain) *UDPEndpoint {
	return &UDPEndpoint{
		laddr:      address,
		chain:      chain,
		lock:       sync.RWMutex{},
		started:    atomic.Bool{},
		inbox:      make(chan question, maxPending),
		bufferPool: sync.Pool{New: func() any { return make([]byte, dto.BufferMaxLength) }},
	}
}

// UDPEndpoint endpoint based on udp protocol
type UDPEndpoint struct {
	laddr      string
	chain      *resolver.ResolverChain
	lock       sync.RWMutex
	started    atomic.Bool
	inbox      chan question
	bufferPool sync.Pool
}

// SetChain implements server.Endpoint
func (e *UDPEndpoint) SetChain(chain *resolver.ResolverChain) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.chain = chain
}

// Start implements server.Endpoint
func (e *UDPEndpoint) Start(ctx context.Context, wg *sync.WaitGroup) {
	if !e.started.CompareAndSwap(false, true) {
		panic("endpoint is already started")
	}
	log.Println("starting udp endpoint on", e.laddr)
	go e.run(ctx, wg)
}

func (e *UDPEndpoint) run(ctx context.Context, ewg *sync.WaitGroup) {
	defer ewg.Done()

	iwg := &sync.WaitGroup{}

	conns := e.populateConn(ctx, workers)
	defer closeAll(conns)

	// start the receiving loop
	// tart the workers
	iwg.Add(workers * 2)
	for i := 0; i < workers; i++ {
		go e.receivingLoop(ctx, conns[i], iwg)
		go e.handler(ctx, conns[i], iwg)
	}

	iwg.Wait()
	log.Println("udp endpoint on", e.laddr, "stopped")
}

func (e *UDPEndpoint) receivingLoop(ctx context.Context, udpConn *net.UDPConn, wg *sync.WaitGroup) {
	// Main loop
	defer wg.Done()
	defer udpConn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			e.receive(udpConn)
		}
	}
}

func (e *UDPEndpoint) receive(udpConn *net.UDPConn) {
	buff := e.getBuffer()
	_ = udpConn.SetReadDeadline(time.Now().Add(udpTimeout))
	n, addr, err := udpConn.ReadFromUDP(buff)
	if err != nil {
		if err, ok := err.(net.Error); ok && (err.Timeout() || errors.Is(err, net.ErrClosed)) {
			return
		}
		panic(err)
	}
	e.inbox <- question{message: buff[0:n], destination: *addr, arrival: time.Now()}
}

func (e *UDPEndpoint) handler(ctx context.Context, udpConn *net.UDPConn, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-e.inbox:
			e.handleRequest(msg.message, &msg.destination, udpConn)
			e.recycle(msg.message)
		}
	}
}

func (e *UDPEndpoint) handleRequest(buffer []byte, dest *net.UDPAddr, udpConn *net.UDPConn) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	message, err := dto.ParseMessage(buffer)
	if err != nil {
		log.Println(err)
		return
	}
	res := e.chain.Resolve(*message)
	send(res, dest, udpConn)
}

func send(message dto.Message, dest *net.UDPAddr, udpConn *net.UDPConn) bool {
	payload := dto.SerializeMessage(message)
	_, err := udpConn.WriteToUDP(payload, dest)
	if err != nil {
		if terr, ok := err.(net.Error); !(ok && terr.Timeout()) {
			log.Println(err)
			return true
		}
	}
	return false
}

func (e *UDPEndpoint) getBuffer() []byte {
	return e.bufferPool.Get().([]byte)
}

func (e *UDPEndpoint) recycle(buff []byte) {
	e.bufferPool.Put(buff[0:dto.BufferMaxLength])
}

func (e *UDPEndpoint) populateConn(ctx context.Context, n int) []*net.UDPConn {
	res := make([]*net.UDPConn, n)

	for i := 0; i < n; i++ {

		conf := net.ListenConfig{
			Control: reusePort,
		}

		conn, err := conf.ListenPacket(ctx, "udp", e.laddr)
		if err != nil {
			panic(err)
		}
		udpConn, ok := conn.(*net.UDPConn)
		if !ok {
			panic("connection is not an udp connection")
		}
		err = udpConn.SetReadBuffer(dto.BufferMaxLength * workers * 2)
		if err != nil {
			panic(err)
		}
		err = udpConn.SetWriteBuffer(dto.BufferMaxLength)
		if err != nil {
			panic(err)
		}

		res[i] = udpConn
	}
	return res
}

func closeAll(r []*net.UDPConn) {
	for _, c := range r {
		_ = c.Close()
	}
}
