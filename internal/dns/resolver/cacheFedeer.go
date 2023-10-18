package resolver

import (
	"github.com/bluguard/dnshield/internal/dns/cache"
	"github.com/bluguard/dnshield/internal/dns/dto"
)

var _ Resolver = &Cachefeeder{}

// Cachefeeder is in charge to feed a cache based on the answer of a resolver
type Cachefeeder struct {
	delegate Resolver
	cache    cache.Feedable
}

func NewCacheFeeder(delegate Resolver, cache cache.Feedable) *Cachefeeder {
	return &Cachefeeder{
		delegate: delegate,
		cache:    cache,
	}
}

// Name implements Resolver
func (r *Cachefeeder) Name() string {
	return r.delegate.Name()
}

// Resolve implements Resolver
func (r *Cachefeeder) Resolve(question dto.Question) (dto.Record, bool) {
	result, ok := r.delegate.Resolve(question)
	if ok {
		r.cache.Feed(result)
	}
	return result, ok
}
