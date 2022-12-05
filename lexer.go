package teal

import (
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	p int // previous index
	i int // current index

	l  int // current line index
	li int // current index in the current line
	lb int // current line begin index in source

	ts   []Token      // tokens
	errs []LexerError // errors

	Source []byte
}

type TokenType string

const (
	TokenEol     = "EOL"   // end of line
	TokenHash    = "#"     // #name
	TokenId      = "Id"    // name
	TokenInt     = "Int"   // int
	TokenBytes   = "Bytes" // bytes
	TokenComment = "Comment"
	TokenString  = "String"
	TokenLabel   = "Label"
)

type Token struct {
	v []byte // value

	l int // line
	b int // begin
	e int // end

	t TokenType
}

func (t Token) String() string {
	return string(t.v)
}

func (t Token) Line() int {
	return t.l
}

func (t Token) Begin() int {
	return t.b
}

func (t Token) End() int {
	return t.e
}

func (t Token) Type() TokenType {
	return t.t
}

func (z *Lexer) inc(n int) {
	z.i += n
	z.li += n
}

func (z *Lexer) emit(t TokenType) {
	z.ts = append(z.ts, Token{
		l: z.l,
		b: z.p - z.lb,
		e: z.i - z.lb,

		v: z.Source[z.p:z.i],
		t: t,
	})

	z.p = z.i

	switch t {
	case TokenEol:
		z.l++
		z.li = 0
		z.lb = z.i
	}
}

func isTerminating(c rune) bool {
	switch c {
	case '\r':
	case '\n':
	case ' ':
	case '\t':
	default:
		return false
	}

	return true
}

type LexerError struct {
	p   int
	msg string
}

func (e LexerError) Error() string {
	return e.msg
}

const (
	numberInt = iota
	numberHex
	numberBin
	numberOct
)

func (z *Lexer) readInt() {
	mode := numberInt

	for {
		if z.i == len(z.Source) {
			z.emit(TokenInt)
			return
		}

		c, n := utf8.DecodeRune(z.Source[z.i:])
		if isTerminating(c) {
			z.emit(TokenInt)
			return
		}

		switch mode {
		case numberInt:
			if !unicode.IsDigit(c) {
				if z.i-z.p == 1 {
					switch c {
					case 'x':
						mode = numberHex
					case 'b':
						mode = numberBin
					default:
						z.errs = append(z.errs, LexerError{p: z.i, msg: "unexpected non-digit"})
						return
					}
				} else {
					z.errs = append(z.errs, LexerError{p: z.i, msg: "unexpected non-digit"})
					return
				}
			}
		case numberBin:
			if c != '0' && c != '1' {
				z.errs = append(z.errs, LexerError{p: z.i, msg: "unexpected non-binary"})
				return
			}
		case numberOct:
			if c < '0' || c > '8' {
				z.errs = append(z.errs, LexerError{p: z.i, msg: "unexpected non-octal"})
				return
			}
		case numberHex:
			if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') && !(c >= 'A' && c <= 'F') {
				z.errs = append(z.errs, LexerError{p: z.i, msg: "unexpected non-hex"})
				return
			}
		}

		z.inc(n)
	}
}

func (z *Lexer) readIdOrLabel() {
	for {
		if z.i == len(z.Source) {
			z.emit(TokenId)
			return
		}

		c, n := utf8.DecodeRune(z.Source[z.i:])
		if z.i-z.p < 1 {
			if unicode.IsDigit(c) {
				z.errs = append(z.errs, LexerError{p: z.i, msg: "unexpected non-letter"})
			}
		} else {
			if c != '_' && !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				if c == ':' {
					z.emit(TokenLabel)
					z.inc(n)
				} else {
					z.emit(TokenId)
				}
				return
			}
		}

		z.inc(n)
	}
}

func (z *Lexer) skipWhitespace() {
	for {
		if z.i == len(z.Source) {
			return
		}

		c, n := utf8.DecodeRune(z.Source[z.i:])
		if c != ' ' && c != '\t' {
			return
		}

		z.inc(n)
		z.p = z.i
	}
}

func (z *Lexer) readComment() {
	for {
		l := z.i - z.p
		if z.i == len(z.Source) {
			if l < 2 {
				z.errs = append(z.errs, LexerError{p: z.p, msg: "incomplete comment"})
				return
			}
			z.p += 2
			z.emit(TokenComment)
			return
		}

		c, n := utf8.DecodeRune(z.Source[z.i:])
		if l < 2 {
			if c != '/' {
				z.errs = append(z.errs, LexerError{p: z.p, msg: "incomplete comment"})
				return
			}
		} else {
			if c == '\r' || c == '\n' {
				z.p += 2
				z.emit(TokenComment)
				return
			}
		}

		z.inc(n)
	}
}

func (z *Lexer) readEol() {
	for {
		if z.i == len(z.Source) {
			z.emit(TokenEol)
			return
		}

		c, n := utf8.DecodeRune(z.Source[z.i:])
		if c == '\r' {
			z.inc(n)
			if z.i < len(z.Source) {
				c2, n2 := utf8.DecodeRune(z.Source[z.i:])
				if c2 == '\n' {
					z.inc(n2)
				}
			}
			z.emit(TokenEol)
			return
		}

		if c == '\n' {
			z.inc(n)
			z.emit(TokenEol)
			return
		}

		z.inc(n)
	}
}

func (z *Lexer) readHash() {
	for {
		l := z.i - z.p
		if z.i == len(z.Source) {
			if l < 2 {
				z.errs = append(z.errs, LexerError{p: z.p, msg: "missing hash name"})
				return
			}
			z.p += 1
			z.emit(TokenHash)
			return
		}

		c, n := utf8.DecodeRune(z.Source[z.i:])

		if l == 0 {
			if c != '#' {
				z.errs = append(z.errs, LexerError{p: z.p, msg: "unexpected non-#"})
				return
			}
		} else {
			if isTerminating(c) {
				z.p += 1
				z.emit(TokenHash)
				return
			}
		}

		z.inc(n)
	}
}

func (z *Lexer) readTokens() {
	z.skipWhitespace()

	if z.i < len(z.Source) {
		c, n := utf8.DecodeRune(z.Source[z.i:])

		var nc rune
		if z.i+n < len(z.Source) {
			nc, _ = utf8.DecodeRune(z.Source[z.i+n:])
		}

		if z.li == 0 {
			if c == '#' {
				z.readHash()
				return
			}
		}

		if c == '/' && nc == '/' {
			z.readComment()
		} else if c == '\n' || c == '\r' {
			z.readEol()
		} else if unicode.IsDigit(c) {
			z.readInt()
		} else if !unicode.IsDigit(c) {
			z.readIdOrLabel()
		} else {
			z.errs = append(z.errs, LexerError{p: z.i, msg: "cannot tokenize"})
		}
	}
}

func (z *Lexer) read() {
	if z.i == len(z.Source) {
		return
	}

	z.readTokens()
}

func (z *Lexer) Next() bool {
	if len(z.ts) > 0 {
		z.ts = z.ts[1:]
	}

	if len(z.ts) == 0 {
		z.read()
	}

	return len(z.ts) > 0
}

func (z *Lexer) Curr() Token {
	if len(z.ts) == 0 {
		return Token{}
	}

	return z.ts[0]
}

func (z *Lexer) Errors() []LexerError {
	return z.errs
}
