package teal

import (
	"testing"
)

func TestLexer(t *testing.T) {
	type test struct {
		i string
		o []TokenType
	}

	tests := []test{
		{
			i: "bytecblock 0x6f 0x65 0x70 0x6131 0x6132 0x6c74 0x73776170 0x6d696e74 0x74 0x7031 0x7032",
			o: []TokenType{TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue, TokenValue},
		},
		{
			i: "12345 0x123",
			o: []TokenType{TokenValue, TokenValue},
		},
		{
			i: "12345",
			o: []TokenType{TokenValue},
		},
		{
			i: "a12345",
			o: []TokenType{TokenValue},
		},
		{
			i: "\r\n",
			o: []TokenType{TokenEol},
		},
		{
			i: "\r",
			o: []TokenType{TokenEol},
		},
		{
			i: "\n\r",
			o: []TokenType{TokenEol, TokenEol},
		},
		{
			i: "\r\n\r\n",
			o: []TokenType{TokenEol, TokenEol},
		},
		{
			i: "\r\n\n\r\n",
			o: []TokenType{TokenEol, TokenEol, TokenEol},
		},
		{
			i: "",
			o: []TokenType{},
		},
		{
			i: "#pragma version 8",
			o: []TokenType{TokenValue, TokenValue, TokenValue},
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
			t.Error("unexpected output length")
		}

		for i := 0; i < len(a); i++ {
			if a[i].Type() != ts.o[i] {
				t.Error("unexpected token")
			}
		}
	}
}
