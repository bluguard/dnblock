package resolver

import (
	"net"
	"reflect"
	"testing"

	"github.com/bluguard/dnshield/internal/dns/dto"
)

var _ Resolver = resolverMock{}

type resolverMock struct {
}

// Name implements Resolver
func (resolverMock) Name() string {
	return "mock"
}

// Resolve implements Resolver
func (resolverMock) Resolve(question dto.Question) (dto.Record, bool) {
	record := dto.Record{
		Name:  question.Name,
		Type:  question.Type,
		Class: question.Class,
		TTL:   12000,
	}
	if question.Type == dto.A {
		record.Data = net.ParseIP("127.0.0.1").To4()
		return record, true
	} else if question.Type == dto.AAAA {
		record.Data = net.ParseIP("::1:").To16()
		return record, true
	}
	return dto.Record{}, false
}

func TestResolverChain_Resolve(t *testing.T) {
	resolverChain := &ResolverChain{
		chain: []Resolver{resolverMock{}},
	}
	tests := []struct {
		name    string
		message dto.Message
		want    dto.Message
	}{
		{
			name: "localhost A",
			message: dto.Message{
				ID:            1,
				Header:        dto.STANDARD_QUERY,
				QuestionCount: 1,
				ResponseCount: 0,
				Question: []dto.Question{
					{
						Name:  "localhost",
						Type:  dto.A,
						Class: dto.IN,
					},
				},
				Response: []dto.Record{},
			},
			want: dto.Message{
				ID:            1,
				Header:        dto.STANDARD_RESPONSE,
				QuestionCount: 1,
				ResponseCount: 1,
				Question: []dto.Question{
					{
						Name:  "localhost",
						Type:  dto.A,
						Class: dto.IN,
					},
				},
				Response: []dto.Record{
					{
						Name:  "localhost",
						Type:  dto.A,
						Class: dto.IN,
						TTL:   12000,
						Data:  net.ParseIP("127.0.0.1").To4(),
					},
				},
			},
		},
		{
			name: "localhost AAAA",
			message: dto.Message{
				ID:            2,
				Header:        dto.STANDARD_QUERY,
				QuestionCount: 1,
				ResponseCount: 0,
				Question: []dto.Question{
					{
						Name:  "localhost",
						Type:  dto.AAAA,
						Class: dto.IN,
					},
				},
				Response: []dto.Record{},
			},
			want: dto.Message{
				ID:            2,
				Header:        dto.STANDARD_RESPONSE,
				QuestionCount: 1,
				ResponseCount: 1,
				Question: []dto.Question{
					{
						Name:  "localhost",
						Type:  dto.AAAA,
						Class: dto.IN,
					},
				},
				Response: []dto.Record{
					{
						Name:  "localhost",
						Type:  dto.AAAA,
						Class: dto.IN,
						TTL:   12000,
						Data:  net.ParseIP("::1:").To16(),
					},
				},
			},
		},
		{
			name: "localhost unknown",
			message: dto.Message{
				ID:            3,
				Header:        dto.STANDARD_QUERY,
				QuestionCount: 1,
				ResponseCount: 0,
				Question: []dto.Question{
					{
						Name:  "localhost",
						Type:  dto.Type(50),
						Class: dto.IN,
					},
				},
				Response: []dto.Record{},
			},
			want: dto.Message{
				ID:            3,
				Header:        dto.STANDARD_RESPONSE,
				QuestionCount: 1,
				ResponseCount: 0,
				Question: []dto.Question{
					{
						Name:  "localhost",
						Type:  dto.Type(50),
						Class: dto.IN,
					},
				},
				Response: []dto.Record{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := resolverChain.Resolve(tt.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResolverChain.Resolve() = %v, want %v", got, tt.want)
			}
		})
	}
}
