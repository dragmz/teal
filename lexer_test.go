package teal

import (
	"testing"
)

func TestLexer(t *testing.T) {
	type test struct {
		i string
		o []TokenType
		v []string
	}

	tests := []test{
		{
			i: "\"\"",
			o: []TokenType{TokenValue},
			v: []string{"\"\""},
		},
		{
			i: "byte \"escape\\\" sequences\\\"\"",
			o: []TokenType{TokenValue, TokenValue},
			v: []string{"byte", "\"escape\" sequences\"\""},
		},
		{
			i: "bytecblock 0x6f 0x65 0x70 0x6131 0x6132 0x6c74 0x73776170 0x6d696e74 0x74 0x7031 0x7032",
			o: []TokenType{TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue},
			v: []string{"bytecblock", "0x6f", "0x65", "0x70", "0x6131", "0x6132", "0x6c74", "0x73776170", "0x6d696e74", "0x74", "0x7031", "0x7032"},
		},
		{
			i: "12345 0x123",
			o: []TokenType{TokenValue, TokenValue},
			v: []string{"12345", "0x123"},
		},
		{
			i: "12345",
			o: []TokenType{TokenValue},
			v: []string{"12345"},
		},
		{
			i: "a12345",
			o: []TokenType{TokenValue},
			v: []string{"a12345"},
		},
		{
			i: "\r\n",
			o: []TokenType{TokenEol},
			v: []string{"\r\n"},
		},
		{
			i: "\r",
			o: []TokenType{TokenEol},
			v: []string{"\r"},
		},
		{
			i: "\n\r",
			o: []TokenType{TokenEol, TokenEol},
			v: []string{"\n", "\r"},
		},
		{
			i: "\r\n\r\n",
			o: []TokenType{TokenEol, TokenEol},
			v: []string{"\r\n", "\r\n"},
		},
		{
			i: "\r\n\n\r\n",
			o: []TokenType{TokenEol, TokenEol, TokenEol},
			v: []string{"\r\n", "\n", "\r\n"},
		},
		{
			i: "",
			o: []TokenType{},
			v: []string{},
		},
		{
			i: "#pragma version 8",
			o: []TokenType{TokenValue, TokenValue, TokenValue},
			v: []string{"#pragma", "version", "8"},
		},
		{
			i: "byte \"some multiword byte string\"",
			o: []TokenType{TokenValue, TokenValue},
			v: []string{"byte", "\"some multiword byte string\""},
		},
	}

	for _, ts := range tests {
		z := Lexer{
			Source: []byte(ts.i),
		}

		var a []Token
		for z.Scan() {
			if len(z.Errors()) > 0 {
				for _, err := range z.Errors() {
					t.Error(err)
				}
			}
			a = append(a, z.Curr())
		}

		if len(a) != len(ts.o) {
			t.Errorf("unexpected output types length: %d != %d in %s", len(a), len(ts.o), ts)
		}

		if len(a) != len(ts.v) {
			t.Errorf("unexpected output values length: %d != %d in %s", len(a), len(ts.v), ts)
		}

		for i := 0; i < len(a); i++ {
			if a[i].Type() != ts.o[i] {
				t.Errorf("unexpected token type: %s != %s in %s", a[i].Type(), ts.o[i], ts)
			}
			if a[i].String() != ts.v[i] {
				t.Errorf("unexpected token value: %s != %s in %s", a[i].String(), ts.v[i], ts)
			}
		}
	}
}
