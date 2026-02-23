package build_test

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/build"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	tk "github.com/spcameron/seanpatrickcameron.com/internal/markdown/testkit"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestBuildAST(t *testing.T) {
	testCases := []struct {
		name    string
		irDoc   ir.Document
		astDoc  ast.Document
		wantErr error
	}{
		{
			name:    "paragraph with normal text",
			irDoc:   tk.IRDoc(tk.IRPara("test text")),
			astDoc:  tk.ASTDoc(tk.ASTPara(tk.ASTText("test text"))),
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := build.AST(tc.irDoc)

			assert.Equal(t, got, tc.astDoc)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
