package teal

import (
	"fmt"
)

type position struct {
	l int
	s int
}

func (p position) Before(other position) bool {
	return p.l < other.l || (p.l == other.l && p.s < other.s)
}

func (p position) After(other position) bool {
	return p.l > other.l || (p.l == other.l && p.s > other.s)
}

type RedundantLine interface {
	Line() int
	Subline() int
	String() string
}

type RedundantLabelLine struct {
	p    position
	name string
}

func (l RedundantLabelLine) Line() int {
	return l.p.l
}

func (l RedundantLabelLine) Subline() int {
	return l.p.s
}

func (l RedundantLabelLine) String() string {
	return fmt.Sprintf("Remove label '%s'", l.name)
}

type RedundantBLine struct {
	p position
}

func (l RedundantBLine) Line() int {
	return l.p.l
}

func (l RedundantBLine) Subline() int {
	return l.p.s
}

func (l RedundantBLine) String() string {
	return "Remove b call"
}

type LineError interface {
	error
	Line() int
	Subline() int
	Severity() DiagnosticSeverity
	Rule() string
}

type Linter struct {
	l [][]Op

	errs []LineError
	reds []RedundantLine
}

func (l *Linter) forEachForward(line, sub int, cb func(line int, sub int) bool) {
	first := true
	for i := line; i < len(l.l); i++ {
		if first {
			first = false
		} else {
			sub = 0
		}
		for j := sub; j < len(l.l[i]); j++ {
			if !cb(i, j) {
				return
			}
		}
	}
}

func (l *Linter) forEachBackward(line, sub int, cb func(int, int) bool) {
	first := true
	for i := line; i >= 0; i-- {
		if first {
			first = false
		} else {
			sub = len(l.l[i]) - 1
		}
		for j := sub; j >= 0; j-- {
			if !cb(i, j) {
				return
			}
		}
	}
}

type DuplicateLabelError struct {
	p    position
	name string
	rule string
}

func (e DuplicateLabelError) Line() int {
	return e.p.l
}

func (e DuplicateLabelError) Subline() int {
	return e.p.s
}

func (e DuplicateLabelError) Error() string {
	return fmt.Sprintf("duplicate label: \"%s\"", e.name)
}

func (e DuplicateLabelError) Severity() DiagnosticSeverity {
	return DiagErr
}

func (e DuplicateLabelError) Rule() string {
	return e.rule
}

type UnusedLabelError struct {
	p    position
	name string
	rule string
}

func (e *UnusedLabelError) Line() int {
	return e.p.l
}

func (e *UnusedLabelError) Subline() int {
	return e.p.s
}

func (e UnusedLabelError) Error() string {
	return fmt.Sprintf("unused label: \"%s\"", e.name)
}

func (e UnusedLabelError) Severity() DiagnosticSeverity {
	return DiagWarn
}

func (e UnusedLabelError) Rule() string {
	return e.rule
}

type UnreachableCodeError struct {
	p    position
	rule string
}

func (e UnreachableCodeError) Line() int {
	return e.p.l
}

func (e UnreachableCodeError) Subline() int {
	return e.p.s
}

func (e UnreachableCodeError) Error() string {
	return "unreachable code"
}

func (e UnreachableCodeError) Severity() DiagnosticSeverity {
	return DiagWarn
}

func (e UnreachableCodeError) Rule() string {
	return e.rule
}

type BJustBeforeLabelError struct {
	p    position
	rule string
}

func (e BJustBeforeLabelError) Line() int {
	return e.p.l
}

func (e BJustBeforeLabelError) Subline() int {
	return e.p.s
}

func (e BJustBeforeLabelError) Error() string {
	return "unconditional branch just before the target label"
}

func (e BJustBeforeLabelError) Severity() DiagnosticSeverity {
	return DiagWarn
}

func (e BJustBeforeLabelError) Rule() string {
	return e.rule
}

type EmptyLoopError struct {
	l    int
	rule string
}

func (e EmptyLoopError) Line() int {
	return e.l
}

func (e EmptyLoopError) Error() string {
	return "empty loop"
}

func (e EmptyLoopError) Severity() DiagnosticSeverity {
	return DiagWarn
}

func (e EmptyLoopError) Rule() string {
	return e.rule
}

type MissingLabelError struct {
	p    position
	name string
	rule string
}

func (e MissingLabelError) Line() int {
	return e.p.l
}

func (e MissingLabelError) Subline() int {
	return e.p.s
}

func (e MissingLabelError) Error() string {
	return fmt.Sprintf("missing label: \"%s\"", e.name)
}

func (e MissingLabelError) Severity() DiagnosticSeverity {
	return DiagErr
}

func (e MissingLabelError) Rule() string {
	return e.rule
}

type InfiniteLoopError struct {
	p    position
	rule string
}

func (e InfiniteLoopError) Line() int {
	return e.p.l
}

func (e InfiniteLoopError) Subline() int {
	return e.p.s
}

func (e InfiniteLoopError) Error() string {
	return "infinite loop"
}

func (e InfiniteLoopError) Severity() DiagnosticSeverity {
	return DiagErr
}

func (e InfiniteLoopError) Rule() string {
	return e.rule
}

type PragmaVersionAfterInstrError struct {
	p    position
	rule string
}

func (e PragmaVersionAfterInstrError) Line() int {
	return e.p.l
}

func (e PragmaVersionAfterInstrError) Subline() int {
	return e.p.s
}

func (e PragmaVersionAfterInstrError) Error() string {
	return "#pragma version is only allowed before instructions"
}

func (e PragmaVersionAfterInstrError) Severity() DiagnosticSeverity {
	return DiagErr
}

func (e PragmaVersionAfterInstrError) Rule() string {
	return e.rule
}

func (l *Linter) getLabelsUsers() map[string][]position {
	used := map[string][]position{}

	for i, lo := range l.l {
		for j, o := range lo {
			switch o2 := o.(type) {
			case usesLabels:
				for _, l := range o2.Labels() {
					used[l.Name] = append(used[l.Name], position{l: i, s: j})
				}
			}
		}
	}

	return used
}

func (l *Linter) getAllLabels() map[string][]position {
	all := map[string][]position{}

	for i, lo := range l.l {
		for j, o := range lo {
			switch o := o.(type) {
			case *LabelExpr:
				all[o.Name] = append(all[o.Name], position{l: i, s: j})
			}
		}
	}

	return all
}

func (l *Linter) canEscape(from, to position) bool {
	labels := l.getAllLabels()

	result := false
	l.forEachForward(from.l, from.s, func(i, j int) bool {
		if to.l == i && to.s == j {
			return false
		}

		op := l.l[i][j]

		switch op := op.(type) {
		case usesLabels:
			for _, lbl := range op.Labels() {
				for _, idx := range labels[lbl.Name] {
					if idx.Before(from) || idx.After(to) {
						// TODO: check if the target label block is escapable
						result = true
						return false
					}
				}
			}
		case Terminator:
			result = true
			return false
		}

		return true
	})

	return result
}

type LintRule interface {
	Id() string
	Desc() string
}

type runnableRule interface {
	Run(l *Linter)
}

type DuplicateLabelsRule struct {
}

func (r DuplicateLabelsRule) Id() string {
	return "LINT0001"
}

func (r DuplicateLabelsRule) Desc() string {
	return "Checks for duplicate labels in a single TEAL source file"
}

func (r DuplicateLabelsRule) Run(l *Linter) {
	labels := l.getAllLabels()
	for name, positions := range labels {
		if len(positions) > 1 {
			for _, pos := range positions {
				l.errs = append(l.errs, DuplicateLabelError{p: pos, name: name, rule: r.Id()})
			}
		}
	}
}

type UnusedLabelsRule struct {
}

func (r UnusedLabelsRule) Id() string {
	return "LINT0002"
}

func (r UnusedLabelsRule) Desc() string {
	return "Checks for labels that are never referenced in the code"
}

func (r UnusedLabelsRule) Run(l *Linter) {
	used := l.getLabelsUsers()
	for name, positions := range l.getAllLabels() {
		if len(used[name]) == 0 {
			for _, p := range positions {
				l.errs = append(l.errs, &UnusedLabelError{p: p, name: name, rule: r.Id()})
				l.reds = append(l.reds, &RedundantLabelLine{p: p, name: name})
			}
		}
	}
}

type OpsAfterUnconditionalBranchRule struct {
}

func (r OpsAfterUnconditionalBranchRule) Id() string {
	return "LINT0003"
}

func (r OpsAfterUnconditionalBranchRule) Desc() string {
	return "Checks for ops after an unconditional branch call"
}

func (r OpsAfterUnconditionalBranchRule) Run(l *Linter) {
	used := l.getLabelsUsers()

	l.forEachForward(0, 0, func(i, j int) bool {
		o := l.l[i][j]

		check := false
		switch o.(type) {
		case *BExpr:
			check = true
		case *ReturnExpr:
			check = true
		case *ErrExpr:
			check = true
		}

		if check {
			l.forEachForward(i, j+1, func(i2, j2 int) bool {
				o2 := l.l[i2][j2]
				switch o2 := o2.(type) {
				case *LabelExpr:
					if len(used[o2.Name]) > 0 {
						return false
					}
				case Nop:
				default:
					l.errs = append(l.errs, UnreachableCodeError{p: position{l: i2, s: j2}, rule: r.Id()})
				}
				return true
			})
		}

		return true
	})
}

type CheckBranchJustBeforeLabelRule struct {
}

func (r CheckBranchJustBeforeLabelRule) Id() string {
	return "LINT0004"
}

func (r CheckBranchJustBeforeLabelRule) Desc() string {
	return "Checks for redundant branch calls that are placed just before their target labels"
}

func (r CheckBranchJustBeforeLabelRule) Run(l *Linter) {
	l.forEachForward(0, 0, func(i, j int) bool {
		o := l.l[i][j]
		switch o := o.(type) {
		case *BExpr:
			l.forEachForward(i, j+1, func(i2, j2 int) bool {
				o2 := l.l[i2][j2]
				switch o2 := o2.(type) {
				case *LabelExpr:
					if o2.Name == o.Label.Name {
						l.errs = append(l.errs, BJustBeforeLabelError{p: position{l: i, s: j}, rule: r.Id()})
						l.reds = append(l.reds, RedundantBLine{p: position{l: i, s: j}})
						return true
					}
					return false
				case Nop:
				default:
					return false
				}
				return true
			})
		}
		return true
	})
}

type CheckLoopsRule struct{}

func (r CheckLoopsRule) Id() string {
	return "LINT0005"
}

func (r CheckLoopsRule) Desc() string {
	return "Checks infinite loops presence"
}

func (r CheckLoopsRule) Run(l *Linter) {
	used := l.getLabelsUsers()
	all := l.getAllLabels()

	for name, users := range used {
		_, ok := all[name]
		if !ok {
			for _, user := range users {
				l.errs = append(l.errs, MissingLabelError{p: user, name: name, rule: r.Id()})
			}
		}
	}

	l.forEachForward(0, 0, func(i, j int) bool {
		o := l.l[i][j]
		switch o := o.(type) {
		case *BExpr:
			l.forEachBackward(i, j-1, func(i2, j2 int) bool {
				o2 := l.l[i2][j2]
				switch o2 := o2.(type) {
				case *LabelExpr:
					if o2.Name == o.Label.Name {
						if !l.canEscape(position{l: i2, s: j2}, position{l: i, s: j}) {
							l.errs = append(l.errs, InfiniteLoopError{p: position{l: i, s: j}, rule: r.Id()})
						}
					}
				}
				return true
			})
		}
		return true
	})

}

type CheckPragmaRule struct{}

func (r CheckPragmaRule) Id() string {
	return "LINT0006"
}

func (r CheckPragmaRule) Desc() string {
	return "Checks proper #pragma usage"
}

func (r CheckPragmaRule) Run(l *Linter) {
	var prev Op
	for i, ops := range l.l {
		for j, op := range ops {
			switch op := op.(type) {
			case *PragmaExpr:
				if prev != nil {
					l.errs = append(l.errs, PragmaVersionAfterInstrError{
						p:    position{l: i, s: j},
						rule: r.Id(),
					})
				}
				prev = op
			case Nop:
			default:
				prev = op
			}
		}
	}
}

type OpCodeAvailabilityInModeRule struct {
}

func (r OpCodeAvailabilityInModeRule) Id() string {
	return "LINT0007"
}

func (r OpCodeAvailabilityInModeRule) Desc() string {
	return "Checks opcode availability in the current mode (app or logicsig)"
}

var OpCodeAvailabilityInModeRuleInstance = OpCodeAvailabilityInModeRule{}

type OpCodeVersionCompatibilityCheckRule struct {
}

func (r OpCodeVersionCompatibilityCheckRule) Id() string {
	return "LINT0008"
}

func (r OpCodeVersionCompatibilityCheckRule) Desc() string {
	return "Checks opcode is available in the current version"
}

var OpCodeVersionCompatibilityCheckRuleInstance = OpCodeVersionCompatibilityCheckRule{}

var LintRules []LintRule

func init() {
	LintRules = append(LintRules, DuplicateLabelsRule{})
	LintRules = append(LintRules, UnusedLabelsRule{})
	LintRules = append(LintRules, OpsAfterUnconditionalBranchRule{})
	LintRules = append(LintRules, CheckBranchJustBeforeLabelRule{})
	LintRules = append(LintRules, CheckLoopsRule{})
	LintRules = append(LintRules, CheckPragmaRule{})
	LintRules = append(LintRules, OpCodeAvailabilityInModeRuleInstance)
	LintRules = append(LintRules, OpCodeVersionCompatibilityCheckRuleInstance)
}

func (l *Linter) Lint() {
	for _, r := range LintRules {
		switch r := r.(type) {
		case runnableRule:
			r.Run(l)
		}
	}
}
