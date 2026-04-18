package lower

import (
	"fmt"
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/block"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/inline"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type Context struct {
	Source      *source.Source
	Definitions map[string]ir.ReferenceDefinition
}

func Document(irDoc ir.Document) (ast.Document, error) {
	ctx := &Context{
		Source:      irDoc.Source,
		Definitions: irDoc.Definitions,
	}

	astDoc := ast.Document{
		Source: irDoc.Source,
		Blocks: make([]ast.Block, 0, len(irDoc.Blocks)),
	}

	for _, v := range irDoc.Blocks {
		block, err := buildBlock(ctx, v)
		if err != nil {
			return ast.Document{}, err
		}

		astDoc.Blocks = append(astDoc.Blocks, block)
	}

	return astDoc, nil
}

func buildBlock(ctx *Context, block ir.Block) (ast.Block, error) {
	switch v := block.(type) {
	case ir.BlockQuote:
		return buildBlockQuote(ctx, v)

	case ir.Header:
		return buildHeader(ctx, v)

	case ir.ThematicBreak:
		return buildThematicBreak(v)

	case ir.OrderedList:
		return buildOrderedList(ctx, v)

	case ir.UnorderedList:
		return buildUnorderedList(ctx, v)

	case ir.ListItem:
		return buildListItem(ctx, v)

	case ir.IndentedCodeBlock:
		return buildIndentedCodeBlock(ctx, v)

	case ir.FencedCodeBlock:
		return buildFencedCodeBlock(ctx, v)

	case ir.HTMLBlock:
		return buildHTMLBlock(v)

	case ir.Paragraph:
		return buildParagraph(ctx, v)

	default:
		return nil, fmt.Errorf("unrecognized block type: %T", block)
	}
}

func buildBlockQuote(ctx *Context, bq ir.BlockQuote) (ast.Block, error) {
	astChildren := make([]ast.Block, 0, len(bq.Children))

	for _, bqChild := range bq.Children {
		astChild, err := buildBlock(ctx, bqChild)
		if err != nil {
			return nil, err
		}

		astChildren = append(astChildren, astChild)
	}

	block := ast.BlockQuote{
		Span:     bq.Span,
		Children: astChildren,
	}

	return block, nil
}

func buildHeader(ctx *Context, h ir.Header) (ast.Block, error) {
	lines := ctx.Source.LineSpansWithin(h.ContentSpan)
	inlines, err := lowerLineSpans(ctx, lines)
	if err != nil {
		return nil, err
	}

	block := ast.Header{
		Span:    h.Span,
		Level:   h.Level,
		Inlines: inlines,
	}

	return block, nil
}

func buildThematicBreak(tb ir.ThematicBreak) (ast.Block, error) {
	block := ast.ThematicBreak{
		Span: tb.Span,
	}

	return block, nil
}

func buildUnorderedList(ctx *Context, ul ir.UnorderedList) (ast.Block, error) {
	astItems := make([]ast.ListItem, 0, len(ul.Items))

	for _, ulItem := range ul.Items {
		astBlock, err := buildBlock(ctx, ulItem)
		if err != nil {
			return nil, err
		}

		astItem, ok := astBlock.(ast.ListItem)
		if !ok {
			panic("ast build requires list item type assertion to pass")
		}

		astItems = append(astItems, astItem)
	}

	block := ast.UnorderedList{
		Span:  ul.Span,
		Items: astItems,
		Tight: ul.Tight,
	}

	return block, nil
}

func buildOrderedList(ctx *Context, ol ir.OrderedList) (ast.Block, error) {
	astItems := make([]ast.ListItem, 0, len(ol.Items))

	for _, olItem := range ol.Items {
		astBlock, err := buildBlock(ctx, olItem)
		if err != nil {
			return nil, err
		}

		astItem, ok := astBlock.(ast.ListItem)
		if !ok {
			panic("ast build requires list item type assertion to pass")
		}

		astItems = append(astItems, astItem)
	}

	block := ast.OrderedList{
		Span:  ol.Span,
		Items: astItems,
		Tight: ol.Tight,
		Start: ol.Start,
	}

	return block, nil
}

func buildListItem(ctx *Context, li ir.ListItem) (ast.Block, error) {
	astChildren := make([]ast.Block, 0, len(li.Children))

	for _, liChild := range li.Children {
		astChild, err := buildBlock(ctx, liChild)
		if err != nil {
			return nil, err
		}

		astChildren = append(astChildren, astChild)
	}

	block := ast.ListItem{
		Span:     li.Span,
		Children: astChildren,
	}

	return block, nil
}

func buildIndentedCodeBlock(ctx *Context, cb ir.IndentedCodeBlock) (ast.Block, error) {
	payload := normalizeCodeBlockPayload(ctx.Source, cb.Lines, block.MinValidCodeBlockIndentation)

	block := ast.CodeBlock{
		Span:              cb.Span,
		Kind:              ast.Indented,
		LanguageTokenSpan: source.ByteSpan{},
		Payload:           payload,
	}

	return block, nil
}

func buildFencedCodeBlock(ctx *Context, cb ir.FencedCodeBlock) (ast.Block, error) {
	payload := normalizeCodeBlockPayload(ctx.Source, cb.Lines, cb.OpenIndentCols)
	languageString := extractLanguageString(ctx.Source, cb.InfoStringSpan)

	block := ast.CodeBlock{
		Span:              cb.Span,
		Kind:              ast.Fenced,
		LanguageTokenSpan: languageString,
		Payload:           payload,
	}

	return block, nil
}

func normalizeCodeBlockPayload(src *source.Source, lines []source.ByteSpan, indent int) []ast.Inline {
	if len(lines) == 0 {
		return []ast.Inline{}
	}

	payload := make([]ast.Inline, 0, len(lines)*2-1)
	last := len(lines) - 1

	for i, ls := range lines {
		s := src.Slice(ls)
		pos := 0
		col := 0

		for pos < len(s) && col < indent {
			b := s[pos]
			if b == ' ' {
				pos++
				col++
				continue
			}
			if b == '\t' {
				pos++
				col += source.TabWidth - (col % source.TabWidth)
				continue
			}
			break
		}

		trimmed := ast.Text{
			Span: source.ByteSpan{
				Start: ls.Start + source.BytePos(pos),
				End:   ls.End,
			},
		}

		payload = append(payload, trimmed)

		if i < last {
			anchor := source.ByteSpan{
				Start: ls.End,
				End:   ls.End,
			}

			payload = append(payload, ast.Newline{Span: anchor})
		}
	}

	return payload
}

func extractLanguageString(src *source.Source, infoSpan source.ByteSpan) source.ByteSpan {
	s := src.Slice(infoSpan)
	pos := 0

	for pos < len(s) {
		b := s[pos]
		switch b {
		case ' ', '\t':
			return source.ByteSpan{
				Start: infoSpan.Start,
				End:   infoSpan.Start + source.BytePos(pos),
			}
		default:
			pos++
			continue
		}
	}

	return infoSpan
}

func buildHTMLBlock(hb ir.HTMLBlock) (ast.Block, error) {
	payload := make([]ast.Inline, 0, len(hb.Lines)*2-1)
	last := len(hb.Lines) - 1

	for i, ls := range hb.Lines {
		payload = append(payload, ast.RawText{Span: ls})

		if i < last {
			anchor := source.ByteSpan{
				Start: ls.End,
				End:   ls.End,
			}

			payload = append(payload, ast.Newline{Span: anchor})
		}
	}

	block := ast.HTMLBlock{
		Span:    hb.Span,
		Payload: payload,
	}

	return block, nil
}

func buildParagraph(ctx *Context, p ir.Paragraph) (ast.Block, error) {
	inlines, err := lowerLineSpans(ctx, p.Lines)
	if err != nil {
		return nil, err
	}

	block := ast.Paragraph{
		Span:    p.Span,
		Inlines: inlines,
	}

	return block, nil
}

func lowerLineSpans(ctx *Context, spans []source.ByteSpan) ([]ast.Inline, error) {
	if len(spans) == 0 {
		return []ast.Inline{}, nil
	}

	inlines := []ast.Inline{}
	last := len(spans) - 1

	for i, ls := range spans {
		ps := ls
		hardBreak := false

		if i < last {
			s := ctx.Source.Slice(ls)
			if strings.HasSuffix(s, "  ") {
				hardBreak = true
				ps.End = ps.End - 2
			} else if strings.HasSuffix(s, "\\") {
				hardBreak = true
				ps.End = ps.End - 1
			}
		}

		lineInlines, err := inline.Parse(ctx.Source, ctx.Definitions, ps)
		if err != nil {
			return nil, err
		}

		inlines = append(inlines, lineInlines...)

		if i < last {
			anchor := source.ByteSpan{
				Start: ls.End,
				End:   ls.End,
			}

			if hardBreak {
				inlines = append(inlines, ast.HardBreak{Span: anchor})
			} else {
				inlines = append(inlines, ast.SoftBreak{Span: anchor})
			}
		}
	}

	return inlines, nil
}
