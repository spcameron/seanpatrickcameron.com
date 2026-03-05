package block

import (
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

// TODO: remove redundant ok checks on Next --> Use MustNext
// TODO: condense ok and line empty checks on initial peeks

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

type MarkerLineResult struct {
	MarkerLine      Line
	ContentLine     Line
	ListIndentCols  int
	ItemContentCols int
}

type UnorderedListRule struct{}

func (r UnorderedListRule) Apply(c *Cursor) (ir.Block, bool, error) {
	// must consume at least one line to apply
	result, ok, err := r.tryConsumeFirstItem(c)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}

	listItems := make([]ir.ListItem, 0, 4)
	tight := true

	// attempt to collect item body lines, append item,
	// and then check for sibling items or break
	for {
		lines, spans, keptBlank, err := r.consumeItemBody(c, result)
		if err != nil {
			return nil, false, err
		}
		if keptBlank {
			tight = false
		}

		children, err := buildBlocks(c.Source, c.Rules, lines, result.ItemContentCols)
		if err != nil {
			return nil, false, err
		}

		// defensive panic
		if len(spans) == 0 {
			panic("unordered list invariant violated: consumed marker line but produced no item spans")
		}

		itemSpan := source.ByteSpan{
			Start: result.MarkerLine.Span.Start,
			End:   spans[len(spans)-1].End,
		}

		item := ir.ListItem{
			Span:     itemSpan,
			Children: children,
		}

		listItems = append(listItems, item)

		sepBlanks := false
		result, sepBlanks, ok, err = r.tryConsumeSiblingItem(c, result.ListIndentCols)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			break
		}
		if sepBlanks {
			tight = false
		}
	}

	// defensive panic
	if len(listItems) == 0 {
		panic("unordered list invariant violated: matched first item but produced no items")
	}

	listSpan := source.ByteSpan{
		Start: listItems[0].Span.Start,
		End:   listItems[len(listItems)-1].Span.End,
	}

	applied := ir.UnorderedList{
		Span:  listSpan,
		Items: listItems,
		Tight: tight,
	}

	return applied, true, nil
}

func (r UnorderedListRule) tryConsumeFirstItem(c *Cursor) (MarkerLineResult, bool, error) {
	// peek next line, reject if EOF or blank
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return MarkerLineResult{}, false, nil
	}

	// measure the leading indentation
	// reject line if indent is greater than 3 visual columns,
	// or if less than the cursor baseline
	relIndentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || relIndentCols > MaxValidIndentation {
		return MarkerLineResult{}, false, nil
	}

	// calculate the list indentation (visual columns)
	listIndentCols, _ := c.AbsBlockIndent(line)

	return r.tryParseMarkerLine(c, line, listIndentCols, indentBytes)
}

func (r UnorderedListRule) tryConsumeSiblingItem(c *Cursor, listIndentCols int) (MarkerLineResult, bool, bool, error) {
	// mark cursor location in case rollback
	m := c.Mark()
	consumedBlanks := false

	// peek next line, reject if EOF
	line, ok := c.Peek()
	if !ok {
		return MarkerLineResult{}, false, false, nil
	}

	// consume trailing blank lines
	for line.IsBlankLine(c.Source) {
		c.MustNext()
		consumedBlanks = true

		line, ok = c.Peek()
		if !ok {
			c.Reset(m)
			return MarkerLineResult{}, false, false, nil
		}
	}

	// measure the line indentation (visual columns)
	// reject if less than listIndentCols (dedent),
	// or if greater than listIndentCols (not a sibling item)
	absIndentCols, indentBytes := c.AbsBlockIndent(line)
	if absIndentCols != listIndentCols {
		c.Reset(m)
		return MarkerLineResult{}, false, false, nil
	}

	// try to parse the next non-blank line
	// if parse fails, roll back the trailing blanks
	result, ok, err := r.tryParseMarkerLine(c, line, listIndentCols, indentBytes)
	if err != nil {
		c.Reset(m)
		return MarkerLineResult{}, false, false, nil
	}
	if !ok {
		c.Reset(m)
		return MarkerLineResult{}, false, false, nil
	}

	return result, consumedBlanks, true, nil
}

func (r UnorderedListRule) consumeItemBody(c *Cursor, start MarkerLineResult) ([]Line, []source.ByteSpan, bool, error) {
	itemSpans := []source.ByteSpan{start.MarkerLine.Span}
	itemLines := []Line{start.ContentLine}

	blankRun := struct {
		active     bool
		cursorMark int
		spanMark   int
		lineMark   int
	}{
		active: false,
	}

	keptBlank := false

	for {
		// peek next line, reject if EOF
		nextLine, ok := c.Peek()
		if !ok {
			break
		}

		// if blank line, tentatively consume
		if nextLine.IsBlankLine(c.Source) {
			// if starting a blank run, mark checkpoints in case of rollback
			if !blankRun.active {
				blankRun.active = true
				blankRun.cursorMark = c.Mark()
				blankRun.spanMark = len(itemSpans)
				blankRun.lineMark = len(itemLines)
			}

			// consume blank line
			line := c.MustNext()
			itemSpans = append(itemSpans, line.Span)
			itemLines = append(itemLines, line)

			continue
		}

		// absolute indentation for the next line
		absIndentCols, _ := c.AbsBlockIndent(nextLine)

		// non-blank and meets the content baseline
		if absIndentCols >= start.ItemContentCols {
			// toggle flag if continuation line follows a blank line
			if blankRun.active {
				keptBlank = true
			}

			// reset blank run flag
			blankRun.active = false

			// consume the next line
			line := c.MustNext()
			itemSpans = append(itemSpans, line.Span)

			// append a derived line trimed to the item baseline for recursive parsing
			trimmed := line.TrimIndentToCols(c.Source, start.ItemContentCols)
			itemLines = append(itemLines, trimmed)

			continue
		}

		// non-blank but does not meet the content baseline
		// stop collecting, roll back any trailing blanks
		if blankRun.active {
			// reset blank run flag
			blankRun.active = false

			// roll back cursor position, spans and items
			c.Reset(blankRun.cursorMark)
			itemSpans = itemSpans[:blankRun.spanMark]
			itemLines = itemLines[:blankRun.lineMark]
		}

		break
	}

	return itemLines, itemSpans, keptBlank, nil
}

func (r UnorderedListRule) tryParseMarkerLine(c *Cursor, line Line, listIndentCols, indentBytes int) (MarkerLineResult, bool, error) {
	s := c.Source.Slice(line.Span)
	pos := indentBytes
	col := listIndentCols

	// validate the first marker character
	if pos >= len(s) {
		return MarkerLineResult{}, false, nil
	}
	switch s[pos] {
	case '-', '*', '+':
	// ok, continue
	default:
		return MarkerLineResult{}, false, nil
	}

	// consume the marker
	pos++
	col++

	// validate the delimiter (at least one space or tab)
	if pos >= len(s) {
		return MarkerLineResult{}, false, nil
	}
	switch s[pos] {
	case ' ', '\t':
	// ok, continue
	default:
		return MarkerLineResult{}, false, nil
	}

	// consume the next line
	markerLine := c.MustNext()

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

	// derive content column and byte starting positions
	itemContentCols := col
	contentOffsetBytes := pos

	contentStart := markerLine.Span.Start + source.BytePos(contentOffsetBytes)

	contentLine := Line{
		Span: source.ByteSpan{
			Start: contentStart,
			End:   markerLine.Span.End,
		},
	}

	result := MarkerLineResult{
		MarkerLine:      markerLine,
		ContentLine:     contentLine,
		ListIndentCols:  listIndentCols,
		ItemContentCols: itemContentCols,
	}

	return result, true, nil
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
