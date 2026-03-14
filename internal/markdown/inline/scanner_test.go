package inline

import (
	"fmt"
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

type EventSummary struct {
	Kind      EventKind
	Lexeme    string
	Delimiter byte
	RunLength int
}

func (es EventSummary) String() string {
	if es.Kind == EventDelimiterRun {
		return fmt.Sprintf("[%s] - Lexeme (%q), Delimiter (%s), Length (%d)", es.Kind, es.Lexeme, string(es.Delimiter), es.RunLength)
	}

	return fmt.Sprintf("[%s] - Lexeme (%q)", es.Kind, es.Lexeme)
}

func TestScan(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []EventSummary
		wantErr error
	}{
		{
			name:    "empty input",
			input:   "",
			want:    []EventSummary{},
			wantErr: nil,
		},
		{
			name:  "plain text",
			input: "hello",
			want: []EventSummary{
				{
					Kind:   EventText,
					Lexeme: "hello",
				},
			},
			wantErr: nil,
		},
		{
			name:  "single star delimiter",
			input: "*",
			want: []EventSummary{
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "*",
					Delimiter: '*',
					RunLength: 1,
				},
			},
			wantErr: nil,
		},
		{
			name:  "double star delimiter",
			input: "**",
			want: []EventSummary{
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "**",
					Delimiter: '*',
					RunLength: 2,
				},
			},
			wantErr: nil,
		},
		{
			name:  "triple star delimiter",
			input: "***",
			want: []EventSummary{
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "***",
					Delimiter: '*',
					RunLength: 3,
				},
			},
			wantErr: nil,
		},
		{
			name:  "text then delimiter",
			input: "abc*",
			want: []EventSummary{
				{
					Kind:   EventText,
					Lexeme: "abc",
				},
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "*",
					Delimiter: '*',
					RunLength: 1,
				},
			},
			wantErr: nil,
		},
		{
			name:  "delimiter then text",
			input: "*abc",
			want: []EventSummary{
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "*",
					Delimiter: '*',
					RunLength: 1,
				},
				{
					Kind:   EventText,
					Lexeme: "abc",
				},
			},
			wantErr: nil,
		},
		{
			name:  "text delimiter text",
			input: "a*b",
			want: []EventSummary{
				{
					Kind:   EventText,
					Lexeme: "a",
				},
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "*",
					Delimiter: '*',
					RunLength: 1,
				},
				{
					Kind:   EventText,
					Lexeme: "b",
				},
			},
			wantErr: nil,
		},
		{
			name:  "text double delimiter text",
			input: "a**b",
			want: []EventSummary{
				{
					Kind:   EventText,
					Lexeme: "a",
				},
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "**",
					Delimiter: '*',
					RunLength: 2,
				},
				{
					Kind:   EventText,
					Lexeme: "b",
				},
			},
			wantErr: nil,
		},
		{
			name:  "emphasis-shaped input",
			input: "*abc*",
			want: []EventSummary{
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "*",
					Delimiter: '*',
					RunLength: 1,
				},
				{
					Kind:   EventText,
					Lexeme: "abc",
				},
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "*",
					Delimiter: '*',
					RunLength: 1,
				},
			},
			wantErr: nil,
		},
		{
			name:  "strong-shaped input",
			input: "**abc**",
			want: []EventSummary{
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "**",
					Delimiter: '*',
					RunLength: 2,
				},
				{
					Kind:   EventText,
					Lexeme: "abc",
				},
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "**",
					Delimiter: '*',
					RunLength: 2,
				},
			},
			wantErr: nil,
		},
		{
			name:  "triple-star wrapped input",
			input: "***abc***",
			want: []EventSummary{
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "***",
					Delimiter: '*',
					RunLength: 3,
				},
				{
					Kind:   EventText,
					Lexeme: "abc",
				},
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "***",
					Delimiter: '*',
					RunLength: 3,
				},
			},
			wantErr: nil,
		},
		{
			name:  "multiple delimiter runs separated by text",
			input: "*a**b***c",
			want: []EventSummary{
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "*",
					Delimiter: '*',
					RunLength: 1,
				},
				{
					Kind:   EventText,
					Lexeme: "a",
				},
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "**",
					Delimiter: '*',
					RunLength: 2,
				},
				{
					Kind:   EventText,
					Lexeme: "b",
				},
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "***",
					Delimiter: '*',
					RunLength: 3,
				},
				{
					Kind:   EventText,
					Lexeme: "c",
				},
			},
			wantErr: nil,
		},
		{
			name:  "space around delimiter",
			input: "a * b",
			want: []EventSummary{
				{
					Kind:   EventText,
					Lexeme: "a ",
				},
				{
					Kind:      EventDelimiterRun,
					Lexeme:    "*",
					Delimiter: '*',
					RunLength: 1,
				},
				{
					Kind:   EventText,
					Lexeme: " b",
				},
			},
			wantErr: nil,
		},
		{
			name:  "newline",
			input: "\n",
			want: []EventSummary{
				{
					Kind:   EventIllegalNewline,
					Lexeme: "\n",
				},
			},
			wantErr: nil,
		},
		{
			name:  "text then newline",
			input: "abc\n",
			want: []EventSummary{
				{
					Kind:   EventText,
					Lexeme: "abc",
				},
				{
					Kind:   EventIllegalNewline,
					Lexeme: "\n",
				},
			},
			wantErr: nil,
		},
		{
			name:  "newline then text",
			input: "\nabc",
			want: []EventSummary{
				{
					Kind:   EventIllegalNewline,
					Lexeme: "\n",
				},
				{
					Kind:   EventText,
					Lexeme: "abc",
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
			got := summarizeEvents(src, events)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func summarizeEvents(src *source.Source, events []Event) []EventSummary {
	es := make([]EventSummary, 0, len(events))

	for _, e := range events {
		s := src.Slice(e.Span)

		es = append(es, EventSummary{
			Kind:      e.Kind,
			Lexeme:    s,
			Delimiter: e.Delimiter,
			RunLength: e.RunLength,
		})
	}

	return es
}
