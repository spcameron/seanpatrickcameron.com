package codegen_test

import (
	"strings"
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/block"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/codegen"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/html"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/lower"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	tk "github.com/spcameron/seanpatrickcameron.com/internal/markdown/testkit"
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
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLTextNode("paragraph"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "hard break renders (two spaces)",
			input: "a  \nb",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLTextNode("a"),
					tk.HTMLVoidNode("br", nil),
					tk.HTMLTextNode("b"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "hard break renders (backslash)",
			input: "a\\\nb",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLTextNode("a"),
					tk.HTMLVoidNode("br", nil),
					tk.HTMLTextNode("b"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "soft break renders as whitespace (space)",
			input: "a\nb",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLTextNode("a b"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: mixed soft and hard breaks across three lines",
			input: "alpha\nbeta  \ngamma",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLTextNode("alpha beta"),
					tk.HTMLVoidNode("br", nil),
					tk.HTMLTextNode("gamma"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: emphasis around hard break",
			input: "*alpha*  \nbeta",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLElementNode(
						"em",
						nil,
						tk.HTMLTextNode("alpha"),
					),
					tk.HTMLVoidNode("br", nil),
					tk.HTMLTextNode("beta"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: code span adjacent to soft break",
			input: "`alpha`\nbeta",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLElementNode(
						"code",
						nil,
						tk.HTMLTextNode("alpha"),
					),
					tk.HTMLTextNode(" beta"),
				),
			),
			wantErr: nil,
		},

		// Headings and simple block forms
		{
			name:  "header with normal text",
			input: "# header",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"h1",
					nil,
					tk.HTMLTextNode("header"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break",
			input: "---",
			want: tk.HTMLFragmentNode(
				tk.HTMLVoidNode("hr", nil),
			),
			wantErr: nil,
		},
		{
			name:  "html block",
			input: "<!-- comment -->",
			want: tk.HTMLFragmentNode(
				tk.HTMLFragmentNode(
					tk.HTMLRawNode("<!-- comment -->"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "header: strong and emphasis",
			input: "# **alpha** *beta*",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"h1",
					nil,
					tk.HTMLElementNode(
						"strong",
						nil,
						tk.HTMLTextNode("alpha"),
					),
					tk.HTMLTextNode(" "),
					tk.HTMLElementNode(
						"em",
						nil,
						tk.HTMLTextNode("beta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "setext header: emphasis",
			input: "*alpha*\n---",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"h2",
					nil,
					tk.HTMLElementNode(
						"em",
						nil,
						tk.HTMLTextNode("alpha"),
					),
				),
			),
			wantErr: nil,
		},

		// Containers
		{
			name:  "block quote: plain text",
			input: "> quote",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"blockquote",
					nil,
					tk.HTMLElementNode(
						"p",
						nil,
						tk.HTMLTextNode("quote"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: two paragraphs",
			input: "> alpha\n>\n> beta",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"blockquote",
					nil,
					tk.HTMLElementNode(
						"p",
						nil,
						tk.HTMLTextNode("alpha"),
					),
					tk.HTMLElementNode(
						"p",
						nil,
						tk.HTMLTextNode("beta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: nested block quote",
			input: "> outer\n> > inner",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"blockquote",
					nil,
					tk.HTMLElementNode(
						"p",
						nil,
						tk.HTMLTextNode("outer"),
					),
					tk.HTMLElementNode(
						"blockquote",
						nil,
						tk.HTMLElementNode(
							"p",
							nil,
							tk.HTMLTextNode("inner"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: contains list",
			input: "> - alpha\n> - beta",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"blockquote",
					nil,
					tk.HTMLElementNode(
						"ul",
						nil,
						tk.HTMLElementNode(
							"li",
							nil,
							tk.HTMLTextNode("alpha"),
						),
						tk.HTMLElementNode(
							"li",
							nil,
							tk.HTMLTextNode("beta"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "unordered list: two items",
			input: "- a\n- b",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"ul",
					nil,
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLTextNode("a"),
					),
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLTextNode("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "unordered list: loose list retains paragraph wrappers",
			input: "- alpha\n\n- beta",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"ul",
					nil,
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLElementNode(
							"p",
							nil,
							tk.HTMLTextNode("alpha"),
						),
					),
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLElementNode(
							"p",
							nil,
							tk.HTMLTextNode("beta"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "unordered list: tight list unwraps single paragraph children",
			input: "- alpha\n- beta",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"ul",
					nil,
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLTextNode("alpha"),
					),
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLTextNode("beta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "unordered list: item with paragraph and nested list",
			input: "- alpha\n  - beta",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"ul",
					nil,
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLTextNode("alpha"),
						tk.HTMLElementNode(
							"ul",
							nil,
							tk.HTMLElementNode(
								"li",
								nil,
								tk.HTMLTextNode("beta"),
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
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"ol",
					nil,
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLTextNode("a"),
					),
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLTextNode("b"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ordered list: non-1 start emits start attribute",
			input: "3. alpha\n4. beta",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"ol",
					html.Attributes{"start": "3"},
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLTextNode("alpha"),
					),
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLTextNode("beta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ordered list: paren delimiter still renders as ol",
			input: "1) alpha\n2) beta",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"ol",
					nil,
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLTextNode("alpha"),
					),
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLTextNode("beta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "list item: indented code block child",
			input: "- alpha\n\n      beta",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"ul",
					nil,
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLElementNode(
							"p",
							nil,
							tk.HTMLTextNode("alpha"),
						),
						tk.HTMLElementNode(
							"pre",
							nil,
							tk.HTMLElementNode(
								"code",
								nil,
								tk.HTMLTextNode("beta"),
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
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"ul",
					nil,
					tk.HTMLElementNode(
						"li",
						nil,
						tk.HTMLTextNode("alpha"),
						tk.HTMLElementNode(
							"blockquote",
							nil,
							tk.HTMLElementNode(
								"p",
								nil,
								tk.HTMLTextNode("beta"),
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
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"pre",
					nil,
					tk.HTMLElementNode(
						"code",
						nil,
						tk.HTMLTextNode(`fmt.Println("hello")`),
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
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"pre",
					nil,
					tk.HTMLElementNode(
						"code",
						nil,
						tk.HTMLTextNode(`fmt.Println("hello")`),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "indented code block: multiple lines",
			input: "    alpha\n    beta",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"pre",
					nil,
					tk.HTMLElementNode(
						"code",
						nil,
						tk.HTMLTextNode("alpha\nbeta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "indented code block: blank line in payload",
			input: "    alpha\n\n    beta",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"pre",
					nil,
					tk.HTMLElementNode(
						"code",
						nil,
						tk.HTMLTextNode("alpha\n\nbeta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "fenced code block: language class emitted",
			input: "```go\nalpha\n```",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"pre",
					nil,
					tk.HTMLElementNode(
						"code",
						html.Attributes{"class": "language-go"},
						tk.HTMLTextNode("alpha"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "fenced code block: info string ignores trailing words in class emission",
			input: "```go linenos\nalpha\n```",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"pre",
					nil,
					tk.HTMLElementNode(
						"code",
						html.Attributes{"class": "language-go"},
						tk.HTMLTextNode("alpha"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "fenced code block: payload preserves leading spaces after indent stripping",
			input: "```\n  alpha\n```",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"pre",
					nil,
					tk.HTMLElementNode(
						"code",
						nil,
						tk.HTMLTextNode("  alpha"),
					),
				),
			),
			wantErr: nil,
		},

		// HTML blocks
		{
			name:  "html block: multi-line comment",
			input: "<!--\nalpha\n-->",
			want: tk.HTMLFragmentNode(
				tk.HTMLFragmentNode(
					tk.HTMLRawNode("<!--"),
					tk.HTMLTextNode("\n"),
					tk.HTMLRawNode("alpha"),
					tk.HTMLTextNode("\n"),
					tk.HTMLRawNode("-->"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "html block: named tag block",
			input: "<div>\nalpha\n</div>",
			want: tk.HTMLFragmentNode(
				tk.HTMLFragmentNode(
					tk.HTMLRawNode("<div>"),
					tk.HTMLTextNode("\n"),
					tk.HTMLRawNode("alpha"),
					tk.HTMLTextNode("\n"),
					tk.HTMLRawNode("</div>"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "html block: processing instruction",
			input: "<?php\necho $a;\n?>",
			want: tk.HTMLFragmentNode(
				tk.HTMLFragmentNode(
					tk.HTMLRawNode("<?php"),
					tk.HTMLTextNode("\n"),
					tk.HTMLRawNode("echo $a;"),
					tk.HTMLTextNode("\n"),
					tk.HTMLRawNode("?>"),
				),
			),
			wantErr: nil,
		},

		// Inline rendering through paragraphs
		{
			name:  "code span",
			input: "`abc`",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLElementNode(
						"code",
						nil,
						tk.HTMLTextNode("abc"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "emphasis",
			input: "*abc*",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLElementNode(
						"em",
						nil,
						tk.HTMLTextNode("abc"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "strong",
			input: "**abc**",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLElementNode(
						"strong",
						nil,
						tk.HTMLTextNode("abc"),
					),
				),
			),
			wantErr: nil,
		},

		// Links and autolinks
		{
			name:  "simple link",
			input: `[x](dest)`,
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLElementNode(
						"a",
						html.Attributes{
							"href": "dest",
						},
						tk.HTMLTextNode("x"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "link with title",
			input: `[x](dest "title")`,
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLElementNode(
						"a",
						html.Attributes{
							"href":  "dest",
							"title": "title",
						},
						tk.HTMLTextNode("x"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "autolink URI",
			input: "<https://google.com>",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLElementNode(
						"a",
						html.Attributes{
							"href": "https://google.com",
						},
						tk.HTMLTextNode("https://google.com"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "autolink email",
			input: "<local@domain.com>",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLElementNode(
						"a",
						html.Attributes{
							"href": "mailto:local@domain.com",
						},
						tk.HTMLTextNode("local@domain.com"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "link with emphasis in label",
			input: "[*x*](dest)",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLElementNode(
						"a",
						html.Attributes{"href": "dest"},
						tk.HTMLElementNode(
							"em",
							nil,
							tk.HTMLTextNode("x"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "link with title and nested strong label",
			input: "[**x**](dest \"title\")",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLElementNode(
						"a",
						html.Attributes{
							"href":  "dest",
							"title": "title",
						},
						tk.HTMLElementNode(
							"strong",
							nil,
							tk.HTMLTextNode("x"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "image",
			input: "![alt](img.png)",
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLVoidNode(
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
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLVoidNode(
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
			want: tk.HTMLFragmentNode(
				tk.HTMLElementNode(
					"p",
					nil,
					tk.HTMLElementNode(
						"a",
						html.Attributes{"href": "mailto:local@domain.com"},
						tk.HTMLTextNode("local@domain.com"),
					),
					tk.HTMLTextNode(" "),
					tk.HTMLElementNode(
						"a",
						html.Attributes{"href": "https://google.com"},
						tk.HTMLTextNode("https://google.com"),
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
