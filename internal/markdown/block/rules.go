package block

import (
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

const MaxValidIndentation = 3
const MinValidCodeBlockIndentation = MaxValidIndentation + 1

type ParagraphTransparentRuleMarker interface {
	isParagraphTransparent()
}

type BuildRule interface {
	Apply(c *Cursor) (ir.Block, bool, error)
}

type BlockQuoteRule struct{}

func (r BlockQuoteRule) Apply(c *Cursor) (ir.Block, bool, error) {
	var spans []source.ByteSpan
	var trimmedLines []Line

	// must consume at least one line to apply
	full, trimmed, ok := r.tryConsumeQuoteLine(c)
	if !ok {
		return nil, false, nil
	}

	spans = append(spans, full.Span)
	trimmedLines = append(trimmedLines, trimmed)

	// consume subsequent quote lines
	for {
		full, trimmed, ok := r.tryConsumeQuoteLine(c)
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

func (BlockQuoteRule) tryConsumeQuoteLine(c *Cursor) (Line, Line, bool) {
	// peek next line, reject if EOF or blank
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return Line{}, Line{}, false
	}

	// count the leading indentation, reject if greater than 3 visual columns
	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return Line{}, Line{}, false
	}

	// derived line guard
	derived := !line.IsPhysicalLineStart(c.Source)
	if derived && indentBytes > 0 {
		return Line{}, Line{}, false
	}

	s := c.Source.Slice(line.Span)
	pos := indentBytes

	// validate the marker
	if pos >= len(s) || s[pos] != '>' {
		return Line{}, Line{}, false
	}

	// commit to consuming the line
	full := c.MustNext()

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

	return full, trimmed, true
}

type OLMarkerLineResult struct {
	MarkerLine      Line
	ContentLine     Line
	ListIndentCols  int
	ItemContentCols int
	MarkerDelim     byte
	StartNumber     int
}

type OrderedListRule struct{}

func (r OrderedListRule) Apply(c *Cursor) (ir.Block, bool, error) {
	//must consume at least one line to apply
	result, ok := r.tryConsumeFirstItem(c)
	if !ok {
		return nil, false, nil
	}

	listItems := make([]ir.ListItem, 0, 4)
	tight := true
	start := result.StartNumber

	// attempt to collect item body lines, append item,
	// and then check for sibling items or break
	for {
		lines, spans, keptBlank := r.consumeItemBody(c, result)
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
		result, sepBlanks, ok = r.tryConsumeSiblingItem(c, result.ListIndentCols, result.MarkerDelim)
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

	applied := ir.OrderedList{
		Span:  listSpan,
		Items: listItems,
		Tight: tight,
		Start: start,
	}

	return applied, true, nil
}

func (r OrderedListRule) tryConsumeFirstItem(c *Cursor) (OLMarkerLineResult, bool) {
	// peek next line, reject if EOF or blank
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return OLMarkerLineResult{}, false
	}

	// measure the leading indentation
	// reject line if indent is greater than 3 visual columns,
	// or if less than the cursor baseline
	relIndentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || relIndentCols > MaxValidIndentation {
		return OLMarkerLineResult{}, false
	}

	// calculate the list indentation (visual columns)
	listIndentCols, _ := c.AbsBlockIndent(line)

	return r.tryParseMarkerLine(c, line, listIndentCols, indentBytes)
}

func (r OrderedListRule) tryConsumeSiblingItem(c *Cursor, listIndentCols int, markerDelim byte) (OLMarkerLineResult, bool, bool) {
	// mark cursor location in case of rollback
	m := c.Mark()
	consumedBlanks := false

	// peek next line, reject if EOF
	line, ok := c.Peek()
	if !ok {
		return OLMarkerLineResult{}, false, false
	}

	// consume trailing blank lines
	for line.IsBlankLine(c.Source) {
		c.MustNext()
		consumedBlanks = true

		line, ok = c.Peek()
		if !ok {
			c.Reset(m)
			return OLMarkerLineResult{}, false, false
		}
	}

	// measure the line indentation (visual columns)
	// reject if less than listIndentCols (dedent),
	// or if greater than listIndentCols (not a sibling item)
	absIndentCols, indentBytes := c.AbsBlockIndent(line)
	if absIndentCols != listIndentCols {
		c.Reset(m)
		return OLMarkerLineResult{}, false, false
	}

	// try to parse the next non-blank line
	// if parse fails, roll back the trailing blanks
	// reject if the sibling item does not share the same delimiter punctuation
	result, ok := r.tryParseMarkerLine(c, line, listIndentCols, indentBytes)
	if !ok || result.MarkerDelim != markerDelim {
		c.Reset(m)
		return OLMarkerLineResult{}, false, false
	}

	return result, consumedBlanks, true
}

func (r OrderedListRule) consumeItemBody(c *Cursor, start OLMarkerLineResult) ([]Line, []source.ByteSpan, bool) {
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

			// append a derived line trimmed to the item baseline for recursive parsing
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

	return itemLines, itemSpans, keptBlank
}

func (r OrderedListRule) tryParseMarkerLine(c *Cursor, line Line, listIndentCols, indentBytes int) (OLMarkerLineResult, bool) {
	s := c.Source.Slice(line.Span)
	pos := indentBytes
	col := listIndentCols
	var delim byte
	var num int

	// validate the marker character
	if pos >= len(s) {
		return OLMarkerLineResult{}, false
	}

	digitStart := pos
	for pos < len(s) {
		b := s[pos]
		if b < '0' || b > '9' {
			break
		}

		digit := int(b - '0')
		num = (num * 10) + digit
		pos++
		col++
	}

	if digitStart == pos {
		return OLMarkerLineResult{}, false
	}

	// reject for absurdly high numbers
	if num > 1e9 {
		return OLMarkerLineResult{}, false
	}

	// validate, consume, and record delimiter punctuation (period or right parens)
	if pos >= len(s) {
		return OLMarkerLineResult{}, false
	}
	switch s[pos] {
	case '.', ')':
		delim = s[pos]
		pos++
		col++
	default:
		return OLMarkerLineResult{}, false
	}

	// validate the delimiter (at least one space or tab)
	if pos >= len(s) {
		return OLMarkerLineResult{}, false
	}
	switch s[pos] {
	case ' ', '\t':
	// ok, continue
	default:
		return OLMarkerLineResult{}, false
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

	result := OLMarkerLineResult{
		MarkerLine:      markerLine,
		ContentLine:     contentLine,
		ListIndentCols:  listIndentCols,
		ItemContentCols: itemContentCols,
		MarkerDelim:     delim,
		StartNumber:     num,
	}

	return result, true
}

type ULMarkerLineResult struct {
	MarkerLine      Line
	ContentLine     Line
	ListIndentCols  int
	ItemContentCols int
}

type UnorderedListRule struct{}

func (r UnorderedListRule) Apply(c *Cursor) (ir.Block, bool, error) {
	// must consume at least one line to apply
	result, ok := r.tryConsumeFirstItem(c)
	if !ok {
		return nil, false, nil
	}

	listItems := make([]ir.ListItem, 0, 4)
	tight := true

	// attempt to collect item body lines, append item,
	// and then check for sibling items or break
	for {
		lines, spans, keptBlank := r.consumeItemBody(c, result)
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
		result, sepBlanks, ok = r.tryConsumeSiblingItem(c, result.ListIndentCols)
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

func (r UnorderedListRule) tryConsumeFirstItem(c *Cursor) (ULMarkerLineResult, bool) {
	// peek next line, reject if EOF or blank
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return ULMarkerLineResult{}, false
	}

	// measure the leading indentation
	// reject line if indent is greater than 3 visual columns,
	// or if less than the cursor baseline
	relIndentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || relIndentCols > MaxValidIndentation {
		return ULMarkerLineResult{}, false
	}

	// calculate the list indentation (visual columns)
	listIndentCols, _ := c.AbsBlockIndent(line)

	return r.tryParseMarkerLine(c, line, listIndentCols, indentBytes)
}

func (r UnorderedListRule) tryConsumeSiblingItem(c *Cursor, listIndentCols int) (ULMarkerLineResult, bool, bool) {
	// mark cursor location in case of rollback
	m := c.Mark()
	consumedBlanks := false

	// peek next line, reject if EOF
	line, ok := c.Peek()
	if !ok {
		return ULMarkerLineResult{}, false, false
	}

	// consume trailing blank lines
	for line.IsBlankLine(c.Source) {
		c.MustNext()
		consumedBlanks = true

		line, ok = c.Peek()
		if !ok {
			c.Reset(m)
			return ULMarkerLineResult{}, false, false
		}
	}

	// measure the line indentation (visual columns)
	// reject if less than listIndentCols (dedent),
	// or if greater than listIndentCols (not a sibling item)
	absIndentCols, indentBytes := c.AbsBlockIndent(line)
	if absIndentCols != listIndentCols {
		c.Reset(m)
		return ULMarkerLineResult{}, false, false
	}

	// try to parse the next non-blank line
	// if parse fails, roll back the trailing blanks
	result, ok := r.tryParseMarkerLine(c, line, listIndentCols, indentBytes)
	if !ok {
		c.Reset(m)
		return ULMarkerLineResult{}, false, false
	}

	return result, consumedBlanks, true
}

func (r UnorderedListRule) consumeItemBody(c *Cursor, start ULMarkerLineResult) ([]Line, []source.ByteSpan, bool) {
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

			// append a derived line trimmed to the item baseline for recursive parsing
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

	return itemLines, itemSpans, keptBlank
}

func (r UnorderedListRule) tryParseMarkerLine(c *Cursor, line Line, listIndentCols, indentBytes int) (ULMarkerLineResult, bool) {
	s := c.Source.Slice(line.Span)
	pos := indentBytes
	col := listIndentCols

	// validate the first marker character
	if pos >= len(s) {
		return ULMarkerLineResult{}, false
	}
	switch s[pos] {
	case '-', '*', '+':
	// ok, continue
	default:
		return ULMarkerLineResult{}, false
	}

	// consume the marker
	pos++
	col++

	// validate the delimiter (at least one space or tab)
	if pos >= len(s) {
		return ULMarkerLineResult{}, false
	}
	switch s[pos] {
	case ' ', '\t':
	// ok, continue
	default:
		return ULMarkerLineResult{}, false
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

	result := ULMarkerLineResult{
		MarkerLine:      markerLine,
		ContentLine:     contentLine,
		ListIndentCols:  listIndentCols,
		ItemContentCols: itemContentCols,
	}

	return result, true
}

type HeaderRule struct{}

func (r HeaderRule) Apply(c *Cursor) (ir.Block, bool, error) {
	level, contentSpan, ok := r.tryParseHeaderLine(c)
	if !ok {
		return nil, false, nil
	}

	// consume next line
	line := c.MustNext()

	applied := ir.Header{
		Level:       level,
		Span:        line.Span,
		ContentSpan: contentSpan,
	}

	return applied, true, nil
}

func (HeaderRule) tryParseHeaderLine(c *Cursor) (int, source.ByteSpan, bool) {
	// peek next line, reject if EOF or blank
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return 0, source.ByteSpan{}, false
	}

	// count the leading indentation, reject if greater than 3 visual columns
	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return 0, source.ByteSpan{}, false
	}

	s := c.Source.Slice(line.Span)
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
	if !r.tryParseThematicBreakLine(c) {
		return nil, false, nil
	}

	line := c.MustNext()

	applied := ir.ThematicBreak{
		Span: line.Span,
	}

	return applied, true, nil

}

func (ThematicBreakRule) tryParseThematicBreakLine(c *Cursor) bool {
	// peek next line, reject if EOF or blank
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return false
	}

	// count the leading indentation, reject if greater than 3 visual columns
	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return false
	}

	s := c.Source.Slice(line.Span)
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

type IndentedCodeBlockRule struct{}

func (IndentedCodeBlockRule) isParagraphTransparent() {}

func (r IndentedCodeBlockRule) Apply(c *Cursor) (ir.Block, bool, error) {
	lineSpans, ok := r.consumeIndentedCodeBlock(c)
	if !ok {
		return nil, false, nil
	}

	blockSpan := source.ByteSpan{
		Start: lineSpans[0].Start,
		End:   lineSpans[len(lineSpans)-1].End,
	}

	applied := ir.IndentedCodeBlock{
		Span:  blockSpan,
		Lines: lineSpans,
	}

	return applied, true, nil

}

func (r IndentedCodeBlockRule) consumeIndentedCodeBlock(c *Cursor) ([]source.ByteSpan, bool) {
	// peek next line, reject if EOF or blank
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return nil, false
	}

	// validate the leading indentation
	if !r.tryParseIndentedCodeBlockLine(c, line) {
		return nil, false
	}

	// consume first line and initialize payload
	line = c.MustNext()
	lineSpans := []source.ByteSpan{line.Span}

	blankRun := struct {
		active     bool
		cursorMark int
		lineMark   int
	}{
		active: false,
	}

	for {
		// peek next line, break if EOF
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
				blankRun.lineMark = len(lineSpans)
			}

			// consume blank line
			line := c.MustNext()
			lineSpans = append(lineSpans, line.Span)

			continue
		}

		// non-blank and meets the indentation baseline
		if r.tryParseIndentedCodeBlockLine(c, nextLine) {
			// reset blank run flag
			blankRun.active = false

			// consume the next line
			line := c.MustNext()
			lineSpans = append(lineSpans, line.Span)

			continue
		}

		// non-blank but does not meet the indentation requirement
		// stop collecting, roll back any trailing blanks
		if blankRun.active {
			// reset blank run flag
			blankRun.active = false

			// roll back cursor position & spans
			c.Reset(blankRun.cursorMark)
			lineSpans = lineSpans[:blankRun.lineMark]
		}

		break
	}

	// defensive panic
	if len(lineSpans) == 0 {
		panic("indented code block invariant violated: matched first item but produced no payload")
	}

	return lineSpans, true
}

func (IndentedCodeBlockRule) tryParseIndentedCodeBlockLine(c *Cursor, line Line) bool {
	// count the leading indentation, reject if less than 4 visual columns
	indentCols, _, ok := c.RelBlockIndent(line)
	if !ok || indentCols < MinValidCodeBlockIndentation {
		return false
	}

	return true
}

type FCBMarkerLineResult struct {
	Marker         byte
	MarkerCount    int
	OpenIndentCols int
	InfoString     source.ByteSpan
}

type FencedCodeBlockRule struct{}

func (r FencedCodeBlockRule) Apply(c *Cursor) (ir.Block, bool, error) {
	result, ok := r.tryParseOpeningFenceLine(c)
	if !ok {
		return nil, false, nil
	}

	// opening fence validated, consume payload until closing fence or EOF
	blockSpan, payload := r.consumeFencedCodeBlock(c, result)

	applied := ir.FencedCodeBlock{
		Span:           blockSpan,
		OpenIndentCols: result.OpenIndentCols,
		InfoStringSpan: result.InfoString,
		Lines:          payload,
	}

	return applied, true, nil
}

func (FencedCodeBlockRule) tryParseOpeningFenceLine(c *Cursor) (FCBMarkerLineResult, bool) {
	// peek next line, reject if EOF or blank
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return FCBMarkerLineResult{}, false
	}

	// count the leading indentation, reject if greater than 3 visual columns
	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return FCBMarkerLineResult{}, false
	}

	s := c.Source.Slice(line.Span)
	pos := indentBytes

	if pos >= len(s) {
		return FCBMarkerLineResult{}, false
	}

	// validate the first marker character
	var marker byte
	switch s[pos] {
	case '`', '~':
		marker = s[pos]
	default:
		return FCBMarkerLineResult{}, false
	}

	// count the marker run, stopping at first non-marker character (including whitespace)
	markerCount := 0
	for pos < len(s) {
		b := s[pos]
		if b == marker {
			pos++
			markerCount++
			continue
		}

		break
	}

	// reject if less than three consecutive marker characters in the marker run
	if markerCount < 3 {
		return FCBMarkerLineResult{}, false
	}

	// consume any delimiter whitespace
	for pos < len(s) {
		b := s[pos]
		if b == ' ' || b == '\t' {
			pos++
			continue
		}

		break
	}

	// if marker is backtick, ensure infostring does not also contain backticks
	if marker == '`' {
		infoPos := pos
		for infoPos < len(s) {
			b := s[infoPos]
			if b == marker {
				return FCBMarkerLineResult{}, false
			}

			infoPos++
		}
	}

	infoStringSpan := source.ByteSpan{
		Start: line.Span.Start + source.BytePos(pos),
		End:   line.Span.End,
	}

	result := FCBMarkerLineResult{
		Marker:         marker,
		MarkerCount:    markerCount,
		OpenIndentCols: indentCols,
		InfoString:     infoStringSpan,
	}

	return result, true
}

func (r FencedCodeBlockRule) consumeFencedCodeBlock(c *Cursor, opener FCBMarkerLineResult) (source.ByteSpan, []source.ByteSpan) {
	// opening fence line, already validated
	line := c.MustNext()
	blockSpanStart := line.Span.Start
	blockSpanEnd := line.Span.End

	// consume all lines until closing fence or EOF
	lineSpans := []source.ByteSpan{}
	for {
		// peek at next line, break if EOF
		line, ok := c.Peek()
		if !ok {
			break
		}

		// if line is closing fence, record span and break
		if r.tryParseClosingFenceLine(c, opener.Marker, opener.MarkerCount) {
			line = c.MustNext()
			blockSpanEnd = line.Span.End

			break
		}

		// otherwise, consume the line
		line = c.MustNext()
		blockSpanEnd = line.Span.End
		lineSpans = append(lineSpans, line.Span)
	}

	span := source.ByteSpan{
		Start: blockSpanStart,
		End:   blockSpanEnd,
	}

	return span, lineSpans
}

func (FencedCodeBlockRule) tryParseClosingFenceLine(c *Cursor, marker byte, markerCount int) bool {
	// peek next line, reject if EOF or blank
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return false
	}

	// count the leading indentation, reject if greater than 3 visual columns
	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return false
	}

	s := c.Source.Slice(line.Span)
	pos := indentBytes

	if pos >= len(s) {
		return false
	}

	// validate the first marker character
	if s[pos] != marker {
		return false
	}

	// consume the marker run
	currentCount := 0
	for pos < len(s) && s[pos] == marker {
		pos++
		currentCount++
	}

	// reject line if current run is shorter than opening run
	if currentCount < markerCount {
		return false
	}

	// consume trailing whitespace and reject if any other character is seen
	for pos < len(s) {
		b := s[pos]
		switch b {
		case ' ', '\t':
			pos++
		default:
			return false
		}
	}

	return true
}

type ParagraphRule struct{}

func (ParagraphRule) isParagraphTransparent() {}

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
			underline := c.MustNext()

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
	// peek next line, reject if EOF or blank line
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return nil, false, nil
	}

	var spans []source.ByteSpan

	// consume the first line
	line = c.MustNext()

	spans = append(spans, line.Span)

	// consume continuation lines
	for {
		line, ok := c.Peek()
		if !ok || line.IsBlankLine(c.Source) {
			break
		}

		_, isSetext := r.tryParseSetextHeadingLine(c, line)
		if isSetext {
			break
		}

		startsBlock, err := c.StartsParagraphInterruptingBlock()
		if err != nil {
			return nil, false, err
		}
		if startsBlock {
			break
		}

		line = c.MustNext()

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
