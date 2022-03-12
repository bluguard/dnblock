package dto

import "net"

type Type uint16
type Class uint16

const (
	A    Type = 1
	AAAA Type = 28

	IN Class = 1

	STANDARD_QUERY    uint16 = 0x0100
	STANDARD_RESPONSE uint16 = 0x8180
)

//Message represent a simplify dns message
type Message struct {
	ID            uint16
	Header        uint16
	QuestionCount uint16
	ResponseCount uint16
	Question      []Question
	Response      []Record
}

//Question is a representation of a dns question
type Question struct {
	Name  string
	Type  Type
	Class Class
}

//Record is a representation of a dns record
type Record struct {
	Name  string
	Type  Type
	Class Class
	TTL   uint32
	Data  net.IP
}
