package dto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"strings"
)

const (
	bufferMaxLength     = 255
	bufferMinLength     = 12
	bufferQuestionStart = 12

	refStartByte = byte(192)
)

var _ error = &BufferTooLongException{0}

//ParseMessage parse a message from a binary representation
func ParseMessage(packet []byte) (*Message, error) {
	if len(packet) > bufferMaxLength {
		return nil, &BufferTooLongException{len(packet)}
	}
	if len(packet) < bufferMinLength {
		return nil, &BufferTooLongException{len(packet)}
	}
	message := &Message{} //create an empty message, it will be filled in future
	if err := parseMetadata(packet, message); err != nil {
		return nil, err
	}
	offset, err := parseQuestion(packet, message)
	if err != nil {
		return nil, err
	}
	if err := parseResponse(packet, message, offset); err != nil {
		return nil, err
	}
	return message, nil
}

func parseMetadata(packet []byte, message *Message) error {
	message.ID = binary.BigEndian.Uint16(packet[0:2])
	message.Header = binary.BigEndian.Uint16(packet[2:4])
	message.QuestionCount = binary.BigEndian.Uint16(packet[4:6])
	message.ResponseCount = binary.BigEndian.Uint16(packet[6:8])
	if binary.BigEndian.Uint16(packet[8:10]) != 0 {
		return errors.New("authority rrs not supported")
	}
	if binary.BigEndian.Uint16(packet[10:12]) != 0 {
		return errors.New("additional rrs not supported")
	}
	return nil
}

func parseQuestion(packet []byte, message *Message) (int, error) {
	readedBytes := 0
	buffer := bytes.NewBuffer(packet[bufferQuestionStart:])

	for i := 0; i < int(message.QuestionCount); i++ {

		question := Question{}

		namestart, err := buffer.ReadByte()
		if err != nil {
			return 0, err
		}
		question.Name, err = readName(namestart, buffer, packet)
		if err != nil {
			return 0, err
		}

		twoBytes := make([]byte, 2)
		n, err := buffer.Read(twoBytes)
		if err != nil {
			return 0, err
		}
		if n != 2 {
			return 0, errors.New("bad read question Type")
		}
		question.Type = Type(binary.BigEndian.Uint16(twoBytes))

		n, err = buffer.Read(twoBytes)
		if err != nil {
			return 0, err
		}
		if n != 2 {
			return 0, errors.New("bad read question class")
		}
		question.Class = Class(binary.BigEndian.Uint16(twoBytes))

		message.Question = append(message.Question, question)
		readedBytes += 6 + len(question.Name)
	}
	return readedBytes + bufferQuestionStart, nil
}

func parseResponse(packet []byte, message *Message, offset int) error {
	buffer := bytes.NewBuffer(packet[offset:])

	for i := 0; i < int(message.ResponseCount); i++ {
		response := Record{}

		namestart, err := buffer.ReadByte()
		if err != nil {
			return err
		}
		response.Name, err = readName(namestart, buffer, packet)
		if err != nil {
			return err
		}

		twoBytes := make([]byte, 2)
		n, err := buffer.Read(twoBytes)
		if err != nil {
			return err
		}
		if n != 2 {
			return errors.New("bad read response type")
		}
		response.Type = Type(binary.BigEndian.Uint16(twoBytes))

		n, err = buffer.Read(twoBytes)
		if err != nil {
			return err
		}
		if n != 2 {
			return errors.New("bad read response class")
		}
		response.Class = Class(binary.BigEndian.Uint16(twoBytes))

		ttlBuffer := make([]byte, 4)
		n, err = buffer.Read(ttlBuffer)
		if err != nil {
			return err
		}
		if n != 4 {
			return errors.New("bad read response TTL")
		}
		response.TTL = binary.BigEndian.Uint32(ttlBuffer)

		n, err = buffer.Read(twoBytes)
		if err != nil {
			return err
		}
		if n != 2 {
			return errors.New("bad read response data length")
		}
		dataLength := binary.BigEndian.Uint16(twoBytes)
		data := make([]byte, dataLength)
		n, err = buffer.Read(data)
		if err != nil {
			return err
		}
		if n != int(dataLength) {
			return errors.New("bad read response data")
		}

		response.Data, err = parseAddress(data, response.Type)
		if err != nil {
			return err
		}

		message.Response = append(message.Response, response)
	}

	return nil
}

func readName(namestart byte, buffer *bytes.Buffer, packet []byte) (string, error) {
	if namestart == refStartByte {
		ref, err := buffer.ReadByte()
		if err != nil {
			return "", err
		}
		parts := make([]byte, 2)
		parts[0] = namestart
		parts[1] = ref
		return parseName(parts, packet), nil
	}
	parts, err := buffer.ReadBytes(0)
	parts = parts[0 : len(parts)-1]
	if err != nil {
		return "", err
	}
	allParts := make([]byte, 0, len(parts)+1)
	allParts = append(allParts, namestart)
	allParts = append(allParts, parts...)
	return parseName(allParts, packet), nil
}

func parseName(parts []byte, packet []byte) string {
	start := parts[0]
	if start == refStartByte {
		buffer := bytes.NewBuffer(packet[parts[1]:])
		bytes, _ := buffer.ReadBytes(0)
		return parseAllParts(bytes[0 : len(bytes)-1])
	}
	return parseAllParts(parts)
}

func parseAllParts(parts []byte) string {
	var sb strings.Builder
	size := int(parts[0])
	sb.Write(parts[1 : size+1])
	if len(parts) > size+1 {
		sb.WriteRune('.')
		sb.WriteString(parseAllParts(parts[size+1:]))
	}
	return sb.String()
}

func parseAddress(data []byte, t Type) (net.IP, error) {
	if t == A && len(data) == net.IPv4len {
		return net.IP(data), nil
	}

	if t == AAAA && len(data) == net.IPv6len {
		return net.IP(data), nil
	}
	return nil, errors.New("bad response type")
}

//BufferTooLongException error returned when the buffer is too long
type BufferTooLongException struct {
	len int
}

//Error returns the string of the current error
func (b *BufferTooLongException) Error() string {
	return "the length of the buffer" + strconv.Itoa(b.len) + "is too long, maximum length is " + strconv.Itoa(bufferMaxLength)
}
