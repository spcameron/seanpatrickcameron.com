package ast

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type Block interface {
	isBlock()
}

type BlockQuote struct {
	Span     source.ByteSpan
	Children []Block
}

func (BlockQuote) isBlock() {}

func (bq BlockQuote) String() string {
	return fmt.Sprintf("[BlockQuote] (Children = %d)", len(bq.Children))
}

type Header struct {
	Span    source.ByteSpan
	Level   int
	Inlines []Inline
}

func (Header) isBlock() {}

func (h Header) String() string {
	return fmt.Sprintf("[Header] (Level = %d)", h.Level)
}

type ThematicBreak struct {
	Span source.ByteSpan
}

func (ThematicBreak) isBlock() {}

func (tb ThematicBreak) String() string {
	return "[Thematic Break]"
}

type OrderedList struct {
	Span  source.ByteSpan
	Items []ListItem
	Tight bool
	Start int
}

func (OrderedList) isBlock() {}

func (ol OrderedList) String() string {
	return fmt.Sprintf("[OrderedList] (Items = %d)", len(ol.Items))
}

type UnorderedList struct {
	Span  source.ByteSpan
	Items []ListItem
	Tight bool
}

func (UnorderedList) isBlock() {}

func (ul UnorderedList) String() string {
	return fmt.Sprintf("[UnorderedList] (Items = %d)", len(ul.Items))
}

type ListItem struct {
	Span     source.ByteSpan
	Children []Block
}

func (ListItem) isBlock() {}

func (li ListItem) String() string {
	return fmt.Sprintf("[ListItem] (Children = %d)", len(li.Children))
}

type CodeBlockKind int

func (k CodeBlockKind) String() string {
	switch k {
	case Indented:
		return "Indented"
	case Fenced:
		return "Fenced"
	default:
		return fmt.Sprintf("Unrecognized CodeBlockKind %d", k)
	}
}

const (
	_ CodeBlockKind = iota
	Indented
	Fenced
)

type CodeBlock struct {
	Span              source.ByteSpan
	Kind              CodeBlockKind
	LanguageTokenSpan source.ByteSpan
	Payload           []Inline
}

func (CodeBlock) isBlock() {}

func (cb CodeBlock) String() string {
	return fmt.Sprintf("[%sCodeBlock] (Lines = %d)", cb.Kind, len(cb.Payload))
}

type HTMLBlock struct {
	Span    source.ByteSpan
	Payload []Inline
}

func (HTMLBlock) isBlock() {}

func (hb HTMLBlock) String() string {
	return fmt.Sprintf("[HTMLBlock] (Lines = %d)", len(hb.Payload))
}

type Paragraph struct {
	Span    source.ByteSpan
	Inlines []Inline
}

func (Paragraph) isBlock() {}

func (p Paragraph) String() string {
	return fmt.Sprintf("[Paragraph] (Lines = %d)", len(p.Inlines))
}
