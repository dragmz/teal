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
		Error:    err,
		Branch:   b,
		Branches: []*VmBranch{b},
	}

	v.onLine()

	return v, nil
}

func (v *Vm) SetBreakpoints(lines []int) map[int]bool {
	clear(v.Branch.b)
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

}

func (v *Vm) Switch(id int) {

}

func (v *Vm) Step() {
	if v.Branch.i >= len(v.Branch.Trace) {
		return
	}

	v.Branch.i++
	v.onLine()
}

func (v *Vm) onLine() {
	if v.Branch.i >= len(v.Branch.Trace) {
		return
	}

	v.Branch.Line = v.Branch.Trace[v.Branch.i].Line
}
