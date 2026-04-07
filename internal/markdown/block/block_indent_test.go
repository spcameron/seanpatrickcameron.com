package block

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestBlockIndent(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		indentCols  int
		indentBytes int
	}{
		{
			name:        "four spaces",
			input:       "    x",
			indentCols:  4,
			indentBytes: 4,
		},
		{
			name:        "one tab",
			input:       "\tx",
			indentCols:  4,
			indentBytes: 1,
		},
		{
			name:        "one space, one tab",
			input:       " \tx",
			indentCols:  4,
			indentBytes: 2,
		},
		{
			name:        "two spaces, one tab",
			input:       "  \tx",
			indentCols:  4,
			indentBytes: 3,
		},
		{
			name:        "three spaces, one tab",
			input:       "   \tx",
			indentCols:  4,
			indentBytes: 4,
		},
		{
			name:        "four spaces, one tab",
			input:       "    \tx",
			indentCols:  8,
			indentBytes: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := source.NewSource(tc.input)
			span := src.LineSpan(0)

			line := Line{span}
			indentCols, indentBytes := line.BlockIndent(src)

			assert.Equal(t, indentCols, tc.indentCols)
			assert.Equal(t, indentBytes, tc.indentBytes)
		})
	}
}
