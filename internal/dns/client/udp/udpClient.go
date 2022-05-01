package udp

import (
	"errors"
	"math"
	"net"

	"github.com/bluguard/dnshield/internal/dns/client"
	"github.com/bluguard/dnshield/internal/dns/dto"
)

var _ client.Client = &UdpClient{}

type UdpClient struct {
	Address string
	id      uint16
}

func (c *UdpClient) ResolveV4(name string) (dto.Record, error) {

	question := dto.Question{
		Name:  name,
		Type:  dto.A,
		Class: dto.IN,
	}

	return c.resolve(question)
}

func (c *UdpClient) ResolveV6(name string) (dto.Record, error) {
	question := dto.Question{
		Name:  name,
		Type:  dto.AAAA,
		Class: dto.IN,
	}
	return c.resolve(question)
}

func (c *UdpClient) resolve(request dto.Question) (dto.Record, error) {

	udpConn, err := net.Dial("udp", c.Address)
	if err != nil {
		return dto.Record{}, err
	}
	defer udpConn.Close()
	message := dto.Message{
		ID:            c.nextId(),
		Header:        dto.STANDARD_QUERY,
		QuestionCount: 1,
		ResponseCount: 0,
		Question:      []dto.Question{request},
		Response:      []dto.Record{},
	}

	payload := dto.SerializeMessage(message)

	_, err = udpConn.Write(payload)
	if err != nil {
		return dto.Record{}, err
	}

	response, err := c.waitResponse(udpConn)
	if err != nil {
		return dto.Record{}, err
	}

	if len(response.Response) < 1 {
		return dto.Record{}, errors.New("no response found")
	}

	return response.Response[0], nil
}

func (c *UdpClient) nextId() uint16 {
	c.id++
	c.id = c.id % math.MaxUint16
	return c.id
}

func (c *UdpClient) waitResponse(udpConn net.Conn) (*dto.Message, error) {
	buffer := make([]byte, dto.BufferMaxLength)
	n, err := udpConn.Read(buffer)
	if err != nil {
		return nil, err
	}
	data := buffer[0:n]
	message, err := dto.ParseMessage(data)
	if err != nil {
		return nil, err
	}
	return message, nil
}
