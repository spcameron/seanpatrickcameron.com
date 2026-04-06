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
			name:  "backslash",
			input: `\`,
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
