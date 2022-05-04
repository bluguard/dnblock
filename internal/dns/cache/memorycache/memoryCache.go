package memorycache

import (
	"context"
	"errors"
	"hash/fnv"
	"log"
	"net"
	"sync"
	"time"

	"github.com/bluguard/dnshield/internal/dns/cache"
	"github.com/bluguard/dnshield/internal/dns/dto"
)

// estimate cost of one entry is 50 bytes
const cost int64 = 50

const defaultTtl = 60

const (
	v4Suffix = "_v4"
	v6Suffix = "_v6"
)

var _ cache.Cache = &MemoryCache{}

type MemoryCache struct {
	memory          map[uint32]net.IP
	lock            *sync.RWMutex
	deadlines       *deadlineFolder
	remainingMemory int64
	baseTtl         uint32
}

func NewMemoryCache(size int64, baseTtl uint32, ctx context.Context, wg *sync.WaitGroup, gcDelay time.Duration) *MemoryCache {
	res := MemoryCache{
		memory:          make(map[uint32]net.IP),
		lock:            &sync.RWMutex{},
		deadlines:       &deadlineFolder{memory: make([]deadline, 0, 50)},
		remainingMemory: size,
		baseTtl:         baseTtl,
	}

	if baseTtl > 0 {
		go gcScheduler(&res, ctx, wg, gcDelay)
	} else {
		wg.Done()
	}

	return &res
}

// ResolveV4 implements cache.Cache
func (c *MemoryCache) ResolveV4(name string) (dto.Record, error) {
	ip, err := c.resolve(name + v4Suffix)
	if err != nil {
		return dto.Record{}, err
	}
	return dto.Record{
		Name:  name,
		Type:  dto.A,
		Class: dto.IN,
		TTL:   defaultTtl,
		Data:  ip.To4(),
	}, nil
}

// ResolveV6 implements cache.Cache
func (c *MemoryCache) ResolveV6(name string) (dto.Record, error) {
	ip, err := c.resolve(name + v6Suffix)
	if err != nil {
		return dto.Record{}, err
	}
	return dto.Record{
		Name:  name,
		Type:  dto.A,
		Class: dto.IN,
		TTL:   defaultTtl,
		Data:  ip.To16(),
	}, nil
}

func (c *MemoryCache) resolve(name string) (net.IP, error) {
	res := c.get(name)
	if res == nil {
		return nil, errors.New("no entry found for " + name)
	}
	return res, nil
}

// Feed implements cache.Cache
func (c *MemoryCache) Feed(record dto.Record) {
	ttl := record.TTL
	if record.TTL < c.baseTtl {
		ttl = c.baseTtl
		log.Println("use default ttl for", record.Name, record.Type, record.TTL)
	}
	c.put(computeName(record.Name, record.Type), record.Data, time.Duration(ttl))
}

// Clear implements cache.Cache
func (c *MemoryCache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for k := range c.memory {
		delete(c.memory, k)
	}
	c.deadlines.shiftLeftOf(len(c.deadlines.memory))
}

func (c *MemoryCache) put(key string, address net.IP, ttl time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.remainingMemory < cost {
		c.freeNextDeadline()
	} else {
		c.remainingMemory -= cost
	}

	hkey := hash(key)
	c.memory[hkey] = address
	c.deadlines.insert(deadline{expiry: time.Now().Add(ttl), key: hkey})
}

func (c *MemoryCache) get(key string) net.IP {
	c.lock.RLock()
	defer c.lock.RUnlock()
	res, ok := c.memory[hash(key)]
	if !ok {
		return nil
	}
	return res
}

func (c *MemoryCache) gc() {
	c.lock.Lock()
	defer c.lock.Unlock()
	count := 0
	now := time.Now()
	for _, d := range c.deadlines.memory {
		if d.expiry.Before(now) {
			count++
			delete(c.memory, d.key)
		} else {
			break //the list of deadlines is sorted, no need to range over all elements
		}
	}
	c.deadlines.shiftLeftOf(count)
	c.remainingMemory += cost * int64(count)
}

func (c *MemoryCache) freeNextDeadline() {
	delete(c.memory, c.deadlines.memory[0].key)
	c.deadlines.shiftLeftOf(1)
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
func computeName(s string, t dto.Type) string {
	switch t {
	case dto.A:
		return s + v4Suffix
	case dto.AAAA:
		return s + v6Suffix
	default:
		return s + v4Suffix
	}
}

func gcScheduler(memoryCache *MemoryCache, ctx context.Context, wg *sync.WaitGroup, gcDelay time.Duration) {
	defer wg.Done()
	ticker := time.NewTicker(gcDelay)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			memoryCache.gc()
		}
	}
}