package teal

import (
	"fmt"
	"strings"
)

type vmFrame struct {
	a uint8
	r uint8
	p uint8
}

type vmSource interface {
	String() string
}

type vmValue struct {
	t vmDataType

	src vmSource
}

type vmConst struct {
	v string
}

type NamedExpr interface {
	Name() string
}

type vmOpSource struct {
	e    NamedExpr
	args []vmSource
	res  string
}

func (s vmOpSource) String() string {
	res := s.e.Name()
	if len(s.args) > 0 {
		var argss []string
		for _, arg := range s.args {
			argss = append(argss, arg.String())
		}

		res += fmt.Sprintf("(%s)", strings.Join(argss, ", "))
	}
	return res
}

func (c vmConst) String() string {
	return c.v
}

func (v vmValue) String() string {
	res := v.t.String()
	if v.src != nil {
		res += ": " + v.src.String()
	}
	return res
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

type vmOp interface {
	Execute(b *VmBranch) error
}

func (t StackType) Vm() vmDataType {
	switch t {
	case StackAny:
		return vmTypeAny
	case StackBytes:
		return vmTypeBytes
	case StackUint64:
		return vmTypeUint64
	case StackNone:
		return vmTypeNone
	default:
		panic("unknown stack type")
	}
}

type vmStack struct {
	Items []vmValue
}

func (b *VmBranch) push(t vmValue) {
	b.Stack.Items = append(b.Stack.Items, t)
}

func (b *VmBranch) prepare(a uint8, r uint8) {
	f := b.fs[len(b.fs)-1]
	f.a = a
	f.r = r
	b.fs[len(b.fs)-1] = f
}

func (b *VmBranch) replace(n uint8, v vmValue) {
	b.Stack.Items[n] = v
}

func (b *VmBranch) pop(t vmDataType) vmValue {
	if len(b.Stack.Items) == 0 {
		panic("empty stack")
	}

	v := b.Stack.Items[len(b.Stack.Items)-1]
	switch t {
	case vmTypeAny:
	default:
		switch v.t {
		case vmTypeAny:
		default:
			if v.t != t {
				panic(fmt.Sprintf("unexpected data type on stack - expected: %s, got: %s", t, v.t))
			}
		}
	}
	b.Stack.Items = b.Stack.Items[:len(b.Stack.Items)-1]
	return v
}

func (s *vmStack) clone() *vmStack {
	return &vmStack{
		Items: append([]vmValue{}, s.Items...),
	}
}

type VmBranch struct {
	Id int

	vm *Vm

	Line int
	cs   []int

	Stack *vmStack

	fs []vmFrame

	Budget int
}

func (b *VmBranch) fork(target string) {
	nb := &VmBranch{
		Id:     b.vm.Id,
		vm:     b.vm,
		Line:   b.vm.find(target),
		Stack:  b.Stack.clone(),
		Budget: b.Budget,
	}

	b.vm.Id++

	b.vm.Branches = append(b.vm.Branches, nb)
}

func (b *VmBranch) jump(target string) {
	b.Line = b.vm.find(target)
}

func (b *VmBranch) call(target string) {
	b.cs = append(b.cs, b.Line+1)
	b.Line = b.vm.find(target)
	b.fs = append(b.fs, vmFrame{a: 0, r: 0, p: uint8(len(b.Stack.Items))})
}

func (b *VmBranch) exit() {
	b.Line = -1
}

type Vm struct {
	Id int

	Line Listing
	syms map[string]int

	Branches []*VmBranch
	Branch   *VmBranch
	Current  int

	Trace string
}

func (v *Vm) find(target string) int {
	ln, ok := v.syms[target]
	if !ok {
		return -1
	}

	return ln
}

func Interpret(l Listing) *Vm {
	syms := map[string]int{}

	for i, op := range l {
		switch op := op.(type) {
		case *LabelExpr:
			syms[op.Name] = i
		}
	}

	v := &Vm{
		Line: l,
		syms: syms,
	}

	b := &VmBranch{
		Id:     v.Id,
		vm:     v,
		Stack:  &vmStack{},
		Budget: 700,
	}

	v.Id++

	v.Branches = append(v.Branches, b)

	v.skipNops()

	return v
}

func (v *Vm) skipNops() {
	var op Op

	for len(v.Branches) > v.Current {
		b := v.Branches[v.Current]

		if b.Line == -1 || b.Line == len(v.Line) {
			v.Current++
			continue
		}

		for b.Line < len(v.Line) {
			op = v.Line[b.Line]

			_, nop := op.(Nop)
			if !nop {
				v.Branch = b
				return
			}

			b.Line++
		}
	}

	v.Branch = nil
}

func (v *Vm) Step() {
	v.skipNops()

	if v.Branch == nil {
		return
	}

	b := v.Branch

	if b.Line < len(v.Line) {
		op := v.Line[b.Line]
		if b.Budget > 0 {
			b.Budget--

			switch op := op.(type) {
			case vmOp:
				op.Execute(b)
			default:
				b.Line++
			}

			v.skipNops()
		} else {
			b.Line = -1
		}
	}
}

func (v *Vm) Run() {
	for len(v.Branches) > 0 {
		v.Step()
	}
}
