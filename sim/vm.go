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
		return fmt.Sprintf("uint(%d)", i.Uint)
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
	Scratch   VmScratch
	Pause     bool
}

func NewVm(src string, args [][]byte) (*Vm, error) {
	ac, err := algod.MakeClient("https://testnet-api.algonode.cloud", "")
	if err != nil {
		return nil, err
	}

	pr := teal.Process(src)

	clear := "int 1"
	if pr.Version > 1 {
		clear = fmt.Sprintf("#pragma version %d\r\n", pr.Version) + clear
	}

	r, err := Run(ac, "F77YBQEP4EAJYCQPS4GYEW2WWJXU6DQ4OJHRYSV74UXHOTRWXYRN7HNP3U", []byte(src), []byte(clear), args)

	b := &VmBranch{
		Trace: r.Call.Approval,
		b:     map[int]bool{},
	}

	v := &Vm{
		Error:     err,
		Branch:    b,
		Branches:  []*VmBranch{b},
		Triggered: map[int][]int{},
		Scratch: VmScratch{
			Items: make([]VmValue, 256),
		},
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
			v.Scratch.Items[s.Slot] = VmValue{s.NewValue}
		}
	}

	v.Branch.PrevTrace = &t
}
