package testkit

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

func IRDoc(blocks ...ir.Block) ir.Document {
	return ir.Document{
		Blocks: blocks,
	}
}

func IRBlockQuote(children ...ir.Block) ir.BlockQuote {
	return ir.BlockQuote{
		Children: children,
		Span:     source.ByteSpan{},
	}
}

func IRHeader(level int, input ...string) ir.Header {
	return ir.Header{
		Level:       level,
		Span:        source.ByteSpan{},
		ContentSpan: source.ByteSpan{},
	}
}

func IRThematicBreak() ir.ThematicBreak {
	return ir.ThematicBreak{
		Span: source.ByteSpan{},
	}
}

func IRPara(input ...string) ir.Paragraph {
	lines := make([]source.ByteSpan, len(input))

	return ir.Paragraph{
		Lines: lines,
		Span:  source.ByteSpan{},
	}
}

func NormalizeIR(doc ir.Document) ir.Document {
	doc.Source = nil
	if doc.Blocks == nil {
		doc.Blocks = []ir.Block{}
	}

	doc.Blocks = normalizeBlocks(doc.Blocks)

	return doc
}

func normalizeBlocks(blocks []ir.Block) []ir.Block {
	for i := range blocks {
		switch b := blocks[i].(type) {
		case ir.BlockQuote:
			b.Span = source.ByteSpan{}
			if b.Children == nil {
				b.Children = []ir.Block{}
			}
			b.Children = normalizeBlocks(b.Children)
			blocks[i] = b
		case ir.Header:
			b.Span = source.ByteSpan{}
			b.ContentSpan = source.ByteSpan{}
			blocks[i] = b
		case ir.ThematicBreak:
			b.Span = source.ByteSpan{}
			blocks[i] = b
		case ir.Paragraph:
			if b.Lines == nil {
				b.Lines = []source.ByteSpan{}
			}

			b.Span = source.ByteSpan{}
			for j := range b.Lines {
				b.Lines[j] = source.ByteSpan{}
			}
			blocks[i] = b
		}
	}

	return blocks
}
