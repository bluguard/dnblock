package resolver

import "github.com/bluguard/dnshield/internal/dns"

//Middleware are parts of the r3esolve process
type Stage interface {
	Execute(dns.Message) dns.Message
	ChainWith(next Stage)
}

//Resolver
type Resolver struct {
	chain Stage
}

func (resolver *Resolver) Resolve(record dns.Message) dns.Message {
	return resolver.chain.Execute(record)
}
