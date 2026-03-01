package block

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	tk "github.com/spcameron/seanpatrickcameron.com/internal/markdown/testkit"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/require"
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
			name:  "only newline emits empty line",
			input: "\n",
			want: []string{
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

func TestBuild(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    ir.Document
		wantErr error
	}{
		{
			name:    "empty input",
			input:   "",
			want:    ir.Document{},
			wantErr: nil,
		},
		{
			name:    "only blank lines",
			input:   " \n\t",
			want:    ir.Document{},
			wantErr: nil,
		},
		{
			name:  "single paragraph, one line",
			input: "a",
			want: tk.IRDoc(
				tk.IRPara("a"),
			),
			wantErr: nil,
		},
		{
			name:  "single paragraph, multiple lines",
			input: "a\nb\nc",
			want: tk.IRDoc(
				tk.IRPara("a", "b", "c"),
			),
			wantErr: nil,
		},
		{
			name:  "leading blank lines ignored",
			input: "\n\na",
			want: tk.IRDoc(
				tk.IRPara("a"),
			),
			wantErr: nil,
		},
		{
			name:  "trailing blank lines ignored",
			input: "a\n\n",
			want: tk.IRDoc(
				tk.IRPara("a"),
			),
			wantErr: nil,
		},
		{
			name:  "two paragraphs separated by one blank line",
			input: "a\n\nb",
			want: tk.IRDoc(
				tk.IRPara("a"),
				tk.IRPara("b"),
			),
			wantErr: nil,
		},
		{
			name:  "two paragraphs separated by two blank lines",
			input: "a\n\n\nb",
			want: tk.IRDoc(
				tk.IRPara("a"),
				tk.IRPara("b"),
			),
			wantErr: nil,
		},
		{
			name:  "two paragraphs separated by whitespace only line",
			input: "a\n \nb",
			want: tk.IRDoc(
				tk.IRPara("a"),
				tk.IRPara("b"),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph stops before header without blank line",
			input: "a\n# h",
			want: tk.IRDoc(
				tk.IRPara("a"),
				tk.IRHeader(1, "h"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1",
			input: "# header",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 2",
			input: "## header",
			want: tk.IRDoc(
				tk.IRHeader(2, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 6",
			input: "###### header",
			want: tk.IRDoc(
				tk.IRHeader(6, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1, 3 leading spaces (max)",
			input: "   # header",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1, tab delimiter",
			input: "#\theader",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1, consumes multiple spaces",
			input: "#     header",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1, consumes multiple tabs",
			input: "#\t\t\theader",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1, trailing whitespace trimmed",
			input: "# header     ",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1, mixed whitespace trimmed",
			input: "# \t header \t ",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1, empty header allowed",
			input: "# ",
			want: tk.IRDoc(
				tk.IRHeader(1, ""),
			),
			wantErr: nil,
		},
		{
			name:  "header and paragraph",
			input: "# h\na",
			want: tk.IRDoc(
				tk.IRHeader(1, "h"),
				tk.IRPara("a"),
			),
			wantErr: nil,
		},
		{
			name:  "header rejected, no marker",
			input: "header",
			want: tk.IRDoc(
				tk.IRPara("header"),
			),
			wantErr: nil,
		},
		{
			name:  "header rejected, too many leading spaces",
			input: "    header",
			want: tk.IRDoc(
				tk.IRPara("    header"),
			),
			wantErr: nil,
		},
		{
			name:  "header rejected, missing delimiter",
			input: "#header",
			want: tk.IRDoc(
				tk.IRPara("#header"),
			),
			wantErr: nil,
		},
		{
			name:  "header rejected, too many hashes",
			input: "####### header",
			want: tk.IRDoc(
				tk.IRPara("####### header"),
			),
			wantErr: nil,
		},
		{
			name:  "header rejected, too many hashes after indent",
			input: "   ####### header",
			want: tk.IRDoc(
				tk.IRPara("   ####### header"),
			),
			wantErr: nil,
		},
		{
			name:  "header rejected, valid marker but missing delimieter",
			input: "##",
			want: tk.IRDoc(
				tk.IRPara("##"),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break (---)",
			input: "---",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break (***)",
			input: "***",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break (___)",
			input: "___",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break, leading spaces",
			input: "   ---",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break, inter-marker whitespace",
			input: "- \t - \t -",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break, trailing whitespace",
			input: "---   ",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break, more than three identical markers",
			input: "-----------------------",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break rejected, too many leading spaces",
			input: "    ---",
			want: tk.IRDoc(
				tk.IRPara("    ---"),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break rejected, tabs in leading whitespace",
			input: "\t---",
			want: tk.IRDoc(
				tk.IRPara("\t---"),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break rejected, mixed marker characters",
			input: "-*-",
			want: tk.IRDoc(
				tk.IRPara("-*-"),
			),
			wantErr: nil,
		},
		{
			name:  "block quote, plain text",
			input: "> text",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("text"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote, leading spaces",
			input: "   > text",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("text"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote, tab delimiter",
			input: ">\ttext",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("text"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote, no delimiter",
			input: ">text",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("text"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote, multiple lines",
			input: "> a\n> b",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("a", "b"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote, containing header",
			input: "> # header",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRHeader(1, "header"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote rejected, too many leading spaces",
			input: "    > text",
			want: tk.IRDoc(
				tk.IRPara("    > text"),
			),
			wantErr: nil,
		},
		{
			name:  "block quote, blank line splits paragraphs",
			input: "> a\n>\n> b",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("a"),
					tk.IRPara("b"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote, starts with a blank line",
			input: ">\n> a",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("a"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote, only blank lines",
			input: ">\n>\n",
			want: tk.IRDoc(
				tk.IRBlockQuote(),
			),
			wantErr: nil,
		},
		{
			name:  "block quote, mixed indentation across lines",
			input: "> a\n > b\n  > c",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("a", "b", "c"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote, terminates on first non-quote line",
			input: "> a\nb",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("a"),
				),
				tk.IRPara("b"),
			),
			wantErr: nil,
		},
		{
			name:  "block quote, terminates on truly blank line",
			input: "> a\n\n> b",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("a"),
				),
				tk.IRBlockQuote(
					tk.IRPara("b"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "nested quotes, via >>",
			input: ">> a",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRBlockQuote(
						tk.IRPara("a"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "nested quotes, via > >",
			input: "> > a",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRBlockQuote(
						tk.IRPara("a"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "nested quotes, via >\t>",
			input: ">\t> a",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRBlockQuote(
						tk.IRPara("a"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "mixed nested quotes across lines",
			input: "> a\n>> b\n> c",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("a"),
					tk.IRBlockQuote(
						tk.IRPara("b"),
					),
					tk.IRPara("c"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "nested quote separated by quoted blank lines",
			input: "> a\n>\n>> b",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("a"),
					tk.IRBlockQuote(
						tk.IRPara("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "setext heading, h1 (minimum)",
			input: "heading\n=",
			want: tk.IRDoc(
				tk.IRHeader(1, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext heading, h1 (typical)",
			input: "heading\n===",
			want: tk.IRDoc(
				tk.IRHeader(1, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext heading, h2 (minimum)",
			input: "heading\n-",
			want: tk.IRDoc(
				tk.IRHeader(2, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext heading, h2 (typical)",
			input: "heading\n---",
			want: tk.IRDoc(
				tk.IRHeader(2, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext heading and paragraph",
			input: "heading\n---\nnext",
			want: tk.IRDoc(
				tk.IRHeader(2, "heading"),
				tk.IRPara("next"),
			),
			wantErr: nil,
		},
		{
			name:  "setext heading, trailing spaces",
			input: "heading\n===   ",
			want: tk.IRDoc(
				tk.IRHeader(1, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext heading, trailing tabs",
			input: "heading\n===\t\t",
			want: tk.IRDoc(
				tk.IRHeader(1, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext heading, mixed trailing spaces and tabs",
			input: "heading\n===\t \t ",
			want: tk.IRDoc(
				tk.IRHeader(1, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext heading, leading spaces",
			input: "heading\n   ---",
			want: tk.IRDoc(
				tk.IRHeader(2, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected, too many leading spaces",
			input: "heading\n    ---",
			want: tk.IRDoc(
				tk.IRPara("heading", "    ---"),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected, underline with internal spaces (dash)",
			input: "heading\n- - -",
			want: tk.IRDoc(
				tk.IRPara("heading"),
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected, underline with internal spaces (equals)",
			input: "heading\n= = =",
			want: tk.IRDoc(
				tk.IRPara("heading", "= = ="),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected, underline with non-marker character",
			input: "heading\n--x--",
			want: tk.IRDoc(
				tk.IRPara("heading", "--x--"),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected, underline with mixed markers",
			input: "heading\n-=-",
			want: tk.IRDoc(
				tk.IRPara("heading", "-=-"),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected, underline with trailing non-space",
			input: "heading\n---x",
			want: tk.IRDoc(
				tk.IRPara("heading", "---x"),
			),
			wantErr: nil,
		},
		{
			name:  "setext, h2 takes precedence over thematic break for '---'",
			input: "heading\n---\nnext",
			want: tk.IRDoc(
				tk.IRHeader(2, "heading"),
				tk.IRPara("next"),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break, '- - -' does not become setext",
			input: "heading\n- - -\nnext",
			want: tk.IRDoc(
				tk.IRPara("heading"),
				tk.IRThematicBreak(),
				tk.IRPara("next"),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break, '***' does not become setext",
			input: "heading\n***\nnext",
			want: tk.IRDoc(
				tk.IRPara("heading"),
				tk.IRThematicBreak(),
				tk.IRPara("next"),
			),
			wantErr: nil,
		},
		{
			name:  "setext, multiline content",
			input: "line1\nline2\n---",
			want: tk.IRDoc(
				tk.IRHeader(2, "line1", "line2"),
			),
			wantErr: nil,
		},
		{
			name:  "setext, multiline content stops before underline",
			input: "line1\nline2\n===\nnext",
			want: tk.IRDoc(
				tk.IRHeader(1, "line1", "line2"),
				tk.IRPara("next"),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected, blank line between content and underline",
			input: "heading\n\n---",
			want: tk.IRDoc(
				tk.IRPara("heading"),
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "setext, underline followed by blank line",
			input: "heading\n---\n\nnext",
			want: tk.IRDoc(
				tk.IRHeader(2, "heading"),
				tk.IRPara("next"),
			),
			wantErr: nil,
		},
		{
			name:  "setext, underline at start of doc is not a heading (dashes)",
			input: "---",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "setext, underline at start of doc is not a heading (equals)",
			input: "===",
			want: tk.IRDoc(
				tk.IRPara("==="),
			),
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := source.NewSource(tc.input)

			lines, err := Scan(src)
			require.NoError(t, err)

			got, err := Build(src, lines)

			got = tk.NormalizeIR(got)
			want := tk.NormalizeIR(tc.want)

			assert.Equal(t, got, want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
