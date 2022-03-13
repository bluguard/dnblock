package inmemoryclient

import (
	"errors"
	"net"

	"github.com/bluguard/dnshield/internal/dns/client"
)

var _ client.Client = &InMemoryClient{}

type InMemoryClient struct {
	v4Store map[string]net.IP
	v6Store map[string]net.IP
}

func (c *InMemoryClient) ResolveV4(name string) (net.IP, error) {
	ip, ok := c.v4Store[name]
	if !ok {
		return nil, errors.New(name + " not found for v4")
	}
	return ip, nil
}
func (c *InMemoryClient) ResolveV6(name string) (net.IP, error) {
	ip, ok := c.v6Store[name]
	if !ok {
		return nil, errors.New(name + " not found for v6")
	}
	return ip, nil
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
		c.v6Store[name] = ip
		return true
	}
	return false
}

func (c *InMemoryClient) tryAddV4(name string, ip net.IP) bool {
	if v4 := ip.To4(); v4 != nil {
		c.v4Store[name] = ip
		return true
	}
	return false
}
