package teal

import (
	"fmt"
)

type RedundantLine interface {
	Line() int
	String() string
}

type RedundantLabelLine struct {
	line int
	name string
}

func (l RedundantLabelLine) Line() int {
	return l.line
}

func (l RedundantLabelLine) String() string {
	return fmt.Sprintf("Remove label '%s'", l.name)
}

type RedundantBLine struct {
	line int
}

func (l RedundantBLine) Line() int {
	return l.line
}

func (l RedundantBLine) String() string {
	return "Remove b call"
}

type LineError interface {
	error
	Line() int
	Severity() DiagnosticSeverity
	Rule() string
}

type Linter struct {
	l Listing

	errs []LineError
	reds []RedundantLine
}

type DuplicateLabelError struct {
	l    int
	name string
	rule string
}

func (e DuplicateLabelError) Line() int {
	return e.l
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
	l    int
	name string
	rule string
}

func (e *UnusedLabelError) Line() int {
	return e.l
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
	l    int
	rule string
}

func (e UnreachableCodeError) Line() int {
	return e.l
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
	l    int
	rule string
}

func (e BJustBeforeLabelError) Line() int {
	return e.l
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
	l    int
	name string
	rule string
}

func (e MissingLabelError) Line() int {
	return e.l
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
	l    int
	rule string
}

func (e InfiniteLoopError) Line() int {
	return e.l
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
	l    int
	rule string
}

func (e PragmaVersionAfterInstrError) Line() int {
	return e.l
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
	for name, lines := range labels {
		if len(lines) > 1 {
			for _, line := range lines {
				l.errs = append(l.errs, DuplicateLabelError{l: line, name: name, rule: r.Id()})
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
	for name, lines := range l.getAllLabels() {
		if len(used[name]) == 0 {
			for _, line := range lines {
				l.errs = append(l.errs, &UnusedLabelError{l: line, name: name, rule: r.Id()})
				l.reds = append(l.reds, &RedundantLabelLine{line: line, name: name})
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
				case *LabelExpr:
					if len(used[o2.Name]) > 0 {
						break loop
					}
				case Nop:
				default:
					l.errs = append(l.errs, UnreachableCodeError{l: i, rule: r.Id()})
				}
			}
		}
	}
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
						case *LabelExpr:
							if n.Name == o.Label.Name {
								l.errs = append(l.errs, BJustBeforeLabelError{l: i, rule: r.Id()})
								l.reds = append(l.reds, RedundantBLine{line: i})
								return
							}
							break loop
						case Nop:
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
				l.errs = append(l.errs, MissingLabelError{l: user, name: name, rule: r.Id()})
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
									l.errs = append(l.errs, InfiniteLoopError{l: i, rule: r.Id()})
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

type CheckPragmaRule struct{}

func (r CheckPragmaRule) Id() string {
	return "LINT0006"
}

func (r CheckPragmaRule) Desc() string {
	return "Checks proper #pragma usage"
}

func (r CheckPragmaRule) Run(l *Linter) {
	var prev Op
	for i, op := range l.l {
		switch op := op.(type) {
		case *PragmaExpr:
			if prev != nil {
				l.errs = append(l.errs, PragmaVersionAfterInstrError{
					l:    i,
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
