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

// TODO:
// generateHTMLCoverageGaps is a curated set of additional codegen-layer cases worth
// adding once expected html.Node trees are filled in. These are organized around the
// places where code generation itself makes decisions: tight-list paragraph unwrapping,
// attribute emission, raw/html passthrough, and preservation of inline/code payload.
var generateHTMLCoverageGaps = []struct {
	name  string
	input string
}{
	// Paragraph/codegen whitespace behavior.
	{
		name:  "paragraph: mixed soft and hard breaks across three lines",
		input: "alpha\nbeta  \ngamma",
	},
	{
		name:  "paragraph: emphasis around hard break",
		input: "*alpha*  \nbeta",
	},
	{
		name:  "paragraph: code span adjacent to soft break",
		input: "`alpha`\nbeta",
	},

	// Headers.
	{
		name:  "header: strong and emphasis",
		input: "# **alpha** *beta*",
	},
	{
		name:  "setext header: emphasis",
		input: "*alpha*\n---",
	},

	// Block quotes and nested containers.
	{
		name:  "block quote: two paragraphs",
		input: "> alpha\n>\n> beta",
	},
	{
		name:  "block quote: nested block quote",
		input: "> outer\n> > inner",
	},
	{
		name:  "block quote: contains list",
		input: "> - alpha\n> - beta",
	},

	// Lists: tight/loose rendering is a real codegen concern.
	{
		name:  "unordered list: loose list retains paragraph wrappers",
		input: "- alpha\n\n- beta",
	},
	{
		name:  "unordered list: tight list unwraps single paragraph children",
		input: "- alpha\n- beta",
	},
	{
		name:  "unordered list: item with paragraph and nested list",
		input: "- alpha\n  - beta",
	},
	{
		name:  "ordered list: non-1 start emits start attribute",
		input: "3. alpha\n4. beta",
	},
	{
		name:  "ordered list: paren delimiter still renders as ol",
		input: "1) alpha\n2) beta",
	},
	{
		name:  "list item: indented code block child",
		input: "- alpha\n\n      beta",
	},
	{
		name:  "list item: block quote child",
		input: "- alpha\n  > beta",
	},

	// Code blocks: codegen should faithfully wrap payload and language class.
	{
		name:  "indented code block: multiple lines",
		input: "    alpha\n    beta",
	},
	{
		name:  "indented code block: blank line in payload",
		input: "    alpha\n\n    beta",
	},
	{
		name:  "fenced code block: language class emitted",
		input: "```go\nalpha\n```",
	},
	{
		name:  "fenced code block: info string ignores trailing words in class emission",
		input: "```go linenos\nalpha\n```",
	},
	{
		name:  "fenced code block: payload preserves leading spaces after indent stripping",
		input: "```\n  alpha\n```",
	},

	// HTML blocks: codegen should emit raw payload as fragment children.
	{
		name:  "html block: multi-line comment",
		input: "<!--\nalpha\n-->",
	},
	{
		name:  "html block: named tag block",
		input: "<div>\nalpha\n</div>",
	},
	{
		name:  "html block: processing instruction",
		input: "<?php\necho $a;\n?>",
	},

	// Links and images: attribute emission matters here.
	{
		name:  "link with emphasis in label",
		input: "[*x*](dest)",
	},
	{
		name:  "link with title and nested strong label",
		input: "[**x**](dest \"title\")",
	},
	{
		name:  "image",
		input: "![alt](img.png)",
	},
	{
		name:  "image with title",
		input: "![alt](img.png \"title\")",
	},
	{
		name:  "autolink email and URI in one paragraph",
		input: "<local@domain.com> <https://google.com>",
	},

	// Mixed document cases.
	{
		name:  "document: header list code block paragraph",
		input: "# alpha\n\n- beta\n- gamma\n\n    delta\n\nepsilon",
	},
	{
		name:  "document: html block followed by paragraph",
		input: "<div>\nalpha\n</div>\n\nbeta",
	},
}

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
			name:  "ul: two items",
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
			name:  "ol: two items",
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
