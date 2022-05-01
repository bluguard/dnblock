package udp

import (
	"net"
	"reflect"
	"testing"

	"github.com/bluguard/dnshield/internal/dns/dto"
)

func TestUdpClient_ResolveV4(t *testing.T) {
	type fields struct {
		Address string
		id      uint16
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
			name: "google.com",
			fields: fields{
				id:      0,
				Address: "1.1.1.1:53",
			},
			args: args{
				name: "google.com",
			},
			wantEmpty: false,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &UdpClient{
				Address: tt.fields.Address,
				id:      tt.fields.id,
			}
			got, err := c.ResolveV4(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("UdpClient.ResolveV4() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantEmpty && !reflect.DeepEqual(got, dto.Record{}) {
				t.Errorf("UdpClient.ResolveV4() = %v, want empty", got)
			}
			if nil == net.ParseIP(got.Data.String()).To4() {
				t.Errorf("ip is not a V4, got %v", got.Data)
			}
		})
	}
}

func TestUdpClient_ResolveV6(t *testing.T) {
	type fields struct {
		Address string
		id      uint16
	}
	type args struct {
		name string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantempty bool
		wantErr   bool
	}{
		{
			name:      "google.com",
			fields:    fields{id: 0, Address: "1.1.1.1:53"},
			args:      args{name: "google.com"},
			wantempty: false,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &UdpClient{
				Address: tt.fields.Address,
				id:      tt.fields.id,
			}
			got, err := c.ResolveV6(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("UdpClient.ResolveV6() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantempty && !reflect.DeepEqual(got, dto.Record{}) {
				t.Errorf("UdpClient.ResolveV6() = %v, want empty", got)
			}
			if nil == net.ParseIP(got.Data.String()).To16() {
				t.Errorf("ip is not a V6, got %v", got.Data)
			}
		})
	}
}
