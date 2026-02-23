package block

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	tk "github.com/spcameron/seanpatrickcameron.com/internal/markdown/testkit"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantIR  ir.Document
		wantErr error
	}{
		{
			name:    "empty input",
			input:   "",
			wantIR:  ir.Document{},
			wantErr: nil,
		},
		{
			name:    "one paragraph",
			input:   "a",
			wantIR:  tk.IRDoc(tk.IRPara("a")),
			wantErr: nil,
		},
		{
			name:    "two paragraphs",
			input:   "a\n\nb",
			wantIR:  tk.IRDoc(tk.IRPara("a"), tk.IRParaAt(2, "b")),
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input)

			assert.Equal(t, got, tc.wantIR)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestScan(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []Line
		wantErr error
	}{
		{
			name:    "empty input",
			input:   "",
			want:    []Line{},
			wantErr: nil,
		},
		{
			name:  "single line, no newline",
			input: "hello",
			want: []Line{
				{"hello"},
			},
			wantErr: nil,
		},
		{
			name:  "single line, trailing newline suppressed",
			input: "hello\n",
			want: []Line{
				{"hello"},
			},
			wantErr: nil,
		},
		{
			name:    "only newline emits empty doc",
			input:   "\n",
			want:    []Line{},
			wantErr: nil,
		},
		{
			name:  "single blank line preserved as delimiter",
			input: "a\n\nb",
			want: []Line{
				{"a"},
				{""},
				{"b"},
			},
			wantErr: nil,
		},
		{
			name:  "leading blank lines preserved",
			input: "\n\na",
			want: []Line{
				{""},
				{""},
				{"a"},
			},
			wantErr: nil,
		},
		{
			name:  "trailing blank line delimiter preserved",
			input: "a\n\n",
			want: []Line{
				{"a"},
				{""},
			},
			wantErr: nil,
		},
		{
			name:  "multiple blank lines preserved",
			input: "a\n\n\nb",
			want: []Line{
				{"a"},
				{""},
				{""},
				{"b"},
			},
			wantErr: nil,
		},
		{
			name:  "CRLF normalized",
			input: "a\r\nb\r\n",
			want: []Line{
				{"a"},
				{"b"},
			},
			wantErr: nil,
		},
		{
			name:  "right trim spaces and tabs",
			input: " indented\t \nnext\t\n",
			want: []Line{
				{" indented"},
				{"next"},
			},
			wantErr: nil,
		},
		{
			name:  "whitespace only line emits blank line",
			input: "a\n \t \n b",
			want: []Line{
				{"a"},
				{""},
				{" b"},
			},
			wantErr: nil,
		},
		{
			name:  "terminal whitespace only line with trailing newline suppressed",
			input: "a\n \t \n",
			want: []Line{
				{"a"},
				{""},
			},
			wantErr: nil,
		},
		{
			name:  "no newline but trailing spaces trimmed",
			input: "a \t",
			want: []Line{
				{"a"},
			},
			wantErr: nil,
		},
		{
			name:  "embedded carriage return is preserved",
			input: "a\rb\n",
			want: []Line{
				{"a\rb"},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Scan(tc.input)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestBuild(t *testing.T) {
	testCases := []struct {
		name    string
		lines   []Line
		want    ir.Document
		wantErr error
	}{
		{
			name:    "empty input",
			lines:   []Line{},
			want:    ir.Document{},
			wantErr: nil,
		},
		{
			name: "only blank lines",
			lines: []Line{
				{" "},
				{"\t"},
			},
			want:    ir.Document{},
			wantErr: nil,
		},
		{
			name: "single paragraph, one line",
			lines: []Line{
				{"a"},
			},
			want:    tk.IRDoc(tk.IRPara("a")),
			wantErr: nil,
		},
		{
			name: "single paragraph, multiple lines",
			lines: []Line{
				{"a"},
				{"b"},
				{"c"},
			},
			want:    tk.IRDoc(tk.IRPara("a", "b", "c")),
			wantErr: nil,
		},
		{
			name: "leading blank lines ignored",
			lines: []Line{
				{""},
				{""},
				{"a"},
			},
			want:    tk.IRDoc(tk.IRParaAt(2, "a")),
			wantErr: nil,
		},
		{
			name: "trailing blank lines ignored",
			lines: []Line{
				{"a"},
				{""},
				{""},
			},
			want:    tk.IRDoc(tk.IRPara("a")),
			wantErr: nil,
		},
		{
			name: "two paragraphs separated by one blank line",
			lines: []Line{
				{"a"},
				{""},
				{"b"},
			},
			want:    tk.IRDoc(tk.IRPara("a"), tk.IRParaAt(2, "b")),
			wantErr: nil,
		},
		{
			name: "two paragraphs separated by two blank lines",
			lines: []Line{
				{"a"},
				{""},
				{""},
				{"b"},
			},
			want:    tk.IRDoc(tk.IRPara("a"), tk.IRParaAt(3, "b")),
			wantErr: nil,
		},
		{
			name: "two paragrapsh separated by whitespace only line",
			lines: []Line{
				{"a"},
				{" "},
				{"b"},
			},
			want:    tk.IRDoc(tk.IRPara("a"), tk.IRParaAt(2, "b")),
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Build(tc.lines)

			assert.Equal(t, got, tc.want)
			assert.Equal(t, err, tc.wantErr)
		})
	}
}
