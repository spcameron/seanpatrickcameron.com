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

type IndentedCodeBlock struct {
	Span  source.ByteSpan
	Lines []source.ByteSpan
}

func (IndentedCodeBlock) isBlock() {}

func (icb IndentedCodeBlock) String() string {
	return fmt.Sprintf("[IndentedCodeBlock] (Lines = %d)", len(icb.Lines))
}

type FencedCodeBlock struct {
	Span           source.ByteSpan
	OpenIndentCols int
	InfoStringSpan source.ByteSpan
	Lines          []source.ByteSpan
}

func (FencedCodeBlock) isBlock() {}

func (fcb FencedCodeBlock) String() string {
	return fmt.Sprintf("[FencedCodeBlock] (Lines = %d)", len(fcb.Lines))
}

type HTMLBlock struct {
	Span  source.ByteSpan
	Lines []source.ByteSpan
}

func (HTMLBlock) isBlock() {}

func (hb HTMLBlock) String() string {
	return fmt.Sprintf("[HTMLBlock] (Lines = %d)", len(hb.Lines))
}

type Paragraph struct {
	Span  source.ByteSpan
	Lines []source.ByteSpan
}

func (Paragraph) isBlock() {}

func (p Paragraph) String() string {
	return fmt.Sprintf("[Paragraph] (Lines = %d)", len(p.Lines))
}
