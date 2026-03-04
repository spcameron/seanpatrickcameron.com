package ir

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

type UnorderedList struct {
	Span  source.ByteSpan
	Items []ListItem
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

type Header struct {
	Span        source.ByteSpan
	ContentSpan source.ByteSpan
	Level       int
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
	return "[ThematicBreak]"
}

type Paragraph struct {
	Span  source.ByteSpan
	Lines []source.ByteSpan
}

func (Paragraph) isBlock() {}

func (p Paragraph) String() string {
	return fmt.Sprintf("[Paragraph] (Lines = %d)", len(p.Lines))
}
