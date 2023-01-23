package teal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImmLexer(t *testing.T) {
	type test struct {
		i string
		o []immArg
	}

	tests := []test{
		{
			i: "{uint8 curve index}",
			o: []immArg{
				{
					kind: immField,
					t:    "uint8",
					n:    "curve index",
				},
			},
		},
		{
			i: "{uint8 transaction group index} {uint8 transaction field index}",
			o: []immArg{
				{
					kind: immField,
					t:    "uint8",
					n:    "transaction group index",
				},
				{
					kind: immField,
					t:    "uint8",
					n:    "transaction field index",
				},
			},
		},
		{
			i: "{uint8 transaction group index} {uint8 transaction field index} {uint8 transaction field array index}",
			o: []immArg{
				{
					kind: immField,
					t:    "uint8",
					n:    "transaction group index",
				},
				{
					kind: immField,
					t:    "uint8",
					n:    "transaction field index",
				},
				{
					kind: immField,
					t:    "uint8",
					n:    "transaction field array index",
				},
			},
		},
		{
			i: "{uint8 branch count} [{int16 branch offset, big-endian}, ...]",
			o: []immArg{
				{
					kind: immField,
					t:    "uint8",
					n:    "branch count",
				},
				{
					kind: immArray,
					r: []immArg{
						{
							kind: immField,
							t:    "int16",
							n:    "branch offset, big-endian",
						},
						{
							kind: immVar,
						},
					},
				},
			},
		},
		{
			i: "{varuint count} [({varuint value length} bytes), ...]",
			o: []immArg{
				{
					kind: immField,
					t:    "varuint",
					n:    "count",
				},
				{
					kind: immArray,
					r: []immArg{
						{
							kind: immField,
							n:    "bytes",
							sub: &immArg{
								kind: immField,
								t:    "varuint",
								n:    "value length",
							},
						},
						{
							kind: immVar,
						},
					},
				},
			},
		},
	}

	for _, ts := range tests {
		l := &immLexer{
			s: []byte(ts.i),
		}

		args, err := l.tokenize()
		if err != nil {
			t.Error(err)
		}

		if !assert.Len(t, args, len(ts.o)) {
			return
		}

		for i, targs := range ts.o {
			a := args[i]
			if targs.kind != a.kind {
				t.Error("unexpected arg kind")
			}

			if len(targs.r) != len(a.r) {
				t.Error("unexpected array length")
			}

			for j, ta := range targs.r {
				ra := a.r[j]

				assert.Equal(t, ta.kind, ra.kind)
				assert.Equal(t, ta.t, ra.t)
				assert.Equal(t, ta.n, ra.n)

				if ta.sub == nil {
					assert.Equal(t, ta.sub, ra.sub)
				}

				if ta.sub != nil {
					assert.Equal(t, ta.kind, ra.kind)
					assert.Equal(t, ta.n, ra.n)
					assert.Equal(t, ta.t, ra.t)
				}
			}
		}
	}
}
