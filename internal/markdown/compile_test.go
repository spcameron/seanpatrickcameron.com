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
		html    string
		wantErr error
	}{
		{
			name:    "plain text: one paragraph",
			md:      "hello",
			html:    `<p>hello</p>`,
			wantErr: nil,
		},
		{
			name:    "plain text: soft break paragraph",
			md:      "a\nb",
			html:    `<p>a b</p>`,
			wantErr: nil,
		},
		{
			name:    "plain text: hard break paragraph (spaces)",
			md:      "a  \nb",
			html:    `<p>a<br>b</p>`,
			wantErr: nil,
		},
		{
			name:    "plain text: hard break paragraph (backslash)",
			md:      "a\\\nb",
			html:    `<p>a<br>b</p>`,
			wantErr: nil,
		},
		{
			name:    "plain text: blank line splits paragraphs",
			md:      "a\n\nb",
			html:    `<p>a</p><p>b</p>`,
			wantErr: nil,
		},
		{
			name:    "header: level 1",
			md:      "# header",
			html:    `<h1>header</h1>`,
			wantErr: nil,
		},
		{
			name:    "header: level 6",
			md:      "###### header",
			html:    `<h6>header</h6>`,
			wantErr: nil,
		},
		{
			name: "header then paragraph",
			md: strings.Join([]string{
				"# h",
				"a",
			}, "\n"),
			html:    `<h1>h</h1><p>a</p>`,
			wantErr: nil,
		},
		{
			name: "paragraph then header",
			md: strings.Join([]string{
				"a",
				"# h",
			}, "\n"),
			html:    `<p>a</p><h1>h</h1>`,
			wantErr: nil,
		},
		{
			name:    "thematic break",
			md:      "---",
			html:    "<hr>",
			wantErr: nil,
		},
		{
			name:    "block quote: plain text",
			md:      "> quote",
			html:    "<blockquote><p>quote</p></blockquote>",
			wantErr: nil,
		},
		{
			name: "block quote: multiple lines",
			md: strings.Join([]string{
				"> a",
				"> b",
			}, "\n"),
			html:    "<blockquote><p>a b</p></blockquote>",
			wantErr: nil,
		},
		{
			name: "block quote: separated by blank line",
			md: strings.Join([]string{
				"> a",
				">",
				"> b",
			}, "\n"),
			html:    "<blockquote><p>a</p><p>b</p></blockquote>",
			wantErr: nil,
		},
		{
			name: "block quote: nested layers",
			md: strings.Join([]string{
				"> a",
				">> nested",
				"> b",
			}, "\n"),
			html:    "<blockquote><p>a</p><blockquote><p>nested</p></blockquote><p>b</p></blockquote>",
			wantErr: nil,
		},
		{
			name:    "block quote: header text",
			md:      "> # h",
			html:    "<blockquote><h1>h</h1></blockquote>",
			wantErr: nil,
		},
		{
			name:    "block quote: thematic break",
			md:      "> ---",
			html:    "<blockquote><hr></blockquote>",
			wantErr: nil,
		},
		{
			name: "setext: level 1",
			md: strings.Join([]string{
				"h",
				"===",
			}, "\n"),
			html:    "<h1>h</h1>",
			wantErr: nil,
		},
		{
			name: "setext: level 2",
			md: strings.Join([]string{
				"h",
				"---",
			}, "\n"),
			html:    "<h2>h</h2>",
			wantErr: nil,
		},
		{
			name: "ul: two items",
			md: strings.Join([]string{
				"- a",
				"- b",
			}, "\n"),
			html:    "<ul><li>a</li><li>b</li></ul>",
			wantErr: nil,
		},
		{
			name: "ul: nested list",
			md: strings.Join([]string{
				"- a",
				"  - b",
				"- c",
			}, "\n"),
			html:    "<ul><li>a<ul><li>b</li></ul></li><li>c</li></ul>",
			wantErr: nil,
		},
		{
			name: "ul: loose list via blank line between items",
			md: strings.Join([]string{
				"- a",
				"",
				"- b",
			}, "\n"),
			html:    "<ul><li><p>a</p></li><li><p>b</p></li></ul>",
			wantErr: nil,
		},
		{
			name: "ul: trailing blank line after last item rolls back",
			md: strings.Join([]string{
				"- a",
				"",
				"x",
			}, "\n"),
			html:    "<ul><li>a</li></ul><p>x</p>",
			wantErr: nil,
		},
		{
			name: "ul: loose list via blank line inside an item",
			md: strings.Join([]string{
				"- a",
				"",
				"  x",
			}, "\n"),
			html:    "<ul><li><p>a</p><p>x</p></li></ul>",
			wantErr: nil,
		},
		{
			name: "ul: nested list does not force looseness",
			md: strings.Join([]string{
				"- a",
				"  - b",
			}, "\n"),
			html:    "<ul><li>a<ul><li>b</li></ul></li></ul>",
			wantErr: nil,
		},
		{
			name: "ul: dedent ends list, next line becomes paragraph",
			md: strings.Join([]string{
				"- a",
				"x",
			}, "\n"),
			html:    "<ul><li>a</li></ul><p>x</p>",
			wantErr: nil,
		},
		{
			name: "ol: tight list unwraps paragraphs",
			md: strings.Join([]string{
				"1. a",
				"2. b",
			}, "\n"),
			html:    "<ol><li>a</li><li>b</li></ol>",
			wantErr: nil,
		},
		{
			name: "ol: loose list keeps paragraph wrappers",
			md: strings.Join([]string{
				"1. a",
				"",
				"2. b",
			}, "\n"),
			html:    "<ol><li><p>a</p></li><li><p>b</p></li></ol>",
			wantErr: nil,
		},
		{
			name: "ol: tight nested list unwraps leading paragraph",
			md: strings.Join([]string{
				"1. a",
				"   1. b",
				"2. c",
			}, "\n"),
			html:    "<ol><li>a<ol><li>b</li></ol></li><li>c</li></ol>",
			wantErr: nil,
		},
		{
			name: "ol: start attribute emitted when first item is not 1",
			md: strings.Join([]string{
				"3. a",
				"4. b",
			}, "\n"),
			html:    `<ol start="3"><li>a</li><li>b</li></ol>`,
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := markdown.HTML(tc.md)

			assert.Equal(t, got, tc.html)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
