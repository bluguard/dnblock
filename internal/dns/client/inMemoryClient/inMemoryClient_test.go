package inmemoryclient

import (
	"net"
	"reflect"
	"testing"
)

var c *InMemoryClient = &InMemoryClient{
	v4Store: map[string]net.IP{
		"localhost": net.IPv4(127, 0, 0, 1),
	},
	v6Store: map[string]net.IP{
		"localhost": net.ParseIP("::1:").To16(),
	},
}

func TestInMemoryClient_ResolveV4(t *testing.T) {
	tests := []struct {
		name    string
		want    net.IP
		wantErr bool
	}{
		{
			name:    "localhost",
			want:    net.IPv4(127, 0, 0, 1),
			wantErr: false,
		},
		{
			name:    "unknown",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := c.ResolveV4(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("InMemoryClient.ResolveV4() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InMemoryClient.ResolveV4() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryClient_ResolveV6(t *testing.T) {
	tests := []struct {
		name    string
		want    net.IP
		wantErr bool
	}{
		{
			name:    "localhost",
			want:    net.ParseIP("::1:").To16(),
			wantErr: false,
		},
		{
			name:    "unknown",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := c.ResolveV6(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("InMemoryClient.ResolveV4() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InMemoryClient.ResolveV4() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryClient_Add(t *testing.T) {
	type args struct {
		name    string
		address net.IP
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		checkMethod func(string) (net.IP, error)
	}{
		{
			name:        "google.com v4",
			args:        args{name: "google.com", address: net.ParseIP("142.250.184.206").To4()},
			wantErr:     false,
			checkMethod: c.ResolveV4,
		},
		{
			name:        "google.lcom v6",
			args:        args{name: "google.com", address: net.ParseIP("2a00:1450:4001:830::200e").To16()},
			wantErr:     false,
			checkMethod: c.ResolveV6,
		},
		{
			name:        "google.com wrong format",
			args:        args{name: "google.com", address: net.ParseIP("142.250.184.206.123")},
			wantErr:     true,
			checkMethod: c.ResolveV4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.args.name, func(t *testing.T) {
			if err := c.Add(tt.args.name, tt.args.address.String()); (err != nil) != tt.wantErr {
				t.Errorf("InMemoryClient.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if ip, err := tt.checkMethod(tt.args.name); err == nil {
				ip.Equal(tt.args.address)
			}
		})
	}
}
