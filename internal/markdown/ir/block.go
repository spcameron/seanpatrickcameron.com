package ir

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

// Block is the marker interface implemented by all IR block nodes.
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

// Header represents a parsed header before inline lowering.
//
// ContentSpan covers the header content as a whole, while ContentLines
// preserves the source lines that contribute that content.
type Header struct {
	Span         source.ByteSpan
	ContentSpan  source.ByteSpan
	ContentLines []source.ByteSpan
	Level        int
}

func (Header) isBlock() {}

func (h Header) String() string {
	if len(h.ContentLines) > 0 {
		return fmt.Sprintf("[Header] (Level = %d, Lines = %d)", h.Level, len(h.ContentLines))
	}
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

// ListItem represents a list item within an ordered or unordered list.
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

// FencedCodeBlock represents a fenced code block in parse-facing form.
//
// OpenIndentCols records the indentation of the opening fence, and
// InfoStringSpan identifies the raw info string when present.
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
