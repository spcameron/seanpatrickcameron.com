package inline

import (
	"fmt"
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type TokenSummary struct {
	Kind   TokenKind
	Lexeme string
}

func (ts TokenSummary) String() string {
	switch ts.Kind {
	case TokenText:
		return fmt.Sprintf("text(%q)", ts.Lexeme)

	case TokenStarDelimiter:
		return fmt.Sprintf("star(%q)", ts.Lexeme)

	case TokenUnderscoreDelimiter:
		return fmt.Sprintf("underscore(%q)", ts.Lexeme)

	case TokenBacktick:
		return fmt.Sprintf("backtick(%q)", ts.Lexeme)

	case TokenOpenBracket:
		return `open_bracket("[")`

	case TokenCloseBracket:
		return `close_bracket("]")`

	case TokenOpenParen:
		return `open_paren("(")`

	case TokenCloseParen:
		return `close_paren(")")`

	case TokenOpenAngle:
		return `open_angle("<")`

	case TokenCloseAngle:
		return `close_angle(">")`

	case TokenBang:
		return `bang("!")`

	case TokenImageOpenBracket:
		return `image_open_bracket("![")`

	case TokenBackslash:
		return `backslash("\")`

	case TokenEOF:
		return "EOF"

	default:
		return fmt.Sprintf("unknown_token(%d, %q)", ts.Kind, ts.Lexeme)
	}
}

func summarizeTokens(src *source.Source, tokens []Token) []TokenSummary {
	summary := make([]TokenSummary, 0, len(tokens))

	for _, t := range tokens {
		s := src.Slice(t.Span)

		summary = append(summary, TokenSummary{
			Kind:   t.Kind,
			Lexeme: s,
		})
	}

	return summary
}

type InlineSummary struct {
	Kind     string
	Lexeme   string
	Children []InlineSummary
}

func (s InlineSummary) String() string {
	switch s.Kind {
	case "text", "raw_text", "hard_break", "soft_break", "newline":
		return fmt.Sprintf("%s(%q)", s.Kind, s.Lexeme)

	default:
		if len(s.Children) == 0 {
			return s.Kind
		}

		var b strings.Builder
		b.WriteString(s.Kind)
		b.WriteString("(")
		for i, child := range s.Children {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(child.String())
		}
		b.WriteString(")")
		return b.String()
	}
}

func summarizeInline(src *source.Source, inl ast.Inline) InlineSummary {
	switch n := inl.(type) {
	case ast.CodeSpan:
		return InlineSummary{
			Kind:   "code_span",
			Lexeme: src.Slice(n.Span),
		}

	case ast.Link:
		return InlineSummary{
			Kind:     "link",
			Lexeme:   src.Slice(n.Span),
			Children: summarizeInlines(src, n.Children),
		}

	case ast.Image:
		return InlineSummary{
			Kind:     "image",
			Lexeme:   src.Slice(n.Span),
			Children: summarizeInlines(src, n.Children),
		}

	case ast.Emph:
		return InlineSummary{
			Kind:     "emphasis",
			Lexeme:   src.Slice(n.Span),
			Children: summarizeInlines(src, n.Children),
		}

	case ast.Strong:
		return InlineSummary{
			Kind:     "strong",
			Lexeme:   src.Slice(n.Span),
			Children: summarizeInlines(src, n.Children),
		}

	case ast.Text:
		return InlineSummary{
			Kind:   "text",
			Lexeme: src.Slice(n.Span),
		}

	case ast.RawText:
		return InlineSummary{
			Kind:   "raw_text",
			Lexeme: src.Slice(n.Span),
		}

	case ast.HardBreak:
		return InlineSummary{
			Kind:   "hard_break",
			Lexeme: src.Slice(n.Span),
		}

	case ast.SoftBreak:
		return InlineSummary{
			Kind:   "soft_break",
			Lexeme: src.Slice(n.Span),
		}

	case ast.Newline:
		return InlineSummary{
			Kind:   "newline",
			Lexeme: src.Slice(n.Span),
		}

	default:
		panic("unknown ast.Inline type")
	}
}

func summarizeInlines(src *source.Source, inlines []ast.Inline) []InlineSummary {
	out := make([]InlineSummary, 0, len(inlines))
	for _, inl := range inlines {
		out = append(out, summarizeInline(src, inl))
	}
	return out
}
