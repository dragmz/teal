package teal

import "sort"

type SemanticToken struct {
	Line      int
	Index     int
	Length    int
	Type      int
	Modifiers int
}

type SemanticTokens []SemanticToken

func (t SemanticTokens) Encode() []uint32 {
	sort.Sort(t)

	res := make([]uint32, len(t)*5)

	prev := SemanticToken{}

	j := 0
	for _, st := range t {
		res[j] = uint32(st.Line - prev.Line)
		if st.Line == prev.Line {
			res[j+1] = uint32(st.Index - prev.Index)
		} else {
			res[j+1] = uint32(st.Index)
		}
		res[j+2] = uint32(st.Length)
		res[j+3] = uint32(st.Type)
		res[j+4] = uint32(st.Modifiers)

		j += 5
		prev = st
	}

	return res
}

func (t SemanticTokens) Len() int {
	return len(t)
}

func (t SemanticTokens) Less(i, j int) bool {
	ti := t[i]
	tj := t[j]

	if ti.Line == tj.Line {
		return ti.Index < tj.Index
	}

	return ti.Line < tj.Line
}

func (t SemanticTokens) Swap(i, j int) {
	temp := t[i]
	t[i] = t[j]
	t[j] = temp
}
