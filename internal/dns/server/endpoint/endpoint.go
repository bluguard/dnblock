package endpoint

import (
	"context"
	"sync"

	"github.com/bluguard/dnshield/internal/dns/resolver"
)

// Endpoint represents a server endpoint to serve dns
type Endpoint interface {
	Start(context.Context, *sync.WaitGroup)
	SetChain(chain *resolver.ResolverChain)
}
