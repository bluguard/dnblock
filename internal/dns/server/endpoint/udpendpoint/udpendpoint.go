package udpendpoint

import (
	"context"
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
	log.Println("starting udp endpoint on ", e.laddr)
	go e.run(ctx, wg)
}

func (e *UDPEndpoint) run(ctx context.Context, ewg *sync.WaitGroup) {
	defer ewg.Done()
	address, err := net.ResolveUDPAddr("udp", e.laddr)
	if err != nil {
		log.Println(err)
		return
	}
	udpConn, err := net.ListenUDP("udp", address)
	if err != nil {
		log.Println(err)
		return
	}
	err = udpConn.SetReadBuffer(dto.BufferMaxLength * workers * 2)
	if err != nil {
		log.Println(err)
		return
	}
	err = udpConn.SetWriteBuffer(dto.BufferMaxLength)
	if err != nil {
		log.Println(err)
		return
	}
	defer udpConn.Close()
	iwg := &sync.WaitGroup{}

	// start the receiving loop
	// tart the workers
	iwg.Add(workers * 2)
	for i := 0; i < workers; i++ {
		go e.receivingLoop(ctx, udpConn, iwg)
		go e.handler(ctx, udpConn, iwg)
	}

	iwg.Wait()
	log.Println("udp endpoint on ", e.laddr, "stopped")
}

func (e *UDPEndpoint) receivingLoop(ctx context.Context, udpConn *net.UDPConn, wg *sync.WaitGroup) {
	// Main loop
	defer wg.Done()
	defer udpConn.Close()

	for {
		select {
		case <-ctx.Done():
			log.Println("udp endpoint on ", e.laddr, " is terminating")
			return
		default:
			e.receive(udpConn)
		}
	}
}

func (e *UDPEndpoint) receive(udpConn *net.UDPConn) {
	start := time.Now()
	buff := e.getBuffer()
	_ = udpConn.SetReadDeadline(time.Now().Add(udpTimeout))
	n, addr, err := udpConn.ReadFromUDP(buff)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return
		}
		panic(err)
	}
	e.inbox <- question{message: buff[0:n], destination: *addr, arrival: time.Now()}

	log.Println("receiving loop iteration took", time.Since(start))
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
			log.Println("message handling took", time.Since(msg.arrival).String())
		}
	}
}

func (e *UDPEndpoint) handleRequest(buffer []byte, dest *net.UDPAddr, udpConn *net.UDPConn) {
	// log.Println("Handling request for ", addr.IP)
	start := time.Now()
	e.lock.RLock()
	defer e.lock.RUnlock()
	message, err := dto.ParseMessage(buffer)
	if err != nil {
		log.Println(err)
		return
	}
	res := e.chain.Resolve(*message)
	send(res, dest, udpConn)
	delay := time.Since(start)
	log.Println("resolving", message.QuestionCount, "questions took", delay.String())
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
