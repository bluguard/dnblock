package ristretoimpl

import (
	"errors"
	"hash/fnv"
	"log"
	"net"
	"time"

	"github.com/bluguard/dnshield/internal/dns/cache"
	"github.com/bluguard/dnshield/internal/dns/dto"
	"github.com/dgraph-io/ristretto"
)

const (
	v4Suffix        = "_v4"
	v6Suffix        = "_v6"
	megabyteInBytes = 1000000
	keySize         = 4
)

var _ cache.Cache = &RistretoCache{}

type RistretoCache struct {
	memory  *ristretto.Cache
	basettl uint32
}

func NewRistrettoCache(cacheSizeMb int64, baseTtl uint32) *RistretoCache {
	cache, err := ristretto.NewCache(&ristretto.Config{
		MaxCost:     cacheSizeMb * megabyteInBytes,
		NumCounters: 1e7,
		Metrics:     false,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}

	return &RistretoCache{
		memory:  cache,
		basettl: baseTtl,
	}
}

// ResolveV4 implements cache.Cache
func (c *RistretoCache) ResolveV4(name string) (dto.Record, error) {
	ip, err := c.resolve(name + v4Suffix)
	if err != nil {
		return dto.Record{}, err
	}
	return dto.Record{
		Name:  name,
		Type:  dto.A,
		Class: dto.IN,
		TTL:   c.basettl,
		Data:  ip.To4(),
	}, nil
}

// ResolveV6 implements cache.Cache
func (c *RistretoCache) ResolveV6(name string) (dto.Record, error) {
	ip, err := c.resolve(name + v6Suffix)
	if err != nil {
		return dto.Record{}, err
	}
	return dto.Record{
		Name:  name,
		Type:  dto.AAAA,
		Class: dto.IN,
		TTL:   c.basettl,
		Data:  ip.To16(),
	}, nil
}

func (c *RistretoCache) resolve(name string) (net.IP, error) {
	key := hash(name)
	record, ok := c.memory.Get(key)
	if !ok {
		return nil, errors.New(name + " not found in cache")
	}
	real, ok := record.(net.IP)
	if !ok {
		c.memory.Del(key)
		return nil, errors.New(name + " entry malformed")
	}
	return real, nil
}

// Feed implements cache.Cache
func (c *RistretoCache) Feed(record dto.Record) {
	if record.TTL == 0 { //one shot response
		return
	}
	key := computeKey(record.Name, record.Type)
	ttl := record.TTL
	if record.TTL < c.basettl {
		ttl = c.basettl
		log.Println("use base ttl for", record.Name, record.TTL)
	}
	if c.basettl == 0 {
		ttl = 0
	}
	cost := int64(keySize + len(record.Data))
	c.memory.SetWithTTL(key, record.Data, cost, time.Second*time.Duration(ttl))
}

func computeKey(s string, t dto.Type) uint32 {
	switch t {
	case dto.A:
		return hash(s + v4Suffix)
	case dto.AAAA:
		return hash(s + v6Suffix)
	default:
		return hash(s + v4Suffix)
	}
}

// Clear implements cache.Cache
func (c *RistretoCache) Clear() {
	c.memory.Clear()
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
