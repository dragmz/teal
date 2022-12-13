package teal

import "testing"

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

		if len(args) != len(ts.o) {
			t.Error("unepxected number of args")
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

				if ta.kind != ra.kind {
					t.Error("unexpected array arg kind")
				}

				if ta.t != ra.t {
					t.Error("unexpected array arg type")
				}

				if ta.n != ra.n {
					t.Error("unexpected array arg name")
				}

				if ta.sub == nil && ra.sub != nil {
					t.Error("ta != ra != nil")
				}

				if ta.sub != nil {
					if ta.kind != ra.kind {
						t.Error("ta kind != ra.kind")
					}

					if ta.n != ra.n {
						t.Error("ta.n != ra.n")
					}

					if ta.t != ra.t {
						t.Error("ta.t != ra.t")
					}
				}
			}
		}
	}
}
