package inline

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	tk "github.com/spcameron/seanpatrickcameron.com/internal/markdown/testkit"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []ast.Inline
		wantErr error
	}{
		{
			name:    "empty events yields empty inlines",
			input:   "",
			want:    []ast.Inline{},
			wantErr: nil,
		},
		{
			name:  "single rune yields one ast.Text",
			input: "a",
			want: []ast.Inline{
				tk.ASTText(),
			},
			wantErr: nil,
		},
		{
			name:  "plain sentence yields one ast.Text",
			input: "this is a test",
			want: []ast.Inline{
				tk.ASTText(),
			},
			wantErr: nil,
		},
		{
			name:  "unicode characters yields one ast.Text",
			input: "café 🎵 — 漢字",
			want: []ast.Inline{
				tk.ASTText(),
			},
			wantErr: nil,
		},
		{
			name:  "whitespace only yields one ast.Text",
			input: " \t ",
			want: []ast.Inline{
				tk.ASTText(),
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// src := source.NewSource(tc.input)
			// span := source.ByteSpan{
			// 	Start: 0,
			// 	End:   src.EOF(),
			// }
			//
			// got, err := Parse(src, span)
			//
			// got = tk.NormalizeASTInlines(got)
			// want := tk.NormalizeASTInlines(tc.want)
			//
			// assert.Equal(t, got, want)
			// assert.ErrorIs(t, err, tc.wantErr)
		})
	}

	spanCases := []struct {
		name    string
		input   string
		span    *source.ByteSpan
		want    []ast.Inline
		wantErr error
	}{
		{
			name:  "windowed span yields ast.Text",
			input: "prefix: body :suffix",
			span:  tk.SpanPtr(8, 12),
			want: []ast.Inline{
				tk.ASTTextAt(8, 12),
			},
			wantErr: nil,
		},
		{
			name:    "windowed empty span yields empty",
			input:   "hello",
			span:    tk.SpanPtr(0, 0),
			want:    []ast.Inline{},
			wantErr: nil,
		},
		{
			name:  "windowed span at beginning",
			input: "hello world",
			span:  tk.SpanPtr(0, 5),
			want: []ast.Inline{
				tk.ASTTextAt(0, 5),
			},
			wantErr: nil,
		},
		{
			name:  "windowed span at end",
			input: "hello world",
			span:  tk.SpanPtr(6, 11),
			want: []ast.Inline{
				tk.ASTTextAt(6, 11),
			},
			wantErr: nil,
		},
	}

	for _, tc := range spanCases {
		t.Run(tc.name, func(t *testing.T) {
			// src := source.NewSource(tc.input)
			//
			// span := source.ByteSpan{
			// 	Start: 0,
			// 	End:   src.EOF(),
			// }
			// if tc.span != nil {
			// 	span = *tc.span
			// }
			//
			// events, err := Scan(src, span)
			// require.NoError(t, err)
			//
			// got, err := Build(src, events)
			//
			// assert.Equal(t, got, tc.want)
			// assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
