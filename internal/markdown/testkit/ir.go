package testkit

import (
	"fmt"

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
		Span:     source.ByteSpan{},
		Children: children,
	}
}

func IRHeader(level int, input ...string) ir.Header {
	return ir.Header{
		Span:        source.ByteSpan{},
		ContentSpan: source.ByteSpan{},
		Level:       level,
	}
}

func IRThematicBreak() ir.ThematicBreak {
	return ir.ThematicBreak{
		Span: source.ByteSpan{},
	}
}

func IROrderedList(tight bool, start int, items ...ir.ListItem) ir.OrderedList {
	return ir.OrderedList{
		Span:  source.ByteSpan{},
		Items: items,
		Tight: tight,
		Start: start,
	}
}

func IRUnorderedList(tight bool, items ...ir.ListItem) ir.UnorderedList {
	return ir.UnorderedList{
		Span:  source.ByteSpan{},
		Items: items,
		Tight: tight,
	}
}

func IRListItem(children ...ir.Block) ir.ListItem {
	return ir.ListItem{
		Span:     source.ByteSpan{},
		Children: children,
	}
}

func IRIndentedCodeBlock(input ...string) ir.IndentedCodeBlock {
	lines := make([]source.ByteSpan, len(input))

	return ir.IndentedCodeBlock{
		Span:  source.ByteSpan{},
		Lines: lines,
	}
}

func IRFencedCodeBlock(fence ir.CodeFence, input ...string) ir.FencedCodeBlock {
	lines := make([]source.ByteSpan, len(input))

	return ir.FencedCodeBlock{
		Span:  source.ByteSpan{},
		Lines: lines,
		Fence: fence,
	}
}

func IRCodeFence(indent int) ir.CodeFence {
	return ir.CodeFence{
		OpenIndentCols: indent,
		OpenFenceSpan:  source.ByteSpan{},
		CloseFenceSpan: source.ByteSpan{},
		InfoStringSpan: source.ByteSpan{},
	}
}

func IRPara(input ...string) ir.Paragraph {
	lines := make([]source.ByteSpan, len(input))

	return ir.Paragraph{
		Span:  source.ByteSpan{},
		Lines: lines,
	}
}

func NormalizeIR(doc ir.Document) ir.Document {
	doc.Source = nil
	if doc.Blocks == nil {
		doc.Blocks = []ir.Block{}
	}

	doc.Blocks = NormalizeIRBLocks(doc.Blocks)

	return doc
}

func NormalizeIRBLocks(blocks []ir.Block) []ir.Block {
	for i := range blocks {
		switch b := blocks[i].(type) {
		case ir.BlockQuote:
			b.Span = source.ByteSpan{}
			if b.Children == nil {
				b.Children = []ir.Block{}
			}
			b.Children = NormalizeIRBLocks(b.Children)
			blocks[i] = b
		case ir.Header:
			b.Span = source.ByteSpan{}
			b.ContentSpan = source.ByteSpan{}
			blocks[i] = b
		case ir.ThematicBreak:
			b.Span = source.ByteSpan{}
			blocks[i] = b
		case ir.UnorderedList:
			b.Span = source.ByteSpan{}
			if b.Items == nil {
				b.Items = []ir.ListItem{}
			}
			for j := range b.Items {
				item := b.Items[j]
				item.Span = source.ByteSpan{}
				if item.Children == nil {
					item.Children = []ir.Block{}
				}
				item.Children = NormalizeIRBLocks(item.Children)
				b.Items[j] = item
			}
			blocks[i] = b
		case ir.OrderedList:
			b.Span = source.ByteSpan{}
			if b.Items == nil {
				b.Items = []ir.ListItem{}
			}
			for j := range b.Items {
				item := b.Items[j]
				item.Span = source.ByteSpan{}
				if item.Children == nil {
					item.Children = []ir.Block{}
				}
				item.Children = NormalizeIRBLocks(item.Children)
				b.Items[j] = item
			}
			blocks[i] = b
		case ir.ListItem:
			b.Span = source.ByteSpan{}
			if b.Children == nil {
				b.Children = []ir.Block{}
			}
			b.Children = NormalizeIRBLocks(b.Children)
			blocks[i] = b
		case ir.IndentedCodeBlock:
			b.Span = source.ByteSpan{}
			if b.Lines == nil {
				b.Lines = []source.ByteSpan{}
			}
			for j := range b.Lines {
				b.Lines[j] = source.ByteSpan{}
			}
			blocks[i] = b
		case ir.Paragraph:
			b.Span = source.ByteSpan{}
			if b.Lines == nil {
				b.Lines = []source.ByteSpan{}
			}
			for j := range b.Lines {
				b.Lines[j] = source.ByteSpan{}
			}
			blocks[i] = b
		default:
			panic(fmt.Sprintf("unhandled block type %T", b))
		}
	}

	return blocks
}
