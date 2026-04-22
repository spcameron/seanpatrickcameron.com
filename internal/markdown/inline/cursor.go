package inline

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/reference"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

// Cursor carries the mutable state used while resolving inline tokens into
// item records and final AST nodes.
type Cursor struct {
	Source      *source.Source
	Definitions map[string]ir.ReferenceDefinition
	Span        source.ByteSpan
	Tokens      []Token
	Index       int
	Items       *ItemList
	Delimiters  *DelimiterList
}

// NewCursor constructs an inline parsing cursor over the given token stream.
func NewCursor(src *source.Source, defs map[string]ir.ReferenceDefinition, span source.ByteSpan, tokens []Token) *Cursor {
	return &Cursor{
		Source:      src,
		Definitions: defs,
		Span:        span,
		Tokens:      tokens,
		Index:       0,
		Items:       NewItemList(),
		Delimiters:  NewDelimiterList(),
	}
}

// Next returns the current token and advances the cursor.
func (c *Cursor) Next() Token {
	out := c.Tokens[c.Index]
	c.Index++
	return out
}

// Peek returns the current token without advancing.
func (c *Cursor) Peek() Token {
	return c.Tokens[c.Index]
}

// advanceToBytePos advances the token cursor to the first token whose span
// begins at or after pos.
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

// Build resolves the token stream into AST inline nodes.
func (c *Cursor) Build() ([]ast.Inline, error) {
	err := c.buildItems()
	if err != nil {
		return []ast.Inline{}, err
	}

	inlines := c.lowerItems(c.Items)

	return inlines, nil
}

// buildItems consumes the token stream into the mutable item and delimiter
// structures used by inline resolution.
func (c *Cursor) buildItems() error {
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
			c.handleTokenCloseBracket()

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

	// process emphasis once again across the entire delimiters stack
	c.processEmphasis(nil)

	return nil
}

// lowerItems converts resolved item records into AST inline nodes.
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

		case ItemImage:
			children := c.lowerItems(item.Children)

			node := ast.Image{
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

// processEmphasis resolves emphasis and strong emphasis by matching
// delimiter runs using the delimiter stack.
func (c *Cursor) processEmphasis(stackBottom *DelimiterRecord) {
	openersTable := newOpenersTable(stackBottom)

	var current *DelimiterRecord
	if stackBottom == nil {
		current = c.Delimiters.Front()
	} else {
		current = stackBottom.Next()
	}

	for current != nil {
		if !current.CanClose {
			current = current.Next()
			continue
		}

		key := openerKeyForCloser(current)
		openerBottom := openersTable[key]

		opener := findMatchingOpener(current, stackBottom, openerBottom)

		if opener != nil {
			strong := opener.Count >= 2 && current.Count >= 2
			current = c.resolveEmphasisMatch(opener, current, strong)
			continue
		}

		openersTable[key] = current.Prev()

		next := current.Next()
		if !current.CanOpen {
			c.Delimiters.Remove(current)
		}
		current = next
	}

	c.removeAllDelimitersAbove(stackBottom)
}

// resolveEmphasisMatch consumes a matched opener/closer pair and
// replaces their contents with an emphasis or strong item.
func (c *Cursor) resolveEmphasisMatch(opener, closer *DelimiterRecord, strong bool) *DelimiterRecord {
	use := 1
	if strong {
		use = 2
	}

	originalSpan := source.ByteSpan{
		Start: opener.Item.LiveSpan.End - source.BytePos(use),
		End:   closer.Item.LiveSpan.Start + source.BytePos(use),
	}

	liveSpan := source.ByteSpan{
		Start: opener.Item.LiveSpan.End,
		End:   closer.Item.LiveSpan.Start,
	}

	nextCurrent := closer.Next()

	c.removeAllDelimitersBetween(opener, closer)

	// guard against underflow
	if opener.Count < use || closer.Count < use {
		panic("resolveEmphasisMatch: delimiter count underflow")
	}

	opener.Count -= use
	closer.Count -= use

	opener.Item.LiveSpan.End -= source.BytePos(use)
	closer.Item.LiveSpan.Start += source.BytePos(use)

	var childList *ItemList
	if opener.Item.Next() == closer.Item {
		childList = NewItemList()
	} else {
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

	if opener.Count == 0 {
		c.Items.Remove(opener.Item)
		c.Delimiters.Remove(opener)
	}
	if closer.Count == 0 {
		c.Items.Remove(closer.Item)
		c.Delimiters.Remove(closer)

		return nextCurrent
	}

	return closer
}

// removeAllDelimitersAbove removes all delimiters above stackBottom.
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

// openerKey identifies a class of potential openers used to bound
// future opener searches.
type openerKey struct {
	kind    DelimiterKind
	mod3    int
	canOpen bool
}

// newOpenersTable initializes the opener search table for emphasis
// resolution.
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

// findMatchingOpener searches backward for a delimiter that can match
// the given closer within the allowed bounds.
func findMatchingOpener(closer, stackBottom, openerBottom *DelimiterRecord) *DelimiterRecord {
	for opener := closer.Prev(); opener != nil &&
		opener != stackBottom &&
		opener != openerBottom; opener = opener.Prev() {
		if delimitersMatch(opener, closer) {
			return opener
		}
	}

	return nil
}

// delimitersMatch reports whether opener and closer can form a valid
// emphasis or strong emphasis pair.
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

	item := c.appendItemRecord(token.Span, ItemText)

	before, beforeOK := c.runeBefore(token.Span)
	after, afterOK := c.runeAfter(token.Span)

	left := leftFlanking(before, beforeOK, after, afterOK)
	right := rightFlanking(before, beforeOK, after, afterOK)

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

	item := c.appendItemRecord(token.Span, ItemText)

	before, beforeOK := c.runeBefore(token.Span)
	after, afterOK := c.runeAfter(token.Span)

	beforeIsPunct := beforeOK && isPunctuation(before)
	afterIsPunct := afterOK && isPunctuation(after)

	left := leftFlanking(before, beforeOK, after, afterOK)
	right := rightFlanking(before, beforeOK, after, afterOK)

	canOpen := left && (!right || beforeIsPunct)
	canClose := right && (!left || afterIsPunct)

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

	if closerIdx >= len(c.Tokens) {
		c.appendItemRecord(openerToken.Span, ItemText)
		return
	}

	originalSpan := source.ByteSpan{
		Start: c.Tokens[openerIdx].Span.Start,
		End:   c.Tokens[closerIdx].Span.End,
	}

	liveSpan := source.ByteSpan{
		Start: c.Tokens[openerIdx].Span.End,
		End:   c.Tokens[closerIdx].Span.Start,
	}

	contentSlice := c.Source.Slice(liveSpan)

	if len(contentSlice) > 0 &&
		isSpace(contentSlice[0]) &&
		isSpace(contentSlice[len(contentSlice)-1]) &&
		!isAllSpaces(contentSlice) {
		liveSpan.Start++
		liveSpan.End--
	}

	item := &ItemRecord{
		OriginalSpan: originalSpan,
		LiveSpan:     liveSpan,
		Kind:         ItemCodeSpan,
	}

	c.Items.PushBack(item)
	c.Index = closerIdx + 1
}

func (c *Cursor) handleTokenOpenBracket() {
	tokenIdx := c.Index - 1
	token := c.Tokens[tokenIdx]

	item := c.appendItemRecord(token.Span, ItemText)

	delim := &DelimiterRecord{
		Item:   item,
		Kind:   DelimOpenBracket,
		Active: true,
	}

	c.Delimiters.PushBack(delim)
}

func (c *Cursor) handleTokenCloseBracket() {
	tokenIdx := c.Index - 1
	token := c.Tokens[tokenIdx]

	// search backward for the nearest unmatched '[' or '!['
	openerDelim := c.Delimiters.Back()
	for openerDelim != nil {
		if openerDelim.Kind == DelimOpenBracket ||
			openerDelim.Kind == DelimImageOpenBracket {
			break
		}

		openerDelim = openerDelim.Prev()
	}

	if openerDelim == nil {
		c.appendItemRecord(token.Span, ItemText)
		return
	}

	// inactive openers cannot form links; treat ']' literally
	if !openerDelim.Active {
		c.Delimiters.Remove(openerDelim)
		c.appendItemRecord(token.Span, ItemText)
		return
	}

	switch openerDelim.Kind {
	case DelimOpenBracket:
		if c.tryResolveBracket(openerDelim, token, ItemLink) {
			return
		}

		c.Delimiters.Remove(openerDelim)
		c.appendItemRecord(token.Span, ItemText)
		return

	case DelimImageOpenBracket:
		if c.tryResolveBracket(openerDelim, token, ItemImage) {
			return
		}

		c.Delimiters.Remove(openerDelim)
		c.appendItemRecord(token.Span, ItemText)
		return

	default:
		panic("unrecognized delimiter kind encountered")
	}
}

func (c *Cursor) handleTokenOpenAngle() {
	openerIdx := c.Index - 1
	openerToken := c.Tokens[openerIdx]

	closerIdx := openerIdx + 1
	for closerIdx < len(c.Tokens) {
		next := c.Tokens[closerIdx]
		if next.Kind != TokenCloseAngle {
			closerIdx++
			continue
		}

		break
	}

	if closerIdx == len(c.Tokens) {
		c.appendItemRecord(openerToken.Span, ItemText)
		return
	}
	closerToken := c.Tokens[closerIdx]

	outerSpan := source.ByteSpan{
		Start: openerToken.Span.Start,
		End:   closerToken.Span.End,
	}

	contentSpan := source.ByteSpan{
		Start: openerToken.Span.End,
		End:   closerToken.Span.Start,
	}

	contentSlice := c.Source.Slice(contentSpan)

	if validateURIAutolink(contentSlice) {
		c.appendItemRecord(outerSpan, ItemAutolinkURI)
		c.Index = closerIdx + 1
		return
	}

	if validateEmailAutolink(contentSlice) {
		c.appendItemRecord(outerSpan, ItemAutolinkEmail)
		c.Index = closerIdx + 1
		return
	}

	candidateSpan := source.ByteSpan{
		Start: openerToken.Span.Start,
		End:   c.Span.End,
	}
	candidate := c.Source.Slice(candidateSpan)

	width, ok := tryInlineHTML(candidate)
	if !ok {
		c.appendItemRecord(openerToken.Span, ItemText)
		return
	}

	candidateSpan.End = openerToken.Span.Start + source.BytePos(width)

	targetSpan := source.ByteSpan{
		Start: candidateSpan.End - 1,
		End:   candidateSpan.End,
	}

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

	item := c.appendItemRecord(token.Span, ItemText)

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

	next := c.Peek()

	switch classifyEscapeTarget(next, c.Source) {
	case EscapeDecompose:
		c.Next()

		// escape "![": literalize '!' and keep '[' bracket-active
		bangSpan := source.ByteSpan{
			Start: next.Span.Start,
			End:   next.Span.Start + 1,
		}

		bracketSpan := source.ByteSpan{
			Start: next.Span.Start + 1,
			End:   next.Span.End,
		}

		c.appendItemRecord(bangSpan, ItemText)

		item := c.appendItemRecord(bracketSpan, ItemText)

		delim := &DelimiterRecord{
			Item:   item,
			Kind:   DelimOpenBracket,
			Active: true,
		}

		c.Delimiters.PushBack(delim)

	case EscapeLiteralize:
		c.Next()
		c.appendItemRecord(next.Span, ItemText)

	case EscapeLiteralizeLeadingByte:
		leadingByteSpan := source.ByteSpan{
			Start: next.Span.Start,
			End:   next.Span.Start + 1,
		}

		c.appendItemRecord(leadingByteSpan, ItemText)

		// consume only the escaped leading byte and leave the remainder of the token in place
		if next.Span.End == leadingByteSpan.End {
			c.Next()
		} else {
			c.Tokens[c.Index].Span.Start = leadingByteSpan.End
		}

	case EscapeNone:
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

// ResolvedBracketResult captures the resolved span and destination/title
// data for a successfully parsed link or image construct.
type ResolvedBracketResult struct {
	FullSpan        source.ByteSpan
	DestinationSpan source.ByteSpan
	TitleSpan       source.ByteSpan
	HasTitle        bool
}

// tryResolveBracket attempts the supported bracket resolution forms in
// precedence order and finalizes the first successful match.
func (c *Cursor) tryResolveBracket(opener *DelimiterRecord, token Token, kind ItemKind) bool {
	tryFns := []func(*DelimiterRecord, Token) (ResolvedBracketResult, bool){
		c.tryParseInlineBracket,
		c.tryParseFullReference,
		c.tryParseCollapsedReference,
		c.tryParseShortcutReference,
	}

	for _, fn := range tryFns {
		if result, ok := fn(opener, token); ok {
			cmd := FinalizeBracketCommand{
				Kind:           kind,
				OpenerDelim:    opener,
				CloseTokenSpan: token.Span,
				Result:         result,
			}
			c.finalizeBracket(cmd)
			return true
		}
	}

	return false
}

// tryParseInlineBracket attempts to parse an inline link or image tail
// following a closing bracket.
func (c *Cursor) tryParseInlineBracket(opener *DelimiterRecord, token Token) (ResolvedBracketResult, bool) {
	tail, ok := c.tryParseInlineLinkTail(token.Span.End)
	if !ok {
		return ResolvedBracketResult{}, false
	}

	result := ResolvedBracketResult{
		FullSpan: source.ByteSpan{
			Start: opener.Item.OriginalSpan.Start,
			End:   tail.FullSpan.End,
		},
		DestinationSpan: tail.DestinationSpan,
		TitleSpan:       tail.TitleSpan,
		HasTitle:        tail.HasTitle,
	}

	return result, true
}

// tryParseFullReference attempts to resolve a full reference tail of the
// form [label] using the definition map.
func (c *Cursor) tryParseFullReference(opener *DelimiterRecord, token Token) (ResolvedBracketResult, bool) {
	tailStart := token.Span.End

	s := c.Source.Slice(source.ByteSpan{
		Start: tailStart,
		End:   c.Source.EOF(),
	})

	pos := 0
	if pos >= len(s) {
		return ResolvedBracketResult{}, false
	}

	if s[pos] != '[' {
		return ResolvedBracketResult{}, false
	}

	labelSpanStart := pos
	pos++

	foundClose := false
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
			foundClose = true
			break
		}

		pos++
	}

	if !foundClose {
		return ResolvedBracketResult{}, false
	}

	labelSpanEnd := pos

	labelSpan := source.ByteSpan{
		Start: tailStart + source.BytePos(labelSpanStart),
		End:   tailStart + source.BytePos(labelSpanEnd),
	}

	contentSpan := source.ByteSpan{
		Start: labelSpan.Start + 1,
		End:   labelSpan.End - 1,
	}

	labelContent := c.Source.Slice(contentSpan)

	if ok := reference.ValidateLabel(labelContent); !ok {
		return ResolvedBracketResult{}, false
	}

	normalizedLabel := reference.NormalizeLabel(labelContent)

	def, exists := c.Definitions[normalizedLabel]
	if !exists {
		return ResolvedBracketResult{}, false
	}

	result := ResolvedBracketResult{
		FullSpan: source.ByteSpan{
			Start: opener.Item.OriginalSpan.Start,
			End:   labelSpan.End,
		},
		DestinationSpan: def.DestinationSpan,
		TitleSpan:       def.TitleSpan,
		HasTitle:        def.HasTitle,
	}

	return result, true
}

// tryParseCollapsedReference attempts to resolve a collapsed reference tail
// of the form [] using the bracket contents as the label.
func (c *Cursor) tryParseCollapsedReference(opener *DelimiterRecord, token Token) (ResolvedBracketResult, bool) {
	tailStart := token.Span.End

	s := c.Source.Slice(source.ByteSpan{
		Start: tailStart,
		End:   c.Source.EOF(),
	})

	if len(s) < 2 {
		return ResolvedBracketResult{}, false
	}

	if s[0] != '[' || s[1] != ']' {
		return ResolvedBracketResult{}, false
	}

	tailEnd := tailStart + 2

	contentSpan := source.ByteSpan{
		Start: opener.Item.OriginalSpan.End,
		End:   token.Span.Start,
	}

	labelContent := c.Source.Slice(contentSpan)

	if ok := reference.ValidateLabel(labelContent); !ok {
		return ResolvedBracketResult{}, false
	}

	normalizedLabel := reference.NormalizeLabel(labelContent)

	def, exists := c.Definitions[normalizedLabel]
	if !exists {
		return ResolvedBracketResult{}, false
	}

	result := ResolvedBracketResult{
		FullSpan: source.ByteSpan{
			Start: opener.Item.OriginalSpan.Start,
			End:   tailEnd,
		},
		DestinationSpan: def.DestinationSpan,
		TitleSpan:       def.TitleSpan,
		HasTitle:        def.HasTitle,
	}

	return result, true
}

// tryParseShortcutReference attempts to resolve a shortcut reference using
// the bracket contents as the label.
func (c *Cursor) tryParseShortcutReference(opener *DelimiterRecord, token Token) (ResolvedBracketResult, bool) {
	contentSpan := source.ByteSpan{
		Start: opener.Item.OriginalSpan.End,
		End:   token.Span.Start,
	}

	labelContent := c.Source.Slice(contentSpan)

	if ok := reference.ValidateLabel(labelContent); !ok {
		return ResolvedBracketResult{}, false
	}

	normalizedLabel := reference.NormalizeLabel(labelContent)

	def, exists := c.Definitions[normalizedLabel]
	if !exists {
		return ResolvedBracketResult{}, false
	}

	result := ResolvedBracketResult{
		FullSpan: source.ByteSpan{
			Start: opener.Item.OriginalSpan.Start,
			End:   token.Span.End,
		},
		DestinationSpan: def.DestinationSpan,
		TitleSpan:       def.TitleSpan,
		HasTitle:        def.HasTitle,
	}

	return result, true
}

// FinalizeBracketCommand describes the data needed to rewrite an opening
// bracket item into a resolved link or image item.
type FinalizeBracketCommand struct {
	Kind           ItemKind
	OpenerDelim    *DelimiterRecord
	CloseTokenSpan source.ByteSpan
	Result         ResolvedBracketResult
}

// finalizeBracket rewrites the opening bracket item into a resolved link or
// image item, attaches its child items, and updates delimiter state.
func (c *Cursor) finalizeBracket(cmd FinalizeBracketCommand) {
	openerItem := cmd.OpenerDelim.Item

	originalSpan := source.ByteSpan{
		Start: openerItem.OriginalSpan.Start,
		End:   cmd.Result.FullSpan.End,
	}

	liveSpan := source.ByteSpan{
		Start: openerItem.OriginalSpan.End,
		End:   cmd.CloseTokenSpan.Start,
	}

	// process emphasis beginning from the opening bracket
	c.processEmphasis(cmd.OpenerDelim)

	firstChild := openerItem.Next()
	lastChild := c.Items.Back()

	var childList *ItemList
	if firstChild == nil {
		childList = NewItemList()
	} else {
		childList = c.Items.DetachRange(firstChild, lastChild)
	}

	openerItem.Kind = cmd.Kind
	openerItem.OriginalSpan = originalSpan
	openerItem.LiveSpan = liveSpan
	openerItem.DestinationSpan = cmd.Result.DestinationSpan
	openerItem.TitleSpan = source.ByteSpan{}
	openerItem.HasTitle = cmd.Result.HasTitle
	if cmd.Result.HasTitle {
		openerItem.TitleSpan = cmd.Result.TitleSpan
	}
	openerItem.Children = childList

	// for links, deactivate all prior '[' delimiters
	if cmd.Kind == ItemLink {
		for delim := cmd.OpenerDelim.Prev(); delim != nil; delim = delim.Prev() {
			if delim.Kind == DelimOpenBracket {
				delim.Active = false
			}
		}
	}

	c.advanceToBytePos(cmd.Result.FullSpan.End)
	c.Delimiters.Remove(cmd.OpenerDelim)
}

// InlineLinkTail represents a parsed inline link tail "(destination title)".
type InlineLinkTail struct {
	FullSpan        source.ByteSpan // from '(' through ')'
	DestinationSpan source.ByteSpan
	TitleSpan       source.ByteSpan
	HasTitle        bool
}

// tryParseInlineLinkTail parses an inline link tail starting at start.
func (c *Cursor) tryParseInlineLinkTail(start source.BytePos) (InlineLinkTail, bool) {
	candidateSpan := source.ByteSpan{
		Start: start,
		End:   c.Source.EOF(),
	}
	s := c.Source.Slice(candidateSpan)
	limit := len(s)

	result := InlineLinkTail{}

	if limit < 2 || s[0] != '(' {
		return InlineLinkTail{}, false
	}

	idx := 1

	idx = consumeSpacesTabs(s, idx, limit)

	if idx < limit && s[idx] == ')' {
		result.FullSpan = source.ByteSpan{
			Start: candidateSpan.Start,
			End:   candidateSpan.Start + source.BytePos(idx+1),
		}

		return result, true
	}

	destinationSpanRel, idx, ok := tryLinkDestination(s, idx, limit)
	if !ok {
		return InlineLinkTail{}, false
	}

	result.DestinationSpan = source.ByteSpan{
		Start: candidateSpan.Start + destinationSpanRel.Start,
		End:   candidateSpan.Start + destinationSpanRel.End,
	}

	sepStart := idx
	idx = consumeSpacesTabs(s, idx, limit)
	sepPresent := idx > sepStart

	if idx < limit && s[idx] == ')' {
		result.FullSpan = source.ByteSpan{
			Start: candidateSpan.Start,
			End:   candidateSpan.Start + source.BytePos(idx+1),
		}

		return result, true
	}

	if idx >= limit {
		return InlineLinkTail{}, false
	}

	if s[idx] == '"' || s[idx] == '\'' || s[idx] == '(' {
		// if no separator exists between destination and title, the tail is invalid
		if !sepPresent {
			return InlineLinkTail{}, false
		}

		titleSpanRel, idx, ok := tryLinkTitle(s, idx, limit)
		if !ok {
			return InlineLinkTail{}, false
		}

		result.TitleSpan = source.ByteSpan{
			Start: candidateSpan.Start + titleSpanRel.Start,
			End:   candidateSpan.Start + titleSpanRel.End,
		}
		result.HasTitle = true

		idx = consumeSpacesTabs(s, idx, limit)

		if idx < limit && s[idx] == ')' {
			result.FullSpan = source.ByteSpan{
				Start: candidateSpan.Start,
				End:   candidateSpan.Start + source.BytePos(idx+1),
			}

			return result, true
		}
	}

	return InlineLinkTail{}, false
}

// tryLinkDestination parses a link destination, selecting angle-bracketed
// or bare forms.
func tryLinkDestination(s string, idx, limit int) (source.ByteSpan, int, bool) {
	if idx >= limit {
		return source.ByteSpan{}, 0, false
	}

	if s[idx] == '<' {
		return tryAngleLinkDestination(s, idx, limit)
	}
	return tryBareLinkDestination(s, idx, limit)
}

// tryAngleLinkDestination parses an angle-bracketed link destination "<...>"
// with stricter character constraints and no nesting.
func tryAngleLinkDestination(s string, idx, limit int) (source.ByteSpan, int, bool) {
	if idx >= limit || s[idx] != '<' {
		return source.ByteSpan{}, 0, false
	}

	idx++
	start := idx

	for idx < limit {
		switch s[idx] {
		case '\n', '\r':
			return source.ByteSpan{}, 0, false

		case '<':
			return source.ByteSpan{}, 0, false

		case '>':
			span := source.ByteSpan{
				Start: source.BytePos(start),
				End:   source.BytePos(idx),
			}

			return span, idx + 1, true

		case '\\':
			if idx+1 < limit {
				idx += 2
				continue
			}
			idx++

		default:
			idx++
		}
	}

	return source.ByteSpan{}, 0, false
}

// tryBareLinkDestination parses a non-angle-bracketed link destination,
// handling nested parentheses.
func tryBareLinkDestination(s string, idx, limit int) (source.ByteSpan, int, bool) {
	if idx >= limit {
		return source.ByteSpan{}, 0, false
	}

	if s[idx] == '<' {
		return source.ByteSpan{}, 0, false
	}

	start := idx
	depth := 0

	for idx < limit {
		b := s[idx]

		// termination conditions (do not consume)
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
			break
		}
		if b == ')' && depth == 0 {
			break
		}

		if b < 0x20 || b == 0x7F {
			return source.ByteSpan{}, 0, false
		}

		switch b {
		case '\\':
			if idx+1 < limit {
				idx += 2
			} else {
				idx++
			}

		case '(':
			depth++
			idx++

		case ')':
			depth--
			idx++

		default:
			idx++
		}
	}

	if idx == start {
		return source.ByteSpan{}, 0, false
	}

	if depth != 0 {
		return source.ByteSpan{}, 0, false
	}

	span := source.ByteSpan{
		Start: source.BytePos(start),
		End:   source.BytePos(idx),
	}

	return span, idx, true
}

// tryLinkTitle parses a link title delimited by quotes or parentheses.
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
			return source.ByteSpan{}, 0, false

		case delim:
			span := source.ByteSpan{
				Start: source.BytePos(start),
				End:   source.BytePos(idx),
			}

			return span, idx + 1, true

		case '\\':
			if idx+1 < limit {
				idx += 2
				continue
			}
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
			return source.ByteSpan{}, 0, false

		case '(':
			depth++
			idx++

		case ')':
			depth--

			if depth == 0 {
				span := source.ByteSpan{
					Start: source.BytePos(start),
					End:   source.BytePos(idx),
				}

				return span, idx + 1, true
			}

			idx++

		case '\\':
			if idx+1 < limit {
				idx += 2
				continue
			}
			idx++

		default:
			idx++
		}
	}

	return source.ByteSpan{}, 0, false
}

// runeBefore returns the rune immediately preceding delimSpan within the
// current inline parse span.
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

// runeAfter returns the rune immediately following delimSpan within the
// current inline parse span.
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

// validateURIAutolink reports whether s is a valid URI autolink body.
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

	if len(scheme) < 2 || len(scheme) > 32 {
		return false
	}

	b := scheme[0]
	if !isAlpha(b) {
		return false
	}

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

// validateEmailAutolink reports whether s is a valid email autolink body.
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

// tryInlineHTML attempts to parse any supported inline HTML construct and
// returns its width on success.
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

// tryHTMLOpenTag attempts to parse an inline HTML opening tag, including
// optional attributes and self-closing suffixes.
func tryHTMLOpenTag(s string) (int, bool) {
	if len(s) < 2 {
		return 0, false
	}

	pos := 1
	insideSingleQuote := false
	insideDoubleQuote := false

	for pos < len(s) {
		b := s[pos]
		if b == '"' && !insideSingleQuote {
			insideDoubleQuote = !insideDoubleQuote
		}

		if b == '\'' && !insideDoubleQuote {
			insideSingleQuote = !insideSingleQuote
		}

		if b == '>' && !insideSingleQuote && !insideDoubleQuote {
			break
		}

		pos++
	}

	if pos == len(s) {
		return 0, false
	}

	last := pos
	width := pos + 1
	candidate := s[:width]

	if candidate[0] != '<' || candidate[last] != '>' {
		return 0, false
	}

	idx := 1
	idx, ok := tryHTMLTagName(s, idx, last)
	if !ok {
		return 0, false
	}

	if idx == last {
		return width, true
	}

	if candidate[idx] == '/' {
		if _, ok := tryHTMLSelfClosingSuffix(candidate, idx, last); ok {
			return width, true
		}
		return 0, false
	}

	for {
		mark := idx

		idx = consumeSpacesTabs(candidate, idx, last)

		if idx == last {
			return width, true
		}

		if candidate[idx] == '/' {
			if _, ok := tryHTMLSelfClosingSuffix(candidate, idx, last); ok {
				return width, true
			}
			return 0, false
		}

		if mark == idx {
			return 0, false
		}

		next, ok := tryHTMLAttribute(candidate, idx, last)

		if !ok || next <= idx {
			return 0, false
		}

		idx = next
	}
}

// tryHTMLClosingTag attempts to parse an inline HTML closing tag.
func tryHTMLClosingTag(s string) (int, bool) {
	if len(s) < 3 {
		return 0, false
	}

	if s[0] != '<' || s[1] != '/' {
		return 0, false
	}

	pos := 2
	for pos < len(s) {
		if s[pos] == '>' {
			break
		}
		pos++
	}

	if pos == len(s) {
		return 0, false
	}

	last := pos
	width := pos + 1
	candidate := s[:width]

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
	if !isAlpha(s[idx]) {
		return 0, false
	}
	idx++

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
	if idx >= last || s[idx] != '/' {
		return 0, false
	}
	idx++

	idx = consumeSpacesTabs(s, idx, last)

	if idx != last {
		return 0, false
	}

	return idx, true

}

// tryHTMLAttribute attempts to parse a single HTML attribute, including an
// optional value.
func tryHTMLAttribute(s string, idx, last int) (int, bool) {
	var ok bool
	idx, ok = tryHTMLAttributeName(s, idx, last)
	if !ok {
		return 0, false
	}

	if idx == last {
		return idx, true
	}

	if s[idx] != ' ' && s[idx] != '\t' && s[idx] != '=' {
		return 0, false
	}

	// if no '=' follows, the attribute is a bare name and trailing whitespace is left for the outer parser
	probe := consumeSpacesTabs(s, idx, last)
	if probe == last || s[probe] != '=' {
		return idx, true
	}

	idx = consumeSpacesTabs(s, probe+1, last)

	if idx == last {
		return 0, false
	}

	return tryHTMLAttributeValue(s, idx, last)
}

func tryHTMLAttributeName(s string, idx, last int) (int, bool) {
	if !isAlpha(s[idx]) && s[idx] != '_' && s[idx] != ':' {
		return 0, false
	}
	idx++

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

	return idx, true
}

func tryHTMLAttributeValue(s string, idx, last int) (int, bool) {
	switch s[idx] {
	case '\'':
		idx++
		for idx < last {
			if s[idx] != '\'' {
				idx++
				continue
			}
			break
		}

		if idx == last {
			return 0, false
		}

		idx++
		return idx, true

	case '"':
		idx++
		for idx < last {
			if s[idx] != '"' {
				idx++
				continue
			}
			break
		}

		if idx == last {
			return 0, false
		}

		idx++
		return idx, true

	default:
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

		if idx == start {
			return 0, false
		}

		return idx, true
	}
}

// tryHTMLDelimited attempts to parse a delimited inline HTML form with the
// given opener and terminator.
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

// consumeSpacesTabs advances idx past consecutive spaces and tabs up to last.
func consumeSpacesTabs(s string, idx, last int) int {
	for idx < last && (s[idx] == ' ' || s[idx] == '\t') {
		idx++
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

// leftFlanking reports whether a delimiter run is left-flanking with
// respect to the surrounding runes.
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

// rightFlanking reports whether a delimiter run is right-flanking with
// respect to the surrounding runes.
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
