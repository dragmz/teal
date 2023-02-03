package teal

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
)

type LangOp struct {
	Opcode  byte
	Name    string
	Args    string `json:",omitempty"`
	Returns string `json:",omitempty"`
	Size    int

	ArgEnum      []string `json:",omitempty"`
	ArgEnumTypes string   `json:",omitempty"`

	Doc           string
	DocExtra      string `json:",omitempty"`
	ImmediateNote string `json:",omitempty"`
	Version       int
	Groups        []string
}

//go:embed specs/langspec_v1.json
var BuiltInLangSpecV1 string

//go:embed specs/langspec_v2.json
var BuiltInLangSpecV2 string

//go:embed specs/langspec_v3.json
var BuiltInLangSpecV3 string

//go:embed specs/langspec_v4.json
var BuiltInLangSpecV4 string

//go:embed specs/langspec_v5.json
var BuiltInLangSpecV5 string

//go:embed specs/langspec_v6.json
var BuiltInLangSpecV6 string

//go:embed specs/langspec_v7.json
var BuiltInLangSpecV7 string

//go:embed specs/langspec_v8.json
var BuiltInLangSpecV8 string

//go:embed specs/langspec_v9.json
var BuiltInLangSpecV9 string

var BuiltInLangSpecs = []string{
	BuiltInLangSpecV1,
	BuiltInLangSpecV2,
	BuiltInLangSpecV3,
	BuiltInLangSpecV4,
	BuiltInLangSpecV5,
	BuiltInLangSpecV6,
	BuiltInLangSpecV7,
	BuiltInLangSpecV8,
	BuiltInLangSpecV9,
}

//go:embed langspec.json
var BuiltInLangSpecJson string

type LangSpec struct {
	EvalMaxVersion  int
	LogicSigVersion uint64
	Ops             []LangOp
}

var BuiltInLangSpec = func() LangSpec {
	spec, err := readLangSpec(strings.NewReader(BuiltInLangSpecJson))
	if err != nil {
		panic(err)
	}

	return spec
}()

func readLangSpec(r io.Reader) (LangSpec, error) {
	var spec LangSpec

	d := json.NewDecoder(r)
	err := d.Decode(&spec)

	return spec, err
}

type immArgKind uint8

const (
	immNone = iota
	immField
	immArray
	immVar
)

type immArg struct {
	kind immArgKind

	t string
	n string

	sub *immArg

	v string
	r []immArg
}

type immLexer struct {
	args []immArg
	i    int
	p    int
	s    []byte
}

func parseField(s []byte) (immArg, int, error) {
	c, _ := utf8.DecodeRune(s)
	switch c {
	case '(':
		return parseFieldBytesArg(s)
	case '{':
		return parseFieldArg(s)
	case '.':
		return parseVarArg(s)
	default:
		return immArg{}, 0, errors.New("failed to parse")
	}
}

func parseVarArg(s []byte) (immArg, int, error) {
	i := 0

	for {
		if i == len(s) {
			return immArg{}, i, errors.New("failed to parse var arg")
		}

		c, n := utf8.DecodeRune(s[i:])
		if c != '.' {
			return immArg{}, 0, errors.New("unexpected character")
		}

		i += n
		if i == 3 {
			return immArg{
				kind: immVar,
				v:    "",
			}, i, nil
		}
	}

}

func parseFieldBytesArg(s []byte) (immArg, int, error) {
	i := 0

	c, n := utf8.DecodeRune(s[i:])
	if c != '(' {
		return immArg{}, i, errors.New("unexpected character")
	}

	i += n
	p := i

	for {
		if i == len(s) {
			return immArg{}, i, errors.New("invalid field")
		}

		c, n := utf8.DecodeRune(s[i:])
		if c == ')' {
			v := s[p:i]
			i += n

			f, n, err := parseField(v)
			if err != nil {
				return immArg{}, i, err
			}

			j := n
			j += skipWhitespace(v[j:])

			arg := immArg{
				kind: immField,
				n:    string(v[j:]),
				sub:  &f,
				v:    fmt.Sprintf("{%s}", string(v[j:])),
			}

			return arg, i, nil
		}

		i += n
	}

}

func parseFieldArg(s []byte) (immArg, int, error) {
	i := 0

	c, n := utf8.DecodeRune(s[i:])
	if c != '{' {
		return immArg{}, i, errors.New("unexpected character")
	}

	i += n
	p := i

	for {
		if i == len(s) {
			return immArg{}, i, errors.New("invalid field")
		}

		c, n := utf8.DecodeRune(s[i:])
		if c == '}' {
			v := string(s[p:i])

			parts := strings.SplitN(v, " ", 2)

			i += n

			var t string
			var n string

			if len(parts) > 0 {
				t = parts[0]
			}

			n = parts[len(parts)-1]

			arg := immArg{
				kind: immField,
				t:    t,
				n:    n,
				v:    fmt.Sprintf("{%s}", string(v)),
			}

			return arg, i, nil
		}

		i += n
	}

}

func (p *immLexer) readField() error {
	arg, n, err := parseField(p.s[p.i:])

	p.i += n
	p.p = n

	p.args = append(p.args, arg)

	if err != nil {
		return err
	}

	return nil
}

func (p *immLexer) readArray() error {
	c, n := utf8.DecodeRune(p.s[p.i:])
	if c != '[' {
		return errors.New("unexpected character")
	}

	p.i += n
	p.p = p.i

	for {
		if p.i == len(p.s) {
			return errors.New("invalid field")
		}

		c, n := utf8.DecodeRune(p.s[p.i:])
		if c == ']' {
			v := p.s[p.p:p.i]

			r := []immArg{}
			{
				i := 0
				for {
					if i == len(v) {
						break
					}

					arg, n, err := parseField(v[i:])
					if err != nil {
						return err
					}

					r = append(r, arg)

					i += n

					i += skipWhitespace(v[i:])

					if i < len(v) {
						c, n := utf8.DecodeRune(v[i:])
						if c != ',' {
							return errors.New("unexpected char")
						}

						i += n
					}

					i += skipWhitespace(v[i:])
				}
			}

			p.i += n

			p.args = append(p.args, immArg{
				kind: immArray,
				v:    fmt.Sprintf("{%s}", string(v)),
				r:    r,
			})

			p.p = p.i
			return nil
		}

		p.i += n
	}

}

func skipWhitespace(s []byte) int {
	i := 0

	if len(s) > 0 {
		for {
			c, n := utf8.DecodeRune(s[i:])
			if c == ' ' {
				i += n
			} else {
				break
			}
		}
	}

	return i
}

// intcblock: {varuint count} [{varuint value}, ...]
// intc: {uint8 int constant index}
func (p *immLexer) tokenize() ([]immArg, error) {
	for p.i < len(p.s) {
		p.i += skipWhitespace(p.s[p.i:])
		p.p = p.i

		c, _ := utf8.DecodeRune(p.s[p.i:])

		var err error
		switch c {
		case '{':
			err = p.readField()
		case '[':
			err = p.readArray()
		default:
			err = errors.Errorf("failed to parse - unexpected char: '%c", c)
		}

		if err != nil {
			return nil, err
		}
	}

	return p.args, nil
}

type ImmArg struct {
	Name  string
	Array bool
}

type ImmNote struct {
	Items []ImmArg
}

func parseImmArgs(s string) ([]ImmArg, error) {
	res := []ImmArg{}

	l := &immLexer{
		s: []byte(s),
	}

	ts, err := l.tokenize()
	if err != nil {
		return res, err
	}

	i := 0
	for {
		if i == len(ts) {
			break
		}

		t := ts[i]
		i += 1

		var array bool

		if i < len(ts) {
			nt := ts[i]
			switch nt.kind {
			case immArray:
				t = nt
				i += 1
				if len(t.r) > 0 {
					f := t.r[0]
					t = f
				}
				array = true
			case immField:
				if nt.n == "bytes" {
					t = nt
					i += 1
				}
			}
		}

		if t.n == "branch offset, big-endian" {
			t.v = "{label}"
		}

		if array {
			t.v = fmt.Sprintf("[%s, ...]", t.v)
		}

		res = append(res, ImmArg{
			Name:  t.v,
			Array: array,
		})
	}

	return res, nil
}
