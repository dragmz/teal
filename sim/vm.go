package sim

import (
	"fmt"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/common/models"
	"github.com/dragmz/teal"
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

func NewVm(src string, config RunConfig) (*Vm, error) {
	pr := teal.Process(src)

	clear := "int 1"
	if pr.Version > 1 {
		clear = fmt.Sprintf("#pragma version %d\r\n", pr.Version) + clear
	}

	r, err := Run([]byte(src), []byte(clear), config)

	createBranch := &VmBranch{
		Id:    0,
		Name:  "create",
		Trace: r.Create.Approval,
		b:     map[int]bool{},
		Scratch: VmScratch{
			Items: make([]VmValue, 256),
		},
	}

	callBranch := &VmBranch{
		Id:    1,
		Name:  "call",
		Trace: r.Call.Approval,
		b:     map[int]bool{},
		Scratch: VmScratch{
			Items: make([]VmValue, 256),
		},
	}

	var branches []*VmBranch

	if config.Create.Debug {
		branches = append(branches, createBranch)
	}

	if config.Call.Debug {
		branches = append(branches, callBranch)
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
		Error:     err,
		Branch:    branch,
		Branches:  branches,
		Triggered: map[int][]int{},
	}

	v.onLine()

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
	v.onLine()

	return true
}

func (v *Vm) onLine() {
	v.Triggered = map[int][]int{}

	if v.Branch.i >= len(v.Branch.Trace) {
		return
	}

	t := v.Branch.Trace[v.Branch.i]

	v.Branch.Line = t.Line
	if v.Branch.b[v.Branch.Line] {
		v.Triggered[v.Branch.Id] = []int{v.Branch.Line}
	}

	pt := v.Branch.PrevTrace

	if pt != nil {
		popCount := int(pt.StackPopCount)

		if popCount > len(v.Branch.Stack.Items) {
			popCount = len(v.Branch.Stack.Items)
			v.Error = fmt.Errorf("stack underflow")
		}

		v.Branch.Stack.Items = v.Branch.Stack.Items[:len(v.Branch.Stack.Items)-popCount]

		for _, a := range pt.StackAdditions {
			v.Branch.Stack.Items = append(v.Branch.Stack.Items, VmValue{a})
		}

		for _, s := range pt.ScratchChanges {
			v.Branch.Scratch.Items[s.Slot] = VmValue{s.NewValue}
		}
	}

	v.Branch.PrevTrace = &t
}
