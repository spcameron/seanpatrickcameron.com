package inline

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/require"
)

func TestBuild(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		span    source.ByteSpan
		defs    map[string]ir.ReferenceDefinition
		want    []InlineSummary
		wantErr error
	}{
		// Empty input and plain text
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

		// Code spans
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
			name:  "code span: only one leading trailing space is stripped",
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
			name:  "code span: unmatched opener falls back to text",
			input: "`foo",
			want: []InlineSummary{
				{
					Kind:   "text",
					Lexeme: "`",
				},
				{
					Kind:   "text",
					Lexeme: "foo",
				},
			},
			wantErr: nil,
		},
		{
			name:  "code span: unmatched closer falls back to text",
			input: "foo`",
			want: []InlineSummary{
				{
					Kind:   "text",
					Lexeme: "foo",
				},
				{
					Kind:   "text",
					Lexeme: "`",
				},
			},
			wantErr: nil,
		},

		// Emphasis and strong
		{
			name:  "emphasis: star",
			input: "*alt*",
			want: []InlineSummary{
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
			wantErr: nil,
		},
		{
			name:  "emphasis: underscore",
			input: "_alt_",
			want: []InlineSummary{
				{
					Kind:   "emphasis",
					Lexeme: "_alt_",
					Children: []InlineSummary{
						{
							Kind:   "text",
							Lexeme: "alt",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "strong: star",
			input: "**alt**",
			want: []InlineSummary{
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
			wantErr: nil,
		},
		{
			name:  "strong: underscore",
			input: "__alt__",
			want: []InlineSummary{
				{
					Kind:   "strong",
					Lexeme: "__alt__",
					Children: []InlineSummary{
						{
							Kind:   "text",
							Lexeme: "alt",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "strong: nested emphasis",
			input: "***alt***",
			want: []InlineSummary{
				{
					Kind:   "emphasis",
					Lexeme: "***alt***",
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
			wantErr: nil,
		},
		{
			name:  "emphasis: unmatched opener falls back to text",
			input: "*alt",
			want: []InlineSummary{
				{Kind: "text", Lexeme: "*"},
				{Kind: "text", Lexeme: "alt"},
			},
			wantErr: nil,
		},
		{
			name:  "emphasis: unmatched closer falls back to text",
			input: "alt*",
			want: []InlineSummary{
				{Kind: "text", Lexeme: "alt"},
				{Kind: "text", Lexeme: "*"},
			},
			wantErr: nil,
		},
		{
			name:  "emphasis: contains code span",
			input: "*a `b` c*",
			want: []InlineSummary{
				{
					Kind:   "emphasis",
					Lexeme: "*a `b` c*",
					Children: []InlineSummary{
						{Kind: "text", Lexeme: "a "},
						{Kind: "code_span", Lexeme: "b"},
						{Kind: "text", Lexeme: " c"},
					},
				},
			},
			wantErr: nil,
		},

		// Autolinks
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
			name:  "autolink: invalid scheme falls back to text",
			input: "<x:y>",
			want: []InlineSummary{
				{Kind: "text", Lexeme: "<"},
				{Kind: "text", Lexeme: "x:y"},
				{Kind: "text", Lexeme: ">"},
			},
			wantErr: nil,
		},
		{
			name:  "autolink: invalid email falls back to text",
			input: "<local@-domain.com>",
			want: []InlineSummary{
				{Kind: "text", Lexeme: "<"},
				{Kind: "text", Lexeme: "local@-domain.com"},
				{Kind: "text", Lexeme: ">"},
			},
			wantErr: nil,
		},

		// Inline HTML
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
			name:  "html: self closing open tag",
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
			name:  "html: self closing open tag with whitespace",
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
			wantErr: nil,
		},
		{
			name:  "html: comment",
			input: "<!-- comment -->",
			want: []InlineSummary{
				{
					Kind:   "raw_text",
					Lexeme: "<!-- comment -->",
				},
			},
			wantErr: nil,
		},
		{
			name:  "html: declaration",
			input: "<!DOCTYPE html>",
			want: []InlineSummary{
				{
					Kind:   "raw_text",
					Lexeme: "<!DOCTYPE html>",
				},
			},
			wantErr: nil,
		},
		{
			name:  "html: processing instruction",
			input: "<?xml version=\"1.0\"?>",
			want: []InlineSummary{
				{
					Kind:   "raw_text",
					Lexeme: "<?xml version=\"1.0\"?>",
				},
			},
			wantErr: nil,
		},
		{
			name:  "html: cdata",
			input: "<![CDATA[hello]]>",
			want: []InlineSummary{
				{
					Kind:   "raw_text",
					Lexeme: "<![CDATA[hello]]>",
				},
			},
			wantErr: nil,
		},
		{
			name:  "html: invalid tag falls back to text",
			input: "<1a>",
			want: []InlineSummary{
				{Kind: "text", Lexeme: "<"},
				{Kind: "text", Lexeme: "1a"},
				{Kind: "text", Lexeme: ">"},
			},
			wantErr: nil,
		},

		// Links
		{
			name:  "link: simple",
			input: "[alt](dest)",
			want: []InlineSummary{
				{
					Kind:   "link",
					Lexeme: "[alt](dest)",
					Children: []InlineSummary{
						{
							Kind:   "text",
							Lexeme: "alt",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "link: with title",
			input: `[alt](dest "title")`,
			want: []InlineSummary{
				{
					Kind:   "link",
					Lexeme: `[alt](dest "title")`,
					Children: []InlineSummary{
						{
							Kind:   "text",
							Lexeme: "alt",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "link: emphasis in label",
			input: "[*alt*](dest)",
			want: []InlineSummary{
				{
					Kind:   "link",
					Lexeme: "[*alt*](dest)",
					Children: []InlineSummary{
						{
							Kind:   "emphasis",
							Lexeme: "*alt*",
							Children: []InlineSummary{
								{Kind: "text", Lexeme: "alt"},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "link: strong in label",
			input: "[**alt**](dest)",
			want: []InlineSummary{
				{
					Kind:   "link",
					Lexeme: "[**alt**](dest)",
					Children: []InlineSummary{
						{
							Kind:   "strong",
							Lexeme: "**alt**",
							Children: []InlineSummary{
								{Kind: "text", Lexeme: "alt"},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "link: code span in label",
			input: "[a `b` c](dest)",
			want: []InlineSummary{
				{
					Kind:   "link",
					Lexeme: "[a `b` c](dest)",
					Children: []InlineSummary{
						{Kind: "text", Lexeme: "a "},
						{Kind: "code_span", Lexeme: "b"},
						{Kind: "text", Lexeme: " c"},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "link: missing tail falls back to text",
			input: "[alt]",
			want: []InlineSummary{
				{Kind: "text", Lexeme: "["},
				{Kind: "text", Lexeme: "alt"},
				{Kind: "text", Lexeme: "]"},
			},
			wantErr: nil,
		},
		{
			name:  "link: malformed tail falls back to text",
			input: "[alt](dest",
			want: []InlineSummary{
				{Kind: "text", Lexeme: "["},
				{Kind: "text", Lexeme: "alt"},
				{Kind: "text", Lexeme: "]"},
				{Kind: "text", Lexeme: "("},
				{Kind: "text", Lexeme: "dest"},
			},
			wantErr: nil,
		},
		{
			name:  "link: escaped opener remains text",
			input: `\[alt](dest)`,
			want: []InlineSummary{
				{Kind: "text", Lexeme: "["},
				{Kind: "text", Lexeme: "alt"},
				{Kind: "text", Lexeme: "]"},
				{Kind: "text", Lexeme: "("},
				{Kind: "text", Lexeme: "dest"},
				{Kind: "text", Lexeme: ")"},
			},
			wantErr: nil,
		},

		// Images
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
			wantErr: nil,
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
			wantErr: nil,
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
			wantErr: nil,
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
		{
			name:  "image: emphasis and code span in alt text",
			input: "![*a* `b`](img.png)",
			want: []InlineSummary{
				{
					Kind:   "image",
					Lexeme: "![*a* `b`](img.png)",
					Children: []InlineSummary{
						{
							Kind:   "emphasis",
							Lexeme: "*a*",
							Children: []InlineSummary{
								{Kind: "text", Lexeme: "a"},
							},
						},
						{Kind: "text", Lexeme: " "},
						{Kind: "code_span", Lexeme: "b"},
					},
				},
			},
		},

		// Reference links
		{
			name:  "reference link: full reference",
			input: "/url [foo][bar]",
			span: source.ByteSpan{
				Start: 5,
				End:   15,
			},
			defs: map[string]ir.ReferenceDefinition{
				"bar": {
					DestinationSpan: source.ByteSpan{Start: 0, End: 4},
					NormalizedKey:   "bar",
				},
			},
			want: []InlineSummary{
				{
					Kind:   "link",
					Lexeme: "[foo][bar]",
					Children: []InlineSummary{
						{Kind: "text", Lexeme: "foo"},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference link: collapsed reference",
			input: "/url [foo][]",
			span: source.ByteSpan{
				Start: 5,
				End:   12,
			},
			defs: map[string]ir.ReferenceDefinition{
				"foo": {
					DestinationSpan: source.ByteSpan{Start: 0, End: 4},
					NormalizedKey:   "foo",
				},
			},
			want: []InlineSummary{
				{
					Kind:   "link",
					Lexeme: "[foo][]",
					Children: []InlineSummary{
						{Kind: "text", Lexeme: "foo"},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference link: shortcut reference",
			input: "/url [foo]",
			span: source.ByteSpan{
				Start: 5,
				End:   10,
			},
			defs: map[string]ir.ReferenceDefinition{
				"foo": {
					DestinationSpan: source.ByteSpan{Start: 0, End: 4},
					NormalizedKey:   "foo",
				},
			},
			want: []InlineSummary{
				{
					Kind:   "link",
					Lexeme: "[foo]",
					Children: []InlineSummary{
						{Kind: "text", Lexeme: "foo"},
					},
				},
			},
			wantErr: nil,
		},

		// Reference images
		{
			name:  "reference image: full reference",
			input: "/url ![foo][bar]",
			span: source.ByteSpan{
				Start: 5,
				End:   16,
			},
			defs: map[string]ir.ReferenceDefinition{
				"bar": {
					DestinationSpan: source.ByteSpan{Start: 0, End: 4},
					NormalizedKey:   "bar",
				},
			},
			want: []InlineSummary{
				{
					Kind:   "image",
					Lexeme: "![foo][bar]",
					Children: []InlineSummary{
						{Kind: "text", Lexeme: "foo"},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference image: collapsed reference",
			input: "/url ![foo][]",
			span: source.ByteSpan{
				Start: 5,
				End:   13,
			},
			defs: map[string]ir.ReferenceDefinition{
				"foo": {
					DestinationSpan: source.ByteSpan{Start: 0, End: 4},
					NormalizedKey:   "foo",
				},
			},
			want: []InlineSummary{
				{
					Kind:   "image",
					Lexeme: "![foo][]",
					Children: []InlineSummary{
						{Kind: "text", Lexeme: "foo"},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "reference image: shortcut reference",
			input: "/url ![foo]",
			span: source.ByteSpan{
				Start: 5,
				End:   11,
			},
			defs: map[string]ir.ReferenceDefinition{
				"foo": {
					DestinationSpan: source.ByteSpan{Start: 0, End: 4},
					NormalizedKey:   "foo",
				},
			},
			want: []InlineSummary{
				{
					Kind:   "image",
					Lexeme: "![foo]",
					Children: []InlineSummary{
						{Kind: "text", Lexeme: "foo"},
					},
				},
			},
			wantErr: nil,
		},

		// Escapes
		{
			name:  "escape: star becomes text",
			input: `\*`,
			want: []InlineSummary{
				{Kind: "text", Lexeme: "*"},
			},
			wantErr: nil,
		},
		{
			name:  "escape: underscore becomes text",
			input: `\_`,
			want: []InlineSummary{
				{Kind: "text", Lexeme: "_"},
			},
			wantErr: nil,
		},
		{
			name:  "escape: backtick becomes text",
			input: "\\`",
			want: []InlineSummary{
				{Kind: "text", Lexeme: "`"},
			},
			wantErr: nil,
		},
		{
			name:  "escape: open bracket becomes text",
			input: `\[`,
			want: []InlineSummary{
				{Kind: "text", Lexeme: "["},
			},
			wantErr: nil,
		},
		{
			name:  "escape: close bracket becomes text",
			input: `\]`,
			want: []InlineSummary{
				{Kind: "text", Lexeme: "]"},
			},
			wantErr: nil,
		},
		{
			name:  "escape: ordinary text remains literal",
			input: `\a`,
			want: []InlineSummary{
				{Kind: "text", Lexeme: "\\"},
				{Kind: "text", Lexeme: "a"},
			},
			wantErr: nil,
		},
		{
			name:  "escape: trailing backslash at eof",
			input: `\`,
			want: []InlineSummary{
				{Kind: "text", Lexeme: `\`},
			},
			wantErr: nil,
		},

		// Interaction and precedence
		{
			name:  "interaction: code span beats emphasis",
			input: "*a `b*`",
			want: []InlineSummary{
				{Kind: "text", Lexeme: "*"},
				{Kind: "text", Lexeme: "a "},
				{Kind: "code_span", Lexeme: "b*"},
			},
			wantErr: nil,
		},
		{
			name:  "interaction: link label resolves emphasis inside label",
			input: "[a *b*](dest)",
			want: []InlineSummary{
				{
					Kind:   "link",
					Lexeme: "[a *b*](dest)",
					Children: []InlineSummary{
						{Kind: "text", Lexeme: "a "},
						{
							Kind:   "emphasis",
							Lexeme: "*b*",
							Children: []InlineSummary{
								{Kind: "text", Lexeme: "b"},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "interaction: malformed emphasis inside link still yields link text children",
			input: "[a *b](dest)",
			want: []InlineSummary{
				{
					Kind:   "link",
					Lexeme: "[a *b](dest)",
					Children: []InlineSummary{
						{Kind: "text", Lexeme: "a "},
						{Kind: "text", Lexeme: "*"},
						{Kind: "text", Lexeme: "b"},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:  "interaction: adjacent link and text",
			input: "[x](y)z",
			want: []InlineSummary{
				{
					Kind:   "link",
					Lexeme: "[x](y)",
					Children: []InlineSummary{
						{Kind: "text", Lexeme: "x"},
					},
				},
				{Kind: "text", Lexeme: "z"},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := source.NewSource(tc.input)

			span := tc.span
			if span == (source.ByteSpan{}) {
				span = source.ByteSpan{
					Start: 0,
					End:   src.EOF(),
				}
			}

			tokens, err := Scan(src, span)
			require.NoError(t, err)

			defs := tc.defs
			if defs == nil {
				defs = map[string]ir.ReferenceDefinition{}
			}

			inlines, err := Build(src, defs, span, tokens)
			got := summarizeInlines(src, inlines)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
