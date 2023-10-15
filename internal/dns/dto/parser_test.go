package dto_test

import (
	"encoding/hex"
	"net"
	"testing"

	"github.com/bluguard/dnshield/internal/dns/dto"
)

func decodeString(s string) []byte {
	res, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return res
}

type testCase struct {
	name string
	in   []byte
	out  dto.Message
}

var parseTest = []testCase{
	{
		name: "google.fr request type A",
		in:   decodeString("da500100000100000000000006676f6f676c650266720000010001"),
		out: dto.Message{
			ID:            55888,
			Header:        256,
			QuestionCount: 1,
			ResponseCount: 0,
			Question:      []dto.Question{{Name: "google.fr", Type: dto.A, Class: dto.IN}},
		},
	},
	{
		name: "google.fr request type AAAA",
		in:   decodeString("e55d0100000100000000000006676f6f676c6502667200001c0001"),
		out: dto.Message{
			ID:            58717,
			Header:        256,
			QuestionCount: 1,
			ResponseCount: 0,
			Question:      []dto.Question{{Name: "google.fr", Type: dto.AAAA, Class: dto.IN}},
		},
	},

	{
		name: "google.com response type A",
		in:   decodeString("96a68180000100010000000006676f6f676c6503636f6d0000010001c00c00010001000000d400048efab8ce"),

		out: dto.Message{
			ID:            0x96a6,
			Header:        0x8180,
			QuestionCount: 1,
			ResponseCount: 1,
			Question:      []dto.Question{{Name: "google.com", Type: dto.A, Class: dto.IN}},
			Response:      []dto.Record{{Name: "google.com", Type: dto.A, Class: dto.IN, TTL: 212, Data: (net.ParseIP("142.250.184.206").To4())}},
		},
	},
	{
		name: "google.com response type AAAA",
		in:   decodeString("80a48180000100010000000006676f6f676c6503636f6d00001c0001c00c001c00010000003200102a00145040010830000000000000200e"),

		out: dto.Message{
			ID:            0x80a4,
			Header:        0x8180,
			QuestionCount: 1,
			ResponseCount: 1,
			Question:      []dto.Question{{Name: "google.com", Type: dto.AAAA, Class: dto.IN}},
			Response:      []dto.Record{{Name: "google.com", Type: dto.AAAA, Class: dto.IN, TTL: 50, Data: (net.ParseIP("2a00:1450:4001:830::200e").To16())}},
		},
	},
	{
		name: "youtube.com",
		in:   decodeString("00038180000100010000000007796f757475626503636f6d000001000107796f757475626503636f6d00000100010000003c00048efac90e"),
		out: dto.Message{
			ID:            0x0003,
			Header:        0x8180,
			QuestionCount: 1,
			ResponseCount: 1,
			Question:      []dto.Question{{Name: "youtube.com", Type: dto.A, Class: dto.IN}},
			Response:      []dto.Record{{Name: "youtube.com", Type: dto.A, Class: dto.IN, TTL: 60, Data: (net.ParseIP("142.250.201.14").To4())}},
		},
	},
}

func TestParseRequest(t *testing.T) {

	for _, test := range parseTest {
		t.Run(test.name, func(t2 *testing.T) { testParse(test.in, test.out, t2) })
	}

}

func testParse(in []byte, out dto.Message, t *testing.T) {
	message, err := dto.ParseMessage(in)
	if err != nil {
		t.Fatal(err)
	}
	assertMessageEquals(message, out, t)
}

func TestSerializeRequest(t *testing.T) {

	for _, test := range parseTest {
		t.Run(test.name, func(t2 *testing.T) { testSerialize(test.out, t2) })
	}

}

func testSerialize(message dto.Message, t *testing.T) {
	m, err := dto.ParseMessage(dto.SerializeMessage(message))
	if err != nil {
		t.Fatal(err)
	}
	assertMessageEquals(m, message, t)
}

func assertMessageEquals(message *dto.Message, out dto.Message, t *testing.T) {
	if message.ID != out.ID {
		t.Fatal("missmatch ID")
	}
	if message.Header != out.Header {
		t.Fatal("missmatch Header")
	}
	if message.QuestionCount != out.QuestionCount {
		t.Fatal("missmatch QuestionCount")
	}
	if message.ResponseCount != out.ResponseCount {
		t.Fatal("missmatch ResponseCount")
	}
	if len(message.Question) != len(out.Question) {
		t.Fatal("Question number missmatch")
	}
	for i, question := range message.Question {
		o := out.Question[i]
		if question.Name != o.Name {
			t.Fatal("missmatch question name")
		}
		if question.Class != o.Class {
			t.Fatal("missmatch question Class")
		}
		if question.Type != o.Type {
			t.Fatal("missmatch question Type")
		}
	}
	if len(message.Response) != len(out.Response) {
		t.Fatal("Question number missmatch")
	}
	for i, response := range message.Response {
		o := out.Response[i]
		if response.Name != o.Name {
			t.Fatal("missmatch response name")
		}
		if response.Class != o.Class {
			t.Fatal("missmatch response Class")
		}
		if response.Type != o.Type {
			t.Fatal("missmatch response Type")
		}
		if response.TTL != o.TTL {
			t.Fatal("missmatch TTL")
		}
		if !net.IP.Equal(o.Data, response.Data) {
			t.Fatal("mismatch data")
		}
	}
}

var benchCase = testCase{
	name: "google.com response type A",
	in:   decodeString("96a68180000100010000000006676f6f676c6503636f6d0000010001c00c00010001000000d400048efab8ce"),

	out: dto.Message{
		ID:            0x96a6,
		Header:        0x8180,
		QuestionCount: 1,
		ResponseCount: 1,
		Question:      []dto.Question{{Name: "google.com", Type: dto.A, Class: dto.IN}},
		Response:      []dto.Record{{Name: "google.com", Type: dto.A, Class: dto.IN, TTL: 212, Data: net.ParseIP("142.250.186.46")}},
	},
}

func BenchmarkParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dto.ParseMessage(benchCase.in)
	}
}
