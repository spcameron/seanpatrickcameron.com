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
			name:  "empty span yields no events",
			input: "",
			span: &source.ByteSpan{
				Start: 0,
				End:   0,
			},
			want:    []Event{},
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

			got = normalizeEvents(got)
			want := normalizeEvents(tc.want)

			assert.Equal(t, got, want)
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
	}{}

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
