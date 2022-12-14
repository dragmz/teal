package teal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	MainName = "(main)"
	ExitLine = -1
	ExitName = "(exited)"
)

type VmFrame struct {
	Return     int
	NumArgs    uint8
	NumReturns uint8
	p          uint8
	Name       string
}

type vmSource interface {
	String() string
}

type VmValue struct {
	T VmDataType

	src vmSource
}

type lenValue interface {
	Lengths() []int
}

func (v VmValue) Lengths() []int {
	switch src := v.src.(type) {
	case lenValue:
		return src.Lengths()
	default:
		return []int{}
	}
}

type vmUint64Const struct {
	v uint64
}

func (c vmUint64Const) Lengths() []int {
	return []int{0}
}

func (c vmUint64Const) String() string {
	return strconv.FormatUint(c.v, 10)
}

type vmSignatureValue struct {
	v string
}

func (v vmSignatureValue) String() string {
	return v.v
}

func (v vmSignatureValue) Lengths() []int {
	return []int{4}
}

type vmByteConst struct {
	v []byte
}

func (c vmByteConst) Lengths() []int {
	return []int{len(c.v)}
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

func (c vmByteConst) String() string {
	return Bytes{Value: c.v}.String()
}

func (v VmValue) String() string {
	res := v.T.String()
	if v.src != nil {
		res += ": " + v.src.String()
	}
	return res
}

type vmValueType int

type VmDataType int

const (
	VmTypeNone = iota
	VmTypeAny
	VmTypeUint64
	VmTypeBytes
)

func (t VmDataType) String() string {
	switch t {
	case VmTypeAny:
		return "any"
	case VmTypeUint64:
		return "uint64"
	case VmTypeBytes:
		return "bytes"
	default:
		return "(none)"
	}
}

type vmOp interface {
	Execute(b *VmBranch) error
}

type costlyOp interface {
	Cost(b *VmBranch) []int
}

func (t StackType) Vm() VmDataType {
	switch t {
	case StackAny:
		return VmTypeAny
	case StackBytes:
		return VmTypeBytes
	case StackUint64:
		return VmTypeUint64
	case StackNone:
		return VmTypeNone
	default:
		panic("unknown stack type")
	}
}

type vmStack struct {
	Items []VmValue
}

func (b *VmBranch) skipNops() bool {
	if b.Line == ExitLine {
		return false
	}

	for b.Line < len(b.vm.Process.Listing) {
		op := b.vm.Process.Listing[b.Line]

		lbl, ok := op.(*LabelExpr)
		if ok {
			b.Name = lbl.Name
		}

		_, isnop := op.(Nop)
		if !isnop {
			return true
		}

		b.Line++
		b.vm.updateBreakpoints(b)
	}

	return false
}

func (b *VmBranch) push(t VmValue) {
	b.Stack.Items = append(b.Stack.Items, t)
}

func (b *VmBranch) prepare(a uint8, r uint8) {
	f := b.Frames[len(b.Frames)-1]
	f.NumArgs = a
	f.NumReturns = r
	b.Frames[len(b.Frames)-1] = f
}

func (b *VmBranch) replace(n uint8, v VmValue) {
	b.Stack.Items[n] = v
}

func (b *VmBranch) store(index VmValue, v VmValue) {
	switch src := index.src.(type) {
	case vmUint64Const:
		b.vm.Scratch.Items[src.v] = v
	}
}

func (b *VmBranch) peek(index int) VmValue {
	return b.Stack.Items[len(b.Stack.Items)-1-index]
}

func (b *VmBranch) pop(t VmDataType) VmValue {
	if len(b.Stack.Items) == 0 {
		panic(errors.Errorf("empty stack - expected: %s", t))
	}

	v := b.Stack.Items[len(b.Stack.Items)-1]
	switch t {
	case VmTypeAny:
	default:
		switch v.T {
		case VmTypeAny:
		default:
			if v.T != t {
				panic(fmt.Sprintf("unexpected data type on stack - expected: %s, got: %s", t, v.T))
			}
		}
	}
	b.Stack.Items = b.Stack.Items[:len(b.Stack.Items)-1]
	return v
}

func (s *vmStack) clone() *vmStack {
	return &vmStack{
		Items: append([]VmValue{}, s.Items...),
	}
}

type VmBranch struct {
	Id int

	vm *Vm

	Line int

	Stack *vmStack

	Frames []VmFrame

	Budget int

	Name  string
	Trace []Op
}

func (b *VmBranch) fork(target string) {
	nb := &VmBranch{
		Id:     b.vm.Id,
		vm:     b.vm,
		Line:   b.vm.find(target),
		Stack:  b.Stack.clone(),
		Budget: b.Budget,
		Name:   target,
		Trace:  append([]Op{}, b.Trace...),
	}

	b.vm.Id++

	nb.skipNops()
	b.vm.Branches = append(b.vm.Branches, nb)
}

func (b *VmBranch) jump(target string) {
	b.Line = b.vm.find(target)
}

func (b *VmBranch) call(target string) {
	b.Frames = append(b.Frames, VmFrame{Return: b.Line, NumArgs: 0, NumReturns: 0, p: uint8(len(b.Stack.Items)), Name: b.Name})
	b.Line = b.vm.find(target)
}

func (b *VmBranch) exit() {
	b.Line = ExitLine
	b.Name = ExitName
}

type VmScratch struct {
	Items [256]VmValue
}

type VmBreakpoint struct {
	Line int
}

type Vm struct {
	Id int

	Process *ProcessResult
	syms    map[string]int

	Scratch VmScratch

	Branches []*VmBranch
	Branch   *VmBranch
	Current  int

	Breakpoints []VmBreakpoint
	Triggered   map[int][]int

	Trace string

	Error any
}

func (v *Vm) find(target string) int {
	ln, ok := v.syms[target]
	if !ok {
		return -1
	}

	return ln
}

func (v *Vm) updateBreakpoints(br *VmBranch) {
	for _, bp := range v.Breakpoints {
		if br.Line == bp.Line {
			v.Triggered[br.Id] = append(v.Triggered[br.Id], br.Line)
		}
	}
}

func NewVm(res *ProcessResult) *Vm {
	syms := map[string]int{}

	for i, op := range res.Listing {
		switch op := op.(type) {
		case *LabelExpr:
			syms[op.Name] = i
		}
	}

	v := &Vm{
		Process:   res,
		Triggered: map[int][]int{},
		syms:      syms,
	}

	b := &VmBranch{
		Id:     v.Id,
		vm:     v,
		Stack:  &vmStack{},
		Budget: 700,
		Name:   MainName,
	}

	v.Id++

	v.Branches = append(v.Branches, b)

	v.skipNops()

	return v
}

func (v *Vm) skipNops() {
	begin := v.Current

	for len(v.Branches) > v.Current {
		b := v.Branches[v.Current]

		if b.Line == ExitLine {
			v.Current++
			if v.Current == len(v.Branches) {
				v.Current = 0
			}
			if v.Current == begin {
				break
			}
			continue
		}

		if ok := b.skipNops(); ok {
			v.Branch = b
			return
		}

		b.exit()
	}

	v.Branch = nil
}

func (v *Vm) Switch(id int) {
	for i, b := range v.Branches {
		if b.Id == id {
			v.Current = i
			v.Branch = b
			break
		}
	}
}

func (v *Vm) SetBreakpoints(lns []int) map[int]bool {
	verified := map[int]bool{}

	var res []VmBreakpoint

	for _, ln := range lns {
		if ln >= len(v.Process.Listing) {
			continue
		}

		if _, isnop := v.Process.Listing[ln].(Nop); !isnop {
			res = append(res, VmBreakpoint{
				Line: ln,
			})
			verified[ln] = true
		}
	}

	v.Breakpoints = res
	return verified
}

func (v *Vm) Step() {
	defer func() {
		switch e := recover().(type) {
		case nil:
		default:
			v.Error = e
		}
	}()

	v.Error = nil
	v.Triggered = map[int][]int{}

	v.skipNops()

	if v.Branch == nil {
		return
	}

	if b := v.Branch; b != nil {
		op := v.Process.Listing[b.Line]

		var costs []int

		switch op := op.(type) {
		case costlyOp:
			costs = op.Cost(b)
		}

		if len(costs) == 0 {
			costs = []int{1}
		}

		cb := b

		for _, cost := range costs {
			if cb == nil {
				cb := &VmBranch{
					Id:     b.vm.Id,
					vm:     b.vm,
					Line:   b.Line,
					Stack:  b.Stack.clone(),
					Budget: b.Budget,
					Name:   b.Name,
					Trace:  append([]Op{}, b.Trace...),
				}

				b.vm.Id++
				b.vm.Branches = append(b.vm.Branches, cb)
			}

			if cb.Budget >= cost {
				cb.Budget -= cost
				cb.Trace = append(cb.Trace, op)

				switch op := op.(type) {
				case vmOp:
					op.Execute(cb)
				default:
					cb.Line++
				}

				cb.skipNops()
				v.skipNops()
				v.updateBreakpoints(cb)
			} else {
				cb.exit()
			}
		}
	}
}

func (v *Vm) Run() {
	for v.Branch != nil && v.Error == nil {
		v.Step()

		if len(v.Triggered) > 0 {
			return
		}
	}
}
