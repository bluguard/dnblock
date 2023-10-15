package udp

import (
	"errors"
	"log"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/bluguard/dnshield/internal/dns/client"
	"github.com/bluguard/dnshield/internal/dns/dto"
)

var _ client.Client = &UDPClient{}

var _ error = &NoResponse{}

type NoResponse struct{}

// Error implements error.
func (*NoResponse) Error() string {
	return "no response found"
}

type UDPClient struct {
	id            uint16
	connexionPool *sync.Pool
	bufferPool    *sync.Pool
	idMutex       sync.Locker
}

// NewUDPClient instantiate a UDPClient for the given address
func NewUDPClient(address string) *UDPClient {
	return &UDPClient{
		id:      0,
		idMutex: &sync.Mutex{},
		connexionPool: &sync.Pool{New: func() any {
			udpConn, err := net.Dial("udp", address)
			if err != nil {
				panic(err)
			}
			return udpConn
		}},
		bufferPool: &sync.Pool{New: func() any {
			return make([]byte, dto.BufferMaxLength)
		}},
	}
}

func (c *UDPClient) ResolveV4(name string) (dto.Record, error) {

	question := dto.Question{
		Name:  name,
		Type:  dto.A,
		Class: dto.IN,
	}

	return c.resolve(question)
}

func (c *UDPClient) ResolveV6(name string) (dto.Record, error) {
	question := dto.Question{
		Name:  name,
		Type:  dto.AAAA,
		Class: dto.IN,
	}
	return c.resolve(question)
}

func (c *UDPClient) resolve(request dto.Question) (dto.Record, error) {

	request.Name = strings.TrimRight(request.Name, ".")

	udpConn := c.getConn()
	defer c.recycleConn(udpConn)

	message := dto.Message{
		ID:            c.nextID(),
		Header:        dto.STANDARD_QUERY,
		QuestionCount: 1,
		ResponseCount: 0,
		Question:      []dto.Question{request},
		Response:      []dto.Record{},
	}

	payload := dto.SerializeMessage(message)

	_, err := udpConn.Write(payload)
	if err != nil {
		return dto.Record{}, err
	}

	response, err := c.waitResponse(udpConn, message.ID)
	if err != nil {
		return dto.Record{}, err
	}

	if len(response.Response) < 1 {
		return dto.Record{}, &NoResponse{}
	}

	return response.Response[0], nil
}

func (c *UDPClient) nextID() uint16 {
	c.idMutex.Lock()
	defer c.idMutex.Unlock()

	c.id++
	c.id = c.id % math.MaxUint16
	return c.id
}

func (c *UDPClient) waitResponse(udpConn net.Conn, id uint16) (*dto.Message, error) {
	buffer := c.getBuffer()
	defer c.recycleBuffer(buffer)
	_ = udpConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	n, err := udpConn.Read(buffer)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		log.Println("client read 0 bytes")
	}
	data := buffer[0:n]
	message, err := dto.ParseMessage(data)
	if err != nil {
		return nil, err
	}
	if id != message.ID {
		return nil, errors.New("id mismatch")
	}
	return message, nil
}

func (c *UDPClient) getConn() net.Conn {
	return c.connexionPool.Get().(net.Conn)
}

func (c *UDPClient) recycleConn(conn net.Conn) {
	c.connexionPool.Put(conn)
}

func (c *UDPClient) getBuffer() []byte {
	return c.bufferPool.Get().([]byte)
}

func (c *UDPClient) recycleBuffer(buff []byte) {
	c.bufferPool.Put(buff)
}
