package teal

import "testing"

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
