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

// TODO:
// lowerDocumentCoverageGaps is a curated set of additional lower-layer cases worth
// adding once expected AST shapes are filled in. These are organized around the
// places where lowering itself makes semantic decisions, rather than merely passing
// IR through unchanged.
var lowerDocumentCoverageGaps = []struct {
	name  string
	input string
}{
	// Paragraph lowering: soft/hard break assembly happens in lower.
	{
		name:  "paragraph: soft break between lines",
		input: "alpha\nbeta",
	},
	{
		name:  "paragraph: hard break via two trailing spaces",
		input: "alpha  \nbeta",
	},
	{
		name:  "paragraph: hard break via trailing backslash",
		input: "alpha\\\nbeta",
	},
	{
		name:  "paragraph: mixed soft and hard breaks across three lines",
		input: "alpha\nbeta  \ngamma",
	},
	{
		name:  "paragraph: emphasis cannot span lowered line boundary",
		input: "*alpha\nbeta*",
	},

	// Header lowering: inline parsing occurs during lowering.
	{
		name:  "header: strong and emphasis",
		input: "# **alpha** *beta*",
	},
	{
		name:  "header: code span",
		input: "# `alpha`",
	},
	{
		name:  "setext header: plain text",
		input: "alpha\n---",
	},
	{
		name:  "setext header: inline emphasis",
		input: "*alpha*\n---",
	},

	// Block quote lowering: recursive lowering through nested children.
	{
		name:  "block quote: two paragraphs",
		input: "> alpha\n>\n> beta",
	},
	{
		name:  "block quote: nested block quote",
		input: "> outer\n> > inner",
	},
	{
		name:  "block quote: contains list",
		input: "> - alpha\n> - beta",
	},
	{
		name:  "block quote: contains code block",
		input: ">     alpha",
	},

	// List lowering: tight/loose metadata and nested child lowering.
	{
		name:  "ordered list: start number preserved",
		input: "3. alpha\n4. beta",
	},
	{
		name:  "ordered list: paren delimiter still lowers as ordered list",
		input: "1) alpha\n2) beta",
	},
	{
		name:  "unordered list: loose list preserves paragraph children",
		input: "- alpha\n\n- beta",
	},
	{
		name:  "unordered list: item with two paragraphs",
		input: "- alpha\n\n  beta",
	},
	{
		name:  "unordered list: nested unordered list",
		input: "- alpha\n  - beta",
	},
	{
		name:  "unordered list: nested ordered list",
		input: "- alpha\n  1. beta",
	},
	{
		name:  "list item: indented code block child",
		input: "- alpha\n\n      beta",
	},
	{
		name:  "list item: block quote child",
		input: "- alpha\n  > beta",
	},

	// Indented code block lowering: indentation stripping and newline assembly.
	{
		name:  "indented code block: two lines preserves internal newline",
		input: "    alpha\n    beta",
	},
	{
		name:  "indented code block: blank line inside payload",
		input: "    alpha\n\n    beta",
	},
	{
		name:  "indented code block: spaces beyond required indent are preserved",
		input: "      alpha",
	},
	{
		name:  "indented code block: tab indentation participates in trim",
		input: "\talpha",
	},
	{
		name:  "indented code block: mixed space and tab indentation",
		input: "  \talpha",
	},

	// Fenced code block lowering: payload reconstruction and info-string handling.
	{
		name:  "fenced code block: two payload lines preserves internal newline",
		input: "```\nalpha\nbeta\n```",
	},
	{
		name:  "fenced code block: blank line inside payload",
		input: "```\nalpha\n\nbeta\n```",
	},
	{
		name:  "fenced code block: indented opener strips up to opener indent",
		input: "  ```\n  alpha\n  beta\n  ```",
	},
	{
		name:  "fenced code block: info string language token only",
		input: "```go\nalpha\n```",
	},
	{
		name:  "fenced code block: info string with trailing words extracts first token only",
		input: "```go linenos\nalpha\n```",
	},
	{
		name:  "fenced code block: tilde fence with info string",
		input: "~~~text\nalpha\n~~~",
	},
	{
		name:  "fenced code block: payload preserves leading spaces after indent stripping",
		input: "```\n  alpha\n```",
	},

	// HTML block lowering: raw payload assembly and newline insertion.
	{
		name:  "html block: multi-line comment preserves internal newline",
		input: "<!--\nalpha\n-->",
	},
	{
		name:  "html block: named tag block preserves multiple raw lines",
		input: "<div>\nalpha\n</div>",
	},
	{
		name:  "html block: processing instruction",
		input: "<?php\necho $a;\n?>",
	},
	{
		name:  "html block: cdata section",
		input: "<![CDATA[\nalpha\n]]>",
	},
	{
		name:  "html block: declaration",
		input: "<!DOCTYPE html>",
	},

	// Mixed document cases: good end-to-end coverage for recursive lowering.
	{
		name:  "document: header list code block paragraph",
		input: "# alpha\n\n- beta\n- gamma\n\n    delta\n\nepsilon",
	},
	{
		name:  "document: html block followed by paragraph",
		input: "<div>\nalpha\n</div>\n\nbeta",
	},
	{
		name:  "document: block quote containing list containing paragraph",
		input: "> - alpha\n>   beta",
	},
}

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
				tk.ASTPara(tk.ASTText()),
			),
			wantErr: nil,
		},

		// Headers
		{
			name:  "header with normal text",
			input: "# header",
			want: tk.ASTDoc(
				tk.ASTHeader(1, tk.ASTText()),
			),
			wantErr: nil,
		},
		{
			name:  "header with emphasis",
			input: "# *header*",
			want: tk.ASTDoc(
				tk.ASTHeader(1, tk.ASTEm(tk.ASTText())),
			),
			wantErr: nil,
		},
		{
			name:  "header and paragraph",
			input: "# header\n\nparagraph",
			want: tk.ASTDoc(
				tk.ASTHeader(1, tk.ASTText()),
				tk.ASTPara(tk.ASTText()),
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
				tk.ASTHTMLBlock(tk.ASTRawText()),
			),
			wantErr: nil,
		},

		// Containers
		{
			name:  "block quote: plain text",
			input: "> quote",
			want: tk.ASTDoc(
				tk.ASTBlockQuote(
					tk.ASTPara(tk.ASTText()),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ul: two items",
			input: "- a\n- b",
			want: tk.ASTDoc(
				tk.ASTUnorderedList(
					true,
					tk.ASTListItem(
						tk.ASTPara(tk.ASTText()),
					),
					tk.ASTListItem(
						tk.ASTPara(tk.ASTText()),
					),
				),
			),
			wantErr: nil,
		},
		{
			name:  "ol: two items",
			input: "1. a\n2. b",
			want: tk.ASTDoc(
				tk.ASTOrderedList(
					true,
					1,
					tk.ASTListItem(
						tk.ASTPara(tk.ASTText()),
					),
					tk.ASTListItem(
						tk.ASTPara(tk.ASTText()),
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
				tk.ASTIndentedCodeBlock(tk.ASTText()),
			),
			wantErr: nil,
		},
		{
			name:  "fenced code block",
			input: "```\ncode\n```",
			want: tk.ASTDoc(
				tk.ASTFencedCodeBlock(tk.ASTText()),
			),
			wantErr: nil,
		},

		// Basic inline lowering through blocks
		{
			name:  "emphasis",
			input: "*abc*",
			want: tk.ASTDoc(
				tk.ASTPara(tk.ASTEm(tk.ASTText())),
			),
			wantErr: nil,
		},
		{
			name:  "strong",
			input: "**abc**",
			want: tk.ASTDoc(
				tk.ASTPara(tk.ASTStrong(tk.ASTText())),
			),
			wantErr: nil,
		},
		{
			name:  "code span",
			input: "`abc`",
			want: tk.ASTDoc(
				tk.ASTPara(tk.ASTCodeSpan()),
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
