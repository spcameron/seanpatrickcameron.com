package inline

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestScan(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []TokenSummary
		wantErr error
	}{
		{
			name:  "empty input",
			input: "",
			want: []TokenSummary{
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "plain text",
			input: "abc",
			want: []TokenSummary{
				{
					Kind:   TokenText,
					Lexeme: "abc",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "plain text, punctuation preserved",
			input: " $.;'",
			want: []TokenSummary{
				{
					Kind:   TokenText,
					Lexeme: " $.;'",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "plain text, spaces preserved",
			input: "a   b   c",
			want: []TokenSummary{
				{
					Kind:   TokenText,
					Lexeme: "a   b   c",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "star delimiter",
			input: "*",
			want: []TokenSummary{
				{
					Kind:   TokenStarDelimiter,
					Lexeme: "*",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "underscore delimiter",
			input: "_",
			want: []TokenSummary{
				{
					Kind:   TokenUnderscoreDelimiter,
					Lexeme: "_",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "backtick",
			input: "`",
			want: []TokenSummary{
				{
					Kind:   TokenBacktick,
					Lexeme: "`",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "open bracket",
			input: "[",
			want: []TokenSummary{
				{
					Kind:   TokenOpenBracket,
					Lexeme: "[",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "close bracket",
			input: "]",
			want: []TokenSummary{
				{
					Kind:   TokenCloseBracket,
					Lexeme: "]",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "open paren",
			input: "(",
			want: []TokenSummary{
				{
					Kind:   TokenOpenParen,
					Lexeme: "(",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "close paren",
			input: ")",
			want: []TokenSummary{
				{
					Kind:   TokenCloseParen,
					Lexeme: ")",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "open angle bracket",
			input: "<",
			want: []TokenSummary{
				{
					Kind:   TokenOpenAngle,
					Lexeme: "<",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "close angle bracket",
			input: ">",
			want: []TokenSummary{
				{
					Kind:   TokenCloseAngle,
					Lexeme: ">",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "bang",
			input: "!",
			want: []TokenSummary{
				{
					Kind:   TokenBang,
					Lexeme: "!",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "image open bracket",
			input: "![",
			want: []TokenSummary{
				{
					Kind:   TokenImageOpenBracket,
					Lexeme: "![",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "backtick run",
			input: "```",
			want: []TokenSummary{
				{
					Kind:   TokenBacktick,
					Lexeme: "```",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "star delimiter run",
			input: "***",
			want: []TokenSummary{
				{
					Kind:   TokenStarDelimiter,
					Lexeme: "***",
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},
		{
			name:  "underscore delimiter run",
			input: "__",
			want: []TokenSummary{
				{
					Kind:   TokenUnderscoreDelimiter,
					Lexeme: "__",
				},
				{
					Kind: TokenEOF,
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

			tokens, err := Scan(src, span)
			got := summarizeTokens(src, tokens)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

// func TestScan(t *testing.T) {
// 	testCases := []struct {
// 		name    string
// 		input   string
// 		want    []EventSummary
// 		wantErr error
// 	}{
// 		{
// 			name:    "empty input",
// 			input:   "",
// 			want:    []EventSummary{},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "plain text",
// 			input: "hello",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventText,
// 					Lexeme: "hello",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "single star delimiter",
// 			input: "*",
// 			want: []EventSummary{
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "*",
// 					Delimiter: '*',
// 					RunLength: 1,
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "double star delimiter",
// 			input: "**",
// 			want: []EventSummary{
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "**",
// 					Delimiter: '*',
// 					RunLength: 2,
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "triple star delimiter",
// 			input: "***",
// 			want: []EventSummary{
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "***",
// 					Delimiter: '*',
// 					RunLength: 3,
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "text then delimiter",
// 			input: "abc*",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventText,
// 					Lexeme: "abc",
// 				},
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "*",
// 					Delimiter: '*',
// 					RunLength: 1,
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "delimiter then text",
// 			input: "*abc",
// 			want: []EventSummary{
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "*",
// 					Delimiter: '*',
// 					RunLength: 1,
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "abc",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "text delimiter text",
// 			input: "a*b",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventText,
// 					Lexeme: "a",
// 				},
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "*",
// 					Delimiter: '*',
// 					RunLength: 1,
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "b",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "text double delimiter text",
// 			input: "a**b",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventText,
// 					Lexeme: "a",
// 				},
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "**",
// 					Delimiter: '*',
// 					RunLength: 2,
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "b",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "emphasis-shaped input",
// 			input: "*abc*",
// 			want: []EventSummary{
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "*",
// 					Delimiter: '*',
// 					RunLength: 1,
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "abc",
// 				},
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "*",
// 					Delimiter: '*',
// 					RunLength: 1,
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "strong-shaped input",
// 			input: "**abc**",
// 			want: []EventSummary{
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "**",
// 					Delimiter: '*',
// 					RunLength: 2,
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "abc",
// 				},
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "**",
// 					Delimiter: '*',
// 					RunLength: 2,
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "triple-star wrapped input",
// 			input: "***abc***",
// 			want: []EventSummary{
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "***",
// 					Delimiter: '*',
// 					RunLength: 3,
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "abc",
// 				},
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "***",
// 					Delimiter: '*',
// 					RunLength: 3,
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "multiple delimiter runs separated by text",
// 			input: "*a**b***c",
// 			want: []EventSummary{
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "*",
// 					Delimiter: '*',
// 					RunLength: 1,
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "a",
// 				},
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "**",
// 					Delimiter: '*',
// 					RunLength: 2,
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "b",
// 				},
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "***",
// 					Delimiter: '*',
// 					RunLength: 3,
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "c",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "space around delimiter",
// 			input: "a * b",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventText,
// 					Lexeme: "a ",
// 				},
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "*",
// 					Delimiter: '*',
// 					RunLength: 1,
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: " b",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "newline",
// 			input: "\n",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventIllegalNewline,
// 					Lexeme: "\n",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "text then newline",
// 			input: "abc\n",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventText,
// 					Lexeme: "abc",
// 				},
// 				{
// 					Kind:   EventIllegalNewline,
// 					Lexeme: "\n",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "newline then text",
// 			input: "\nabc",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventIllegalNewline,
// 					Lexeme: "\n",
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "abc",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "open bracket",
// 			input: "[",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventOpenBracket,
// 					Lexeme: "[",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "close bracket",
// 			input: "]",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventCloseBracket,
// 					Lexeme: "]",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "open paren",
// 			input: "(",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventOpenParen,
// 					Lexeme: "(",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "close paren",
// 			input: ")",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventCloseParen,
// 					Lexeme: ")",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "adjacent brackets",
// 			input: "[]",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventOpenBracket,
// 					Lexeme: "[",
// 				},
// 				{
// 					Kind:   EventCloseBracket,
// 					Lexeme: "]",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "adjacent parens",
// 			input: "()",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventOpenParen,
// 					Lexeme: "(",
// 				},
// 				{
// 					Kind:   EventCloseParen,
// 					Lexeme: ")",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "text open bracket text",
// 			input: "a[b",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventText,
// 					Lexeme: "a",
// 				},
// 				{
// 					Kind:   EventOpenBracket,
// 					Lexeme: "[",
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "b",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "text close bracket text",
// 			input: "a]b",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventText,
// 					Lexeme: "a",
// 				},
// 				{
// 					Kind:   EventCloseBracket,
// 					Lexeme: "]",
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "b",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "text open paren text",
// 			input: "a(b",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventText,
// 					Lexeme: "a",
// 				},
// 				{
// 					Kind:   EventOpenParen,
// 					Lexeme: "(",
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "b",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "text close paren text",
// 			input: "a)b",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventText,
// 					Lexeme: "a",
// 				},
// 				{
// 					Kind:   EventCloseParen,
// 					Lexeme: ")",
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "b",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "simple bracketed label",
// 			input: "[abc]",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventOpenBracket,
// 					Lexeme: "[",
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "abc",
// 				},
// 				{
// 					Kind:   EventCloseBracket,
// 					Lexeme: "]",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "simple paren group",
// 			input: "(abc)",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventOpenParen,
// 					Lexeme: "(",
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "abc",
// 				},
// 				{
// 					Kind:   EventCloseParen,
// 					Lexeme: ")",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "simple inline link skeleton",
// 			input: "[label](dest)",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventOpenBracket,
// 					Lexeme: "[",
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "label",
// 				},
// 				{
// 					Kind:   EventCloseBracket,
// 					Lexeme: "]",
// 				},
// 				{
// 					Kind:   EventOpenParen,
// 					Lexeme: "(",
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "dest",
// 				},
// 				{
// 					Kind:   EventCloseParen,
// 					Lexeme: ")",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "inline link with surrounding text",
// 			input: "go to [home](index)",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventText,
// 					Lexeme: "go to ",
// 				},
// 				{
// 					Kind:   EventOpenBracket,
// 					Lexeme: "[",
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "home",
// 				},
// 				{
// 					Kind:   EventCloseBracket,
// 					Lexeme: "]",
// 				},
// 				{
// 					Kind:   EventOpenParen,
// 					Lexeme: "(",
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "index",
// 				},
// 				{
// 					Kind:   EventCloseParen,
// 					Lexeme: ")",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "emphasis inside label",
// 			input: "[a *b* c](url)",
// 			want: []EventSummary{
// 				{
// 					Kind:   EventOpenBracket,
// 					Lexeme: "[",
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "a ",
// 				},
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "*",
// 					Delimiter: '*',
// 					RunLength: 1,
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "b",
// 				},
// 				{
// 					Kind:      EventDelimiterRun,
// 					Lexeme:    "*",
// 					Delimiter: '*',
// 					RunLength: 1,
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: " c",
// 				},
// 				{
// 					Kind:   EventCloseBracket,
// 					Lexeme: "]",
// 				},
// 				{
// 					Kind:   EventOpenParen,
// 					Lexeme: "(",
// 				},
// 				{
// 					Kind:   EventText,
// 					Lexeme: "url",
// 				},
// 				{
// 					Kind:   EventCloseParen,
// 					Lexeme: ")",
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 	}
//
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			src := source.NewSource(tc.input)
// 			span := source.ByteSpan{
// 				Start: 0,
// 				End:   src.EOF(),
// 			}
//
// 			events, err := Scan(src, span)
// 			got := summarizeEvents(src, events)
//
// 			assert.Equal(t, got, tc.want)
// 			assert.ErrorIs(t, err, tc.wantErr)
// 		})
// 	}
// }
