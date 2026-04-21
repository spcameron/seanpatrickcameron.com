package markdown

import (
	"strings"
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestCompile_EndToEnd(t *testing.T) {
	testCases := []struct {
		name     string
		markdown string
		wantHTML string
		wantErr  error
	}{
		// paragraphs

		{
			name:     "paragraph: plain single line",
			markdown: "hello world",
			wantHTML: `<p>hello world</p>`,
			wantErr:  nil,
		},
		{
			name: "paragraph: multiple lines join into one paragraph",
			markdown: md(
				"hello",
				"world",
			),
			wantHTML: `<p>hello world</p>`,
			wantErr:  nil,
		},
		{
			name: "paragraph: blank line separates paragraph",
			markdown: md(
				"hello",
				"",
				"world",
			),
			wantHTML: `<p>hello</p><p>world</p>`,
			wantErr:  nil,
		},
		{
			name: "paragaph: whitespace only line separates paragraph",
			markdown: md(
				"hello",
				"    ",
				"world",
			),
			wantHTML: `<p>hello</p><p>world</p>`,
			wantErr:  nil,
		},
		{
			name: "paragraph: leading blank lines are ignored",
			markdown: md(
				"",
				"",
				"hello world",
			),
			wantHTML: `<p>hello world</p>`,
			wantErr:  nil,
		},
		{
			name: "paragraph: trailing blank lines are ignored",
			markdown: md(
				"hello world",
				"",
				"",
			),
			wantHTML: `<p>hello world</p>`,
			wantErr:  nil,
		},
		{
			name: "paragraph: indented line continues paragraph",
			markdown: md(
				"hello",
				"    world",
			),
			wantHTML: `<p>hello     world</p>`,
			wantErr:  nil,
		},
		{
			name: "paragraph: ends before atx heading",
			markdown: md(
				"hello",
				"# heading",
			),
			wantHTML: `<p>hello</p><h1>heading</h1>`,
			wantErr:  nil,
		},
		{
			name: "paragraph: ends before block quote",
			markdown: md(
				"hello",
				"> world",
			),
			wantHTML: `<p>hello</p><blockquote><p>world</p></blockquote>`,
			wantErr:  nil,
		},
		{
			name: "paragraph: ends before unordered list",
			markdown: md(
				"hello",
				"- world",
			),
			wantHTML: `<p>hello</p><ul><li>world</li></ul>`,
			wantErr:  nil,
		},
		{
			name: "paragraph: ends before ordered list",
			markdown: md(
				"hello",
				"1. world",
			),
			wantHTML: `<p>hello</p><ol><li>world</li></ol>`,
			wantErr:  nil,
		},
		{
			name: "paragraph: hard break via two trailing spaces",
			markdown: md(
				"hello  ",
				"world",
			),
			wantHTML: `<p>hello<br>world</p>`,
			wantErr:  nil,
		},
		{
			name: "paragraph: hard break via trailing backslash",
			markdown: md(
				"hello\\",
				"world",
			),
			wantHTML: `<p>hello<br>world</p>`,
			wantErr:  nil,
		},

		// ATX headers

		{
			name:     "atx heading: level 1 plain text",
			markdown: "# hello",
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: level 2 plain text",
			markdown: "## hello",
			wantHTML: `<h2>hello</h2>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: level 3 plain text",
			markdown: "### hello",
			wantHTML: `<h3>hello</h3>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: level 4 plain text",
			markdown: "#### hello",
			wantHTML: `<h4>hello</h4>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: level 5 plain text",
			markdown: "##### hello",
			wantHTML: `<h5>hello</h5>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: level 6 plain text",
			markdown: "###### hello",
			wantHTML: `<h6>hello</h6>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: empty content after required delimiter",
			markdown: "# ",
			wantHTML: `<h1></h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: empty content after multiple spaces",
			markdown: "#   ",
			wantHTML: `<h1></h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: empty context after tab delimiter",
			markdown: "#\t",
			wantHTML: `<h1></h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: level 1 empty heading at end of line",
			markdown: "#",
			wantHTML: `<h1></h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: level 6 empty heading at end of line",
			markdown: "######",
			wantHTML: `<h6></h6>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: multiple spaces after marker allowed",
			markdown: "#   hello",
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: tab after marker allowed",
			markdown: "#\thello",
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: spaces and tabs after marker are consumed before content",
			markdown: "# \t   hello",
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: content may contain internal hash characters",
			markdown: "# hello # world",
			wantHTML: `<h1>hello # world</h1>`,
			wantErr:  nil,
		},
		// NOTE: subject to change upon implementing trim ATX closers
		{
			name:     "atx heading: trailing closing marker run is trimmed",
			markdown: "# hello ###",
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: content may be empty after trimming closing marker run",
			markdown: "# ###",
			wantHTML: `<h1></h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: trailing spaces are trimmed from content",
			markdown: "# hello   ",
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: trailing tabs are trimmed from content",
			markdown: "# hello\t\t",
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: trailing whitespace after closing marker run is trimmed",
			markdown: "# hello ###   ",
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: single closing marker is trimmed",
			markdown: "# hello #",
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: closing marker run need not match opening length",
			markdown: "##### hello ##",
			wantHTML: `<h5>hello</h5>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: closing marker run may be separated by multiple spaces",
			markdown: "###   bar    ###",
			wantHTML: `<h3>bar</h3>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: closing marker run may be followed by spaces",
			markdown: "### hello ###     ",
			wantHTML: `<h3>hello</h3>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: closing marker run may be followed by tabs",
			markdown: "### hello ###\t\t",
			wantHTML: `<h3>hello</h3>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: hash run without separating whitespace remains content",
			markdown: "# hello###",
			wantHTML: `<h1>hello###</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: single trailing hash without separating whitespace remains content",
			markdown: "# hello#",
			wantHTML: `<h1>hello#</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: hash run with non-whitespace following remains content",
			markdown: "### hello ### b",
			wantHTML: `<h3>hello ### b</h3>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: escaped closing marker run remains content",
			markdown: "### hello \\###",
			wantHTML: `<h3>hello ###</h3>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: partially escaped trailing hash run remains content",
			markdown: "## hello #\\##",
			wantHTML: `<h2>hello ###</h2>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: escaped single trailing hash remains content",
			markdown: "# hello \\#",
			wantHTML: `<h1>hello #</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: interior hashes before valid closing run remain content",
			markdown: "### foo # bar ###",
			wantHTML: `<h3>foo # bar</h3>`,
			wantErr:  nil,
		},
		// NOTE: end trim ATX closers cases
		{
			name:     "atx heading: leading indentation of one space allowed",
			markdown: " # hello",
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: leading indentation of two spaces allowed",
			markdown: "  # hello",
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: leading indentation of three spaces allowed",
			markdown: "   # hello",
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: leading indentation of four spaces is not a heading",
			markdown: "    # hello",
			wantHTML: `<pre><code># hello</code></pre>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: marker must be followed by space not plain text",
			markdown: "#hello",
			wantHTML: `<p>#hello</p>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: seven markers is not a heading",
			markdown: "####### hello",
			wantHTML: `<p>####### hello</p>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: non-hash opener is not a heading",
			markdown: "hello",
			wantHTML: `<p>hello</p>`,
			wantErr:  nil,
		},
		{
			name: "atx heading: heading may follow blank line",
			markdown: md(
				"",
				"# hello",
			),
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name: "atx heading: consecutive headings of different levels",
			markdown: md(
				"# one",
				"## two",
				"### three",
			),
			wantHTML: `<h1>one</h1><h2>two</h2><h3>three</h3>`,
			wantErr:  nil,
		},
		{
			name:     "atx heading: escaped hash in content remains content",
			markdown: "# \\# hello",
			wantHTML: `<h1># hello</h1>`,
			wantErr:  nil,
		},

		// Setext headers

		{
			name: "setext heading: level 1 plain text with equals underline",
			markdown: md(
				"hello",
				"=====",
			),
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: level 2 plain text with dash underline",
			markdown: md(
				"hello",
				"-----",
			),
			wantHTML: `<h2>hello</h2>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: single equals marker is valid underline",
			markdown: md(
				"hello",
				"=",
			),
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: single dash marker is valid underline",
			markdown: md(
				"hello",
				"-",
			),
			wantHTML: `<h2>hello</h2>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: underline may be longer than content",
			markdown: md(
				"hi",
				"==========",
			),
			wantHTML: `<h1>hi</h1>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: dash underline may be longer than content",
			markdown: md(
				"hi",
				"----------",
			),
			wantHTML: `<h2>hi</h2>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: leading indentation of one space on underline allowed",
			markdown: md(
				"hello",
				" =====",
			),
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: leading indentation of two spaces on underline allowed",
			markdown: md(
				"hello",
				"  =====",
			),
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: leading indentation of three spaces on underline allowed",
			markdown: md(
				"hello",
				"   =====",
			),
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: leading indentation of four spaces on underline is not valid",
			markdown: md(
				"hello",
				"    =====",
			),
			wantHTML: `<p>hello     =====</p>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: trailing spaces after equals underline allowed",
			markdown: md(
				"hello",
				"=====   ",
			),
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: trailing tabs after equals underline allowed",
			markdown: md(
				"hello",
				"=====\t\t",
			),
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: trailing mixed spaces and tabs after underline allowed",
			markdown: md(
				"hello",
				"===== \t \t",
			),
			wantHTML: `<h1>hello</h1>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: content line may contain punctuation",
			markdown: md(
				"hello, world",
				"-----",
			),
			wantHTML: `<h2>hello, world</h2>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: content line may contain hash characters",
			markdown: md(
				"hello # world",
				"-----",
			),
			wantHTML: `<h2>hello # world</h2>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: content line may contain inline markers",
			markdown: md(
				"hello *world*",
				"-----",
			),
			wantHTML: `<h2>hello <em>world</em></h2>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: content line may contain leading and trailing spaces",
			markdown: md(
				" hello world ",
				"-----",
			),
			wantHTML: `<h2> hello world </h2>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: blank line between content and underline prevents heading",
			markdown: md(
				"hello",
				"",
				"-----",
			),
			wantHTML: `<p>hello</p><hr>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: underline line alone at document start is not a heading",
			markdown: md(
				"-----",
			),
			wantHTML: `<hr>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: equals underline alone at document start is not a heading",
			markdown: md(
				"=====",
			),
			wantHTML: `<p>=====</p>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: mixed equals and dash markers are not valid",
			markdown: md(
				"hello",
				"=-=-=",
			),
			wantHTML: `<p>hello =-=-=</p>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: mixed dash and equals markers are not valid",
			markdown: md(
				"hello",
				"-=-=-",
			),
			wantHTML: `<p>hello -=-=-</p>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: internal spaces between markers are not valid",
			markdown: md(
				"hello",
				"= = =",
			),
			wantHTML: `<p>hello = = =</p>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: internal tabs between markers are not valid",
			markdown: md(
				"hello",
				"=\t=\t=",
			),
			wantHTML: `<p>hello =	=	=</p>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: trailing non-whitespace after equals markers is not valid",
			markdown: md(
				"hello",
				"=====x",
			),
			wantHTML: `<p>hello =====x</p>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: trailing non-whitespace after dash markers is not valid",
			markdown: md(
				"hello",
				"-----x",
			),
			wantHTML: `<p>hello -----x</p>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: leading non-marker text is not valid underline",
			markdown: md(
				"hello",
				"x----",
			),
			wantHTML: `<p>hello x----</p>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: non-whitespace mixed into underline is not valid",
			markdown: md(
				"hello",
				"--x--",
			),
			wantHTML: `<p>hello --x--</p>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: level 2 underline disambiguates from thematic break after paragraph text",
			markdown: md(
				"hello",
				"---",
			),
			wantHTML: `<h2>hello</h2>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: heading text must immediately precede underline",
			markdown: md(
				"hello",
				"",
				"world",
				"-----",
			),
			wantHTML: `<p>hello</p><h2>world</h2>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: multiline content normalizes newline to space",
			markdown: md(
				"hello",
				"world",
				"-----",
			),
			wantHTML: `<h2>hello world</h2>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: multiline content preserves hard break via trailing backslash",
			markdown: md(
				"hello\\",
				"world",
				"-----",
			),
			wantHTML: `<h2>hello<br>world</h2>`,
			wantErr:  nil,
		},
		{
			name: "setext heading: multiline content preserves hard break via trailing spaces",
			markdown: md(
				"hello  ",
				"world",
				"-----",
			),
			wantHTML: `<h2>hello<br>world</h2>`,
			wantErr:  nil,
		},

		// thematic breaks

		{
			name:     "thematic break: three hyphens",
			markdown: "---",
			wantHTML: `<hr>`,
			wantErr:  nil,
		},
		{
			name:     "thematic break: three asterisks",
			markdown: "***",
			wantHTML: `<hr>`,
			wantErr:  nil,
		},
		{
			name:     "thematic break: three underscores",
			markdown: "___",
			wantHTML: `<hr>`,
			wantErr:  nil,
		},
		{
			name:     "thematic break: spaces between markers allowed",
			markdown: "- - -",
			wantHTML: `<hr>`,
			wantErr:  nil,
		},
		{
			name:     "thematic break: tabs between markers allowed",
			markdown: "-\t-\t-",
			wantHTML: `<hr>`,
			wantErr:  nil,
		},
		{
			name:     "thematic break: trailing spaces allowed",
			markdown: "---   ",
			wantHTML: `<hr>`,
			wantErr:  nil,
		},
		{
			name:     "thematic break: leading indentation of three spaces allowed",
			markdown: "   ---",
			wantHTML: `<hr>`,
			wantErr:  nil,
		},
		{
			name:     "thematic break: leading indentation of four spaces is not thematic break",
			markdown: "    ---",
			wantHTML: `<pre><code>---</code></pre>`,
			wantErr:  nil,
		},
		{
			name:     "thematic break: exactly two markers is not thematic break",
			markdown: "--",
			wantHTML: `<p>--</p>`,
			wantErr:  nil,
		},
		{
			name:     "thematic break: mixed marker families are not allowed",
			markdown: "-*-",
			wantHTML: `<p>-*-</p>`,
			wantErr:  nil,
		},
		{
			name:     "thematic break: internal non-whitespace invalidates line",
			markdown: "--x--",
			wantHTML: `<p>--x--</p>`,
			wantErr:  nil,
		},
		{
			name:     "thematic break: leading text invalidates line",
			markdown: "x---",
			wantHTML: `<p>x---</p>`,
			wantErr:  nil,
		},
		{
			name: "thematic break: between paragraphs",
			markdown: md(
				"hello",
				"",
				"---",
				"",
				"world",
			),
			wantHTML: `<p>hello</p><hr><p>world</p>`,
			wantErr:  nil,
		},
		{
			name: "thematic break: dash line after paragraph text becomes setext heading not thematic break",
			markdown: md(
				"hello",
				"---",
			),
			wantHTML: `<h2>hello</h2>`,
			wantErr:  nil,
		},
		{
			name: "thematic break: dash line after multiline paragraph text becomes setext heading not thematic break",
			markdown: md(
				"hello",
				"world",
				"---",
			),
			wantHTML: `<h2>hello world</h2>`,
			wantErr:  nil,
		},

		// block quotes

		{
			name:     "block quote: single quoted paragraph line",
			markdown: "> hello",
			wantHTML: `<blockquote><p>hello</p></blockquote>`,
			wantErr:  nil,
		},
		{
			name: "block quote: multiple quoted lines form one paragraph",
			markdown: md(
				"> hello",
				"> world",
			),
			wantHTML: `<blockquote><p>hello world</p></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: marker without following space is allowed",
			markdown: ">hello",
			wantHTML: `<blockquote><p>hello</p></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: marker with following tab is allowed",
			markdown: ">\thello",
			wantHTML: `<blockquote><p>hello</p></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: leading indentation of two spaces allowed",
			markdown: `  > hello`,
			wantHTML: `<blockquote><p>hello</p></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: leading indentation of three spaces allowed",
			markdown: `   > hello`,
			wantHTML: `<blockquote><p>hello</p></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: leading indentation of four spaces is not a block quote",
			markdown: `    > hello`,
			wantHTML: `<pre><code>&gt; hello</code></pre>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: empty quoted line is allowed",
			markdown: ">",
			wantHTML: `<blockquote></blockquote>`,
			wantErr:  nil,
		},
		{
			name: "block quote: quoted blank line separates inner paragraphs",
			markdown: md(
				"> hello",
				">",
				"> world",
			),
			wantHTML: `<blockquote><p>hello</p><p>world</p></blockquote>`,
			wantErr:  nil,
		},
		{
			name: "block quote: non-quoted following line ends block quote",
			markdown: md(
				"> hello",
				"world",
			),
			wantHTML: `<blockquote><p>hello</p></blockquote><p>world</p>`,
			wantErr:  nil,
		},
		{
			name: "block quote: blank physical line ends block quote",
			markdown: md(
				"> hello",
				"",
				"> world",
			),
			wantHTML: `<blockquote><p>hello</p></blockquote><blockquote><p>world</p></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: nested quote with double marker",
			markdown: ">> hello",
			wantHTML: `<blockquote><blockquote><p>hello</p></blockquote></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: nested quote with space between markers",
			markdown: "> > hello",
			wantHTML: `<blockquote><blockquote><p>hello</p></blockquote></blockquote>`,
			wantErr:  nil,
		},
		{
			name: "block quote: nested quote across multiple lines",
			markdown: md(
				"> > hello",
				"> > world",
			),
			wantHTML: `<blockquote><blockquote><p>hello world</p></blockquote></blockquote>`,
		},
		{
			name: "block quote: mixed nesting depths across lines",
			markdown: md(
				"> outer",
				"> > inner",
				"> outer again",
			),
			wantHTML: `<blockquote><p>outer</p><blockquote><p>inner</p></blockquote><p>outer again</p></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: triple nesting",
			markdown: "> > > hello",
			wantHTML: `<blockquote><blockquote><blockquote><p>hello</p></blockquote></blockquote></blockquote>`,
			wantErr:  nil,
		},
		{
			name: "block quote: inner paragraph after nested blank line",
			markdown: md(
				"> > hello",
				"> >",
				"> > world",
			),
			wantHTML: `<blockquote><blockquote><p>hello</p><p>world</p></blockquote></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: quoted atx heading",
			markdown: "> # hello",
			wantHTML: `<blockquote><h1>hello</h1></blockquote>`,
			wantErr:  nil,
		},
		{
			name: "block quote: quoted setext heading",
			markdown: md(
				"> hello",
				"> -----",
			),
			wantHTML: `<blockquote><h2>hello</h2></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: quoted thematic break",
			markdown: "> ---",
			wantHTML: `<blockquote><hr></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: quoted unordered list item",
			markdown: "> - item",
			wantHTML: `<blockquote><ul><li>item</li></ul></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: quoted ordered list item",
			markdown: "> 1. item",
			wantHTML: "<blockquote><ol><li>item</li></ol></blockquote>",
			wantErr:  nil,
		},
		{
			name: "block quote: quoted fenced code block",
			markdown: md(
				"> ```",
				"> code",
				"> ```",
			),
			wantHTML: `<blockquote><pre><code>code</code></pre></blockquote>`,
			wantErr:  nil,
		},
		{
			name: "block quote: quoted indented code block after quoted blank line",
			markdown: md(
				">",
				">     code",
			),
			wantHTML: `<blockquote><pre><code>code</code></pre></blockquote>`,
			wantErr:  nil,
		},
		{
			name: "block quote: nested block quote contains heading and paragraph",
			markdown: md(
				"> # title",
				">",
				"> body",
			),
			wantHTML: `<blockquote><h1>title</h1><p>body</p></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: line with only marker and delimiter is empty quote line",
			markdown: "> ",
			wantHTML: `<blockquote></blockquote>`,
			wantErr:  nil,
		},
		{
			name:     "block quote: line with only marker and tab delimiter is empty quote line",
			markdown: ">\t",
			wantHTML: `<blockquote></blockquote>`,
			wantErr:  nil,
		},

		// unordered lists

		{
			name:     "unordered list: single hyphen item",
			markdown: "- item",
			wantHTML: `<ul><li>item</li></ul>`,
			wantErr:  nil,
		},
		{
			name:     "unordered list: single asterisk item",
			markdown: "* item",
			wantHTML: `<ul><li>item</li></ul>`,
			wantErr:  nil,
		},
		{
			name:     "unordered list: single plus item",
			markdown: "+ item",
			wantHTML: `<ul><li>item</li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: multiple sibling hyphen items",
			markdown: md(
				"- one",
				"- two",
				"- three",
			),
			wantHTML: `<ul><li>one</li><li>two</li><li>three</li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: multiple sibling asterisk items",
			markdown: md(
				"* one",
				"* two",
				"* three",
			),
			wantHTML: `<ul><li>one</li><li>two</li><li>three</li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: multiple sibling plus items",
			markdown: md(
				"+ one",
				"+ two",
				"+ three",
			),
			wantHTML: `<ul><li>one</li><li>two</li><li>three</li></ul>`,
			wantErr:  nil,
		},
		{
			name:     "unordered list: marker requires following tab or space",
			markdown: "-item",
			wantHTML: `<p>-item</p>`,
			wantErr:  nil,
		},
		{
			name:     "unordered list: tab after marker allowed",
			markdown: "-\titem",
			wantHTML: `<ul><li>item</li></ul>`,
			wantErr:  nil,
		},
		{
			name:     "unordered list: multiple spaces after marker allowed",
			markdown: "-   item",
			wantHTML: `<ul><li>item</li></ul>`,
			wantErr:  nil,
		},
		{
			name:     "unordered list: multiple tabs and spaces after marker allowed",
			markdown: "- \t  item",
			wantHTML: `<ul><li>item</li></ul>`,
			wantErr:  nil,
		},
		{
			name:     "unordered list: empty item content after required delimiter",
			markdown: "- ",
			wantHTML: `<ul><li></li></ul>`,
			wantErr:  nil,
		},
		{
			name:     "unordered list: leading indentation of one space allowed",
			markdown: " - item",
			wantHTML: `<ul><li>item</li></ul>`,
			wantErr:  nil,
		},
		{
			name:     "unordered list: leading indentation of two spaces allowed",
			markdown: "  - item",
			wantHTML: `<ul><li>item</li></ul>`,
			wantErr:  nil,
		},
		{
			name:     "unordered list: leading indentation of three spaces allowed",
			markdown: "   - item",
			wantHTML: `<ul><li>item</li></ul>`,
			wantErr:  nil,
		},
		{
			name:     "unordered list: leading indentation of four spaces is not a list",
			markdown: "    - item",
			wantHTML: `<pre><code>- item</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: continuation line at content baseline stays in item",
			markdown: md(
				"- one",
				"  two",
			),
			wantHTML: `<ul><li>one two</li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: continuation line beyond content baseline stays in item",
			markdown: md(
				"- one",
				"    two",
			),
			wantHTML: `<ul><li>one   two</li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: dedented nonblank line ends list",
			markdown: md(
				"- one",
				"two",
			),
			wantHTML: `<ul><li>one</li></ul><p>two</p>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: two sibling items form tight list",
			markdown: md(
				"- one",
				"- two",
			),
			wantHTML: `<ul><li>one</li><li>two</li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: blank line between siblings makes loose list",
			markdown: md(
				"- one",
				"",
				"- two",
			),
			wantHTML: `<ul><li><p>one</p></li><li><p>two</p></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: blank line within item followed by continuation makes loose list",
			markdown: md(
				"- one",
				"",
				"  two",
			),
			wantHTML: `<ul><li><p>one</p><p>two</p></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: trailing blank line after final item does not become loose by rollback",
			markdown: md(
				"- one",
				"",
			),
			wantHTML: `<ul><li>one</li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: blank line after item followed by dedented line rolls back blank",
			markdown: md(
				"- one",
				"",
				"two",
			),
			wantHTML: `<ul><li>one</li></ul><p>two</p>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: continuation line may contain emphasis",
			markdown: md(
				"- one",
				"  *two*",
			),
			wantHTML: `<ul><li>one <em>two</em></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: item may contain atx heading in body",
			markdown: md(
				"- one",
				"  # two",
			),
			wantHTML: `<ul><li>one<h1>two</h1></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: item body may contain single-line setext heading",
			markdown: md(
				"- one",
				"  ---",
			),
			wantHTML: `<ul><li><h2>one</h2></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: item body may contain multiline setext heading",
			markdown: md(
				"- one",
				"  two",
				"  ---",
			),
			wantHTML: `<ul><li><h2>one two</h2></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: item may contain thematic break after blank line",
			markdown: md(
				"- one",
				"",
				"  ---",
			),
			wantHTML: `<ul><li><p>one</p><hr></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: block quote nested inside item",
			markdown: md(
				"- outer",
				"  > quote",
			),
			wantHTML: `<ul><li>outer<blockquote><p>quote</p></blockquote></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: item may contain fenced code block in body",
			markdown: md(
				"- one",
				"  ```",
				"  code",
				"  ```",
			),
			wantHTML: `<ul><li>one<pre><code>code</code></pre></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: sibling item at different absolute indent does not join list",
			markdown: md(
				"- one",
				" - two",
			),
			wantHTML: `<ul><li>one</li></ul><ul><li>two</li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: mixed unordered marker families may still form sibling items",
			markdown: md(
				"- one",
				"* two",
				"+ three",
			),
			wantHTML: `<ul><li>one</li><li>two</li><li>three</li></ul>`,
			wantErr:  nil,
		},
		{
			name:     "unordered list: marker line with only spaces after marker creates empty item",
			markdown: "-    ",
			wantHTML: `<ul><li></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: continuation line trimmed to item baseline before recursive parsing",
			markdown: md(
				"- one",
				"    > two",
			),
			wantHTML: `<ul><li>one   &gt; two</li></ul>`,
			wantErr:  nil,
		},

		// ordered lists

		{
			name:     "ordered list: single item with period delimiter",
			markdown: "1. item",
			wantHTML: `<ol><li>item</li></ol>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: single item with right paren delimiter",
			markdown: "1) item",
			wantHTML: `<ol><li>item</li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: multiple sibling items with period delimiter",
			markdown: md(
				"1. one",
				"2. two",
				"3. three",
			),
			wantHTML: `<ol><li>one</li><li>two</li><li>three</li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: multiple sibling items with right paren delimiter",
			markdown: md(
				"1) one",
				"2) two",
				"3) three",
			),
			wantHTML: `<ol><li>one</li><li>two</li><li>three</li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: start number preserved from first marker",
			markdown: md(
				"3. one",
				"4. two",
			),
			wantHTML: `<ol start="3"><li>one</li><li>two</li></ol>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: zero start number allowed",
			markdown: "0. item",
			wantHTML: `<ol start="0"><li>item</li></ol>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: multi-digit marker allowed",
			markdown: "12. item",
			wantHTML: `<ol start="12"><li>item</li></ol>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: absurdly high marker rejected",
			markdown: "1000000001. item",
			wantHTML: `<p>1000000001. item</p>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: delimiter requires following space",
			markdown: "1.item",
			wantHTML: `<p>1.item</p>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: right paren delimiter requires following space",
			markdown: "1)item",
			wantHTML: `<p>1)item</p>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: tab after delimiter allowed",
			markdown: "1.\titem",
			wantHTML: `<ol><li>item</li></ol>`,
		},
		{
			name:     "ordered list: multiple spaces after delimiter allowed",
			markdown: "1.    item",
			wantHTML: `<ol><li>item</li></ol>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: multiple tabs and spaces after delimiter allowed",
			markdown: "1. \t  item",
			wantHTML: `<ol><li>item</li></ol>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: empty item content after required delimiter",
			markdown: "1. ",
			wantHTML: `<ol><li></li></ol>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: leading indentation of one space allowed",
			markdown: " 1. item",
			wantHTML: `<ol><li>item</li></ol>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: leading indentation of two spaces allowed",
			markdown: "  1. item",
			wantHTML: `<ol><li>item</li></ol>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: leading indentation of three spaces allowed",
			markdown: "   1. item",
			wantHTML: `<ol><li>item</li></ol>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: leading indentation of four spaces is not a list",
			markdown: "    1. item",
			wantHTML: `<pre><code>1. item</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: continuation line at content baseline stays in item",
			markdown: md(
				"1. one",
				"   two",
			),
			wantHTML: `<ol><li>one two</li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: continuation line beyond content baseline stays in item",
			markdown: md(
				"1. one",
				"     two",
			),
			wantHTML: `<ol><li>one   two</li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: dedented nonblank line ends list",
			markdown: md(
				"1. one",
				"two",
			),
			wantHTML: `<ol><li>one</li></ol><p>two</p>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: blank line between siblings makes loose list",
			markdown: md(
				"1. one",
				"",
				"2. two",
			),
			wantHTML: `<ol><li><p>one</p></li><li><p>two</p></li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: blank line within item followed by continuation makes loose list",
			markdown: md(
				"1. one",
				"",
				"   two",
			),
			wantHTML: `<ol><li><p>one</p><p>two</p></li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: trailing blank line after final item rolls back",
			markdown: md(
				"1. one",
				"",
			),
			wantHTML: `<ol><li>one</li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: sibling item must match period delimiter family",
			markdown: md(
				"1. one",
				"2) two",
			),
			wantHTML: `<ol><li>one</li></ol><ol start="2"><li>two</li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: sibling item must match right paren delimiter family",
			markdown: md(
				"1) one",
				"2. two",
			),
			wantHTML: `<ol><li>one</li></ol><ol start="2"><li>two</li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: item may contain atx heading in body",
			markdown: md(
				"1. one",
				"   # two",
			),
			wantHTML: `<ol><li>one<h1>two</h1></li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: item may contain setext heading in body",
			markdown: md(
				"1. one",
				"   two",
				"   ---",
			),
			wantHTML: `<ol><li><h2>one two</h2></li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: item may contain thematic break after blank line",
			markdown: md(
				"1. one",
				"",
				"   ---",
			),
			wantHTML: `<ol><li><p>one</p><hr></li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: block quote nested inside item",
			markdown: md(
				"1. outer",
				"   > quote",
			),
			wantHTML: `<ol><li>outer<blockquote><p>quote</p></blockquote></li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: item may contain fenced code block in body",
			markdown: md(
				"1. one",
				"   ```",
				"   code",
				"   ```",
			),
			wantHTML: `<ol><li>one<pre><code>code</code></pre></li></ol>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: marker line with only spaces after delimiter creates empty item",
			markdown: "1.      ",
			wantHTML: `<ol><li></li></ol>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: nonnumeric marker is not ordered list",
			markdown: "x. item",
			wantHTML: `<p>x. item</p>`,
			wantErr:  nil,
		},
		{
			name:     "ordered list: missing delimiter punctuation is not ordered list",
			markdown: "1 item",
			wantHTML: `<p>1 item</p>`,
			wantErr:  nil,
		},

		// nested lists and list interactions

		{
			name: "unordered list: nested unordered list in second line of item",
			markdown: md(
				"- outer",
				"  - inner",
			),
			wantHTML: `<ul><li>outer<ul><li>inner</li></ul></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: nested ordered list in second line of item",
			markdown: md(
				"- outer",
				"  1. inner",
			),
			wantHTML: `<ul><li>outer<ol><li>inner</li></ol></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: nested unordered list in second line of item",
			markdown: md(
				"1. outer",
				"   - inner",
			),
			wantHTML: `<ol><li>outer<ul><li>inner</li></ul></li></ol>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: nested ordered list in second line of item",
			markdown: md(
				"1. outer",
				"   1. inner",
			),
			wantHTML: `<ol><li>outer<ol><li>inner</li></ol></li></ol>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: nested sibling list items under one parent item",
			markdown: md(
				"- outer",
				"  - inner one",
				"  - inner two",
			),
			wantHTML: `<ul><li>outer<ul><li>inner one</li><li>inner two</li></ul></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: nested sibling list items under one parent item",
			markdown: md(
				"1. outer",
				"   1. inner one",
				"   2. inner two",
			),
			wantHTML: `<ol><li>outer<ol><li>inner one</li><li>inner two</li></ol></li></ol>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: nested list followed by parent continuation",
			markdown: md(
				"- outer",
				"  - inner",
				"  tail",
			),
			wantHTML: `<ul><li>outer<ul><li>inner</li></ul>tail</li></ul>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: nested list followed by parent continuation",
			markdown: md(
				"1. outer",
				"   1. inner",
				"   tail",
			),
			wantHTML: `<ol><li>outer<ol><li>inner</li></ol>tail</li></ol>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: two top-level items each with nested list",
			markdown: md(
				"- outer one",
				"  - inner one",
				"- outer two",
				"  - inner two",
			),
			wantHTML: `<ul><li>outer one<ul><li>inner one</li></ul></li><li>outer two<ul><li>inner two</li></ul></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: two top-level items each with nested list",
			markdown: md(
				"1. outer one",
				"   1. inner one",
				"2. outer two",
				"   1. inner two",
			),
			wantHTML: `<ol><li>outer one<ol><li>inner one</li></ol></li><li>outer two<ol><li>inner two</li></ol></li></ol>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: nested list separated by blank line makes parent loose",
			markdown: md(
				"- outer",
				"",
				"  - inner",
			),
			wantHTML: `<ul><li><p>outer</p><ul><li>inner</li></ul></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: nested list separated by blank line makes parent loose",
			markdown: md(
				"1. outer",
				"",
				"   1. inner",
			),
			wantHTML: `<ol><li><p>outer</p><ol><li>inner</li></ol></li></ol>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: nested list item may itself contain continuation paragraph",
			markdown: md(
				"- outer",
				"  - inner",
				"    tail",
			),
			wantHTML: `<ul><li>outer<ul><li>inner tail</li></ul></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: nested list item may itself contain continuation paragraph",
			markdown: md(
				"1. outer",
				"   1. inner",
				"      tail",
			),
			wantHTML: `<ol><li>outer<ol><li>inner tail</li></ol></li></ol>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: child item not meeting parent content baseline does not nest",
			markdown: md(
				"- outer",
				" - inner",
			),
			wantHTML: `<ul><li>outer</li></ul><ul><li>inner</li></ul>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: child item not meeting parent content baseline does not nest",
			markdown: md(
				"1. outer",
				"  1. inner",
			),
			wantHTML: `<ol><li>outer</li></ol><ol><li>inner</li></ol>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: top-level sibling resumes after nested list",
			markdown: md(
				"- outer",
				"  - inner",
				"- next outer",
			),
			wantHTML: `<ul><li>outer<ul><li>inner</li></ul></li><li>next outer</li></ul>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: top-level sibling resumes after nested list",
			markdown: md(
				"1. outer",
				"   1. inner",
				"2. next outer",
			),
			wantHTML: `<ol><li>outer<ol><li>inner</li></ol></li><li>next outer</li></ol>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: nested ordered list preserves start number",
			markdown: md(
				"- outer",
				"  3. inner",
			),
			wantHTML: `<ul><li>outer<ol start="3"><li>inner</li></ol></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: nested ordered list with right paren delimiter",
			markdown: md(
				"1. outer",
				"   1) inner",
			),
			wantHTML: `<ol><li>outer<ol><li>inner</li></ol></li></ol>`,
			wantErr:  nil,
		},
		{
			name: "unordered list: mixed nested unordered marker families are allowed",
			markdown: md(
				"- outer",
				"  * inner",
				"  + inner two",
			),
			wantHTML: `<ul><li>outer<ul><li>inner</li><li>inner two</li></ul></li></ul>`,
			wantErr:  nil,
		},
		{
			name: "ordered list: nested ordered sibling delimiter mismatch splits structure",
			markdown: md(
				"1. outer",
				"   1. inner one",
				"   2) inner two",
			),
			wantHTML: `<ol><li>outer<ol><li>inner one</li></ol><ol start="2"><li>inner two</li></ol></li></ol>`,
			wantErr:  nil,
		},

		// fenced code blocks

		{
			name: "fenced code: backtick fence minimum opener and closer",
			markdown: md(
				"```",
				"code",
				"```",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: tilde fence minimum opener and closer",
			markdown: md(
				"~~~",
				"code",
				"~~~",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: longer backtick opener and matching closer",
			markdown: md(
				"````",
				"code",
				"````",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: longer tilde opener and matching closer",
			markdown: md(
				"~~~~",
				"code",
				"~~~~",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: closer longer than opener",
			markdown: md(
				"```",
				"code",
				"````",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: closer shorter than opener does not close",
			markdown: md(
				"````",
				"code",
				"```",
			),
			wantHTML: "<pre><code>code\n```</code></pre>",
			wantErr:  nil,
		},
		{
			name: "fenced code: opener with one leading space",
			markdown: md(
				" ```",
				"code",
				" ```",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: opener with two leading spaces",
			markdown: md(
				"  ```",
				"code",
				"  ```",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: opener with three leading spaces",
			markdown: md(
				"   ```",
				"code",
				"   ```",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: opener with four leading spaces is not fenced code",
			markdown: md(
				"    ```",
				"    code",
				"    ```",
			),
			wantHTML: "<pre><code>```\ncode\n```</code></pre>",
			wantErr:  nil,
		},
		{
			name: "fenced code: closer with one leading space",
			markdown: md(
				"```",
				"code",
				" ```",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: closer with two leading spaces",
			markdown: md(
				"```",
				"code",
				"  ```",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: closer with three leading spaces",
			markdown: md(
				"```",
				"code",
				"   ```",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: closer with four leading spaces is not closing fence",
			markdown: md(
				"```",
				"code",
				"    ```",
				"```",
			),
			wantHTML: "<pre><code>code\n    ```</code></pre>",
			wantErr:  nil,
		},
		{
			name: "fenced code: blank line inside block",
			markdown: md(
				"```",
				"one",
				"",
				"two",
				"```",
			),
			wantHTML: "<pre><code>one\n\ntwo</code></pre>",
			wantErr:  nil,
		},
		{
			name: "fenced code: empty fenced block",
			markdown: md(
				"```",
				"```",
			),
			wantHTML: `<pre><code></code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: unclosed backtick fence runs to eof",
			markdown: md(
				"```",
				"code",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: unclosed tilde fence runs to eof",
			markdown: md(
				"~~~",
				"code",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: payload line equal to shorter fence is literal content",
			markdown: md(
				"````",
				"```",
				"````",
			),
			wantHTML: "<pre><code>```</code></pre>",
			wantErr:  nil,
		},
		{
			name: "fenced code: closing fence may have trailing spaces",
			markdown: md(
				"```",
				"code",
				"```   ",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: closing fence may have trailing tabs",
			markdown: md(
				"```",
				"code",
				"```\t\t",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: closing fence with trailing nonwhitespace is not valid closer",
			markdown: md(
				"```",
				"code",
				"```x",
				"```",
			),
			wantHTML: "<pre><code>code\n```x</code></pre>",
			wantErr:  nil,
		},
		{
			name: "fenced code: backtick opener with info string",
			markdown: md(
				"```go",
				"code",
				"```",
			),
			wantHTML: `<pre><code class="language-go">code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: backtick opener with delimiter whitespace before info string",
			markdown: md(
				"```   go",
				"code",
				"```",
			),
			wantHTML: `<pre><code class="language-go">code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: tilde opener may contain backticks in info string",
			markdown: md(
				"~~~ ```",
				"code",
				"~~~",
			),
			wantHTML: "<pre><code class=\"language-```\">code</code></pre>",
			wantErr:  nil,
		},
		{
			name: "fenced code: backtick opener rejects info string containing backtick",
			markdown: md(
				"``` `",
				"code",
				"```",
			),
			wantHTML: "<p>``` ` code</p><pre><code></code></pre>",
			wantErr:  nil,
		},
		{
			name: "fenced code: opener with fewer than three markers is not fenced code",
			markdown: md(
				"``",
				"code",
				"``",
			),
			wantHTML: "<p>`` code ``</p>",
			wantErr:  nil,
		},
		{
			name: "fenced code: mixed marker family does not close block",
			markdown: md(
				"```",
				"code",
				"~~~",
				"```",
			),
			wantHTML: "<pre><code>code\n~~~</code></pre>",
			wantErr:  nil,
		},
		{
			name: "fenced code: marker-looking content line is literal until valid closer",
			markdown: md(
				"```",
				"~~~",
				"```",
			),
			wantHTML: `<pre><code>~~~</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: html-looking content inside block",
			markdown: md(
				"```",
				"<div>",
				"```",
			),
			wantHTML: `<pre><code>&lt;div&gt;</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: block-quote-looking content inside block",
			markdown: md(
				"```",
				"> hello",
				"```",
			),
			wantHTML: `<pre><code>&gt; hello</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: list-looking content inside block",
			markdown: md(
				"```",
				"- hello",
				"```",
			),
			wantHTML: `<pre><code>- hello</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: fenced opener interrupts paragraph without blank line",
			markdown: md(
				"one",
				"```",
				"two",
				"```",
			),
			wantHTML: `<p>one</p><pre><code>two</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: fenced block followed by paragraph",
			markdown: md(
				"```",
				"code",
				"```",
				"tail",
			),
			wantHTML: `<pre><code>code</code></pre><p>tail</p>`,
			wantErr:  nil,
		},
		{
			name: "fenced code: opener with only delimiter whitespace and no info string",
			markdown: md(
				"```   ",
				"code",
				"```",
			),
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},

		// indented code blocks

		{
			name:     "indented code: single line with four spaces",
			markdown: `    code`,
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name:     "indented code: single line with more than four spaces preserves remainder",
			markdown: `      code`,
			wantHTML: `<pre><code>  code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "indented code: multiple indented lines",
			markdown: md(
				"    one",
				"    two",
			),
			wantHTML: "<pre><code>one\ntwo</code></pre>",
			wantErr:  nil,
		},
		{
			name: "indented code: blank line inside block",
			markdown: md(
				"    one",
				"",
				"    two",
			),
			wantHTML: "<pre><code>one\n\ntwo</code></pre>",
			wantErr:  nil,
		},
		{
			name: "indented code: trailing blank lines rolled back before dedented line",
			markdown: md(
				"    one",
				"",
				"two",
			),
			wantHTML: `<pre><code>one</code></pre><p>two</p>`,
			wantErr:  nil,
		},
		{
			name: "indented code: trailing blank lines at eof",
			markdown: md(
				"    one",
				"",
			),
			wantHTML: "<pre><code>one\n</code></pre>",
			wantErr:  nil,
		},
		{
			name:     "indented code: line with three leading spaces is not code block",
			markdown: `   code`,
			wantHTML: `<p>   code</p>`,
			wantErr:  nil,
		},
		{
			name:     "indented code: tab reaching four columns",
			markdown: "\tcode",
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name:     "indented code: mixed indentation reaching four columns",
			markdown: "  \tcode",
			wantHTML: `<pre><code>code</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "indented code: paragraph transparency with continuation line",
			markdown: md(
				"one",
				"    two",
			),
			wantHTML: `<p>one     two</p>`,
			wantErr:  nil,
		},
		{
			name: "indented code: begins after blank line following paragraph",
			markdown: md(
				"one",
				"",
				"    two",
			),
			wantHTML: `<p>one</p><pre><code>two</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "indented code: dedented nonblank line ends block",
			markdown: md(
				"    one",
				"    two",
				"three",
			),
			wantHTML: "<pre><code>one\ntwo</code></pre><p>three</p>",
			wantErr:  nil,
		},
		{
			name:     "indented code: thematic-break-looking content is literal",
			markdown: `    ---`,
			wantHTML: `<pre><code>---</code></pre>`,
			wantErr:  nil,
		},
		{
			name:     "indented code: block-quote-looking content is literal",
			markdown: `    > hello`,
			wantHTML: `<pre><code>&gt; hello</code></pre>`,
			wantErr:  nil,
		},
		{
			name:     "indented code: list-looking content is literal",
			markdown: `    - hello`,
			wantHTML: `<pre><code>- hello</code></pre>`,
			wantErr:  nil,
		},
		{
			name:     "indented code: atx-heading-looking content is literal",
			markdown: `    # hello`,
			wantHTML: `<pre><code># hello</code></pre>`,
			wantErr:  nil,
		},
		{
			name:     "indented code: fenced-opener-looking content is literal",
			markdown: `    ~~~`,
			wantHTML: `<pre><code>~~~</code></pre>`,
			wantErr:  nil,
		},
		{
			name: "indented code: multiple blank lines inside block",
			markdown: md(
				"    one",
				"",
				"",
				"    two",
			),
			wantHTML: "<pre><code>one\n\n\ntwo</code></pre>",
			wantErr:  nil,
		},
		{
			name:     "indented code: trailing spaces in content line preserved",
			markdown: "    code  ",
			wantHTML: `<pre><code>code  </code></pre>`,
			wantErr:  nil,
		},
		{
			name:     "indented code: html-looking content escaped",
			markdown: `    <div>`,
			wantHTML: `<pre><code>&lt;div&gt;</code></pre>`,
			wantErr:  nil,
		},

		// strong & emphasis

		{
			name:     "emphasis: single delimiter produces emphasis",
			markdown: "*a*",
			wantHTML: `<p><em>a</em></p>`,
			wantErr:  nil,
		},
		{
			name:     "strong: double delimiter produces strong emphasis",
			markdown: "**a**",
			wantHTML: `<p><strong>a</strong></p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: underscore produces emphasis",
			markdown: "_a_",
			wantHTML: `<p><em>a</em></p>`,
			wantErr:  nil,
		},
		{
			name:     "strong: underscore produces strong emphasis",
			markdown: "__a__",
			wantHTML: `<p><strong>a</strong></p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: delimiter must be left and right flanking",
			markdown: "a *b* c",
			wantHTML: `<p>a <em>b</em> c</p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: opening delimiter cannot be followed by whitespace",
			markdown: "* b*",
			wantHTML: `<ul><li>b*</li></ul>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: closing delimiter cannot be preceded by whitespace",
			markdown: "*b *",
			wantHTML: `<p>*b *</p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: delimiter adjacent to punctuation can open",
			markdown: "(*a*)",
			wantHTML: `<p>(<em>a</em>)</p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: delimiter adjacent to punctuation can close",
			markdown: "(*a*)",
			wantHTML: `<p>(<em>a</em>)</p>`,
			wantErr:  nil,
		},
		{
			name:     "underscore: intraword emphasis is disallowed",
			markdown: "foo_bar_baz",
			wantHTML: `<p>foo_bar_baz</p>`,
			wantErr:  nil,
		},
		{
			name:     "underscore: emphasis allowed when separated by punctuation",
			markdown: "foo _bar_ baz",
			wantHTML: `<p>foo <em>bar</em> baz</p>`,
			wantErr:  nil,
		},
		{
			name:     "underscore: can open when preceded by punctuation",
			markdown: "(_a_)",
			wantHTML: `<p>(<em>a</em>)</p>`,
			wantErr:  nil,
		},
		{
			name:     "underscore: can close when followed by punctuation",
			markdown: "(_a_)",
			wantHTML: `<p>(<em>a</em>)</p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: triple delimiter can resolve to emphasis and strong nesting",
			markdown: "***a***",
			wantHTML: `<p><em><strong>a</strong></em></p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: unmatched extra delimiter remains as literal text",
			markdown: "**a*",
			wantHTML: `<p>*<em>a</em></p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: unmatched closing delimiter remains as literal text",
			markdown: "*a**",
			wantHTML: `<p><em>a</em>*</p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: delimiter pairing blocked by multiple of three rule",
			markdown: "***a**",
			wantHTML: `<p>*<strong>a</strong></p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: mod-3 rule prevents pairing across runs",
			markdown: "**a***",
			wantHTML: `<p><strong>a</strong>*</p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: nested emphasis resolves inside outer emphasis",
			markdown: "*a *b* c*",
			wantHTML: `<p><em>a <em>b</em> c</em></p>`,
			wantErr:  nil,
		},
		{
			name:     "strong: nested strong inside emphasis",
			markdown: "*a **b** c*",
			wantHTML: `<p><em>a <strong>b</strong> c</em></p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: nested emphasis inside strong",
			markdown: "**a *b* c**",
			wantHTML: `<p><strong>a <em>b</em> c</strong></p>`,
			wantErr:  nil,
		},
		{
			name:     "strong: nested strong resolves correctly",
			markdown: "**a **b** c**",
			wantHTML: `<p><strong>a <strong>b</strong> c</strong></p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: crossing delimiters do not form valid nesting",
			markdown: "*a **b* c**",
			wantHTML: `<p><em>a <em><em>b</em> c</em></em></p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: multiple independent emphasis runs",
			markdown: "*a* *b*",
			wantHTML: `<p><em>a</em> <em>b</em></p>`,
			wantErr:  nil,
		},
		{
			name:     "strong: multiple independent strong runs",
			markdown: "**a** **b**",
			wantHTML: `<p><strong>a</strong> <strong>b</strong></p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: mixed delimiter kinds do not match",
			markdown: "*a_",
			wantHTML: `<p>*a_</p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: mixed delimiter kinds do not match",
			markdown: "_a*",
			wantHTML: `<p>_a*</p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: delimiter at start of input",
			markdown: "*a*",
			wantHTML: `<p><em>a</em></p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: delimiter at end of input",
			markdown: "a *b*",
			wantHTML: `<p>a <em>b</em></p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: isolated delimiter produces literal text",
			markdown: "*",
			wantHTML: `<p>*</p>`,
			wantErr:  nil,
		},
		{
			name:     "strong: isolated double delimiter produces literal text",
			markdown: "**",
			wantHTML: `<p>**</p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: empty content does not form emphasis",
			markdown: "**",
			wantHTML: `<p>**</p>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: empty content between delimiters is ignored",
			markdown: "* *",
			wantHTML: `<ul><li>*</li></ul>`,
			wantErr:  nil,
		},
		{
			name:     "emphasis: emphasis spans do not consume surrounding text",
			markdown: "a *b* c",
			wantHTML: `<p>a <em>b</em> c</p>`,
			wantErr:  nil,
		},
		{
			name:     "strong: strong spans do not consume surrounding text",
			markdown: "a **b** c",
			wantHTML: `<p>a <strong>b</strong> c</p>`,
			wantErr:  nil,
		},

		// code spans

		{
			name:     "code span: matching single delimiters produce code span",
			markdown: "`code`",
			wantHTML: `<p><code>code</code></p>`,
			wantErr:  nil,
		},
		{
			name:     "code span: matching multi-delimiters produce code span",
			markdown: "``code``",
			wantHTML: `<p><code>code</code></p>`,
			wantErr:  nil,
		},
		{
			name:     "code span: content may contain single backticks when wrapped in wider delimiters",
			markdown: "``a ` b``",
			wantHTML: "<p><code>a ` b</code></p>",
			wantErr:  nil,
		},
		{
			name:     "code span: content may contain double backticks when wrapped in wider delimiters",
			markdown: "```a `` b```",
			wantHTML: "<p><code>a `` b</code></p>",
			wantErr:  nil,
		},
		{
			name:     "code span: opening delimiter without matching closer remains literal text",
			markdown: "`code",
			wantHTML: "<p>`code</p>",
			wantErr:  nil,
		},
		{
			name:     "code span: closing delimiter with different width does not close span",
			markdown: "`code``",
			wantHTML: "<p>`code``</p>",
			wantErr:  nil,
		},
		{
			name:     "code span: opener skips non-matching backtick runs and closes on matching width",
			markdown: "``code`more``",
			wantHTML: "<p><code>code`more</code></p>",
			wantErr:  nil,
		},
		{
			name:     "code span: empty content is allowed",
			markdown: "````",
			wantHTML: "<pre><code></code></pre>",
			wantErr:  nil,
		},
		{
			name:     "code span: single leading and trailing spaces are trimmed when content is not all spaces",
			markdown: "` code `",
			wantHTML: "<p><code>code</code></p>",
			wantErr:  nil,
		},
		{
			name:     "code span: only one leading and trailing space are trimmed",
			markdown: "`  code  `",
			wantHTML: "<p><code> code </code></p>",
			wantErr:  nil,
		},
		{
			name:     "code span: all-space content is not trimmed",
			markdown: "`   `",
			wantHTML: "<p><code>   </code></p>",
			wantErr:  nil,
		},
		{
			name:     "code span: interior spaces are preserved",
			markdown: "`a  b`",
			wantHTML: "<p><code>a  b</code></p>",
			wantErr:  nil,
		},
		{
			name:     "code span: punctuation is treated as literal content",
			markdown: "`<a>*&_[]()`",
			wantHTML: "<p><code>&lt;a&gt;*&amp;_[]()</code></p>",
			wantErr:  nil,
		},
		{
			name:     "code span: emphasis delimiters inside span are literal content",
			markdown: "`*a*`",
			wantHTML: "<p><code>*a*</code></p>",
			wantErr:  nil,
		},
		{
			name:     "code span: brackets and parentheses inside span are literal content",
			markdown: "`[a](b)`",
			wantHTML: "<p><code>[a](b)</code></p>",
			wantErr:  nil,
		},
		{
			name:     "code span: span does not consume surrounding text",
			markdown: "a `code` b",
			wantHTML: "<p>a <code>code</code> b</p>",
			wantErr:  nil,
		},
		{
			name:     "code span: multiple code spans may appear in one line",
			markdown: "`a` and `b`",
			wantHTML: "<p><code>a</code> and <code>b</code></p>",
			wantErr:  nil,
		},
		{
			name:     "code span: unmatched wider opener remains literal text",
			markdown: "``code`",
			wantHTML: "<p>``code`</p>",
			wantErr:  nil,
		},
		{
			name:     "code span: later matching closer forms span after earlier mismatched run",
			markdown: "```code``more```",
			wantHTML: "<p><code>code``more</code></p>",
			wantErr:  nil,
		},
		{
			name:     "code span: lone backtick remains literal text",
			markdown: "`",
			wantHTML: "<p>`</p>",
			wantErr:  nil,
		},

		// inline links

		{
			name:     "link: inline destination resolves to link",
			markdown: "[label](/url)",
			wantHTML: `<p><a href="/url">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: empty destination is allowed",
			markdown: "[label]()",
			wantHTML: `<p><a href="">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: empty label is allowed",
			markdown: "[](/url)",
			wantHTML: `<p><a href="/url"></a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: angle-delimited destination resolves to link",
			markdown: "[label](</url>)",
			wantHTML: `<p><a href="/url">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: bare destination resolves to link",
			markdown: "[label](/a/b)",
			wantHTML: `<p><a href="/a/b">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: bare destination may contain balanced parentheses",
			markdown: "[label](a(b)c)",
			wantHTML: `<p><a href="a(b)c">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: bare destination may contain escaped parentheses",
			markdown: "[label](a\\(b\\)c)",
			wantHTML: `<p><a href="a(b)c">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: escaped punctuation is unescaped in destination and title",
			markdown: `[label](a\(b\)c "ti\"tle")`,
			wantHTML: `<p><a href="a(b)c" title="ti&#34;tle">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: destination may include a double-quoted title",
			markdown: "[label](/url \"title\")",
			wantHTML: `<p><a href="/url" title="title">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: destination may include a single-quoted title",
			markdown: "[label](/url 'title')",
			wantHTML: `<p><a href="/url" title="title">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: destination may include a parenthesized title",
			markdown: "[label](/url (title))",
			wantHTML: `<p><a href="/url" title="title">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: title may contain balanced parentheses",
			markdown: "[label](/url (a (b) c))",
			wantHTML: `<p><a href="/url" title="a (b) c">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: whitespace between destination and title is required",
			markdown: "[label](/url\"title\")",
			wantHTML: `<p><a href="/url&#34;title&#34;">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: spaces and tabs around destination are allowed",
			markdown: "[label]( \t/url\t )",
			wantHTML: `<p><a href="/url">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: spaces and tabs around destination and title are allowed",
			markdown: "[label]( \t/url \t \"title\"\t )",
			wantHTML: `<p><a href="/url" title="title">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: nested inline content is allowed in label",
			markdown: "[a *b* c](/url)",
			wantHTML: `<p><a href="/url">a <em>b</em> c</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: code span is allowed inside label",
			markdown: "[a `b` c](/url)",
			wantHTML: `<p><a href="/url">a <code>b</code> c</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: image is allowed inside label",
			markdown: "[![alt](/img.png)](/url)",
			wantHTML: `<p><a href="/url"><img alt="alt" src="/img.png"></a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: surrounding text is preserved",
			markdown: "a [label](/url) b",
			wantHTML: `<p>a <a href="/url">label</a> b</p>`,
			wantErr:  nil,
		},
		{
			name:     "link: missing tail leaves brackets as literal text",
			markdown: "[label]",
			wantHTML: `<p>[label]</p>`,
			wantErr:  nil,
		},
		{
			name:     "link: missing closing parenthesis leaves construct as literal text",
			markdown: "[label](/url",
			wantHTML: `<p>[label](/url</p>`,
			wantErr:  nil,
		},
		{
			name:     "link: invalid angle destination leaves construct as literal text",
			markdown: "[label](<a<>)",
			wantHTML: `<p>[label](&lt;a&lt;&gt;)</p>`,
			wantErr:  nil,
		},
		{
			name:     "link: unbalanced bare destination leaves construct as literal text",
			markdown: "[label](a(b)",
			wantHTML: `<p>[label](a(b)</p>`,
			wantErr:  nil,
		},
		{
			name:     "link: quoted text without separator is parsed as bare destination",
			markdown: "[label](\"title\")",
			wantHTML: `<p><a href="&#34;title&#34;">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "link: newline in angle destination leaves construct as literal text",
			markdown: "[label](<a\nb>)",
			wantHTML: `<p>[label](&lt;a b&gt;)</p>`,
			wantErr:  nil,
		},
		{
			name:     "link: newline in quoted title leaves construct as literal text",
			markdown: "[label](/url \"a\nb\")",
			wantHTML: `<p>[label](/url &#34;a b&#34;)</p>`,
			wantErr:  nil,
		},
		{
			name:     "link: newline in parenthesized title leaves construct as literal text",
			markdown: "[label](/url (a\nb))",
			wantHTML: `<p>[label](/url (a b))</p>`,
			wantErr:  nil,
		},
		{
			name:     "link: literal closing bracket without opener remains text",
			markdown: "label](/url)",
			wantHTML: `<p>label](/url)</p>`,
			wantErr:  nil,
		},
		{
			name:     "link: nested links are rejected",
			markdown: "[outer [inner](/in)](/out)",
			wantHTML: `<p>[outer <a href="/in">inner</a>](/out)</p>`,
			wantErr:  nil,
		},
		{
			name:     "link: inner links may still resolve when outer link is rejected",
			markdown: "[outer [inner](/in)]",
			wantHTML: `<p>[outer <a href="/in">inner</a>]</p>`,
			wantErr:  nil,
		},

		// inline images

		{
			name:     "image: inline destination resolves to image",
			markdown: "![alt](/img.png)",
			wantHTML: `<p><img alt="alt" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: empty destination is allowed",
			markdown: "![alt]()",
			wantHTML: `<p><img alt="alt" src=""></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: empty label is allowed",
			markdown: "![](/img.png)",
			wantHTML: `<p><img alt="" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: angle-delimited destination resolves to image",
			markdown: "![alt](</img.png>)",
			wantHTML: `<p><img alt="alt" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: bare destination resolves to image",
			markdown: "![alt](/a/b.png)",
			wantHTML: `<p><img alt="alt" src="/a/b.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: bare destination may contain balanced parentheses",
			markdown: "![alt](a(b)c.png)",
			wantHTML: `<p><img alt="alt" src="a(b)c.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: bare destination may contain escaped parentheses",
			markdown: "![alt](a\\(b\\)c.png)",
			wantHTML: `<p><img alt="alt" src="a(b)c.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: escaped punctuation is unescaped in destination and title",
			markdown: `![alt](a\(b\)c "ti\"tle")`,
			wantHTML: `<p><img alt="alt" src="a(b)c" title="ti&#34;tle"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: destination may include a double-quoted title",
			markdown: "![alt](/img.png \"title\")",
			wantHTML: `<p><img alt="alt" src="/img.png" title="title"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: destination may include a single-quoted title",
			markdown: "![alt](/img.png 'title')",
			wantHTML: `<p><img alt="alt" src="/img.png" title="title"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: destination may include a parenthesized title",
			markdown: "![alt](/img.png (title))",
			wantHTML: `<p><img alt="alt" src="/img.png" title="title"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: title may contain balanced parentheses",
			markdown: "![alt](/img.png (a (b) c))",
			wantHTML: `<p><img alt="alt" src="/img.png" title="a (b) c"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: whitespace between destination and title is required",
			markdown: "![alt](/img.png\"title\")",
			wantHTML: `<p><img alt="alt" src="/img.png&#34;title&#34;"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: spaces and tabs around destination are allowed",
			markdown: "![alt]( \t/img.png\t )",
			wantHTML: `<p><img alt="alt" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: spaces and tabs around destination and title are allowed",
			markdown: "![alt]( \t/img.png \t \"title\"\t )",
			wantHTML: `<p><img alt="alt" src="/img.png" title="title"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: nested inline content is allowed in label",
			markdown: "![a *b* c](/img.png)",
			wantHTML: `<p><img alt="a b c" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: code span is allowed inside label",
			markdown: "![a `b` c](/img.png)",
			wantHTML: `<p><img alt="a b c" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: link is allowed inside image label",
			markdown: "![see [this](/url)](/img.png)",
			wantHTML: `<p><img alt="see this" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: surrounding text is preserved",
			markdown: "a ![alt](/img.png) b",
			wantHTML: `<p>a <img alt="alt" src="/img.png"> b</p>`,
			wantErr:  nil,
		},
		{
			name:     "image: missing tail leaves brackets as literal text",
			markdown: "![alt]",
			wantHTML: `<p>![alt]</p>`,
			wantErr:  nil,
		},
		{
			name:     "image: missing closing parenthesis leaves construct as literal text",
			markdown: "![alt](/img.png",
			wantHTML: `<p>![alt](/img.png</p>`,
			wantErr:  nil,
		},
		{
			name:     "image: invalid angle destination leaves construct as literal text",
			markdown: "![alt](<a<>)",
			wantHTML: `<p>![alt](&lt;a&lt;&gt;)</p>`,
			wantErr:  nil,
		},
		{
			name:     "image: unbalanced bare destination leaves construct as literal text",
			markdown: "![alt](a(b)",
			wantHTML: `<p>![alt](a(b)</p>`,
			wantErr:  nil,
		},
		{
			name:     "image: quoted text without separator is parsed as bare destination",
			markdown: "![alt](\"title\")",
			wantHTML: `<p><img alt="alt" src="&#34;title&#34;"></p>`,
			wantErr:  nil,
		},
		{
			name:     "image: newline in angle destination leaves construct as literal text",
			markdown: "![alt](<a\nb>)",
			wantHTML: `<p>![alt](&lt;a b&gt;)</p>`,
			wantErr:  nil,
		},
		{
			name:     "image: newline in quoted title leaves construct as literal text",
			markdown: "![alt](/img.png \"a\nb\")",
			wantHTML: `<p>![alt](/img.png &#34;a b&#34;)</p>`,
			wantErr:  nil,
		},
		{
			name:     "image: newline in parenthesized title leaves construct as literal text",
			markdown: "![alt](/img.png (a\nb))",
			wantHTML: `<p>![alt](/img.png (a b))</p>`,
			wantErr:  nil,
		},
		{
			name:     "image: literal closing bracket without opener remains text",
			markdown: "alt](/img.png)",
			wantHTML: `<p>alt](/img.png)</p>`,
			wantErr:  nil,
		},
		{
			name:     "image: nested images are allowed",
			markdown: "![outer ![inner](/in.png)](/out.png)",
			wantHTML: `<p><img alt="outer inner" src="/out.png"></p>`,
			wantErr:  nil,
		},

		// inline autolinks

		{
			name:     "autolink: uri with alphabetic scheme resolves to link",
			markdown: "<https://example.com>",
			wantHTML: `<p><a href="https://example.com">https://example.com</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: uri scheme may include plus period and hyphen",
			markdown: "<a+b.c-d:xyz>",
			wantHTML: `<p><a href="a+b.c-d:xyz">a+b.c-d:xyz</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: uri scheme may include digits after the first character",
			markdown: "<x1:abc>",
			wantHTML: `<p><a href="x1:abc">x1:abc</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: uri scheme must begin with a letter",
			markdown: "<1x:abc>",
			wantHTML: `<p>&lt;1x:abc&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: uri scheme must contain a colon",
			markdown: "<https//example.com>",
			wantHTML: `<p>&lt;https//example.com&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: uri scheme must be at least two characters",
			markdown: "<h:abc>",
			wantHTML: `<p>&lt;h:abc&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: uri scheme of thirty two characters is allowed",
			markdown: "<abcdefghijklmnopqrstuvwxyzabcdef:abc>",
			wantHTML: `<p><a href="abcdefghijklmnopqrstuvwxyzabcdef:abc">abcdefghijklmnopqrstuvwxyzabcdef:abc</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: uri scheme may not exceed thirty two characters",
			markdown: "<abcdefghijklmnopqrstuvwxyzabcdefg:abc>",
			wantHTML: `<p>&lt;abcdefghijklmnopqrstuvwxyzabcdefg:abc&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: uri may contain query punctuation",
			markdown: "<https://example.com?a=1&b=2>",
			wantHTML: `<p><a href="https://example.com?a=1&amp;b=2">https://example.com?a=1&amp;b=2</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: uri may not contain spaces",
			markdown: "<https://example .com>",
			wantHTML: `<p>&lt;https://example .com&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: uri may not contain angle brackets in content",
			markdown: "<https://exa<mple.com>",
			wantHTML: `<p>&lt;https://exa&lt;mple.com&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: uri without closing angle bracket remains literal text",
			markdown: "<https://example.com",
			wantHTML: `<p>&lt;https://example.com</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: surrounding text is preserved for uri autolink",
			markdown: "a <https://example.com> b",
			wantHTML: `<p>a <a href="https://example.com">https://example.com</a> b</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: email address resolves to mail link",
			markdown: "<user@example.com>",
			wantHTML: `<p><a href="mailto:user@example.com">user@example.com</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: email local part may contain permitted punctuation",
			markdown: "<a.b+c_d-test@example.com>",
			wantHTML: `<p><a href="mailto:a.b+c_d-test@example.com">a.b+c_d-test@example.com</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: email domain may contain hyphen within label",
			markdown: "<user@exa-mple.com>",
			wantHTML: `<p><a href="mailto:user@exa-mple.com">user@exa-mple.com</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: email requires exactly one at sign",
			markdown: "<a@b@c.com>",
			wantHTML: `<p>&lt;a@b@c.com&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: email requires nonempty local part",
			markdown: "<@example.com>",
			wantHTML: `<p>&lt;@example.com&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: email requires nonempty domain",
			markdown: "<user@>",
			wantHTML: `<p>&lt;user@&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: email domain labels may not begin with hyphen",
			markdown: "<user@-example.com>",
			wantHTML: `<p>&lt;user@-example.com&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: email domain labels may not end with hyphen",
			markdown: "<user@example-.com>",
			wantHTML: `<p>&lt;user@example-.com&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: email domain labels may not contain underscore",
			markdown: "<user@exa_mple.com>",
			wantHTML: `<p>&lt;user@exa_mple.com&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: email domain labels may not be empty",
			markdown: "<user@example..com>",
			wantHTML: `<p>&lt;user@example..com&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: email without closing angle bracket remains literal text",
			markdown: "<user@example.com",
			wantHTML: `<p>&lt;user@example.com</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: surrounding text is preserved for email autolink",
			markdown: "a <user@example.com> b",
			wantHTML: `<p>a <a href="mailto:user@example.com">user@example.com</a> b</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: email autolink does not require dotted domain",
			markdown: "<local@domain>",
			wantHTML: `<p><a href="mailto:local@domain">local@domain</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: invalid email content falls back to literal text",
			markdown: "<local@do_main>",
			wantHTML: `<p>&lt;local@do_main&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: invalid uri content falls back to literal text",
			markdown: "<http:exa mple>",
			wantHTML: `<p>&lt;http:exa mple&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: invalid email content falls back to literal text",
			markdown: "<user@exa_mple.com>",
			wantHTML: `<p>&lt;user@exa_mple.com&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: lone opening angle bracket remains literal text",
			markdown: "<",
			wantHTML: `<p>&lt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "autolink: empty angle pair remains literal text",
			markdown: "<>",
			wantHTML: `<p>&lt;&gt;</p>`,
			wantErr:  nil,
		},

		// raw inline HTML

		{
			name:     "inline html: comment resolves as raw html",
			markdown: "<!-- comment -->",
			wantHTML: `<!-- comment -->`,
			wantErr:  nil,
		},
		{
			name:     "inline html: empty comment resolves as raw html",
			markdown: "<!---->",
			wantHTML: `<!---->`,
			wantErr:  nil,
		},
		{
			name:     "inline html: processing instruction resolves as raw html",
			markdown: "<?php?>",
			wantHTML: `<?php?>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: declaration resolves as raw html",
			markdown: "<!DOCTYPE html>",
			wantHTML: `<!DOCTYPE html>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: cdata section resolves as raw html",
			markdown: "<![CDATA[hello]]>",
			wantHTML: `<![CDATA[hello]]>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: opening tag resolves as raw html",
			markdown: "<span>",
			wantHTML: `<p><span></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: closing tag resolves as raw html",
			markdown: "</span>",
			wantHTML: `<p></span></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: opening tag may contain attributes",
			markdown: "<a href=\"/url\" title=\"x\">",
			wantHTML: `<p><a href="/url" title="x"></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: opening tag may contain single quoted attributes",
			markdown: "<a href='/url' title='x'>",
			wantHTML: `<p><a href='/url' title='x'></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: opening tag may contain unquoted attributes",
			markdown: "<a href=/url title=x>",
			wantHTML: `<p><a href=/url title=x></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: opening tag may contain bare attributes",
			markdown: "<input disabled>",
			wantHTML: `<p><input disabled></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: opening tag may contain spaces before attributes",
			markdown: "<a  href=\"/url\">",
			wantHTML: `<p><a  href="/url"></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: self closing tag resolves as raw html",
			markdown: "<br/>",
			wantHTML: `<p><br/></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: self closing tag may contain spaces before slash",
			markdown: "<br />",
			wantHTML: `<p><br /></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: self closing tag may contain attributes",
			markdown: "<img src=\"/img.png\" alt=\"x\" />",
			wantHTML: `<p><img src="/img.png" alt="x" /></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: tag attributes may mix quoting styles",
			markdown: "<a href=\"/url\" data-x='y' rel=noopener>",
			wantHTML: `<p><a href="/url" data-x='y' rel=noopener></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: quoted attribute values may contain closing angle brackets",
			markdown: "<a title=\"1 > 0\">",
			wantHTML: `<p><a title="1 > 0"></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: quoted attribute values may contain opposite quote kind",
			markdown: "<a title='\"x\"'>",
			wantHTML: `<p><a title='"x"'></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: surrounding text is preserved",
			markdown: "a <span> b",
			wantHTML: `<p>a <span> b</p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: opening tag requires alphabetic tag name start",
			markdown: "<1span>",
			wantHTML: `<p>&lt;1span&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: closing tag requires alphabetic tag name start",
			markdown: "</1span>",
			wantHTML: `<p>&lt;/1span&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: tag name may contain digits after the first character",
			markdown: "<h1>",
			wantHTML: `<h1>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: tag name may contain hyphen",
			markdown: "<custom-tag>",
			wantHTML: `<p><custom-tag></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: attribute name may begin with underscore",
			markdown: "<a _x=\"y\">",
			wantHTML: `<p><a _x="y"></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: attribute name may begin with colon",
			markdown: "<a :x=\"y\">",
			wantHTML: `<p><a :x="y"></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: attribute name may contain punctuation",
			markdown: "<a data.x:y-z=\"v\">",
			wantHTML: `<p><a data.x:y-z="v"></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: unterminated comment remains literal text",
			markdown: "<!-- comment",
			wantHTML: `<!-- comment`,
			wantErr:  nil,
		},
		{
			name:     "inline html: unterminated processing instruction remains literal text",
			markdown: "<?php",
			wantHTML: `<?php`,
			wantErr:  nil,
		},
		{
			name:     "inline html: unterminated declaration remains literal text",
			markdown: "<!DOCTYPE html",
			wantHTML: `<!DOCTYPE html`,
			wantErr:  nil,
		},
		{
			name:     "inline html: unterminated cdata remains literal text",
			markdown: "<![CDATA[hello",
			wantHTML: `<![CDATA[hello`,
			wantErr:  nil,
		},
		{
			name:     "inline html: unterminated opening tag remains literal text",
			markdown: "<span",
			wantHTML: `<p>&lt;span</p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: unterminated closing tag remains literal text",
			markdown: "</span",
			wantHTML: `<p>&lt;/span</p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: attribute requires separator whitespace",
			markdown: "<ahref=\"/url\">",
			wantHTML: `<p>&lt;ahref=&#34;/url&#34;&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: attribute value requires nonempty content when unquoted",
			markdown: "<a href=>",
			wantHTML: `<p>&lt;a href=&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: attribute value requires closing quote",
			markdown: "<a href=\"/url>",
			wantHTML: `<p>&lt;a href=&#34;/url&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: single quoted attribute value requires closing quote",
			markdown: "<a href='/url>",
			wantHTML: `<p>&lt;a href=&#39;/url&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: self closing suffix must run directly to closing bracket",
			markdown: "<br/ x>",
			wantHTML: `<p>&lt;br/ x&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: closing tag allows trailing spaces only",
			markdown: "</span >",
			wantHTML: `<p></span ></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: closing tag rejects extra content",
			markdown: "</span x>",
			wantHTML: `<p>&lt;/span x&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: unquoted attribute value may not contain less than",
			markdown: "<a href=x<y>",
			wantHTML: `<p>&lt;a href=x<y></p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: unquoted attribute value may not contain greater than",
			markdown: "<a href=x>y>",
			wantHTML: `<p><a href=x>y&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "inline html: unquoted attribute value may not contain backtick",
			markdown: "<a href=x`y>",
			wantHTML: "<p>&lt;a href=x`y&gt;</p>",
			wantErr:  nil,
		},
		{
			name:     "inline html: lone opening angle bracket remains literal text",
			markdown: "<",
			wantHTML: `<p>&lt;</p>`,
			wantErr:  nil,
		},

		// HTML blocks

		{
			name:     "html block: single line comment block",
			markdown: "<!-- hello -->",
			wantHTML: "<!-- hello -->",
			wantErr:  nil,
		},
		{
			name: "html block: multiline comment block",
			markdown: md(
				"<!--",
				"hello",
				"-->",
			),
			wantHTML: "<!--\nhello\n-->",
			wantErr:  nil,
		},
		{
			name: "html block: comment terminator on opening line closes block immediately",
			markdown: md(
				"<!-- hello -->",
				"tail",
			),
			wantHTML: "<!-- hello --><p>tail</p>",
			wantErr:  nil,
		},
		{
			name: "html block: comment block runs to eof when unterminated",
			markdown: md(
				"<!--",
				"hello",
			),
			wantHTML: "<!--\nhello",
			wantErr:  nil,
		},
		{
			name:     "html block: single line cdata block",
			markdown: "<![CDATA[hello]]>",
			wantHTML: "<![CDATA[hello]]>",
			wantErr:  nil,
		},
		{
			name: "html block: multiline cdata block",
			markdown: md(
				"<![CDATA[",
				"hello",
				"]]>",
			),
			wantHTML: "<![CDATA[\nhello\n]]>",
			wantErr:  nil,
		},
		{
			name: "html block: cdata terminator on opening line closes block immediately",
			markdown: md(
				"<![CDATA[hello]]>",
				"tail",
			),
			wantHTML: "<![CDATA[hello]]><p>tail</p>",
			wantErr:  nil,
		},
		{
			name: "html block: cdata block runs to eof when unterminated",
			markdown: md(
				"<![CDATA[",
				"hello",
			),
			wantHTML: "<![CDATA[\nhello",
			wantErr:  nil,
		},
		{
			name:     "html block: single line processing instruction block",
			markdown: "<?php?>",
			wantHTML: "<?php?>",
			wantErr:  nil,
		},
		{
			name: "html block: multiline processing instruction block",
			markdown: md(
				"<?",
				"hello",
				"?>",
			),
			wantHTML: "<?\nhello\n?>",
			wantErr:  nil,
		},
		{
			name: "html block: processing instruction terminator on opening line closes block immediately",
			markdown: md(
				"<?hello?>",
				"tail",
			),
			wantHTML: "<?hello?><p>tail</p>",
			wantErr:  nil,
		},
		{
			name: "html block: processing instruction block runs to eof when unterminated",
			markdown: md(
				"<?",
				"hello",
			),
			wantHTML: "<?\nhello",
			wantErr:  nil,
		},
		{
			name:     "html block: single line declaration block",
			markdown: "<!DOCTYPE html>",
			wantHTML: "<!DOCTYPE html>",
			wantErr:  nil,
		},
		{
			name: "html block: multiline declaration block terminates on first greater-than",
			markdown: md(
				"<!DOCTYPE",
				"html>",
				"tail",
			),
			wantHTML: "<!DOCTYPE\nhtml><p>tail</p>",
			wantErr:  nil,
		},
		{
			name: "html block: declaration terminator on opening line closes block immediately",
			markdown: md(
				"<!DOCTYPE html>",
				"tail",
			),
			wantHTML: "<!DOCTYPE html><p>tail</p>",
			wantErr:  nil,
		},
		{
			name: "html block: declaration block runs to eof when unterminated",
			markdown: md(
				"<!DOCTYPE",
				"html",
			),
			wantHTML: "<!DOCTYPE\nhtml",
			wantErr:  nil,
		},
		{
			name:     "html block: named opening tag alone starts block",
			markdown: "<div>",
			wantHTML: "<div>",
			wantErr:  nil,
		},
		{
			name:     "html block: named closing tag alone starts block",
			markdown: "</div>",
			wantHTML: "</div>",
			wantErr:  nil,
		},
		{
			name:     "html block: named self-closing tag starts block",
			markdown: "<div/>",
			wantHTML: "<div/>",
			wantErr:  nil,
		},
		{
			name:     "html block: named self-closing tag with space before closer starts block",
			markdown: "<div / >",
			wantHTML: "<div / >",
			wantErr:  nil,
		},
		{
			name:     "html block: named tag with attributes starts block",
			markdown: `<div class="x">`,
			wantHTML: `<div class="x">`,
			wantErr:  nil,
		},
		{
			name:     "html block: named tag with trailing text on same line still starts block",
			markdown: "<div>hello",
			wantHTML: "<div>hello",
			wantErr:  nil,
		},
		{
			name: "html block: named tag block continues until blank line",
			markdown: md(
				"<div>",
				"hello",
				"world",
				"",
				"tail",
			),
			wantHTML: "<div>\nhello\nworld<p>tail</p>",
			wantErr:  nil,
		},
		{
			name: "html block: named tag block runs to eof without blank line",
			markdown: md(
				"<div>",
				"hello",
			),
			wantHTML: "<div>\nhello",
			wantErr:  nil,
		},
		{
			name: "html block: named tag block does not terminate on another tag line alone",
			markdown: md(
				"<div>",
				"</div>",
				"tail",
			),
			wantHTML: "<div>\n</div>\ntail",
			wantErr:  nil,
		},
		{
			name:     "html block: named tag with one leading space allowed",
			markdown: " <div>",
			wantHTML: " <div>",
			wantErr:  nil,
		},
		{
			name:     "html block: named tag with two leading spaces allowed",
			markdown: "  <div>",
			wantHTML: "  <div>",
			wantErr:  nil,
		},
		{
			name:     "html block: named tag with three leading spaces allowed",
			markdown: "   <div>",
			wantHTML: "   <div>",
			wantErr:  nil,
		},
		{
			name:     "html block: named tag with four leading spaces is not html block",
			markdown: "    <div>",
			wantHTML: "<pre><code>&lt;div&gt;</code></pre>",
			wantErr:  nil,
		},
		{
			name:     "html block: comment with three leading spaces allowed",
			markdown: "   <!-- hello -->",
			wantHTML: "   <!-- hello -->",
			wantErr:  nil,
		},
		{
			name:     "html block: comment with four leading spaces is not html block",
			markdown: "    <!-- hello -->",
			wantHTML: "<pre><code>&lt;!-- hello --&gt;</code></pre>",
			wantErr:  nil,
		},
		{
			name:     "html block: cdata with three leading spaces allowed",
			markdown: "   <![CDATA[hello]]>",
			wantHTML: "   <![CDATA[hello]]>",
			wantErr:  nil,
		},
		{
			name:     "html block: processing instruction with three leading spaces allowed",
			markdown: "    <?hello?>",
			wantHTML: "<pre><code>&lt;?hello?&gt;</code></pre>",
			wantErr:  nil,
		},
		{
			name:     "html block: declaration with three leading spaces allowed",
			markdown: "   <!DOCTYPE html>",
			wantHTML: "   <!DOCTYPE html>",
			wantErr:  nil,
		},
		{
			name:     "html block: unknown named tag is not recognized as block html",
			markdown: "<not-a-tag>",
			wantHTML: "<p><not-a-tag></p>",
			wantErr:  nil,
		},
		{
			name:     "html block: non-alpha tag opener is not recognized as block html",
			markdown: "<1div>",
			wantHTML: "<p>&lt;1div&gt;</p>",
			wantErr:  nil,
		},
		{
			name:     "html block: non-alpha tag name is not recognized as block html",
			markdown: "</1div>",
			wantHTML: "<p>&lt;/1div&gt;</p>",
			wantErr:  nil,
		},
		{
			name:     "html block: named tag without closing angle bracket is not html block",
			markdown: "<div",
			wantHTML: "<p>&lt;div</p>",
			wantErr:  nil,
		},
		{
			name:     "html block: named closing tag without closing angle bracket is not html block",
			markdown: "</div",
			wantHTML: "<p>&lt;/div</p>",
			wantErr:  nil,
		},
		{
			name:     "html block: named tag with invalid punctuation in head is not html block",
			markdown: "<div!>",
			wantHTML: "<p>&lt;div!&gt;</p>",
			wantErr:  nil,
		},
		{
			name:     "html block: named tag with slash not followed by optional whitespace and closer is not html block",
			markdown: "<div/x>",
			wantHTML: "<p>&lt;div/x&gt;</p>",
			wantErr:  nil,
		},
		{
			name:     "html block: named tag with attributes but no closing angle bracket is not html block",
			markdown: `<div class="x"`,
			wantHTML: `<p>&lt;div class=&#34;x&#34;</p>`,
			wantErr:  nil,
		},
		{
			name:     "html block: non-html less-than text is not html block",
			markdown: "< hello",
			wantHTML: "<p>&lt; hello</p>",
			wantErr:  nil,
		},
		{
			name:     "html block: bare less-than is not html block",
			markdown: "<",
			wantHTML: "<p>&lt;</p>",
			wantErr:  nil,
		},
		{
			name: "html block: less-than followed by question mark without terminator still starts processing instruction block",
			markdown: md(
				"<?hello",
				"tail",
			),
			wantHTML: "<?hello\ntail",
			wantErr:  nil,
		},
		{
			name: "html block: less-than bang without greater-than still starts declaration block",
			markdown: md(
				"<!hello",
				"tail",
			),
			wantHTML: "<!hello\ntail",
			wantErr:  nil,
		},
		{
			name: "html block: named tag interrupted by blank line then paragraph",
			markdown: md(
				"<div>",
				"hello",
				"",
				"world",
			),
			wantHTML: "<div>\nhello<p>world</p>",
			wantErr:  nil,
		},
		{
			name: "html block: comment block may contain blank lines before terminator",
			markdown: md(
				"<!--",
				"",
				"hello",
				"-->",
				"tail",
			),
			wantHTML: "<!--\n\nhello\n--><p>tail</p>",
			wantErr:  nil,
		},
		{
			name: "html block: cdata block may contain blank lines before terminator",
			markdown: md(
				"<![CDATA[",
				"",
				"hello",
				"]]>",
				"tail",
			),
			wantHTML: "<![CDATA[\n\nhello\n]]><p>tail</p>",
			wantErr:  nil,
		},
		{
			name: "html block: processing instruction block may contain blank lines before terminator",
			markdown: md(
				"<?",
				"",
				"hello",
				"?>",
				"tail",
			),
			wantHTML: "<?\n\nhello\n?><p>tail</p>",
			wantErr:  nil,
		},
		{
			name: "html block: declaration block may contain blank lines before terminator",
			markdown: md(
				"<!DOCTYPE",
				"",
				"html>",
				"tail",
			),
			wantHTML: "<!DOCTYPE\n\nhtml><p>tail</p>",
			wantErr:  nil,
		},
		{
			name: "html block: named tag block may contain html-looking lines until blank line",
			markdown: md(
				"<div>",
				"<span>",
				"</span>",
				"",
				"tail",
			),
			wantHTML: "<div>\n<span>\n</span><p>tail</p>",
			wantErr:  nil,
		},
		{
			name: "html block: named closing tag block continues until blank line",
			markdown: md(
				"</div>",
				"hello",
				"",
				"tail",
			),
			wantHTML: "</div>\nhello<p>tail</p>",
			wantErr:  nil,
		},
		{
			name:     "html block: whitelisted tag name is case-insensitive",
			markdown: "<DIV>",
			wantHTML: "<DIV>",
			wantErr:  nil,
		},
		{
			name:     "html block: closing whitelisted tag name is case-insensitive",
			markdown: "</DIV>",
			wantHTML: "</DIV>",
			wantErr:  nil,
		},
		{
			name:     "html block: alphanumeric tag name allowed when whitelisted",
			markdown: "<h1>",
			wantHTML: "<h1>",
			wantErr:  nil,
		},
		{
			name:     "html block: non-whitelisted alphanumeric tag name rejected",
			markdown: "<x1>",
			wantHTML: "<p><x1></p>",
			wantErr:  nil,
		},

		// escapes

		{
			name:     "escape: escaped emphasis opener remains literal text",
			markdown: "\\*a*",
			wantHTML: `<p>*a*</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped emphasis closer remains literal text",
			markdown: "*a\\*",
			wantHTML: `<p>*a*</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped strong delimiter remains literal text",
			markdown: "\\**a**",
			wantHTML: `<p>**a**</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped underscore remains literal text",
			markdown: "\\_a_",
			wantHTML: `<p>_a_</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped backtick remains literal text",
			markdown: "\\`code`",
			wantHTML: "<p>`code`</p>",
			wantErr:  nil,
		},
		{
			name:     "escape: escaped opening bracket prevents link formation",
			markdown: "\\[label](/url)",
			wantHTML: `<p>[label](/url)</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped closing bracket prevents link formation",
			markdown: "[label\\](/url)",
			wantHTML: `<p>[label](/url)</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped opening parenthesis prevents inline link tail",
			markdown: "[label]\\(/url)",
			wantHTML: `<p>[label](/url)</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped closing parenthesis remains literal text",
			markdown: "[label](/url\\)",
			wantHTML: `<p>[label](/url)</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped opening angle remains literal text",
			markdown: "\\<span>",
			wantHTML: `<p>&lt;span&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped closing angle remains literal text",
			markdown: "<span\\>",
			wantHTML: `<p>&lt;span&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped bang prevents image formation",
			markdown: "\\![alt](/img.png)",
			wantHTML: `<p>!<a href="/img.png">alt</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped image opener becomes literal punctuation",
			markdown: "!\\[alt](/img.png)",
			wantHTML: `<p>![alt](/img.png)</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped delimiter inside emphasis is literal text",
			markdown: "*a \\* b*",
			wantHTML: `<p><em>a * b</em></p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped delimiter inside strong is literal text",
			markdown: "**a \\* b**",
			wantHTML: `<p><strong>a * b</strong></p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped backticks prevent code span formation",
			markdown: "\\`code\\`",
			wantHTML: "<p>`code`</p>",
			wantErr:  nil,
		},
		{
			name:     "escape: escaped brackets are literal inside text",
			markdown: "\\[a\\]",
			wantHTML: `<p>[a]</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped parentheses are literal inside text",
			markdown: "\\(a\\)",
			wantHTML: `<p>(a)</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped angle brackets are literal inside text",
			markdown: "\\<a\\>",
			wantHTML: `<p>&lt;a&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped backslash yields literal backslash",
			markdown: "\\\\",
			wantHTML: `<p>\</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: trailing backslash remains literal text",
			markdown: "\\",
			wantHTML: `<p>\</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: backslash before ordinary text remains literal",
			markdown: "\\a",
			wantHTML: `<p>\a</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped delimiter does not prevent later valid emphasis",
			markdown: "\\*a* *b*",
			wantHTML: `<p>*a* <em>b</em></p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped opener leaves following link syntax literal",
			markdown: "\\[x](y)",
			wantHTML: `<p>[x](y)</p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped bang leaves following bracket construct as link syntax",
			markdown: "\\![x](y)",
			wantHTML: `<p>!<a href="y">x</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "escape: escaped punctuation is preserved in surrounding text",
			markdown: "a \\* b",
			wantHTML: `<p>a * b</p>`,
			wantErr:  nil,
		},

		// reference links

		{
			name:     "reference link: full reference resolves to link",
			markdown: "[label][ref]\n\n[ref]: /url",
			wantHTML: `<p><a href="/url">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: collapsed reference resolves to link",
			markdown: "[label][]\n\n[label]: /url",
			wantHTML: `<p><a href="/url">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: shortcut reference resolves to link",
			markdown: "[label]\n\n[label]: /url",
			wantHTML: `<p><a href="/url">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: visible label text is independent of referenced definition label",
			markdown: "[visible][ref]\n\n[ref]: /url",
			wantHTML: `<p><a href="/url">visible</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: reference definition may provide title",
			markdown: "[label][ref]\n\n[ref]: /url \"title\"",
			wantHTML: `<p><a href="/url" title="title">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: collapsed reference may use definition title",
			markdown: "[label][]\n\n[label]: /url \"title\"",
			wantHTML: `<p><a href="/url" title="title">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: shortcut reference may use definition title",
			markdown: "[label]\n\n[label]: /url \"title\"",
			wantHTML: `<p><a href="/url" title="title">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: nested inline content is allowed in label",
			markdown: "[a *b* c][ref]\n\n[ref]: /url",
			wantHTML: `<p><a href="/url">a <em>b</em> c</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: code span is allowed in label",
			markdown: "[a `b` c][ref]\n\n[ref]: /url",
			wantHTML: `<p><a href="/url">a <code>b</code> c</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: image is allowed in label",
			markdown: "[![alt](/img.png)][ref]\n\n[ref]: /url",
			wantHTML: `<p><a href="/url"><img alt="alt" src="/img.png"></a></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: surrounding text is preserved",
			markdown: "a [label][ref] b\n\n[ref]: /url",
			wantHTML: `<p>a <a href="/url">label</a> b</p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: full reference falls back to literal text when definition is missing",
			markdown: "[label][ref]",
			wantHTML: `<p>[label][ref]</p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: collapsed reference falls back to literal text when definition is missing",
			markdown: "[label][]",
			wantHTML: `<p>[label][]</p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: shortcut reference falls back to literal text when definition is missing",
			markdown: "[label]",
			wantHTML: `<p>[label]</p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: full reference falls back to literal text when closing bracket is missing",
			markdown: "[label][ref",
			wantHTML: `<p>[label][ref</p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: collapsed reference takes precedence over shortcut reference",
			markdown: "[label][]\n\n[label]: /url",
			wantHTML: `<p><a href="/url">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: full reference takes precedence over shortcut reference",
			markdown: "[label][ref]\n\n[label]: /wrong\n[ref]: /right",
			wantHTML: `<p><a href="/right">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: inner reference link may still resolve when outer link is rejected",
			markdown: "[outer [inner][in]]\n\n[in]: /in",
			wantHTML: `<p>[outer <a href="/in">inner</a>]</p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: first matching definition wins",
			markdown: "[label]\n\n[label]: /first\n[label]: /second",
			wantHTML: `<p><a href="/first">label</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference link: definition may be indented up to three spaces",
			markdown: "[label]\n\n   [label]: /url",
			wantHTML: `<p><a href="/url">label</a></p>`,
			wantErr:  nil,
		},

		// reference images

		{
			name:     "reference image: full reference resolves to image",
			markdown: "![alt][ref]\n\n[ref]: /img.png",
			wantHTML: `<p><img alt="alt" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: collapsed reference resolves to image",
			markdown: "![alt][]\n\n[alt]: /img.png",
			wantHTML: `<p><img alt="alt" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: shortcut reference resolves to image",
			markdown: "![alt]\n\n[alt]: /img.png",
			wantHTML: `<p><img alt="alt" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: full reference uses referenced label rather than visible label",
			markdown: "![visible][ref]\n\n[ref]: /img.png",
			wantHTML: `<p><img alt="visible" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: reference definition may provide title",
			markdown: "![alt][ref]\n\n[ref]: /img.png \"title\"",
			wantHTML: `<p><img alt="alt" src="/img.png" title="title"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: collapsed reference may use definition title",
			markdown: "![alt][]\n\n[alt]: /img.png \"title\"",
			wantHTML: `<p><img alt="alt" src="/img.png" title="title"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: shortcut reference may use definition title",
			markdown: "![alt]\n\n[alt]: /img.png \"title\"",
			wantHTML: `<p><img alt="alt" src="/img.png" title="title"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: nested inline content is allowed in label",
			markdown: "![a *b* c][ref]\n\n[ref]: /img.png",
			wantHTML: `<p><img alt="a b c" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: code span is allowed in label",
			markdown: "![a `b` c][ref]\n\n[ref]: /img.png",
			wantHTML: `<p><img alt="a b c" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: link is allowed in image label",
			markdown: "![see [this](/url)][ref]\n\n[ref]: /img.png",
			wantHTML: `<p><img alt="see this" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: surrounding text is preserved",
			markdown: "a ![alt][ref] b\n\n[ref]: /img.png",
			wantHTML: `<p>a <img alt="alt" src="/img.png"> b</p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: full reference falls back to literal text when definition is missing",
			markdown: "![alt][ref]",
			wantHTML: `<p>![alt][ref]</p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: collapsed reference falls back to literal text when definition is missing",
			markdown: "![alt][]",
			wantHTML: `<p>![alt][]</p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: shortcut reference falls back to literal text when definition is missing",
			markdown: "![alt]",
			wantHTML: `<p>![alt]</p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: full reference falls back to literal text when closing bracket is missing",
			markdown: "![alt][ref",
			wantHTML: `<p>![alt][ref</p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: collapsed reference takes precedence over shortcut reference",
			markdown: "![alt][]\n\n[alt]: /img.png",
			wantHTML: `<p><img alt="alt" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: nested images are allowed in reference form",
			markdown: "![outer ![inner][in]][out]\n\n[in]: /in.png\n[out]: /out.png",
			wantHTML: `<p><img alt="outer inner" src="/out.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: full reference takes precedence over shortcut reference",
			markdown: "![alt][ref]\n\n[alt]: /wrong.png\n[ref]: /right.png",
			wantHTML: `<p><img alt="alt" src="/right.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: first matching definition wins",
			markdown: "![alt]\n\n[alt]: /first.png\n[alt]: /second.png",
			wantHTML: `<p><img alt="alt" src="/first.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "reference image: definition may be indented up to three spaces",
			markdown: "![alt]\n\n   [alt]: /img.png",
			wantHTML: `<p><img alt="alt" src="/img.png"></p>`,
			wantErr:  nil,
		},

		// precedence and ambiguity

		{
			name:     "precedence: code span suppresses emphasis parsing",
			markdown: "*a `*`*",
			wantHTML: `<p><em>a <code>*</code></em></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: code span suppresses link parsing",
			markdown: "[`[x](y)`](/url)",
			wantHTML: `<p><a href="/url"><code>[x](y)</code></a></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: code span suppresses image parsing",
			markdown: "`![alt](/img.png)`",
			wantHTML: `<p><code>![alt](/img.png)</code></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: code span suppresses raw html parsing",
			markdown: "`<span>`",
			wantHTML: `<p><code>&lt;span&gt;</code></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: uri autolink takes precedence over raw html fallback",
			markdown: "<https://example.com>",
			wantHTML: `<p><a href="https://example.com">https://example.com</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: email autolink takes precedence over raw html fallback",
			markdown: "<user@example.com>",
			wantHTML: `<p><a href="mailto:user@example.com">user@example.com</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: raw html resolves when angle construct is not an autolink",
			markdown: "<span>",
			wantHTML: `<p><span></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: invalid angle construct falls back to literal text",
			markdown: "<local@domain>",
			wantHTML: `<p><a href="mailto:local@domain">local@domain</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: emphasis is parsed inside link label",
			markdown: "[a *b* c](/url)",
			wantHTML: `<p><a href="/url">a <em>b</em> c</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: strong emphasis is parsed inside image label",
			markdown: "![a **b** c](/img.png)",
			wantHTML: `<p><img alt="a b c" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: nested link is rejected inside link label",
			markdown: "[outer [inner](/in)](/out)",
			wantHTML: `<p>[outer <a href="/in">inner</a>](/out)</p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: link is allowed inside image label",
			markdown: "![see [this](/url)](/img.png)",
			wantHTML: `<p><img alt="see this" src="/img.png"></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: escape prevents emphasis from binding",
			markdown: "\\*a*",
			wantHTML: `<p>*a*</p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: escape prevents link formation",
			markdown: "\\[x](y)",
			wantHTML: `<p>[x](y)</p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: escaped bang prevents image formation but preserves link parsing",
			markdown: "\\![x](y)",
			wantHTML: `<p>!<a href="y">x</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: escape prevents html recognition from opening angle bracket",
			markdown: "\\<span>",
			wantHTML: `<p>&lt;span&gt;</p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: emphasis may wrap a link",
			markdown: "*[x](/url)*",
			wantHTML: `<p><em><a href="/url">x</a></em></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: strong emphasis may wrap an image",
			markdown: "**![alt](/img.png)**",
			wantHTML: `<p><strong><img alt="alt" src="/img.png"></strong></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: link label may contain code span and emphasis together",
			markdown: "[a `b` *c*](/url)",
			wantHTML: `<p><a href="/url">a <code>b</code> <em>c</em></a></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: code span prevents delimiter participation inside emphasis run",
			markdown: "*a `*` b*",
			wantHTML: `<p><em>a <code>*</code> b</em></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: failed inline link falls back without preventing later emphasis",
			markdown: "[x](a(b) *c*",
			wantHTML: `<p>[x](a(b) <em>c</em></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: failed autolink falls back to text and allows later emphasis",
			markdown: "<local@domain> *x*",
			wantHTML: `<p><a href="mailto:local@domain">local@domain</a> <em>x</em></p>`,
			wantErr:  nil,
		},
		{
			name:     "precedence: failed raw html falls back to text and allows later link",
			markdown: "<1tag> [x](/url)",
			wantHTML: `<p>&lt;1tag&gt; <a href="/url">x</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "composite: emphasis code span and link may coexist in one paragraph",
			markdown: "a *b* `c` [d](/url) e",
			wantHTML: `<p>a <em>b</em> <code>c</code> <a href="/url">d</a> e</p>`,
			wantErr:  nil,
		},
		{
			name:     "composite: image link and autolink may coexist in one paragraph",
			markdown: "![alt](/img.png) [x](/url) <https://example.com>",
			wantHTML: `<p><img alt="alt" src="/img.png"> <a href="/url">x</a> <a href="https://example.com">https://example.com</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "composite: escaped punctuation does not disrupt neighboring inline constructs",
			markdown: "\\* a *b* [c](/url) `d`",
			wantHTML: `<p>* a <em>b</em> <a href="/url">c</a> <code>d</code></p>`,
			wantErr:  nil,
		},
		{
			name:     "adjacency: consecutive links parse independently",
			markdown: "[a](/x)[b](/y)",
			wantHTML: `<p><a href="/x">a</a><a href="/y">b</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "adjacency: code span adjacent to link parses independently",
			markdown: "`a`[b](/url)",
			wantHTML: `<p><code>a</code><a href="/url">b</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "adjacency: html adjacent to autolink parses independently",
			markdown: "<span><https://example.com>",
			wantHTML: `<p><span><a href="https://example.com">https://example.com</a></p>`,
			wantErr:  nil,
		},
		{
			name:     "nesting: emphasis inside link inside emphasis resolves correctly",
			markdown: "*[a **b** c](/url)*",
			wantHTML: `<p><em><a href="/url">a <strong>b</strong> c</a></em></p>`,
			wantErr:  nil,
		},
		{
			name:     "nesting: image inside link inside emphasis resolves correctly",
			markdown: "*[![alt](/img.png)](/url)*",
			wantHTML: `<p><em><a href="/url"><img alt="alt" src="/img.png"></a></em></p>`,
			wantErr:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := HTML(tc.markdown)

			assert.Equal(t, got, tc.wantHTML)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func md(xs ...string) string {
	return strings.Join(xs, "\n")
}
