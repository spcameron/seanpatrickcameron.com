package block

import (
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/reference"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

// MaxValidIndentation is the greatest leading indentation, in columns, at
// which a block may still start without being treated as an indented code
// block.
const MaxValidIndentation = 3

// MinValidCodeBlockIndentation is the least leading indentation, in
// columns, required for an indented code block.
const MinValidCodeBlockIndentation = MaxValidIndentation + 1

// ParagraphTransparentRuleMarker is implemented by block rules that do not
// interrupt an in-progress paragraph.
type ParagraphTransparentRuleMarker interface {
	isParagraphTransparent()
}

// BuildRule is implemented by block parsing rules.
//
// Apply attempts to recognize and consume a block at the cursor position.
// It returns the applied block, whether the rule matched, and any error
// encountered during parsing.
type BuildRule interface {
	Apply(c *Cursor) (ir.Block, bool, error)
}

// ReferenceDefinitionRule parses link reference definitions and records
// them in the build metadata without producing a block node.
type ReferenceDefinitionRule struct{}

func (r ReferenceDefinitionRule) isParagraphTransparent() {}

func (r ReferenceDefinitionRule) Apply(c *Cursor) (ir.Block, bool, error) {
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return nil, false, nil
	}

	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return nil, false, nil
	}

	s := c.Source.Slice(line.Span)
	pos := indentBytes
	lineBase := line.Span.Start

	if pos >= len(s) {
		return nil, false, nil
	}

	if s[pos] != '[' {
		return nil, false, nil
	}

	labelSpanStart := pos
	pos++

	for pos < len(s) {
		if s[pos] == '\\' {
			if pos+1 < len(s) {
				pos += 2
			} else {
				pos++
			}
			continue
		}

		if s[pos] == ']' {
			pos++
			break
		}

		pos++
	}

	if pos >= len(s) {
		return nil, false, nil
	}

	labelSpanEnd := pos

	labelSpan := source.ByteSpan{
		Start: lineBase + source.BytePos(labelSpanStart),
		End:   lineBase + source.BytePos(labelSpanEnd),
	}

	contentSpan := source.ByteSpan{
		Start: labelSpan.Start + 1,
		End:   labelSpan.End - 1,
	}

	labelContent := c.Source.Slice(contentSpan)

	if ok := reference.ValidateLabel(labelContent); !ok {
		return nil, false, nil
	}

	normalizedLabel := reference.NormalizeLabel(labelContent)

	if s[pos] != ':' {
		return nil, false, nil
	}
	pos++

	pos = consumeSpacesTabs(s, pos)
	if pos >= len(s) {
		return nil, false, nil
	}

	destinationSpanRel, pos, ok := tryLinkDestination(s, pos)
	if !ok {
		return nil, false, nil
	}

	destinationSpan := source.ByteSpan{
		Start: lineBase + destinationSpanRel.Start,
		End:   lineBase + destinationSpanRel.End,
	}

	sepStart := pos
	pos = consumeSpacesTabs(s, pos)
	sepPresent := pos > sepStart

	if pos >= len(s) {
		fullSpan := source.ByteSpan{
			Start: lineBase + source.BytePos(indentBytes),
			End:   lineBase + source.BytePos(pos),
		}

		def := ir.ReferenceDefinition{
			FullSpan:        fullSpan,
			LabelSpan:       labelSpan,
			DestinationSpan: destinationSpan,
			NormalizedKey:   normalizedLabel,
		}

		if _, exists := c.Metadata.Definitions[normalizedLabel]; !exists {
			c.Metadata.Definitions[normalizedLabel] = def
		}

		c.MustNext()
		return nil, true, nil
	}

	if s[pos] == '"' || s[pos] == '\'' || s[pos] == '(' {
		// if no separator exists between destination and title, the definition is invalid
		if !sepPresent {
			return nil, false, nil
		}

		titleSpanRel, pos, ok := tryLinkTitle(s, pos)
		if !ok {
			return nil, false, nil
		}

		titleSpan := source.ByteSpan{
			Start: lineBase + titleSpanRel.Start,
			End:   lineBase + titleSpanRel.End,
		}

		pos = consumeSpacesTabs(s, pos)

		// if any non-space/tab byte encountered before reaching end of line, the definition is invalid
		if pos < len(s) {
			return nil, false, nil
		}

		fullSpan := source.ByteSpan{
			Start: lineBase + source.BytePos(indentBytes),
			End:   lineBase + source.BytePos(pos),
		}

		def := ir.ReferenceDefinition{
			FullSpan:        fullSpan,
			LabelSpan:       labelSpan,
			DestinationSpan: destinationSpan,
			TitleSpan:       titleSpan,
			HasTitle:        true,
			NormalizedKey:   normalizedLabel,
		}

		if _, exists := c.Metadata.Definitions[normalizedLabel]; !exists {
			c.Metadata.Definitions[normalizedLabel] = def
		}

		c.MustNext()
		return nil, true, nil
	}

	// trailing non-title content after destination means this is not a valid definition
	return nil, false, nil
}

func tryLinkDestination(s string, pos int) (source.ByteSpan, int, bool) {
	if pos >= len(s) {
		return source.ByteSpan{}, 0, false
	}

	if s[pos] == '<' {
		return tryAngleLinkDestination(s, pos)
	}
	return tryBareLinkDestination(s, pos)
}

func tryAngleLinkDestination(s string, pos int) (source.ByteSpan, int, bool) {
	if pos >= len(s) || s[pos] != '<' {
		return source.ByteSpan{}, 0, false
	}

	pos++
	start := pos

	for pos < len(s) {
		switch s[pos] {
		case '\n', '\r':
			// newlines are not permitted inside the destination
			return source.ByteSpan{}, 0, false

		case '<':
			// an unescaped '<' is an invalid destination
			return source.ByteSpan{}, 0, false

		case '>':
			// an unescaped '>' is the end of the destination
			span := source.ByteSpan{
				Start: source.BytePos(start),
				End:   source.BytePos(pos),
			}

			return span, pos + 1, true

		case '\\':
			if pos+1 < len(s) {
				pos += 2
				continue
			}
			pos++

		default:
			pos++
		}
	}

	// no closing '>' was encountered, invalid angle link destination
	return source.ByteSpan{}, 0, false
}

func tryBareLinkDestination(s string, pos int) (source.ByteSpan, int, bool) {
	if pos >= len(s) {
		return source.ByteSpan{}, 0, false
	}

	// must not start with '<'
	if s[pos] == '<' {
		return source.ByteSpan{}, 0, false
	}

	start := pos
	depth := 0

	for pos < len(s) {
		b := s[pos]

		// termination conditions (do not consume)
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
			break
		}

		// invalid bytes
		if b < 0x20 || b == 0x7F {
			return source.ByteSpan{}, 0, false
		}

		switch b {
		case '\\':
			if pos+1 < len(s) {
				pos += 2
			} else {
				pos++
			}

		case '(':
			depth++
			pos++

		case ')':
			if depth == 0 {
				return source.ByteSpan{}, 0, false
			}
			depth--
			pos++

		default:
			pos++
		}
	}

	if pos == start {
		return source.ByteSpan{}, 0, false
	}

	if depth != 0 {
		return source.ByteSpan{}, 0, false
	}

	span := source.ByteSpan{
		Start: source.BytePos(start),
		End:   source.BytePos(pos),
	}

	// do not advance the pos, leave trailing whitespace for consumption upstream
	return span, pos, true
}

func tryLinkTitle(s string, pos int) (source.ByteSpan, int, bool) {
	if pos >= len(s) {
		return source.ByteSpan{}, 0, false
	}

	switch s[pos] {
	case '"', '\'':
		return tryQuotedLinkTitle(s, pos, s[pos])

	case '(':
		return tryParenLinkTitle(s, pos)

	default:
		return source.ByteSpan{}, 0, false
	}
}

func tryQuotedLinkTitle(s string, pos int, delim byte) (source.ByteSpan, int, bool) {
	if pos >= len(s) || s[pos] != delim {
		return source.ByteSpan{}, 0, false
	}

	pos++
	start := pos

	for pos < len(s) {
		switch s[pos] {
		case '\n', '\r':
			// newlines are not permitted inside the title
			return source.ByteSpan{}, 0, false

		case delim:
			// an unescaped delimiter ends the title
			span := source.ByteSpan{
				Start: source.BytePos(start),
				End:   source.BytePos(pos),
			}

			return span, pos + 1, true

		case '\\':
			if pos+1 < len(s) {
				pos += 2
				continue
			}
			pos++

		default:
			pos++
		}
	}

	return source.ByteSpan{}, 0, false
}

func tryParenLinkTitle(s string, pos int) (source.ByteSpan, int, bool) {
	if pos >= len(s) || s[pos] != '(' {
		return source.ByteSpan{}, 0, false
	}

	pos++
	start := pos
	depth := 1

	for pos < len(s) {
		switch s[pos] {
		case '\n', '\r':
			// newlines are not permitted inside the title
			return source.ByteSpan{}, 0, false

		case '(':
			depth++
			pos++

		case ')':
			depth--

			if depth == 0 {
				span := source.ByteSpan{
					Start: source.BytePos(start),
					End:   source.BytePos(pos),
				}

				return span, pos + 1, true
			}

			pos++

		case '\\':
			if pos+1 < len(s) {
				pos += 2
				continue
			}
			pos++

		default:
			pos++
		}
	}

	return source.ByteSpan{}, 0, false
}

func consumeSpacesTabs(s string, pos int) int {
	for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
		pos++
	}

	return pos
}

// BlockQuoteRule parses contiguous block quote lines and recursively builds
// block structure from their marker-trimmed content.
type BlockQuoteRule struct{}

func (r BlockQuoteRule) Apply(c *Cursor) (ir.Block, bool, error) {
	var spans []source.ByteSpan
	var trimmedLines []Line

	full, trimmed, ok := r.tryConsumeQuoteLine(c)
	if !ok {
		return nil, false, nil
	}

	spans = append(spans, full.Span)
	trimmedLines = append(trimmedLines, trimmed)

	for {
		full, trimmed, ok := r.tryConsumeQuoteLine(c)
		if !ok {
			break
		}

		spans = append(spans, full.Span)
		trimmedLines = append(trimmedLines, trimmed)
	}

	innerBlocks, err := buildBlocks(c.Source, c.Rules, trimmedLines, c.BaselineCols, c.Metadata)
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

// tryConsumeQuoteLine consumes a single block quote line and returns both
// the full line and a marker-trimmed line for recursive parsing.
func (BlockQuoteRule) tryConsumeQuoteLine(c *Cursor) (Line, Line, bool) {
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return Line{}, Line{}, false
	}

	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return Line{}, Line{}, false
	}

	derived := !line.IsPhysicalLineStart(c.Source)
	if derived && indentBytes > 0 {
		return Line{}, Line{}, false
	}

	s := c.Source.Slice(line.Span)
	pos := indentBytes

	if pos >= len(s) || s[pos] != '>' {
		return Line{}, Line{}, false
	}

	full := c.MustNext()

	pos++

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

// OLMarkerLineResult captures the parsed structure of an ordered list
// marker line and the derived positions used to parse its item body.
type OLMarkerLineResult struct {
	MarkerLine      Line
	ContentLine     Line
	ListIndentCols  int
	ItemContentCols int
	MarkerDelim     byte
	StartNumber     int
}

// OrderedListRule parses ordered list blocks, collecting list items and
// recursively building their contents.
type OrderedListRule struct{}

func (r OrderedListRule) Apply(c *Cursor) (ir.Block, bool, error) {
	result, ok := r.tryConsumeFirstItem(c)
	if !ok {
		return nil, false, nil
	}

	listItems := make([]ir.ListItem, 0, 4)
	tight := true
	start := result.StartNumber

	for {
		lines, spans, keptBlank := r.consumeItemBody(c, result)
		if keptBlank {
			tight = false
		}

		children, err := buildBlocks(c.Source, c.Rules, lines, 0, c.Metadata)
		if err != nil {
			return nil, false, err
		}

		// defensive panic
		if len(spans) == 0 {
			panic("ordered list invariant violated: consumed marker line but produced no item spans")
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
		panic("ordered list invariant violated: matched first item but produced no items")
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
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return OLMarkerLineResult{}, false
	}

	relIndentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || relIndentCols > MaxValidIndentation {
		return OLMarkerLineResult{}, false
	}

	listIndentCols, _ := c.AbsBlockIndent(line)

	return r.tryParseMarkerLine(c, line, listIndentCols, indentBytes)
}

// tryConsumeSiblingItem attempts to consume a subsequent list item at the
// same indentation and with matching delimiter, rolling back on failure.
func (r OrderedListRule) tryConsumeSiblingItem(c *Cursor, listIndentCols int, markerDelim byte) (OLMarkerLineResult, bool, bool) {
	m := c.Mark()
	consumedBlanks := false

	line, ok := c.Peek()
	if !ok {
		return OLMarkerLineResult{}, false, false
	}

	for line.IsBlankLine(c.Source) {
		c.MustNext()
		consumedBlanks = true

		line, ok = c.Peek()
		if !ok {
			c.Reset(m)
			return OLMarkerLineResult{}, false, false
		}
	}

	absIndentCols, indentBytes := c.AbsBlockIndent(line)
	if absIndentCols != listIndentCols {
		c.Reset(m)
		return OLMarkerLineResult{}, false, false
	}

	result, ok := r.tryParseMarkerLine(c, line, listIndentCols, indentBytes)
	if !ok || result.MarkerDelim != markerDelim {
		c.Reset(m)
		return OLMarkerLineResult{}, false, false
	}

	return result, consumedBlanks, true
}

// consumeItemBody collects the lines belonging to a list item, rebasing
// content lines to the item baseline and handling trailing blank runs.
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
		nextLine, ok := c.Peek()
		if !ok {
			break
		}

		if nextLine.IsBlankLine(c.Source) {
			if !blankRun.active {
				blankRun.active = true
				blankRun.cursorMark = c.Mark()
				blankRun.spanMark = len(itemSpans)
				blankRun.lineMark = len(itemLines)
			}

			line := c.MustNext()
			itemSpans = append(itemSpans, line.Span)
			itemLines = append(itemLines, line)

			continue
		}

		absIndentCols, _ := c.AbsBlockIndent(nextLine)

		if absIndentCols >= start.ItemContentCols {
			if blankRun.active {
				keptBlank = true
			}

			blankRun.active = false

			line := c.MustNext()
			itemSpans = append(itemSpans, line.Span)

			trimmed := line.TrimIndentToCols(c.Source, start.ItemContentCols)
			itemLines = append(itemLines, trimmed)

			continue
		}

		if blankRun.active {
			blankRun.active = false

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

	if num > 1e9 {
		return OLMarkerLineResult{}, false
	}

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

	if pos >= len(s) {
		return OLMarkerLineResult{}, false
	}

	switch s[pos] {
	case ' ', '\t':
	// ok, continue

	default:
		return OLMarkerLineResult{}, false
	}

	markerLine := c.MustNext()

	for pos < len(s) {
		b := s[pos]

		if b == ' ' {
			col++
			pos++
			continue
		}

		if b == '\t' {
			col += source.TabWidth - (col % source.TabWidth)
			pos++
			continue
		}

		break
	}

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

// ULMarkerLineResult captures the parsed structure of an unordered list
// marker line and the derived positions used to parse its item body.
type ULMarkerLineResult struct {
	MarkerLine      Line
	ContentLine     Line
	ListIndentCols  int
	ItemContentCols int
}

// UnorderedListRule parses unordered list blocks, collecting list items
// and recursively building their contents.
type UnorderedListRule struct{}

func (r UnorderedListRule) Apply(c *Cursor) (ir.Block, bool, error) {
	result, ok := r.tryConsumeFirstItem(c)
	if !ok {
		return nil, false, nil
	}

	listItems := make([]ir.ListItem, 0, 4)
	tight := true

	for {
		lines, spans, keptBlank := r.consumeItemBody(c, result)
		if keptBlank {
			tight = false
		}

		children, err := buildBlocks(c.Source, c.Rules, lines, 0, c.Metadata)
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
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return ULMarkerLineResult{}, false
	}

	relIndentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || relIndentCols > MaxValidIndentation {
		return ULMarkerLineResult{}, false
	}

	listIndentCols, _ := c.AbsBlockIndent(line)

	return r.tryParseMarkerLine(c, line, listIndentCols, indentBytes)
}

// tryConsumeSiblingItem attempts to consume a subsequent list item at the
// same indentation, rolling back on failure.
func (r UnorderedListRule) tryConsumeSiblingItem(c *Cursor, listIndentCols int) (ULMarkerLineResult, bool, bool) {
	m := c.Mark()
	consumedBlanks := false

	line, ok := c.Peek()
	if !ok {
		return ULMarkerLineResult{}, false, false
	}

	for line.IsBlankLine(c.Source) {
		c.MustNext()
		consumedBlanks = true

		line, ok = c.Peek()
		if !ok {
			c.Reset(m)
			return ULMarkerLineResult{}, false, false
		}
	}

	absIndentCols, indentBytes := c.AbsBlockIndent(line)
	if absIndentCols != listIndentCols {
		c.Reset(m)
		return ULMarkerLineResult{}, false, false
	}

	result, ok := r.tryParseMarkerLine(c, line, listIndentCols, indentBytes)
	if !ok {
		c.Reset(m)
		return ULMarkerLineResult{}, false, false
	}

	return result, consumedBlanks, true
}

// consumeItemBody collects the lines belonging to a list item, rebasing
// content lines to the item baseline and handling trailing blank runs.
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
		nextLine, ok := c.Peek()
		if !ok {
			break
		}

		if nextLine.IsBlankLine(c.Source) {
			if !blankRun.active {
				blankRun.active = true
				blankRun.cursorMark = c.Mark()
				blankRun.spanMark = len(itemSpans)
				blankRun.lineMark = len(itemLines)
			}

			line := c.MustNext()
			itemSpans = append(itemSpans, line.Span)
			itemLines = append(itemLines, line)

			continue
		}

		absIndentCols, _ := c.AbsBlockIndent(nextLine)

		if absIndentCols >= start.ItemContentCols {
			if blankRun.active {
				keptBlank = true
			}

			blankRun.active = false

			line := c.MustNext()
			itemSpans = append(itemSpans, line.Span)

			trimmed := line.TrimIndentToCols(c.Source, start.ItemContentCols)
			itemLines = append(itemLines, trimmed)

			continue
		}

		if blankRun.active {
			blankRun.active = false

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

	if pos >= len(s) {
		return ULMarkerLineResult{}, false
	}

	switch s[pos] {
	case '-', '*', '+':
	// ok, continue

	default:
		return ULMarkerLineResult{}, false
	}

	pos++
	col++

	if pos >= len(s) {
		return ULMarkerLineResult{}, false
	}

	switch s[pos] {
	case ' ', '\t':
	// ok, continue

	default:
		return ULMarkerLineResult{}, false
	}

	markerLine := c.MustNext()

	for pos < len(s) {
		b := s[pos]

		if b == ' ' {
			col++
			pos++
			continue
		}

		if b == '\t' {
			col += source.TabWidth - (col % source.TabWidth)
			pos++
			continue
		}

		break
	}

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

// HeaderRule parses ATX headings.
type HeaderRule struct{}

func (r HeaderRule) Apply(c *Cursor) (ir.Block, bool, error) {
	level, contentSpan, ok := r.tryParseHeaderLine(c)
	if !ok {
		return nil, false, nil
	}

	line := c.MustNext()

	applied := ir.Header{
		Span:         line.Span,
		ContentSpan:  contentSpan,
		ContentLines: []source.ByteSpan{contentSpan},
		Level:        level,
	}

	return applied, true, nil
}

func (HeaderRule) tryParseHeaderLine(c *Cursor) (int, source.ByteSpan, bool) {
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return 0, source.ByteSpan{}, false
	}

	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return 0, source.ByteSpan{}, false
	}

	s := c.Source.Slice(line.Span)
	pos := indentBytes
	level := 0

	if pos >= len(s) || s[pos] != '#' {
		return 0, source.ByteSpan{}, false
	}

	for pos < len(s) && s[pos] == '#' {
		pos++
		level++
		if level == 7 {
			return 0, source.ByteSpan{}, false
		}
	}

	if pos < len(s) && s[pos] != ' ' && s[pos] != '\t' {
		return 0, source.ByteSpan{}, false
	}

	fieldStart := pos
	field := s[fieldStart:]

	fieldEnd := len(strings.TrimRight(field, " \t"))
	if end, ok := atxContentEnd(field); ok {
		fieldEnd = end
	}

	lead := 0
	for lead < fieldEnd && (field[lead] == ' ' || field[lead] == '\t') {
		lead++
	}

	contentStart := line.Span.Start + source.BytePos(fieldStart+lead)
	contentEnd := line.Span.Start + source.BytePos(fieldStart+fieldEnd)

	contentSpan := source.ByteSpan{
		Start: contentStart,
		End:   contentEnd,
	}

	return level, contentSpan, true
}

// atxContentEnd reports the end offset of the heading content field after
// removing a valid closing ATX marker run, if present.
func atxContentEnd(s string) (int, bool) {
	if len(s) == 0 {
		return 0, false
	}

	pos := len(s) - 1

	for pos >= 0 && (s[pos] == ' ' || s[pos] == '\t') {
		pos--
	}

	markerConsumed := false
	for pos >= 0 && s[pos] == '#' {
		if isEscaped(s, pos) {
			break
		}
		markerConsumed = true
		pos--
	}

	if !markerConsumed {
		return 0, false
	}

	sepStart := pos
	for pos >= 0 && (s[pos] == ' ' || s[pos] == '\t') {
		pos--
	}

	if pos == sepStart {
		return 0, false
	}

	return pos + 1, true
}

// isEscaped reports whether s[i] is escaped by an odd-length run of
// preceding backslashes.
func isEscaped(s string, i int) bool {
	slashes := 0
	for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
		slashes++
	}

	return slashes%2 == 1
}

// ThematicBreakRule parses thematic break lines (horizontal rules).
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

// tryParseThematicBreakLine reports whether the current line forms a valid
// thematic break.
func (ThematicBreakRule) tryParseThematicBreakLine(c *Cursor) bool {
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return false
	}

	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return false
	}

	s := c.Source.Slice(line.Span)
	pos := indentBytes

	if pos >= len(s) {
		return false
	}

	var marker byte
	switch s[pos] {
	case '-', '*', '_':
		marker = s[pos]

	default:
		return false
	}

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

// IndentedCodeBlockRule parses indented code blocks.
//
// It is paragraph-transparent, so it does not interrupt an in-progress
// paragraph.
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

// consumeIndentedCodeBlock collects the contiguous lines of an indented
// code block, rolling back trailing blank lines that are not followed by
// additional code block content.
func (r IndentedCodeBlockRule) consumeIndentedCodeBlock(c *Cursor) ([]source.ByteSpan, bool) {
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return nil, false
	}

	if !r.tryParseIndentedCodeBlockLine(c, line) {
		return nil, false
	}

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
		nextLine, ok := c.Peek()
		if !ok {
			break
		}

		if nextLine.IsBlankLine(c.Source) {
			if !blankRun.active {
				blankRun.active = true
				blankRun.cursorMark = c.Mark()
				blankRun.lineMark = len(lineSpans)
			}

			line := c.MustNext()
			lineSpans = append(lineSpans, line.Span)

			continue
		}

		if r.tryParseIndentedCodeBlockLine(c, nextLine) {
			blankRun.active = false

			line := c.MustNext()
			lineSpans = append(lineSpans, line.Span)

			continue
		}

		if blankRun.active {
			blankRun.active = false

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

// tryParseIndentedCodeBlockLine reports whether line satisfies the
// indentation requirement for an indented code block.
func (IndentedCodeBlockRule) tryParseIndentedCodeBlockLine(c *Cursor, line Line) bool {
	indentCols, _, ok := c.RelBlockIndent(line)
	if !ok || indentCols < MinValidCodeBlockIndentation {
		return false
	}

	return true
}

// FCBMarkerLineResult captures the parsed structure of a fenced code block
// opening fence.
type FCBMarkerLineResult struct {
	Marker         byte
	MarkerCount    int
	OpenIndentCols int
	InfoString     source.ByteSpan
}

// FencedCodeBlockRule parses fenced code blocks and collects their payload
// lines until a matching closing fence or EOF.
type FencedCodeBlockRule struct{}

func (r FencedCodeBlockRule) Apply(c *Cursor) (ir.Block, bool, error) {
	result, ok := r.tryParseOpeningFenceLine(c)
	if !ok {
		return nil, false, nil
	}

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
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return FCBMarkerLineResult{}, false
	}

	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return FCBMarkerLineResult{}, false
	}

	s := c.Source.Slice(line.Span)
	pos := indentBytes

	if pos >= len(s) {
		return FCBMarkerLineResult{}, false
	}

	var marker byte
	switch s[pos] {
	case '`', '~':
		marker = s[pos]

	default:
		return FCBMarkerLineResult{}, false
	}

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

	if markerCount < 3 {
		return FCBMarkerLineResult{}, false
	}

	for pos < len(s) {
		b := s[pos]
		if b == ' ' || b == '\t' {
			pos++
			continue
		}

		break
	}

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

// consumeFencedCodeBlock consumes the opening fence and subsequent payload
// lines, stopping at a matching closing fence or EOF.
func (r FencedCodeBlockRule) consumeFencedCodeBlock(c *Cursor, opener FCBMarkerLineResult) (source.ByteSpan, []source.ByteSpan) {
	line := c.MustNext()
	blockSpanStart := line.Span.Start
	blockSpanEnd := line.Span.End

	lineSpans := []source.ByteSpan{}
	for {
		line, ok := c.Peek()
		if !ok {
			break
		}

		if r.tryParseClosingFenceLine(c, opener.Marker, opener.MarkerCount) {
			line = c.MustNext()
			blockSpanEnd = line.Span.End

			break
		}

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

// tryParseClosingFenceLine reports whether the current line is a valid
// closing fence matching the given marker family and minimum run length.
func (FencedCodeBlockRule) tryParseClosingFenceLine(c *Cursor, marker byte, markerCount int) bool {
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return false
	}

	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return false
	}

	s := c.Source.Slice(line.Span)
	pos := indentBytes

	if pos >= len(s) {
		return false
	}

	if s[pos] != marker {
		return false
	}

	currentCount := 0
	for pos < len(s) && s[pos] == marker {
		pos++
		currentCount++
	}

	if currentCount < markerCount {
		return false
	}

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

// HTMLBlockRule parses block-level HTML constructs and consumes their
// lines according to the recognized block terminator.
type HTMLBlockRule struct{}

func (r HTMLBlockRule) Apply(c *Cursor) (ir.Block, bool, error) {
	terminator, ok := r.tryParseHTMLBlockLine(c)
	if !ok {
		return nil, false, nil
	}

	lineSpans := r.consumeHTMLBlock(c, terminator)

	// defensive panic
	if len(lineSpans) == 0 {
		panic("html block invariant violated: matched opening prefix but produced no payload")
	}

	blockSpan := source.ByteSpan{
		Start: lineSpans[0].Start,
		End:   lineSpans[len(lineSpans)-1].End,
	}

	applied := ir.HTMLBlock{
		Span:  blockSpan,
		Lines: lineSpans,
	}

	return applied, true, nil
}

// tryParseHTMLBlockLine classifies the current line as an HTML block
// opener and returns the corresponding terminator, if any.
func (r HTMLBlockRule) tryParseHTMLBlockLine(c *Cursor) (string, bool) {
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return "", false
	}

	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return "", false
	}

	s := c.Source.Slice(line.Span)
	pos := indentBytes

	if pos >= len(s) || s[pos] != '<' {
		return "", false
	}

	rest := s[pos:]
	var terminator string

	switch {
	case strings.HasPrefix(rest, "<!--"):
		terminator = "-->"

	case strings.HasPrefix(rest, "<![CDATA["):
		terminator = "]]>"

	case strings.HasPrefix(rest, "<?"):
		terminator = "?>"

	case strings.HasPrefix(rest, "<!"):
		terminator = ">"

	default:
		if !r.tryParseNamedTagLine(rest) {
			return "", false
		}
	}

	return terminator, true
}

// tryParseNamedTagLine reports whether s begins with a whitelisted HTML
// tag opener or closer eligible for named-tag HTML block parsing.
func (HTMLBlockRule) tryParseNamedTagLine(s string) bool {
	pos := 0

	if pos >= len(s) {
		return false
	}

	if s[pos] != '<' {
		return false
	}
	pos++

	if pos >= len(s) {
		return false
	}

	if s[pos] == '/' {
		pos++
	}

	if pos >= len(s) || !isAlpha(s[pos]) {
		return false
	}

	start := pos

	for pos < len(s) && (isAlpha(s[pos]) || isDigit(s[pos])) {
		pos++
	}

	name := strings.ToLower(s[start:pos])

	if !validateTagName(name) {
		return false
	}

	if pos >= len(s) {
		return false
	}

	switch s[pos] {
	case '>':
		return true

	case '/':
		pos++

		for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
			pos++
		}

		if pos >= len(s) || s[pos] != '>' {
			return false
		}

		return true

	case ' ', '\t':
		for pos < len(s) {
			if s[pos] == '>' {
				return true
			}
			pos++
		}

		return false

	default:
		return false
	}
}

// consumeHTMLBLock consumes an HTML block beginning at the current line.
//
// For fixed-terminator forms, consumption continues through the line
// containing the terminator. For named-tag forms, consumption continues
// until EOF or the next blank line.
func (HTMLBlockRule) consumeHTMLBlock(c *Cursor, terminator string) []source.ByteSpan {
	lineSpans := make([]source.ByteSpan, 0, 4)

	for {
		line := c.MustNext()
		lineSpans = append(lineSpans, line.Span)

		s := c.Source.Slice(line.Span)

		if terminator != "" && strings.Contains(s, terminator) {
			break
		}

		nextLine, ok := c.Peek()
		if !ok {
			break
		}

		if terminator == "" && nextLine.IsBlankLine(c.Source) {
			break
		}
	}

	return lineSpans
}

func isAlpha(b byte) bool {
	return 'A' <= b && b <= 'Z' || 'a' <= b && b <= 'z'
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// validateTagName reports whether name is recognized as a block-level HTML
// tag for named-tag HTML block parsing.
func validateTagName(name string) bool {
	_, ok := htmlBlockTags[name]
	return ok
}

// ParagraphRule parses paragraph runs and promotes them to setext headings
// when followed by a valid underline line.
type ParagraphRule struct{}

func (ParagraphRule) isParagraphTransparent() {}

func (r ParagraphRule) Apply(c *Cursor) (ir.Block, bool, error) {
	lineSpans, ok, err := r.consumeParagraphRun(c)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}

	contentSpan := source.ByteSpan{
		Start: lineSpans[0].Start,
		End:   lineSpans[len(lineSpans)-1].End,
	}

	if line, ok := c.Peek(); ok {
		level, isSetext := r.tryParseSetextHeadingLine(c, line)
		if isSetext {
			underline := c.MustNext()

			headerSpan := source.ByteSpan{
				Start: lineSpans[0].Start,
				End:   underline.Span.End,
			}

			applied := ir.Header{
				Span:         headerSpan,
				ContentSpan:  contentSpan,
				ContentLines: lineSpans,
				Level:        level,
			}

			return applied, true, nil
		}
	}

	applied := ir.Paragraph{
		Span:  contentSpan,
		Lines: lineSpans,
	}

	return applied, true, nil
}

// consumeParagraphRun collects consecutive paragraph lines until a blank
// line, a setext underline, or a paragraph-interrupting block is reached.
func (r ParagraphRule) consumeParagraphRun(c *Cursor) ([]source.ByteSpan, bool, error) {
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return nil, false, nil
	}

	line = c.MustNext()
	lineSpans := []source.ByteSpan{line.Span}

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

		lineSpans = append(lineSpans, line.Span)
	}

	return lineSpans, true, nil

}

// tryParseSetextHeadingLine reports whether line is a valid setext heading
// underline and returns the corresponding header level.
func (ParagraphRule) tryParseSetextHeadingLine(c *Cursor, line Line) (int, bool) {
	if line.IsBlankLine(c.Source) {
		return 0, false
	}

	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return 0, false
	}

	s := c.Source.Slice(line.Span)
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
