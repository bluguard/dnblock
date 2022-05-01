package cache

import (
	"github.com/bluguard/dnshield/internal/dns/client"
	"github.com/bluguard/dnshield/internal/dns/dto"
)

type Feedable interface {
	Feed(dto.Record)
}

type Cache interface {
	client.Client
	Feedable
	Clear()
}
