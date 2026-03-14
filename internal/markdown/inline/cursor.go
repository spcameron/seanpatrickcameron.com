package inline

import (
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
				Span:         ev.Span,
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
	return nil
}

func (c *Cursor) Finalize() ([]ast.Inline, error) {
	return []ast.Inline{}, nil
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
