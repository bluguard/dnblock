package ristretoimpl

import (
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/bluguard/dnshield/internal/dns/dto"
)

func TestRistretoCache(t *testing.T) {
	c := NewRistrettoCache(1, 60)

	time.Sleep(2 * time.Second)

	type args struct {
		record  dto.Record
		wantErr bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test A",
			args: args{
				record: dto.Record{
					Name:  "localhost",
					Type:  dto.A,
					Class: dto.IN,
					TTL:   60,
					Data:  net.IPv4(127, 0, 0, 1).To4(),
				},
			},
		},
		{
			name: "test AAAA",
			args: args{
				record: dto.Record{
					Name:  "localhost",
					Type:  dto.AAAA,
					Class: dto.IN,
					TTL:   60,
					Data:  net.ParseIP("::1").To16(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Feed(tt.args.record)
			time.Sleep(2 * time.Second)
			var record dto.Record
			var err error
			switch tt.args.record.Type {
			case dto.A:
				irecord, ierr := c.ResolveV4(tt.args.record.Name)
				record = irecord
				err = ierr
			case dto.AAAA:
				irecord, ierr := c.ResolveV6(tt.args.record.Name)
				record = irecord
				err = ierr
			}
			if (err != nil) && !tt.args.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.args.wantErr)
			}
			record.TTL = tt.args.record.TTL
			if !reflect.DeepEqual(record, tt.args.record) {
				t.Errorf("Resolve() = %v, wantErr %v", tt.args.record, record)
			}
		})
	}
}
