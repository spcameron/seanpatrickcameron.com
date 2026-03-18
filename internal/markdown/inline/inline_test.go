package inline

import (
	"fmt"
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type CursorSummary struct {
	WorkingItems []WorkingItemSummary
	Delimiters   []DelimiterSummary
	Brackets     []BracketSummary
}

func (s CursorSummary) String() string {
	var b strings.Builder

	b.WriteString("Working:\n")
	for i, w := range s.WorkingItems {
		fmt.Fprintf(&b, "  [%d] %s\n", i, w.String())
	}

	b.WriteString("Delimiters:\n")
	for i, d := range s.Delimiters {
		fmt.Fprintf(&b, "  [%d] %s\n", i, d.String())
	}

	b.WriteString("Brackets:\n")
	for i, br := range s.Brackets {
		fmt.Fprintf(&b, "  [%d] %s\n", i, br.String())
	}

	return b.String()
}

type WorkingItemSummary struct {
	Kind      string
	Lexeme    string
	Delimiter byte
	Token     string
	Node      *InlineSummary
}

func (w WorkingItemSummary) String() string {
	switch w.Kind {
	case "text":
		return fmt.Sprintf("text(%q)", w.Lexeme)

	case "delimiter":
		return fmt.Sprintf("delimiter(%q)", w.Lexeme)

	case "token":
		return fmt.Sprintf("%s(%q)", w.Token, w.Lexeme)

	case "node":
		if w.Node != nil {
			return fmt.Sprintf("node(%s, %q)", w.Node.String(), w.Lexeme)
		}
		return fmt.Sprintf("node(%q)", w.Lexeme)

	case "consumed":
		return "consumed"

	default:
		return fmt.Sprintf("unknown(%s)", w.Kind)
	}
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

type DelimiterSummary struct {
	Lexeme       string
	Delimiter    byte
	OriginalRun  int
	RemainingRun int
	CanOpen      bool
	CanClose     bool
	ItemIndex    int
}

func (d DelimiterSummary) String() string {
	return fmt.Sprintf(
		"%q run=%d rem=%d open=%t close=%t item=%d",
		d.Lexeme,
		d.OriginalRun,
		d.RemainingRun,
		d.CanOpen,
		d.CanClose,
		d.ItemIndex,
	)
}

type BracketSummary struct {
	Lexeme    string
	ItemIndex int
	Active    bool
}

func (b BracketSummary) String() string {
	return fmt.Sprintf(
		"%q active=%t item=%d",
		b.Lexeme,
		b.Active,
		b.ItemIndex,
	)
}

type EventSummary struct {
	Kind      EventKind
	Lexeme    string
	Delimiter byte
	RunLength int
}

func (es EventSummary) String() string {
	switch es.Kind {
	case EventText:
		return fmt.Sprintf("text(%q)", es.Lexeme)

	case EventDelimiterRun:
		return fmt.Sprintf("delimiter(%q)", es.Lexeme)

	case EventOpenBracket:
		return `open_bracket("[")`

	case EventCloseBracket:
		return `close_bracket("]")`

	case EventOpenParen:
		return `open_paren("(")`

	case EventCloseParen:
		return `close_paren(")")`

	case EventIllegalNewline:
		return `illegal_newline("\n")`

	default:
		return fmt.Sprintf("unknown_event(%s, %q)", es.Kind, es.Lexeme)
	}
}

func summarizeCursor(c *Cursor) CursorSummary {
	summary := CursorSummary{
		WorkingItems: summarizeWorkingItems(c.Source, c.WorkingItems),
		Delimiters:   summarizeDelimiters(c.Source, c.DelimiterRecords),
		Brackets:     summarizeBrackets(c.Source, c.BracketRecords),
	}

	return summary
}

func summarizeWorkingItems(src *source.Source, items []WorkingItem) []WorkingItemSummary {
	summary := make([]WorkingItemSummary, 0, len(items))

	for _, item := range items {
		switch item := item.(type) {
		case *TextItem:
			summary = append(summary, WorkingItemSummary{
				Kind:   "text",
				Lexeme: src.Slice(item.Span),
			})

		case *DelimiterItem:
			summary = append(summary, WorkingItemSummary{
				Kind:      "delimiter",
				Lexeme:    src.Slice(item.Span),
				Delimiter: item.Delimiter,
			})

		case *TokenItem:
			summary = append(summary, WorkingItemSummary{
				Kind:   "token",
				Token:  item.Kind.String(),
				Lexeme: src.Slice(item.Span),
			})

		case *NodeItem:
			summary = append(summary, WorkingItemSummary{
				Kind:   "node",
				Lexeme: src.Slice(item.Span),
				Node:   summarizeInline(src, item.Node),
			})

		case *ConsumedItem:
			summary = append(summary, WorkingItemSummary{
				Kind: "consumed",
			})

		default:
			panic("unknown item type")
		}
	}

	return summary
}

func summarizeInline(src *source.Source, inl ast.Inline) *InlineSummary {
	switch n := inl.(type) {
	case ast.Link:
		return &InlineSummary{
			Kind:     "link",
			Lexeme:   src.Slice(n.Span),
			Children: summarizeInlines(src, n.Children),
		}

	case ast.Em:
		return &InlineSummary{
			Kind:     "emphasis",
			Lexeme:   src.Slice(n.Span),
			Children: summarizeInlines(src, n.Children),
		}

	case ast.Strong:
		return &InlineSummary{
			Kind:     "strong",
			Lexeme:   src.Slice(n.Span),
			Children: summarizeInlines(src, n.Children),
		}

	case ast.Text:
		return &InlineSummary{
			Kind:   "text",
			Lexeme: src.Slice(n.Span),
		}

	case ast.RawText:
		return &InlineSummary{
			Kind:   "raw_text",
			Lexeme: src.Slice(n.Span),
		}

	case ast.HardBreak:
		return &InlineSummary{
			Kind:   "hard_break",
			Lexeme: src.Slice(n.Span),
		}

	case ast.SoftBreak:
		return &InlineSummary{
			Kind:   "soft_break",
			Lexeme: src.Slice(n.Span),
		}

	case ast.Newline:
		return &InlineSummary{
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
		out = append(out, *summarizeInline(src, inl))
	}
	return out
}

func summarizeDelimiters(src *source.Source, delims []*DelimiterRecord) []DelimiterSummary {
	summary := make([]DelimiterSummary, 0, len(delims))

	for _, delim := range delims {
		s := src.Slice(delim.OriginalSpan)

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

func summarizeBrackets(src *source.Source, brackets []*BracketRecord) []BracketSummary {
	summary := make([]BracketSummary, 0, len(brackets))

	for _, br := range brackets {
		summary = append(summary, BracketSummary{
			Lexeme:    src.Slice(br.Span),
			ItemIndex: br.ItemIndex,
			Active:    br.Active,
		})
	}

	return summary
}
