package markdown_test

import (
	"strings"
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestCompile(t *testing.T) {
	testCases := []struct {
		name    string
		md      string
		want    string
		wantErr error
	}{
		{
			name:    "plain text: one paragraph",
			md:      "hello",
			want:    `<p>hello</p>`,
			wantErr: nil,
		},
		{
			name:    "plain text: soft break paragraph",
			md:      "a\nb",
			want:    `<p>a b</p>`,
			wantErr: nil,
		},
		{
			name:    "plain text: hard break paragraph (spaces)",
			md:      "a  \nb",
			want:    `<p>a<br>b</p>`,
			wantErr: nil,
		},
		{
			name:    "plain text: hard break paragraph (backslash)",
			md:      "a\\\nb",
			want:    `<p>a<br>b</p>`,
			wantErr: nil,
		},
		{
			name:    "plain text: blank line splits paragraphs",
			md:      "a\n\nb",
			want:    `<p>a</p><p>b</p>`,
			wantErr: nil,
		},
		{
			name:    "header: level 1",
			md:      "# header",
			want:    `<h1>header</h1>`,
			wantErr: nil,
		},
		{
			name:    "header: level 6",
			md:      "###### header",
			want:    `<h6>header</h6>`,
			wantErr: nil,
		},
		{
			name: "header then paragraph",
			md: strings.Join([]string{
				"# h",
				"a",
			}, "\n"),
			want:    `<h1>h</h1><p>a</p>`,
			wantErr: nil,
		},
		{
			name: "paragraph then header",
			md: strings.Join([]string{
				"a",
				"# h",
			}, "\n"),
			want:    `<p>a</p><h1>h</h1>`,
			wantErr: nil,
		},
		{
			name:    "thematic break",
			md:      "---",
			want:    "<hr>",
			wantErr: nil,
		},
		{
			name:    "block quote: plain text",
			md:      "> quote",
			want:    "<blockquote><p>quote</p></blockquote>",
			wantErr: nil,
		},
		{
			name: "block quote: multiple lines",
			md: strings.Join([]string{
				"> a",
				"> b",
			}, "\n"),
			want:    "<blockquote><p>a b</p></blockquote>",
			wantErr: nil,
		},
		{
			name: "block quote: separated by blank line",
			md: strings.Join([]string{
				"> a",
				">",
				"> b",
			}, "\n"),
			want:    "<blockquote><p>a</p><p>b</p></blockquote>",
			wantErr: nil,
		},
		{
			name: "block quote: nested layers",
			md: strings.Join([]string{
				"> a",
				">> nested",
				"> b",
			}, "\n"),
			want:    "<blockquote><p>a</p><blockquote><p>nested</p></blockquote><p>b</p></blockquote>",
			wantErr: nil,
		},
		{
			name:    "block quote: header text",
			md:      "> # h",
			want:    "<blockquote><h1>h</h1></blockquote>",
			wantErr: nil,
		},
		{
			name:    "block quote: thematic break",
			md:      "> ---",
			want:    "<blockquote><hr></blockquote>",
			wantErr: nil,
		},
		{
			name: "setext: level 1",
			md: strings.Join([]string{
				"h",
				"===",
			}, "\n"),
			want:    "<h1>h</h1>",
			wantErr: nil,
		},
		{
			name: "setext: level 2",
			md: strings.Join([]string{
				"h",
				"---",
			}, "\n"),
			want:    "<h2>h</h2>",
			wantErr: nil,
		},
		{
			name: "ul: two items",
			md: strings.Join([]string{
				"- a",
				"- b",
			}, "\n"),
			want:    "<ul><li>a</li><li>b</li></ul>",
			wantErr: nil,
		},
		{
			name: "ul: nested list",
			md: strings.Join([]string{
				"- a",
				"  - b",
				"- c",
			}, "\n"),
			want:    "<ul><li>a<ul><li>b</li></ul></li><li>c</li></ul>",
			wantErr: nil,
		},
		{
			name: "ul: loose list via blank line between items",
			md: strings.Join([]string{
				"- a",
				"",
				"- b",
			}, "\n"),
			want:    "<ul><li><p>a</p></li><li><p>b</p></li></ul>",
			wantErr: nil,
		},
		{
			name: "ul: trailing blank line after last item rolls back",
			md: strings.Join([]string{
				"- a",
				"",
				"x",
			}, "\n"),
			want:    "<ul><li>a</li></ul><p>x</p>",
			wantErr: nil,
		},
		{
			name: "ul: loose list via blank line inside an item",
			md: strings.Join([]string{
				"- a",
				"",
				"  x",
			}, "\n"),
			want:    "<ul><li><p>a</p><p>x</p></li></ul>",
			wantErr: nil,
		},
		{
			name: "ul: nested list does not force looseness",
			md: strings.Join([]string{
				"- a",
				"  - b",
			}, "\n"),
			want:    "<ul><li>a<ul><li>b</li></ul></li></ul>",
			wantErr: nil,
		},
		{
			name: "ul: dedent ends list, next line becomes paragraph",
			md: strings.Join([]string{
				"- a",
				"x",
			}, "\n"),
			want:    "<ul><li>a</li></ul><p>x</p>",
			wantErr: nil,
		},
		{
			name: "ol: tight list unwraps paragraphs",
			md: strings.Join([]string{
				"1. a",
				"2. b",
			}, "\n"),
			want:    "<ol><li>a</li><li>b</li></ol>",
			wantErr: nil,
		},
		{
			name: "ol: loose list keeps paragraph wrappers",
			md: strings.Join([]string{
				"1. a",
				"",
				"2. b",
			}, "\n"),
			want:    "<ol><li><p>a</p></li><li><p>b</p></li></ol>",
			wantErr: nil,
		},
		{
			name: "ol: tight nested list unwraps leading paragraph",
			md: strings.Join([]string{
				"1. a",
				"   1. b",
				"2. c",
			}, "\n"),
			want:    "<ol><li>a<ol><li>b</li></ol></li><li>c</li></ol>",
			wantErr: nil,
		},
		{
			name: "ol: start attribute emitted when first item is not 1",
			md: strings.Join([]string{
				"3. a",
				"4. b",
			}, "\n"),
			want:    `<ol start="3"><li>a</li><li>b</li></ol>`,
			wantErr: nil,
		},
		{
			name:    "icb: single line",
			md:      `    fmt.Println("hello")`,
			want:    "<pre><code>fmt.Println(&#34;hello&#34;)</code></pre>",
			wantErr: nil,
		},
		{
			name: "icb: multiple lines",
			md: strings.Join([]string{
				"    a := 1",
				"    b := 2",
				"    fmt.Println(a + b)",
			}, "\n"),
			want:    "<pre><code>a := 1\nb := 2\nfmt.Println(a + b)</code></pre>",
			wantErr: nil,
		},
		{
			name: "icb: preserves extra indentation",
			md: strings.Join([]string{
				"    if x {",
				"        y()",
				"    }",
			}, "\n"),
			want:    "<pre><code>if x {\n    y()\n}</code></pre>",
			wantErr: nil,
		},
		{
			name: "icb: preserves internal blank line",
			md: strings.Join([]string{
				"    line one",
				"",
				"    line two",
			}, "\n"),
			want:    "<pre><code>line one\n\nline two</code></pre>",
			wantErr: nil,
		},
		{
			name: "icb: trailing blanks only preserved if followed by valid payload lines",
			md: strings.Join([]string{
				"    line one",
				"",
				"",
				"paragraph",
			}, "\n"),
			want:    "<pre><code>line one</code></pre><p>paragraph</p>",
			wantErr: nil,
		},
		{
			name: "icb: ends at EOF",
			md: strings.Join([]string{
				"    last line",
				"    still code",
			}, "\n"),
			want:    "<pre><code>last line\nstill code</code></pre>",
			wantErr: nil,
		},
		{
			name: "fcb: backtick fence, single line",
			md: strings.Join([]string{
				"```",
				`fmt.Println("hello")`,
				"```",
			}, "\n"),
			want:    "<pre><code>fmt.Println(&#34;hello&#34;)</code></pre>",
			wantErr: nil,
		},
		{
			name: "fcb: tilde fence, single line",
			md: strings.Join([]string{
				"~~~",
				`fmt.Println("hello")`,
				"~~~",
			}, "\n"),
			want:    "<pre><code>fmt.Println(&#34;hello&#34;)</code></pre>",
			wantErr: nil,
		},
		{
			name: "fcb: language token",
			md: strings.Join([]string{
				"```go",
				`fmt.Println("hello")`,
				"```",
			}, "\n"),
			want:    `<pre><code class="language-go">fmt.Println(&#34;hello&#34;)</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fcb: language token followed by extra text",
			md: strings.Join([]string{
				"```go linenos",
				`fmt.Println("hello")`,
				"```",
			}, "\n"),
			want:    `<pre><code class="language-go">fmt.Println(&#34;hello&#34;)</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fcb: language token with trailing whitespace",
			md: strings.Join([]string{
				"```go    ",
				`fmt.Println("hello")`,
				"```",
			}, "\n"),
			want:    `<pre><code class="language-go">fmt.Println(&#34;hello&#34;)</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fcb: opener indentation strips matching payload indentation",
			md: strings.Join([]string{
				"  ```go",
				"  x := 1",
				"  y := 2",
				"  ```",
			}, "\n"),
			want:    "<pre><code class=\"language-go\">x := 1\ny := 2</code></pre>",
			wantErr: nil,
		},
		{
			name: "fcb: preserves extra indentation",
			md: strings.Join([]string{
				"  ```go",
				"  if x {",
				"      y()",
				"  }",
				"```",
			}, "\n"),
			want:    "<pre><code class=\"language-go\">if x {\n    y()\n}</code></pre>",
			wantErr: nil,
		},
		{
			name: "fcb: payload line with less indentation than opener indent",
			md: strings.Join([]string{
				"  ```",
				"x := 1",
				"  y := 2",
				"  ```",
			}, "\n"),
			want:    "<pre><code>x := 1\ny := 2</code></pre>",
			wantErr: nil,
		},
		{
			name: "fcb: blank line inside payload is preserved",
			md: strings.Join([]string{
				"```go",
				"line one",
				"",
				"line three",
				"```",
			}, "\n"),
			want:    "<pre><code class=\"language-go\">line one\n\nline three</code></pre>",
			wantErr: nil,
		},
		{
			name: "fcb: multiple payload lines preserve exact line boundaries",
			md: strings.Join([]string{
				"```",
				"a",
				"b",
				"c",
				"```",
			}, "\n"),
			want:    "<pre><code>a\nb\nc</code></pre>",
			wantErr: nil,
		},
		{
			name: "fcb: longer closing fence is accepted",
			md: strings.Join([]string{
				"```",
				"x := 1",
				"`````",
			}, "\n"),
			want:    "<pre><code>x := 1</code></pre>",
			wantErr: nil,
		},
		{
			name: "fcb: no closing fence runs to EOF",
			md: strings.Join([]string{
				"```",
				"x := 1",
				"y := 2",
			}, "\n"),
			want:    "<pre><code>x := 1\ny := 2</code></pre>",
			wantErr: nil,
		},
		{
			name: "fcb: markdown content in payload remains literal",
			md: strings.Join([]string{
				"```md",
				"# not a heading",
				"* not a list item",
				"**not emphasis**",
				"```",
			}, "\n"),
			want:    "<pre><code class=\"language-md\"># not a heading\n* not a list item\n**not emphasis**</code></pre>",
			wantErr: nil,
		},
		{
			name: "fcb: HTML-sensitive content is escaped",
			md: strings.Join([]string{
				"```html",
				`<div class="note">hi</div>`,
				"```",
			}, "\n"),
			want:    `<pre><code class="language-html">&lt;div class=&#34;note&#34;&gt;hi&lt;/div&gt;</code></pre>`,
			wantErr: nil,
		},
		{
			name: "fcb: backslashes and entities remain literal",
			md: strings.Join([]string{
				"```",
				`\*literal asterisk\*`,
				`&amp;`,
				"```",
			}, "\n"),
			want:    "<pre><code>\\*literal asterisk\\*\n&amp;amp;</code></pre>",
			wantErr: nil,
		},
		{
			name: "fcb: paragraph before and after block",
			md: strings.Join([]string{
				"before",
				"",
				"```go",
				"x := 1",
				"```",
				"",
				"after",
			}, "\n"),
			want:    "<p>before</p><pre><code class=\"language-go\">x := 1</code></pre><p>after</p>",
			wantErr: nil,
		},
		{
			name:    "html block: comment",
			md:      "<!-- comment -->",
			want:    "<!-- comment -->",
			wantErr: nil,
		},
		{
			name: "html block: comment, multi-line",
			md: strings.Join([]string{
				"<!--",
				"comment",
				"-->",
			}, "\n"),
			want:    "<!--\ncomment\n-->",
			wantErr: nil,
		},
		{
			name: "html block: comment, no terminator",
			md: strings.Join([]string{
				"<!--",
				"hello",
				"world",
			}, "\n"),
			want:    "<!--\nhello\nworld",
			wantErr: nil,
		},
		{
			name:    "html block: processing instruction",
			md:      `<?xml version="1.0"?>`,
			want:    `<?xml version="1.0"?>`,
			wantErr: nil,
		},
		{
			name:    "html block: declaration",
			md:      "<!DOCTYPE html>",
			want:    "<!DOCTYPE html>",
			wantErr: nil,
		},
		{
			name: "html block: cdata, multi-line",
			md: strings.Join([]string{
				"<![CDATA[",
				"a < b",
				"]]>",
			}, "\n"),
			want:    "<![CDATA[\na < b\n]]>",
			wantErr: nil,
		},
		{
			name: "html block: named tag block",
			md: strings.Join([]string{
				"<div>",
				"hello",
				"</div>",
			}, "\n"),
			want:    "<div>\nhello\n</div>",
			wantErr: nil,
		},
		{
			name: "html block: named tag with class",
			md: strings.Join([]string{
				`<section class="note">`,
				"<p>hello</p>",
				"</section>",
			}, "\n"),
			want:    "<section class=\"note\">\n<p>hello</p>\n</section>",
			wantErr: nil,
		},
		{
			name: "html block: named tag terminates at blank line",
			md: strings.Join([]string{
				"<div>",
				"hello",
				"</div>",
				"",
				"world",
			}, "\n"),
			want:    "<div>\nhello\n</div><p>world</p>",
			wantErr: nil,
		},
		{
			name: "html block: html interrupts paragraph",
			md: strings.Join([]string{
				"alpha",
				"<div>",
				"beta",
				"</div>",
				"omega",
			}, "\n"),
			want:    "<p>alpha</p><div>\nbeta\n</div>\nomega",
			wantErr: nil,
		},
		{
			name: "html block: html interrupts paragraph but terminates at new line",
			md: strings.Join([]string{
				"alpha",
				"<div>",
				"beta",
				"</div>",
				"",
				"omega",
			}, "\n"),
			want:    "<p>alpha</p><div>\nbeta\n</div><p>omega</p>",
			wantErr: nil,
		},
		{
			name: "html block: markdown looking text is preserved literally",
			md: strings.Join([]string{
				"<div>",
				"# not a heading",
				"* not a list",
				"</div>",
			}, "\n"),
			want:    "<div>\n# not a heading\n* not a list\n</div>",
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := markdown.HTML(tc.md)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
