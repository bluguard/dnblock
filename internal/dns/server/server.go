package server

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bluguard/dnshield/internal/dns/cache/memorycache"
	"github.com/bluguard/dnshield/internal/dns/client"
	"github.com/bluguard/dnshield/internal/dns/client/doh"
	inmemoryclient "github.com/bluguard/dnshield/internal/dns/client/inMemoryClient"
	"github.com/bluguard/dnshield/internal/dns/client/udp"
	"github.com/bluguard/dnshield/internal/dns/resolver"
	"github.com/bluguard/dnshield/internal/dns/server/configuration"
	"github.com/bluguard/dnshield/internal/dns/server/endpoint"
	"github.com/bluguard/dnshield/internal/dns/server/endpoint/udpendpoint"
)

type Server struct {
	chain     resolver.ResolverChain
	endpoints []endpoint.Endpoint
	started   bool
	//http controller
	cancelFunc context.CancelFunc
}

func (s *Server) Start(conf configuration.ServerConf) *sync.WaitGroup {
	if s.started {
		log.Println("server already started")
	}
	log.Println("starting server ...")
	s.started = true

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-ch
		if s.cancelFunc != nil {
			s.cancelFunc()
		}
	}()

	wg := s.Reconfigure(conf)
	log.Println("server started")
	return wg

}

func (s *Server) Stop() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
}

func (s *Server) Reconfigure(conf configuration.ServerConf) *sync.WaitGroup {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	s.cancelFunc = cancelFunc

	wg := sync.WaitGroup{}

	cache := memorycache.NewMemoryCache(conf.Cache.Size, conf.Cache.Basettl, ctx, &wg, 1*time.Minute)

	s.chain = *resolver.NewResolverChain([]resolver.Resolver{
		resolver.NewClientresolver(buildBlocker(conf), "Block"),
		resolver.NewClientresolver(buildCustom(conf), "Custom"),
		resolver.NewClientresolver(cache, "Cache"),
		resolver.NewCacheFeeder(resolver.NewClientresolver(buildExternal(conf), "External"), cache),
	})

	s.endpoints = createEndpoints(conf, &s.chain)

	for _, endpoint := range s.endpoints {
		wg.Add(1)
		endpoint.Start(ctx, &wg)
	}
	return &wg
}

func createEndpoints(conf configuration.ServerConf, chain *resolver.ResolverChain) []endpoint.Endpoint {
	return []endpoint.Endpoint{
		udpendpoint.NewUdpEndpoint(conf.Endpoint.Address, chain),
	}
}

func buildExternal(conf configuration.ServerConf) client.Client {
	if !conf.AllowExternal {
		panic("unexpected")
	}
	switch conf.External.Type {
	case "DOH":
		return doh.NewDOHClient(conf.External.Endpoint)
	default:
		return &udp.UdpClient{Address: conf.External.Endpoint}
	}
}

func buildCustom(conf configuration.ServerConf) client.Client {
	res := inmemoryclient.InMemoryClient{}
	for _, v := range conf.Custom {
		err := res.Add(v.Name, v.Address)
		if err != nil {
			log.Println("error creating inmemory source ", err)
		}
	}

	return &res
}

func buildBlocker(conf configuration.ServerConf) client.Client {
	res := inmemoryclient.InMemoryClient{}
	for _, v := range conf.BlockingList {
		err := res.Add(v, "127.0.0.1")
		if err != nil {
			log.Println("error creating Blocker source ", err)
		}
		err = res.Add(v, "::1")
		if err != nil {
			log.Println("error creating Blocker source ", err)
		}
	}

	return &res
}

//The optimal chain is
// Client(Blocker) -> Client(Memory) -> Client(Cache) -> CacheFeeder((Multiple(Client(udp/https))))
