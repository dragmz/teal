package teal

import (
	"testing"
)

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
