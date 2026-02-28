package ir

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type Block interface {
	isBlock()
}

type BlockQuote struct {
	Children []Block
	Span     source.ByteSpan
}

func (BlockQuote) isBlock() {}

func (bq BlockQuote) String() string {
	return fmt.Sprintf("[BlockQuote] (Children = %d)", len(bq.Children))
}

type Header struct {
	Level       int
	Span        source.ByteSpan
	ContentSpan source.ByteSpan
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
	return fmt.Sprint("[ThematicBreak]")
}

type Paragraph struct {
	Span  source.ByteSpan
	Lines []source.ByteSpan
}

func (Paragraph) isBlock() {}

func (p Paragraph) String() string {
	return fmt.Sprintf("[Paragraph] (Lines = %d)", len(p.Lines))
}
