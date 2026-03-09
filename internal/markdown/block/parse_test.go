package block

import (
	"strings"
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
			want:    tk.IRDoc(),
			wantErr: nil,
		},
		{
			name:    "only blank lines",
			input:   " \n\t",
			want:    tk.IRDoc(),
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
				tk.IRIndentedCodeBlock("    header"),
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
				tk.IRIndentedCodeBlock("    ---"),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break rejected, tabs in leading whitespace",
			input: "\t---",
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock("\t---"),
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
				tk.IRIndentedCodeBlock("    > text"),
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
		{
			name:  "ul: single item, single line",
			input: "- a",
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ul: rejects missing delimiter",
			input: "-a",
			want: tk.IRDoc(
				tk.IRPara("-a"),
			),
			wantErr: nil,
		},
		{
			name:  "ul: accepts '*' markers",
			input: "* a",
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ul: accepts '+' markers",
			input: "+ a",
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ul: accepts 0-3 indentation at scope",
			input: "   - a",
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ul: rejects 4+ indentation at scope",
			input: "    - a",
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock("    - a"),
			),
			wantErr: nil,
		},
		{
			name: "ul: two sibling items",
			input: strings.Join([]string{
				"- a",
				"- b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
					),
					tk.IRListItem(
						tk.IRPara("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ul: sibling items may mix markers",
			input: strings.Join([]string{
				"- a",
				"* b",
				"+ c",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
					),
					tk.IRListItem(
						tk.IRPara("b"),
					),
					tk.IRListItem(
						tk.IRPara("c"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ul: list terminates on non-item line at same indent",
			input: strings.Join([]string{
				"- a",
				"x",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
				tk.IRPara("x"),
			),
			wantErr: nil,
		},
		{
			name:  "ul: item content is parsed as child blocks",
			input: "- # h",
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRHeader(1, "h"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ul: continuation line indented equal to content baseline accepted",
			input: strings.Join([]string{
				"- a",
				"  b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a", "b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ul: continuation line dedented to content baseline terminates list",
			input: strings.Join([]string{
				"- a",
				" b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
				tk.IRPara(" b"),
			),
			wantErr: nil,
		},
		{
			name: "ul: continuation line indented greater than content baseline accepted",
			input: strings.Join([]string{
				"- a",
				"    b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a", "  b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ul: nested list via indentation",
			input: strings.Join([]string{
				"- a",
				"  - b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
						tk.IRUnorderedList(
							true,
							tk.IRListItem(
								tk.IRPara("b"),
							),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ul: sibling item after nested list",
			input: strings.Join([]string{
				"- a",
				"  - b",
				"- c",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
						tk.IRUnorderedList(
							true,
							tk.IRListItem(
								tk.IRPara("b"),
							),
						),
					),
					tk.IRListItem(
						tk.IRPara("c"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ul: blank line inside items separates paragraphs",
			input: strings.Join([]string{
				"- a",
				"",
				"  b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					false,
					tk.IRListItem(
						tk.IRPara("a"),
						tk.IRPara("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ul: trailing blank not followed by continuation rolls back",
			input: strings.Join([]string{
				"- a",
				"",
				"x",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
				tk.IRPara("x"),
			),
			wantErr: nil,
		},
		{
			name: "ul: blank line between sibling items ends list",
			input: strings.Join([]string{
				"- a",
				"",
				"- b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					false,
					tk.IRListItem(
						tk.IRPara("a"),
					),
					tk.IRListItem(
						tk.IRPara("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ol: single item, single line, dot delimiter",
			input: "1. a",
			want: tk.IRDoc(
				tk.IROrderedList(
					true,
					1,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ol: single item, single line, paren delimiter",
			input: "1) a",
			want: tk.IRDoc(
				tk.IROrderedList(
					true,
					1,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ol: rejects missing delimiter after punctuation",
			input: "1.a",
			want: tk.IRDoc(
				tk.IRPara("1.a"),
			),
			wantErr: nil,
		},
		{
			name:  "ol: rejects missing digits",
			input: ".a",
			want: tk.IRDoc(
				tk.IRPara(".a"),
			),
			wantErr: nil,
		},
		{
			name:  "ol: two sibling items",
			input: "1. a\n2. b",
			want: tk.IRDoc(
				tk.IROrderedList(
					true,
					1,
					tk.IRListItem(
						tk.IRPara("a"),
					),
					tk.IRListItem(
						tk.IRPara("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ol: numbering need not be sequential",
			input: "1. a\n7. b\n42. c",
			want: tk.IRDoc(
				tk.IROrderedList(
					true,
					1,
					tk.IRListItem(
						tk.IRPara("a"),
					),
					tk.IRListItem(
						tk.IRPara("b"),
					),
					tk.IRListItem(
						tk.IRPara("c"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ol: continuation line indented past content baseline",
			input: strings.Join([]string{
				"1. a",
				"       b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IROrderedList(
					true,
					1,
					tk.IRListItem(
						tk.IRPara("a", "    b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ol: line below content baseline does not continue item",
			input: strings.Join([]string{
				"1. a",
				" b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IROrderedList(
					true,
					1,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
				tk.IRPara(" b"),
			),
			wantErr: nil,
		},
		{
			name: "ol: nested ordered list",
			input: strings.Join([]string{
				"1. a",
				"   1. a",
				"2. b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IROrderedList(
					true,
					1,
					tk.IRListItem(
						tk.IRPara("a"),
						tk.IROrderedList(
							true,
							1,
							tk.IRListItem(
								tk.IRPara("a"),
							),
						),
					),
					tk.IRListItem(
						tk.IRPara("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ol: ordered item contains nested unordered list",
			input: strings.Join([]string{
				"1. a",
				"   - a",
				"2. b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IROrderedList(
					true,
					1,
					tk.IRListItem(
						tk.IRPara("a"),
						tk.IRUnorderedList(
							true,
							tk.IRListItem(
								tk.IRPara("a"),
							),
						),
					),
					tk.IRListItem(
						tk.IRPara("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ol: unordered item contains nested ordered list",
			input: strings.Join([]string{
				"- a",
				"  1. a",
				"- b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
						tk.IROrderedList(
							true,
							1,
							tk.IRListItem(
								tk.IRPara("a"),
							),
						),
					),
					tk.IRListItem(
						tk.IRPara("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ol: loose list via blank line between items",
			input: strings.Join([]string{
				"1. a",
				"",
				"2. b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IROrderedList(
					false,
					1,
					tk.IRListItem(
						tk.IRPara("a"),
					),
					tk.IRListItem(
						tk.IRPara("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ol: trailing blank not followed by sibling rolls back",
			input: strings.Join([]string{
				"1. a",
				"",
				"x",
			}, "\n"),
			want: tk.IRDoc(
				tk.IROrderedList(
					true,
					1,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
				tk.IRPara("x"),
			),
			wantErr: nil,
		},
		{
			name: "ol: loose list via blank line inside item",
			input: strings.Join([]string{
				"1. a",
				"",
				"   x",
			}, "\n"),
			want: tk.IRDoc(
				tk.IROrderedList(
					false,
					1,
					tk.IRListItem(
						tk.IRPara("a"),
						tk.IRPara("x"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "indented code block: single line",
			input: `	fmt.Println("hello")`,
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock(
					`	fmt.Println("hello")`,
				),
			),
			wantErr: nil,
		},
		{
			name: "indented code block: multiple lines",
			input: strings.Join([]string{
				`	func(main) {`,
				`		fmt.Println("hello")`,
				`	}`,
			}, "\n"),
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock(
					`	func(main) {`,
					`		fmt.Println("hello")`,
					`	}`,
				),
			),
			wantErr: nil,
		},
		{
			name: "indented code block: indentation is preserved",
			input: strings.Join([]string{
				"		code",
				"	block",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock(
					"		code",
					"	block",
				),
			),
			wantErr: nil,
		},
		{
			name: "indented code block: blank lines are allowed",
			input: strings.Join([]string{
				"	code",
				"",
				"",
				"	block",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock(
					"	code",
					"",
					"",
					"	block",
				),
			),
			wantErr: nil,
		},
		{
			name: "indented code block: terminates on non-blank line indented less than 4 visual columns",
			input: strings.Join([]string{
				"	line 1",
				"	line 2",
				"end",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock(
					"	line 1",
					"	line 2",
				),
				tk.IRPara("end"),
			),
			wantErr: nil,
		},
		{
			name: "indented code block: excludes trailing blank lines",
			input: strings.Join([]string{
				"	line 1",
				"",
				"",
				"end",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock(
					"	line 1",
				),
				tk.IRPara("end"),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: backtick, single line",
			input: strings.Join([]string{
				"```",
				"code",
				"```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: tilde, single line",
			input: strings.Join([]string{
				"~~~",
				"code",
				"~~~",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: multiple lines",
			input: strings.Join([]string{
				"```",
				"code 1",
				"code 2",
				"```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code 1",
					"code 2",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: no payload lines",
			input: strings.Join([]string{
				"```",
				"```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: rejects only two backticks",
			input: strings.Join([]string{
				"``",
				"code",
				"``",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRPara("``", "code", "``"),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: rejects only two tildes",
			input: strings.Join([]string{
				"~~",
				"code",
				"~~",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRPara("~~", "code", "~~"),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: rejects mismatched closing marker",
			input: strings.Join([]string{
				"```",
				"code",
				"~~~",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code",
					"~~~",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: accepts closing run longer than opening",
			input: strings.Join([]string{
				"```",
				"code",
				"`````",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: rejects closing run shorter than opening",
			input: strings.Join([]string{
				"`````",
				"code",
				"```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code",
					"```",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: single word info string",
			input: strings.Join([]string{
				"```go",
				"code",
				"```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: multiple word info string",
			input: strings.Join([]string{
				"```go linenos",
				"code",
				"```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: delimiter before info string",
			input: strings.Join([]string{
				"```     go",
				"code",
				"```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: backtick marker rejects backtick in info string",
			input: strings.Join([]string{
				"```go`bad",
				"code",
				"```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRPara("```go`bad", "code"),
				tk.IRFencedCodeBlock(0),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: tilde marker accepts backtick in info string",
			input: strings.Join([]string{
				"~~~go~ok",
				"code",
				"~~~",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: accepts up to 3 indented spaces",
			input: strings.Join([]string{
				"   ```",
				"code",
				"   ```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					3,
					"code",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: rejects more than 3 indented spaces",
			input: strings.Join([]string{
				"    ```",
				"code",
				"    ```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock(
					"    ```",
				),
				tk.IRPara(
					"code",
					"    ```",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: closing fence accepts up to 3 indented spaces",
			input: strings.Join([]string{
				"```",
				"code",
				"   ```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: closing fence rejects more than 3 indented spaces",
			input: strings.Join([]string{
				"```",
				"code",
				"    ```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code",
					"    ```",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: payload preserves blank lines",
			input: strings.Join([]string{
				"```",
				"code 1",
				"",
				"code 2",
				"```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code 1",
					"",
					"code 2",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: runs to EOF when no closer",
			input: strings.Join([]string{
				"```",
				"code 1",
				"code 2",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code 1",
					"code 2",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: accepts closer with trailing whitespace",
			input: strings.Join([]string{
				"```",
				"code",
				"```   ",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code",
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block: rejects closer with trailing characters (non-whitespace)",
			input: strings.Join([]string{
				"```",
				"code",
				"```bad",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRFencedCodeBlock(
					0,
					"code",
					"```bad",
				),
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
