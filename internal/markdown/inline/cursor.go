package inline

import (
	"fmt"
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type Cursor struct {
	Source     *source.Source
	Span       source.ByteSpan
	Tokens     []Token
	Index      int
	Items      *ItemList
	Delimiters *DelimiterStack
}

func NewCursor(src *source.Source, span source.ByteSpan, tokens []Token) *Cursor {
	return &Cursor{
		Source: src,
		Span:   span,
		Tokens: tokens,
		Index:  0,
		Items:  NewItemList(),
	}
}

func (c *Cursor) Next() Token {
	out := c.Tokens[c.Index]
	c.Index++
	return out
}

func (c *Cursor) Build() ([]ast.Inline, error) {
	// traverse tokens and dispatch as needed
	for {
		token := c.Next()
		if token.Kind == TokenEOF {
			break
		}

		switch token.Kind {
		case TokenText:
			c.appendItemRecord(token.Span, ItemText)

		case TokenBacktick:
			c.handleTokenBacktick()

		case TokenOpenAngle:
			openerIdx := c.Index - 1
			openerToken := c.Tokens[openerIdx]

			// search for the first close angle token
			closerIdx := openerIdx + 1
			for closerIdx < len(c.Tokens) {
				next := c.Tokens[closerIdx]
				if next.Kind != TokenCloseAngle {
					closerIdx++
					continue
				}

				break
			}

			// if no close angle token found, append the open angle token as text
			if closerIdx == len(c.Tokens) {
				c.appendItemRecord(openerToken.Span, ItemText)
				continue
			}

			// define the outer span (including angle brackets)
			outerSpan := source.ByteSpan{
				Start: c.Tokens[openerIdx].Span.Start,
				End:   c.Tokens[closerIdx].Span.End,
			}

			// define the content span (excluding angle brackets)
			contentSpan := source.ByteSpan{
				Start: c.Tokens[openerIdx].Span.End,
				End:   c.Tokens[closerIdx].Span.Start,
			}
			contentSlice := c.Source.Slice(contentSpan)

			// check if the content span is a valid URI autolink
			if validateURI(contentSlice) {
				c.appendItemRecord(outerSpan, ItemAutolinkURI)
				c.Index = closerIdx + 1
				continue
			}

			// check if the content span is a valid email autolink
			if validateEmail(contentSlice) {
				c.appendItemRecord(outerSpan, ItemAutolinkEmail)
				c.Index = closerIdx + 1
				continue
			}

			// extract the span and string from the opening token to the end of the inline
			candidateSpan := source.ByteSpan{
				Start: openerToken.Span.Start,
				End:   c.Span.End,
			}
			s := c.Source.Slice(candidateSpan)

			// try to find a valid inline HTML construct, and otherwise return the token as plain text
			width, ok := tryInlineHTML(s)
			if !ok {
				c.appendItemRecord(openerToken.Span, ItemText)
			}

			// if valid inline HTML is found, update the candidate span to match
			candidateSpan.End = openerToken.Span.Start + source.BytePos(width)

			// recreate the closing angle bracket's span based on the candidate span
			targetSpan := source.ByteSpan{
				Start: candidateSpan.End - 1,
				End:   candidateSpan.End,
			}

			// locate the token for this closing angle bracket by matching spans
			candidateCloserIdx := -1
			for i := openerIdx; i < len(c.Tokens); i++ {
				tok := c.Tokens[i]
				if tok.Span == targetSpan {
					candidateCloserIdx = i
					break
				}
			}

			if candidateCloserIdx == -1 {
				panic("candidate closing index found during byte-traversal search, but no matching token could be found in the token stream")
			}

			c.appendItemRecord(candidateSpan, ItemHTML)
			c.Index = candidateCloserIdx + 1
			continue

		case TokenCloseAngle:
			c.appendItemRecord(token.Span, ItemText)

		default:
			panic(fmt.Sprintf("unknown token kind encountered (%d)", token.Kind))
		}
	}

	inlines := []ast.Inline{}

	// finalize inline items and construct ast.Inlines
	item := c.Items.Front()
	for item != nil {
		switch item.Kind {
		case ItemText:
			node := ast.Text{
				Span: item.LiveSpan,
			}

			inlines = append(inlines, node)

		case ItemCodeSpan:
			contentSlice := c.Source.Slice(item.OriginalSpan)

			if len(contentSlice) > 0 &&
				isSpace(contentSlice[0]) &&
				isSpace(contentSlice[len(contentSlice)-1]) &&
				!isAllSpaces(contentSlice) {
				item.LiveSpan.Start++
				item.LiveSpan.End--
			}

			node := ast.CodeSpan{
				Span: item.LiveSpan,
			}

			inlines = append(inlines, node)

		case ItemAutolinkURI:
			contentSpan := source.ByteSpan{
				Start: item.LiveSpan.Start + 1,
				End:   item.LiveSpan.End - 1,
			}

			node := ast.Link{
				Span:        item.LiveSpan,
				Destination: contentSpan,
				Children: []ast.Inline{
					ast.Text{
						Span: contentSpan,
					},
				},
			}

			inlines = append(inlines, node)

		case ItemAutolinkEmail:
			contentSpan := source.ByteSpan{
				Start: item.LiveSpan.Start + 1,
				End:   item.LiveSpan.End - 1,
			}

			node := ast.Link{
				Span:        item.LiveSpan,
				Destination: contentSpan,
				MailTo:      true,
				Children: []ast.Inline{
					ast.Text{
						Span: contentSpan,
					},
				},
			}

			inlines = append(inlines, node)

		case ItemHTML:
			node := ast.RawText{
				Span: item.LiveSpan,
			}

			inlines = append(inlines, node)

		default:
			panic(fmt.Sprintf("unknown item kind encountered (%d)", item.Kind))
		}

		item = item.Next()
	}

	return inlines, nil
}

func (c *Cursor) handleTokenBacktick() {
	openerIdx := c.Index - 1
	openerToken := c.Tokens[openerIdx]
	openerWidth := openerToken.Span.Width()

	// search for a closing backtick token with the same length as the opener
	closerIdx := openerIdx + 1
	for closerIdx < len(c.Tokens) {
		next := c.Tokens[closerIdx]
		if next.Kind != TokenBacktick {
			closerIdx++
			continue
		}

		if next.Span.Width() != openerWidth {
			closerIdx++
			continue
		}

		break
	}

	// if no matching closer is found, append the opening token as plain text and return
	if closerIdx >= len(c.Tokens) {
		c.appendItemRecord(openerToken.Span, ItemText)
		return
	}

	// define the exclusive span between the backtick delimiters
	contentSpan := source.ByteSpan{
		Start: c.Tokens[openerIdx].Span.End,
		End:   c.Tokens[closerIdx].Span.Start,
	}

	// append the item to the item record list and advance the index past the closer backtick token
	c.appendItemRecord(contentSpan, ItemCodeSpan)
	c.Index = closerIdx + 1
}

func (c *Cursor) appendItemRecord(span source.ByteSpan, kind ItemKind) *ItemRecord {
	item := &ItemRecord{
		OriginalSpan: span,
		LiveSpan:     span,
		Kind:         kind,
	}

	return c.Items.PushBack(item)
}

// func (c *Cursor) Gather() error {
// 	for {
// 		if c.EOF() {
// 			break
// 		}
//
// 		ev, ok := c.Next()
// 		if !ok {
// 			break
// 		}
//
// 		switch ev.Kind {
// 		case EventText:
// 			c.gatherText(ev)
//
// 		case EventDelimiterRun:
// 			c.gatherDelimiter(ev)
//
// 		case EventOpenBracket:
// 			c.gatherToken(ev, TokenOpenBracket)
//
// 		case EventCloseBracket:
// 			c.gatherCloseBracket(ev)
//
// 		case EventOpenParen:
// 			c.gatherToken(ev, TokenOpenParen)
//
// 		case EventCloseParen:
// 			c.gatherToken(ev, TokenCloseParen)
//
// 		case EventIllegalNewline:
// 			panic("illegal newline encountered during inline gather")
//
// 		default:
// 			panic("unhandled inline event kind in gather")
// 		}
// 	}
//
// 	return nil
// }

// func (c *Cursor) Resolve() error {
// 	snapshot := slices.Clone(c.WorkingItems)
//
// 	if err := c.resolveEmphasis(); err != nil {
// 		return err
// 	}
//
// 	if err := c.resolveInlineLinks(snapshot); err != nil {
// 		return err
// 	}
//
// 	return nil
// }

// func (c *Cursor) Finalize() ([]ast.Inline, error) {
// 	inlines := make([]ast.Inline, 0, len(c.WorkingItems))
//
// 	for i, item := range c.WorkingItems {
// 		inl, ok, err := c.finalizeItem(i, item)
// 		if err != nil {
// 			return []ast.Inline{}, err
// 		}
// 		if !ok {
// 			continue
// 		}
//
// 		inlines = append(inlines, inl)
// 	}
//
// 	return inlines, nil
// }

// func (c *Cursor) gatherText(ev Event) {
// 	item := &TextItem{
// 		Span: ev.Span,
// 	}
//
// 	c.WorkingItems = append(c.WorkingItems, item)
// }

// func (c *Cursor) gatherDelimiter(ev Event) {
// 	item := &DelimiterItem{
// 		Span:      ev.Span,
// 		Delimiter: ev.Delimiter,
// 	}
//
// 	c.WorkingItems = append(c.WorkingItems, item)
// 	idx := len(c.WorkingItems) - 1
//
// 	canOpen, canClose := c.delimiterEligibility(ev.Span)
//
// 	record := &DelimiterRecord{
// 		OriginalSpan: ev.Span,
// 		LiveSpan:     ev.Span,
// 		Delimiter:    ev.Delimiter,
// 		OriginalRun:  ev.RunLength,
// 		RemainingRun: ev.RunLength,
// 		CanOpen:      canOpen,
// 		CanClose:     canClose,
// 		ItemIndex:    idx,
// 	}
//
// 	c.DelimiterRecords = append(c.DelimiterRecords, record)
// }

// func (c *Cursor) gatherToken(ev Event, t TokenKind) {
// 	item := &TokenItem{
// 		Span: ev.Span,
// 		Kind: t,
// 	}
//
// 	c.WorkingItems = append(c.WorkingItems, item)
// }

// func (c *Cursor) gatherCloseBracket(ev Event) {
// 	c.gatherToken(ev, TokenCloseBracket)
//
// 	idx := len(c.WorkingItems) - 1
//
// 	record := &BracketRecord{
// 		Span:      ev.Span,
// 		ItemIndex: idx,
// 		Active:    true,
// 	}
//
// 	c.BracketRecords = append(c.BracketRecords, record)
// }

// func (c *Cursor) delimiterEligibility(span source.ByteSpan) (canOpen, canClose bool) {
// 	before, beforeOK := c.runeBefore(span)
// 	after, afterOK := c.runeAfter(span)
//
// 	canOpen = leftFlanking(before, beforeOK, after, afterOK)
// 	canClose = rightFlanking(before, beforeOK, after, afterOK)
//
// 	return canOpen, canClose
// }

// func (c *Cursor) resolveEmphasis() error {
// 	for closerIdx := 0; closerIdx < len(c.DelimiterRecords); closerIdx++ {
// 		closerRecord := c.DelimiterRecords[closerIdx]
//
// 		for closerRecord.CanClose && closerRecord.RemainingRun > 0 {
// 			openerIdx, ok := c.findOpenerForCloserDelimiter(closerIdx)
// 			if !ok {
// 				break
// 			}
//
// 			use := c.delimiterUse(openerIdx, closerIdx)
// 			didReduce, err := c.resolveDelimiterPair(openerIdx, closerIdx, use)
// 			if err != nil {
// 				return err
// 			}
// 			if !didReduce {
// 				break
// 			}
// 		}
// 	}
//
// 	return nil
// }

// func (c *Cursor) findOpenerForCloserDelimiter(closerIdx int) (int, bool) {
// 	closer := c.DelimiterRecords[closerIdx]
//
// 	for openerIdx := closerIdx - 1; openerIdx >= 0; openerIdx-- {
// 		d := c.DelimiterRecords[openerIdx]
// 		if d.Delimiter == closer.Delimiter && d.CanOpen && d.RemainingRun > 0 {
// 			return openerIdx, true
// 		}
// 	}
//
// 	return -1, false
// }

// func (c *Cursor) delimiterUse(openerIdx, closerIdx int) int {
// 	use := 1
// 	if c.DelimiterRecords[openerIdx].RemainingRun >= 2 &&
// 		c.DelimiterRecords[closerIdx].RemainingRun >= 2 {
// 		use = 2
// 	}
//
// 	return use
// }

// func (c *Cursor) resolveDelimiterPair(openerIdx, closerIdx, use int) (bool, error) {
// 	// retrieve the delimiter records
// 	openerRecord := c.DelimiterRecords[openerIdx]
// 	closerRecord := c.DelimiterRecords[closerIdx]
//
// 	// retrieve the corresponding working item indexes
// 	openItemIdx := openerRecord.ItemIndex
// 	closeItemIdx := closerRecord.ItemIndex
//
// 	// determine the node span based on the consumed spans
// 	nodeSpan := source.ByteSpan{
// 		Start: openerRecord.LiveSpan.End - source.BytePos(use),
// 		End:   closerRecord.LiveSpan.Start + source.BytePos(use),
// 	}
//
// 	// build the children ast inline nodes and their cumulative span
// 	children, childSpan := c.buildChildren(openItemIdx+1, closeItemIdx)
//
// 	// if no content between delimiters, do not reduce
// 	if len(children) == 0 {
// 		return false, nil
// 	}
//
// 	// build the inline node with the content-only span
// 	var node ast.Inline
// 	switch use {
// 	case 1:
// 		node = ast.Em{
// 			Span:     childSpan,
// 			Children: children,
// 		}
// 	case 2:
// 		node = ast.Strong{
// 			Span:     childSpan,
// 			Children: children,
// 		}
// 	default:
// 		panic("resolvePair: unsupported delimiter consumption size")
// 	}
//
// 	// consume from the matched delimiter records
// 	openerRecord.RemainingRun -= use
// 	closerRecord.RemainingRun -= use
//
// 	// trim live spans according to their roles
// 	// opener consumes from the right
// 	openerRecord.LiveSpan.End -= source.BytePos(use)
//
// 	// closer consumes from the left
// 	closerRecord.LiveSpan.Start += source.BytePos(use)
//
// 	// rewrite working items without changing indices
// 	//
// 	// anchor the resolved node at the first interior slot
// 	anchor := openItemIdx + 1
// 	c.WorkingItems[anchor] = &NodeItem{
// 		Span: nodeSpan,
// 		Node: node,
// 	}
//
// 	// any other interior slots absorbed by the node are marked consumed
// 	for i := anchor + 1; i < closeItemIdx; i++ {
// 		err := c.consumeWorkingItemAt(i)
// 		if err != nil {
// 			return true, err
// 		}
// 	}
//
// 	// if either delimiter run has been fully exhausted, it is also consumed
// 	// otherwise it remains a live DelimiterItem with a narrower LiveSpan
// 	if openerRecord.RemainingRun == 0 {
// 		err := c.consumeWorkingItemAt(openItemIdx)
// 		if err != nil {
// 			return true, err
// 		}
// 	}
//
// 	if closerRecord.RemainingRun == 0 {
// 		err := c.consumeWorkingItemAt(closeItemIdx)
// 		if err != nil {
// 			return true, err
// 		}
// 	}
//
// 	return true, nil
// }

// func (c *Cursor) resolveInlineLinks(snapshot []WorkingItem) error {
// 	for closerIdx := 0; closerIdx < len(c.BracketRecords); closerIdx++ {
// 		// extract the next bracket record and ensure it's active
// 		rec := c.BracketRecords[closerIdx]
// 		if !rec.Active {
// 			continue
// 		}
// 		closeBracketIdx := rec.ItemIndex
//
// 		// extract the corresponding working closeBracketItem record from the snapshot
// 		closeBracketItem := snapshot[closeBracketIdx]
//
// 		// verify the close item is a closing bracket token
// 		closeToken, ok := closeBracketItem.(*TokenItem)
// 		if !ok || closeToken.Kind != TokenCloseBracket {
// 			continue
// 		}
//
// 		// attempt to parse a valid tail from the next live item
// 		tail, ok, err := c.tryParseInlineLinkTail(snapshot, closeBracketIdx)
// 		if err != nil {
// 			return err
// 		}
// 		if !ok {
// 			continue
// 		}
//
// 		// search backward for nearest live TokenOpenBracket
// 		openBracketIdx, ok := c.findOpenerForCloserBracket(snapshot, closeBracketIdx)
// 		if !ok {
// 			continue
// 		}
//
// 		// extract the corresponding working openBracketItem record from the snapshot
// 		openBracketItem := snapshot[openBracketIdx]
//
// 		// verify the open item is an opening bracket token
// 		openToken, ok := openBracketItem.(*TokenItem)
// 		if !ok || openToken.Kind != TokenOpenBracket {
// 			continue
// 		}
//
// 		// build label children, disregard the childspan
// 		children, _ := c.buildChildren(openBracketIdx+1, closeBracketIdx)
//
// 		// construct the full link span, tip-to-tail
// 		fullSpan := source.ByteSpan{
// 			Start: openToken.Span.Start,
// 			End:   tail.FullSpan.End,
// 		}
//
// 		// construct the label span directly from the opener and closer token spans
// 		labelSpan := source.ByteSpan{
// 			Start: openToken.Span.End,
// 			End:   closeToken.Span.Start,
// 		}
//
// 		// construct link node
// 		node := ast.Link{
// 			Span:        fullSpan,
// 			Label:       labelSpan,
// 			Destination: tail.DestinationSpan,
// 			Title:       tail.TitleSpan,
// 			Children:    children,
// 		}
//
// 		// anchor node at openItemIdx
// 		anchor := openBracketIdx
// 		c.WorkingItems[anchor] = &NodeItem{
// 			Span: node.Span,
// 			Node: node,
// 		}
//
// 		// mark all absorbed items through tail close as consumed
// 		for i := anchor + 1; i <= tail.CloseParenItemIndex; i++ {
// 			err := c.consumeWorkingItemAt(i)
// 			if err != nil {
// 				return err
// 			}
// 		}
//
// 		// deactivate affected bracket records inside consumed region
// 		for i := range c.BracketRecords {
// 			rec := c.BracketRecords[i]
// 			if rec.ItemIndex >= openBracketIdx && rec.ItemIndex <= tail.CloseParenItemIndex {
// 				rec.Active = false
// 			}
// 		}
// 	}
//
// 	return nil
// }

// func (c *Cursor) tryParseInlineLinkTail(snapshot []WorkingItem, closeItemIdx int) (InlineLinkTail, bool, error) {
// 	// initialize empty full span early, update Start and End when validated
// 	fullSpan := source.ByteSpan{}
// 	contentSpan := source.ByteSpan{}
//
// 	// verify that there exists an item past the closeItemIdx provided
// 	openParenIdx := closeItemIdx + 1
// 	if openParenIdx == len(snapshot) {
// 		return InlineLinkTail{}, false, nil
// 	}
//
// 	// verify that the next item is an open paren token
// 	openParenItem, ok := snapshot[openParenIdx].(*TokenItem)
// 	if !ok {
// 		return InlineLinkTail{}, false, nil
// 	}
// 	if openParenItem.Kind != TokenOpenParen {
// 		return InlineLinkTail{}, false, nil
// 	}
//
// 	fullSpan.Start = openParenItem.Span.Start
// 	contentSpan.Start = openParenItem.Span.End
//
// 	closeParenIdx := openParenIdx + 1
// 	for closeParenIdx < len(snapshot) {
// 		// inspect the next item, and break on the first close paren token encountered
// 		// NOTE: this behavior changes after V1
// 		closeParenItem, ok := snapshot[closeParenIdx].(*TokenItem)
// 		if ok && closeParenItem.Kind == TokenCloseParen {
// 			fullSpan.End = closeParenItem.Span.End
// 			contentSpan.End = closeParenItem.Span.Start
// 			break
// 		}
// 		closeParenIdx++
// 	}
//
// 	// if no close paren token encountered, report failure
// 	if closeParenIdx == len(snapshot) {
// 		return InlineLinkTail{}, false, nil
// 	}
//
// 	// parse the tail span for the destination and optional title
// 	s := c.Source.Slice(contentSpan)
// 	pos := 0
//
// 	// consume optional leading whitespace
// 	for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
// 		pos++
// 	}
//
// 	// if no valid destination, report failure
// 	if pos == len(s) {
// 		return InlineLinkTail{}, false, nil
// 	}
//
// 	// mark the start of the destination span
// 	destinationSpan := source.ByteSpan{
// 		Start: contentSpan.Start + source.BytePos(pos),
// 	}
//
// 	// reject a destination that begins with a quote
// 	if s[pos] == '\'' || s[pos] == '"' {
// 		return InlineLinkTail{}, false, nil
// 	}
//
// 	// traverse the tail span and break on the first white space
// 	// reject paren characters inside destinations
// 	// NOTE: this behavior changes after V1
// destLoop:
// 	for pos < len(s) {
// 		b := s[pos]
// 		switch b {
// 		case ' ', '\t':
// 			break destLoop
// 		case '(', ')', '\'', '"':
// 			return InlineLinkTail{}, false, nil
// 		default:
// 			pos++
// 		}
// 	}
//
// 	// record the end of the destination span
// 	destinationSpan.End = contentSpan.Start + source.BytePos(pos)
//
// 	// at this point, we have a legal destination prefix, so construct what we have
// 	tail := InlineLinkTail{
// 		OpenParenItemIndex:  openParenIdx,
// 		CloseParenItemIndex: closeParenIdx,
// 		FullSpan:            fullSpan,
// 		DestinationSpan:     destinationSpan,
// 	}
//
// 	// if at the end of content span, there is no title and the tail is valid
// 	if pos == len(s) {
// 		return tail, true, nil
// 	}
//
// 	// any remaining content must begin with whitespace separating destination from title
// 	if s[pos] != ' ' && s[pos] != '\t' {
// 		return InlineLinkTail{}, false, nil
// 	}
//
// 	// consume trailing whitespace after destination
// 	for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
// 		pos++
// 	}
//
// 	// if only trailing whitespace remains, the tail is valid with no title
// 	if pos == len(s) {
// 		return tail, true, nil
// 	}
//
// 	// remaining content must begin a quote title
// 	if s[pos] != '\'' && s[pos] != '"' {
// 		return InlineLinkTail{}, false, nil
// 	}
//
// 	quoteMarker := s[pos]
// 	quoteStart := pos
// 	pos++
//
// 	// traverse the string searching for a matching quote
// 	for pos < len(s) && s[pos] != quoteMarker {
// 		pos++
// 	}
//
// 	// if at end of content span, no matching closing quote was found
// 	if pos == len(s) {
// 		return InlineLinkTail{}, false, nil
// 	}
//
// 	quoteEnd := pos
// 	pos++
//
// 	// consume trailing whitespace
// 	for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
// 		pos++
// 	}
//
// 	// if not at end of span, break (reject junk after title)
// 	if pos != len(s) {
// 		return InlineLinkTail{}, false, nil
// 	}
//
// 	tail.TitleSpan = source.ByteSpan{
// 		Start: contentSpan.Start + source.BytePos(quoteStart+1),
// 		End:   contentSpan.Start + source.BytePos(quoteEnd),
// 	}
//
// 	return tail, true, nil
// }

// func (c *Cursor) findOpenerForCloserBracket(snapshot []WorkingItem, closerIdx int) (int, bool) {
// 	for openerIdx := closerIdx - 1; openerIdx >= 0; openerIdx-- {
// 		snapItem := snapshot[openerIdx]
// 		snapToken, ok := snapItem.(*TokenItem)
// 		if ok && snapToken.Kind == TokenOpenBracket {
// 			return openerIdx, true
// 		}
// 	}
//
// 	return -1, false
// }

// type InlineLinkTail struct {
// 	OpenParenItemIndex  int
// 	CloseParenItemIndex int
// 	FullSpan            source.ByteSpan
// 	DestinationSpan     source.ByteSpan
// 	TitleSpan           source.ByteSpan
// }
//
// func (c *Cursor) buildChildren(start, end int) ([]ast.Inline, source.ByteSpan) {
// 	children := make([]ast.Inline, 0, end-start)
//
// 	for i := start; i < end; i++ {
// 		switch v := c.WorkingItems[i].(type) {
// 		case *TextItem:
// 			inl := ast.Text{
// 				Span: v.Span,
// 			}
// 			children = append(children, inl)
//
// 		case *DelimiterItem:
// 			idx, ok := c.delimiterRecordForItem(i)
// 			if !ok {
// 				panic(fmt.Sprintf("no matching delimiter record found for the working item at index %d", i))
// 			}
//
// 			rec := c.DelimiterRecords[idx]
// 			if rec.RemainingRun > 0 {
// 				inl := ast.Text{
// 					Span: rec.LiveSpan,
// 				}
// 				children = append(children, inl)
// 			}
//
// 		case *TokenItem:
// 			inl := ast.Text{
// 				Span: v.Span,
// 			}
// 			children = append(children, inl)
//
// 		case *NodeItem:
// 			inl := v.Node
// 			children = append(children, inl)
//
// 		case *ConsumedItem:
// 			continue
//
// 		default:
// 			panic("unknown working item encountered")
// 		}
// 	}
//
// 	if len(children) == 0 {
// 		return children, source.ByteSpan{}
// 	}
//
// 	first, ok := inlineSpan(children[0])
// 	if !ok {
// 		panic("could not determine span of first child inline")
// 	}
//
// 	last, ok := inlineSpan(children[len(children)-1])
// 	if !ok {
// 		panic("could not determine span of last child inline")
// 	}
//
// 	span := source.ByteSpan{
// 		Start: first.Start,
// 		End:   last.End,
// 	}
//
// 	return children, span
// }

// func (c *Cursor) finalizeItem(i int, item WorkingItem) (ast.Inline, bool, error) {
// 	switch v := item.(type) {
// 	case *TextItem:
// 		return ast.Text{
// 			Span: v.Span,
// 		}, true, nil
//
// 	case *DelimiterItem:
// 		recIdx, ok := c.delimiterRecordForItem(i)
// 		if !ok {
// 			return nil, false, fmt.Errorf("no matching delimiter record found for working item at index %d", i)
// 		}
//
// 		rec := c.DelimiterRecords[recIdx]
// 		if rec.RemainingRun == 0 {
// 			return nil, false, nil
// 		}
//
// 		return ast.Text{
// 			Span: rec.LiveSpan,
// 		}, true, nil
//
// 	case *TokenItem:
// 		return ast.Text{
// 			Span: v.Span,
// 		}, true, nil
//
// 	case *NodeItem:
// 		return v.Node, true, nil
//
// 	case *ConsumedItem:
// 		return nil, false, nil
//
// 	default:
// 		return nil, false, fmt.Errorf("unknown working item type %T", item)
// 	}
// }

// func (c *Cursor) consumeWorkingItemAt(idx int) error {
// 	if idx < 0 || idx >= len(c.WorkingItems) {
// 		return fmt.Errorf("index %d is out of bounds", idx)
// 	}
//
// 	consumedItem := &ConsumedItem{}
// 	switch item := c.WorkingItems[idx].(type) {
// 	case *TextItem:
// 		consumedItem.Span = item.Span
//
// 	case *NodeItem:
// 		consumedItem.Span = item.Span
//
// 	case *TokenItem:
// 		consumedItem.Span = item.Span
//
// 	case *DelimiterItem:
// 		recordIdx, ok := c.delimiterRecordForItem(idx)
// 		if !ok {
// 			return fmt.Errorf("no matching delimiter record found for working item at index %d", idx)
// 		}
// 		rec := c.DelimiterRecords[recordIdx]
// 		consumedItem.Span = rec.LiveSpan
//
// 	case *ConsumedItem:
// 		consumedItem.Span = item.Span
//
// 	default:
// 		return fmt.Errorf("unknown working item type %T", item)
// 	}
//
// 	c.WorkingItems[idx] = consumedItem
// 	return nil
// }

// func (c *Cursor) delimiterRecordForItem(index int) (int, bool) {
// 	for i := 0; i < len(c.DelimiterRecords); i++ {
// 		if c.DelimiterRecords[i].ItemIndex == index {
// 			return i, true
// 		}
// 	}
//
// 	return -1, false
// }

// func (c *Cursor) runeBefore(delimSpan source.ByteSpan) (rune, bool) {
// 	if delimSpan.Start == c.Span.Start {
// 		return 0, false
// 	}
//
// 	leftWindow := source.ByteSpan{
// 		Start: c.Span.Start,
// 		End:   delimSpan.Start,
// 	}
//
// 	s := c.Source.Slice(leftWindow)
// 	r, width := utf8.DecodeLastRuneInString(s)
// 	if width == 0 {
// 		return 0, false
// 	}
//
// 	return r, true
// }

// func (c *Cursor) runeAfter(delimSpan source.ByteSpan) (rune, bool) {
// 	if delimSpan.End == c.Span.End {
// 		return 0, false
// 	}
//
// 	rightWindow := source.ByteSpan{
// 		Start: delimSpan.End,
// 		End:   c.Span.End,
// 	}
//
// 	s := c.Source.Slice(rightWindow)
// 	r, width := utf8.DecodeRuneInString(s)
// 	if width == 0 {
// 		return 0, false
// 	}
//
// 	return r, true
// }

// func leftFlanking(before rune, beforeOK bool, after rune, afterOK bool) bool {
// 	if !afterOK {
// 		return false
// 	}
// 	if isWhitespace(after) {
// 		return false
// 	}
// 	if !isPunctuation(after) {
// 		return true
// 	}
// 	if !beforeOK {
// 		return true
// 	}
// 	if isWhitespace(before) {
// 		return true
// 	}
// 	if isPunctuation(before) {
// 		return true
// 	}
// 	return false
// }

// func rightFlanking(before rune, beforeOK bool, after rune, afterOK bool) bool {
// 	if !beforeOK {
// 		return false
// 	}
// 	if isWhitespace(before) {
// 		return false
// 	}
// 	if !isPunctuation(before) {
// 		return true
// 	}
// 	if !afterOK {
// 		return true
// 	}
// 	if isWhitespace(after) {
// 		return true
// 	}
// 	if isPunctuation(after) {
// 		return true
// 	}
// 	return false
// }

func validateURI(s string) bool {
	idx := 0
	for idx < len(s) {
		if s[idx] == ':' {
			break
		}
		idx++
	}

	if idx == len(s) {
		return false
	}

	scheme := s[:idx]
	rest := s[idx+1:]

	// validate the scheme length
	if len(scheme) < 2 || len(scheme) > 32 {
		return false
	}

	// validate the first scheme character
	b := scheme[0]
	if !isAlpha(b) {
		return false
	}

	// validate the scheme characters
	for i := 1; i < len(scheme); i++ {
		b := scheme[i]
		if isAlpha(b) || isDigit(b) {
			continue
		}

		if b == '+' || b == '.' || b == '-' {
			continue
		}

		return false
	}

	// validate the rest of the URI
	for i := 0; i < len(rest); i++ {
		b := rest[i]
		if b < 0x20 || b == 0x7F {
			return false
		}

		if b == ' ' || b == '<' || b == '>' {
			return false
		}
	}

	return true
}

func validateEmail(s string) bool {
	idx := -1
	for i := 0; i < len(s); i++ {
		if s[i] != '@' {
			continue
		}
		if idx != -1 {
			return false
		}
		idx = i
	}

	if idx == -1 {
		return false
	}

	local := s[:idx]
	domain := s[idx+1:]

	if len(local) == 0 || len(domain) == 0 {
		return false
	}

	for i := 0; i < len(local); i++ {
		b := local[i]
		if isAlpha(b) || isDigit(b) || isEmailLocalSpecial(b) {
			continue
		}

		return false
	}

	for label := range strings.SplitSeq(domain, ".") {
		if len(label) < 1 || len(label) > 63 {
			return false
		}

		firstByte := label[0]
		if !isAlpha(firstByte) && !isDigit(firstByte) {
			return false
		}

		lastByte := label[len(label)-1]
		if !isAlpha(lastByte) && !isDigit(lastByte) {
			return false
		}

		for i := 0; i < len(label); i++ {
			b := label[i]
			if isAlpha(b) || isDigit(b) || b == '-' {
				continue
			}

			return false
		}
	}

	return true
}

func tryInlineHTML(s string) (int, bool) {
	if len(s) < 2 {
		return 0, false
	}

	switch s[1] {
	case '!':
		if n, ok := tryHTMLComment(s); ok {
			return n, true
		}

		if n, ok := tryHTMLCDATA(s); ok {
			return n, true
		}

		if n, ok := tryHTMLDeclaration(s); ok {
			return n, true
		}

	case '?':
		if n, ok := tryHTMLProcessingInstruction(s); ok {
			return n, true
		}

	case '/':
		if n, ok := tryHTMLClosingTag(s); ok {
			return n, true
		}

	default:
		if n, ok := tryHTMLOpenTag(s); ok {
			return n, true
		}
	}

	return 0, false
}

func tryHTMLComment(s string) (int, bool) {
	return 0, false
}

func tryHTMLProcessingInstruction(s string) (int, bool) {
	return 0, false
}

func tryHTMLDeclaration(s string) (int, bool) {
	return 0, false
}

func tryHTMLCDATA(s string) (int, bool) {
	return 0, false
}

func tryHTMLOpenTag(s string) (int, bool) {
	// NOTE: 1: candidate end detection
	insideSingleQuote := false
	insideDoubleQuote := false

	// traverse the string and break on the first unquoted closing angle bracket
	pos := 1
	for pos < len(s) {
		b := s[pos]
		// only toggle the double quote status if not inside single quotes
		if b == '"' && !insideSingleQuote {
			insideDoubleQuote = !insideDoubleQuote
		}

		// only toggle the single quote status if not inside double quotes
		if b == '\'' && !insideDoubleQuote {
			insideSingleQuote = !insideSingleQuote
		}

		// only break on closing angle brackets if not inside quoted material
		if b == '>' && !insideSingleQuote && !insideDoubleQuote {
			break
		}

		pos++
	}

	// no terminating '>' found outside of quotes, not a valid tag candidate
	if pos == len(s) {
		return 0, false
	}

	last := pos      // index of the closing '>'
	width := pos + 1 // total width of the candidate slice
	candidate := s[:width]

	// local helper
	consumeSpacesTabs := func(i int) int {
		for i < last && (candidate[i] == ' ' || candidate[i] == '\t') {
			i++
		}
		return i
	}

	// must begin with '<' and end with '>'
	if candidate[0] != '<' || candidate[last] != '>' {
		return 0, false
	}

	// NOTE: 2: tag name parsing

	// validate the tag name
	// the first byte after '<' must be an ASCII letter
	idx := 1
	if !isAlpha(candidate[idx]) {
		return 0, false
	}
	idx++

	// consume maximal tag name
	for idx < last {
		b := candidate[idx]
		if isAlpha(b) || isDigit(b) || b == '-' {
			idx++
			continue
		}
		break
	}

	// NOTE: 3: tail parsing

	// form is <tag>
	if idx == last {
		return width, true
	}

	// form is <tag/> or <tag /...>
	if candidate[idx] == '/' {
		idx++
		idx = consumeSpacesTabs(idx)
		if idx == last {
			return width, true
		}
		return 0, false
	}

	// general tail, attributes must be preceded by at least one space/tab
	for {
		mark := idx

		// consume separator whitespace between elements
		idx = consumeSpacesTabs(idx)

		// valid end
		if idx == last {
			return width, true
		}

		// self-closing suffix after whitespace
		if candidate[idx] == '/' {
			idx++
			idx = consumeSpacesTabs(idx)
			if idx == last {
				return width, true
			}
			return 0, false
		}

		// no separator consumed, cannot begin an attribute
		if mark == idx {
			return 0, false
		}

		// parse a single attribute (name + optional value)
		next, ok := tryHTMLAttribute(candidate, idx, last)

		// must succeed and must make forward progress
		if !ok || next <= idx {
			return 0, false
		}

		idx = next
	}
}

func tryHTMLClosingTag(s string) (int, bool) {
	return 0, false
}

// TODO:
func tryHTMLAttribute(string, int, int) (int, bool) {
	return 0, false
}

func isSpace(b byte) bool {
	return b == ' '
}

func isAllSpaces(s string) bool {
	for i := range len(s) {
		if !isSpace(s[i]) {
			return false
		}
	}

	return true
}

func isAlpha(b byte) bool {
	return 'A' <= b && b <= 'Z' || 'a' <= b && b <= 'z'
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isEmailLocalSpecial(b byte) bool {
	switch b {
	case '.', '!', '#', '$', '%', '&', '\'', '*', '+',
		'/', '=', '?', '^', '_', '`', '{', '|', '}', '~', '-':
		return true
	}
	return false
}

// func isWhitespace(r rune) bool {
// 	return unicode.IsSpace(r)
// }

// func isPunctuation(r rune) bool {
// 	return unicode.IsPunct(r)
// }

// func inlineSpan(inl ast.Inline) (source.ByteSpan, bool) {
// 	switch n := inl.(type) {
// 	case ast.Link:
// 		return n.Span, true
//
// 	case ast.Em:
// 		return n.Span, true
//
// 	case ast.Strong:
// 		return n.Span, true
//
// 	case ast.Text:
// 		return n.Span, true
//
// 	case ast.RawText:
// 		return n.Span, true
//
// 	case ast.HardBreak:
// 		return n.Span, true
//
// 	case ast.SoftBreak:
// 		return n.Span, true
//
// 	case ast.Newline:
// 		return n.Span, true
//
// 	default:
// 		return source.ByteSpan{}, false
// 	}
// }
