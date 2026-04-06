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
			name:  "thematic break",
			input: "---",
			want: html.FragmentNode(
				html.VoidNode("hr", nil),
			),
			wantErr: nil,
		},
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
