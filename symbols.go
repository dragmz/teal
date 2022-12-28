package teal

type Symbol interface {
	Name() string
	Line() int
	Begin() int
	End() int

	StartLine() int
	StartCharacter() int

	EndLine() int
	EndCharacter() int

	Docs() string
}

type labelSymbol struct {
	n    string
	l    int
	b    int
	e    int
	docs string
}

func (s labelSymbol) Name() string {
	return s.n
}

func (s labelSymbol) Line() int {
	return s.l
}

func (s labelSymbol) Begin() int {
	return s.b
}

func (s labelSymbol) End() int {
	return s.e
}

func (s labelSymbol) StartLine() int {
	return s.l
}

func (s labelSymbol) StartCharacter() int {
	return s.b
}

func (s labelSymbol) EndLine() int {
	return s.l
}

func (s labelSymbol) EndCharacter() int {
	return s.e
}

func (s labelSymbol) Docs() string {
	return s.docs
}
