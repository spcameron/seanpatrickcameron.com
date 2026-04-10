package codegen_test

import (
	"strings"
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/block"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/codegen"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/html"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/lower"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/require"
)

func TestGenerateHTML(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    html.Node
		wantErr error
	}{
		// Paragraphs and line breaks
		{
			name:  "paragraph with normal text",
			input: "paragraph",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.TextNode("paragraph"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "hard break renders (two spaces)",
			input: "a  \nb",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.TextNode("a"),
					html.VoidNode("br", nil),
					html.TextNode("b"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "hard break renders (backslash)",
			input: "a\\\nb",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.TextNode("a"),
					html.VoidNode("br", nil),
					html.TextNode("b"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "soft break renders as whitespace (space)",
			input: "a\nb",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.TextNode("a b"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: mixed soft and hard breaks across three lines",
			input: "alpha\nbeta  \ngamma",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.TextNode("alpha beta"),
					html.VoidNode("br", nil),
					html.TextNode("gamma"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: emphasis around hard break",
			input: "*alpha*  \nbeta",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.ElemNode(
						"em",
						nil,
						html.TextNode("alpha"),
					),
					html.VoidNode("br", nil),
					html.TextNode("beta"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: code span adjacent to soft break",
			input: "`alpha`\nbeta",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.ElemNode(
						"code",
						nil,
						html.TextNode("alpha"),
					),
					html.TextNode(" beta"),
				),
			),
			wantErr: nil,
		},

		// Headings and simple block forms
		{
			name:  "header with normal text",
			input: "# header",
			want: html.FragmentNode(
				html.ElemNode(
					"h1",
					nil,
					html.TextNode("header"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break",
			input: "---",
			want: html.FragmentNode(
				html.VoidNode("hr", nil),
			),
			wantErr: nil,
		},
		{
			name:  "html block",
			input: "<!-- comment -->",
			want: html.FragmentNode(
				html.FragmentNode(
					html.RawNode("<!-- comment -->"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "header: strong and emphasis",
			input: "# **alpha** *beta*",
			want: html.FragmentNode(
				html.ElemNode(
					"h1",
					nil,
					html.ElemNode(
						"strong",
						nil,
						html.TextNode("alpha"),
					),
					html.TextNode(" "),
					html.ElemNode(
						"em",
						nil,
						html.TextNode("beta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "setext header: emphasis",
			input: "*alpha*\n---",
			want: html.FragmentNode(
				html.ElemNode(
					"h2",
					nil,
					html.ElemNode(
						"em",
						nil,
						html.TextNode("alpha"),
					),
				),
			),
			wantErr: nil,
		},

		// Containers
		{
			name:  "block quote: plain text",
			input: "> quote",
			want: html.FragmentNode(
				html.ElemNode(
					"blockquote",
					nil,
					html.ElemNode(
						"p",
						nil,
						html.TextNode("quote"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: two paragraphs",
			input: "> alpha\n>\n> beta",
			want: html.FragmentNode(
				html.ElemNode(
					"blockquote",
					nil,
					html.ElemNode(
						"p",
						nil,
						html.TextNode("alpha"),
					),
					html.ElemNode(
						"p",
						nil,
						html.TextNode("beta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: nested block quote",
			input: "> outer\n> > inner",
			want: html.FragmentNode(
				html.ElemNode(
					"blockquote",
					nil,
					html.ElemNode(
						"p",
						nil,
						html.TextNode("outer"),
					),
					html.ElemNode(
						"blockquote",
						nil,
						html.ElemNode(
							"p",
							nil,
							html.TextNode("inner"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: contains list",
			input: "> - alpha\n> - beta",
			want: html.FragmentNode(
				html.ElemNode(
					"blockquote",
					nil,
					html.ElemNode(
						"ul",
						nil,
						html.ElemNode(
							"li",
							nil,
							html.TextNode("alpha"),
						),
						html.ElemNode(
							"li",
							nil,
							html.TextNode("beta"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "unordered list: two items",
			input: "- a\n- b",
			want: html.FragmentNode(
				html.ElemNode(
					"ul",
					nil,
					html.ElemNode(
						"li",
						nil,
						html.TextNode("a"),
					),
					html.ElemNode(
						"li",
						nil,
						html.TextNode("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "unordered list: loose list retains paragraph wrappers",
			input: "- alpha\n\n- beta",
			want: html.FragmentNode(
				html.ElemNode(
					"ul",
					nil,
					html.ElemNode(
						"li",
						nil,
						html.ElemNode(
							"p",
							nil,
							html.TextNode("alpha"),
						),
					),
					html.ElemNode(
						"li",
						nil,
						html.ElemNode(
							"p",
							nil,
							html.TextNode("beta"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "unordered list: tight list unwraps single paragraph children",
			input: "- alpha\n- beta",
			want: html.FragmentNode(
				html.ElemNode(
					"ul",
					nil,
					html.ElemNode(
						"li",
						nil,
						html.TextNode("alpha"),
					),
					html.ElemNode(
						"li",
						nil,
						html.TextNode("beta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "unordered list: item with paragraph and nested list",
			input: "- alpha\n  - beta",
			want: html.FragmentNode(
				html.ElemNode(
					"ul",
					nil,
					html.ElemNode(
						"li",
						nil,
						html.TextNode("alpha"),
						html.ElemNode(
							"ul",
							nil,
							html.ElemNode(
								"li",
								nil,
								html.TextNode("beta"),
							),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ordered list: two items",
			input: "1. a\n2. b",
			want: html.FragmentNode(
				html.ElemNode(
					"ol",
					nil,
					html.ElemNode(
						"li",
						nil,
						html.TextNode("a"),
					),
					html.ElemNode(
						"li",
						nil,
						html.TextNode("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ordered list: non-1 start emits start attribute",
			input: "3. alpha\n4. beta",
			want: html.FragmentNode(
				html.ElemNode(
					"ol",
					html.Attributes{"start": "3"},
					html.ElemNode(
						"li",
						nil,
						html.TextNode("alpha"),
					),
					html.ElemNode(
						"li",
						nil,
						html.TextNode("beta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ordered list: paren delimiter still renders as ol",
			input: "1) alpha\n2) beta",
			want: html.FragmentNode(
				html.ElemNode(
					"ol",
					nil,
					html.ElemNode(
						"li",
						nil,
						html.TextNode("alpha"),
					),
					html.ElemNode(
						"li",
						nil,
						html.TextNode("beta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "list item: indented code block child",
			input: "- alpha\n\n      beta",
			want: html.FragmentNode(
				html.ElemNode(
					"ul",
					nil,
					html.ElemNode(
						"li",
						nil,
						html.ElemNode(
							"p",
							nil,
							html.TextNode("alpha"),
						),
						html.ElemNode(
							"pre",
							nil,
							html.ElemNode(
								"code",
								nil,
								html.TextNode("beta"),
							),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "list item: block quote child",
			input: "- alpha\n  > beta",
			want: html.FragmentNode(
				html.ElemNode(
					"ul",
					nil,
					html.ElemNode(
						"li",
						nil,
						html.TextNode("alpha"),
						html.ElemNode(
							"blockquote",
							nil,
							html.ElemNode(
								"p",
								nil,
								html.TextNode("beta"),
							),
						),
					),
				),
			),
			wantErr: nil,
		},

		// Code blocks
		{
			name:  "indented code block",
			input: `    fmt.Println("hello")`,
			want: html.FragmentNode(
				html.ElemNode(
					"pre",
					nil,
					html.ElemNode(
						"code",
						nil,
						html.TextNode(`fmt.Println("hello")`),
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "fenced code block",
			input: strings.Join([]string{
				"```",
				`fmt.Println("hello")`,
				"```",
			}, "\n"),
			want: html.FragmentNode(
				html.ElemNode(
					"pre",
					nil,
					html.ElemNode(
						"code",
						nil,
						html.TextNode(`fmt.Println("hello")`),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "indented code block: multiple lines",
			input: "    alpha\n    beta",
			want: html.FragmentNode(
				html.ElemNode(
					"pre",
					nil,
					html.ElemNode(
						"code",
						nil,
						html.TextNode("alpha\nbeta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "indented code block: blank line in payload",
			input: "    alpha\n\n    beta",
			want: html.FragmentNode(
				html.ElemNode(
					"pre",
					nil,
					html.ElemNode(
						"code",
						nil,
						html.TextNode("alpha\n\nbeta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "fenced code block: language class emitted",
			input: "```go\nalpha\n```",
			want: html.FragmentNode(
				html.ElemNode(
					"pre",
					nil,
					html.ElemNode(
						"code",
						html.Attributes{"class": "language-go"},
						html.TextNode("alpha"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "fenced code block: info string ignores trailing words in class emission",
			input: "```go linenos\nalpha\n```",
			want: html.FragmentNode(
				html.ElemNode(
					"pre",
					nil,
					html.ElemNode(
						"code",
						html.Attributes{"class": "language-go"},
						html.TextNode("alpha"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "fenced code block: payload preserves leading spaces after indent stripping",
			input: "```\n  alpha\n```",
			want: html.FragmentNode(
				html.ElemNode(
					"pre",
					nil,
					html.ElemNode(
						"code",
						nil,
						html.TextNode("  alpha"),
					),
				),
			),
			wantErr: nil,
		},

		// HTML blocks
		{
			name:  "html block: multi-line comment",
			input: "<!--\nalpha\n-->",
			want: html.FragmentNode(
				html.FragmentNode(
					html.RawNode("<!--"),
					html.TextNode("\n"),
					html.RawNode("alpha"),
					html.TextNode("\n"),
					html.RawNode("-->"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "html block: named tag block",
			input: "<div>\nalpha\n</div>",
			want: html.FragmentNode(
				html.FragmentNode(
					html.RawNode("<div>"),
					html.TextNode("\n"),
					html.RawNode("alpha"),
					html.TextNode("\n"),
					html.RawNode("</div>"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "html block: processing instruction",
			input: "<?php\necho $a;\n?>",
			want: html.FragmentNode(
				html.FragmentNode(
					html.RawNode("<?php"),
					html.TextNode("\n"),
					html.RawNode("echo $a;"),
					html.TextNode("\n"),
					html.RawNode("?>"),
				),
			),
			wantErr: nil,
		},

		// Inline rendering through paragraphs
		{
			name:  "code span",
			input: "`abc`",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.ElemNode(
						"code",
						nil,
						html.TextNode("abc"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "emphasis",
			input: "*abc*",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.ElemNode(
						"em",
						nil,
						html.TextNode("abc"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "strong",
			input: "**abc**",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.ElemNode(
						"strong",
						nil,
						html.TextNode("abc"),
					),
				),
			),
			wantErr: nil,
		},

		// Links and autolinks
		{
			name:  "simple link",
			input: `[x](dest)`,
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.ElemNode(
						"a",
						html.Attributes{
							"href": "dest",
						},
						html.TextNode("x"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "link with title",
			input: `[x](dest "title")`,
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.ElemNode(
						"a",
						html.Attributes{
							"href":  "dest",
							"title": "title",
						},
						html.TextNode("x"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "autolink URI",
			input: "<https://google.com>",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.ElemNode(
						"a",
						html.Attributes{
							"href": "https://google.com",
						},
						html.TextNode("https://google.com"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "autolink email",
			input: "<local@domain.com>",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.ElemNode(
						"a",
						html.Attributes{
							"href": "mailto:local@domain.com",
						},
						html.TextNode("local@domain.com"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "link with emphasis in label",
			input: "[*x*](dest)",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.ElemNode(
						"a",
						html.Attributes{"href": "dest"},
						html.ElemNode(
							"em",
							nil,
							html.TextNode("x"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "link with title and nested strong label",
			input: "[**x**](dest \"title\")",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.ElemNode(
						"a",
						html.Attributes{
							"href":  "dest",
							"title": "title",
						},
						html.ElemNode(
							"strong",
							nil,
							html.TextNode("x"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "image",
			input: "![alt](img.png)",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.VoidNode(
						"img",
						html.Attributes{
							"alt": "alt",
							"src": "img.png",
						},
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "image with title",
			input: "![alt](img.png \"title\")",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.VoidNode(
						"img",
						html.Attributes{
							"alt":   "alt",
							"src":   "img.png",
							"title": "title",
						},
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "autolink email and URI in one paragraph",
			input: "<local@domain.com> <https://google.com>",
			want: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.ElemNode(
						"a",
						html.Attributes{"href": "mailto:local@domain.com"},
						html.TextNode("local@domain.com"),
					),
					html.TextNode(" "),
					html.ElemNode(
						"a",
						html.Attributes{"href": "https://google.com"},
						html.TextNode("https://google.com"),
					),
				),
			),
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := source.NewSource(tc.input)

			irDoc, err := block.Parse(src)
			require.NoError(t, err)

			astDoc, err := lower.Document(irDoc)
			require.NoError(t, err)

			got, err := codegen.HTML(astDoc)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
