package teal

import (
	"fmt"
	"strings"
)

type compiler struct {
	labels map[string]*compiledLabelExpr
}

type Program []Expr
type Listing []Op

type usesLabels interface {
	Labels() []*compiledLabelExpr
}

type Terminator interface {
	IsTerminator()
}

type Op interface {
	fmt.Stringer
}

type complexExpr interface {
	Compile(c *compiler) []Op
}

type Expr interface {
}

func (c *compiler) getLabel(name string) *compiledLabelExpr {
	l, ok := c.labels[name]

	if !ok {
		l = &compiledLabelExpr{Name: name}
		c.labels[name] = l
	}

	return l
}

func Compile(exprs []Expr) Listing {
	c := &compiler{
		labels: make(map[string]*compiledLabelExpr),
	}

	l := Program{exprs}.Compile(c)

	return l
}

func removeOpsAfterUnconditionalBranch(l Listing) Listing {
	var res Listing

	for i := 0; i < len(l); i++ {
		o := l[i]
		skip := false
		switch o.(type) {
		case *compiledBExpr:
			skip = true
		case *ReturnExpr:
			skip = true
		case *ErrExpr:
			skip = true
		}

		res = append(res, o)

		if skip {
		loop:
			for i = i + 1; i < len(l); i++ {
				o2 := l[i]
				switch o2.(type) {
				case Nop:
					res = append(res, o2)
				case *compiledLabelExpr:
					res = append(res, o2)
					break loop
				}
			}
		}
	}
	return res
}

func removeBJustBeforeItsTargetLabel(l Listing) Listing {
	var res Listing

	for i, o := range l {
		func() {
			if i >= len(l)-1 {
				res = append(res, o)
				return
			}

			switch o := o.(type) {
			case *compiledBExpr:
				j := i + 1

				func() {
				loop:
					for {
						if j >= len(l) {
							break
						}

						n := l[j]
						j += 1

						switch n := n.(type) {
						case Nop:
						case *compiledLabelExpr:
							if n.Name == o.Label.Name {
								return
							}
							break loop
						default:
							break loop
						}
					}

					res = append(res, o)
				}()
			default:
				res = append(res, o)
			}
		}()
	}

	return res
}

func removeUnused(l Listing) Listing {
	var res Listing

	used := map[string]bool{}

	for _, o := range l {
		switch o := o.(type) {
		case usesLabels:
			for _, l := range o.Labels() {
				used[l.Name] = true
			}
		}
	}

	var prev Expr

	for i, o := range l {
		switch o := o.(type) {
		case *compiledLabelExpr:
			keep := used[o.Name]
			if !keep {
			loop:
				for j := i + 1; j < len(l); j++ {
					switch l[j].(type) {
					case *RetSubExpr:
						keep = true
						break loop
					case *ProtoExpr:
						keep = true
						break loop
					case Nop:
						continue
					default:
						break loop
					}
				}

				if !keep {
					continue
				}
			}
			res = append(res, o)
		case *RetSubExpr:
			switch o2 := prev.(type) {
			case *LabelExpr:
				if !used[o2.Name] {
					continue
				}

				res = append(res, o)
			}
		default:
			res = append(res, o)
		}

		prev = o
	}

	return res
}

func mergeLabels(l Listing) Listing {
	var res Listing

	var prev Op
	for _, o := range l {
		func() {
			switch o := o.(type) {
			case *compiledLabelExpr:
				if prev != nil {
					switch p := prev.(type) {
					case *compiledLabelExpr:
						o.Name = p.Name
						return
					}
				}
			}

			res = append(res, o)
		}()

		switch o.(type) {
		case Nop:
		default:
			prev = o
		}
	}

	return res
}

func (l Listing) Lint() []LineError {
	lnt := &Linter{l: l}
	lnt.Lint()

	return lnt.res
}

func (l Listing) Reconstruct() Program {
	var p Program
	for _, op := range l {
		switch op := op.(type) {
		default:
			fmt.Print(op)
		}
	}
	return p
}

func (l Listing) Optimize() Listing {
	res := l

	res = removeUnused(res)
	res = removeOpsAfterUnconditionalBranch(res)
	res = removeBJustBeforeItsTargetLabel(res)
	res = mergeLabels(res)

	return res
}

func (l Listing) String() string {
	var b strings.Builder

	for _, op := range l {
		b.WriteString(op.String())
		b.WriteString("\n")
	}

	return b.String()
}

func (p Program) String() string {
	return Compile(p).String()
}

func (p Program) Compile(c *compiler) Listing {
	var l Listing

	for _, e := range p {
		l = c.compile(l, e)
	}

	return l
}

func (c *compiler) compile(to []Op, exprs ...Expr) []Op {
	for i, e := range exprs {
		switch e := e.(type) {
		case []Expr:
			for _, e := range e {
				to = c.compile(to, e)
			}
		case complexExpr:
			to = append(to, e.Compile(c)...)
		default:
			switch e := e.(type) {
			case Op:
				to = append(to, e)
			default:
				panic(fmt.Sprintf("unsupported expr: %d %#v", i, e))
			}
		}
	}
	return to
}
