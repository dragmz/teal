package sim

import (
	"fmt"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/dragmz/teal"
)

type VmStackItem struct {
	None bool
}

func (i VmStackItem) String() string {
	return "TODO"
}

type VmFrame struct {
	Return int
	Name   string
}

type VmBranch struct {
	Id     int
	Budget int
	Stack  VmStack
	Line   int
	Name   string
	Frames []VmFrame
	Trace  []ProgramExecutionTrace

	i int
	b map[int]bool
}

type VmStack struct {
	Items []VmStackItem
}

type VmScratch struct {
	Items []VmStackItem
}

type Vm struct {
	Error     error
	Triggered map[int][]int
	Branch    *VmBranch
	Branches  []*VmBranch
	Scratch   VmScratch
	Pause     bool
}

func NewVm(src string) (*Vm, error) {
	ac, err := algod.MakeClient("https://testnet-api.algonode.cloud", "")
	if err != nil {
		return nil, err
	}

	pr := teal.Process(src)

	clear := "int 1"
	if pr.Version > 1 {
		clear = fmt.Sprintf("#pragma version %d\r\n", pr.Version) + clear
	}

	r, err := Run(ac, "F77YBQEP4EAJYCQPS4GYEW2WWJXU6DQ4OJHRYSV74UXHOTRWXYRN7HNP3U", []byte(src), []byte(clear))

	b := &VmBranch{
		Trace: r.Call.Approval,
		b:     map[int]bool{},
	}

	v := &Vm{
		Error:     err,
		Branch:    b,
		Branches:  []*VmBranch{b},
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

	v.Branch.Line = v.Branch.Trace[v.Branch.i].Line
	if v.Branch.b[v.Branch.Line] {
		v.Triggered[v.Branch.Id] = []int{v.Branch.Line}
	}
}
