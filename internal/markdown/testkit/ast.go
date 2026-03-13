package testkit

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

func ASTDoc(blocks ...ast.Block) ast.Document {
	return ast.Document{
		Blocks: blocks,
	}
}

func ASTBlockQuote(children ...ast.Block) ast.BlockQuote {
	return ast.BlockQuote{
		Span:     source.ByteSpan{},
		Children: children,
	}
}

func ASTHeader(level int, inlines ...ast.Inline) ast.Header {
	return ast.Header{
		Span:    source.ByteSpan{},
		Level:   level,
		Inlines: inlines,
	}
}

func ASTThematicBreak() ast.ThematicBreak {
	return ast.ThematicBreak{
		Span: source.ByteSpan{},
	}
}

func ASTOrderedList(tight bool, start int, items ...ast.ListItem) ast.OrderedList {
	return ast.OrderedList{
		Span:  source.ByteSpan{},
		Items: items,
		Tight: tight,
		Start: start,
	}
}

func ASTUnorderedList(tight bool, items ...ast.ListItem) ast.UnorderedList {
	return ast.UnorderedList{
		Span:  source.ByteSpan{},
		Items: items,
		Tight: tight,
	}
}

func ASTListItem(children ...ast.Block) ast.ListItem {
	return ast.ListItem{
		Span:     source.ByteSpan{},
		Children: children,
	}
}

func ASTIndentedCodeBlock(inlines ...ast.Inline) ast.CodeBlock {
	return ast.CodeBlock{
		Span:              source.ByteSpan{},
		Kind:              ast.Indented,
		LanguageTokenSpan: source.ByteSpan{},
		Payload:           inlines,
	}
}

func ASTFencedCodeBlock(inlines ...ast.Inline) ast.CodeBlock {
	return ast.CodeBlock{
		Span:              source.ByteSpan{},
		Kind:              ast.Fenced,
		LanguageTokenSpan: source.ByteSpan{},
		Payload:           inlines,
	}
}

func ASTHTMLBlock(inlines ...ast.Inline) ast.HTMLBlock {
	return ast.HTMLBlock{
		Span:    source.ByteSpan{},
		Payload: inlines,
	}
}

func ASTPara(inlines ...ast.Inline) ast.Paragraph {
	return ast.Paragraph{
		Span:    source.ByteSpan{},
		Inlines: inlines,
	}
}

func ASTText() ast.Text {
	return ast.Text{
		Span: source.ByteSpan{},
	}
}

func ASTTextAt(start, end int) ast.Text {
	span := source.ByteSpan{
		Start: source.BytePos(start),
		End:   source.BytePos(end),
	}

	return ast.Text{
		Span: span,
	}
}

func ASTRawText() ast.RawText {
	return ast.RawText{
		Span: source.ByteSpan{},
	}
}

func NormalizeAST(doc ast.Document) ast.Document {
	doc.Source = nil
	if doc.Blocks == nil {
		doc.Blocks = []ast.Block{}
	}

	doc.Blocks = NormalizeASTBlocks(doc.Blocks)

	return doc
}

func NormalizeASTBlocks(blocks []ast.Block) []ast.Block {
	for i := range blocks {
		switch b := blocks[i].(type) {
		case ast.BlockQuote:
			b.Span = source.ByteSpan{}
			if b.Children == nil {
				b.Children = []ast.Block{}
			}
			b.Children = NormalizeASTBlocks(b.Children)
			blocks[i] = b
		case ast.Header:
			b.Span = source.ByteSpan{}
			b.Inlines = NormalizeASTInlines(b.Inlines)
			blocks[i] = b
		case ast.ThematicBreak:
			b.Span = source.ByteSpan{}
			blocks[i] = b
		case ast.OrderedList:
			b.Span = source.ByteSpan{}
			if b.Items == nil {
				b.Items = []ast.ListItem{}
			}
			for j := range b.Items {
				item := b.Items[j]
				item.Span = source.ByteSpan{}
				if item.Children == nil {
					item.Children = []ast.Block{}
				}
				item.Children = NormalizeASTBlocks(item.Children)
				b.Items[j] = item
			}
			blocks[i] = b
		case ast.UnorderedList:
			b.Span = source.ByteSpan{}
			if b.Items == nil {
				b.Items = []ast.ListItem{}
			}
			for j := range b.Items {
				item := b.Items[j]
				item.Span = source.ByteSpan{}
				if item.Children == nil {
					item.Children = []ast.Block{}
				}
				item.Children = NormalizeASTBlocks(item.Children)
				b.Items[j] = item
			}
			blocks[i] = b
		case ast.ListItem:
			b.Span = source.ByteSpan{}
			if b.Children == nil {
				b.Children = []ast.Block{}
			}
			b.Children = NormalizeASTBlocks(b.Children)
			blocks[i] = b
		case ast.CodeBlock:
			b.Span = source.ByteSpan{}
			b.LanguageTokenSpan = source.ByteSpan{}
			b.Payload = NormalizeASTInlines(b.Payload)
			blocks[i] = b
		case ast.HTMLBlock:
			b.Span = source.ByteSpan{}
			b.Payload = NormalizeASTInlines(b.Payload)
			blocks[i] = b
		case ast.Paragraph:
			b.Span = source.ByteSpan{}
			b.Inlines = NormalizeASTInlines(b.Inlines)
			blocks[i] = b
		default:
			panic(fmt.Sprintf("unhandled block type %T", b))
		}
	}

	return blocks
}

func NormalizeASTInlines(inl []ast.Inline) []ast.Inline {
	out := make([]ast.Inline, 0, len(inl))
	for i := range inl {
		switch v := inl[i].(type) {
		case ast.Text:
			v.Span = source.ByteSpan{}
			out = append(out, v)
		case ast.RawText:
			v.Span = source.ByteSpan{}
			out = append(out, v)
		case ast.HardBreak:
			v.Span = source.ByteSpan{}
			out = append(out, v)
		case ast.SoftBreak:
			v.Span = source.ByteSpan{}
			out = append(out, v)
		case ast.Newline:
			v.Span = source.ByteSpan{}
			out = append(out, v)
		default:
			panic(fmt.Sprintf("unhandled inline type %T", v))
		}
	}

	return out
}
