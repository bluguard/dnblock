package udpendpoint

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	inmemoryclient "github.com/bluguard/dnshield/internal/dns/client/inMemoryClient"
	"github.com/bluguard/dnshield/internal/dns/client/udp"
	"github.com/bluguard/dnshield/internal/dns/resolver"
)

const addr = "127.0.0.1:12349"

var client udp.UdpClient

func TestMain(m *testing.M) {

	client.Address = addr

	memoryClient := inmemoryclient.InMemoryClient{}
	memoryClient.Add("localhost", "127.0.0.1")
	memoryClient.Add("localhost", "::1")

	chain := resolver.NewResolverChain([]resolver.Resolver{
		resolver.NewClientresolver(&memoryClient, "inMemory"),
	})

	endpoint := NewUdpEndpoint(addr, chain)

	endpoint.SetChain(chain)

	//start endpoint
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	endpoint.Start(ctx, &wg)

	time.Sleep(100 * time.Millisecond)

	res := m.Run()
	cancel()
	wg.Wait()
	os.Exit(res)
}

func TestUdpEndpoint(t *testing.T) {
	res, err := client.ResolveV4("localhost")
	if err != nil {
		t.Fatalf("error resolving localhost in v4 %v", err)
	}
	if res.Name != "localhost" || res.Data.String() != "127.0.0.1" {
		t.Fatalf("Expecting localhost -> 127.0.0.1, got %v", res)
	}

	res, err = client.ResolveV6("localhost")
	if err != nil {
		t.Fatalf("error resolving localhost in v6 %v", err)
	}
	if res.Name != "localhost" || res.Data.String() != "::1" {
		t.Fatalf("Expecting localhost -> ::1, got %v", res)
	}
}
