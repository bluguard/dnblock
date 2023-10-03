package udpendpoint

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/bluguard/dnshield/internal/dns/dto"
	"github.com/bluguard/dnshield/internal/dns/resolver"
	"github.com/bluguard/dnshield/internal/dns/server/endpoint"
)

const (
	udpTimeout = 200 * time.Millisecond
	workers    = 10
)

var _ endpoint.Endpoint = &UDPEndpoint{}

type response struct {
	message     dto.Message
	destination net.UDPAddr
}

// NewUDPEndpoint create a new udp enpoint with the given chain
func NewUDPEndpoint(address string, chain *resolver.ResolverChain) *UDPEndpoint {
	return &UDPEndpoint{
		laddr:    address,
		chain:    chain,
		lock:     sync.RWMutex{},
		started:  false,
		sendChan: make(chan response),
	}
}

// UDPEndpoint endpoint based on udp protocol
type UDPEndpoint struct {
	laddr    string
	chain    *resolver.ResolverChain
	lock     sync.RWMutex
	started  bool
	sendChan chan response
}

// SetChain implements server.Endpoint
func (e *UDPEndpoint) SetChain(chain *resolver.ResolverChain) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.chain = chain
}

// Start implements server.Endpoint
func (e *UDPEndpoint) Start(ctx context.Context, wg *sync.WaitGroup) {
	if e.started {
		panic("endpoint is already started")
	}
	log.Println("starting udp endpoint on ", e.laddr)
	e.started = true
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
	err = udpConn.SetReadBuffer(dto.BufferMaxLength)
	if err != nil {
		log.Println(err)
		return
	}
	defer udpConn.Close()
	iwg := sync.WaitGroup{}

	for i := 0; i < workers; i++ {
		go e.receivingLoop(ctx, udpConn, &iwg)
	}
	iwg.Add(workers)
	iwg.Add(1)
	go e.sendingLoop(ctx, udpConn, &iwg)

	iwg.Wait()
	log.Println("udp endpoint on ", e.laddr, "stopped")
}

func (e *UDPEndpoint) receivingLoop(ctx context.Context, udpConn *net.UDPConn, wg *sync.WaitGroup) {
	//Main loop
	defer wg.Done()
	defer udpConn.Close()
	for {
		start := time.Now()
		select {
		case <-ctx.Done():
			log.Println("udp endpoint on ", e.laddr, " is terminating")
			return
		default:
			// if timeout loop
			shouldReturn := e.receive(udpConn)
			if shouldReturn {
				return
			}
			log.Println("receiving loop iteration took", time.Since(start))
		}
	}
}

func (e *UDPEndpoint) receive(udpConn *net.UDPConn) bool {
	buffer := make([]byte, dto.BufferMaxLength)
	n, addr, err := udpConn.ReadFromUDP(buffer)
	if err != nil {
		if terr, ok := err.(net.Error); !(ok && terr.Timeout()) {
			log.Println(err)
			return true
		}
	}
	if n == 0 {
		return false
	}
	data := buffer[0:n]
	go e.handleRequest(data, addr)
	return false
}

func (e *UDPEndpoint) sendingLoop(ctx context.Context, udpConn *net.UDPConn, iwg *sync.WaitGroup) {
	defer iwg.Done()

	defer udpConn.Close()
	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-e.sendChan:
			// if timeout loop
			shouldReturn := send(resp, udpConn)
			if shouldReturn {
				return
			}
		}
	}
}

func send(resp response, udpConn *net.UDPConn) bool {
	payload := dto.SerializeMessage(resp.message)
	err := udpConn.SetWriteDeadline(time.Now().Add(udpTimeout))
	if err != nil {
		log.Println(err)
		return true
	}
	_, err = udpConn.WriteToUDP(payload, &resp.destination)
	if err != nil {
		if terr, ok := err.(net.Error); !(ok && terr.Timeout()) {
			log.Println(err)
			return true
		}
	}
	return false
}

func (e *UDPEndpoint) handleRequest(buffer []byte, addr *net.UDPAddr) {
	//log.Println("Handling request for ", addr.IP)
	start := time.Now()
	e.lock.RLock()
	defer e.lock.RUnlock()
	message, err := dto.ParseMessage(buffer)
	if err != nil {
		log.Println(err)
		return
	}
	res := e.chain.Resolve(*message)
	//log.Println("Handling request for ", addr.IP, message, " -> ", res)
	e.sendChan <- response{
		message:     res,
		destination: *addr,
	}
	delay := time.Since(start)
	log.Println("resolving", message.QuestionCount, "questions took", delay.String())
}
