package markdown

import (
	"strings"
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

// TODO:
// - trim ATX trailing markers

func TestCompile_EndToEnd(t *testing.T) {
	// TODO: rename fields input -> md, want -> html
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

		{
			name:    "unordered list: single hyphen item",
			input:   "- item",
			want:    `<ul><li>item</li></ul>`,
			wantErr: nil,
		},
		{
			name:    "unordered list: single asterisk item",
			input:   "* item",
			want:    `<ul><li>item</li></ul>`,
			wantErr: nil,
		},
		{
			name:    "unordered list: single plus item",
			input:   "+ item",
			want:    `<ul><li>item</li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: multiple sibling hyphen items",
			input: md(
				"- one",
				"- two",
				"- three",
			),
			want:    `<ul><li>one</li><li>two</li><li>three</li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: multiple sibling asterisk items",
			input: md(
				"* one",
				"* two",
				"* three",
			),
			want:    `<ul><li>one</li><li>two</li><li>three</li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: multiple sibling plus items",
			input: md(
				"+ one",
				"+ two",
				"+ three",
			),
			want:    `<ul><li>one</li><li>two</li><li>three</li></ul>`,
			wantErr: nil,
		},
		{
			name:    "unordered list: marker requires following tab or space",
			input:   "-item",
			want:    `<p>-item</p>`,
			wantErr: nil,
		},
		{
			name:    "unordered list: tab after marker allowed",
			input:   "-\titem",
			want:    `<ul><li>item</li></ul>`,
			wantErr: nil,
		},
		{
			name:    "unordered list: multiple spaces after marker allowed",
			input:   "-   item",
			want:    `<ul><li>item</li></ul>`,
			wantErr: nil,
		},
		{
			name:    "unordered list: multiple tabs and spaces after marker allowed",
			input:   "- \t  item",
			want:    `<ul><li>item</li></ul>`,
			wantErr: nil,
		},
		{
			name:    "unordered list: empty item content after required delimiter",
			input:   "- ",
			want:    `<ul><li></li></ul>`,
			wantErr: nil,
		},
		{
			name:    "unordered list: leading indentation of one space allowed",
			input:   " - item",
			want:    `<ul><li>item</li></ul>`,
			wantErr: nil,
		},
		{
			name:    "unordered list: leading indentation of two spaces allowed",
			input:   "  - item",
			want:    `<ul><li>item</li></ul>`,
			wantErr: nil,
		},
		{
			name:    "unordered list: leading indentation of three spaces allowed",
			input:   "   - item",
			want:    `<ul><li>item</li></ul>`,
			wantErr: nil,
		},
		{
			name:    "unordered list: leading indentation of four spaces is not a list",
			input:   "    - item",
			want:    `<pre><code>- item</code></pre>`,
			wantErr: nil,
		},
		{
			name: "unordered list: continuation line at content baseline stays in item",
			input: md(
				"- one",
				"  two",
			),
			want:    `<ul><li>one two</li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: continuation line beyond content baseline stays in item",
			input: md(
				"- one",
				"    two",
			),
			want:    `<ul><li>one   two</li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: dedented nonblank line ends list",
			input: md(
				"- one",
				"two",
			),
			want:    `<ul><li>one</li></ul><p>two</p>`,
			wantErr: nil,
		},
		{
			name: "unordered list: two sibling items form tight list",
			input: md(
				"- one",
				"- two",
			),
			want:    `<ul><li>one</li><li>two</li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: blank line between siblings makes loose list",
			input: md(
				"- one",
				"",
				"- two",
			),
			want:    `<ul><li><p>one</p></li><li><p>two</p></li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: blank line within item followed by continuation makes loose list",
			input: md(
				"- one",
				"",
				"  two",
			),
			want:    `<ul><li><p>one</p><p>two</p></li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: trailing blank line after final item does not become loose by rollback",
			input: md(
				"- one",
				"",
			),
			want:    `<ul><li>one</li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: blank line after item followed by dedented line rolls back blank",
			input: md(
				"- one",
				"",
				"two",
			),
			want:    `<ul><li>one</li></ul><p>two</p>`,
			wantErr: nil,
		},
		{
			name: "unordered list: continuation line may contain emphasis",
			input: md(
				"- one",
				"  *two*",
			),
			want:    `<ul><li>one <em>two</em></li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: item may contain atx heading in body",
			input: md(
				"- one",
				"  # two",
			),
			want:    `<ul><li>one<h1>two</h1></li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: item body may contain single-line setext heading",
			input: md(
				"- one",
				"  ---",
			),
			want:    `<ul><li><h2>one</h2></li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: item body may contain multiline setext heading",
			input: md(
				"- one",
				"  two",
				"  ---",
			),
			want:    `<ul><li><h2>one two</h2></li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: item may contain thematic break after blank line",
			input: md(
				"- one",
				"",
				"  ---",
			),
			want:    `<ul><li><p>one</p><hr></li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: block quote nested inside item",
			input: md(
				"- outer",
				"  > quote",
			),
			want:    `<ul><li>outer<blockquote><p>quote</p></blockquote></li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: item may contain fenced code block in body",
			input: md(
				"- one",
				"  ```",
				"  code",
				"  ```",
			),
			want:    `<ul><li>one<pre><code>code</code></pre></li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: sibling item at different absolute indent does not join list",
			input: md(
				"- one",
				" - two",
			),
			want:    `<ul><li>one</li></ul><ul><li>two</li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: mixed unordered marker families may still form sibling items",
			input: md(
				"- one",
				"* two",
				"+ three",
			),
			want:    `<ul><li>one</li><li>two</li><li>three</li></ul>`,
			wantErr: nil,
		},
		{
			name:    "unordered list: marker line with only spaces after marker creates empty item",
			input:   "-    ",
			want:    `<ul><li></li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: continuation line trimmed to item baseline before recursive parsing",
			input: md(
				"- one",
				"    > two",
			),
			want:    `<ul><li>one   &gt; two</li></ul>`,
			wantErr: nil,
		},

		// ordered lists

		{
			name:    "ordered list: single item with period delimiter",
			input:   "1. item",
			want:    `<ol><li>item</li></ol>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: single item with right paren delimiter",
			input:   "1) item",
			want:    `<ol><li>item</li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: multiple sibling items with period delimiter",
			input: md(
				"1. one",
				"2. two",
				"3. three",
			),
			want:    `<ol><li>one</li><li>two</li><li>three</li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: multiple sibling items with right paren delimiter",
			input: md(
				"1) one",
				"2) two",
				"3) three",
			),
			want:    `<ol><li>one</li><li>two</li><li>three</li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: start number preserved from first marker",
			input: md(
				"3. one",
				"4. two",
			),
			want:    `<ol start="3"><li>one</li><li>two</li></ol>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: zero start number allowed",
			input:   "0. item",
			want:    `<ol start="0"><li>item</li></ol>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: multi-digit marker allowed",
			input:   "12. item",
			want:    `<ol start="12"><li>item</li></ol>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: absurdly high marker rejected",
			input:   "1000000001. item",
			want:    `<p>1000000001. item</p>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: delimiter requires following space",
			input:   "1.item",
			want:    `<p>1.item</p>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: right paren delimiter requires following space",
			input:   "1)item",
			want:    `<p>1)item</p>`,
			wantErr: nil,
		},
		{
			name:  "ordered list: tab after delimiter allowed",
			input: "1.\titem",
			want:  `<ol><li>item</li></ol>`,
		},
		{
			name:    "ordered list: multiple spaces after delimiter allowed",
			input:   "1.    item",
			want:    `<ol><li>item</li></ol>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: multiple tabs and spaces after delimiter allowed",
			input:   "1. \t  item",
			want:    `<ol><li>item</li></ol>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: empty item content after required delimiter",
			input:   "1. ",
			want:    `<ol><li></li></ol>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: leading indentation of one space allowed",
			input:   " 1. item",
			want:    `<ol><li>item</li></ol>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: leading indentation of two spaces allowed",
			input:   "  1. item",
			want:    `<ol><li>item</li></ol>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: leading indentation of three spaces allowed",
			input:   "   1. item",
			want:    `<ol><li>item</li></ol>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: leading indentation of four spaces is not a list",
			input:   "    1. item",
			want:    `<pre><code>1. item</code></pre>`,
			wantErr: nil,
		},
		{
			name: "ordered list: continuation line at content baseline stays in item",
			input: md(
				"1. one",
				"   two",
			),
			want:    `<ol><li>one two</li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: continuation line beyond content baseline stays in item",
			input: md(
				"1. one",
				"     two",
			),
			want:    `<ol><li>one   two</li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: dedented nonblank line ends list",
			input: md(
				"1. one",
				"two",
			),
			want:    `<ol><li>one</li></ol><p>two</p>`,
			wantErr: nil,
		},
		{
			name: "ordered list: blank line between siblings makes loose list",
			input: md(
				"1. one",
				"",
				"2. two",
			),
			want:    `<ol><li><p>one</p></li><li><p>two</p></li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: blank line within item followed by continuation makes loose list",
			input: md(
				"1. one",
				"",
				"   two",
			),
			want:    `<ol><li><p>one</p><p>two</p></li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: trailing blank line after final item rolls back",
			input: md(
				"1. one",
				"",
			),
			want:    `<ol><li>one</li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: sibling item must match period delimiter family",
			input: md(
				"1. one",
				"2) two",
			),
			want:    `<ol><li>one</li></ol><ol start="2"><li>two</li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: sibling item must match right paren delimiter family",
			input: md(
				"1) one",
				"2. two",
			),
			want:    `<ol><li>one</li></ol><ol start="2"><li>two</li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: item may contain atx heading in body",
			input: md(
				"1. one",
				"   # two",
			),
			want:    `<ol><li>one<h1>two</h1></li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: item may contain setext heading in body",
			input: md(
				"1. one",
				"   two",
				"   ---",
			),
			want:    `<ol><li><h2>one two</h2></li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: item may contain thematic break after blank line",
			input: md(
				"1. one",
				"",
				"   ---",
			),
			want:    `<ol><li><p>one</p><hr></li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: block quote nested inside item",
			input: md(
				"1. outer",
				"   > quote",
			),
			want:    `<ol><li>outer<blockquote><p>quote</p></blockquote></li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: item may contain fenced code block in body",
			input: md(
				"1. one",
				"   ```",
				"   code",
				"   ```",
			),
			want:    `<ol><li>one<pre><code>code</code></pre></li></ol>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: marker line with only spaces after delimiter creates empty item",
			input:   "1.      ",
			want:    `<ol><li></li></ol>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: nonnumeric marker is not ordered list",
			input:   "x. item",
			want:    `<p>x. item</p>`,
			wantErr: nil,
		},
		{
			name:    "ordered list: missing delimiter punctuation is not ordered list",
			input:   "1 item",
			want:    `<p>1 item</p>`,
			wantErr: nil,
		},

		// nested lists and list interactions

		{
			name: "unordered list: nested unordered list in second line of item",
			input: md(
				"- outer",
				"  - inner",
			),
			want:    `<ul><li>outer<ul><li>inner</li></ul></li></ul>`,
			wantErr: nil,
		},
		{
			name: "unordered list: nested ordered list in second line of item",
			input: md(
				"- outer",
				"  1. inner",
			),
			want:    `<ul><li>outer<ol><li>inner</li></ol></li></ul>`,
			wantErr: nil,
		},
		{
			name: "ordered list: nested unordered list in second line of item",
			input: md(
				"1. outer",
				"   - inner",
			),
			want:    `<ol><li>outer<ul><li>inner</li></ul></li></ol>`,
			wantErr: nil,
		},
		{
			name: "ordered list: nested ordered list in second line of item",
			input: md(
				"1. outer",
				"   1. inner",
			),
			want:    `<ol><li>outer<ol><li>inner</li></ol></li></ol>`,
			wantErr: nil,
		},
		{
			name: "unordered list: nested sibling list items under one parent item",
			input: md(
				"- outer",
				"  - inner one",
				"  - inner two",
			),
			want:    `<ul><li>outer<ul><li>inner one</li><li>inner two</li></ul></li></ul>`,
			wantErr: nil,
		},
		{
			name: "ordered list: nested sibling list items under one parent item",
			input: md(
				"1. outer",
				"   1. inner one",
				"   2. inner two",
			),
			want:    `<ol><li>outer<ol><li>inner one</li><li>inner two</li></ol></li></ol>`,
			wantErr: nil,
		},
		{
			name: "unordered list: nested list followed by parent continuation",
			input: md(
				"- outer",
				"  - inner",
				"  tail",
			),
			want:    `<ul><li>outer<ul><li>inner</li></ul>tail</li></ul>`,
			wantErr: nil,
		},
		{
			name: "ordered list: nested list followed by parent continuation",
			input: md(
				"1. outer",
				"   1. inner",
				"   tail",
			),
			want:    `<ol><li>outer<ol><li>inner</li></ol>tail</li></ol>`,
			wantErr: nil,
		},
		{
			name: "unordered list: two top-level items each with nested list",
			input: md(
				"- outer one",
				"  - inner one",
				"- outer two",
				"  - inner two",
			),
			want:    `<ul><li>outer one<ul><li>inner one</li></ul></li><li>outer two<ul><li>inner two</li></ul></li></ul>`,
			wantErr: nil,
		},
		{
			name: "ordered list: two top-level items each with nested list",
			input: md(
				"1. outer one",
				"   1. inner one",
				"2. outer two",
				"   1. inner two",
			),
			want:    `<ol><li>outer one<ol><li>inner one</li></ol></li><li>outer two<ol><li>inner two</li></ol></li></ol>`,
			wantErr: nil,
		},
		{
			name: "unordered list: nested list separated by blank line makes parent loose",
			input: md(
				"- outer",
				"",
				"  - inner",
			),
			want:    `<ul><li><p>outer</p><ul><li>inner</li></ul></li></ul>`,
			wantErr: nil,
		},
		{
			name: "ordered list: nested list separated by blank line makes parent loose",
			input: md(
				"1. outer",
				"",
				"   1. inner",
			),
			want:    `<ol><li><p>outer</p><ol><li>inner</li></ol></li></ol>`,
			wantErr: nil,
		},
		{
			name: "unordered list: nested list item may itself contain continuation paragraph",
			input: md(
				"- outer",
				"  - inner",
				"    tail",
			),
			want:    `<ul><li>outer<ul><li>inner tail</li></ul></li></ul>`,
			wantErr: nil,
		},
		{
			name: "ordered list: nested list item may itself contain continuation paragraph",
			input: md(
				"1. outer",
				"   1. inner",
				"      tail",
			),
			want:    `<ol><li>outer<ol><li>inner tail</li></ol></li></ol>`,
			wantErr: nil,
		},
		{
			name: "unordered list: child item not meeting parent content baseline does not nest",
			input: md(
				"- outer",
				" - inner",
			),
			want:    `<ul><li>outer</li></ul><ul><li>inner</li></ul>`,
			wantErr: nil,
		},
		{
			name: "ordered list: child item not meeting parent content baseline does not nest",
			input: md(
				"1. outer",
				"  1. inner",
			),
			want:    `<ol><li>outer</li></ol><ol><li>inner</li></ol>`,
			wantErr: nil,
		},
		{
			name: "unordered list: top-level sibling resumes after nested list",
			input: md(
				"- outer",
				"  - inner",
				"- next outer",
			),
			want:    `<ul><li>outer<ul><li>inner</li></ul></li><li>next outer</li></ul>`,
			wantErr: nil,
		},
		{
			name: "ordered list: top-level sibling resumes after nested list",
			input: md(
				"1. outer",
				"   1. inner",
				"2. next outer",
			),
			want:    `<ol><li>outer<ol><li>inner</li></ol></li><li>next outer</li></ol>`,
			wantErr: nil,
		},
		{
			name: "unordered list: nested ordered list preserves start number",
			input: md(
				"- outer",
				"  3. inner",
			),
			want:    `<ul><li>outer<ol start="3"><li>inner</li></ol></li></ul>`,
			wantErr: nil,
		},
		{
			name: "ordered list: nested ordered list with right paren delimiter",
			input: md(
				"1. outer",
				"   1) inner",
			),
			want:    `<ol><li>outer<ol><li>inner</li></ol></li></ol>`,
			wantErr: nil,
		},
		{
			name: "unordered list: mixed nested unordered marker families are allowed",
			input: md(
				"- outer",
				"  * inner",
				"  + inner two",
			),
			want:    `<ul><li>outer<ul><li>inner</li><li>inner two</li></ul></li></ul>`,
			wantErr: nil,
		},
		{
			name: "ordered list: nested ordered sibling delimiter mismatch splits structure",
			input: md(
				"1. outer",
				"   1. inner one",
				"   2) inner two",
			),
			want:    `<ol><li>outer<ol><li>inner one</li></ol><ol start="2"><li>inner two</li></ol></li></ol>`,
			wantErr: nil,
		},

		// fenced code blocks

		{
			name: "fenced code: backtick fence minimum opener and closer",
			input: md(
				"```",
				"code",
				"```",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: tilde fence minimum opener and closer",
			input: md(
				"~~~",
				"code",
				"~~~",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: longer backtick opener and matching closer",
			input: md(
				"````",
				"code",
				"````",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: longer tilde opener and matching closer",
			input: md(
				"~~~~",
				"code",
				"~~~~",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: closer longer than opener",
			input: md(
				"```",
				"code",
				"````",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: closer shorter than opener does not close",
			input: md(
				"````",
				"code",
				"```",
			),
			want:    "<pre><code>code\n```</code></pre>",
			wantErr: nil,
		},
		{
			name: "fenced code: opener with one leading space",
			input: md(
				" ```",
				"code",
				" ```",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: opener with two leading spaces",
			input: md(
				"  ```",
				"code",
				"  ```",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: opener with three leading spaces",
			input: md(
				"   ```",
				"code",
				"   ```",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: opener with four leading spaces is not fenced code",
			input: md(
				"    ```",
				"    code",
				"    ```",
			),
			want:    "<pre><code>```\ncode\n```</code></pre>",
			wantErr: nil,
		},
		{
			name: "fenced code: closer with one leading space",
			input: md(
				"```",
				"code",
				" ```",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: closer with two leading spaces",
			input: md(
				"```",
				"code",
				"  ```",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: closer with three leading spaces",
			input: md(
				"```",
				"code",
				"   ```",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: closer with four leading spaces is not closing fence",
			input: md(
				"```",
				"code",
				"    ```",
				"```",
			),
			want:    "<pre><code>code\n    ```</code></pre>",
			wantErr: nil,
		},
		{
			name: "fenced code: blank line inside block",
			input: md(
				"```",
				"one",
				"",
				"two",
				"```",
			),
			want:    "<pre><code>one\n\ntwo</code></pre>",
			wantErr: nil,
		},
		{
			name: "fenced code: empty fenced block",
			input: md(
				"```",
				"```",
			),
			want:    `<pre><code></code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: unclosed backtick fence runs to eof",
			input: md(
				"```",
				"code",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: unclosed tilde fence runs to eof",
			input: md(
				"~~~",
				"code",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: payload line equal to shorter fence is literal content",
			input: md(
				"````",
				"```",
				"````",
			),
			want:    "<pre><code>```</code></pre>",
			wantErr: nil,
		},
		{
			name: "fenced code: closing fence may have trailing spaces",
			input: md(
				"```",
				"code",
				"```   ",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: closing fence may have trailing tabs",
			input: md(
				"```",
				"code",
				"```\t\t",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: closing fence with trailing nonwhitespace is not valid closer",
			input: md(
				"```",
				"code",
				"```x",
				"```",
			),
			want:    "<pre><code>code\n```x</code></pre>",
			wantErr: nil,
		},
		{
			name: "fenced code: backtick opener with info string",
			input: md(
				"```go",
				"code",
				"```",
			),
			want:    `<pre><code class="language-go">code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: backtick opener with delimiter whitespace before info string",
			input: md(
				"```   go",
				"code",
				"```",
			),
			want:    `<pre><code class="language-go">code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: tilde opener may contain backticks in info string",
			input: md(
				"~~~ ```",
				"code",
				"~~~",
			),
			want:    "<pre><code class=\"language-```\">code</code></pre>",
			wantErr: nil,
		},
		{
			name: "fenced code: backtick opener rejects info string containing backtick",
			input: md(
				"``` `",
				"code",
				"```",
			),
			want:    "<p>``` ` code</p><pre><code></code></pre>",
			wantErr: nil,
		},
		{
			name: "fenced code: opener with fewer than three markers is not fenced code",
			input: md(
				"``",
				"code",
				"``",
			),
			want:    "<p>`` code ``</p>",
			wantErr: nil,
		},
		{
			name: "fenced code: mixed marker family does not close block",
			input: md(
				"```",
				"code",
				"~~~",
				"```",
			),
			want:    "<pre><code>code\n~~~</code></pre>",
			wantErr: nil,
		},
		{
			name: "fenced code: marker-looking content line is literal until valid closer",
			input: md(
				"```",
				"~~~",
				"```",
			),
			want:    `<pre><code>~~~</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: html-looking content inside block",
			input: md(
				"```",
				"<div>",
				"```",
			),
			want:    `<pre><code>&lt;div&gt;</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: block-quote-looking content inside block",
			input: md(
				"```",
				"> hello",
				"```",
			),
			want:    `<pre><code>&gt; hello</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: list-looking content inside block",
			input: md(
				"```",
				"- hello",
				"```",
			),
			want:    `<pre><code>- hello</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: fenced opener interrupts paragraph without blank line",
			input: md(
				"one",
				"```",
				"two",
				"```",
			),
			want:    `<p>one</p><pre><code>two</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fenced code: fenced block followed by paragraph",
			input: md(
				"```",
				"code",
				"```",
				"tail",
			),
			want:    `<pre><code>code</code></pre><p>tail</p>`,
			wantErr: nil,
		},
		{
			name: "fenced code: opener with only delimiter whitespace and no info string",
			input: md(
				"```   ",
				"code",
				"```",
			),
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},

		// indented code blocks

		{
			name:    "indented code: single line with four spaces",
			input:   `    code`,
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name:    "indented code: single line with more than four spaces preserves remainder",
			input:   `      code`,
			want:    `<pre><code>  code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "indented code: multiple indented lines",
			input: md(
				"    one",
				"    two",
			),
			want:    "<pre><code>one\ntwo</code></pre>",
			wantErr: nil,
		},
		{
			name: "indented code: blank line inside block",
			input: md(
				"    one",
				"",
				"    two",
			),
			want:    "<pre><code>one\n\ntwo</code></pre>",
			wantErr: nil,
		},
		{
			name: "indented code: trailing blank lines rolled back before dedented line",
			input: md(
				"    one",
				"",
				"two",
			),
			want:    `<pre><code>one</code></pre><p>two</p>`,
			wantErr: nil,
		},
		{
			name: "indented code: trailing blank lines at eof",
			input: md(
				"    one",
				"",
			),
			want:    "<pre><code>one\n</code></pre>",
			wantErr: nil,
		},
		{
			name:    "indented code: line with three leading spaces is not code block",
			input:   `   code`,
			want:    `<p>   code</p>`,
			wantErr: nil,
		},
		{
			name:    "indented code: tab reaching four columns",
			input:   "\tcode",
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name:    "indented code: mixed indentation reaching four columns",
			input:   "  \tcode",
			want:    `<pre><code>code</code></pre>`,
			wantErr: nil,
		},
		{
			name: "indented code: paragraph transparency with continuation line",
			input: md(
				"one",
				"    two",
			),
			want:    `<p>one     two</p>`,
			wantErr: nil,
		},
		{
			name: "indented code: begins after blank line following paragraph",
			input: md(
				"one",
				"",
				"    two",
			),
			want:    `<p>one</p><pre><code>two</code></pre>`,
			wantErr: nil,
		},
		{
			name: "indented code: dedented nonblank line ends block",
			input: md(
				"    one",
				"    two",
				"three",
			),
			want:    "<pre><code>one\ntwo</code></pre><p>three</p>",
			wantErr: nil,
		},
		{
			name:    "indented code: thematic-break-looking content is literal",
			input:   `    ---`,
			want:    `<pre><code>---</code></pre>`,
			wantErr: nil,
		},
		{
			name:    "indented code: block-quote-looking content is literal",
			input:   `    > hello`,
			want:    `<pre><code>&gt; hello</code></pre>`,
			wantErr: nil,
		},
		{
			name:    "indented code: list-looking content is literal",
			input:   `    - hello`,
			want:    `<pre><code>- hello</code></pre>`,
			wantErr: nil,
		},
		{
			name:    "indented code: atx-heading-looking content is literal",
			input:   `    # hello`,
			want:    `<pre><code># hello</code></pre>`,
			wantErr: nil,
		},
		{
			name:    "indented code: fenced-opener-looking content is literal",
			input:   `    ~~~`,
			want:    `<pre><code>~~~</code></pre>`,
			wantErr: nil,
		},
		{
			name: "indented code: multiple blank lines inside block",
			input: md(
				"    one",
				"",
				"",
				"    two",
			),
			want:    "<pre><code>one\n\n\ntwo</code></pre>",
			wantErr: nil,
		},
		{
			name:    "indented code: trailing spaces in content line preserved",
			input:   "    code  ",
			want:    `<pre><code>code  </code></pre>`,
			wantErr: nil,
		},
		{
			name:    "indented code: html-looking content escaped",
			input:   `    <div>`,
			want:    `<pre><code>&lt;div&gt;</code></pre>`,
			wantErr: nil,
		},

		// strong & emphasis

		{
			name:    "emphasis: single delimiter produces emphasis",
			input:   "*a*",
			want:    `<p><em>a</em></p>`,
			wantErr: nil,
		},
		{
			name:    "strong: double delimiter produces strong emphasis",
			input:   "**a**",
			want:    `<p><strong>a</strong></p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: underscore produces emphasis",
			input:   "_a_",
			want:    `<p><em>a</em></p>`,
			wantErr: nil,
		},
		{
			name:    "strong: underscore produces strong emphasis",
			input:   "__a__",
			want:    `<p><strong>a</strong></p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: delimiter must be left and right flanking",
			input:   "a *b* c",
			want:    `<p>a <em>b</em> c</p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: opening delimiter cannot be followed by whitespace",
			input:   "* b*",
			want:    `<ul><li>b*</li></ul>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: closing delimiter cannot be preceded by whitespace",
			input:   "*b *",
			want:    `<p>*b *</p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: delimiter adjacent to punctuation can open",
			input:   "(*a*)",
			want:    `<p>(<em>a</em>)</p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: delimiter adjacent to punctuation can close",
			input:   "(*a*)",
			want:    `<p>(<em>a</em>)</p>`,
			wantErr: nil,
		},
		{
			name:    "underscore: intraword emphasis is disallowed",
			input:   "foo_bar_baz",
			want:    `<p>foo_bar_baz</p>`,
			wantErr: nil,
		},
		{
			name:    "underscore: emphasis allowed when separated by punctuation",
			input:   "foo _bar_ baz",
			want:    `<p>foo <em>bar</em> baz</p>`,
			wantErr: nil,
		},
		{
			name:    "underscore: can open when preceded by punctuation",
			input:   "(_a_)",
			want:    `<p>(<em>a</em>)</p>`,
			wantErr: nil,
		},
		{
			name:    "underscore: can close when followed by punctuation",
			input:   "(_a_)",
			want:    `<p>(<em>a</em>)</p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: triple delimiter can resolve to emphasis and strong nesting",
			input:   "***a***",
			want:    `<p><em><strong>a</strong></em></p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: unmatched extra delimiter remains as literal text",
			input:   "**a*",
			want:    `<p>*<em>a</em></p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: unmatched closing delimiter remains as literal text",
			input:   "*a**",
			want:    `<p><em>a</em>*</p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: delimiter pairing blocked by multiple of three rule",
			input:   "***a**",
			want:    `<p>*<strong>a</strong></p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: mod-3 rule prevents pairing across runs",
			input:   "**a***",
			want:    `<p><strong>a</strong>*</p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: nested emphasis resolves inside outer emphasis",
			input:   "*a *b* c*",
			want:    `<p><em>a <em>b</em> c</em></p>`,
			wantErr: nil,
		},
		{
			name:    "strong: nested strong inside emphasis",
			input:   "*a **b** c*",
			want:    `<p><em>a <strong>b</strong> c</em></p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: nested emphasis inside strong",
			input:   "**a *b* c**",
			want:    `<p><strong>a <em>b</em> c</strong></p>`,
			wantErr: nil,
		},
		{
			name:    "strong: nested strong resolves correctly",
			input:   "**a **b** c**",
			want:    `<p><strong>a <strong>b</strong> c</strong></p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: crossing delimiters do not form valid nesting",
			input:   "*a **b* c**",
			want:    `<p><em>a <em><em>b</em> c</em></em></p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: multiple independent emphasis runs",
			input:   "*a* *b*",
			want:    `<p><em>a</em> <em>b</em></p>`,
			wantErr: nil,
		},
		{
			name:    "strong: multiple independent strong runs",
			input:   "**a** **b**",
			want:    `<p><strong>a</strong> <strong>b</strong></p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: mixed delimiter kinds do not match",
			input:   "*a_",
			want:    `<p>*a_</p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: mixed delimiter kinds do not match",
			input:   "_a*",
			want:    `<p>_a*</p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: delimiter at start of input",
			input:   "*a*",
			want:    `<p><em>a</em></p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: delimiter at end of input",
			input:   "a *b*",
			want:    `<p>a <em>b</em></p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: isolated delimiter produces literal text",
			input:   "*",
			want:    `<p>*</p>`,
			wantErr: nil,
		},
		{
			name:    "strong: isolated double delimiter produces literal text",
			input:   "**",
			want:    `<p>**</p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: empty content does not form emphasis",
			input:   "**",
			want:    `<p>**</p>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: empty content between delimiters is ignored",
			input:   "* *",
			want:    `<ul><li>*</li></ul>`,
			wantErr: nil,
		},
		{
			name:    "emphasis: emphasis spans do not consume surrounding text",
			input:   "a *b* c",
			want:    `<p>a <em>b</em> c</p>`,
			wantErr: nil,
		},
		{
			name:    "strong: strong spans do not consume surrounding text",
			input:   "a **b** c",
			want:    `<p>a <strong>b</strong> c</p>`,
			wantErr: nil,
		},

		// code spans

		{
			name:    "code span: matching single delimiters produce code span",
			input:   "`code`",
			want:    `<p><code>code</code></p>`,
			wantErr: nil,
		},
		{
			name:    "code span: matching multi-delimiters produce code span",
			input:   "``code``",
			want:    `<p><code>code</code></p>`,
			wantErr: nil,
		},
		{
			name:    "code span: content may contain single backticks when wrapped in wider delimiters",
			input:   "``a ` b``",
			want:    "<p><code>a ` b</code></p>",
			wantErr: nil,
		},
		{
			name:    "code span: content may contain double backticks when wrapped in wider delimiters",
			input:   "```a `` b```",
			want:    "<p><code>a `` b</code></p>",
			wantErr: nil,
		},
		{
			name:    "code span: opening delimiter without matching closer remains literal text",
			input:   "`code",
			want:    "<p>`code</p>",
			wantErr: nil,
		},
		{
			name:    "code span: closing delimiter with different width does not close span",
			input:   "`code``",
			want:    "<p>`code``</p>",
			wantErr: nil,
		},
		{
			name:    "code span: opener skips non-matching backtick runs and closes on matching width",
			input:   "``code`more``",
			want:    "<p><code>code`more</code></p>",
			wantErr: nil,
		},
		{
			name:    "code span: empty content is allowed",
			input:   "````",
			want:    "<pre><code></code></pre>",
			wantErr: nil,
		},
		{
			name:    "code span: single leading and trailing spaces are trimmed when content is not all spaces",
			input:   "` code `",
			want:    "<p><code>code</code></p>",
			wantErr: nil,
		},
		{
			name:    "code span: only one leading and trailing space are trimmed",
			input:   "`  code  `",
			want:    "<p><code> code </code></p>",
			wantErr: nil,
		},
		{
			name:    "code span: all-space content is not trimmed",
			input:   "`   `",
			want:    "<p><code>   </code></p>",
			wantErr: nil,
		},
		{
			name:    "code span: interior spaces are preserved",
			input:   "`a  b`",
			want:    "<p><code>a  b</code></p>",
			wantErr: nil,
		},
		{
			name:    "code span: punctuation is treated as literal content",
			input:   "`<a>*&_[]()`",
			want:    "<p><code>&lt;a&gt;*&amp;_[]()</code></p>",
			wantErr: nil,
		},
		{
			name:    "code span: emphasis delimiters inside span are literal content",
			input:   "`*a*`",
			want:    "<p><code>*a*</code></p>",
			wantErr: nil,
		},
		{
			name:    "code span: brackets and parentheses inside span are literal content",
			input:   "`[a](b)`",
			want:    "<p><code>[a](b)</code></p>",
			wantErr: nil,
		},
		{
			name:    "code span: span does not consume surrounding text",
			input:   "a `code` b",
			want:    "<p>a <code>code</code> b</p>",
			wantErr: nil,
		},
		{
			name:    "code span: multiple code spans may appear in one line",
			input:   "`a` and `b`",
			want:    "<p><code>a</code> and <code>b</code></p>",
			wantErr: nil,
		},
		{
			name:    "code span: unmatched wider opener remains literal text",
			input:   "``code`",
			want:    "<p>``code`</p>",
			wantErr: nil,
		},
		{
			name:    "code span: later matching closer forms span after earlier mismatched run",
			input:   "```code``more```",
			want:    "<p><code>code``more</code></p>",
			wantErr: nil,
		},
		{
			name:    "code span: lone backtick remains literal text",
			input:   "`",
			want:    "<p>`</p>",
			wantErr: nil,
		},

		// inline links

		{
			name:    "link: inline destination resolves to link",
			input:   "[label](/url)",
			want:    `<p><a href="/url">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: empty destination is allowed",
			input:   "[label]()",
			want:    `<p><a href="">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: empty label is allowed",
			input:   "[](/url)",
			want:    `<p><a href="/url"></a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: angle-delimited destination resolves to link",
			input:   "[label](</url>)",
			want:    `<p><a href="/url">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: bare destination resolves to link",
			input:   "[label](/a/b)",
			want:    `<p><a href="/a/b">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: bare destination may contain balanced parentheses",
			input:   "[label](a(b)c)",
			want:    `<p><a href="a(b)c">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: bare destination may contain escaped parentheses",
			input:   "[label](a\\(b\\)c)",
			want:    `<p><a href="a(b)c">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: escaped punctuation is unescaped in destination and title",
			input:   `[label](a\(b\)c "ti\"tle")`,
			want:    `<p><a href="a(b)c" title="ti&#34;tle">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: destination may include a double-quoted title",
			input:   "[label](/url \"title\")",
			want:    `<p><a href="/url" title="title">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: destination may include a single-quoted title",
			input:   "[label](/url 'title')",
			want:    `<p><a href="/url" title="title">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: destination may include a parenthesized title",
			input:   "[label](/url (title))",
			want:    `<p><a href="/url" title="title">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: title may contain balanced parentheses",
			input:   "[label](/url (a (b) c))",
			want:    `<p><a href="/url" title="a (b) c">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: whitespace between destination and title is required",
			input:   "[label](/url\"title\")",
			want:    `<p><a href="/url&#34;title&#34;">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: spaces and tabs around destination are allowed",
			input:   "[label]( \t/url\t )",
			want:    `<p><a href="/url">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: spaces and tabs around destination and title are allowed",
			input:   "[label]( \t/url \t \"title\"\t )",
			want:    `<p><a href="/url" title="title">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: nested inline content is allowed in label",
			input:   "[a *b* c](/url)",
			want:    `<p><a href="/url">a <em>b</em> c</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: code span is allowed inside label",
			input:   "[a `b` c](/url)",
			want:    `<p><a href="/url">a <code>b</code> c</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: image is allowed inside label",
			input:   "[![alt](/img.png)](/url)",
			want:    `<p><a href="/url"><img alt="alt" src="/img.png"></a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: surrounding text is preserved",
			input:   "a [label](/url) b",
			want:    `<p>a <a href="/url">label</a> b</p>`,
			wantErr: nil,
		},
		{
			name:    "link: missing tail leaves brackets as literal text",
			input:   "[label]",
			want:    `<p>[label]</p>`,
			wantErr: nil,
		},
		{
			name:    "link: missing closing parenthesis leaves construct as literal text",
			input:   "[label](/url",
			want:    `<p>[label](/url</p>`,
			wantErr: nil,
		},
		{
			name:    "link: invalid angle destination leaves construct as literal text",
			input:   "[label](<a<>)",
			want:    `<p>[label](&lt;a&lt;&gt;)</p>`,
			wantErr: nil,
		},
		{
			name:    "link: unbalanced bare destination leaves construct as literal text",
			input:   "[label](a(b)",
			want:    `<p>[label](a(b)</p>`,
			wantErr: nil,
		},
		{
			name:    "link: quoted text without separator is parsed as bare destination",
			input:   "[label](\"title\")",
			want:    `<p><a href="&#34;title&#34;">label</a></p>`,
			wantErr: nil,
		},
		{
			name:    "link: newline in angle destination leaves construct as literal text",
			input:   "[label](<a\nb>)",
			want:    `<p>[label](&lt;a b&gt;)</p>`,
			wantErr: nil,
		},
		{
			name:    "link: newline in quoted title leaves construct as literal text",
			input:   "[label](/url \"a\nb\")",
			want:    `<p>[label](/url &#34;a b&#34;)</p>`,
			wantErr: nil,
		},
		{
			name:    "link: newline in parenthesized title leaves construct as literal text",
			input:   "[label](/url (a\nb))",
			want:    `<p>[label](/url (a b))</p>`,
			wantErr: nil,
		},
		{
			name:    "link: literal closing bracket without opener remains text",
			input:   "label](/url)",
			want:    `<p>label](/url)</p>`,
			wantErr: nil,
		},
		{
			name:    "link: nested links are rejected",
			input:   "[outer [inner](/in)](/out)",
			want:    `<p>[outer <a href="/in">inner</a>](/out)</p>`,
			wantErr: nil,
		},
		{
			name:    "link: inner links may still resolve when outer link is rejected",
			input:   "[outer [inner](/in)]",
			want:    `<p>[outer <a href="/in">inner</a>]</p>`,
			wantErr: nil,
		},

		// inline images

		// {
		// 	name:  "image: inline destination resolves to image",
		// 	input: "![alt](/img.png)",
		// },
		// {
		// 	name:  "image: empty destination is allowed",
		// 	input: "![alt]()",
		// },
		// {
		// 	name:  "image: empty label is allowed",
		// 	input: "![](/img.png)",
		// },
		// {
		// 	name:  "image: angle-delimited destination resolves to image",
		// 	input: "![alt](</img.png>)",
		// },
		// {
		// 	name:  "image: bare destination resolves to image",
		// 	input: "![alt](/a/b.png)",
		// },
		// {
		// 	name:  "image: bare destination may contain balanced parentheses",
		// 	input: "![alt](a(b)c.png)",
		// },
		// {
		// 	name:  "image: bare destination may contain escaped parentheses",
		// 	input: "![alt](a\\(b\\)c.png)",
		// },
		{
			name:    "image: escaped punctuation is unescaped in destination and title",
			input:   `![alt](a\(b\)c "ti\"tle")`,
			want:    `<p><img alt="alt" src="a(b)c" title="ti&#34;tle"></p>`,
			wantErr: nil,
		},
		// {
		// 	name:  "image: destination may include a double-quoted title",
		// 	input: "![alt](/img.png \"title\")",
		// },
		// {
		// 	name:  "image: destination may include a single-quoted title",
		// 	input: "![alt](/img.png 'title')",
		// },
		// {
		// 	name:  "image: destination may include a parenthesized title",
		// 	input: "![alt](/img.png (title))",
		// },
		// {
		// 	name:  "image: title may contain balanced parentheses",
		// 	input: "![alt](/img.png (a (b) c))",
		// },
		// {
		// 	name:  "image: whitespace between destination and title is required",
		// 	input: "![alt](/img.png\"title\")",
		// },
		// {
		// 	name:  "image: spaces and tabs around destination are allowed",
		// 	input: "![alt]( \t/img.png\t )",
		// },
		// {
		// 	name:  "image: spaces and tabs around destination and title are allowed",
		// 	input: "![alt]( \t/img.png \t \"title\"\t )",
		// },
		// {
		// 	name:  "image: nested inline content is allowed in label",
		// 	input: "![a *b* c](/img.png)",
		// },
		// {
		// 	name:  "image: code span is allowed inside label",
		// 	input: "![a `b` c](/img.png)",
		// },
		// {
		// 	name:  "image: link is allowed inside image label",
		// 	input: "![see [this](/url)](/img.png)",
		// },
		// {
		// 	name:  "image: surrounding text is preserved",
		// 	input: "a ![alt](/img.png) b",
		// },
		// {
		// 	name:  "image: missing tail leaves brackets as literal text",
		// 	input: "![alt]",
		// },
		// {
		// 	name:  "image: missing closing parenthesis leaves construct as literal text",
		// 	input: "![alt](/img.png",
		// },
		// {
		// 	name:  "image: invalid angle destination leaves construct as literal text",
		// 	input: "![alt](<a<>)",
		// },
		// {
		// 	name:  "image: unbalanced bare destination leaves construct as literal text",
		// 	input: "![alt](a(b)",
		// },
		// {
		// 	name:  "image: title without destination leaves construct as literal text",
		// 	input: "![alt](\"title\")",
		// },
		// {
		// 	name:  "image: newline in angle destination leaves construct as literal text",
		// 	input: "![alt](<a\nb>)",
		// },
		// {
		// 	name:  "image: newline in quoted title leaves construct as literal text",
		// 	input: "![alt](/img.png \"a\nb\")",
		// },
		// {
		// 	name:  "image: newline in parenthesized title leaves construct as literal text",
		// 	input: "![alt](/img.png (a\nb))",
		// },
		// {
		// 	name:  "image: literal closing bracket without opener remains text",
		// 	input: "alt](/img.png)",
		// },
		// {
		// 	name:  "image: nested images are allowed",
		// 	input: "![outer ![inner](/in.png)](/out.png)",
		// },

		// inline autolinks

		// 		{
		// 	name:  "autolink: uri with alphabetic scheme resolves to link",
		// 	input: "<https://example.com>",
		// },
		// {
		// 	name:  "autolink: uri scheme may include plus period and hyphen",
		// 	input: "<a+b.c-d:xyz>",
		// },
		// {
		// 	name:  "autolink: uri scheme may include digits after the first character",
		// 	input: "<x1:abc>",
		// },
		// {
		// 	name:  "autolink: uri scheme must begin with a letter",
		// 	input: "<1x:abc>",
		// },
		// {
		// 	name:  "autolink: uri scheme must contain a colon",
		// 	input: "<https//example.com>",
		// },
		// {
		// 	name:  "autolink: uri scheme must be at least two characters",
		// 	input: "<h:abc>",
		// },
		// {
		// 	name:  "autolink: uri scheme may not exceed thirty two characters",
		// 	input: "<abcdefghijklmnopqrstuvwxyzabcdef:abc>",
		// },
		// {
		// 	name:  "autolink: uri may contain query punctuation",
		// 	input: "<https://example.com?a=1&b=2>",
		// },
		// {
		// 	name:  "autolink: uri may not contain spaces",
		// 	input: "<https://example .com>",
		// },
		// {
		// 	name:  "autolink: uri may not contain angle brackets in content",
		// 	input: "<https://exa<mple.com>",
		// },
		// {
		// 	name:  "autolink: uri without closing angle bracket remains literal text",
		// 	input: "<https://example.com",
		// },
		// {
		// 	name:  "autolink: surrounding text is preserved for uri autolink",
		// 	input: "a <https://example.com> b",
		// },
		//
		// {
		// 	name:  "autolink: email address resolves to mail link",
		// 	input: "<user@example.com>",
		// },
		// {
		// 	name:  "autolink: email local part may contain permitted punctuation",
		// 	input: "<a.b+c_d-test@example.com>",
		// },
		// {
		// 	name:  "autolink: email domain may contain hyphen within label",
		// 	input: "<user@exa-mple.com>",
		// },
		// {
		// 	name:  "autolink: email requires exactly one at sign",
		// 	input: "<a@b@c.com>",
		// },
		// {
		// 	name:  "autolink: email requires nonempty local part",
		// 	input: "<@example.com>",
		// },
		// {
		// 	name:  "autolink: email requires nonempty domain",
		// 	input: "<user@>",
		// },
		// {
		// 	name:  "autolink: email domain labels may not begin with hyphen",
		// 	input: "<user@-example.com>",
		// },
		// {
		// 	name:  "autolink: email domain labels may not end with hyphen",
		// 	input: "<user@example-.com>",
		// },
		// {
		// 	name:  "autolink: email domain labels may not contain underscore",
		// 	input: "<user@exa_mple.com>",
		// },
		// {
		// 	name:  "autolink: email domain labels may not be empty",
		// 	input: "<user@example..com>",
		// },
		// {
		// 	name:  "autolink: email without closing angle bracket remains literal text",
		// 	input: "<user@example.com",
		// },
		// {
		// 	name:  "autolink: surrounding text is preserved for email autolink",
		// 	input: "a <user@example.com> b",
		// },
		//
		// {
		// 	name:  "autolink: invalid angle content falls back to literal text",
		// 	input: "<local@domain>",
		// },
		// {
		// 	name:  "autolink: invalid uri content falls back to literal text",
		// 	input: "<http:exa mple>",
		// },
		// {
		// 	name:  "autolink: invalid email content falls back to literal text",
		// 	input: "<user@exa_mple.com>",
		// },
		// {
		// 	name:  "autolink: lone opening angle bracket remains literal text",
		// 	input: "<",
		// },
		// {
		// 	name:  "autolink: empty angle pair remains literal text",
		// 	input: "<>",
		// },

		// raw inline HTML

		// 		{
		// 	name:  "inline html: comment resolves as raw html",
		// 	input: "<!-- comment -->",
		// },
		// {
		// 	name:  "inline html: empty comment resolves as raw html",
		// 	input: "<!---->",
		// },
		// {
		// 	name:  "inline html: processing instruction resolves as raw html",
		// 	input: "<?php?>",
		// },
		// {
		// 	name:  "inline html: declaration resolves as raw html",
		// 	input: "<!DOCTYPE html>",
		// },
		// {
		// 	name:  "inline html: cdata section resolves as raw html",
		// 	input: "<![CDATA[hello]]>",
		// },
		// {
		// 	name:  "inline html: opening tag resolves as raw html",
		// 	input: "<span>",
		// },
		// {
		// 	name:  "inline html: closing tag resolves as raw html",
		// 	input: "</span>",
		// },
		// {
		// 	name:  "inline html: opening tag may contain attributes",
		// 	input: "<a href=\"/url\" title=\"x\">",
		// },
		// {
		// 	name:  "inline html: opening tag may contain single quoted attributes",
		// 	input: "<a href='/url' title='x'>",
		// },
		// {
		// 	name:  "inline html: opening tag may contain unquoted attributes",
		// 	input: "<a href=/url title=x>",
		// },
		// {
		// 	name:  "inline html: opening tag may contain bare attributes",
		// 	input: "<input disabled>",
		// },
		// {
		// 	name:  "inline html: opening tag may contain spaces before attributes",
		// 	input: "<a  href=\"/url\">",
		// },
		// {
		// 	name:  "inline html: self closing tag resolves as raw html",
		// 	input: "<br/>",
		// },
		// {
		// 	name:  "inline html: self closing tag may contain spaces before slash",
		// 	input: "<br />",
		// },
		// {
		// 	name:  "inline html: self closing tag may contain attributes",
		// 	input: "<img src=\"/img.png\" alt=\"x\" />",
		// },
		// {
		// 	name:  "inline html: tag attributes may mix quoting styles",
		// 	input: "<a href=\"/url\" data-x='y' rel=noopener>",
		// },
		// {
		// 	name:  "inline html: quoted attribute values may contain closing angle brackets",
		// 	input: "<a title=\"1 > 0\">",
		// },
		// {
		// 	name:  "inline html: quoted attribute values may contain opposite quote kind",
		// 	input: "<a title='\"x\"'>",
		// },
		// {
		// 	name:  "inline html: surrounding text is preserved",
		// 	input: "a <span> b",
		// },
		//
		// {
		// 	name:  "inline html: opening tag requires alphabetic tag name start",
		// 	input: "<1span>",
		// },
		// {
		// 	name:  "inline html: closing tag requires alphabetic tag name start",
		// 	input: "</1span>",
		// },
		// {
		// 	name:  "inline html: tag name may contain digits after the first character",
		// 	input: "<h1>",
		// },
		// {
		// 	name:  "inline html: tag name may contain hyphen",
		// 	input: "<custom-tag>",
		// },
		// {
		// 	name:  "inline html: attribute name may begin with underscore",
		// 	input: "<a _x=\"y\">",
		// },
		// {
		// 	name:  "inline html: attribute name may begin with colon",
		// 	input: "<a :x=\"y\">",
		// },
		// {
		// 	name:  "inline html: attribute name may contain punctuation",
		// 	input: "<a data.x:y-z=\"v\">",
		// },
		//
		// {
		// 	name:  "inline html: unterminated comment remains literal text",
		// 	input: "<!-- comment",
		// },
		// {
		// 	name:  "inline html: unterminated processing instruction remains literal text",
		// 	input: "<?php",
		// },
		// {
		// 	name:  "inline html: unterminated declaration remains literal text",
		// 	input: "<!DOCTYPE html",
		// },
		// {
		// 	name:  "inline html: unterminated cdata remains literal text",
		// 	input: "<![CDATA[hello",
		// },
		// {
		// 	name:  "inline html: unterminated opening tag remains literal text",
		// 	input: "<span",
		// },
		// {
		// 	name:  "inline html: unterminated closing tag remains literal text",
		// 	input: "</span",
		// },
		// {
		// 	name:  "inline html: attribute requires separator whitespace",
		// 	input: "<ahref=\"/url\">",
		// },
		// {
		// 	name:  "inline html: attribute value requires nonempty content when unquoted",
		// 	input: "<a href=>",
		// },
		// {
		// 	name:  "inline html: attribute value requires closing quote",
		// 	input: "<a href=\"/url>",
		// },
		// {
		// 	name:  "inline html: single quoted attribute value requires closing quote",
		// 	input: "<a href='/url>",
		// },
		// {
		// 	name:  "inline html: self closing suffix must run directly to closing bracket",
		// 	input: "<br/ x>",
		// },
		// {
		// 	name:  "inline html: closing tag allows trailing spaces only",
		// 	input: "</span >",
		// },
		// {
		// 	name:  "inline html: closing tag rejects extra content",
		// 	input: "</span x>",
		// },
		// {
		// 	name:  "inline html: unquoted attribute value may not contain less than",
		// 	input: "<a href=x<y>",
		// },
		// {
		// 	name:  "inline html: unquoted attribute value may not contain greater than",
		// 	input: "<a href=x>y>",
		// },
		// {
		// 	name:  "inline html: unquoted attribute value may not contain backtick",
		// 	input: "<a href=x`y>",
		// },
		// {
		// 	name:  "inline html: lone opening angle bracket remains literal text",
		// 	input: "<",
		// },
		// {
		// 	name:  "inline html: unknown angle construct remains literal text",
		// 	input: "<!>",
		// },

		// HTML blocks

		{
			name:    "html block: single line comment block",
			input:   "<!-- hello -->",
			want:    "<!-- hello -->",
			wantErr: nil,
		},
		{
			name: "html block: multiline comment block",
			input: md(
				"<!--",
				"hello",
				"-->",
			),
			want:    "<!--\nhello\n-->",
			wantErr: nil,
		},
		{
			name: "html block: comment terminator on opening line closes block immediately",
			input: md(
				"<!-- hello -->",
				"tail",
			),
			want:    "<!-- hello --><p>tail</p>",
			wantErr: nil,
		},
		{
			name: "html block: comment block runs to eof when unterminated",
			input: md(
				"<!--",
				"hello",
			),
			want:    "<!--\nhello",
			wantErr: nil,
		},
		{
			name:    "html block: single line cdata block",
			input:   "<![CDATA[hello]]>",
			want:    "<![CDATA[hello]]>",
			wantErr: nil,
		},
		{
			name: "html block: multiline cdata block",
			input: md(
				"<![CDATA[",
				"hello",
				"]]>",
			),
			want:    "<![CDATA[\nhello\n]]>",
			wantErr: nil,
		},
		{
			name: "html block: cdata terminator on opening line closes block immediately",
			input: md(
				"<![CDATA[hello]]>",
				"tail",
			),
			want:    "<![CDATA[hello]]><p>tail</p>",
			wantErr: nil,
		},
		{
			name: "html block: cdata block runs to eof when unterminated",
			input: md(
				"<![CDATA[",
				"hello",
			),
			want:    "<![CDATA[\nhello",
			wantErr: nil,
		},
		{
			name:    "html block: single line processing instruction block",
			input:   "<?php?>",
			want:    "<?php?>",
			wantErr: nil,
		},
		{
			name: "html block: multiline processing instruction block",
			input: md(
				"<?",
				"hello",
				"?>",
			),
			want:    "<?\nhello\n?>",
			wantErr: nil,
		},
		{
			name: "html block: processing instruction terminator on opening line closes block immediately",
			input: md(
				"<?hello?>",
				"tail",
			),
			want:    "<?hello?><p>tail</p>",
			wantErr: nil,
		},
		{
			name: "html block: processing instruction block runs to eof when unterminated",
			input: md(
				"<?",
				"hello",
			),
			want:    "<?\nhello",
			wantErr: nil,
		},
		{
			name:    "html block: single line declaration block",
			input:   "<!DOCTYPE html>",
			want:    "<!DOCTYPE html>",
			wantErr: nil,
		},
		{
			name: "html block: multiline declaration block terminates on first greater-than",
			input: md(
				"<!DOCTYPE",
				"html>",
				"tail",
			),
			want:    "<!DOCTYPE\nhtml><p>tail</p>",
			wantErr: nil,
		},
		{
			name: "html block: declaration terminator on opening line closes block immediately",
			input: md(
				"<!DOCTYPE html>",
				"tail",
			),
			want:    "<!DOCTYPE html><p>tail</p>",
			wantErr: nil,
		},
		{
			name: "html block: declaration block runs to eof when unterminated",
			input: md(
				"<!DOCTYPE",
				"html",
			),
			want:    "<!DOCTYPE\nhtml",
			wantErr: nil,
		},
		{
			name:    "html block: named opening tag alone starts block",
			input:   "<div>",
			want:    "<div>",
			wantErr: nil,
		},
		{
			name:    "html block: named closing tag alone starts block",
			input:   "</div>",
			want:    "</div>",
			wantErr: nil,
		},
		{
			name:    "html block: named self-closing tag starts block",
			input:   "<div/>",
			want:    "<div/>",
			wantErr: nil,
		},
		{
			name:    "html block: named self-closing tag with space before closer starts block",
			input:   "<div / >",
			want:    "<div / >",
			wantErr: nil,
		},
		{
			name:    "html block: named tag with attributes starts block",
			input:   `<div class="x">`,
			want:    `<div class="x">`,
			wantErr: nil,
		},
		{
			name:    "html block: named tag with trailing text on same line still starts block",
			input:   "<div>hello",
			want:    "<div>hello",
			wantErr: nil,
		},
		{
			name: "html block: named tag block continues until blank line",
			input: md(
				"<div>",
				"hello",
				"world",
				"",
				"tail",
			),
			want:    "<div>\nhello\nworld<p>tail</p>",
			wantErr: nil,
		},
		{
			name: "html block: named tag block runs to eof without blank line",
			input: md(
				"<div>",
				"hello",
			),
			want:    "<div>\nhello",
			wantErr: nil,
		},
		{
			name: "html block: named tag block does not terminate on another tag line alone",
			input: md(
				"<div>",
				"</div>",
				"tail",
			),
			want:    "<div>\n</div>\ntail",
			wantErr: nil,
		},
		{
			name:    "html block: named tag with one leading space allowed",
			input:   " <div>",
			want:    " <div>",
			wantErr: nil,
		},
		{
			name:    "html block: named tag with two leading spaces allowed",
			input:   "  <div>",
			want:    "  <div>",
			wantErr: nil,
		},
		{
			name:    "html block: named tag with three leading spaces allowed",
			input:   "   <div>",
			want:    "   <div>",
			wantErr: nil,
		},
		{
			name:    "html block: named tag with four leading spaces is not html block",
			input:   "    <div>",
			want:    "<pre><code>&lt;div&gt;</code></pre>",
			wantErr: nil,
		},
		{
			name:    "html block: comment with three leading spaces allowed",
			input:   "   <!-- hello -->",
			want:    "   <!-- hello -->",
			wantErr: nil,
		},
		{
			name:    "html block: comment with four leading spaces is not html block",
			input:   "    <!-- hello -->",
			want:    "<pre><code>&lt;!-- hello --&gt;</code></pre>",
			wantErr: nil,
		},
		{
			name:    "html block: cdata with three leading spaces allowed",
			input:   "   <![CDATA[hello]]>",
			want:    "   <![CDATA[hello]]>",
			wantErr: nil,
		},
		{
			name:    "html block: processing instruction with three leading spaces allowed",
			input:   "    <?hello?>",
			want:    "<pre><code>&lt;?hello?&gt;</code></pre>",
			wantErr: nil,
		},
		{
			name:    "html block: declaration with three leading spaces allowed",
			input:   "   <!DOCTYPE html>",
			want:    "   <!DOCTYPE html>",
			wantErr: nil,
		},
		{
			name:    "html block: unknown named tag is not recognized as block html",
			input:   "<not-a-tag>",
			want:    "<p><not-a-tag></p>",
			wantErr: nil,
		},
		{
			name:    "html block: non-alpha tag opener is not recognized as block html",
			input:   "<1div>",
			want:    "<p>&lt;1div&gt;</p>",
			wantErr: nil,
		},
		{
			name:    "html block: non-alpha tag name is not recognized as block html",
			input:   "</1div>",
			want:    "<p>&lt;/1div&gt;</p>",
			wantErr: nil,
		},
		{
			name:    "html block: named tag without closing angle bracket is not html block",
			input:   "<div",
			want:    "<p>&lt;div</p>",
			wantErr: nil,
		},
		{
			name:    "html block: named closing tag without closing angle bracket is not html block",
			input:   "</div",
			want:    "<p>&lt;/div</p>",
			wantErr: nil,
		},
		{
			name:    "html block: named tag with invalid punctuation in head is not html block",
			input:   "<div!>",
			want:    "<p>&lt;div!&gt;</p>",
			wantErr: nil,
		},
		{
			name:    "html block: named tag with slash not followed by optional whitespace and closer is not html block",
			input:   "<div/x>",
			want:    "<p>&lt;div/x&gt;</p>",
			wantErr: nil,
		},
		{
			name:    "html block: named tag with attributes but no closing angle bracket is not html block",
			input:   `<div class="x"`,
			want:    `<p>&lt;div class=&#34;x&#34;</p>`,
			wantErr: nil,
		},
		{
			name:    "html block: non-html less-than text is not html block",
			input:   "< hello",
			want:    "<p>&lt; hello</p>",
			wantErr: nil,
		},
		{
			name:    "html block: bare less-than is not html block",
			input:   "<",
			want:    "<p>&lt;</p>",
			wantErr: nil,
		},
		{
			name: "html block: less-than followed by question mark without terminator still starts processing instruction block",
			input: md(
				"<?hello",
				"tail",
			),
			want:    "<?hello\ntail",
			wantErr: nil,
		},
		{
			name: "html block: less-than bang without greater-than still starts declaration block",
			input: md(
				"<!hello",
				"tail",
			),
			want:    "<!hello\ntail",
			wantErr: nil,
		},
		{
			name: "html block: named tag interrupted by blank line then paragraph",
			input: md(
				"<div>",
				"hello",
				"",
				"world",
			),
			want:    "<div>\nhello<p>world</p>",
			wantErr: nil,
		},
		{
			name: "html block: comment block may contain blank lines before terminator",
			input: md(
				"<!--",
				"",
				"hello",
				"-->",
				"tail",
			),
			want:    "<!--\n\nhello\n--><p>tail</p>",
			wantErr: nil,
		},
		{
			name: "html block: cdata block may contain blank lines before terminator",
			input: md(
				"<![CDATA[",
				"",
				"hello",
				"]]>",
				"tail",
			),
			want:    "<![CDATA[\n\nhello\n]]><p>tail</p>",
			wantErr: nil,
		},
		{
			name: "html block: processing instruction block may contain blank lines before terminator",
			input: md(
				"<?",
				"",
				"hello",
				"?>",
				"tail",
			),
			want:    "<?\n\nhello\n?><p>tail</p>",
			wantErr: nil,
		},
		{
			name: "html block: declaration block may contain blank lines before terminator",
			input: md(
				"<!DOCTYPE",
				"",
				"html>",
				"tail",
			),
			want:    "<!DOCTYPE\n\nhtml><p>tail</p>",
			wantErr: nil,
		},
		{
			name: "html block: named tag block may contain html-looking lines until blank line",
			input: md(
				"<div>",
				"<span>",
				"</span>",
				"",
				"tail",
			),
			want:    "<div>\n<span>\n</span><p>tail</p>",
			wantErr: nil,
		},
		{
			name: "html block: named closing tag block continues until blank line",
			input: md(
				"</div>",
				"hello",
				"",
				"tail",
			),
			want:    "</div>\nhello<p>tail</p>",
			wantErr: nil,
		},
		{
			name:    "html block: whitelisted tag name is case-insensitive",
			input:   "<DIV>",
			want:    "<DIV>",
			wantErr: nil,
		},
		{
			name:    "html block: closing whitelisted tag name is case-insensitive",
			input:   "</DIV>",
			want:    "</DIV>",
			wantErr: nil,
		},
		{
			name:    "html block: alphanumeric tag name allowed when whitelisted",
			input:   "<h1>",
			want:    "<h1>",
			wantErr: nil,
		},
		{
			name:    "html block: non-whitelisted alphanumeric tag name rejected",
			input:   "<x1>",
			want:    "<p><x1></p>",
			wantErr: nil,
		},

		// escapes

		// 		{
		// 	name:  "escape: escaped emphasis opener remains literal text",
		// 	input: "\\*a*",
		// },
		// {
		// 	name:  "escape: escaped emphasis closer remains literal text",
		// 	input: "*a\\*",
		// },
		// {
		// 	name:  "escape: escaped strong delimiter remains literal text",
		// 	input: "\\**a**",
		// },
		// {
		// 	name:  "escape: escaped underscore remains literal text",
		// 	input: "\\_a_",
		// },
		// {
		// 	name:  "escape: escaped backtick remains literal text",
		// 	input: "\\`code`",
		// },
		// {
		// 	name:  "escape: escaped opening bracket prevents link formation",
		// 	input: "\\[label](/url)",
		// },
		// {
		// 	name:  "escape: escaped closing bracket prevents link formation",
		// 	input: "[label\\](/url)",
		// },
		// {
		// 	name:  "escape: escaped opening parenthesis prevents inline link tail",
		// 	input: "[label]\\(/url)",
		// },
		// {
		// 	name:  "escape: escaped closing parenthesis remains literal text",
		// 	input: "[label](/url\\)",
		// },
		// {
		// 	name:  "escape: escaped opening angle remains literal text",
		// 	input: "\\<span>",
		// },
		// {
		// 	name:  "escape: escaped closing angle remains literal text",
		// 	input: "<span\\>",
		// },
		// {
		// 	name:  "escape: escaped bang prevents image formation",
		// 	input: "\\![alt](/img.png)",
		// },
		// {
		// 	name:  "escape: escaped image opener becomes literal punctuation",
		// 	input: "!\\[alt](/img.png)",
		// },
		// {
		// 	name:  "escape: escaped delimiter inside emphasis is literal text",
		// 	input: "*a \\* b*",
		// },
		// {
		// 	name:  "escape: escaped delimiter inside strong is literal text",
		// 	input: "**a \\* b**",
		// },
		// {
		// 	name:  "escape: escaped backticks prevent code span formation",
		// 	input: "\\`code\\`",
		// },
		// {
		// 	name:  "escape: escaped brackets are literal inside text",
		// 	input: "\\[a\\]",
		// },
		// {
		// 	name:  "escape: escaped parentheses are literal inside text",
		// 	input: "\\(a\\)",
		// },
		// {
		// 	name:  "escape: escaped angle brackets are literal inside text",
		// 	input: "\\<a\\>",
		// },
		// {
		// 	name:  "escape: escaped backslash yields literal backslash",
		// 	input: "\\\\",
		// },
		// {
		// 	name:  "escape: trailing backslash remains literal text",
		// 	input: "\\",
		// },
		// {
		// 	name:  "escape: backslash before ordinary text remains literal",
		// 	input: "\\a",
		// },
		// {
		// 	name:  "escape: escaped delimiter does not prevent later valid emphasis",
		// 	input: "\\*a* *b*",
		// },
		// {
		// 	name:  "escape: escaped opener leaves following link syntax literal",
		// 	input: "\\[x](y)",
		// },
		// {
		// 	name:  "escape: escaped bang leaves following bracket construct as link syntax",
		// 	input: "\\![x](y)",
		// },
		// {
		// 	name:  "escape: escaped punctuation is preserved in surrounding text",
		// 	input: "a \\* b",
		// },

		// reference links and images

		// 		{
		// 	name:  "reference link: full reference resolves to link",
		// 	input: "[label][ref]\n\n[ref]: /url",
		// },
		// {
		// 	name:  "reference link: collapsed reference resolves to link",
		// 	input: "[label][]\n\n[label]: /url",
		// },
		// {
		// 	name:  "reference link: shortcut reference resolves to link",
		// 	input: "[label]\n\n[label]: /url",
		// },
		// {
		// 	name:  "reference link: full reference may use empty label",
		// 	input: "[label][]\n\n[]: /url",
		// },
		// {
		// 	name:  "reference link: full reference uses referenced label rather than visible label",
		// 	input: "[visible][ref]\n\n[ref]: /url",
		// },
		// {
		// 	name:  "reference link: reference definition may provide title",
		// 	input: "[label][ref]\n\n[ref]: /url \"title\"",
		// },
		// {
		// 	name:  "reference link: collapsed reference may use definition title",
		// 	input: "[label][]\n\n[label]: /url \"title\"",
		// },
		// {
		// 	name:  "reference link: shortcut reference may use definition title",
		// 	input: "[label]\n\n[label]: /url \"title\"",
		// },
		// {
		// 	name:  "reference link: nested inline content is allowed in label",
		// 	input: "[a *b* c][ref]\n\n[ref]: /url",
		// },
		// {
		// 	name:  "reference link: code span is allowed in label",
		// 	input: "[a `b` c][ref]\n\n[ref]: /url",
		// },
		// {
		// 	name:  "reference link: image is allowed in label",
		// 	input: "[![alt](/img.png)][ref]\n\n[ref]: /url",
		// },
		// {
		// 	name:  "reference link: surrounding text is preserved",
		// 	input: "a [label][ref] b\n\n[ref]: /url",
		// },
		// {
		// 	name:  "reference link: full reference falls back to literal text when definition is missing",
		// 	input: "[label][ref]",
		// },
		// {
		// 	name:  "reference link: collapsed reference falls back to literal text when definition is missing",
		// 	input: "[label][]",
		// },
		// {
		// 	name:  "reference link: shortcut reference falls back to literal text when definition is missing",
		// 	input: "[label]",
		// },
		// {
		// 	name:  "reference link: full reference falls back to literal text when closing bracket is missing",
		// 	input: "[label][ref",
		// },
		// {
		// 	name:  "reference link: full reference falls back to literal text when label is invalid",
		// 	input: "[label][a\nb]\n\n[a b]: /url",
		// },
		// {
		// 	name:  "reference link: collapsed reference takes precedence over shortcut reference",
		// 	input: "[label][]\n\n[label]: /url",
		// },
		// {
		// 	name:  "reference link: full reference takes precedence over shortcut reference",
		// 	input: "[label][ref]\n\n[label]: /wrong\n[ref]: /right",
		// },
		// {
		// 	name:  "reference link: nested links are rejected in reference form",
		// 	input: "[outer [inner][in]][out]\n\n[in]: /in\n[out]: /out",
		// },
		// {
		// 	name:  "reference link: inner reference link may still resolve when outer link is rejected",
		// 	input: "[outer [inner][in]]\n\n[in]: /in",
		// },
		//
		// {
		// 	name:  "reference image: full reference resolves to image",
		// 	input: "![alt][ref]\n\n[ref]: /img.png",
		// },
		// {
		// 	name:  "reference image: collapsed reference resolves to image",
		// 	input: "![alt][]\n\n[alt]: /img.png",
		// },
		// {
		// 	name:  "reference image: shortcut reference resolves to image",
		// 	input: "![alt]\n\n[alt]: /img.png",
		// },
		// {
		// 	name:  "reference image: full reference may use empty label",
		// 	input: "![alt][]\n\n[]: /img.png",
		// },
		// {
		// 	name:  "reference image: full reference uses referenced label rather than visible label",
		// 	input: "![visible][ref]\n\n[ref]: /img.png",
		// },
		// {
		// 	name:  "reference image: reference definition may provide title",
		// 	input: "![alt][ref]\n\n[ref]: /img.png \"title\"",
		// },
		// {
		// 	name:  "reference image: collapsed reference may use definition title",
		// 	input: "![alt][]\n\n[alt]: /img.png \"title\"",
		// },
		// {
		// 	name:  "reference image: shortcut reference may use definition title",
		// 	input: "![alt]\n\n[alt]: /img.png \"title\"",
		// },
		// {
		// 	name:  "reference image: nested inline content is allowed in label",
		// 	input: "![a *b* c][ref]\n\n[ref]: /img.png",
		// },
		// {
		// 	name:  "reference image: code span is allowed in label",
		// 	input: "![a `b` c][ref]\n\n[ref]: /img.png",
		// },
		// {
		// 	name:  "reference image: link is allowed in image label",
		// 	input: "![see [this](/url)][ref]\n\n[ref]: /img.png",
		// },
		// {
		// 	name:  "reference image: surrounding text is preserved",
		// 	input: "a ![alt][ref] b\n\n[ref]: /img.png",
		// },
		// {
		// 	name:  "reference image: full reference falls back to literal text when definition is missing",
		// 	input: "![alt][ref]",
		// },
		// {
		// 	name:  "reference image: collapsed reference falls back to literal text when definition is missing",
		// 	input: "![alt][]",
		// },
		// {
		// 	name:  "reference image: shortcut reference falls back to literal text when definition is missing",
		// 	input: "![alt]",
		// },
		// {
		// 	name:  "reference image: full reference falls back to literal text when closing bracket is missing",
		// 	input: "![alt][ref",
		// },
		// {
		// 	name:  "reference image: full reference falls back to literal text when label is invalid",
		// 	input: "![alt][a\nb]\n\n[a b]: /img.png",
		// },
		// {
		// 	name:  "reference image: collapsed reference takes precedence over shortcut reference",
		// 	input: "![alt][]\n\n[alt]: /img.png",
		// },
		// {
		// 	name:  "reference image: nested images are allowed in reference form",
		// 	input: "![outer ![inner][in]][out]\n\n[in]: /in.png\n[out]: /out.png",
		// },
		// {
		// 	name:  "reference image: full reference takes precedence over shortcut reference",
		// 	input: "![alt][ref]\n\n[alt]: /wrong.png\n[ref]: /right.png",
		// },

		// precedence and ambiguity

		// 		{
		// 	name:  "precedence: code span suppresses emphasis parsing",
		// 	input: "*a `*`*",
		// },
		// {
		// 	name:  "precedence: code span suppresses link parsing",
		// 	input: "[`[x](y)`](/url)",
		// },
		// {
		// 	name:  "precedence: code span suppresses image parsing",
		// 	input: "`![alt](/img.png)`",
		// },
		// {
		// 	name:  "precedence: code span suppresses raw html parsing",
		// 	input: "`<span>`",
		// },
		//
		// 		{
		// 	name:  "precedence: uri autolink takes precedence over raw html fallback",
		// 	input: "<https://example.com>",
		// },
		// {
		// 	name:  "precedence: email autolink takes precedence over raw html fallback",
		// 	input: "<user@example.com>",
		// },
		// {
		// 	name:  "precedence: raw html resolves when angle construct is not an autolink",
		// 	input: "<span>",
		// },
		// {
		// 	name:  "precedence: invalid angle construct falls back to literal text",
		// 	input: "<local@domain>",
		// },
		//
		// 		{
		// 	name:  "precedence: emphasis is parsed inside link label",
		// 	input: "[a *b* c](/url)",
		// },
		// {
		// 	name:  "precedence: strong emphasis is parsed inside image label",
		// 	input: "![a **b** c](/img.png)",
		// },
		// {
		// 	name:  "precedence: nested link is rejected inside link label",
		// 	input: "[outer [inner](/in)](/out)",
		// },
		// {
		// 	name:  "precedence: link is allowed inside image label",
		// 	input: "![see [this](/url)](/img.png)",
		// },
		//
		// 		{
		// 	name:  "precedence: escape prevents emphasis from binding",
		// 	input: "\\*a*",
		// },
		// {
		// 	name:  "precedence: escape prevents link formation",
		// 	input: "\\[x](y)",
		// },
		// {
		// 	name:  "precedence: escape prevents image formation",
		// 	input: "\\![x](y)",
		// },
		// {
		// 	name:  "precedence: escape prevents html recognition from opening angle bracket",
		// 	input: "\\<span>",
		// },
		//
		// 		{
		// 	name:  "precedence: emphasis may wrap a link",
		// 	input: "*[x](/url)*",
		// },
		// {
		// 	name:  "precedence: strong emphasis may wrap an image",
		// 	input: "**![alt](/img.png)**",
		// },
		// {
		// 	name:  "precedence: link label may contain code span and emphasis together",
		// 	input: "[a `b` *c*](/url)",
		// },
		// {
		// 	name:  "precedence: code span prevents delimiter participation inside emphasis run",
		// 	input: "*a `*` b*",
		// },
		//
		// 		{
		// 	name:  "precedence: failed inline link falls back without preventing later emphasis",
		// 	input: "[x](a(b) *c*",
		// },
		// {
		// 	name:  "precedence: failed autolink falls back to text and allows later emphasis",
		// 	input: "<local@domain> *x*",
		// },
		// {
		// 	name:  "precedence: failed raw html falls back to text and allows later link",
		// 	input: "<1tag> [x](/url)",
		// },

		// 		{
		// 	name:  "composite: emphasis code span and link may coexist in one paragraph",
		// 	input: "a *b* `c` [d](/url) e",
		// },
		// {
		// 	name:  "composite: image link and autolink may coexist in one paragraph",
		// 	input: "![alt](/img.png) [x](/url) <https://example.com>",
		// },
		// {
		// 	name:  "composite: escaped punctuation does not disrupt neighboring inline constructs",
		// 	input: "\\* a *b* [c](/url) `d`",
		// },
		//
		// 		{
		// 	name:  "adjacency: consecutive links parse independently",
		// 	input: "[a](/x)[b](/y)",
		// },
		// {
		// 	name:  "adjacency: consecutive emphasis runs parse independently",
		// 	input: "*a**b*",
		// },
		// {
		// 	name:  "adjacency: code span adjacent to link parses independently",
		// 	input: "`a`[b](/url)",
		// },
		// {
		// 	name:  "adjacency: html adjacent to autolink parses independently",
		// 	input: "<span><https://example.com>",
		// },
		//
		// 		{
		// 	name:  "nesting: emphasis inside link inside emphasis resolves correctly",
		// 	input: "*[a **b** c](/url)*",
		// },
		// {
		// 	name:  "nesting: image inside link inside emphasis resolves correctly",
		// 	input: "*[![alt](/img.png)](/url)*",
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := HTML(tc.input)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func md(xs ...string) string {
	return strings.Join(xs, "\n")
}
