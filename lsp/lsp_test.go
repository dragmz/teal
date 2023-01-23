package lsp

import (
	"fmt"
	"testing"

	"github.com/dragmz/teal"
	"github.com/stretchr/testify/assert"
)

func TestPrepareVersionEditForNil(t *testing.T) {
	type test struct {
		token   teal.Range
		version uint64

		text string
		rg   lspRange
	}

	tests := []test{
		{
			token:   nil,
			version: 8,
			text:    "#pragma version 8\r\n",
		},
		{
			token: lspRange{
				Start: lspPosition{
					Line:      1,
					Character: 10,
				},
				End: lspPosition{
					Line:      1,
					Character: 11,
				},
			},
			version: 9,
			text:    "9",
			rg: lspRange{
				Start: lspPosition{
					Line:      1,
					Character: 10,
				},
				End: lspPosition{
					Line:      1,
					Character: 11,
				},
			},
		},
	}

	for i, ts := range tests {
		name := fmt.Sprintf("test #%d", i)

		e := prepareVersionEdit(ts.token, ts.version)

		assert.Equal(t, ts.text, e.NewText, name)
		assert.Equal(t, ts.rg, e.Range, name)
	}
}
