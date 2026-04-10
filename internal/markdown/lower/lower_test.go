package lower_test

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/block"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/lower"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
	tk "github.com/spcameron/seanpatrickcameron.com/internal/markdown/testkit"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/require"
)

func TestLowerDocument(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    ast.Document
		wantErr error
	}{
		// Paragraphs
		{
			name:  "paragraph with normal text",
			input: "paragraph",
			want: tk.ASTDoc(
				tk.ASTPara(
					tk.ASTText("paragraph"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: soft break between lines",
			input: "alpha\nbeta",
			want: tk.ASTDoc(
				tk.ASTPara(
					tk.ASTText("alpha"),
					tk.ASTSoftBreak(),
					tk.ASTText("beta"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: hard break via two trailing spaces",
			input: "alpha  \nbeta",
			want: tk.ASTDoc(
				tk.ASTPara(
					tk.ASTText("alpha"),
					tk.ASTHardBreak(),
					tk.ASTText("beta"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: hard break via trailing backslash",
			input: "alpha\\\nbeta",
			want: tk.ASTDoc(
				tk.ASTPara(
					tk.ASTText("alpha"),
					tk.ASTHardBreak(),
					tk.ASTText("beta"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: mixed soft and hard breaks across three lines",
			input: "alpha\nbeta  \ngamma",
			want: tk.ASTDoc(
				tk.ASTPara(
					tk.ASTText("alpha"),
					tk.ASTSoftBreak(),
					tk.ASTText("beta"),
					tk.ASTHardBreak(),
					tk.ASTText("gamma"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "paragraph: emphasis cannot span lowered line boundary",
			input: "*alpha\nbeta*",
			want: tk.ASTDoc(
				tk.ASTPara(
					tk.ASTText("*"),
					tk.ASTText("alpha"),
					tk.ASTSoftBreak(),
					tk.ASTText("beta"),
					tk.ASTText("*"),
				),
			),
			wantErr: nil,
		},

		// Headers
		{
			name:  "header with normal text",
			input: "# header",
			want: tk.ASTDoc(
				tk.ASTHeader(
					1,
					tk.ASTText("header"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "header with emphasis",
			input: "# *header*",
			want: tk.ASTDoc(
				tk.ASTHeader(
					1,
					tk.ASTEm(
						tk.ASTText("header"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "header and paragraph",
			input: "# header\n\nparagraph",
			want: tk.ASTDoc(
				tk.ASTHeader(
					1,
					tk.ASTText("header"),
				),
				tk.ASTPara(
					tk.ASTText("paragraph"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "header: strong and emphasis",
			input: "# **alpha** *beta*",
			want: tk.ASTDoc(
				tk.ASTHeader(
					1,
					tk.ASTStrong(
						tk.ASTText("alpha"),
					),
					tk.ASTText(" "),
					tk.ASTEm(
						tk.ASTText("beta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "header: code span",
			input: "# `alpha`",
			want: tk.ASTDoc(
				tk.ASTHeader(
					1,
					tk.ASTCodeSpan("alpha"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "setext header: plain text",
			input: "alpha\n---",
			want: tk.ASTDoc(
				tk.ASTHeader(
					2,
					tk.ASTText("alpha"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "setext header: inline emphasis",
			input: "*alpha*\n---",
			want: tk.ASTDoc(
				tk.ASTHeader(
					2,
					tk.ASTEm(
						tk.ASTText("alpha"),
					),
				),
			),
			wantErr: nil,
		},

		// Simple block forms
		{
			name:  "thematic break",
			input: "---",
			want: tk.ASTDoc(
				tk.ASTThematicBreak(),
			),
			wantErr: nil,
		},
		{
			name:  "html block",
			input: "<!-- comment -->",
			want: tk.ASTDoc(
				tk.ASTHTMLBlock(
					tk.ASTRawText("<!-- comment -->"),
				),
			),
			wantErr: nil,
		},

		// Block quotes
		{
			name:  "block quote: plain text",
			input: "> quote",
			want: tk.ASTDoc(
				tk.ASTBlockQuote(
					tk.ASTPara(
						tk.ASTText("quote"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: two paragraphs",
			input: "> alpha\n>\n> beta",
			want: tk.ASTDoc(
				tk.ASTBlockQuote(
					tk.ASTPara(
						tk.ASTText("alpha"),
					),
					tk.ASTPara(
						tk.ASTText("beta"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: nested block quote",
			input: "> outer\n> > inner",
			want: tk.ASTDoc(
				tk.ASTBlockQuote(
					tk.ASTPara(
						tk.ASTText("outer"),
					),
					tk.ASTBlockQuote(
						tk.ASTPara(
							tk.ASTText("inner"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: contains list",
			input: "> - alpha\n> - beta",
			want: tk.ASTDoc(
				tk.ASTBlockQuote(
					tk.ASTUnorderedList(
						true,
						tk.ASTListItem(
							tk.ASTPara(
								tk.ASTText("alpha"),
							),
						),
						tk.ASTListItem(
							tk.ASTPara(
								tk.ASTText("beta"),
							),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "block quote: contains code block",
			input: ">     alpha",
			want: tk.ASTDoc(
				tk.ASTBlockQuote(
					tk.ASTIndentedCodeBlock(
						tk.ASTText("alpha"),
					),
				),
			),
			wantErr: nil,
		},

		// Lists
		{
			name:  "unordered list: two items",
			input: "- a\n- b",
			want: tk.ASTDoc(
				tk.ASTUnorderedList(
					true,
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("a"),
						),
					),
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("b"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "unordered list: loose list preserves paragraph children",
			input: "- alpha\n\n- beta",
			want: tk.ASTDoc(
				tk.ASTUnorderedList(
					false,
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("alpha"),
						),
					),
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("beta"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "unordered list: item with two paragraphs",
			input: "- alpha\n\n  beta",
			want: tk.ASTDoc(
				tk.ASTUnorderedList(
					false,
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("alpha"),
						),
						tk.ASTPara(
							tk.ASTText("beta"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "unordered list: nested unordered list",
			input: "- alpha\n  - beta",
			want: tk.ASTDoc(
				tk.ASTUnorderedList(
					true,
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("alpha"),
						),
						tk.ASTUnorderedList(
							true,
							tk.ASTListItem(
								tk.ASTPara(
									tk.ASTText("beta"),
								),
							),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "unordered list: nested ordered list",
			input: "- alpha\n  1. beta",
			want: tk.ASTDoc(
				tk.ASTUnorderedList(
					true,
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("alpha"),
						),
						tk.ASTOrderedList(
							true,
							1,
							tk.ASTListItem(
								tk.ASTPara(
									tk.ASTText("beta"),
								),
							),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ordered list: two items",
			input: "1. a\n2. b",
			want: tk.ASTDoc(
				tk.ASTOrderedList(
					true,
					1,
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("a"),
						),
					),
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("b"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ordered list: start number preserved",
			input: "3. alpha\n4. beta",
			want: tk.ASTDoc(
				tk.ASTOrderedList(
					true,
					3,
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("alpha"),
						),
					),
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("beta"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ordered list: paren delimiter still lowers as ordered list",
			input: "1) alpha\n2) beta",
			want: tk.ASTDoc(
				tk.ASTOrderedList(
					true,
					1,
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("alpha"),
						),
					),
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("beta"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "list item: indented code block child",
			input: "- alpha\n\n      beta",
			want: tk.ASTDoc(
				tk.ASTUnorderedList(
					false,
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("alpha"),
						),
						tk.ASTIndentedCodeBlock(
							tk.ASTText("beta"),
						),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "list item: block quote child",
			input: "- alpha\n  > beta",
			want: tk.ASTDoc(
				tk.ASTUnorderedList(
					true,
					tk.ASTListItem(
						tk.ASTPara(
							tk.ASTText("alpha"),
						),
						tk.ASTBlockQuote(
							tk.ASTPara(
								tk.ASTText("beta"),
							),
						),
					),
				),
			),
			wantErr: nil,
		},

		// Code blocks
		{
			name:  "indented code block",
			input: "    code",
			want: tk.ASTDoc(
				tk.ASTIndentedCodeBlock(
					tk.ASTText("code"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "indented code block: two lines preserves internal newline",
			input: "    alpha\n    beta",
			want: tk.ASTDoc(
				tk.ASTIndentedCodeBlock(
					tk.ASTText("alpha"),
					tk.ASTNewline(),
					tk.ASTText("beta"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "indented code block: blank line inside payload",
			input: "    alpha\n\n    beta",
			want: tk.ASTDoc(
				tk.ASTIndentedCodeBlock(
					tk.ASTText("alpha"),
					tk.ASTNewline(),
					tk.ASTText(""),
					tk.ASTNewline(),
					tk.ASTText("beta"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "fenced code block",
			input: "```\ncode\n```",
			want: tk.ASTDoc(
				tk.ASTFencedCodeBlock(
					tk.ASTText("code"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "fenced code block: two payload lines preserves internal newline",
			input: "```\nalpha\nbeta\n```",
			want: tk.ASTDoc(
				tk.ASTFencedCodeBlock(
					tk.ASTText("alpha"),
					tk.ASTNewline(),
					tk.ASTText("beta"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "fenced code block: blank line inside payload",
			input: "```\nalpha\n\nbeta\n```",
			want: tk.ASTDoc(
				tk.ASTFencedCodeBlock(
					tk.ASTText("alpha"),
					tk.ASTNewline(),
					tk.ASTText(""),
					tk.ASTNewline(),
					tk.ASTText("beta"),
				),
			),
			wantErr: nil,
		},

		// Basic inline lowering through blocks
		{
			name:  "emphasis",
			input: "*abc*",
			want: tk.ASTDoc(
				tk.ASTPara(
					tk.ASTEm(
						tk.ASTText("abc"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "strong",
			input: "**abc**",
			want: tk.ASTDoc(
				tk.ASTPara(
					tk.ASTStrong(
						tk.ASTText("abc"),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "code span",
			input: "`abc`",
			want: tk.ASTDoc(
				tk.ASTPara(
					tk.ASTCodeSpan("abc"),
				),
			),
			wantErr: nil,
		},

		// HTML
		{
			name:  "html block: multi-line comment preserves internal newline",
			input: "<!--\nalpha\n-->",
			want: tk.ASTDoc(
				tk.ASTHTMLBlock(
					tk.ASTRawText("<!--"),
					tk.ASTNewline(),
					tk.ASTRawText("alpha"),
					tk.ASTNewline(),
					tk.ASTRawText("-->"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "html block: named tag block preserves multiple raw lines",
			input: "<div>\nalpha\n</div>",
			want: tk.ASTDoc(
				tk.ASTHTMLBlock(
					tk.ASTRawText("<div>"),
					tk.ASTNewline(),
					tk.ASTRawText("alpha"),
					tk.ASTNewline(),
					tk.ASTRawText("</div>"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "html block: processing instruction",
			input: "<?php\necho $a;\n?>",
			want: tk.ASTDoc(
				tk.ASTHTMLBlock(
					tk.ASTRawText("<?php"),
					tk.ASTNewline(),
					tk.ASTRawText("echo $a;"),
					tk.ASTNewline(),
					tk.ASTRawText("?>"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "html block: cdata section",
			input: "<![CDATA[\nalpha\n]]>",
			want: tk.ASTDoc(
				tk.ASTHTMLBlock(
					tk.ASTRawText("<![CDATA["),
					tk.ASTNewline(),
					tk.ASTRawText("alpha"),
					tk.ASTNewline(),
					tk.ASTRawText("]]>"),
				),
			),
			wantErr: nil,
		},
		{
			name:  "html block: declaration",
			input: "<!DOCTYPE html>",
			want: tk.ASTDoc(
				tk.ASTHTMLBlock(
					tk.ASTRawText("<!DOCTYPE html>"),
				),
			),
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := source.NewSource(tc.input)

			irDoc, err := block.Parse(src)
			require.NoError(t, err)

			got, err := lower.Document(irDoc)

			got = tk.NormalizeAST(got)
			want := tk.NormalizeAST(tc.want)

			assert.Equal(t, got, want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
