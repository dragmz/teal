package teal

import (
	"testing"
)

type testRange struct {
	sl int
	sc int
	el int
	ec int
}

func (r testRange) StartLine() int {
	return r.sl
}

func (r testRange) StartCharacter() int {
	return r.sc
}

func (r testRange) EndLine() int {
	return r.el
}

func (r testRange) EndCharacter() int {
	return r.ec
}

func TestProcessEmpty(t *testing.T) {
	Process("")
}

func TestRedundantLabelLine(t *testing.T) {
	res := Process("test_label:")

	if len(res.Redundants) != 1 {
		t.Error("len mismatch")
	}

	r := res.Redundants[0]

	if r.Line() != 0 {
		t.Error("line mismatch")
	}

	if r.String() != "Remove label 'test_label'" {
		t.Error("title mismatch")
	}
}

func TestRedundantBCallLine(t *testing.T) {
	res := Process("b a\na:")

	if len(res.Redundants) != 1 {
		t.Error("len mismatch")
	}

	r := res.Redundants[0]

	if r.Line() != 0 {
		t.Error("line mismatch")
	}

	if r.String() != "Remove b call" {
		t.Error("title mismatch")
	}
}

func TestIntArgVals(t *testing.T) {
	res := Process("int ")

	vals := res.ArgValsAt(0, 4)

	m := map[string]bool{}
	for _, v := range vals {
		m[v.Name] = true
	}

	if _, ok := m["DeleteApplication"]; !ok {
		t.Error("missing DeleteApplication")
	}
}

func TestSymOrRefAt(t *testing.T) {
	res := Process(`test_label:
test_label2:
b test_label
`)

	type test struct {
		i Range
		o string
	}

	tests := []test{
		{testRange{}, "test_label"},
		{testRange{1, 1, 1, 1}, "test_label2"},
		{testRange{2, 2, 2, 2}, "test_label"},
	}

	for i, test := range tests {
		name := res.SymOrRefAt(test.i)
		if name != test.o {
			t.Errorf("unexpected name - test: %d, actual: %s, expected: %s", i, name, test.o)
		}
	}
}
