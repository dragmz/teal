package teal

type DiagnosticSeverity int

func (s DiagnosticSeverity) String() string {
	switch s {
	case DiagInfo:
		return "info"
	case DiagWarn:
		return "warn"
	case DiagErr:
		return "error"
	default:
		panic("unsupported severity")
	}
}

const (
	DiagErr  = 1
	DiagWarn = 2
	DiagInfo = 3
	DiagHint = 4
)

type Diagnostic interface {
	Line() int

	Begin() int
	End() int

	String() string
	Severity() DiagnosticSeverity

	Rule() string
}
