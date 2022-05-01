package client

import (
	"github.com/bluguard/dnshield/internal/dns/dto"
)

type Client interface {
	ResolveV4(name string) (dto.Record, error)
	ResolveV6(name string) (dto.Record, error)
}

type ReversableClient interface {
	Client
	ReverseResolve(ip string)
}
