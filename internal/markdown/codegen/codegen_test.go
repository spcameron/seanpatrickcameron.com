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
