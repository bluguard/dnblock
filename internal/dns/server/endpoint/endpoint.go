package endpoint

import (
	"context"
	"sync"

	"github.com/bluguard/dnshield/internal/dns/resolver"
)

type Endpoint interface {
	Start(context.Context, *sync.WaitGroup)
	SetChain(chain *resolver.ResolverChain)
}
