package teal

import (
	"fmt"
	"strings"
)

type VM struct {
	l Listing
}

type vmFrame struct {
	a uint8
	r uint8
	p uint8
}

type vmValue struct {
	t vmDataType

	ref string
}

func (v vmValue) String() string {
	return v.t.String()
}

type vmValueType int

type vmDataType int

const (
	vmTypeNone = iota
	vmTypeAny
	vmTypeUint64
	vmTypeBytes
)

func (t vmDataType) String() string {
	switch t {
	case vmTypeAny:
		return "any"
	case vmTypeUint64:
		return "uint64"
	case vmTypeBytes:
		return "bytes"
	default:
		return "(none)"
	}
}

type vmStack struct {
	items []vmValue
}

func (b *vmBranch) push(t vmValue) {
	b.s.items = append(b.s.items, t)
}

func (b *vmBranch) prepare(a uint8, r uint8) {
	b.fs = append(b.fs, vmFrame{a: a, r: r, p: uint8(len(b.s.items))})
}

func (b *vmBranch) replace(n uint8, v vmValue) {
	b.s.items[n] = v
}

func (b *vmBranch) at(n uint8) vmValue {
	return b.s.items[uint8(len(b.s.items))-1-n]
}

func (b *vmBranch) pop(t vmDataType) vmValue {
	if len(b.s.items) == 0 {
		panic("empty stack")
	}

	v := b.s.items[len(b.s.items)-1]
	switch t {
	case vmTypeAny:
	default:
		switch v.t {
		case vmTypeAny:
		default:
			if v.t != t {
				panic("unexpected data type on stack")
			}
		}
	}
	b.s.items = b.s.items[:len(b.s.items)-1]
	return v
}

func (s *vmStack) clone() *vmStack {
	return &vmStack{
		items: append([]vmValue{}, s.items...),
	}
}

type vmBranch struct {
	id int

	v *vm

	ln int
	cs []int

	s *vmStack

	fs []vmFrame

	budget int
}

func (b *vmBranch) retsub() {
	b.ln = b.cs[len(b.cs)-1]
	b.cs = b.cs[:len(b.cs)-1]
}

func (b *vmBranch) fork(target string) {
	nb := &vmBranch{
		id:     b.v.id,
		v:      b.v,
		ln:     b.v.find(target),
		s:      b.s.clone(),
		budget: b.budget,
	}

	b.v.id++

	b.v.br = append(b.v.br, nb)
	b.ln = -1
}

func (b *vmBranch) jump(target string) {
	b.ln = b.v.find(target)
}

func (b *vmBranch) call(target string) {
	b.cs = append(b.cs, b.ln+1)
	b.ln = b.v.find(target)
}

func (b *vmBranch) exit() {
	b.ln = -1
}

type vm struct {
	id int

	l    Listing
	syms map[string]int

	br []*vmBranch
}

func (v *vm) find(target string) int {
	ln, ok := v.syms[target]
	if !ok {
		return -1
	}

	return ln
}

func Interpret(l Listing) error {
	syms := map[string]int{}

	for i, op := range l {
		switch op := op.(type) {
		case *LabelExpr:
			syms[op.Name] = i
		}
	}

	v := &vm{
		l:    l,
		syms: syms,
	}

	b := &vmBranch{
		id:     v.id,
		v:      v,
		s:      &vmStack{},
		budget: 700,
	}

	v.id++

	v.br = append(v.br, b)

	for len(v.br) > 0 {
		b := v.br[0]

		var op Op
		for b.ln < len(v.l) {
			op = v.l[b.ln]

			_, nop := op.(Nop)
			if !nop {
				break
			}

			b.ln++
		}

		if b.ln < len(v.l) {
			if b.budget > 0 {
				budget := b.budget
				ln := b.ln

				b.budget--

				switch op := op.(type) {
				case vmOp:
					op.Execute(b)
				default:
					b.ln++
				}

				ss := []string{}
				for i := len(b.s.items) - 1; i >= 0; i-- {
					ss = append(ss, b.s.items[i].String())
				}

				fmt.Printf("[%d: %d, %d] %s | [%s]\n", b.id, ln, budget, op, strings.Join(ss, ", "))
			} else {
				fmt.Printf("[%d: %d, %d] out of budget\n", b.id, b.ln, b.budget)
				b.ln = -1
			}
		}

		if b.ln == -1 || b.ln == len(v.l) {
			v.br = v.br[1:]
		}

	}

	return nil
}
