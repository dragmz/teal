package teal

type Range interface {
	StartLine() int
	StartCharacter() int
	EndLine() int
	EndCharacter() int
}

func Contains(r Range, l, ch int) bool {
	if r.StartLine() > l || r.StartLine() == l && r.StartCharacter() > ch {
		return false
	}

	if r.EndLine() < l || r.EndLine() == l && r.EndCharacter() < ch {
		return false
	}

	return true
}

func Within(what, where Range) bool {
	return Contains(where, what.StartLine(), what.StartCharacter()) && Contains(where, what.EndLine(), what.EndCharacter())
}

func Overlaps(a, b Range) bool {
	if Within(a, b) || Within(b, a) {
		return true
	}

	if Contains(a, b.StartLine(), b.StartCharacter()) || Contains(a, b.EndLine(), b.EndCharacter()) {
		return true
	}

	if Contains(b, a.StartLine(), a.StartCharacter()) || Contains(b, a.EndLine(), a.EndCharacter()) {
		return true
	}

	return false
}

type LineRange int

func (r LineRange) StartLine() int {
	return int(r)
}
