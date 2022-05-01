package resolver

import (
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/bluguard/dnshield/internal/dns/client"
	"github.com/bluguard/dnshield/internal/dns/dto"
)

var _ client.Client = MockClient{}

type MockClient struct {
	v4Count int
	v6Count int
}

// ResolveV4 implements client.Client
func (m MockClient) ResolveV4(name string) (dto.Record, error) {
	m.v4Count++
	if name == "localhost" {
		return dto.Record{
			Name:  "localhost",
			Type:  dto.A,
			Class: dto.IN,
			TTL:   200,
			Data:  net.ParseIP("127.0.0.1").To4(),
		}, nil
	}
	return dto.Record{}, errors.New("unknown")
}

// ResolveV6 implements client.Client
func (m MockClient) ResolveV6(name string) (dto.Record, error) {
	m.v6Count++
	return dto.Record{}, errors.New("unsuported")
}

func TestClientResolver_Resolve(t *testing.T) {

	resolver := &ClientResolver{
		name:   "test",
		client: MockClient{},
	}

	tests := []struct {
		name     string
		question dto.Question
		want     dto.Record
		ok       bool
	}{
		{
			name: "localhost v4",
			question: dto.Question{
				Name:  "localhost",
				Type:  dto.A,
				Class: dto.IN,
			},
			want: dto.Record{
				Name:  "localhost",
				Type:  dto.A,
				Class: dto.IN,
				TTL:   200,
				Data:  net.ParseIP("127.0.0.1").To4(),
			},
			ok: true,
		},
		{
			name: "localhost v6",
			question: dto.Question{
				Name:  "localhost",
				Type:  dto.AAAA,
				Class: dto.IN,
			},
			want: dto.Record{},
			ok:   false,
		},
		{
			name: "localhost unknown",
			question: dto.Question{
				Name:  "localhost",
				Type:  dto.Type(50),
				Class: dto.IN,
			},
			want: dto.Record{},
			ok:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, ok := resolver.Resolve(tt.question)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientResolver.Resolve() got = %v, want %v", got, tt.want)
			}
			if ok != tt.ok {
				t.Errorf("ClientResolver.Resolve() ok = %v, want %v", ok, tt.ok)
			}
		})
	}
}
