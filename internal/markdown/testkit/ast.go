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

func NormalizeAST(doc ast.Document) ast.Document {
	doc.Source = nil
	if doc.Blocks == nil {
		doc.Blocks = []ast.Block{}
	}

	for i := range doc.Blocks {
		b := doc.Blocks[i]
		doc.Blocks[i] = NormalizeASTBlock(b)
	}

	return doc
}

func NormalizeASTBlock(b ast.Block) ast.Block {
	switch v := b.(type) {
	case ast.BlockQuote:
		v.Span = source.ByteSpan{}
		for i := range v.Children {
			block := v.Children[i]
			v.Children[i] = NormalizeASTBlock(block)
		}
		return v
	case ast.Header:
		v.Span = source.ByteSpan{}
		v.Inlines = NormalizeASTInlines(v.Inlines)
		return v
	case ast.ThematicBreak:
		v.Span = source.ByteSpan{}
		return v
	case ast.Paragraph:
		v.Span = source.ByteSpan{}
		v.Inlines = NormalizeASTInlines(v.Inlines)
		return v
	default:
		panic(fmt.Sprintf("unhandled block type %T", v))
	}
}

func NormalizeASTInlines(inl []ast.Inline) []ast.Inline {
	out := make([]ast.Inline, 0, len(inl))
	for i := range inl {
		switch v := inl[i].(type) {
		case ast.Text:
			v.Span = source.ByteSpan{}
			out = append(out, v)
		case ast.HardBreak:
			v.Span = source.ByteSpan{}
			out = append(out, v)
		case ast.SoftBreak:
			v.Span = source.ByteSpan{}
			out = append(out, v)
		default:
			panic(fmt.Sprintf("unhandled inline type %T", v))
		}
	}

	return out
}
