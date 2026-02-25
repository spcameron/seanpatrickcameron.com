package inline

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	tk "github.com/spcameron/seanpatrickcameron.com/internal/markdown/testkit"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/require"
)

func TestScan(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		span    *source.ByteSpan // nil == whole input
		want    []Event
		wantErr error
	}{
		{
			name:    "empty span yields no events",
			input:   "",
			span:    nil,
			want:    []Event{},
			wantErr: nil,
		},
		{
			name:  "non-empty span yield one EventText covering span",
			input: "hello",
			span:  nil,
			want: []Event{
				{
					Kind: EventText,
					Span: source.ByteSpan{Start: 0, End: 5},
				},
			},
			wantErr: nil,
		},
		{
			name:  "windowed span yields one EventText covering window",
			input: "hello",
			span:  tk.SpanPtr(1, 3),
			want: []Event{
				{
					Kind: EventText,
					Span: source.ByteSpan{Start: 1, End: 3},
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := source.NewSource(tc.input)

			span := source.ByteSpan{
				Start: 0,
				End:   src.EOF(),
			}
			if tc.span != nil {
				span = *tc.span
			}

			got, err := Scan(src, span)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestBuild(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		span    *source.ByteSpan // nil == whole input
		want    []ast.Inline
		wantErr error
	}{
		{
			name:    "empty events yields empty inlines",
			input:   "",
			span:    nil,
			want:    []ast.Inline{},
			wantErr: nil,
		},
		{
			name:  "single rune yields one ast.Text",
			input: "a",
			span:  nil,
			want: []ast.Inline{
				tk.ASTText(),
			},
			wantErr: nil,
		},
		{
			name:  "plain sentence yields one ast.Text",
			input: "this is a test",
			span:  nil,
			want: []ast.Inline{
				tk.ASTText(),
			},
			wantErr: nil,
		},
		{
			name:  "unicode characters yields one ast.Text",
			input: "café 🎵 — 漢字",
			span:  nil,
			want: []ast.Inline{
				tk.ASTText(),
			},
			wantErr: nil,
		},
		{
			name:  "whitespace only yields one ast.Text",
			input: " \t ",
			span:  nil,
			want: []ast.Inline{
				tk.ASTText(),
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := source.NewSource(tc.input)

			span := source.ByteSpan{
				Start: 0,
				End:   src.EOF(),
			}
			if tc.span != nil {
				span = *tc.span
			}

			events, err := Scan(src, span)
			require.NoError(t, err)

			got, err := Build(src, events)

			got = tk.NormalizeASTInlines(got)
			want := tk.NormalizeASTInlines(tc.want)

			assert.Equal(t, got, want)
			assert.ErrorIs(t, err, tc.wantErr)
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
			src := source.NewSource(tc.input)

			span := source.ByteSpan{
				Start: 0,
				End:   src.EOF(),
			}
			if tc.span != nil {
				span = *tc.span
			}

			events, err := Scan(src, span)
			require.NoError(t, err)

			got, err := Build(src, events)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func normalizeEvents(events []Event) []Event {
	out := make([]Event, 0, len(events))
	for i := range events {
		ev := events[i]
		ev.Span = source.ByteSpan{}
		out = append(out, ev)
	}

	return out
}
