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

	for i := range doc.Blocks {
		switch b := doc.Blocks[i].(type) {
		case ir.Header:
			b.Span = source.ByteSpan{}
			b.ContentSpan = source.ByteSpan{}
			doc.Blocks[i] = b
		case ir.ThematicBreak:
			b.Span = source.ByteSpan{}
			doc.Blocks[i] = b
		case ir.Paragraph:
			if b.Lines == nil {
				b.Lines = []source.ByteSpan{}
			}

			b.Span = source.ByteSpan{}
			for j := range b.Lines {
				b.Lines[j] = source.ByteSpan{}
			}
			doc.Blocks[i] = b
		}
	}

	return doc
}
