package inline

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/require"
)

func TestBuild(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []InlineSummary
		wantErr error
	}{
		{
			name:    "empty input",
			input:   "",
			want:    []InlineSummary{},
			wantErr: nil,
		},
		{
			name:  "plain text",
			input: "abc",
			want: []InlineSummary{
				{
					Kind:   "text",
					Lexeme: "abc",
				},
			},
			wantErr: nil,
		},
		{
			name:  "code span",
			input: "`foo`",
			want: []InlineSummary{
				{
					Kind:   "code_span",
					Lexeme: "foo",
				},
			},
			wantErr: nil,
		},
		{
			name:  "code span: two backticks",
			input: "``foo`bar``",
			want: []InlineSummary{
				{
					Kind:   "code_span",
					Lexeme: "foo`bar",
				},
			},
			wantErr: nil,
		},
		{
			name:  "code span: leading and trailing spaces",
			input: "` `` `",
			want: []InlineSummary{
				{
					Kind:   "code_span",
					Lexeme: "``",
				},
			},
			wantErr: nil,
		},
		{
			name:  "code span: only one leading/trailing space is stripped",
			input: "`  ``  `",
			want: []InlineSummary{
				{
					Kind:   "code_span",
					Lexeme: " `` ",
				},
			},
			wantErr: nil,
		},
		{
			name:  "code span: stripping only if the space is on both sides",
			input: "` a`",
			want: []InlineSummary{
				{
					Kind:   "code_span",
					Lexeme: " a",
				},
			},
			wantErr: nil,
		},
		{
			name:  "code span: interior spaces are not collapsed",
			input: "`foo   bar`",
			want: []InlineSummary{
				{
					Kind:   "code_span",
					Lexeme: "foo   bar",
				},
			},
			wantErr: nil,
		},
		{
			name:  "autolink: absolute URI",
			input: "<http://foo.bar.baz>",
			want: []InlineSummary{
				{
					Kind:   "link",
					Lexeme: "<http://foo.bar.baz>",
					Children: []InlineSummary{
						{
							Kind:   "text",
							Lexeme: "http://foo.bar.baz",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "autolink: email",
			input: "<local@domain.com>",
			want: []InlineSummary{
				{
					Kind:   "link",
					Lexeme: "<local@domain.com>",
					Children: []InlineSummary{
						{
							Kind:   "text",
							Lexeme: "local@domain.com",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "html: simple open tag",
			input: "<a>",
			want: []InlineSummary{
				{
					Kind:   "raw_text",
					Lexeme: "<a>",
				},
			},
			wantErr: nil,
		},
		{
			name:  "html: open tag with trailing whitespace",
			input: "<a >",
			want: []InlineSummary{
				{
					Kind:   "raw_text",
					Lexeme: "<a >",
				},
			},
			wantErr: nil,
		},
		{
			name:  "html: self-closing open tag",
			input: "<a/>",
			want: []InlineSummary{
				{
					Kind:   "raw_text",
					Lexeme: "<a/>",
				},
			},
			wantErr: nil,
		},
		{
			name:  "html: self-closing open tag with whitespace",
			input: "<a />",
			want: []InlineSummary{
				{
					Kind:   "raw_text",
					Lexeme: "<a />",
				},
			},
			wantErr: nil,
		},
		{
			name:  "html: closing tag",
			input: "</a>",
			want: []InlineSummary{
				{
					Kind:   "raw_text",
					Lexeme: "</a>",
				},
			},
		},
		{
			name:  "image: simple",
			input: "![alt](img.png)",
			want: []InlineSummary{
				{
					Kind:   "image",
					Lexeme: "![alt](img.png)",
					Children: []InlineSummary{
						{
							Kind:   "text",
							Lexeme: "alt",
						},
					},
				},
			},
		},
		{
			name:  "image: with title",
			input: `![alt](img.png "title")`,
			want: []InlineSummary{
				{
					Kind:   "image",
					Lexeme: `![alt](img.png "title")`,
					Children: []InlineSummary{
						{
							Kind:   "text",
							Lexeme: "alt",
						},
					},
				},
			},
		},
		{
			name:  "image: empty alt",
			input: "![](img.png)",
			want: []InlineSummary{
				{
					Kind:     "image",
					Lexeme:   "![](img.png)",
					Children: []InlineSummary{},
				},
			},
		},
		{
			name:  "image: emphasis in alt text",
			input: "![*alt*](img.png)",
			want: []InlineSummary{
				{
					Kind:   "image",
					Lexeme: "![*alt*](img.png)",
					Children: []InlineSummary{
						{
							Kind:   "emphasis",
							Lexeme: "*alt*",
							Children: []InlineSummary{
								{
									Kind:   "text",
									Lexeme: "alt",
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "image: strong in alt text",
			input: "![**alt**](img.png)",
			want: []InlineSummary{
				{
					Kind:   "image",
					Lexeme: "![**alt**](img.png)",
					Children: []InlineSummary{
						{
							Kind:   "strong",
							Lexeme: "**alt**",
							Children: []InlineSummary{
								{
									Kind:   "text",
									Lexeme: "alt",
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "image: mixed alt children",
			input: "![a `b` c](img.png)",
			want: []InlineSummary{
				{
					Kind:   "image",
					Lexeme: "![a `b` c](img.png)",
					Children: []InlineSummary{
						{Kind: "text", Lexeme: "a "},
						{Kind: "code_span", Lexeme: "b"},
						{Kind: "text", Lexeme: " c"},
					},
				},
			},
		},
		{
			name:  "image: escaped bang becomes text plus link",
			input: `\![alt](img.png)`,
			want: []InlineSummary{
				{
					Kind:   "text",
					Lexeme: "!",
				},
				{
					Kind:   "link",
					Lexeme: "[alt](img.png)",
					Children: []InlineSummary{
						{
							Kind:   "text",
							Lexeme: "alt",
						},
					},
				},
			},
		},
		{
			name:  "image: missing tail falls back to text",
			input: "![alt]",
			want: []InlineSummary{
				{Kind: "text", Lexeme: "!["},
				{Kind: "text", Lexeme: "alt"},
				{Kind: "text", Lexeme: "]"},
			},
		},
		{
			name:  "image: link inside alt text",
			input: "![see [x](y)](img.png)",
			want: []InlineSummary{
				{
					Kind:   "image",
					Lexeme: "![see [x](y)](img.png)",
					Children: []InlineSummary{
						{Kind: "text", Lexeme: "see "},
						{
							Kind:   "link",
							Lexeme: "[x](y)",
							Children: []InlineSummary{
								{Kind: "text", Lexeme: "x"},
							},
						},
					},
				},
			},
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
			require.NoError(t, err)

			inlines, err := Build(src, span, tokens)
			got := summarizeInlines(src, inlines)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
