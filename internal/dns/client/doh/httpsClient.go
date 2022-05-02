package doh

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/bluguard/dnshield/internal/dns/client"
	"github.com/bluguard/dnshield/internal/dns/dto"
)

var _ client.Client = &DOHClient{}

type DOHClient struct {
	endpoint string
}

func NewDOHClient(endpoint string) *DOHClient {
	return &DOHClient{
		endpoint: endpoint,
	}
}

// ResolveV4 implements client.Client
func (c *DOHClient) ResolveV4(name string) (dto.Record, error) {
	return c.resolve(name, dto.A)
}

// ResolveV6 implements client.Client
func (c *DOHClient) ResolveV6(name string) (dto.Record, error) {
	return c.resolve(name, dto.AAAA)
}

func (c *DOHClient) resolve(name string, t dto.Type) (dto.Record, error) {
	req, err := http.NewRequest("GET", c.endpoint+"?name="+name+"&type="+strconv.Itoa(int(t)), nil)
	if err != nil {
		return dto.Record{}, err
	}
	req.Header.Add("accept", "application/dns-json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return dto.Record{}, err
	}
	var message Message
	err = json.NewDecoder(resp.Body).Decode(&message)
	if err != nil {
		return dto.Record{}, err
	}
	if message.Status > 0 {
		return dto.Record{}, errors.New("status is " + strconv.Itoa(message.Status))
	}
	if len(message.Answer) < 1 {
		return dto.Record{}, errors.New("no answer in response")
	}
	if message.Answer[0].Type == 5 {
		record, err := c.resolve(message.Answer[0].Data, t)
		record.Name = name //Keep the Answer consistent with the initial Question
		return record, err
	}
	if message.Answer[0].Type != uint16(dto.A) && message.Answer[0].Type != uint16(dto.AAAA) {
		log.Println("receive message of type", message.Answer[0].Type)
		return dto.Record{}, errors.New("answer with unknown type in response")
	}

	return message.Answer[0].ToRecord(), nil
}
