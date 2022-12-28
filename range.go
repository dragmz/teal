package teal

type Range interface {
	StartLine() int
	StartCharacter() int
	EndLine() int
	EndCharacter() int
}

func Overlaps(a, b Range) bool {
	return a.StartCharacter() <= b.EndCharacter() && b.StartCharacter() <= a.EndCharacter() &&
		a.StartLine() <= b.EndLine() && b.StartLine() <= b.EndLine()
}
