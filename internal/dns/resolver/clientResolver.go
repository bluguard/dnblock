package resolver

import (
	"github.com/bluguard/dnshield/internal/dns/client"
	"github.com/bluguard/dnshield/internal/dns/dto"
)

var _ Resolver = &ClientResolver{}

func NewClientresolver(c client.Client, name string) *ClientResolver {
	return &ClientResolver{
		name:   name,
		client: c,
	}
}

// ClientResolver is a resolver who delegates to a dns client interface
type ClientResolver struct {
	name   string
	client client.Client
}

// Name implements Resolver
func (resolver *ClientResolver) Name() string {
	return resolver.name
}

// Resolve implements Resolver
// Use the client to get the records
func (resolver *ClientResolver) Resolve(question dto.Question) (dto.Record, bool) {
	var callClient func(string) (dto.Record, error)
	if question.Type == dto.A {
		callClient = resolver.client.ResolveV4
	} else if question.Type == dto.AAAA {
		callClient = resolver.client.ResolveV6
	}
	if callClient == nil {
		return dto.Record{}, false
	}
	record, err := callClient(question.Name)
	if err != nil {
		return dto.Record{}, false
	}
	return record, true
}
