package teal

import (
	"fmt"
)

type LineError interface {
	error
	Line() int
}

type Linter struct {
	l   Listing
	res []LineError
}

type UnusedLabelError struct {
	l    int
	name string
}

func (e *UnusedLabelError) Line() int {
	return e.l
}

func (e UnusedLabelError) Error() string {
	return fmt.Sprintf("unused label: \"%s\"", e.name)
}

type UnreachableCodeError struct {
	l int
}

func (e UnreachableCodeError) Line() int {
	return e.l
}

func (e UnreachableCodeError) Error() string {
	return "unreachable code"
}

type BJustBeforeLabelError struct {
	l int
}

func (e BJustBeforeLabelError) Line() int {
	return e.l
}

func (e BJustBeforeLabelError) Error() string {
	return "unconditional branch just before the target label"
}

type EmptyLoopError struct {
	l int
}

func (e EmptyLoopError) Line() int {
	return e.l
}

func (e EmptyLoopError) Error() string {
	return "empty loop"
}

type MissingLabelError struct {
	l    int
	name string
}

func (e MissingLabelError) Line() int {
	return e.l
}

func (e MissingLabelError) Error() string {
	return fmt.Sprintf("missing label: \"%s\"", e.name)
}

type InfiniteLoopError struct {
	l int
}

func (e InfiniteLoopError) Line() int {
	return e.l
}

func (e InfiniteLoopError) Error() string {
	return "infinite loop"
}

func (l *Linter) getLabelsUsers() map[string][]int {
	used := map[string][]int{}

	for i, o := range l.l {
		switch o2 := o.(type) {
		case usesLabels:
			for _, l := range o2.Labels() {
				used[l.Name] = append(used[l.Name], i)
			}
		}
	}

	return used
}

func (l *Linter) getAllLabels() map[string][]int {
	all := map[string][]int{}

	for i, o := range l.l {
		switch o := o.(type) {
		case *LabelExpr:
			all[o.Name] = append(all[o.Name], i)
		}
	}

	return all
}

func (l *Linter) checkUnusedLabels() {
	used := l.getLabelsUsers()
	for name, lines := range l.getAllLabels() {
		if len(used[name]) == 0 {
			for _, line := range lines {
				l.res = append(l.res, &UnusedLabelError{l: line, name: name})
			}
		}
	}
}

func (l *Linter) checkOpsAfterUnconditionalBranch() {
	used := l.getLabelsUsers()

	for i := 0; i < len(l.l); i++ {
		o := l.l[i]
		unused := false
		switch o.(type) {
		case *BExpr:
			unused = true
		case *ReturnExpr:
			unused = true
		case *ErrExpr:
			unused = true
		}

		if unused {
		loop:
			for i = i + 1; i < len(l.l); i++ {
				o2 := l.l[i]
				switch o2 := o2.(type) {
				case Nop:
				case *LabelExpr:
					if len(used[o2.Name]) > 0 {
						break loop
					}
				default:
					l.res = append(l.res, UnreachableCodeError{i})
				}
			}
		}
	}
}

func (l *Linter) checkBranchJustBeforeLabel() {
	for i, o := range l.l {
		func() {
			if i >= len(l.l)-1 {
				return
			}

			switch o := o.(type) {
			case *BExpr:
				j := i + 1

				func() {
				loop:
					for {
						if j >= len(l.l) {
							break
						}

						n := l.l[j]
						j += 1

						switch n := n.(type) {
						case Nop:
						case *LabelExpr:
							if n.Name == o.Label.Name {
								l.res = append(l.res, BJustBeforeLabelError{l: i})
								return
							}
							break loop
						default:
							break loop
						}
					}
				}()
			default:
			}
		}()
	}
}

func (l *Linter) checkLoops() {
	used := l.getLabelsUsers()
	all := l.getAllLabels()

	for name, users := range used {
		_, ok := all[name]
		if !ok {
			for _, user := range users {
				l.res = append(l.res, MissingLabelError{l: user, name: name})
			}
		}
	}

	for i, o := range l.l {
		if i == 0 {
			continue
		}

		func() {
			if i == 0 {
				return
			}

			switch o := o.(type) {
			case *BExpr:
				j := i - 1

				func() {
					for {
						if j < 0 {
							break
						}

						n := l.l[j]

						switch n := n.(type) {
						case *LabelExpr:
							if n.Name == o.Label.Name {
								if !l.canEscape(j, i) {
									l.res = append(l.res, InfiniteLoopError{l: i})
								}
								return
							}
						}
						j -= 1
					}
				}()
			default:
			}
		}()
	}
}

func (l *Linter) canEscape(from, to int) bool {
	labels := l.getAllLabels()

	for i := from; i <= to; i++ {
		switch op := l.l[i].(type) {
		case usesLabels:
			for _, lbl := range op.Labels() {
				for _, idx := range labels[lbl.Name] {
					if idx < from || idx > to {
						// TODO: check if the target label block is escapable
						return true
					}
				}
			}
		case Terminator:
			return true
		}
	}

	return false
}

func (l *Linter) Lint() {
	l.checkUnusedLabels()
	l.checkOpsAfterUnconditionalBranch()
	l.checkBranchJustBeforeLabel()
	l.checkLoops()
}
