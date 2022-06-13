package blocker

import (
	"errors"
	"net"

	"github.com/bluguard/dnshield/internal/dns/client"
	"github.com/bluguard/dnshield/internal/dns/dto"
)

var _ client.Client = &Blocker{}

var (
	v4Block = net.ParseIP("0.0.0.0").To4()
	v6Block = net.ParseIP("::1").To16()
)

const defaultTTl uint32 = 600

type Blocker map[string]struct{}

// ResolveV4 implements client.Client
func (b *Blocker) ResolveV4(name string) (dto.Record, error) {
	if b.contains(name) {
		return dto.Record{
			Name:  name,
			Type:  dto.A,
			Class: dto.IN,
			TTL:   defaultTTl,
			Data:  v4Block,
		}, nil
	}
	return dto.Record{}, errors.New("not blocking")
}

// ResolveV6 implements client.Client
func (b *Blocker) ResolveV6(name string) (dto.Record, error) {
	if b.contains(name) {
		return dto.Record{
			Name:  name,
			Type:  dto.AAAA,
			Class: dto.IN,
			TTL:   defaultTTl,
			Data:  v6Block,
		}, nil
	}
	return dto.Record{}, errors.New("not blocking")
}

func (b Blocker) contains(name string) bool {
	_, ok := b[name]
	return ok
}

func (b Blocker) add(name string) {
	b[name] = struct{}{}
}

func (b *Blocker) Init(i Initializer) {
	i(b.add)
}

type Initializer func(func(string))
