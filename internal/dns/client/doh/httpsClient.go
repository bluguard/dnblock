package doh

import (
	"bytes"
	"errors"
	"log"
	"strconv"

	json "github.com/goccy/go-json"
	"github.com/valyala/fasthttp"

	"github.com/bluguard/dnshield/internal/dns/client"
	"github.com/bluguard/dnshield/internal/dns/dto"
)

var _ client.Client = &DOHClient{}

// DOHClient Dns Pver Http clien, resolve request by requesting it to an http server
type DOHClient struct {
	endpoint   string
	httpClient *fasthttp.Client
}

// NewDOHClient instantiate a new DOHClient
func NewDOHClient(endpoint string) *DOHClient {
	// t := http.DefaultTransport.(*http.Transport).Clone()
	// t.MaxIdleConns = 100
	// t.MaxConnsPerHost = 100
	// t.MaxIdleConnsPerHost = 100

	httpClient := &fasthttp.Client{
		MaxConnsPerHost: 100,
	}
	// &http.Client{
	// 	Timeout:   10 * time.Second,
	// 	Transport: t,
	// }

	return &DOHClient{
		endpoint:   endpoint,
		httpClient: httpClient,
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
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(c.endpoint + "?name=" + name + "&type=" + strconv.Itoa(int(t)))
	req.Header.Add("accept", "application/dns-json")
	req.Header.SetMethod("GET")

	c.httpClient.Do(req, resp)

	var message Message
	err := json.NewDecoder(bytes.NewReader(resp.Body())).Decode(&message)

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
		record.Name = name // Keep the Answer consistent with the initial Question
		return record, err
	}
	if message.Answer[0].Type != uint16(dto.A) && message.Answer[0].Type != uint16(dto.AAAA) {
		log.Println("receive message of type", message.Answer[0].Type)
		return dto.Record{}, errors.New("answer with unknown type in response")
	}

	return message.Answer[0].ToRecord(), nil
}
