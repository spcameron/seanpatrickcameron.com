package inline

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type Cursor struct {
	Source *source.Source
	Span   source.ByteSpan
	Events []Event
	Index  int

	WorkingItems     []WorkingItem
	DelimiterRecords []*DelimiterRecord
}

func NewCursor(src *source.Source, span source.ByteSpan, events []Event) *Cursor {
	return &Cursor{
		Source:           src,
		Span:             span,
		Events:           events,
		Index:            0,
		WorkingItems:     []WorkingItem{},
		DelimiterRecords: []*DelimiterRecord{},
	}
}

func (c *Cursor) Peek() (Event, bool) {
	if c.EOF() {
		return Event{}, false
	}

	return c.Events[c.Index], true
}

func (c *Cursor) Next() (Event, bool) {
	if c.EOF() {
		return Event{}, false
	}

	out := c.Events[c.Index]
	c.Index++
	return out, true
}

func (c *Cursor) Mark() int {
	return c.Index
}

func (c *Cursor) Reset(i int) {
	c.Index = i
}

func (c *Cursor) EOF() bool {
	return c.Index >= len(c.Events)
}

func (c *Cursor) Gather() error {
	for {
		if c.EOF() {
			break
		}

		ev, ok := c.Next()
		if !ok {
			break
		}

		switch ev.Kind {
		case EventText:
			item := &TextItem{
				Span: ev.Span,
			}

			c.WorkingItems = append(c.WorkingItems, item)

		case EventDelimiterRun:
			item := &DelimiterItem{
				Span:      ev.Span,
				Delimiter: ev.Delimiter,
			}

			c.WorkingItems = append(c.WorkingItems, item)
			idx := len(c.WorkingItems) - 1

			before, beforeOK := c.runeBefore(ev.Span)
			after, afterOK := c.runeAfter(ev.Span)

			canOpen := leftFlanking(before, beforeOK, after, afterOK)
			canClose := rightFlanking(before, beforeOK, after, afterOK)

			record := &DelimiterRecord{
				OriginalSpan: ev.Span,
				LiveSpan:     ev.Span,
				Delimiter:    ev.Delimiter,
				OriginalRun:  ev.RunLength,
				RemainingRun: ev.RunLength,
				CanOpen:      canOpen,
				CanClose:     canClose,
				ItemIndex:    idx,
			}

			c.DelimiterRecords = append(c.DelimiterRecords, record)

		case EventIllegalNewline:
			panic("illegal newline encountered during inline gather")

		default:
			panic("unhandled inline event kind in gather")
		}
	}

	return nil
}

func (c *Cursor) Resolve() error {
	for closerIdx := 0; closerIdx < len(c.DelimiterRecords); closerIdx++ {
		closerRecord := c.DelimiterRecords[closerIdx]

		for closerRecord.CanClose && closerRecord.RemainingRun > 0 {
			openerIdx, ok := c.findOpenerForCloser(closerIdx)
			if !ok {
				break
			}

			use := 1
			if c.DelimiterRecords[openerIdx].RemainingRun >= 2 &&
				c.DelimiterRecords[closerIdx].RemainingRun >= 2 {
				use = 2
			}

			c.resolvePair(openerIdx, closerIdx, use)
		}
	}

	return nil
}

func (c *Cursor) Finalize() ([]ast.Inline, error) {
	inlines := make([]ast.Inline, 0, len(c.WorkingItems))

	for i, item := range c.WorkingItems {
		switch v := item.(type) {
		case *TextItem:
			inlines = append(inlines, ast.Text{
				Span: v.Span,
			})

		case *DelimiterItem:
			recIdx, ok := c.delimiterRecordForItem(i)
			if !ok {
				return []ast.Inline{}, fmt.Errorf("no matching delimiter record found for working item at index %d", i)
			}

			rec := c.DelimiterRecords[recIdx]
			if rec.RemainingRun == 0 {
				continue
			}

			inlines = append(inlines, ast.Text{
				Span: rec.LiveSpan,
			})

		case *NodeItem:
			inlines = append(inlines, v.Node)

		case *ConsumedItem:
			continue

		default:
			return []ast.Inline{}, fmt.Errorf("unknown working item type %T", item)
		}
	}

	return inlines, nil
}

func (c *Cursor) findOpenerForCloser(closerIdx int) (int, bool) {
	closer := c.DelimiterRecords[closerIdx]

	openerIdx := closerIdx - 1
	for openerIdx >= 0 {
		d := c.DelimiterRecords[openerIdx]
		if d.Delimiter == closer.Delimiter && d.CanOpen && d.RemainingRun > 0 {
			return openerIdx, true
		}
		openerIdx--
	}

	return -1, false
}

func (c *Cursor) resolvePair(openerIdx, closerIdx, use int) {
	// retrieve the delimiter records
	openerRecord := c.DelimiterRecords[openerIdx]
	closerRecord := c.DelimiterRecords[closerIdx]

	// retrieve the corresponding working item indexes
	openItemIdx := openerRecord.ItemIndex
	closeItemIdx := closerRecord.ItemIndex

	// determine the node span based on the consumed spans
	nodeSpan := source.ByteSpan{
		Start: openerRecord.LiveSpan.End - source.BytePos(use),
		End:   closerRecord.LiveSpan.Start + source.BytePos(use),
	}

	// build the children ast inline nodes and their cumulative span
	children, childSpan := c.buildChildren(openItemIdx+1, closeItemIdx)

	// reject empty children
	if len(children) == 0 {
		panic("resolvePair: cannot resolve pair with no child content")
	}

	// build the inline node with the content-only span
	var node ast.Inline
	switch use {
	case 1:
		node = ast.Em{
			Span:     childSpan,
			Children: children,
		}
	case 2:
		node = ast.Strong{
			Span:     childSpan,
			Children: children,
		}
	default:
		panic("resolvePair: unsupported delimiter consumption size")
	}

	// consume from the matched delimiter records
	openerRecord.RemainingRun -= use
	closerRecord.RemainingRun -= use

	// trim live spans according to their roles
	// opener consumes from the right
	openerRecord.LiveSpan.End -= source.BytePos(use)

	// closer consumes from the left
	closerRecord.LiveSpan.Start += source.BytePos(use)

	// rewrite working items without changing indices
	//
	// anchor the resolved node at the first interior slot
	anchor := openItemIdx + 1
	c.WorkingItems[anchor] = &NodeItem{
		Span: nodeSpan,
		Node: node,
	}

	// any other interior slots absorbed by the node are marked consumed
	for i := anchor + 1; i < closeItemIdx; i++ {
		c.WorkingItems[i] = &ConsumedItem{}
	}

	// if either delimiter run has been fully exhausted, it is also consumed
	// otherwise it remains a live DelimiterItem with a narrower LiveSpan
	if openerRecord.RemainingRun == 0 {
		c.WorkingItems[openItemIdx] = &ConsumedItem{}
	}

	if closerRecord.RemainingRun == 0 {
		c.WorkingItems[closeItemIdx] = &ConsumedItem{}
	}
}

func (c *Cursor) buildChildren(start, end int) ([]ast.Inline, source.ByteSpan) {
	children := make([]ast.Inline, 0, end-start)

	for i := start; i < end; i++ {
		switch v := c.WorkingItems[i].(type) {
		case *TextItem:
			inl := ast.Text{
				Span: v.Span,
			}
			children = append(children, inl)

		case *DelimiterItem:
			idx, ok := c.delimiterRecordForItem(i)
			if !ok {
				panic(fmt.Sprintf("no matching delimiter record found for the working item at index %d", i))
			}

			rec := c.DelimiterRecords[idx]
			if rec.RemainingRun > 0 {
				inl := ast.Text{
					Span: rec.LiveSpan,
				}
				children = append(children, inl)
			}

		case *NodeItem:
			inl := v.Node
			children = append(children, inl)

		case *ConsumedItem:
			continue

		default:
			panic("unknown working item encountered")
		}
	}

	if len(children) == 0 {
		return children, source.ByteSpan{}
	}

	first, ok := inlineSpan(children[0])
	if !ok {
		panic("could not determine span of first child inline")
	}

	last, ok := inlineSpan(children[len(children)-1])
	if !ok {
		panic("could not determine span of last child inline")
	}

	span := source.ByteSpan{
		Start: first.Start,
		End:   last.End,
	}

	return children, span
}

func (c *Cursor) delimiterRecordForItem(index int) (int, bool) {
	for i := 0; i < len(c.DelimiterRecords); i++ {
		if c.DelimiterRecords[i].ItemIndex == index {
			return i, true
		}
	}

	return -1, false
}

func (c *Cursor) runeBefore(delimSpan source.ByteSpan) (rune, bool) {
	if delimSpan.Start == c.Span.Start {
		return 0, false
	}

	leftWindow := source.ByteSpan{
		Start: c.Span.Start,
		End:   delimSpan.Start,
	}

	s := c.Source.Slice(leftWindow)
	r, width := utf8.DecodeLastRuneInString(s)
	if width == 0 {
		return 0, false
	}

	return r, true
}

func (c *Cursor) runeAfter(delimSpan source.ByteSpan) (rune, bool) {
	if delimSpan.End == c.Span.End {
		return 0, false
	}

	rightWindow := source.ByteSpan{
		Start: delimSpan.End,
		End:   c.Span.End,
	}

	s := c.Source.Slice(rightWindow)
	r, width := utf8.DecodeRuneInString(s)
	if width == 0 {
		return 0, false
	}

	return r, true
}

func leftFlanking(before rune, beforeOK bool, after rune, afterOK bool) bool {
	if !afterOK {
		return false
	}
	if isWhitespace(after) {
		return false
	}
	if !isPunctuation(after) {
		return true
	}
	if !beforeOK {
		return true
	}
	if isWhitespace(before) {
		return true
	}
	if isPunctuation(before) {
		return true
	}
	return false
}

func rightFlanking(before rune, beforeOK bool, after rune, afterOK bool) bool {
	if !beforeOK {
		return false
	}
	if isWhitespace(before) {
		return false
	}
	if !isPunctuation(before) {
		return true
	}
	if !afterOK {
		return true
	}
	if isWhitespace(after) {
		return true
	}
	if isPunctuation(after) {
		return true
	}
	return false
}

func isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

func isPunctuation(r rune) bool {
	return unicode.IsPunct(r)
}

func inlineSpan(inl ast.Inline) (source.ByteSpan, bool) {
	switch n := inl.(type) {
	case ast.Text:
		return n.Span, true
	case ast.RawText:
		return n.Span, true
	case ast.HardBreak:
		return n.Span, true
	case ast.SoftBreak:
		return n.Span, true
	case ast.Newline:
		return n.Span, true
	case ast.Em:
		return n.Span, true
	case ast.Strong:
		return n.Span, true
	default:
		return source.ByteSpan{}, false
	}
}
