package inline

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/require"
)

func TestGather(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    CursorSummary
		wantErr error
	}{
		{
			name:  "empty input",
			input: "",
			want: CursorSummary{
				WorkingItems: []WorkingItemSummary{},
				Delimiters:   []DelimiterSummary{},
			},
			wantErr: nil,
		},
		{
			name:  "plain text",
			input: "abc",
			want: CursorSummary{
				WorkingItems: []WorkingItemSummary{
					{
						Kind:   "text",
						Lexeme: "abc",
					},
				},
				Delimiters: []DelimiterSummary{},
			},
			wantErr: nil,
		},
		{
			name:  "opener only",
			input: "*abc",
			want: CursorSummary{
				WorkingItems: []WorkingItemSummary{
					{
						Kind:      "delimiter",
						Lexeme:    "*",
						Delimiter: '*',
					},
					{
						Kind:   "text",
						Lexeme: "abc",
					},
				},
				Delimiters: []DelimiterSummary{
					{
						Lexeme:       "*",
						Delimiter:    '*',
						OriginalRun:  1,
						RemainingRun: 1,
						CanOpen:      true,
						CanClose:     false,
						ItemIndex:    0,
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "closer only",
			input: "abc*",
			want: CursorSummary{
				WorkingItems: []WorkingItemSummary{
					{
						Kind:   "text",
						Lexeme: "abc",
					},
					{
						Kind:      "delimiter",
						Lexeme:    "*",
						Delimiter: '*',
					},
				},
				Delimiters: []DelimiterSummary{
					{
						Lexeme:       "*",
						Delimiter:    '*',
						OriginalRun:  1,
						RemainingRun: 1,
						CanOpen:      false,
						CanClose:     true,
						ItemIndex:    1,
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "opener and closer",
			input: "a*b",
			want: CursorSummary{
				WorkingItems: []WorkingItemSummary{
					{
						Kind:   "text",
						Lexeme: "a",
					},
					{
						Kind:      "delimiter",
						Lexeme:    "*",
						Delimiter: '*',
					},
					{
						Kind:   "text",
						Lexeme: "b",
					},
				},
				Delimiters: []DelimiterSummary{
					{
						Lexeme:       "*",
						Delimiter:    '*',
						OriginalRun:  1,
						RemainingRun: 1,
						CanOpen:      true,
						CanClose:     true,
						ItemIndex:    1,
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "neither opener nor closer",
			input: "a * b",
			want: CursorSummary{
				WorkingItems: []WorkingItemSummary{
					{
						Kind:   "text",
						Lexeme: "a ",
					},
					{
						Kind:      "delimiter",
						Lexeme:    "*",
						Delimiter: '*',
					},
					{
						Kind:   "text",
						Lexeme: " b",
					},
				},
				Delimiters: []DelimiterSummary{
					{
						Lexeme:       "*",
						Delimiter:    '*',
						OriginalRun:  1,
						RemainingRun: 1,
						CanOpen:      false,
						CanClose:     false,
						ItemIndex:    1,
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "triple star delimiter",
			input: "***abc***",
			want: CursorSummary{
				WorkingItems: []WorkingItemSummary{
					{
						Kind:      "delimiter",
						Lexeme:    "***",
						Delimiter: '*',
					},
					{
						Kind:   "text",
						Lexeme: "abc",
					},
					{
						Kind:      "delimiter",
						Lexeme:    "***",
						Delimiter: '*',
					},
				},
				Delimiters: []DelimiterSummary{
					{
						Lexeme:       "***",
						Delimiter:    '*',
						OriginalRun:  3,
						RemainingRun: 3,
						CanOpen:      true,
						CanClose:     false,
						ItemIndex:    0,
					},
					{
						Lexeme:       "***",
						Delimiter:    '*',
						OriginalRun:  3,
						RemainingRun: 3,
						CanOpen:      false,
						CanClose:     true,
						ItemIndex:    2,
					},
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

			events, err := Scan(src, span)
			require.NoError(t, err)

			cursor := NewCursor(src, span, events)

			err = cursor.Gather()

			summary := summarizeCursor(cursor)

			assert.Equal(t, summary, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
