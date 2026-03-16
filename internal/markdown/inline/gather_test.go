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
				Brackets:     []BracketSummary{},
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
				Brackets:   []BracketSummary{},
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
				Brackets: []BracketSummary{},
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
				Brackets: []BracketSummary{},
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
				Brackets: []BracketSummary{},
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
				Brackets: []BracketSummary{},
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
				Brackets: []BracketSummary{},
			},
			wantErr: nil,
		},
		{
			name:  "open bracket only",
			input: "[",
			want: CursorSummary{
				WorkingItems: []WorkingItemSummary{
					{
						Kind:   "token",
						Lexeme: "[",
						Token:  "open_bracket",
					},
				},
				Delimiters: []DelimiterSummary{},
				Brackets:   []BracketSummary{},
			},
			wantErr: nil,
		},
		{
			name:  "close bracket only",
			input: "]",
			want: CursorSummary{
				WorkingItems: []WorkingItemSummary{
					{
						Kind:   "token",
						Lexeme: "]",
						Token:  "close_bracket",
					},
				},
				Delimiters: []DelimiterSummary{},
				Brackets: []BracketSummary{
					{
						Lexeme:    "]",
						ItemIndex: 0,
						Active:    true,
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "open paren only",
			input: "(",
			want: CursorSummary{
				WorkingItems: []WorkingItemSummary{
					{
						Kind:   "token",
						Lexeme: "(",
						Token:  "open_paren",
					},
				},
				Delimiters: []DelimiterSummary{},
				Brackets:   []BracketSummary{},
			},
			wantErr: nil,
		},
		{
			name:  "close paren only",
			input: ")",
			want: CursorSummary{
				WorkingItems: []WorkingItemSummary{
					{
						Kind:   "token",
						Lexeme: ")",
						Token:  "close_paren",
					},
				},
				Delimiters: []DelimiterSummary{},
				Brackets:   []BracketSummary{},
			},
			wantErr: nil,
		},
		{
			name:  "simple bracketed label",
			input: "[label]",
			want: CursorSummary{
				WorkingItems: []WorkingItemSummary{
					{
						Kind:   "token",
						Lexeme: "[",
						Token:  "open_bracket",
					},
					{
						Kind:   "text",
						Lexeme: "label",
					},
					{
						Kind:   "token",
						Lexeme: "]",
						Token:  "close_bracket",
					},
				},
				Delimiters: []DelimiterSummary{},
				Brackets: []BracketSummary{
					{
						Lexeme:    "]",
						ItemIndex: 2,
						Active:    true,
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
