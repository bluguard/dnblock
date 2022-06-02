package memorycache

import (
	"context"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/bluguard/dnshield/internal/dns/cache"
	"github.com/bluguard/dnshield/internal/dns/client"
	"github.com/bluguard/dnshield/internal/dns/dto"
)

func TestMemoryCache(t *testing.T) {
	ctx, cancelfunc := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	memCache := NewMemoryCache(1000, 1, ctx, wg, time.Second*1)

	feedable := cache.Feedable(memCache)

	cl := client.Client(memCache)

	wantv6 := dto.Record{Name: "google.com", Type: dto.AAAA, Class: dto.IN, TTL: 60, Data: net.ParseIP("::1").To16()}
	wantv4 := dto.Record{Name: "google.com", Type: dto.A, Class: dto.IN, TTL: 1, Data: net.ParseIP("127.0.0.1").To4()}

	feedable.Feed(wantv6)
	feedable.Feed(wantv4)
	wantv4.TTL = 60

	res, err := cl.ResolveV4("google.com")
	if err != nil {
		t.Fatalf("error resolving v4 " + err.Error())
	}

	if !reflect.DeepEqual(res, wantv4) {
		t.Fatalf("error resolving v4 %v, got %v ", wantv4, res)
	}

	res, err = cl.ResolveV6("google.com")
	if err != nil {
		t.Fatalf("error resolving v6 " + err.Error())
	}

	if !reflect.DeepEqual(res, wantv6) {
		t.Fatalf("error resolving v6 %v, got %v ", wantv6, res)
	}

	time.Sleep(1 * time.Second)

	_, err = cl.ResolveV4("google.com")
	if err == nil {
		t.Fatalf("it should have no more v4 entry in the cache")
	}

	_, err = cl.ResolveV6("google.com")
	if err != nil {
		t.Fatalf("it should still have v6 entry in the cache")
	}

	cancelfunc()
	wg.Wait()
}
