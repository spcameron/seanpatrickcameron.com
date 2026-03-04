package block

import (
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

// TODO: remove redundant ok checks on Next

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
	innerBlocks, err := buildBlocks(c.Source, c.Rules, trimmedLines, c.BaselineCols)
	if err != nil {
		return nil, false, err
	}

	span := source.ByteSpan{
		Start: spans[0].Start,
		End:   spans[len(spans)-1].End,
	}

	applied := ir.BlockQuote{
		Children: innerBlocks,
		Span:     span,
	}

	return applied, true, nil
}

func (BlockQuoteRule) tryConsumeQuoteLine(c *Cursor) (Line, Line, bool, error) {
	line, ok := c.Peek()
	if !ok {
		return Line{}, Line{}, false, nil
	}

	// blank line terminates the quote run
	if line.IsBlankLine(c.Source) {
		return Line{}, Line{}, false, nil
	}

	// count the leading indentation, reject if greater than 3 visual columns
	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return Line{}, Line{}, false, nil
	}

	// derived line guard
	derived := !line.IsPhysicalLineStart(c.Source)
	if derived && indentBytes > 0 {
		return Line{}, Line{}, false, nil
	}

	s := c.Source.Slice(line.Span)
	pos := indentBytes

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

type UnorderedListRule struct{}

func (r UnorderedListRule) Apply(c *Cursor) (ir.Block, bool, error) {
	line, ok := c.Peek()
	if !ok {
		return nil, false, nil
	}
	if line.IsBlankLine(c.Source) {
		return nil, false, nil
	}

	// count the leading indentation
	// reject line if indent is greater than 3 visual columns,
	// or if the indentation is less than the cursor baseline
	relIndentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || relIndentCols > MaxValidIndentation {
		return nil, false, nil
	}

	// calculate the list indentation (visual columns)
	listIndentCols, _ := c.AbsBlockIndent(line)

	s := c.Source.Slice(line.Span)
	pos := indentBytes
	col := listIndentCols

	// validate the first marker character
	if pos >= len(s) {
		return nil, false, nil
	}
	switch s[pos] {
	case '-', '*', '+':
	// ok
	default:
		return nil, false, nil
	}

	// consume the marker
	pos++
	col++

	// validate the delimiter (at least one space or tab)
	if pos >= len(s) || (s[pos] != ' ' && s[pos] != '\t') {
		return nil, false, nil
	}

	// NOTE: committed to building the list at this point

	lineSpans := []source.ByteSpan{}
	listItems := []ir.ListItem{}

	// NOTE: outer loop
	// parse one list item at a time (sibling)
buildList:
	for {
		// consume the next line
		markerLine, _ := c.Next()
		itemLines := []Line{}

		lineSpans = append(lineSpans, markerLine.Span)

		// consume the delimiter run
		for pos < len(s) {
			b := s[pos]
			if b == ' ' {
				col++
				pos++
				continue
			}
			if b == '\t' {
				col += tabWidth - (col % tabWidth)
				pos++
				continue
			}
			break
		}

		itemContentCols := col
		lineContentByte := pos

		contentStart := markerLine.Span.Start + source.BytePos(lineContentByte)

		line := Line{
			Span: source.ByteSpan{
				Start: contentStart,
				End:   markerLine.Span.End,
			},
		}

		itemLines = append(itemLines, line)

		// NOTE: inner loop
		// collect the current item's body lines (continuation)
		for {
			nextLine, ok := c.Peek()
			// if EOF, stop collecting, proceed to finalization
			if !ok {
				break
			}

			// if blank line, tenatively consume
			if nextLine.IsBlankLine(c.Source) {
				_ = "quiet staticcheck complaint about empty branch"
				// TODO: blank line policy
			}

			// absolute indentation for the next line
			absIndentCols, _ := c.AbsBlockIndent(nextLine)

			// non-blank and meets the content baseline
			if absIndentCols >= itemContentCols {
				line, _ := c.Next()
				lineSpans = append(lineSpans, line.Span)
				itemLines = append(itemLines, line)

				// reset trailing blanks to zero here

				continue
			}

			// non-blank but does not meet the content baseline
			// stop collecting, roll back any trailing blanks
			break
		}

		// recursively parse children
		children, err := buildBlocks(c.Source, c.Rules, itemLines, itemContentCols)
		if err != nil {
			return nil, false, err
		}

		// finalize the list item
		var itemSpan source.ByteSpan
		if len(itemLines) > 0 {
			itemSpan.Start = markerLine.Span.Start
			itemSpan.End = itemLines[len(itemLines)-1].Span.End
		}

		item := ir.ListItem{
			Span:     itemSpan,
			Children: children,
		}

		// update listItems
		listItems = append(listItems, item)

		// peek for sibling item
		nextLine, ok := c.Peek()
		if !ok {
			break
		}
		if nextLine.IsBlankLine(c.Source) {
			break
			// possible to allow blank lines between items later
		}

		// calculate the next line indentation (visual columns)
		absIndentCols, indentBytes := c.AbsBlockIndent(nextLine)

		// dedent, list ends
		if absIndentCols < listIndentCols {
			break
		}

		// not a sibling item, list ends
		if absIndentCols != listIndentCols {
			break
		}

		// the following is repeated from first-line validation
		// extract to helper
		//
		// potential sibling item
		// validate marker & delimiter as before
		// if valid, continue; if not, list ends
		s = c.Source.Slice(nextLine.Span)
		pos = indentBytes
		col = listIndentCols

		// validate the first marker character
		if pos >= len(s) {
			break
		}
		switch s[pos] {
		case '-', '*', '+':
		//ok
		default:
			break buildList
		}

		// consume the marker
		pos++
		col++

		// validate the delimiter (at least one space or tab)
		if pos >= len(s) || (s[pos] != ' ' && s[pos] != '\t') {
			break buildList
		}

		// valid sibling, continue loop
	}

	var listSpan source.ByteSpan
	if len(lineSpans) > 0 {
		listSpan.Start = lineSpans[0].Start
		listSpan.End = lineSpans[len(lineSpans)-1].End
	}

	applied := ir.UnorderedList{
		Span:  listSpan,
		Items: listItems,
	}

	return applied, true, nil
}

type HeaderRule struct{}

func (r HeaderRule) Apply(c *Cursor) (ir.Block, bool, error) {
	line, ok := c.Peek()
	if !ok {
		return nil, false, nil
	}

	level, contentSpan, ok := r.tryParseHeaderLine(c, line)
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

func (HeaderRule) tryParseHeaderLine(c *Cursor, line Line) (int, source.ByteSpan, bool) {
	src := c.Source

	if line.IsBlankLine(src) {
		return 0, source.ByteSpan{}, false
	}

	// count the leading indentation, reject if greater than 3 visual columns
	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return 0, source.ByteSpan{}, false
	}

	s := src.Slice(line.Span)
	pos := indentBytes
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

	// consume the delimiter and leading whitespace
	for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
		pos++
	}

	// TODO: consider trimming suffix '#' characters
	// separate flow, not simply added to TrimRight

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

	if !r.tryParseThematicBreakLine(c, line) {
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

func (ThematicBreakRule) tryParseThematicBreakLine(c *Cursor, line Line) bool {
	src := c.Source

	if line.IsBlankLine(src) {
		return false
	}

	// count the leading indentation, reject if greater than 3 visual columns
	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return false
	}

	s := src.Slice(line.Span)
	pos := indentBytes

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

	contentSpan := source.ByteSpan{
		Start: spans[0].Start,
		End:   spans[len(spans)-1].End,
	}

	if line, ok := c.Peek(); ok {
		level, isSetext := r.tryParseSetextHeadingLine(c, line)
		if isSetext {
			underline, ok := c.Next()
			if !ok {
				return nil, false, nil
			}

			headerSpan := source.ByteSpan{
				Start: spans[0].Start,
				End:   underline.Span.End,
			}

			applied := ir.Header{
				Level:       level,
				Span:        headerSpan,
				ContentSpan: contentSpan,
			}

			return applied, true, nil
		}
	}

	applied := ir.Paragraph{
		Lines: spans,
		Span:  contentSpan,
	}

	return applied, true, nil
}

func (r ParagraphRule) consumeParagraphRun(c *Cursor) ([]source.ByteSpan, bool, error) {
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

		_, isSetext := r.tryParseSetextHeadingLine(c, line)
		if isSetext {
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

func (ParagraphRule) tryParseSetextHeadingLine(c *Cursor, line Line) (int, bool) {
	src := c.Source

	if line.IsBlankLine(src) {
		return 0, false
	}

	// count the leading indentation, reject if greater than 3 visual columns
	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return 0, false
	}

	s := src.Slice(line.Span)
	pos := indentBytes

	if pos >= len(s) {
		return 0, false
	}

	var marker byte
	var level int
	switch s[pos] {
	case '=':
		marker = s[pos]
		level = 1
	case '-':
		marker = s[pos]
		level = 2
	default:
		return 0, false
	}

	for pos < len(s) {
		b := s[pos]
		if b == ' ' || b == '\t' {
			break
		}

		if b != marker {
			return 0, false
		}

		pos++
	}

	for pos < len(s) {
		b := s[pos]
		switch b {
		case ' ', '\t':
			pos++
			continue
		default:
			return 0, false
		}
	}

	return level, true
}
