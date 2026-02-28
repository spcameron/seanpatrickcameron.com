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

type Paragraph struct {
	Span    source.ByteSpan
	Inlines []Inline
}

func (Paragraph) isBlock() {}

func (p Paragraph) String() string {
	return fmt.Sprintf("[Paragraph] (Lines = %d)", len(p.Inlines))
}
