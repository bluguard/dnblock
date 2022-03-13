package cache

import "github.com/bluguard/dnshield/internal/dns/dto"

type Cache interface {
	Insert(dto.Record)
	Retrieve(name string) []dto.Record
}
