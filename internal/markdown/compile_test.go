package markdown_test

import (
	"strings"
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestCompile_EndToEnd(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		// paragraphs

		{
			name:    "paragraph: plain single line",
			input:   "hello world",
			want:    `<p>hello world</p>`,
			wantErr: nil,
		},
		{
			name: "paragraph: multiple lines join into one paragraph",
			input: md(
				"hello",
				"world",
			),
			want:    `<p>hello world</p>`,
			wantErr: nil,
		},
		{
			name: "paragraph: blank line separates paragraph",
			input: md(
				"hello",
				"",
				"world",
			),
			want:    `<p>hello</p><p>world</p>`,
			wantErr: nil,
		},
		{
			name: "paragaph: whitespace only line separates paragraph",
			input: md(
				"hello",
				"    ",
				"world",
			),
			want:    `<p>hello</p><p>world</p>`,
			wantErr: nil,
		},
		{
			name: "paragraph: leading blank lines are ignored",
			input: md(
				"",
				"",
				"hello world",
			),
			want:    `<p>hello world</p>`,
			wantErr: nil,
		},
		{
			name: "paragraph: trailing blank lines are ignored",
			input: md(
				"hello world",
				"",
				"",
			),
			want:    `<p>hello world</p>`,
			wantErr: nil,
		},
		{
			name: "paragraph: indented line continues paragraph",
			input: md(
				"hello",
				"    world",
			),
			want:    `<p>hello     world</p>`,
			wantErr: nil,
		},
		{
			name: "paragraph: ends before atx heading",
			input: md(
				"hello",
				"# heading",
			),
			want:    `<p>hello</p><h1>heading</h1>`,
			wantErr: nil,
		},
		{
			name: "paragraph: ends before block quote",
			input: md(
				"hello",
				"> world",
			),
			want:    `<p>hello</p><blockquote><p>world</p></blockquote>`,
			wantErr: nil,
		},
		{
			name: "paragraph: ends before unordered list",
			input: md(
				"hello",
				"- world",
			),
			want:    `<p>hello</p><ul><li>world</li></ul>`,
			wantErr: nil,
		},
		{
			name: "paragraph: ends before ordered list",
			input: md(
				"hello",
				"1. world",
			),
			want:    `<p>hello</p><ol><li>world</li></ol>`,
			wantErr: nil,
		},
		{
			name: "paragraph: hard break via two trailing spaces",
			input: md(
				"hello  ",
				"world",
			),
			want:    `<p>hello<br>world</p>`,
			wantErr: nil,
		},
		{
			name: "paragraph: hard break via trailing backslash",
			input: md(
				"hello\\",
				"world",
			),
			want:    `<p>hello<br>world</p>`,
			wantErr: nil,
		},

		// ATX headers

		{
			name:    "atx heading: level 1 plain text",
			input:   "# hello",
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: level 2 plain text",
			input:   "## hello",
			want:    `<h2>hello</h2>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: level 3 plain text",
			input:   "### hello",
			want:    `<h3>hello</h3>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: level 4 plain text",
			input:   "#### hello",
			want:    `<h4>hello</h4>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: level 5 plain text",
			input:   "##### hello",
			want:    `<h5>hello</h5>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: level 6 plain text",
			input:   "###### hello",
			want:    `<h6>hello</h6>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: empty content after required delimiter",
			input:   "# ",
			want:    `<h1></h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: empty content after multiple spaces",
			input:   "#   ",
			want:    `<h1></h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: empty context after tab delimiter",
			input:   "#\t",
			want:    `<h1></h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: level 1 heading missing delimiter",
			input:   "#",
			want:    `<p>#</p>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: level 6 heading missing delimiter",
			input:   "######",
			want:    `<p>######</p>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: multiple spaces after marker allowed",
			input:   "#   hello",
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: tab after marker allowed",
			input:   "#\thello",
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: spaces and tabs after marker are consumed before content",
			input:   "# \t   hello",
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: content may contain internal hash characters",
			input:   "# hello # world",
			want:    `<h1>hello # world</h1>`,
			wantErr: nil,
		},
		// NOTE: subject to change upon implementing trim ATX closers
		{
			name:    "atx heading: trailing hash characters remain content",
			input:   "# hello ###",
			want:    `<h1>hello ###</h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: content may consist only of hash characters",
			input:   "# ###",
			want:    `<h1>###</h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: trailing spaces are trimmed from content span",
			input:   "# hello   ",
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: trailing tabs are trimmed from content span",
			input:   "# hello\t\t",
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: trailing spaces after hash content are trimmed but hashes remain",
			input:   "# hello ###   ",
			want:    `<h1>hello ###</h1>`,
			wantErr: nil,
		},
		// NOTE: end trim ATX closers cases
		{
			name:    "atx heading: leading indentation of one space allowed",
			input:   " # hello",
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: leading indentation of two spaces allowed",
			input:   "  # hello",
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: leading indentation of three spaces allowed",
			input:   "   # hello",
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: leading indentation of four spaces is not a heading",
			input:   "    # hello",
			want:    `<pre><code># hello</code></pre>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: marker must be followed by space not plain text",
			input:   "#hello",
			want:    `<p>#hello</p>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: seven markers is not a heading",
			input:   "####### hello",
			want:    `<p>####### hello</p>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: non-hash opener is not a heading",
			input:   "hello",
			want:    `<p>hello</p>`,
			wantErr: nil,
		},
		{
			name: "atx heading: heading may follow blank line",
			input: md(
				"",
				"# hello",
			),
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name: "atx heading: consecutive headings of different levels",
			input: md(
				"# one",
				"## two",
				"### three",
			),
			want:    `<h1>one</h1><h2>two</h2><h3>three</h3>`,
			wantErr: nil,
		},
		{
			name:    "atx heading: escaped hash in content remains content",
			input:   "# \\# hello",
			want:    `<h1># hello</h1>`,
			wantErr: nil,
		},

		// Setext headers

		{
			name: "setext heading: level 1 plain text with equals underline",
			input: md(
				"hello",
				"=====",
			),
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name: "setext heading: level 2 plain text with dash underline",
			input: md(
				"hello",
				"-----",
			),
			want:    `<h2>hello</h2>`,
			wantErr: nil,
		},
		{
			name: "setext heading: single equals marker is valid underline",
			input: md(
				"hello",
				"=",
			),
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name: "setext heading: single dash marker is valid underline",
			input: md(
				"hello",
				"-",
			),
			want:    `<h2>hello</h2>`,
			wantErr: nil,
		},
		{
			name: "setext heading: underline may be longer than content",
			input: md(
				"hi",
				"==========",
			),
			want:    `<h1>hi</h1>`,
			wantErr: nil,
		},
		{
			name: "setext heading: dash underline may be longer than content",
			input: md(
				"hi",
				"----------",
			),
			want:    `<h2>hi</h2>`,
			wantErr: nil,
		},
		{
			name: "setext heading: leading indentation of one space on underline allowed",
			input: md(
				"hello",
				" =====",
			),
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name: "setext heading: leading indentation of two spaces on underline allowed",
			input: md(
				"hello",
				"  =====",
			),
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name: "setext heading: leading indentation of three spaces on underline allowed",
			input: md(
				"hello",
				"   =====",
			),
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name: "setext heading: leading indentation of four spaces on underline is not valid",
			input: md(
				"hello",
				"    =====",
			),
			want:    `<p>hello     =====</p>`,
			wantErr: nil,
		},
		{
			name: "setext heading: trailing spaces after equals underline allowed",
			input: md(
				"hello",
				"=====   ",
			),
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name: "setext heading: trailing tabs after equals underline allowed",
			input: md(
				"hello",
				"=====\t\t",
			),
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name: "setext heading: trailing mixed spaces and tabs after underline allowed",
			input: md(
				"hello",
				"===== \t \t",
			),
			want:    `<h1>hello</h1>`,
			wantErr: nil,
		},
		{
			name: "setext heading: content line may contain punctuation",
			input: md(
				"hello, world",
				"-----",
			),
			want:    `<h2>hello, world</h2>`,
			wantErr: nil,
		},
		{
			name: "setext heading: content line may contain hash characters",
			input: md(
				"hello # world",
				"-----",
			),
			want:    `<h2>hello # world</h2>`,
			wantErr: nil,
		},
		{
			name: "setext heading: content line may contain inline markers",
			input: md(
				"hello *world*",
				"-----",
			),
			want:    `<h2>hello <em>world</em></h2>`,
			wantErr: nil,
		},
		{
			name: "setext heading: content line may contain leading and trailing spaces",
			input: md(
				" hello world ",
				"-----",
			),
			want:    `<h2> hello world </h2>`,
			wantErr: nil,
		},
		{
			name: "setext heading: blank line between content and underline prevents heading",
			input: md(
				"hello",
				"",
				"-----",
			),
			want:    `<p>hello</p><hr>`,
			wantErr: nil,
		},
		{
			name: "setext heading: underline line alone at document start is not a heading",
			input: md(
				"-----",
			),
			want:    `<hr>`,
			wantErr: nil,
		},
		{
			name: "setext heading: equals underline alone at document start is not a heading",
			input: md(
				"=====",
			),
			want:    `<p>=====</p>`,
			wantErr: nil,
		},
		{
			name: "setext heading: mixed equals and dash markers are not valid",
			input: md(
				"hello",
				"=-=-=",
			),
			want:    `<p>hello =-=-=</p>`,
			wantErr: nil,
		},
		{
			name: "setext heading: mixed dash and equals markers are not valid",
			input: md(
				"hello",
				"-=-=-",
			),
			want:    `<p>hello -=-=-</p>`,
			wantErr: nil,
		},
		{
			name: "setext heading: internal spaces between markers are not valid",
			input: md(
				"hello",
				"= = =",
			),
			want:    `<p>hello = = =</p>`,
			wantErr: nil,
		},
		{
			name: "setext heading: internal tabs between markers are not valid",
			input: md(
				"hello",
				"=\t=\t=",
			),
			want:    `<p>hello =	=	=</p>`,
			wantErr: nil,
		},
		{
			name: "setext heading: trailing non-whitespace after equals markers is not valid",
			input: md(
				"hello",
				"=====x",
			),
			want:    `<p>hello =====x</p>`,
			wantErr: nil,
		},
		{
			name: "setext heading: trailing non-whitespace after dash markers is not valid",
			input: md(
				"hello",
				"-----x",
			),
			want:    `<p>hello -----x</p>`,
			wantErr: nil,
		},
		{
			name: "setext heading: leading non-marker text is not valid underline",
			input: md(
				"hello",
				"x----",
			),
			want:    `<p>hello x----</p>`,
			wantErr: nil,
		},
		{
			name: "setext heading: non-whitespace mixed into underline is not valid",
			input: md(
				"hello",
				"--x--",
			),
			want:    `<p>hello --x--</p>`,
			wantErr: nil,
		},
		{
			name: "setext heading: level 2 underline disambiguates from thematic break after paragraph text",
			input: md(
				"hello",
				"---",
			),
			want:    `<h2>hello</h2>`,
			wantErr: nil,
		},
		{
			name: "setext heading: heading text must immediately precede underline",
			input: md(
				"hello",
				"",
				"world",
				"-----",
			),
			want:    `<p>hello</p><h2>world</h2>`,
			wantErr: nil,
		},
		{
			name: "setext heading: multiline content normalizes newline to space",
			input: md(
				"hello",
				"world",
				"-----",
			),
			want:    `<h2>hello world</h2>`,
			wantErr: nil,
		},
		{
			name: "setext heading: multiline content preserves hard break via trailing backslash",
			input: md(
				"hello\\",
				"world",
				"-----",
			),
			want:    `<h2>hello<br>world</h2>`,
			wantErr: nil,
		},
		{
			name: "setext heading: multiline content preserves hard break via trailing spaces",
			input: md(
				"hello  ",
				"world",
				"-----",
			),
			want:    `<h2>hello<br>world</h2>`,
			wantErr: nil,
		},

		// thematic breaks

		{
			name:    "thematic break: three hyphens",
			input:   "---",
			want:    `<hr>`,
			wantErr: nil,
		},
		{
			name:    "thematic break: three asterisks",
			input:   "***",
			want:    `<hr>`,
			wantErr: nil,
		},
		{
			name:    "thematic break: three underscores",
			input:   "___",
			want:    `<hr>`,
			wantErr: nil,
		},
		{
			name:    "thematic break: spaces between markers allowed",
			input:   "- - -",
			want:    `<hr>`,
			wantErr: nil,
		},
		{
			name:    "thematic break: tabs between markers allowed",
			input:   "-\t-\t-",
			want:    `<hr>`,
			wantErr: nil,
		},
		{
			name:    "thematic break: trailing spaces allowed",
			input:   "---   ",
			want:    `<hr>`,
			wantErr: nil,
		},
		{
			name:    "thematic break: leading indentation of three spaces allowed",
			input:   "   ---",
			want:    `<hr>`,
			wantErr: nil,
		},
		{
			name:    "thematic break: leading indentation of four spaces is not thematic break",
			input:   "    ---",
			want:    `<pre><code>---</code></pre>`,
			wantErr: nil,
		},
		{
			name:    "thematic break: exactly two markers is not thematic break",
			input:   "--",
			want:    `<p>--</p>`,
			wantErr: nil,
		},
		{
			name:    "thematic break: mixed marker families are not allowed",
			input:   "-*-",
			want:    `<p>-*-</p>`,
			wantErr: nil,
		},
		{
			name:    "thematic break: internal non-whitespace invalidates line",
			input:   "--x--",
			want:    `<p>--x--</p>`,
			wantErr: nil,
		},
		{
			name:    "thematic break: leading text invalidates line",
			input:   "x---",
			want:    `<p>x---</p>`,
			wantErr: nil,
		},
		{
			name: "thematic break: between paragraphs",
			input: md(
				"hello",
				"",
				"---",
				"",
				"world",
			),
			want:    `<p>hello</p><hr><p>world</p>`,
			wantErr: nil,
		},
		{
			name: "thematic break: dash line after paragraph text becomes setext heading not thematic break",
			input: md(
				"hello",
				"---",
			),
			want:    `<h2>hello</h2>`,
			wantErr: nil,
		},
		{
			name: "thematic break: dash line after multiline paragraph text becomes setext heading not thematic break",
			input: md(
				"hello",
				"world",
				"---",
			),
			want:    `<h2>hello world</h2>`,
			wantErr: nil,
		},

		// block quotes

		{
			name:    "block quote: single quoted paragraph line",
			input:   "> hello",
			want:    `<blockquote><p>hello</p></blockquote>`,
			wantErr: nil,
		},
		{
			name: "block quote: multiple quoted lines form one paragraph",
			input: md(
				"> hello",
				"> world",
			),
			want:    `<blockquote><p>hello world</p></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: marker without following space is allowed",
			input:   ">hello",
			want:    `<blockquote><p>hello</p></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: marker with following tab is allowed",
			input:   ">\thello",
			want:    `<blockquote><p>hello</p></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: leading indentation of two spaces allowed",
			input:   `  > hello`,
			want:    `<blockquote><p>hello</p></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: leading indentation of three spaces allowed",
			input:   `   > hello`,
			want:    `<blockquote><p>hello</p></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: leading indentation of four spaces is not a block quote",
			input:   `    > hello`,
			want:    `<pre><code>&gt; hello</code></pre>`,
			wantErr: nil,
		},
		{
			name:    "block quote: empty quoted line is allowed",
			input:   ">",
			want:    `<blockquote></blockquote>`,
			wantErr: nil,
		},
		{
			name: "block quote: quoted blank line separates inner paragraphs",
			input: md(
				"> hello",
				">",
				"> world",
			),
			want:    `<blockquote><p>hello</p><p>world</p></blockquote>`,
			wantErr: nil,
		},
		{
			name: "block quote: non-quoted following line ends block quote",
			input: md(
				"> hello",
				"world",
			),
			want:    `<blockquote><p>hello</p></blockquote><p>world</p>`,
			wantErr: nil,
		},
		{
			name: "block quote: blank physical line ends block quote",
			input: md(
				"> hello",
				"",
				"> world",
			),
			want:    `<blockquote><p>hello</p></blockquote><blockquote><p>world</p></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: nested quote with double marker",
			input:   ">> hello",
			want:    `<blockquote><blockquote><p>hello</p></blockquote></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: nested quote with space between markers",
			input:   "> > hello",
			want:    `<blockquote><blockquote><p>hello</p></blockquote></blockquote>`,
			wantErr: nil,
		},
		{
			name: "block quote: nested quote across multiple lines",
			input: md(
				"> > hello",
				"> > world",
			),
			want: `<blockquote><blockquote><p>hello world</p></blockquote></blockquote>`,
		},
		{
			name: "block quote: mixed nesting depths across lines",
			input: md(
				"> outer",
				"> > inner",
				"> outer again",
			),
			want:    `<blockquote><p>outer</p><blockquote><p>inner</p></blockquote><p>outer again</p></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: triple nesting",
			input:   "> > > hello",
			want:    `<blockquote><blockquote><blockquote><p>hello</p></blockquote></blockquote></blockquote>`,
			wantErr: nil,
		},
		{
			name: "block quote: inner paragraph after nested blank line",
			input: md(
				"> > hello",
				"> >",
				"> > world",
			),
			want:    `<blockquote><blockquote><p>hello</p><p>world</p></blockquote></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: quoted atx heading",
			input:   "> # hello",
			want:    `<blockquote><h1>hello</h1></blockquote>`,
			wantErr: nil,
		},
		{
			name: "block quote: quoted setext heading",
			input: md(
				"> hello",
				"> -----",
			),
			want:    `<blockquote><h2>hello</h2></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: quoted thematic break",
			input:   "> ---",
			want:    `<blockquote><hr></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: quoted unordered list item",
			input:   "> - item",
			want:    `<blockquote><ul><li>item</li></ul></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: quoted ordered list item",
			input:   "> 1. item",
			want:    "<blockquote><ol><li>item</li></ol></blockquote>",
			wantErr: nil,
		},
		{
			name: "block quote: quoted fenced code block",
			input: md(
				"> ```",
				"> code",
				"> ```",
			),
			want:    `<blockquote><pre><code>code</code></pre></blockquote>`,
			wantErr: nil,
		},
		{
			name: "block quote: quoted indented code block after quoted blank line",
			input: md(
				">",
				">     code",
			),
			want:    `<blockquote><pre><code>code</code></pre></blockquote>`,
			wantErr: nil,
		},
		{
			name: "block quote: nested block quote contains heading and paragraph",
			input: md(
				"> # title",
				">",
				"> body",
			),
			want:    `<blockquote><h1>title</h1><p>body</p></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: line with only marker and delimiter is empty quote line",
			input:   "> ",
			want:    `<blockquote></blockquote>`,
			wantErr: nil,
		},
		{
			name:    "block quote: line with only marker and tab delimiter is empty quote line",
			input:   ">\t",
			want:    `<blockquote></blockquote>`,
			wantErr: nil,
		},

		// unordered lists

		// ordered lists

		// nested lists

		// fenced code blocks

		// indented code blocks

		// inline emphasis

		// strong emphasis

		// mixed emphasis

		// code spans

		// inline links

		// images

		// autolinks

		// raw html

		// html blocks

		// escapes

		// reference links and images

		// tight and loose lists

		// precedence and ambiguity

		// malformed input fallback
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := markdown.HTML(tc.input)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func md(xs ...string) string {
	return strings.Join(xs, "\n")
}
