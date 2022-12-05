package lsp

var (
	EOL = []string{"\n", "\r\n", "\r"}
)

type doc struct {
	s     string
	lines []string
}

func newDoc(s string) *doc {
	return &doc{
		s:     s,
		lines: splitLines(s),
	}
}

func splitLines(s string) []string {
	res := []string{}

	p := 0
	l := len(s)
	i := 0

	for {
		if i == l {
			if p == l {
				res = append(res, "")
			}
			break
		}

		c := s[i]

		split := c == '\r' || c == '\n'

		if split {
			res = append(res, s[p:i])
		}

		if c == '\r' {
			if i < l-1 && s[i+1] == '\n' {
				i++
			}
		}

		if split {
			i++
			p = i
		} else {
			i++
		}
	}

	if p < l {
		res = append(res, s[p:i])
	}

	return res
}
