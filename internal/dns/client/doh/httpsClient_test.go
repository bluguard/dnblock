package doh

import (
	"reflect"
	"testing"

	"github.com/bluguard/dnshield/internal/dns/dto"
)

func TestDOHClient_ResolveV4(t *testing.T) {
	type fields struct {
		endpoint string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantEmpty bool
		wantErr   bool
	}{
		{
			name:      "google.com v4",
			fields:    fields{endpoint: "https://cloudflare-dns.com/dns-query"},
			args:      args{name: "google.com"},
			wantEmpty: false,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewDOHClient(tt.fields.endpoint)
			got, err := c.ResolveV4(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DOHClient.ResolveV4() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantEmpty && !reflect.DeepEqual(got, dto.Record{}) {
				t.Errorf("DOHClient.ResolveV4() = %v, want empty", got)
			}
		})
	}
}

func TestDOHClient_ResolveV6(t *testing.T) {
	type fields struct {
		endpoint string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantEmpty bool
		wantErr   bool
	}{
		{
			name:      "google.com v6",
			fields:    fields{endpoint: "https://cloudflare-dns.com/dns-query"},
			args:      args{name: "google.com"},
			wantEmpty: false,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewDOHClient(tt.fields.endpoint)
			got, err := c.ResolveV6(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DOHClient.ResolveV4() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantEmpty && !reflect.DeepEqual(got, dto.Record{}) {
				t.Errorf("DOHClient.ResolveV4() = %v, want empty", got)
			}
		})
	}
}
