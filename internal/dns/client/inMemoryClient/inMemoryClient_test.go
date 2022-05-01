package inmemoryclient

import (
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/bluguard/dnshield/internal/dns/dto"
)

var c *InMemoryClient

func TestMain(m *testing.M) {
	c = &InMemoryClient{}
	c.Add("localhost", "127.0.0.1")  //ipv4
	c.Add("localhost", "::1")        //ipv6
	c.Add("unknown", "192897347459") //not an ip
	os.Exit(m.Run())
}

func TestInMemoryClient_ResolveV4(t *testing.T) {

	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    dto.Record
		wantErr bool
	}{
		{
			name: "localhost v4",
			args: args{name: "localhost"},
			want: dto.Record{
				Name:  "localhost",
				Type:  dto.A,
				Class: dto.IN,
				TTL:   200,
				Data:  net.ParseIP("127.0.0.1").To4(),
			},
			wantErr: false,
		},
		{
			name:    "unknown v4",
			args:    args{name: "unknown"},
			want:    dto.Record{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.ResolveV4(tt.args.name)
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
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    dto.Record
		wantErr bool
	}{
		{
			name: "localhost v6",
			args: args{name: "localhost"},
			want: dto.Record{
				Name:  "localhost",
				Type:  dto.AAAA,
				Class: dto.IN,
				TTL:   200,
				Data:  net.ParseIP("::1").To16(),
			},
			wantErr: false,
		},
		{
			name:    "unknown v6",
			args:    args{name: "unknown"},
			want:    dto.Record{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.ResolveV6(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("InMemoryClient.ResolveV6() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InMemoryClient.ResolveV6() = %v, want %v", got, tt.want)
			}
		})
	}
}
