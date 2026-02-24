package render_test

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/html"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/render"
	tk "github.com/spcameron/seanpatrickcameron.com/internal/markdown/testkit"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestRenderHTML(t *testing.T) {
	testCases := []struct {
		name     string
		astDoc   ast.Document
		htmlNode html.Node
		wantErr  error
	}{
		{
			name:   "paragraph with normal text",
			astDoc: tk.ASTDoc(tk.ASTPara(tk.ASTText("paragraph"))),
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
			name:   "header with normal text",
			astDoc: tk.ASTDoc(tk.ASTHeader(1, tk.ASTText("header"))),
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
			name:   "hard break renders",
			astDoc: tk.ASTDoc(tk.ASTPara(tk.ASTHardBreak())),
			htmlNode: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.VoidNode(
						"br",
						nil,
					),
				),
			),
			wantErr: nil,
		},
		{
			name: "soft break renders as whitespace (space)",
			astDoc: tk.ASTDoc(
				tk.ASTPara(
					tk.ASTText("a"),
					tk.ASTSoftBreak(),
					tk.ASTText("b"),
				),
			),
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
			got, err := render.HTML(tc.astDoc)

			assert.Equal(t, got, tc.htmlNode)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
