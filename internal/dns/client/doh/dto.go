package doh

import (
	"net"

	"github.com/bluguard/dnshield/internal/dns/dto"
)

type Message struct {
	Status   int        `json:"status,omitempty"`
	TC       bool       `json:"TC,omitempty"`
	RD       bool       `json:"RD,omitempty"`
	RA       bool       `json:"RA,omitempty"`
	AD       bool       `json:"AD,omitempty"`
	CD       bool       `json:"CD,omitempty"`
	Question []Question `json:"Question,omitempty"`
	Answer   []Answer   `json:"Answer,omitempty"`
}

type Question struct {
	Name string `json:"name,omitempty"`
	Type int    `json:"type,omitempty"`
}

type Answer struct {
	Name string `json:"name,omitempty"`
	Type uint16 `json:"type,omitempty"`
	Ttl  uint32 `json:"TTL,omitempty"`
	Data string `json:"data,omitempty"`
}

func (a Answer) ToRecord() dto.Record {
	ip := parseIp(a.Data)
	return dto.Record{
		Name:  a.Name,
		Type:  dto.Type(a.Type),
		Class: dto.IN,
		TTL:   a.Ttl,
		Data:  ip,
	}
}

func parseIp(addr string) net.IP {
	ip := net.ParseIP(addr)
	v4 := ip.To4()
	if v4 != nil {
		return v4
	}
	return ip.To16()
}
