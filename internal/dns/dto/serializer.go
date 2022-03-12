package dto

import (
	"bytes"
	"encoding/binary"
	"net"
	"strings"
)

//SerializeMessage serialize a DNS message into a binary representation
func SerializeMessage(message Message) []byte {
	var buffer bytes.Buffer

	writeUint16(message.ID, &buffer)
	writeUint16(message.Header, &buffer)
	writeUint16(message.QuestionCount, &buffer)
	writeUint16(message.ResponseCount, &buffer)
	writeUint32(0, &buffer) // additionals rrs and authority rrs
	for _, question := range message.Question {
		writeQuestion(question, &buffer)
	}

	for _, response := range message.Response {
		writeResponse(response, &buffer)
	}

	return buffer.Bytes()
}

func writeQuestion(question Question, buffer *bytes.Buffer) {
	writeName(question.Name, buffer)
	writeUint16(uint16(question.Type), buffer)
	writeUint16(uint16(question.Class), buffer)
}

func writeResponse(response Record, buffer *bytes.Buffer) {
	writeName(response.Name, buffer)
	writeUint16(uint16(response.Type), buffer)
	writeUint16(uint16(response.Class), buffer)
	writeUint32(response.TTL, buffer)
	writeData(response.Type, response.Data, buffer)
}

func writeName(s string, buffer *bytes.Buffer) {
	nameParts := strings.Split(s, ".")
	for _, p := range nameParts {
		buffer.WriteByte(uint8(len(p)))
		buffer.Write([]byte(p))
	}
	buffer.WriteByte(0)
}

func writeData(t Type, iP net.IP, buffer *bytes.Buffer) {
	switch t {
	case AAAA:
		writeUint16(net.IPv6len, buffer)
		break
	case A:
		writeUint16(net.IPv4len, buffer)
		break
	default:
		writeUint16(net.IPv4len, buffer)
		break
	}
	buffer.Write(iP)
}

func writeUint16(u uint16, buffer *bytes.Buffer) {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, u)
	buffer.Write(bytes)
}

func writeUint32(u uint32, buffer *bytes.Buffer) {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, u)
	buffer.Write(bytes)
}
