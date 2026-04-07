package block

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestScan(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []string
		wantErr error
	}{
		{
			name:    "empty input",
			input:   "",
			want:    nil,
			wantErr: nil,
		},
		{
			name:  "single line, no newline",
			input: "hello",
			want: []string{
				"hello",
			},
			wantErr: nil,
		},
		{
			name:  "single line, trailing newline preserved",
			input: "hello\n",
			want: []string{
				"hello",
				"",
			},
			wantErr: nil,
		},
		{
			name:  "multiple nonblank lines without trailing newline",
			input: "a\nb\nc",
			want: []string{
				"a",
				"b",
				"c",
			},
			wantErr: nil,
		},
		{
			name:  "multiple nonblank lines with trailing newline",
			input: "a\nb\nc\n",
			want: []string{
				"a",
				"b",
				"c",
				"",
			},
			wantErr: nil,
		},
		{
			name:  "only newline emits empty line",
			input: "\n",
			want: []string{
				"",
				"",
			},
			wantErr: nil,
		},
		{
			name:  "only blank lines",
			input: "\n\n",
			want: []string{
				"",
				"",
				"",
			},
			wantErr: nil,
		},
		{
			name:  "single blank line preserved as delimiter",
			input: "a\n\nb",
			want: []string{
				"a",
				"",
				"b",
			},
			wantErr: nil,
		},
		{
			name:  "leading blank lines preserved",
			input: "\n\na",
			want: []string{
				"",
				"",
				"a",
			},
			wantErr: nil,
		},
		{
			name:  "trailing blank line delimiter preserved",
			input: "a\n\n",
			want: []string{
				"a",
				"",
				"",
			},
			wantErr: nil,
		},
		{
			name:  "multiple blank lines preserved",
			input: "a\n\n\nb",
			want: []string{
				"a",
				"",
				"",
				"b",
			},
			wantErr: nil,
		},
		{
			name:  "three terminal newlines produce trailing empty logical lines",
			input: "a\n\n\n",
			want: []string{
				"a",
				"",
				"",
				"",
			},
			wantErr: nil,
		},
		{
			name:  "CRLF normalized",
			input: "a\r\nb\r\n",
			want: []string{
				"a",
				"b",
				"",
			},
			wantErr: nil,
		},
		{
			name:  "mixed newline styles across multiple lines",
			input: "a\r\nb\rc\n",
			want: []string{
				"a",
				"b",
				"c",
				"",
			},
			wantErr: nil,
		},
		{
			name:  "multiple CRLF blank lines preserved after normalization",
			input: "a\r\n\r\nb",
			want: []string{
				"a",
				"",
				"b",
			},
			wantErr: nil,
		},
		{
			name:  "only carriage returns normalize into blank lines",
			input: "\r\r",
			want: []string{
				"",
				"",
				"",
			},
			wantErr: nil,
		},
		{
			name:  "trailing spaces are preserved",
			input: "a \n",
			want: []string{
				"a ",
				"",
			},
			wantErr: nil,
		},
		{
			name:  "trailing spaces and tabs are preserved",
			input: " indented\t \nnext\t\n",
			want: []string{
				" indented\t ",
				"next\t",
				"",
			},
			wantErr: nil,
		},
		{
			name:  "trailing carriage return is normalized",
			input: "a\r\n",
			want: []string{
				"a",
				"",
			},
			wantErr: nil,
		},
		{
			name:  "whitespace only line preserves spaces and tabs",
			input: "a\n \t \n b",
			want: []string{
				"a",
				" \t ",
				" b",
			},
			wantErr: nil,
		},
		{
			name:  "terminal whitespace only line preserves spaces and tab",
			input: "a\n \t \n",
			want: []string{
				"a",
				" \t ",
				"",
			},
			wantErr: nil,
		},
		{
			name:  "single whitespace only line without trailing newline",
			input: " \t ",
			want: []string{
				" \t ",
			},
			wantErr: nil,
		},
		{
			name:  "no newline still preserves trailing spaces and tabs",
			input: "a \t",
			want: []string{
				"a \t",
			},
			wantErr: nil,
		},
		{
			name:  "embedded carriage return is normalized",
			input: "a\rb\n",
			want: []string{
				"a",
				"b",
				"",
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := source.NewSource(tc.input)
			gotLines, err := Scan(src)

			var got []string
			for _, line := range gotLines {
				got = append(got, src.Slice(line.Span))
			}

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestScanSpans(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  []source.ByteSpan
	}{
		{
			name:  "single line no newline",
			input: "hello",
			want: []source.ByteSpan{
				{Start: 0, End: 5},
			},
		},
		{
			name:  "single line with trailing newline includes final empty line at EOF",
			input: "hello\n",
			want: []source.ByteSpan{
				{Start: 0, End: 5},
				{Start: 6, End: 6},
			},
		},
		{
			name:  "blank line between content lines",
			input: "a\n\nb",
			want: []source.ByteSpan{
				{Start: 0, End: 1},
				{Start: 2, End: 2},
				{Start: 3, End: 4},
			},
		},
		{
			name:  "only newline produces two empty spans",
			input: "\n",
			want: []source.ByteSpan{
				{Start: 0, End: 0},
				{Start: 1, End: 1},
			},
		},
		{
			name:  "trailing blank lines each become explicit EOF-adjacent spans",
			input: "a\n\n",
			want: []source.ByteSpan{
				{Start: 0, End: 1},
				{Start: 2, End: 2},
				{Start: 3, End: 3},
			},
		},
		{
			name:  "normalized CRLF input produces spans over normalized buffer",
			input: "a\r\nb\r\n",
			want: []source.ByteSpan{
				{Start: 0, End: 1},
				{Start: 2, End: 3},
				{Start: 4, End: 4},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := source.NewSource(tc.input)
			gotLines, err := Scan(src)
			assert.ErrorIs(t, err, nil)

			var got []source.ByteSpan
			for _, line := range gotLines {
				got = append(got, line.Span)
			}

			assert.Equal(t, got, tc.want)
		})
	}
}
