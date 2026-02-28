package block

import (
	"bytes"
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

// TODO: extract out duplication into helpers

const MaxValidIndentation = 3

type BuildRule interface {
	Apply(c *Cursor) (ir.Block, bool, error)
}

type BlockQuoteRule struct{}

func (r BlockQuoteRule) Apply(c *Cursor) (ir.Block, bool, error) {
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

	derived := true
	start := line.Span.Start
	if start == 0 || c.Source.Raw[start-1] == '\n' {
		derived = false
	}

	if derived && offset > 0 {
		return nil, false, nil
	}

	s := c.Source.Slice(line.Span)
	pos := offset

	// validate the marker
	if pos >= len(s) || s[pos] != '>' {
		return nil, false, nil
	}

	var spans []source.ByteSpan
	var trimmedLines []Line

	line, ok = c.Next()
	if !ok {
		return nil, false, nil
	}

	pos++

	// consume a single, optional delimiter
	if pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
		pos++
	}

	// define a new line with a trimmed span (no indent or marker run)
	tl := Line{
		source.ByteSpan{
			Start: line.Span.Start + source.BytePos(pos),
			End:   line.Span.End,
		},
	}

	spans = append(spans, line.Span)
	trimmedLines = append(trimmedLines, tl)

	for {
		line, ok := c.Peek()
		if !ok {
			break
		}
		if line.IsBlankLine(c.Source) {
			break
		}

		offset := line.BlockIndentSpaces(c.Source)
		if offset > MaxValidIndentation {
			break
		}

		s = c.Source.Slice(line.Span)
		pos = offset

		if pos >= len(s) || s[pos] != '>' {
			break
		}

		line, ok = c.Next()
		if !ok {
			return nil, false, nil
		}

		pos++

		if pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
			pos++
		}

		tl := Line{
			source.ByteSpan{
				Start: line.Span.Start + source.BytePos(pos),
				End:   line.Span.End,
			},
		}

		spans = append(spans, line.Span)
		trimmedLines = append(trimmedLines, tl)
	}

	// call recursive build with trimmed lines
	innerDoc, err := Build(c.Source, trimmedLines)
	if err != nil {
		return nil, false, err
	}

	children := innerDoc.Blocks

	span := source.ByteSpan{
		Start: spans[0].Start,
		End:   spans[len(spans)-1].End,
	}

	applied := ir.BlockQuote{
		Children: children,
		Span:     span,
	}

	return applied, true, nil
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

	// TODO: consider trimming suffix '#' characters
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
