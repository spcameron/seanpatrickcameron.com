package lower_test

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/block"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/lower"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	tk "github.com/spcameron/seanpatrickcameron.com/internal/markdown/testkit"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/require"
)

func TestLowerDocument(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    ast.Document
		wantErr error
	}{
		{
			name:  "paragraph with normal text",
			input: "paragraph",
			want: tk.ASTDoc(
				tk.ASTPara(tk.ASTText()),
			),
			wantErr: nil,
		},
		{
			name:  "header with normal text",
			input: "# header",
			want: tk.ASTDoc(
				tk.ASTHeader(1, tk.ASTText()),
			),
			wantErr: nil,
		},
		{
			name:  "header and paragraph",
			input: "# header\n\nparagraph",
			want: tk.ASTDoc(
				tk.ASTHeader(1, tk.ASTText()),
				tk.ASTPara(tk.ASTText()),
			),
			wantErr: nil,
		},
		{
			name:  "thematic break",
			input: "---",
			want: tk.ASTDoc(
				tk.ASTThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: plain text",
			input: "> quote",
			want: tk.ASTDoc(
				tk.ASTBlockQuote(
					tk.ASTPara(tk.ASTText()),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ul: two items",
			input: "- a\n- b",
			want: tk.ASTDoc(
				tk.ASTUnorderedList(
					true,
					tk.ASTListItem(
						tk.ASTPara(tk.ASTText()),
					),
					tk.ASTListItem(
						tk.ASTPara((tk.ASTText())),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ol: two items",
			input: "1. a\n2. b",
			want: tk.ASTDoc(
				tk.ASTOrderedList(
					true,
					1,
					tk.ASTListItem(
						tk.ASTPara(tk.ASTText()),
					),
					tk.ASTListItem(
						tk.ASTPara(tk.ASTText()),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "indented code block",
			input: "    code",
			want: tk.ASTDoc(
				tk.ASTIndentedCodeBlock(
					tk.ASTText(),
				),
			),
			wantErr: nil,
		},
		{
			name:  "fenced code block",
			input: "```\ncode\n```",
			want: tk.ASTDoc(
				tk.ASTFencedCodeBlock(
					tk.ASTText(),
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

			got, err := lower.Document(irDoc)

			got = tk.NormalizeAST(got)
			want := tk.NormalizeAST(tc.want)

			assert.Equal(t, got, want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
