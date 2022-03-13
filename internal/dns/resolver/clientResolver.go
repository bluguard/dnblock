package resolver

import (
	"log"
	"net"

	"github.com/bluguard/dnshield/internal/dns/client"
	"github.com/bluguard/dnshield/internal/dns/dto"
)

var _ Resolver = &ClientResolver{}

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
	var callClient func(string) (net.IP, error)
	if question.Type == dto.A {
		callClient = resolver.client.ResolveV4
	} else if question.Type == dto.AAAA {
		callClient = resolver.client.ResolveV6
	}
	if callClient == nil {
		return dto.Record{}, false
	}
	ip, err := callClient(question.Name)
	if err != nil {
		log.Println(err)
		return dto.Record{}, false
	}
	return dto.Record{
		Name:  question.Name,
		Type:  question.Type,
		Class: question.Class,
		TTL:   200, // get the value from the actual record
		Data:  ip,
	}, true
}
