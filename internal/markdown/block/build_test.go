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

func TestBuild(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    ir.Document
		wantErr error
	}{
		// Empty input and paragraph formation

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
			name:  "single paragraph: one line",
			input: "a",
			want: tk.IRDoc(
				tk.IRPara("a"),
			),
			wantErr: nil,
		},
		{
			name:  "single paragraph: multiple lines",
			input: "a\nb\nc",
			want: tk.IRDoc(
				tk.IRPara("a", "b", "c"),
			),
			wantErr: nil,
		},
		{
			name:  "leading blank lines: ignored",
			input: "\n\na",
			want: tk.IRDoc(
				tk.IRPara("a"),
			),
			wantErr: nil,
		},
		{
			name:  "trailing blank lines: ignored",
			input: "a\n\n",
			want: tk.IRDoc(
				tk.IRPara("a"),
			),
			wantErr: nil,
		},
		{
			name:  "two paragraphs: separated by one blank line",
			input: "a\n\nb",
			want: tk.IRDoc(
				tk.IRPara("a"),
				tk.IRPara("b"),
			),
			wantErr: nil,
		},
		{
			name:  "two paragraphs: separated by two blank lines",
			input: "a\n\n\nb",
			want: tk.IRDoc(
				tk.IRPara("a"),
				tk.IRPara("b"),
			),
			wantErr: nil,
		},
		{
			name:  "two paragraphs: separated by whitespace only line",
			input: "a\n \nb",
			want: tk.IRDoc(
				tk.IRPara("a"),
				tk.IRPara("b"),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: stops before header without blank line",
			input: "a\n# h",
			want: tk.IRDoc(
				tk.IRPara("a"),
				tk.IRHeader(1, "h"),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: interrupted by thematic break",
			input: "a\n---",
			want: tk.IRDoc(
				tk.IRHeader(2, "a"),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: interrupted by unordered list",
			input: "a\n- b",
			want: tk.IRDoc(
				tk.IRPara("a"),
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "paragraph: interrupted by fenced code block",
			input: strings.Join([]string{
				"a",
				"```",
				"code",
				"```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRPara("a"),
				tk.IRFencedCodeBlock(
					0,
					"code",
				),
			),
			wantErr: nil,
		},
		{
			name: "paragraph: interrupted by HTML block",
			input: strings.Join([]string{
				"a",
				"<!--",
				"b",
				"-->",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRPara("a"),
				tk.IRHTMLBlock("<!--", "b", "-->"),
			),
			wantErr: nil,
		},

		// ATX headers

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
			name:  "header level 1: 3 leading spaces (max)",
			input: "   # header",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: tab delimiter",
			input: "#\theader",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: consumes multiple spaces",
			input: "#     header",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: consumes multiple tabs",
			input: "#\t\t\theader",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: trailing whitespace trimmed",
			input: "# header     ",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: mixed whitespace trimmed",
			input: "# \t header \t ",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: empty header allowed",
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
			name:  "header rejected: no marker",
			input: "header",
			want: tk.IRDoc(
				tk.IRPara("header"),
			),
			wantErr: nil,
		},
		{
			name:  "header rejected: too many leading spaces",
			input: "    header",
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock("    header"),
			),
			wantErr: nil,
		},
		{
			name:  "header rejected: missing delimiter",
			input: "#header",
			want: tk.IRDoc(
				tk.IRPara("#header"),
			),
			wantErr: nil,
		},
		{
			name:  "header rejected: too many hashes",
			input: "####### header",
			want: tk.IRDoc(
				tk.IRPara("####### header"),
			),
			wantErr: nil,
		},
		{
			name:  "header rejected: too many hashes after indent",
			input: "   ####### header",
			want: tk.IRDoc(
				tk.IRPara("   ####### header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: empty header at end of line allowed",
			input: "#",
			want: tk.IRDoc(
				tk.IRHeader(1, ""),
			),
			wantErr: nil,
		},
		{
			name:  "header level 2: empty header at end of line allowed",
			input: "##",
			want: tk.IRDoc(
				tk.IRHeader(2, ""),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: closing marker run trimmed",
			input: "# header #",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: multiple closing markers trimmed",
			input: "# header ###",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 3: closing marker run trimmed with internal separation",
			input: "###   bar    ###",
			want: tk.IRDoc(
				tk.IRHeader(3, "bar"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: closing marker run allows trailing whitespace",
			input: "# header ###   \t ",
			want: tk.IRDoc(
				tk.IRHeader(1, "header"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: closing marker rejected without separating whitespace",
			input: "# header###",
			want: tk.IRDoc(
				tk.IRHeader(1, "header###"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: single trailing hash rejected without separating whitespace",
			input: "# header#",
			want: tk.IRDoc(
				tk.IRHeader(1, "header#"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: escaped closing marker not trimmed",
			input: "# header \\###",
			want: tk.IRDoc(
				tk.IRHeader(1, "header \\###"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: partially escaped trailing hashes not trimmed as closer",
			input: "# header #\\##",
			want: tk.IRDoc(
				tk.IRHeader(1, "header #\\##"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: even backslashes leave closing marker unescaped and trimmed",
			input: "# header \\\\###",
			want: tk.IRDoc(
				tk.IRHeader(1, "header \\\\"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: odd backslashes escape first trailing hash and prevent closer",
			input: "# header \\\\\\###",
			want: tk.IRDoc(
				tk.IRHeader(1, "header \\\\\\###"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 1: closing marker trims to empty content",
			input: "# ###",
			want: tk.IRDoc(
				tk.IRHeader(1, ""),
			),
			wantErr: nil,
		},
		{
			name:  "header level 3: content may itself contain hashes before valid closer",
			input: "### foo # bar ###",
			want: tk.IRDoc(
				tk.IRHeader(3, "foo # bar"),
			),
			wantErr: nil,
		},
		{
			name:  "header level 3: nonterminal hashes remain content",
			input: "### foo ### bar",
			want: tk.IRDoc(
				tk.IRHeader(3, "foo ### bar"),
			),
			wantErr: nil,
		},

		// Thematic breaks

		{
			name:  "thematic break: ---",
			input: "---",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break: ***",
			input: "***",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break: ___",
			input: "___",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break: leading spaces",
			input: "   ---",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break: inter-marker whitespace",
			input: "- \t - \t -",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break: trailing whitespace",
			input: "---   ",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break: more than three identical markers",
			input: "-----------------------",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break rejected: too many leading spaces",
			input: "    ---",
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock("    ---"),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break rejected: tabs in leading whitespace",
			input: "\t---",
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock("\t---"),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break rejected: mixed marker characters",
			input: "-*-",
			want: tk.IRDoc(
				tk.IRPara("-*-"),
			),
			wantErr: nil,
		},

		// Block quotes

		{
			name:  "block quote: plain text",
			input: "> text",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("text"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: leading spaces",
			input: "   > text",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("text"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: tab delimiter",
			input: ">\ttext",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("text"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: no delimiter",
			input: ">text",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("text"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: multiple lines",
			input: "> a\n> b",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("a", "b"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: containing header",
			input: "> # header",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRHeader(1, "header"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote rejected: too many leading spaces",
			input: "    > text",
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock("    > text"),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: blank line splits paragraphs",
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
			name:  "block quote: starts with a blank line",
			input: ">\n> a",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("a"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: only blank lines",
			input: ">\n>\n",
			want: tk.IRDoc(
				tk.IRBlockQuote(),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: mixed indentation across lines",
			input: "> a\n > b\n  > c",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("a", "b", "c"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: terminates on first non-quote line",
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
			name:  "block quote: terminates on truly blank line",
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
			name:  "block quote: terminates on dedented non-quote line",
			input: "> a\n b",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRPara("a"),
				),
				tk.IRPara(" b"),
			),
			wantErr: nil,
		},
		{
			name:  "nested quotes: via >>",
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
			name:  "nested quotes: via > >",
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
			name:  "nested quotes: via >\t>",
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
			name:  "block quote: contains unordered list",
			input: "> - a\n> - b",
			want: tk.IRDoc(
				tk.IRBlockQuote(
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
			),
			wantErr: nil,
		},
		{
			name:  "block quote: contains indented code block",
			input: ">     code",
			want: tk.IRDoc(
				tk.IRBlockQuote(
					tk.IRIndentedCodeBlock("    code"),
				),
			),
			wantErr: nil,
		},

		// Setext headings

		{
			name:  "setext: h1 (minimum)",
			input: "heading\n=",
			want: tk.IRDoc(
				tk.IRHeader(1, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext: h1 (typical)",
			input: "heading\n===",
			want: tk.IRDoc(
				tk.IRHeader(1, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext: h2 (minimum)",
			input: "heading\n-",
			want: tk.IRDoc(
				tk.IRHeader(2, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext: h2 (typical)",
			input: "heading\n---",
			want: tk.IRDoc(
				tk.IRHeader(2, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext: heading and paragraph",
			input: "heading\n---\nnext",
			want: tk.IRDoc(
				tk.IRHeader(2, "heading"),
				tk.IRPara("next"),
			),
			wantErr: nil,
		},
		{
			name:  "setext: trailing spaces",
			input: "heading\n===   ",
			want: tk.IRDoc(
				tk.IRHeader(1, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext: trailing tabs",
			input: "heading\n===\t\t",
			want: tk.IRDoc(
				tk.IRHeader(1, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext: mixed trailing spaces and tabs",
			input: "heading\n===\t \t ",
			want: tk.IRDoc(
				tk.IRHeader(1, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext: leading spaces",
			input: "heading\n   ---",
			want: tk.IRDoc(
				tk.IRHeader(2, "heading"),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected: too many leading spaces",
			input: "heading\n    ---",
			want: tk.IRDoc(
				tk.IRPara("heading", "    ---"),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected: underline with internal spaces (dash)",
			input: "heading\n- - -",
			want: tk.IRDoc(
				tk.IRPara("heading"),
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected: underline with internal spaces (equals)",
			input: "heading\n= = =",
			want: tk.IRDoc(
				tk.IRPara("heading", "= = ="),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected: underline with non-marker character",
			input: "heading\n--x--",
			want: tk.IRDoc(
				tk.IRPara("heading", "--x--"),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected: underline with mixed markers",
			input: "heading\n-=-",
			want: tk.IRDoc(
				tk.IRPara("heading", "-=-"),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected: underline with trailing non-space",
			input: "heading\n---x",
			want: tk.IRDoc(
				tk.IRPara("heading", "---x"),
			),
			wantErr: nil,
		},
		{
			name:  "setext: h2 takes precedence over thematic break for ---",
			input: "heading\n---\nnext",
			want: tk.IRDoc(
				tk.IRHeader(2, "heading"),
				tk.IRPara("next"),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break: - - - does not become setext",
			input: "heading\n- - -\nnext",
			want: tk.IRDoc(
				tk.IRPara("heading"),
				tk.IRThematicBreak(),
				tk.IRPara("next"),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break: *** does not become setext",
			input: "heading\n***\nnext",
			want: tk.IRDoc(
				tk.IRPara("heading"),
				tk.IRThematicBreak(),
				tk.IRPara("next"),
			),
			wantErr: nil,
		},
		{
			name:  "setext: multiline content",
			input: "line1\nline2\n---",
			want: tk.IRDoc(
				tk.IRHeader(2, "line1", "line2"),
			),
			wantErr: nil,
		},
		{
			name:  "setext: multiline content stops before underline",
			input: "line1\nline2\n===\nnext",
			want: tk.IRDoc(
				tk.IRHeader(1, "line1", "line2"),
				tk.IRPara("next"),
			),
			wantErr: nil,
		},
		{
			name:  "setext rejected: blank line between content and underline",
			input: "heading\n\n---",
			want: tk.IRDoc(
				tk.IRPara("heading"),
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "setext: underline followed by blank line",
			input: "heading\n---\n\nnext",
			want: tk.IRDoc(
				tk.IRHeader(2, "heading"),
				tk.IRPara("next"),
			),
			wantErr: nil,
		},
		{
			name:  "setext: underline at start of doc is not a heading (dashes)",
			input: "---",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "setext: underline at start of doc is not a heading (equals)",
			input: "===",
			want: tk.IRDoc(
				tk.IRPara("==="),
			),
			wantErr: nil,
		},

		// Unordered lists

		{
			name:  "ul: single item: single line",
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
			name:  "ul: accepts * markers",
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
			name:  "ul: accepts + markers",
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
			name: "ul: mixed ordered list at same indent terminates unordered list",
			input: strings.Join([]string{
				"- a",
				"1. b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
				tk.IROrderedList(
					true,
					1,
					tk.IRListItem(
						tk.IRPara("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ul: empty list item (bare marker) is rejected",
			input: "-",
			want: tk.IRDoc(
				tk.IRPara("-"),
			),
			wantErr: nil,
		},
		{
			name:  "ul: empty list item (marker and space) produces empty paragraph child",
			input: "- ",
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(),
				),
			),
			wantErr: nil,
		},
		{
			name: "ul: deep nesting",
			input: strings.Join([]string{
				"- a",
				"  - b",
				"    - c",
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
								tk.IRUnorderedList(
									true,
									tk.IRListItem(
										tk.IRPara("c"),
									),
								),
							),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ul: item with thematic break child",
			input: "- ---",
			want: tk.IRDoc(
				tk.IRThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name: "ul: item with fenced code block child",
			input: strings.Join([]string{
				"- ```",
				"  code",
				"  ```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRFencedCodeBlock(
							0,
							"code",
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ul: item with HTML block child",
			input: strings.Join([]string{
				"- <div>",
				"  html",
				"",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRHTMLBlock("<div>", "html"),
					),
				),
			),
			wantErr: nil,
		},

		// Ordered lists

		{
			name:  "ol: single item: single line: dot delimiter",
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
			name:  "ol: single item: single line: paren delimiter",
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
			name: "ol: mixed unordered list at same indent terminates ordered list",
			input: strings.Join([]string{
				"1. a",
				"- b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IROrderedList(
					true,
					1,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "ol: mixed delimiter punctuation terminates list",
			input: strings.Join([]string{
				"1. a",
				"2) b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IROrderedList(
					true,
					1,
					tk.IRListItem(
						tk.IRPara("a"),
					),
				),
				tk.IROrderedList(
					true,
					2,
					tk.IRListItem(
						tk.IRPara("b"),
					),
				),
			),
			wantErr: nil,
		},

		// Indented code blocks

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
			name:  "indented code block: exactly three spaces is paragraph",
			input: "   a",
			want: tk.IRDoc(
				tk.IRPara("   a"),
			),
			wantErr: nil,
		},
		{
			name:  "indented code block: exactly four spaces is code block",
			input: "    a",
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock("    a"),
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
			name: "indented code block: does not interrupt paragraph inside list item",
			input: strings.Join([]string{
				"- a",
				"      code",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a", "    code"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "indented code block: inside list item at baseline plus four columns",
			input: strings.Join([]string{
				"- a",
				"",
				"      code",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					false,
					tk.IRListItem(
						tk.IRPara("a"),
						tk.IRIndentedCodeBlock("    code"),
					),
				),
			),
			wantErr: nil,
		},

		// Fenced code blocks

		{
			name: "fenced code block: backtick: single line",
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
			name: "fenced code block: tilde: single line",
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
		{
			name: "fenced code block: inside list item",
			input: strings.Join([]string{
				"- a",
				"  ```",
				"  code",
				"  ```",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRUnorderedList(
					true,
					tk.IRListItem(
						tk.IRPara("a"),
						tk.IRFencedCodeBlock(
							0,
							"code",
						),
					),
				),
			),
			wantErr: nil,
		},

		// HTML blocks

		{
			name:  "html block: comment",
			input: "<!-- comment -->",
			want: tk.IRDoc(
				tk.IRHTMLBlock("<!-- comment -->"),
			),
			wantErr: nil,
		},
		{
			name: "html block: comment: multiple lines",
			input: strings.Join([]string{
				"<!--",
				"comment",
				"-->",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<!--", "comment", "-->"),
			),
			wantErr: nil,
		},
		{
			name: "html block: comment: blank lines permitted",
			input: strings.Join([]string{
				"<!--",
				"line 1",
				"",
				"line 3",
				"-->",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<!--", "line 1", "", "line 3", "-->"),
			),
			wantErr: nil,
		},
		{
			name: "html block: comment: consumes to EOF when unterminated",
			input: strings.Join([]string{
				"<!--",
				"line 1",
				"line 2",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<!--", "line 1", "line 2"),
			),
			wantErr: nil,
		},
		{
			name: "html block: comment: whole terminating line absorbed",
			input: strings.Join([]string{
				"<!--",
				"hello --> trailing",
				"next",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<!--", "hello --> trailing"),
				tk.IRPara("next"),
			),
			wantErr: nil,
		},
		{
			name:  "html block: comment: accepts leading spaces",
			input: "   <!-- comment -->",
			want: tk.IRDoc(
				tk.IRHTMLBlock("   <!-- comment -->"),
			),
			wantErr: nil,
		},
		{
			name:  "html block: comment: rejects 4+ leading spaces",
			input: "    <!-- not comment -->",
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock("    <!-- not comment -->"),
			),
			wantErr: nil,
		},
		{
			name:  "html block: rejects opener with wrong spacing",
			input: "< !-- not comment -->",
			want: tk.IRDoc(
				tk.IRPara("< !-- not comment -->"),
			),
			wantErr: nil,
		},
		{
			name:  "html block: processing instructions",
			input: `<?xml version="1.0"?>`,
			want: tk.IRDoc(
				tk.IRHTMLBlock(`<?xml version="1.0"?>`),
			),
			wantErr: nil,
		},
		{
			name:  "html block: declarations",
			input: "<!DOCTYPE html>",
			want: tk.IRDoc(
				tk.IRHTMLBlock("<!DOCTYPE html>"),
			),
			wantErr: nil,
		},
		{
			name:  "html block: cdata",
			input: "<![CDATA[hello]]>",
			want: tk.IRDoc(
				tk.IRHTMLBlock("<![CDATA[hello]]>"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: simple opening tag",
			input: strings.Join([]string{
				"<div>",
				"hello",
				"</div>",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<div>", "hello", "</div>"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: opening tag with attributes",
			input: strings.Join([]string{
				`<div class="note">`,
				"hello",
				"</div>",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock(`<div class="note">`, "hello", "</div>"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: uppercase tag name normalized",
			input: strings.Join([]string{
				"<DIV>",
				"hello",
				"</DIV>",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<div>", "hello", "</div>"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: mixed case tag name normalized",
			input: strings.Join([]string{
				`<Section class="callout">`,
				"hello",
				"</Section>",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock(`<section class="callout">`, "hello", "</section>"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: closing tag line as opener",
			input: strings.Join([]string{
				"</div>",
				"tail",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("</div>", "tail"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: self closing tag",
			input: strings.Join([]string{
				"<hr />",
				"tail",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<hr />", "tail"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: self closing without space",
			input: strings.Join([]string{
				"<hr/>",
				"tail",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<hr/>", "tail"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: terminates before first blank line",
			input: strings.Join([]string{
				"<div>",
				"hello",
				"",
				"world",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<div>", "hello"),
				tk.IRPara("world"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: single opening line followed by blank line",
			input: strings.Join([]string{
				"<div>",
				"",
				"world",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<div>"),
				tk.IRPara("world"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: consumes through non blank lines",
			input: strings.Join([]string{
				"<div>",
				"line 1",
				"line 2",
				"line 3",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<div>", "line 1", "line 2", "line 3"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: closing tag does not itself terminate",
			input: strings.Join([]string{
				"<div>",
				"line 1",
				"</div>",
				"line 2",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<div>", "line 1", "</div>", "line 2"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: blank line after closing tag controls termination",
			input: strings.Join([]string{
				"<div>",
				"hello",
				"</div>",
				"",
				"tail",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<div>", "hello", "</div>"),
				tk.IRPara("tail"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: followed by content without blank line remains in block",
			input: strings.Join([]string{
				"<div>",
				"a",
				"</div>",
				"b",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRHTMLBlock("<div>", "a", "</div>", "b"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: rejects inline span tag",
			input: strings.Join([]string{
				"<span>",
				"hello",
				"</span>",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRPara("<span>", "hello", "</span>"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: rejects custom element",
			input: strings.Join([]string{
				"<custom-element>",
				"hello",
				"</custom-element>",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRPara("<custom-element>", "hello", "</custom-element>"),
			),
			wantErr: nil,
		},
		{
			name:  "html block: named tag: rejects missing name tag",
			input: "<>",
			want: tk.IRDoc(
				tk.IRPara("<>"),
			),
			wantErr: nil,
		},
		{
			name:  "html block: named tag: rejects closing slash with no name",
			input: "</>",
			want: tk.IRDoc(
				tk.IRPara("</>"),
			),
			wantErr: nil,
		},
		{
			name:  "html block: named tag: reject numeric tag start",
			input: "<1div>",
			want: tk.IRDoc(
				tk.IRPara("<1div>"),
			),
			wantErr: nil,
		},
		{
			name:  "html block: named tag: rejects space after opener",
			input: "< div>",
			want: tk.IRDoc(
				tk.IRPara("< div>"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: rejects incomplete opener tag",
			input: strings.Join([]string{
				"<div",
				"hello",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRPara("<div", "hello"),
			),
			wantErr: nil,
		},
		{
			name:  "html block: named tag: rejects self closing garbage",
			input: "<hr/garbage>",
			want: tk.IRDoc(
				tk.IRPara("<hr/garbage>"),
			),
			wantErr: nil,
		},
		{
			name:  "html block: named tag: rejects invalid character after tag name",
			input: "<div-foo>",
			want: tk.IRDoc(
				tk.IRPara("<div-foo>"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: rejects punctuation pseudo alpha",
			input: strings.Join([]string{
				"<`div>",
				"hello",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRPara("<`div>", "hello"),
			),
			wantErr: nil,
		},
		{
			name: "html block: interrupts paragraph consumption",
			input: strings.Join([]string{
				"line 1",
				"<!--",
				"line 2",
				"-->",
				"line 3",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRPara("line 1"),
				tk.IRHTMLBlock("<!--", "line 2", "-->"),
				tk.IRPara("line 3"),
			),
			wantErr: nil,
		},
		{
			name: "html block: named tag: interrupts paragraph consumption",
			input: strings.Join([]string{
				"line 1",
				"<div>",
				"line 2",
				"</div>",
				"line 3",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRPara("line 1"),
				tk.IRHTMLBlock("<div>", "line 2", "</div>", "line 3"),
			),
			wantErr: nil,
		},

		// Reference link definitions

		{
			name:  "reference definition: bare destination, no title",
			input: "[foo]: /url",
			want: ir.Document{
				Definitions: map[string]ir.ReferenceDefinition{
					"foo": tk.IRRefDef("foo", false),
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference definition: angle destination, no title",
			input: "[foo]: <https://example.com>",
			want: ir.Document{
				Definitions: map[string]ir.ReferenceDefinition{
					"foo": tk.IRRefDef("foo", false),
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference definition: quoted title with double quotes",
			input: `[foo]: /url "title"`,
			want: ir.Document{
				Definitions: map[string]ir.ReferenceDefinition{
					"foo": tk.IRRefDef("foo", true),
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference definition: quoted title with single quotes",
			input: "[foo]: /url 'title'",
			want: ir.Document{
				Definitions: map[string]ir.ReferenceDefinition{
					"foo": tk.IRRefDef("foo", true),
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference definition: paren title",
			input: "[foo]: /url (title)",
			want: ir.Document{
				Definitions: map[string]ir.ReferenceDefinition{
					"foo": tk.IRRefDef("foo", true),
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference definition: up to three spaces indentation allowed",
			input: "   [foo]: /url",
			want: ir.Document{
				Definitions: map[string]ir.ReferenceDefinition{
					"foo": tk.IRRefDef("foo", false),
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference definition: four spaces indentation rejected",
			input: "    [foo]: /url",
			want: tk.IRDoc(
				tk.IRIndentedCodeBlock("[foo]: /url"),
			),
			wantErr: nil,
		},
		{
			name:  "reference definition: label whitespace normalizes in key",
			input: "[  Foo  Bar  ]: /url",
			want: ir.Document{
				Definitions: map[string]ir.ReferenceDefinition{
					"foo bar": tk.IRRefDef("foo bar", false),
				},
			},
			wantErr: nil,
		},
		{
			name: "reference definition: duplicate key accumulates only one definition",
			input: strings.Join([]string{
				"[Foo]: /one",
				"[ foo ]: /two",
			}, "\n"),
			want: ir.Document{
				Definitions: map[string]ir.ReferenceDefinition{
					"foo": tk.IRRefDef("foo", false),
				},
			},
			wantErr: nil,
		},
		{
			name: "reference definition: multiple definitions accumulate",
			input: strings.Join([]string{
				"[foo]: /one",
				`[bar]: /two "title"`,
				"[baz qux]: <https://example.com>",
			}, "\n"),
			want: ir.Document{
				Definitions: map[string]ir.ReferenceDefinition{
					"foo":     tk.IRRefDef("foo", false),
					"bar":     tk.IRRefDef("bar", true),
					"baz qux": tk.IRRefDef("baz qux", false),
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference definition: escaped opening bracket normalizes without backslash",
			input: `[foo\[]: /url`,
			want: ir.Document{
				Definitions: map[string]ir.ReferenceDefinition{
					"foo[": tk.IRRefDef("foo[", false),
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference definition: escaped closing bracket normalizes without backslash",
			input: `[foo\]]: /url`,
			want: ir.Document{
				Definitions: map[string]ir.ReferenceDefinition{
					"foo]": tk.IRRefDef("foo]", false),
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference definition: escaped whitespace in label is preserved",
			input: `[foo\ bar]: /url`,
			want: ir.Document{
				Definitions: map[string]ir.ReferenceDefinition{
					"foo bar": tk.IRRefDef("foo bar", false),
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference definition: collapsed whitespace before escaped whitespace is preserved separately",
			input: `[foo   \ bar]: /url`,
			want: ir.Document{
				Definitions: map[string]ir.ReferenceDefinition{
					"foo  bar": tk.IRRefDef("foo  bar", false),
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference definition: all-whitespace label is rejected",
			input: "[   ]: /url",
			want: tk.IRDoc(
				tk.IRPara("[   ]: /url"),
			),
			wantErr: nil,
		},
		{
			name:  "reference definition: missing colon is rejected",
			input: "[foo] /url",
			want: tk.IRDoc(
				tk.IRPara("[foo] /url"),
			),
			wantErr: nil,
		},
		{
			name:  "reference definition: missing destination is rejected",
			input: "[foo]:",
			want: tk.IRDoc(
				tk.IRPara("[foo]:"),
			),
			wantErr: nil,
		},
		{
			name:  "reference definition: trailing junk after destination is rejected",
			input: "[foo]: /url garbage",
			want: tk.IRDoc(
				tk.IRPara("[foo]: /url garbage"),
			),
			wantErr: nil,
		},
		{
			name:  "reference definition: title requires separating space",
			input: `[foo]: <url>"title"`,
			want: tk.IRDoc(
				tk.IRPara(`[foo]: <url>"title"`),
			),
			wantErr: nil,
		},
		{
			name:  "reference definition: trailing junk after title is rejected",
			input: `[foo]: /url "title" garbage`,
			want: tk.IRDoc(
				tk.IRPara(`[foo]: /url "title" garbage`),
			),
			wantErr: nil,
		},
		{
			name: "reference definition: cannot interrupt paragraph",
			input: strings.Join([]string{
				"Foo",
				"[bar]: /url",
				"",
				"[bar]",
			}, "\n"),
			want: tk.IRDoc(
				tk.IRPara("Foo", "[bar]: /url"),
				tk.IRPara("[bar]"),
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
