package inline

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type Cursor struct {
	Source     *source.Source
	Span       source.ByteSpan
	Tokens     []Token
	Index      int
	Items      *ItemList
	Delimiters *DelimiterList
}

func NewCursor(src *source.Source, span source.ByteSpan, tokens []Token) *Cursor {
	return &Cursor{
		Source:     src,
		Span:       span,
		Tokens:     tokens,
		Index:      0,
		Items:      NewItemList(),
		Delimiters: NewDelimiterList(),
	}
}

func (c *Cursor) Next() Token {
	out := c.Tokens[c.Index]
	c.Index++
	return out
}

func (c *Cursor) Peek() Token {
	return c.Tokens[c.Index]
}

func (c *Cursor) advanceToBytePos(pos source.BytePos) {
	for c.Index < len(c.Tokens) {
		tok := c.Tokens[c.Index]

		if tok.Span.Start >= pos {
			return
		}

		if tok.Span.Start < pos && tok.Span.End > pos {
			panic("advanceToBytePos: token straddles consumed boundary")
		}

		c.Index++
	}
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

		case TokenStarDelimiter:
			c.handleStarDelimiter()

		case TokenUnderscoreDelimiter:
			c.handleUnderscoreDelimiter()

		case TokenBacktick:
			c.handleTokenBacktick()

		case TokenOpenBracket:
			c.handleTokenOpenBracket()

		case TokenCloseBracket:
			// on encountering a ']' character, call the look for link or image procedure (below)

			// starting at the top of the delimiter stack, look backwards through the stack for an opening '[' or '![' delimiter
			openerDelim := c.Delimiters.Back()
			for openerDelim != nil {
				if openerDelim.Kind == DelimOpenBracket || openerDelim.Kind == DelimImageOpenBracket {
					break
				}

				openerDelim = openerDelim.prev
			}

			// if no opening bracket found, append a literal text node
			if openerDelim == nil {
				c.appendItemRecord(token.Span, ItemText)
				// NOTE: becomes return in helper version
				continue
			}

			// if an opening bracket is found, but it's not active, remove the inactive delimiter from the stack and append a literal text node
			if !openerDelim.Active {
				c.Delimiters.Remove(openerDelim)
				c.appendItemRecord(token.Span, ItemText)
				// NOTE: becomes return in helper version
				continue
			}

			// if an active opening bracket is found, then parse ahead to see whether we have:
			// - inline link/image
			// TODO:
			// - reference link/image (not yet implemented)
			// - collapsed reference link/image (not yet implemented)
			// - short cut reference link/image (not yet implemented)

			switch openerDelim.Kind {
			case DelimOpenBracket:
				// TODO: tryInlineLink helper

				// try to parse the inline link tail
				// if parse fails, remove the delimiter from the stack and append the the closing bracket as plain text
				tail, ok := c.tryParseInlineLinkTail(token.Span.End)
				if !ok {
					c.Delimiters.Remove(openerDelim)
					c.appendItemRecord(token.Span, ItemText)
					// NOTE: becomes return in helper version
					continue
				}

				openerItem := openerDelim.Item

				linkOriginalSpan := source.ByteSpan{
					Start: openerItem.OriginalSpan.Start,
					End:   tail.FullSpan.End,
				}

				linkLiveSpan := source.ByteSpan{
					Start: openerItem.OriginalSpan.End,
					End:   token.Span.Start,
				}

				// process emphasis beginning from the opening bracket
				c.processEmphasis(openerDelim)

				// identify the first and last child items
				firstChild := openerItem.Next()
				lastChild := c.Items.Back()

				// define the list of child items
				var childList *ItemList
				if firstChild == nil {
					// if there are no child items, initialize an empty list
					childList = NewItemList()
				} else {
					// otherwise, detach the contiguous span [firstChild, lastChild] and extract to a new *ItemList
					childList = c.Items.DetachRange(firstChild, lastChild)
				}

				// mutate the opener item and update metadata
				openerItem.Kind = ItemLink
				openerItem.OriginalSpan = linkOriginalSpan
				openerItem.LiveSpan = linkLiveSpan
				openerItem.DestinationSpan = tail.DestinationSpan
				openerItem.TitleSpan = source.ByteSpan{}
				openerItem.HasTitle = tail.HasTitle
				if tail.HasTitle {
					openerItem.TitleSpan = tail.TitleSpan
				}
				openerItem.Children = childList

				// deactivate all prior '[' delimiters
				for delim := openerDelim.Prev(); delim != nil; delim = delim.Prev() {
					if delim.Kind == DelimOpenBracket {
						delim.Active = false
					}
				}

				// advance the cursor past the tail
				c.advanceToBytePos(tail.FullSpan.End)

				// remove the opening delimiter from the stack
				c.Delimiters.Remove(openerDelim)

			case DelimImageOpenBracket:
				// TODO: tryInlineImage helper

			default:
				// TODO: revise panic text
				panic("unrecognized delimiter kind encountered")
			}

		case TokenOpenParen:
			c.appendItemRecord(token.Span, ItemText)

		case TokenCloseParen:
			c.appendItemRecord(token.Span, ItemText)

		case TokenOpenAngle:
			c.handleTokenOpenAngle()

		case TokenCloseAngle:
			c.appendItemRecord(token.Span, ItemText)

		case TokenBang:
			c.appendItemRecord(token.Span, ItemText)

		case TokenImageOpenBracket:
			c.handleTokenImageOpenBracket()

		case TokenBackslash:
			c.handleTokenBackslash()

		default:
			panic(fmt.Sprintf("unknown token kind encountered (%d)", token.Kind))
		}
	}

	// process emphasis once again across the entire stack
	c.processEmphasis(nil)

	inlines := c.lowerItems(c.Items)

	return inlines, nil
}

// TODO:
func (c *Cursor) buildItems() error {
	return nil
}

func (c *Cursor) lowerItems(items *ItemList) []ast.Inline {
	inlines := []ast.Inline{}

	item := items.Front()
	for item != nil {
		switch item.Kind {
		case ItemText:
			node := ast.Text{
				Span: item.LiveSpan,
			}

			inlines = append(inlines, node)

		case ItemCodeSpan:
			node := ast.CodeSpan{
				Span: item.LiveSpan,
			}

			inlines = append(inlines, node)

		case ItemAutolinkURI:
			contentSpan := source.ByteSpan{
				Start: item.OriginalSpan.Start + 1,
				End:   item.OriginalSpan.End - 1,
			}

			node := ast.Link{
				Span:        item.OriginalSpan,
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
				Start: item.OriginalSpan.Start + 1,
				End:   item.OriginalSpan.End - 1,
			}

			node := ast.Link{
				Span:        item.OriginalSpan,
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

		case ItemEmphasis:
			children := c.lowerItems(item.Children)

			node := ast.Emph{
				Span:     item.OriginalSpan,
				Children: children,
			}

			inlines = append(inlines, node)

		case ItemStrong:
			children := c.lowerItems(item.Children)

			node := ast.Strong{
				Span:     item.OriginalSpan,
				Children: children,
			}

			inlines = append(inlines, node)

		case ItemLink:
			children := c.lowerItems(item.Children)

			node := ast.Link{
				Span:        item.OriginalSpan,
				Destination: item.DestinationSpan,
				Children:    children,
			}

			if item.HasTitle {
				node.Title = item.TitleSpan
			}

			inlines = append(inlines, node)

		default:
			panic(fmt.Sprintf("unknown item kind encountered (%d)", item.Kind))
		}

		item = item.Next()
	}

	return inlines
}

func (c *Cursor) processEmphasis(stackBottom *DelimiterRecord) {
	// initialize the openers table
	openersTable := newOpenersTable(stackBottom)

	// define the starting delimiter record for traversal
	var current *DelimiterRecord
	if stackBottom == nil {
		// if nil is passed as argument, begin from the first delimiter in the list
		current = c.Delimiters.Front()
	} else {
		// otherwise, start from the first delimiter after the delimiter passed as argument
		current = stackBottom.Next()
	}

	// traverse the delimiters stack
	for current != nil {
		// if current cannot close, advance to the next delimiter
		if !current.CanClose {
			current = current.Next()
			continue
		}

		// obtain the openerKey for the current delimiter
		key := openerKeyForCloser(current)
		openerBottom := openersTable[key]

		// find a matching opener
		opener := findMatchingOpener(current, stackBottom, openerBottom)

		// if a matching opener is found, derive the emphasis level and resolve the match
		if opener != nil {
			strong := opener.Count >= 2 && current.Count >= 2
			current = c.resolveEmphasisMatch(opener, current, strong)
			continue
		}

		// if no matching opener is found, update the table for future searches
		openersTable[key] = current.Prev()

		// if current also cannot open, remove it from the delimiter stack (since it also cannot be a closer)
		next := current.Next()
		if !current.CanOpen {
			c.Delimiters.Remove(current)
		}
		current = next
	}

	c.removeAllDelimitersAbove(stackBottom)
}

func (c *Cursor) resolveEmphasisMatch(opener, closer *DelimiterRecord, strong bool) *DelimiterRecord {
	// determine the number of delimiter characters to consume
	use := 1
	if strong {
		use = 2
	}

	// compute the original and live spans for the item record to be created
	originalSpan := source.ByteSpan{
		Start: opener.Item.LiveSpan.End - source.BytePos(use),
		End:   closer.Item.LiveSpan.Start + source.BytePos(use),
	}

	liveSpan := source.ByteSpan{
		Start: opener.Item.LiveSpan.End,
		End:   closer.Item.LiveSpan.Start,
	}

	// save the next delimiter record to be potentially returned
	nextCurrent := closer.Next()

	// remove all delimiter records strictly between opener and closer
	c.removeAllDelimitersBetween(opener, closer)

	// guard against underflow
	if opener.Count < use || closer.Count < use {
		panic("resolveEmphasisMatch: delimiter count underflow")
	}

	// decrement opener and closer counts by use
	opener.Count -= use
	closer.Count -= use

	// update the ItemRecord live spans as well
	opener.Item.LiveSpan.End -= source.BytePos(use)
	closer.Item.LiveSpan.Start += source.BytePos(use)

	// define the list of child items
	var childList *ItemList
	if opener.Item.Next() == closer.Item {
		// if there is no item strictly between the opener and closer items, initialize an empty list
		childList = NewItemList()
	} else {
		// otherwise, detach the contiguous span [firstChild, lastChild] and extract to a new *ItemList
		firstChild := opener.Item.Next()
		lastChild := closer.Item.Prev()
		childList = c.Items.DetachRange(firstChild, lastChild)
	}

	item := &ItemRecord{
		OriginalSpan: originalSpan,
		LiveSpan:     liveSpan,
		Children:     childList,
	}

	if use == 1 {
		item.Kind = ItemEmphasis
	} else {
		item.Kind = ItemStrong
	}

	c.Items.InsertAfter(item, opener.Item)

	// remove the items and delimiters whose runs are fully consumed
	if opener.Count == 0 {
		c.Items.Remove(opener.Item)
		c.Delimiters.Remove(opener)
	}
	if closer.Count == 0 {
		c.Items.Remove(closer.Item)
		c.Delimiters.Remove(closer)
		// if closer count reached zero, return the nextCurrent record
		return nextCurrent
	}

	// if closer count is > 0, return the closer
	return closer
}

func (c *Cursor) removeAllDelimitersAbove(stackBottom *DelimiterRecord) {
	var current *DelimiterRecord
	if stackBottom == nil {
		current = c.Delimiters.Front()
	} else {
		current = stackBottom.Next()
	}

	for current != nil {
		next := current.Next()
		c.Delimiters.Remove(current)
		current = next
	}
}

func (c *Cursor) removeAllDelimitersBetween(opener, closer *DelimiterRecord) {
	first := opener.Next()
	if first == nil || first == closer {
		return
	}

	last := closer.Prev()
	c.Delimiters.RemoveRange(first, last)
}

type openerKey struct {
	kind    DelimiterKind
	mod3    int
	canOpen bool
}

func newOpenersTable(bottom *DelimiterRecord) map[openerKey]*DelimiterRecord {
	m := make(map[openerKey]*DelimiterRecord)

	kinds := []DelimiterKind{
		DelimAsterisk,
		DelimUnderscore,
	}

	for _, kind := range kinds {
		for mod3 := range 3 {
			for _, canOpen := range []bool{false, true} {
				key := openerKey{
					kind:    kind,
					mod3:    mod3,
					canOpen: canOpen,
				}

				m[key] = bottom
			}
		}
	}

	return m
}

func openerKeyForCloser(delim *DelimiterRecord) openerKey {
	return openerKey{
		kind:    delim.Kind,
		mod3:    delim.Count % 3,
		canOpen: delim.CanOpen,
	}
}

func findMatchingOpener(closer, stackBottom, openerBottom *DelimiterRecord) *DelimiterRecord {
	for opener := closer.Prev(); opener != nil && opener != stackBottom && opener != openerBottom; opener = opener.Prev() {
		if delimitersMatch(opener, closer) {
			return opener
		}
	}

	return nil
}

func delimitersMatch(opener, closer *DelimiterRecord) bool {
	if opener == nil || closer == nil {
		return false
	}

	if opener.Kind != closer.Kind {
		return false
	}

	if !opener.CanOpen || !closer.CanClose {
		return false
	}

	if (opener.CanClose || closer.CanOpen) &&
		(opener.Count+closer.Count)%3 == 0 {
		return false
	}

	return true

}

func (c *Cursor) handleStarDelimiter() {
	tokenIdx := c.Index - 1
	token := c.Tokens[tokenIdx]

	// provisionally append the delimiter run as plain text
	item := c.appendItemRecord(token.Span, ItemText)

	// extract information about the preceding and following runes
	before, beforeOK := c.runeBefore(token.Span)
	after, afterOK := c.runeAfter(token.Span)

	// determine the flanking for determining opening/closing capability
	left := leftFlanking(before, beforeOK, after, afterOK)
	right := rightFlanking(before, beforeOK, after, afterOK)

	// create and append the corresponding delimiter record
	delim := &DelimiterRecord{
		Item:     item,
		Kind:     DelimAsterisk,
		Count:    token.Span.Width(),
		Active:   true,
		CanOpen:  left,
		CanClose: right,
	}

	c.Delimiters.PushBack(delim)
}

func (c *Cursor) handleUnderscoreDelimiter() {
	tokenIdx := c.Index - 1
	token := c.Tokens[tokenIdx]

	// provisionally append the delimiter run as plain text
	item := c.appendItemRecord(token.Span, ItemText)

	// extract information about the preceding and following runes
	before, beforeOK := c.runeBefore(token.Span)
	after, afterOK := c.runeAfter(token.Span)

	beforeIsPunct := beforeOK && isPunctuation(before)
	afterIsPunct := afterOK && isPunctuation(after)

	// determine flanking
	left := leftFlanking(before, beforeOK, after, afterOK)
	right := rightFlanking(before, beforeOK, after, afterOK)

	// define opening/closing capability
	canOpen := left && (!right || beforeIsPunct)
	canClose := right && (!left || afterIsPunct)

	// create and append the corresponding delimiter record
	delim := &DelimiterRecord{
		Item:     item,
		Kind:     DelimUnderscore,
		Count:    token.Span.Width(),
		Active:   true,
		CanOpen:  canOpen,
		CanClose: canClose,
	}

	c.Delimiters.PushBack(delim)
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

	// define the original span (delimiter inclusive)
	originalSpan := source.ByteSpan{
		Start: c.Tokens[openerIdx].Span.Start,
		End:   c.Tokens[closerIdx].Span.End,
	}
	// define the live span (delimiter exclusive)
	liveSpan := source.ByteSpan{
		Start: c.Tokens[openerIdx].Span.End,
		End:   c.Tokens[closerIdx].Span.Start,
	}

	// examine the content slice, and if there is both leading and trailing spaces and at least one non-space character, trim one space from both ends
	contentSlice := c.Source.Slice(liveSpan)

	if len(contentSlice) > 0 &&
		isSpace(contentSlice[0]) &&
		isSpace(contentSlice[len(contentSlice)-1]) &&
		!isAllSpaces(contentSlice) {
		liveSpan.Start++
		liveSpan.End--
	}

	// define the new ItemRecord
	item := &ItemRecord{
		OriginalSpan: originalSpan,
		LiveSpan:     liveSpan,
		Kind:         ItemCodeSpan,
	}

	// append the item to the item record list and advance the index past the closer backtick token
	c.Items.PushBack(item)
	c.Index = closerIdx + 1
}

func (c *Cursor) handleTokenOpenBracket() {
	tokenIdx := c.Index - 1
	token := c.Tokens[tokenIdx]

	// provisionally append the open bracket as plain text
	item := c.appendItemRecord(token.Span, ItemText)

	// create and append the corresponding delimiter record
	delim := &DelimiterRecord{
		Item:   item,
		Kind:   DelimOpenBracket,
		Active: true,
	}

	c.Delimiters.PushBack(delim)
}

func (c *Cursor) handleTokenOpenAngle() {
	openerIdx := c.Index - 1
	openerToken := c.Tokens[openerIdx]

	// search for the very next close angle token
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
		return
	}
	closerToken := c.Tokens[closerIdx]

	// define the outer span (including angle brackets)
	outerSpan := source.ByteSpan{
		Start: openerToken.Span.Start,
		End:   closerToken.Span.End,
	}

	// define the content span (excluding angle brackets)
	contentSpan := source.ByteSpan{
		Start: openerToken.Span.End,
		End:   closerToken.Span.Start,
	}

	// extract the content slice
	contentSlice := c.Source.Slice(contentSpan)

	// check if the content span is a valid URI autolink
	if validateURIAutolink(contentSlice) {
		c.appendItemRecord(outerSpan, ItemAutolinkURI)
		c.Index = closerIdx + 1
		return
	}

	// check if the content span is a valid email autolink
	if validateEmailAutolink(contentSlice) {
		c.appendItemRecord(outerSpan, ItemAutolinkEmail)
		c.Index = closerIdx + 1
		return
	}

	// if autolinks fail, extract the span and matching string from the opening token
	// through the very end of the line for use in a byte-wise search
	candidateSpan := source.ByteSpan{
		Start: openerToken.Span.Start,
		End:   c.Span.End,
	}
	candidate := c.Source.Slice(candidateSpan)

	// try to find a valid inline HTML construct, otherwise return the opening token as plain text
	width, ok := tryInlineHTML(candidate)
	if !ok {
		c.appendItemRecord(openerToken.Span, ItemText)
		return
	}

	// if valid inline HTML is detected, update the candidate span to match the width returned
	candidateSpan.End = openerToken.Span.Start + source.BytePos(width)

	// recreate the closing angle bracket token's span based on the candidate span
	targetSpan := source.ByteSpan{
		Start: candidateSpan.End - 1,
		End:   candidateSpan.End,
	}

	// locate the matching token for this closing angle bracket by inspecting spans
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
}

func (c *Cursor) handleTokenImageOpenBracket() {
	tokenIdx := c.Index - 1
	token := c.Tokens[tokenIdx]

	// provisionally append the open image bracket as plain text
	item := c.appendItemRecord(token.Span, ItemText)

	// create and append the corresponding delimiter record
	delim := &DelimiterRecord{
		Item:   item,
		Kind:   DelimImageOpenBracket,
		Active: true,
	}

	c.Delimiters.PushBack(delim)
}

func (c *Cursor) handleTokenBackslash() {
	tokenIdx := c.Index - 1
	token := c.Tokens[tokenIdx]

	// peek the next token
	next := c.Peek()

	switch classifyEscapeTarget(next) {
	case EscapeDecompose:
		// consume the next token
		c.Next()

		// decompose the image open bracket
		bangSpan := source.ByteSpan{
			Start: next.Span.Start,
			End:   next.Span.Start + 1,
		}
		bracketSpan := source.ByteSpan{
			Start: next.Span.Start + 1,
			End:   next.Span.End,
		}

		// append the bang as plain text
		c.appendItemRecord(bangSpan, ItemText)

		// provisionally append the open bracket as plain text
		item := c.appendItemRecord(bracketSpan, ItemText)

		// create and append teh corresponding delimiter record
		delim := &DelimiterRecord{
			Item:   item,
			Kind:   DelimOpenBracket,
			Active: true,
		}

		c.Delimiters.PushBack(delim)

	case EscapeLiteralize:
		// consume the next token and append as plain text
		c.Next()
		c.appendItemRecord(next.Span, ItemText)

	case EscapeNone:
		// append the backslash as plain text
		c.appendItemRecord(token.Span, ItemText)
	}
}

func (c *Cursor) appendItemRecord(span source.ByteSpan, kind ItemKind) *ItemRecord {
	item := &ItemRecord{
		OriginalSpan: span,
		LiveSpan:     span,
		Kind:         kind,
	}

	return c.Items.PushBack(item)
}

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

type InlineLinkTail struct {
	FullSpan        source.ByteSpan // from '(' through ')'
	DestinationSpan source.ByteSpan
	TitleSpan       source.ByteSpan
	HasTitle        bool
}

func (c *Cursor) tryParseInlineLinkTail(start source.BytePos) (InlineLinkTail, bool) {
	candidateSpan := source.ByteSpan{
		Start: start,
		End:   c.Source.EOF(),
	}
	s := c.Source.Slice(candidateSpan)
	limit := len(s)

	result := InlineLinkTail{}

	// validate the candidate length and that the first byte is an open paren
	if limit < 2 || s[0] != '(' {
		return InlineLinkTail{}, false
	}

	// advance past the open paren
	idx := 1

	// consume any spaces and tabs
	idx = consumeSpacesTabs(s, idx, limit)

	if idx < limit && s[idx] == ')' {
		// valid link title, no destination & no title
		result.FullSpan = source.ByteSpan{
			Start: candidateSpan.Start,
			End:   candidateSpan.Start + source.BytePos(idx+1),
		}

		return result, true
	}

	// validate and consume the link destination
	destinationSpanRel, idx, ok := tryLinkDestination(s, idx, limit)
	if !ok {
		return InlineLinkTail{}, false
	}

	// update the result struct
	result.DestinationSpan = source.ByteSpan{
		Start: candidateSpan.Start + destinationSpanRel.Start,
		End:   candidateSpan.Start + destinationSpanRel.End,
	}

	// mark the index and consume any spaces or tabs
	sepStart := idx
	idx = consumeSpacesTabs(s, idx, limit)
	sepPresent := idx > sepStart

	if idx < limit && s[idx] == ')' {
		// valid link title, no title
		result.FullSpan = source.ByteSpan{
			Start: candidateSpan.Start,
			End:   candidateSpan.Start + source.BytePos(idx+1),
		}

		return result, true
	}

	if idx >= limit {
		return InlineLinkTail{}, false
	}

	// check whether the next byte can start a link title
	if s[idx] == '"' || s[idx] == '\'' || s[idx] == '(' {
		// if no separator exists between destination and title, the tail is invalid
		if !sepPresent {
			return InlineLinkTail{}, false
		}

		titleSpanRel, idx, ok := tryLinkTitle(s, idx, limit)
		if !ok {
			return InlineLinkTail{}, false
		}

		// update the result struct
		result.TitleSpan = source.ByteSpan{
			Start: candidateSpan.Start + titleSpanRel.Start,
			End:   candidateSpan.Start + titleSpanRel.End,
		}

		// consume any spaces or tabs
		idx = consumeSpacesTabs(s, idx, limit)

		if idx < limit && s[idx] == ')' {
			// valid link title
			result.FullSpan = source.ByteSpan{
				Start: candidateSpan.Start,
				End:   candidateSpan.Start + source.BytePos(idx+1),
			}

			return result, true
		}
	}

	return InlineLinkTail{}, false
}

func tryLinkDestination(s string, idx, limit int) (source.ByteSpan, int, bool) {
	if idx >= limit {
		return source.ByteSpan{}, 0, false
	}

	if s[idx] == '<' {
		return tryAngleLinkDestination(s, idx, limit)
	}
	return tryBareLinkDestination(s, idx, limit)
}

func tryAngleLinkDestination(s string, idx, limit int) (source.ByteSpan, int, bool) {
	// validate that the first byte is '<'
	if idx >= limit || s[idx] != '<' {
		return source.ByteSpan{}, 0, false
	}

	// advance past the opening angle bracket
	idx++
	start := idx

	for idx < limit {
		switch s[idx] {
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
				End:   source.BytePos(idx),
			}

			return span, idx + 1, true

		case '\\':
			// on a backslash, advance two bytes if within span limit
			if idx+1 < limit {
				idx += 2
				continue
			}
			// otherwise, trailing backslash is just ordinary content
			idx++

		default:
			idx++
		}
	}

	return source.ByteSpan{}, 0, false
}

// TODO:
func tryBareLinkDestination(s string, idx, limit int) (source.ByteSpan, int, bool) {
	return source.ByteSpan{}, 0, false
}

func tryLinkTitle(s string, idx, limit int) (source.ByteSpan, int, bool) {
	if idx >= limit {
		return source.ByteSpan{}, 0, false
	}

	switch s[idx] {
	case '"', '\'':
		return tryQuotedLinkTitle(s, idx, limit, s[idx])
	case '(':
		return tryParenLinkTitle(s, idx, limit)
	default:
		return source.ByteSpan{}, 0, false
	}
}

func tryQuotedLinkTitle(s string, idx, limit int, delim byte) (source.ByteSpan, int, bool) {
	if idx >= limit || s[idx] != delim {
		return source.ByteSpan{}, 0, false
	}

	idx++
	start := idx

	for idx < limit {
		switch s[idx] {
		case '\n', '\r':
			// newlines are not permitted inside the title
			return source.ByteSpan{}, 0, false

		case delim:
			// an unescaped closer ends the title
			span := source.ByteSpan{
				Start: source.BytePos(start),
				End:   source.BytePos(idx),
			}

			return span, idx + 1, true

		case '\\':
			// on a backslash, advance two bytes if within span limit
			if idx+1 < limit {
				idx += 2
				continue
			}
			// otherwise, trailing backslash is just ordinary content
			idx++

		default:
			idx++
		}
	}

	return source.ByteSpan{}, 0, false
}

func tryParenLinkTitle(s string, idx, limit int) (source.ByteSpan, int, bool) {
	if idx >= limit || s[idx] != '(' {
		return source.ByteSpan{}, 0, false
	}

	idx++
	start := idx
	depth := 1

	for idx < limit {
		switch s[idx] {
		case '\n', '\r':
			// newlines are not permitted inside the title
			return source.ByteSpan{}, 0, false

		case '(':
			// an unescaped open paren increases the paren depth
			depth++
			idx++

		case ')':
			// an unescaped close paren decreases the paren depth
			depth--

			// reaching the depth 0 ends the title
			if depth == 0 {
				span := source.ByteSpan{
					Start: source.BytePos(start),
					End:   source.BytePos(idx),
				}

				return span, idx + 1, true
			}

			idx++

		case '\\':
			// on a backslash, advance two bytes if within span limit
			if idx+1 < limit {
				idx += 2
				continue
			}
			// otherwise, trailing backslash is just ordinary content
			idx++

		default:
			idx++
		}
	}

	return source.ByteSpan{}, 0, false
}

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

func validateURIAutolink(s string) bool {
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

func validateEmailAutolink(s string) bool {
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
	return tryHTMLDelimited(s, "<!--", "-->")
}

func tryHTMLProcessingInstruction(s string) (int, bool) {
	return tryHTMLDelimited(s, "<?", "?>")
}

func tryHTMLDeclaration(s string) (int, bool) {
	return tryHTMLDelimited(s, "<!", ">")
}

func tryHTMLCDATA(s string) (int, bool) {
	return tryHTMLDelimited(s, "<![CDATA[", "]]>")
}

func tryHTMLOpenTag(s string) (int, bool) {
	if len(s) < 2 {
		return 0, false
	}

	// traverse the string and break on the first unquoted closing angle bracket
	pos := 1
	insideSingleQuote := false
	insideDoubleQuote := false

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

	// must begin with '<' and end with '>'
	if candidate[0] != '<' || candidate[last] != '>' {
		return 0, false
	}

	// validate and consume the tag name
	idx := 1
	idx, ok := tryHTMLTagName(s, idx, last)
	if !ok {
		return 0, false
	}

	// form is <tag>
	if idx == last {
		return width, true
	}

	// form is <tag/> or <tag /...>
	if candidate[idx] == '/' {
		if _, ok := tryHTMLSelfClosingSuffix(candidate, idx, last); ok {
			return width, true
		}
		return 0, false
	}

	// general tail, attributes must be preceded by at least one space/tab
	for {
		mark := idx

		// consume separator whitespace between elements
		idx = consumeSpacesTabs(candidate, idx, last)

		// valid end
		if idx == last {
			return width, true
		}

		// self-closing suffix after whitespace
		if candidate[idx] == '/' {
			if _, ok := tryHTMLSelfClosingSuffix(candidate, idx, last); ok {
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
	if len(s) < 3 {
		return 0, false
	}

	// validate the open angle bracket
	if s[0] != '<' || s[1] != '/' {
		return 0, false
	}

	// traverse the string and break on the first closing angle bracket
	pos := 2
	for pos < len(s) {
		if s[pos] == '>' {
			break
		}
		pos++
	}

	// no terminating '>' found, not a valid candidate
	if pos == len(s) {
		return 0, false
	}

	last := pos
	width := pos + 1
	candidate := s[:width]

	// validate and consume the tag name
	idx := 2
	idx, ok := tryHTMLTagName(candidate, idx, last)
	if !ok {
		return 0, false
	}

	idx = consumeSpacesTabs(candidate, idx, last)

	if idx == last {
		return width, true
	}

	return 0, false
}

func tryHTMLTagName(s string, idx, last int) (int, bool) {
	// the first byte must be an ASCII letter
	if !isAlpha(s[idx]) {
		return 0, false
	}
	idx++

	// consume maximal tag name
	for idx < last {
		b := s[idx]
		if isAlpha(b) || isDigit(b) || b == '-' {
			idx++
			continue
		}
		break
	}

	return idx, true
}

func tryHTMLSelfClosingSuffix(s string, idx, last int) (int, bool) {
	// the suffix must being with a forward slash
	if idx >= last || s[idx] != '/' {
		return 0, false
	}
	idx++

	// consume any optional spaces or tabs after the slice
	idx = consumeSpacesTabs(s, idx, last)

	// the suffix is valid only if it runs directly up to the closing angle bracket
	if idx != last {
		return 0, false
	}

	return idx, true

}

func tryHTMLAttribute(s string, idx, last int) (int, bool) {
	// TODO: extract to tryHTMLAttributeName

	// validate the attribute name
	// the first character must be an ASCII letter, '_', or ':'
	if !isAlpha(s[idx]) && s[idx] != '_' && s[idx] != ':' {
		return 0, false
	}
	idx++

	// consume maximal attribute name:
	// ASCII letters, digits, '_', '.', ':', or '-'
	for idx < last {
		b := s[idx]
		if isAlpha(b) ||
			isDigit(b) ||
			b == '_' ||
			b == '.' ||
			b == ':' ||
			b == '-' {
			idx++
			continue
		}
		break
	}

	// if attribute name advances up to the closing angle bracket,
	// the attribute is a bare name with no value specification
	if idx == last {
		return idx, true
	}

	// after the name, only spaces, tabs, or '=' may appear
	if s[idx] != ' ' && s[idx] != '\t' && s[idx] != '=' {
		return 0, false
	}

	// TODO: leave this as-is within function, probing for '='

	// scan ahead for an attribute value specification
	//
	// consume any spaces or tabs after the name:
	// - if no '=' follows, the attribute is a bare name and trailing
	//   whitespace is left for the outer parser
	// - if '=' follows, continue into value parsing
	probe := consumeSpacesTabs(s, idx, last)
	if probe == last || s[probe] != '=' {
		return idx, true
	}

	// consume spaces or tabs after '=' and position idx
	// at the first byte of the attribute value
	idx = consumeSpacesTabs(s, probe+1, last)

	// an '=' must by followed by an attribute value
	if idx == last {
		return 0, false
	}

	// TODO: extract to tryHTMLAttributeValue

	// parse one of the three attribute value forms:
	// single-quoted, double-quoted, or unquoted
	switch s[idx] {
	case '\'':
		// single-quoted value
		idx++
		for idx < last {
			if s[idx] != '\'' {
				idx++
				continue
			}
			break
		}

		// no closing single quote found
		if idx == last {
			return 0, false
		}

		// consume the closing quote
		idx++

	case '"':
		// double-quoted value
		idx++
		for idx < last {
			if s[idx] != '"' {
				idx++
				continue
			}
			break
		}

		// no closing double quote found
		if idx == last {
			return 0, false
		}

		// consume the closing quote
		idx++

	default:
		// unquoted value
		// a nonempty string of characters excluding spaces, tabs, ", ', =, <, >, and `
		start := idx
		for idx < last {
			b := s[idx]
			if b == ' ' ||
				b == '\t' ||
				b == '"' ||
				b == '\'' ||
				b == '=' ||
				b == '<' ||
				b == '>' ||
				b == '`' {
				break
			}
			idx++
		}

		// unquoted values must be nonempty
		if idx == start {
			return 0, false
		}
	}

	return idx, true
}

// TODO:
func tryHTMLAttributeName() {}

// TODO:
func tryHTMLAttributeValue() {}

func tryHTMLDelimited(s, opener, terminator string) (int, bool) {
	if !strings.HasPrefix(s, opener) {
		return 0, false
	}

	i := strings.Index(s, terminator)
	if i == -1 {
		return 0, false
	}

	return i + len(terminator), true
}

func consumeSpacesTabs(s string, idx, last int) int {
	for idx < last {
		b := s[idx]
		if b == ' ' || b == '\t' {
			idx++
			continue
		}
		break
	}

	return idx
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

func isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

func isPunctuation(r rune) bool {
	return unicode.IsPunct(r)
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
