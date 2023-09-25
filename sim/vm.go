package sim

import (
	"fmt"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/common/models"
)

type VmValue struct {
	models.AvmValue
}

func (i VmValue) IsNone() bool {
	return i.Type == 0
}

func (i VmValue) String() string {
	switch i.Type {
	case 1:
		return fmt.Sprintf("bytes(%s)", string(i.Bytes))
	case 2:
		return fmt.Sprintf("uint64(%d)", i.Uint)
	default:
		return fmt.Sprintf("unknown(%d)", i.Type)
	}
}

type VmFrame struct {
	Return int
	Name   string
}

type VmBranch struct {
	Id        int
	Budget    int
	Stack     VmStack
	Scratch   VmScratch
	PrevTrace *ProgramExecutionTrace
	Line      int
	Name      string
	Frames    []VmFrame
	Trace     []ProgramExecutionTrace

	i int
	b map[int]bool
}

type VmStack struct {
	Items []VmValue
}

type VmScratch struct {
	Items []VmValue
}

type Vm struct {
	Error     error
	Triggered map[int][]int
	Branch    *VmBranch
	Branches  []*VmBranch
	Pause     bool
}

type VmConfig struct {
	Ac *algod.Client

	Args     [][]byte
	Accounts []string
	Apps     []uint64
	Assets   []uint64
}

func NewVm(r Result) (*Vm, error) {
	var branches []*VmBranch

	for i, e := range r.Executions {
		branch := &VmBranch{
			Id:    i,
			Name:  "", // TODO: name
			Trace: e.Approval,
			b:     map[int]bool{},
			Scratch: VmScratch{
				Items: make([]VmValue, 256),
			},
		}

		branches = append(branches, branch)
	}

	for _, b := range branches {
		if len(b.Trace) > 0 {
			b.Line = b.Trace[0].Line
		}
	}

	var branch *VmBranch

	if len(branches) > 0 {
		branch = branches[0]
	}

	v := &Vm{
		Branch:    branch,
		Branches:  branches,
		Triggered: map[int][]int{},
	}

	for _, b := range branches {
		v.onLine(b)
	}

	return v, nil
}

func (v *Vm) SetBreakpoints(lines []int) map[int]bool {
	v.Branch.b = map[int]bool{}
	for _, l := range lines {
		for _, t := range v.Branch.Trace {
			if t.Line != l {
				continue
			}

			v.Branch.b[l] = true
		}
	}

	return v.Branch.b
}

func (v *Vm) Run() {
	for !v.Pause {
		if !v.Step() {
			break
		}

		if len(v.Triggered) > 0 {
			break
		}
	}

	v.Pause = false
}

func (v *Vm) Switch(id int) {
	for _, b := range v.Branches {
		if b.Id == id {
			v.Branch = b
			break
		}
	}
}

func (v *Vm) Step() bool {
	if v.Branch.i >= len(v.Branch.Trace) {
		return false
	}

	v.Branch.i++
	v.onLine(v.Branch)

	return true
}

func (v *Vm) onLine(b *VmBranch) {
	v.Triggered = map[int][]int{}

	if b.i >= len(b.Trace) {
		return
	}

	t := b.Trace[b.i]

	b.Line = t.Line
	if b.b[b.Line] {
		v.Triggered[b.Id] = []int{b.Line}
	}

	pt := b.PrevTrace

	if pt != nil {
		popCount := int(pt.StackPopCount)

		if popCount > len(b.Stack.Items) {
			popCount = len(b.Stack.Items)
			v.Error = fmt.Errorf("stack underflow")
		}

		b.Stack.Items = b.Stack.Items[:len(b.Stack.Items)-popCount]

		for _, a := range pt.StackAdditions {
			b.Stack.Items = append(b.Stack.Items, VmValue{a})
		}

		for _, s := range pt.ScratchChanges {
			b.Scratch.Items[s.Slot] = VmValue{s.NewValue}
		}
	}

	b.PrevTrace = &t
}
