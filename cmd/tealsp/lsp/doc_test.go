package lsp

import (
	"testing"
)

func TestSplit(t *testing.T) {
	type test struct {
		i string
		o []string
	}

	tests := []test{
		{
			i: "",
			o: []string{""},
		},
		{
			i: "\r",
			o: []string{"", ""},
		},
		{
			i: "\r\n",
			o: []string{"", ""},
		},
		{
			i: "\n",
			o: []string{"", ""},
		},
		{
			i: "\n\n",
			o: []string{"", "", ""},
		},
		{
			i: "\n\n\r",
			o: []string{"", "", "", ""},
		},
		{
			i: "11\n22\n33",
			o: []string{"11", "22", "33"},
		},
		{
			i: "\n11\n22\n33",
			o: []string{"", "11", "22", "33"},
		},
		{
			i: "\n11\n22\n33\n",
			o: []string{"", "11", "22", "33", ""},
		},
	}

	for _, ts := range tests {
		o := splitLines(ts.i)

		if len(o) != len(ts.o) {
			t.Error("unexpected output")
		}

		for i := 0; i < len(o); i++ {
			if o[i] != ts.o[i] {
				t.Error("element mismatch")
			}
		}
	}
}
