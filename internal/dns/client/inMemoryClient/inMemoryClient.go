package inmemoryclient

import (
	"errors"
	"net"
	"sync"

	"github.com/bluguard/dnshield/internal/dns/client"
	"github.com/bluguard/dnshield/internal/dns/dto"
)

var _ client.Client = &InMemoryClient{}

//Concurrent safe client, storing data in memory
type InMemoryClient struct {
	v4Store sync.Map
	v6Store sync.Map
}

func (c *InMemoryClient) ResolveV4(name string) (dto.Record, error) {
	ip, ok := c.v4Store.Load(name)
	if !ok {
		return dto.Record{}, errors.New(name + " not found for v4")
	}
	return dto.Record{
		Name:  name,
		Type:  dto.A,
		Class: dto.IN,
		TTL:   200,
		Data:  ip.(net.IP),
	}, nil
}
func (c *InMemoryClient) ResolveV6(name string) (dto.Record, error) {
	ip, ok := c.v6Store.Load(name)
	if !ok {
		return dto.Record{}, errors.New(name + " not found for v6")
	}
	return dto.Record{
		Name:  name,
		Type:  dto.AAAA,
		Class: dto.IN,
		TTL:   200,
		Data:  ip.(net.IP),
	}, nil
}

func (c *InMemoryClient) Add(name, address string) error {
	ip := net.ParseIP(address)
	if !(c.tryAddV4(name, ip) || c.tryAddV6(name, ip)) {
		return errors.New("unknown address format for " + ip.String())
	}
	return nil
}

func (c *InMemoryClient) tryAddV6(name string, ip net.IP) bool {
	if v6 := ip.To16(); v6 != nil {
		c.v6Store.Store(name, v6)
		return true
	}
	return false
}

func (c *InMemoryClient) tryAddV4(name string, ip net.IP) bool {
	if v4 := ip.To4(); v4 != nil {
		c.v4Store.Store(name, v4)
		return true
	}
	return false
}
