package inline

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/require"
)

type GatherSummary struct {
	WorkingItems []WorkingItemSummary
	Delimiters   []DelimiterSummary
}

type WorkingItemSummary struct {
	Kind      string
	Lexeme    string
	Delimiter byte
}

type DelimiterSummary struct {
	Lexeme       string
	Delimiter    byte
	OriginalRun  int
	RemainingRun int
	CanOpen      bool
	CanClose     bool
	ItemIndex    int
}

func TestGather(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    GatherSummary
		wantErr error
	}{
		{
			name:  "empty input",
			input: "",
			want: GatherSummary{
				WorkingItems: []WorkingItemSummary{},
				Delimiters:   []DelimiterSummary{},
			},
			wantErr: nil,
		},
		{
			name:  "plain text",
			input: "abc",
			want: GatherSummary{
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
			want: GatherSummary{
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
			want: GatherSummary{
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
			want: GatherSummary{
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
			want: GatherSummary{
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
			want: GatherSummary{
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

			summary := summarizeGather(cursor)

			assert.Equal(t, summary, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func summarizeWorkingItems(src *source.Source, items []WorkingItem) []WorkingItemSummary {
	summary := make([]WorkingItemSummary, 0, len(items))

	for _, item := range items {
		switch v := item.(type) {
		case *TextItem:
			s := src.Slice(v.Span)

			summary = append(summary, WorkingItemSummary{
				Kind:   "text",
				Lexeme: s,
			})

		case *DelimiterItem:
			s := src.Slice(v.Span)

			summary = append(summary, WorkingItemSummary{
				Kind:      "delimiter",
				Lexeme:    s,
				Delimiter: v.Delimiter,
			})

		case *NodeItem:
			panic("node item encountered during gather")

		default:
			panic("unknown item type")
		}
	}

	return summary
}

func summarizeDelimiters(src *source.Source, delims []*DelimiterRecord) []DelimiterSummary {
	summary := make([]DelimiterSummary, 0, len(delims))

	for _, delim := range delims {
		s := src.Slice(delim.Span)

		summary = append(summary, DelimiterSummary{
			Lexeme:       s,
			Delimiter:    delim.Delimiter,
			OriginalRun:  delim.OriginalRun,
			RemainingRun: delim.RemainingRun,
			CanOpen:      delim.CanOpen,
			CanClose:     delim.CanClose,
			ItemIndex:    delim.ItemIndex,
		})
	}

	return summary
}

func summarizeGather(c *Cursor) GatherSummary {
	summary := GatherSummary{
		WorkingItems: summarizeWorkingItems(c.Source, c.WorkingItems),
		Delimiters:   summarizeDelimiters(c.Source, c.DelimiterRecords),
	}

	return summary
}
