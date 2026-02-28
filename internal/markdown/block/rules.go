package block

import (
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

const MaxValidIndentation = 3

type BuildRule interface {
	Apply(c *Cursor) (ir.Block, bool, error)
}

type BlockQuoteRule struct{}

func (r BlockQuoteRule) Apply(c *Cursor) (ir.Block, bool, error) {
	var spans []source.ByteSpan
	var trimmedLines []Line

	// must consume at least one line to apply
	full, trimmed, ok, err := r.tryConsumeQuoteLine(c)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}

	spans = append(spans, full.Span)
	trimmedLines = append(trimmedLines, trimmed)

	// consume subsequent quote lines
	for {
		full, trimmed, ok, err := r.tryConsumeQuoteLine(c)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			break
		}

		spans = append(spans, full.Span)
		trimmedLines = append(trimmedLines, trimmed)
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

func (BlockQuoteRule) tryConsumeQuoteLine(c *Cursor) (Line, Line, bool, error) {
	line, ok := c.Peek()
	if !ok {
		return Line{}, Line{}, false, nil
	}

	// truly blank line terminates the quote run
	if line.IsBlankLine(c.Source) {
		return Line{}, Line{}, false, nil
	}

	// count the leading spaces, reject if more than 3
	offset := line.BlockIndentSpaces(c.Source)
	if offset > MaxValidIndentation {
		return Line{}, Line{}, false, nil
	}

	// derived line guard
	derived := !line.IsPhysicalLineStart(c.Source)
	if derived && offset > 0 {
		return Line{}, Line{}, false, nil
	}

	s := c.Source.Slice(line.Span)
	pos := offset

	// validate the marker
	if pos >= len(s) || s[pos] != '>' {
		return Line{}, Line{}, false, nil
	}

	// commit to consuming the line
	full, ok := c.Next()
	if !ok {
		return Line{}, Line{}, false, nil
	}

	// consume '>'
	pos++

	// consume a single, optional delimiter
	if pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
		pos++
	}

	trimmed := Line{
		Span: source.ByteSpan{
			Start: full.Span.Start + source.BytePos(pos),
			End:   full.Span.End,
		},
	}

	return full, trimmed, true, nil
}

type HeaderRule struct{}

func (r HeaderRule) Apply(c *Cursor) (ir.Block, bool, error) {
	line, ok := c.Peek()
	if !ok {
		return nil, false, nil
	}

	level, contentSpan, ok := r.tryParseHeaderLine(c.Source, line)
	if !ok {
		return nil, false, nil
	}

	line, ok = c.Next()
	if !ok {
		return nil, false, nil
	}

	applied := ir.Header{
		Level:       level,
		Span:        line.Span,
		ContentSpan: contentSpan,
	}

	return applied, true, nil
}

func (HeaderRule) tryParseHeaderLine(src *source.Source, line Line) (int, source.ByteSpan, bool) {
	if line.IsBlankLine(src) {
		return 0, source.ByteSpan{}, false
	}

	// count the leading spaces, reject if more than 3
	offset := line.BlockIndentSpaces(src)
	if offset > MaxValidIndentation {
		return 0, source.ByteSpan{}, false
	}

	s := src.Slice(line.Span)
	pos := offset
	level := 0

	// validate the marker
	if pos >= len(s) || s[pos] != '#' {
		return 0, source.ByteSpan{}, false
	}

	// count the marker run, reject if more than 6
	for pos < len(s) && s[pos] == '#' {
		pos++
		level++
		if level == 7 {
			return 0, source.ByteSpan{}, false
		}
	}

	// validate the delimiter (space or tab)
	if pos >= len(s) || (s[pos] != ' ' && s[pos] != '\t') {
		return 0, source.ByteSpan{}, false
	}

	// consume the delimiter
	for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
		pos++
	}

	// TODO: consider trimming suffix '#' characters

	// trim tailing spaces and tabs
	rawContent := s[pos:]
	trimmed := strings.TrimRight(rawContent, " \t")

	contentStart := line.Span.Start + source.BytePos(pos)
	contentEnd := contentStart + source.BytePos(len(trimmed))

	contentSpan := source.ByteSpan{
		Start: contentStart,
		End:   contentEnd,
	}

	return level, contentSpan, true
}

type ThematicBreakRule struct{}

func (r ThematicBreakRule) Apply(c *Cursor) (ir.Block, bool, error) {
	line, ok := c.Peek()
	if !ok {
		return nil, false, nil
	}

	if !r.tryParseThematicBreakLine(c.Source, line) {
		return nil, false, nil
	}

	line, ok = c.Next()
	if !ok {
		return nil, false, nil
	}

	applied := ir.ThematicBreak{
		Span: line.Span,
	}

	return applied, true, nil

}

func (ThematicBreakRule) tryParseThematicBreakLine(src *source.Source, line Line) bool {
	if line.IsBlankLine(src) {
		return false
	}

	// count the leading spaces, reject if more than 3
	offset := line.BlockIndentSpaces(src)
	if offset > MaxValidIndentation {
		return false
	}

	s := src.Slice(line.Span)
	pos := offset

	if pos >= len(s) {
		return false
	}

	// validate the first marker character
	var marker byte
	switch s[pos] {
	case '-', '*', '_':
		marker = s[pos]
	default:
		return false
	}

	// count the marker run, skipping whitespace and rejecting mixed markers
	markerCount := 0
	for pos < len(s) {
		b := s[pos]
		switch b {
		case ' ', '\t':
			pos++
			continue
		default:
			if b != marker {
				return false
			}
			pos++
			markerCount++
		}
	}

	return markerCount >= 3
}

type ParagraphRuleMarker interface {
	isParagraphRule()
}

type ParagraphRule struct{}

func (ParagraphRule) isParagraphRule() {}

func (r ParagraphRule) Apply(c *Cursor) (ir.Block, bool, error) {
	spans, ok, err := r.consumeParagraphRun(c)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
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

func (ParagraphRule) consumeParagraphRun(c *Cursor) ([]source.ByteSpan, bool, error) {
	line, ok := c.Peek()
	if !ok {
		return nil, false, nil
	}
	if line.IsBlankLine(c.Source) {
		return nil, false, nil
	}

	var spans []source.ByteSpan

	// consume the first line
	line, ok = c.Next()
	if !ok {
		return nil, false, nil
	}

	spans = append(spans, line.Span)

	// consume continuation lines
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

	return spans, true, nil

}
