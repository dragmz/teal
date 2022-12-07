package teal

type Symbol interface {
	Name() string
	Line() int
	Begin() int
	End() int
}

type labelSymbol struct {
	n string
	l int
}

func (s labelSymbol) Name() string {
	return s.n
}

func (s labelSymbol) Line() int {
	return s.l
}

func (s labelSymbol) Begin() int {
	return 0
}

func (s labelSymbol) End() int {
	return len(s.n) + 1
}
