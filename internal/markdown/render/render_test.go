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
			astDoc: tk.ASTDoc(tk.ASTPara(tk.ASTText("test text"))),
			htmlNode: html.FragmentNode(
				html.ElemNode(
					"p",
					nil,
					html.TextNode("test text"),
				)),
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
