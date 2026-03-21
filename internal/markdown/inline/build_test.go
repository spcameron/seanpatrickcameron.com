package inline

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/require"
)

func TestBuild(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []InlineSummary
		wantErr error
	}{
		{
			name:    "empty input",
			input:   "",
			want:    []InlineSummary{},
			wantErr: nil,
		},
		{
			name:  "plain text",
			input: "abc",
			want: []InlineSummary{
				{
					Kind:   "text",
					Lexeme: "abc",
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := source.NewSource(tc.input)
			span := source.ByteSpan{
				Start: 0,
				End:   src.EOF(),
			}

			tokens, err := Scan(src, span)
			require.NoError(t, err)

			inlines, err := Build(src, span, tokens)
			got := summarizeInlines(src, inlines)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
