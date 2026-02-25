package render_test

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/block"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/build"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/html"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/render"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/require"
)

func TestRenderHTML(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		htmlNode html.Node
		wantErr  error
	}{
		{
			name:  "paragraph with normal text",
			input: "paragraph",
			htmlNode: html.FragmentNode(
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
			htmlNode: html.FragmentNode(
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
			htmlNode: html.FragmentNode(
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
			htmlNode: html.FragmentNode(
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
			htmlNode: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.TextNode("a b"),
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

			astDoc, err := build.AST(irDoc)
			require.NoError(t, err)

			got, err := render.HTML(astDoc)

			assert.Equal(t, got, tc.htmlNode)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
