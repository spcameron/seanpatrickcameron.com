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
		span    source.ByteSpan
		want    []TokenSummary
		wantErr error
	}{
		// Plain text
		{
			name:  "empty input",
			input: "",
			span:  source.ByteSpan{Start: 0, End: 0},
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
			span:  source.ByteSpan{Start: 0, End: 3},
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
			span:  source.ByteSpan{Start: 0, End: 5},
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
			span:  source.ByteSpan{Start: 0, End: 9},
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

		// Single-token forms
		{
			name:  "star delimiter",
			input: "*",
			span:  source.ByteSpan{Start: 0, End: 1},
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
			span:  source.ByteSpan{Start: 0, End: 1},
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
			span:  source.ByteSpan{Start: 0, End: 1},
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
			span:  source.ByteSpan{Start: 0, End: 1},
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
			span:  source.ByteSpan{Start: 0, End: 1},
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
			span:  source.ByteSpan{Start: 0, End: 1},
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
			span:  source.ByteSpan{Start: 0, End: 1},
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
			span:  source.ByteSpan{Start: 0, End: 1},
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
			span:  source.ByteSpan{Start: 0, End: 1},
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
			span:  source.ByteSpan{Start: 0, End: 1},
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
			span:  source.ByteSpan{Start: 0, End: 2},
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
			name:  "backslash",
			input: `\`,
			span:  source.ByteSpan{Start: 0, End: 1},
			want: []TokenSummary{
				{
					Kind:   TokenBackslash,
					Lexeme: `\`,
				},
				{
					Kind: TokenEOF,
				},
			},
			wantErr: nil,
		},

		// Run tokens
		{
			name:  "backtick run",
			input: "```",
			span:  source.ByteSpan{Start: 0, End: 3},
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
			span:  source.ByteSpan{Start: 0, End: 3},
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
			span:  source.ByteSpan{Start: 0, End: 2},
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

		// Mixed token sequences
		{
			name:  "text then star then text",
			input: "a*b",
			span:  source.ByteSpan{Start: 0, End: 3},
			want: []TokenSummary{
				{Kind: TokenText, Lexeme: "a"},
				{Kind: TokenStarDelimiter, Lexeme: "*"},
				{Kind: TokenText, Lexeme: "b"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "text then underscore then text",
			input: "a_b",
			span:  source.ByteSpan{Start: 0, End: 3},
			want: []TokenSummary{
				{Kind: TokenText, Lexeme: "a"},
				{Kind: TokenUnderscoreDelimiter, Lexeme: "_"},
				{Kind: TokenText, Lexeme: "b"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "text then backtick then text",
			input: "a`b",
			span:  source.ByteSpan{Start: 0, End: 3},
			want: []TokenSummary{
				{Kind: TokenText, Lexeme: "a"},
				{Kind: TokenBacktick, Lexeme: "`"},
				{Kind: TokenText, Lexeme: "b"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "brackets inside text",
			input: "a[b]c",
			span:  source.ByteSpan{Start: 0, End: 5},
			want: []TokenSummary{
				{Kind: TokenText, Lexeme: "a"},
				{Kind: TokenOpenBracket, Lexeme: "["},
				{Kind: TokenText, Lexeme: "b"},
				{Kind: TokenCloseBracket, Lexeme: "]"},
				{Kind: TokenText, Lexeme: "c"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "parens inside text",
			input: "a(b)c",
			span:  source.ByteSpan{Start: 0, End: 5},
			want: []TokenSummary{
				{Kind: TokenText, Lexeme: "a"},
				{Kind: TokenOpenParen, Lexeme: "("},
				{Kind: TokenText, Lexeme: "b"},
				{Kind: TokenCloseParen, Lexeme: ")"},
				{Kind: TokenText, Lexeme: "c"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "angle brackets inside text",
			input: "a<b>c",
			span:  source.ByteSpan{Start: 0, End: 5},
			want: []TokenSummary{
				{Kind: TokenText, Lexeme: "a"},
				{Kind: TokenOpenAngle, Lexeme: "<"},
				{Kind: TokenText, Lexeme: "b"},
				{Kind: TokenCloseAngle, Lexeme: ">"},
				{Kind: TokenText, Lexeme: "c"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "bang followed by text",
			input: "!a",
			span:  source.ByteSpan{Start: 0, End: 2},
			want: []TokenSummary{
				{Kind: TokenBang, Lexeme: "!"},
				{Kind: TokenText, Lexeme: "a"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "image opener followed by label text and close bracket",
			input: "![x]",
			span:  source.ByteSpan{Start: 0, End: 4},
			want: []TokenSummary{
				{Kind: TokenImageOpenBracket, Lexeme: "!["},
				{Kind: TokenText, Lexeme: "x"},
				{Kind: TokenCloseBracket, Lexeme: "]"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "plain text around multiple token kinds",
			input: "ab[c](d)!",
			span:  source.ByteSpan{Start: 0, End: 9},
			want: []TokenSummary{
				{Kind: TokenText, Lexeme: "ab"},
				{Kind: TokenOpenBracket, Lexeme: "["},
				{Kind: TokenText, Lexeme: "c"},
				{Kind: TokenCloseBracket, Lexeme: "]"},
				{Kind: TokenOpenParen, Lexeme: "("},
				{Kind: TokenText, Lexeme: "d"},
				{Kind: TokenCloseParen, Lexeme: ")"},
				{Kind: TokenBang, Lexeme: "!"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "spaces preserved around tokens",
			input: "a [ b ] c",
			span:  source.ByteSpan{Start: 0, End: 9},
			want: []TokenSummary{
				{Kind: TokenText, Lexeme: "a "},
				{Kind: TokenOpenBracket, Lexeme: "["},
				{Kind: TokenText, Lexeme: " b "},
				{Kind: TokenCloseBracket, Lexeme: "]"},
				{Kind: TokenText, Lexeme: " c"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},

		// Special-case precedence and boundaries
		{
			name:  "bang not followed by open bracket",
			input: "!x",
			span:  source.ByteSpan{Start: 0, End: 2},
			want: []TokenSummary{
				{Kind: TokenBang, Lexeme: "!"},
				{Kind: TokenText, Lexeme: "x"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "double bang before bracket",
			input: "!![",
			span:  source.ByteSpan{Start: 0, End: 3},
			want: []TokenSummary{
				{Kind: TokenBang, Lexeme: "!"},
				{Kind: TokenImageOpenBracket, Lexeme: "!["},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "bang separated from bracket by space",
			input: "! [",
			span:  source.ByteSpan{Start: 0, End: 3},
			want: []TokenSummary{
				{Kind: TokenBang, Lexeme: "!"},
				{Kind: TokenText, Lexeme: " "},
				{Kind: TokenOpenBracket, Lexeme: "["},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "backslash before image opener remains separate tokens in scan",
			input: `\![`,
			span:  source.ByteSpan{Start: 0, End: 3},
			want: []TokenSummary{
				{Kind: TokenBackslash, Lexeme: `\`},
				{Kind: TokenImageOpenBracket, Lexeme: "!["},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},

		// Run boundaries
		{
			name:  "star run followed by text",
			input: "***a",
			span:  source.ByteSpan{Start: 0, End: 4},
			want: []TokenSummary{
				{Kind: TokenStarDelimiter, Lexeme: "***"},
				{Kind: TokenText, Lexeme: "a"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "text followed by star run",
			input: "a***",
			span:  source.ByteSpan{Start: 0, End: 4},
			want: []TokenSummary{
				{Kind: TokenText, Lexeme: "a"},
				{Kind: TokenStarDelimiter, Lexeme: "***"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "star run between text",
			input: "a***b",
			span:  source.ByteSpan{Start: 0, End: 5},
			want: []TokenSummary{
				{Kind: TokenText, Lexeme: "a"},
				{Kind: TokenStarDelimiter, Lexeme: "***"},
				{Kind: TokenText, Lexeme: "b"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "underscore run followed by text",
			input: "__a",
			span:  source.ByteSpan{Start: 0, End: 3},
			want: []TokenSummary{
				{Kind: TokenUnderscoreDelimiter, Lexeme: "__"},
				{Kind: TokenText, Lexeme: "a"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "backtick run followed by text",
			input: "```a",
			span:  source.ByteSpan{Start: 0, End: 4},
			want: []TokenSummary{
				{Kind: TokenBacktick, Lexeme: "```"},
				{Kind: TokenText, Lexeme: "a"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "mixed delimiter bytes are not one run",
			input: "*_",
			span:  source.ByteSpan{Start: 0, End: 2},
			want: []TokenSummary{
				{Kind: TokenStarDelimiter, Lexeme: "*"},
				{Kind: TokenUnderscoreDelimiter, Lexeme: "_"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "backticks separated by text do not merge",
			input: "a`b`c",
			span:  source.ByteSpan{Start: 0, End: 5},
			want: []TokenSummary{
				{Kind: TokenText, Lexeme: "a"},
				{Kind: TokenBacktick, Lexeme: "`"},
				{Kind: TokenText, Lexeme: "b"},
				{Kind: TokenBacktick, Lexeme: "`"},
				{Kind: TokenText, Lexeme: "c"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},

		// Span-restricted scanning
		{
			name:  "scan middle slice of plain text",
			input: "abcdef",
			span:  source.ByteSpan{Start: 2, End: 4},
			want: []TokenSummary{
				{Kind: TokenText, Lexeme: "cd"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "scan slice beginning at delimiter",
			input: "xx*abc*yy",
			span:  source.ByteSpan{Start: 2, End: 7},
			want: []TokenSummary{
				{Kind: TokenStarDelimiter, Lexeme: "*"},
				{Kind: TokenText, Lexeme: "abc"},
				{Kind: TokenStarDelimiter, Lexeme: "*"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "scan slice containing only one token from larger source",
			input: "xx![a](b)yy",
			span:  source.ByteSpan{Start: 2, End: 4},
			want: []TokenSummary{
				{Kind: TokenImageOpenBracket, Lexeme: "!["},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "scan slice over bracketed text in larger source",
			input: "a[b]c",
			span:  source.ByteSpan{Start: 1, End: 4},
			want: []TokenSummary{
				{Kind: TokenOpenBracket, Lexeme: "["},
				{Kind: TokenText, Lexeme: "b"},
				{Kind: TokenCloseBracket, Lexeme: "]"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
		{
			name:  "scan slice ending before EOF",
			input: "abc*def",
			span:  source.ByteSpan{Start: 0, End: 4},
			want: []TokenSummary{
				{Kind: TokenText, Lexeme: "abc"},
				{Kind: TokenStarDelimiter, Lexeme: "*"},
				{Kind: TokenEOF},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := source.NewSource(tc.input)
			tokens, err := Scan(src, tc.span)
			got := summarizeTokens(src, tokens)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
