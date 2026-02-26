package block

import (
	"bytes"
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

const MaxValidIndentation = 3

type BuildRule interface {
	Apply(c *Cursor) (ir.Block, bool, error)
}

type HeaderRule struct{}

func (r HeaderRule) Apply(c *Cursor) (ir.Block, bool, error) {
	line, ok := c.Peek()
	if !ok {
		return nil, false, nil
	}
	if line.IsBlankLine(c.Source) {
		return nil, false, nil
	}

	// count the leading spaces, reject if more than 3
	offset := line.BlockIndentSpaces(c.Source)
	if offset > MaxValidIndentation {
		return nil, false, nil
	}

	s := c.Source.Slice(line.Span)
	pos := offset
	level := 0

	// validate the marker
	if pos >= len(s) || s[pos] != '#' {
		return nil, false, nil
	}

	// count the marker run, reject if more than 6
	for pos < len(s) && s[pos] == '#' {
		pos++
		level++

		if level == 7 {
			return nil, false, nil
		}
	}

	// validate the delimiter
	if pos >= len(s) || (s[pos] != ' ' && s[pos] != '\t') {
		return nil, false, nil
	}

	// consume the delimiter
	for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
		pos++
	}

	// trim trailing spaces and tabs
	content := strings.TrimRight(s[pos:], " \t")
	contentStart := line.Span.Start + source.BytePos(pos)
	contentEnd := contentStart + source.BytePos(len(content))

	span := line.Span

	line, ok = c.Next()
	if !ok {
		return nil, false, nil
	}

	applied := ir.Header{
		Level: level,
		Span:  span,
		ContentSpan: source.ByteSpan{
			Start: contentStart,
			End:   contentEnd,
		},
	}

	return applied, true, nil
}

type ThematicBreakRule struct{}

func (r ThematicBreakRule) Apply(c *Cursor) (ir.Block, bool, error) {
	line, ok := c.Peek()
	if !ok {
		return nil, false, nil
	}
	if line.IsBlankLine(c.Source) {
		return nil, false, nil
	}

	// count the leading spaces, reject if more than 3
	offset := line.BlockIndentSpaces(c.Source)
	if offset > MaxValidIndentation {
		return nil, false, nil
	}

	s := c.Source.Slice(line.Span)
	pos := offset

	// validate the first marker character
	validMarkers := []byte{'-', '*', '_'}
	if pos >= len(s) || !bytes.Contains(validMarkers, []byte{s[pos]}) {
		return nil, false, nil
	}

	markerChar := s[pos]
	markerCount := 0

	// count the marker run, skipping whitespace and rejecting mixed markers
	for pos < len(s) {
		b := s[pos]
		if b == ' ' || b == '\t' {
			pos++
			continue
		}
		if b != markerChar {
			return nil, false, nil
		}

		pos++
		markerCount++
	}

	span := line.Span

	line, ok = c.Next()
	if !ok {
		return nil, false, nil
	}

	applied := ir.ThematicBreak{
		Span: span,
	}

	return applied, true, nil

}

type ParagraphRuleMarker interface {
	isParagraphRule()
}

type ParagraphRule struct{}

func (ParagraphRule) isParagraphRule() {}

func (r ParagraphRule) Apply(c *Cursor) (ir.Block, bool, error) {
	line, ok := c.Peek()
	if !ok {
		return nil, false, nil
	}
	if line.IsBlankLine(c.Source) {
		return nil, false, nil
	}

	var spans []source.ByteSpan

	line, ok = c.Next()
	if !ok {
		return nil, false, nil
	}

	spans = append(spans, line.Span)

	for {
		line, ok := c.Peek()
		if !ok {
			break
		}
		if line.IsBlankLine(c.Source) {
			break
		}

		startsBlock, err := c.StartsNonParagraphBlock()
		if err != nil {
			return nil, false, err
		}
		if startsBlock {
			break
		}

		line, ok = c.Next()
		if !ok {
			return nil, false, nil
		}

		spans = append(spans, line.Span)
	}

	span := source.ByteSpan{
		Start: spans[0].Start,
		End:   spans[len(spans)-1].End,
	}

	applied := ir.Paragraph{
		Lines: spans,
		Span:  span,
	}

	return applied, true, nil
}
