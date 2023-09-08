package teal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
	res := Process("")

	assert.Equal(t, ModeApp, res.Mode)
	assert.Equal(t, uint64(1), res.Version)

	assert.Equal(t, 0, len(res.Diagnostics))
	assert.Equal(t, 0, len(res.Keywords))
	assert.Equal(t, 0, len(res.Lines))
	assert.Equal(t, 0, len(res.Listing))
	assert.Equal(t, 0, len(res.Macros))
	assert.Equal(t, 0, len(res.MissRefs))
	assert.Equal(t, 0, len(res.Numbers))
	assert.Equal(t, 0, len(res.Ops))
	assert.Equal(t, 0, len(res.Redundants))
	assert.Equal(t, 0, len(res.RefCounts))
	assert.Equal(t, 0, len(res.Strings))
	assert.Equal(t, 0, len(res.SymbolRefs))
	assert.Equal(t, 0, len(res.Symbols))
	assert.Equal(t, 0, len(res.Tokens))
	assert.Equal(t, 0, len(res.Versions))
}

func TestRedundantLabelLine(t *testing.T) {
	res := Process("test_label:")

	assert.Len(t, res.Redundants, 1)

	r := res.Redundants[0]

	assert.Equal(t, 0, r.Line())
	assert.Equal(t, "Remove label 'test_label'", r.String())
}

func TestRedundantBCallLine(t *testing.T) {
	res := Process("b a\na:")

	if len(res.Redundants) != 1 {
		t.Error("len mismatch")
	}

	r := res.Redundants[0]

	assert.Equal(t, 0, r.Line())
	assert.Equal(t, "Remove b call", r.String())
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
		assert.Equal(t, test.o, name, fmt.Sprintf("test #%d", i))
	}
}

func TestOverlaps(t *testing.T) {
	type test struct {
		a Range
		b Range

		o bool
	}

	tests := []test{
		{testRange{1, 1, 1, 2}, testRange{0, 3, 1, 3}, true},

		{testRange{1, 1, 1, 2}, testRange{0, 0, 1, 0}, false},
		{testRange{1, 1, 1, 2}, testRange{0, 2, 1, 2}, true},

		{testRange{1, 1, 1, 2}, testRange{1, 0, 1, 0}, false},
		{testRange{0, 5, 0, 11}, testRange{0, 0, 1, 0}, true},

		{testRange{}, testRange{}, true},
		{testRange{}, testRange{0, 1, 0, 1}, false},
		{testRange{0, 1, 0, 1}, testRange{0, 1, 0, 1}, true},
		{testRange{1, 1, 1, 1}, testRange{1, 1, 1, 1}, true},
		{testRange{1, 1, 1, 2}, testRange{1, 1, 1, 1}, true},
		{testRange{1, 1, 1, 2}, testRange{1, 2, 1, 2}, true},
		{testRange{1, 1, 1, 2}, testRange{1, 3, 1, 3}, false},
		{testRange{1, 1, 1, 2}, testRange{0, 1, 1, 1}, true},
		{testRange{1, 1, 1, 2}, testRange{0, 1, 0, 1}, false},
		{testRange{1, 1, 1, 2}, testRange{0, 2, 0, 2}, false},
		{testRange{1, 1, 1, 2}, testRange{0, 0, 0, 0}, false},
		{testRange{1, 1, 1, 2}, testRange{0, 3, 0, 3}, false},
		{testRange{1, 1, 1, 2}, testRange{2, 1, 2, 1}, false},
		{testRange{1, 1, 1, 2}, testRange{2, 2, 2, 2}, false},
		{testRange{1, 1, 1, 2}, testRange{2, 0, 2, 0}, false},
		{testRange{1, 1, 1, 2}, testRange{2, 3, 2, 3}, false},
	}

	for i, test := range tests {
		o := Overlaps(test.a, test.b)
		assert.Equal(t, test.o, o, fmt.Sprintf("test #%d", i))
	}
}

func TestInlayHints(t *testing.T) {
	tests := []string{
		"byte 0x3031",
		"byte 0x3031\n",
	}

	for i, ts := range tests {
		name := fmt.Sprintf("test #%d", i)

		res := Process(ts)
		ihs := res.InlayHints(testRange{0, 0, 1, 0})

		if !assert.Equal(t, 1, len(ihs.Decoded), name) {
			return
		}
		assert.Equal(t, "01", ihs.Decoded[0].Value, name)
	}
}

func TestRefCounts(t *testing.T) {
	res := Process("b a\nb a\nb a\na:")
	assert.Equal(t, 3, res.RefCounts["a"])
}

func TestVersion(t *testing.T) {
	res := Process("#pragma version 8")
	assert.Equal(t, uint64(8), res.Version)
}

func TestRequiredVersion(t *testing.T) {
	res := Process("box_create")
	assert.Len(t, res.Versions, 1)

	v := res.Versions[0]
	assert.Equal(t, uint64(8), v.Version)
}

func TestInvalidByteInt(t *testing.T) {
	res := Process(`#pragma version 8
	byte \"test\"int 123
	int 1
	int 2`)

	assert.Len(t, res.Lines, 4)
	assert.Len(t, res.Listing, 4)
}

func TestSemicolon(t *testing.T) {
	res := Process("int 1; int 2")
	assert.Len(t, res.Lines, 1)

	assert.Len(t, res.Lines[0].Subs, 2)
	assert.Len(t, res.Lines[0].Subs[0].Tokens, 2)
	assert.Len(t, res.Lines[0].Subs[1].Tokens, 2)

	assert.Len(t, res.Lines[0].Tokens, 5)
}

func TestSemicolonEmptySubs(t *testing.T) {
	res := Process("int 1;")
	assert.Len(t, res.Lines[0].Subs, 2)
}

func TestMultiSemicolon(t *testing.T) {
	res := Process(";;;")
	assert.Len(t, res.Lines, 1)

	assert.Len(t, res.Lines[0].Subs, 4)
	assert.Len(t, res.Lines[0].Subs[0].Tokens, 0)
	assert.Len(t, res.Lines[0].Subs[1].Tokens, 0)
	assert.Len(t, res.Lines[0].Subs[2].Tokens, 0)
	assert.Len(t, res.Lines[0].Subs[3].Tokens, 0)

	assert.Len(t, res.Lines[0].Tokens, 3)
}

func TestGithubIssueVsCodeTeal3Regression(t *testing.T) {
	Process(`int 1 /
	b a`)
}

func TestBranchToSameLine(t *testing.T) {
	Process("a:;b a")
}

func TestInfiniteLoopLinting(t *testing.T) {
	res := Process(`#pragma version 8
	l1:
	b l1
	l2:
	b l2`)

	assert.Len(t, res.Diagnostics, 2)
	assert.Equal(t, res.Diagnostics[0].Line(), 2)
	assert.Equal(t, res.Diagnostics[1].Line(), 4)
}

func TestDefine(t *testing.T) {
	res := Process(`#pragma version 8;
	#define test123 b a`)
	assert.Len(t, res.Diagnostics, 0)
}

func TestLogicSigMode(t *testing.T) {
	res := Process(`//#pragma mode logicsig`)

	assert.Equal(t, ModeSig, res.Mode)
}

func TestAppMode(t *testing.T) {
	req := Process(``)
	assert.Equal(t, ModeApp, req.Mode)
}
