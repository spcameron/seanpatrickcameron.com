package build_test

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/block"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/build"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	tk "github.com/spcameron/seanpatrickcameron.com/internal/markdown/testkit"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/require"
)

func TestBuildAST(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		astDoc  ast.Document
		wantErr error
	}{
		{
			name:    "paragraph with normal text",
			input:   "paragraph",
			astDoc:  tk.ASTDoc(tk.ASTPara(tk.ASTText())),
			wantErr: nil,
		},
		{
			name:    "header with normal text",
			input:   "# header",
			astDoc:  tk.ASTDoc(tk.ASTHeader(1, tk.ASTText())),
			wantErr: nil,
		},
		{
			name:  "header and paragraph",
			input: "# header\n\nparagraph",
			astDoc: tk.ASTDoc(
				tk.ASTHeader(1, tk.ASTText()),
				tk.ASTPara(tk.ASTText()),
			),
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := source.NewSource(tc.input)

			irDoc, err := block.Parse(src)
			require.NoError(t, err)

			got, err := build.AST(irDoc)

			got = tk.NormalizeAST(got)
			want := tk.NormalizeAST(tc.astDoc)

			assert.Equal(t, got, want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
