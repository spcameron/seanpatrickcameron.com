package block

import (
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/reference"
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

type ReferenceDefinitionRule struct{}

func (r ReferenceDefinitionRule) isParagraphTransparent() {}

func (r ReferenceDefinitionRule) Apply(c *Cursor) (ir.Block, bool, error) {
	// peek next line, reject if EOF or blank
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return nil, false, nil
	}

	// count the leading indentation, reject if greater than 3 visual columns
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

	// validate the first byte is an opening square bracket
	if s[pos] != '[' {
		return nil, false, nil
	}

	labelSpanStart := pos
	pos++

	// probe ahead for an unescaped closing square bracket
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

	// validate the link label
	if ok := reference.ValidateLabel(labelContent); !ok {
		return nil, false, nil
	}

	// normalize the link label for use as key in definitions map
	normalizedLabel := reference.NormalizeLabel(labelContent)

	// validate that the link label is immediately followed by a colon
	if s[pos] != ':' {
		return nil, false, nil
	}
	pos++

	// consume any optional spaces or tabs
	pos = consumeSpacesTabs(s, pos)
	if pos >= len(s) {
		return nil, false, nil
	}

	// validate the link destination
	destinationSpanRel, pos, ok := tryLinkDestination(s, pos)
	if !ok {
		return nil, false, nil
	}

	destinationSpan := source.ByteSpan{
		Start: lineBase + destinationSpanRel.Start,
		End:   lineBase + destinationSpanRel.End,
	}

	// mark the position and consume any spaces or tabs
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

	// attempt to validate and parse an (optional) link title
	if s[pos] == '"' || s[pos] == '\'' || s[pos] == '(' {
		// if no separator exists between destination and title, the definition is invalid
		if !sepPresent {
			return nil, false, nil
		}

		// try to parse the link title
		titleSpanRel, pos, ok := tryLinkTitle(s, pos)
		if !ok {
			return nil, false, nil
		}

		titleSpan := source.ByteSpan{
			Start: lineBase + titleSpanRel.Start,
			End:   lineBase + titleSpanRel.End,
		}

		// consume any spaces or tabs
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

	// trailing non-title content after destination makes this not a valid definition
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
	// validate that the byte at pos is '<'
	if pos >= len(s) || s[pos] != '<' {
		return source.ByteSpan{}, 0, false
	}

	// advance past the opening angle bracket
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
			// on a backslash, advance two bytes if within span limit
			if pos+1 < len(s) {
				pos += 2
				continue
			}
			// otherwise, trailing backslash is just ordinary content
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
			// escaped byte
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
			// on a backslash, advance two bytes if possible
			if pos+1 < len(s) {
				pos += 2
				continue
			}
			// otherwise, trailing backslash is just ordinary content
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
			// an unescaped open paren increases the paren depth
			depth++
			pos++

		case ')':
			// an unescaped close paren decreases the paren depth
			depth--

			// reaching depth 0 ends the title
			if depth == 0 {
				span := source.ByteSpan{
					Start: source.BytePos(start),
					End:   source.BytePos(pos),
				}

				return span, pos + 1, true
			}

			pos++

		case '\\':
			// on a backslash, advance two bytes if possible
			if pos+1 < len(s) {
				pos += 2
				continue
			}
			// otherwise, trailing backslash is just ordinary content
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

		children, err := buildBlocks(c.Source, c.Rules, lines, result.ItemContentCols, c.Metadata)
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
			col += source.TabWidth - (col % source.TabWidth)
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

		children, err := buildBlocks(c.Source, c.Rules, lines, result.ItemContentCols, c.Metadata)
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
			col += source.TabWidth - (col % source.TabWidth)
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
		Span:         line.Span,
		ContentSpan:  contentSpan,
		ContentLines: []source.ByteSpan{contentSpan},
		Level:        level,
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

type HTMLBlockRule struct{}

func (r HTMLBlockRule) Apply(c *Cursor) (ir.Block, bool, error) {
	terminator, ok := r.tryParseHTMLBlockLine(c)
	if !ok {
		return nil, false, nil
	}

	lineSpans := r.consumeHTMLBLock(c, terminator)

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

func (r HTMLBlockRule) tryParseHTMLBlockLine(c *Cursor) (string, bool) {
	// peek first line, reject if EOF or blank
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return "", false
	}

	// count the leading indentation, reject if greater than 3 visual columns
	indentCols, indentBytes, ok := c.RelBlockIndent(line)
	if !ok || indentCols > MaxValidIndentation {
		return "", false
	}

	s := c.Source.Slice(line.Span)
	pos := indentBytes

	// validate the first marker character
	if pos >= len(s) || s[pos] != '<' {
		return "", false
	}

	rest := s[pos:]
	var terminator string

	// validate the prefix and determine the terminator
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
		// try named tag, and if that fails, reject the line
		if !r.tryParseNamedTagLine(rest) {
			return "", false
		}
	}

	return terminator, true
}

func (HTMLBlockRule) tryParseNamedTagLine(s string) bool {
	pos := 0

	// reject an empty string
	if pos >= len(s) {
		return false
	}

	// validate and consume < opener
	if s[pos] != '<' {
		return false
	}
	pos++

	// validate and consume an optional /
	if pos >= len(s) {
		return false
	}
	if s[pos] == '/' {
		pos++
	}

	// validate tag name starts with alpha character
	if pos >= len(s) || !isAlpha(s[pos]) {
		return false
	}

	// mark start of tag name
	start := pos

	// consume valid tag name characters (alphanumeric)
	for pos < len(s) && (isAlpha(s[pos]) || isDigit(s[pos])) {
		pos++
	}

	// normalize and record tag name
	name := strings.ToLower(s[start:pos])

	// reject if the tag name is not whitelisted
	if !validateTagName(name) {
		return false
	}

	// validate following byte forms plausible tag head
	if pos >= len(s) {
		return false
	}

	switch s[pos] {
	case '>':
		return true

	case '/':
		pos++

		// consume any whitespace
		for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
			pos++
		}

		// reject if the next non-whitespace character is not a closer
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
		// any other character is ineligible in a tag head
		return false
	}
}

func (HTMLBlockRule) consumeHTMLBLock(c *Cursor, terminator string) []source.ByteSpan {
	lineSpans := make([]source.ByteSpan, 0, 4)

	// consume maximal HTML block
	for {
		line := c.MustNext()
		lineSpans = append(lineSpans, line.Span)

		s := c.Source.Slice(line.Span)

		if terminator != "" && strings.Contains(s, terminator) {
			break
		}

		// peek next line, break if EOF
		nextLine, ok := c.Peek()
		if !ok {
			break
		}

		// if terminator is "" (from named-tag case), break on blank lines
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

func validateTagName(name string) bool {
	_, ok := htmlBlockTags[name]
	return ok
}

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

func (r ParagraphRule) consumeParagraphRun(c *Cursor) ([]source.ByteSpan, bool, error) {
	// peek next line, reject if EOF or blank line
	line, ok := c.Peek()
	if !ok || line.IsBlankLine(c.Source) {
		return nil, false, nil
	}

	// consume the first line
	line = c.MustNext()
	lineSpans := []source.ByteSpan{line.Span}

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

		lineSpans = append(lineSpans, line.Span)
	}

	return lineSpans, true, nil

}

func (ParagraphRule) tryParseSetextHeadingLine(c *Cursor, line Line) (int, bool) {
	// reject blank line
	if line.IsBlankLine(c.Source) {
		return 0, false
	}

	// count the leading indentation, reject if greater than 3 visual columns
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
