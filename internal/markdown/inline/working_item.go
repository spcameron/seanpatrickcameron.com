package inline

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type InlineItem interface {
	isInlineItem()
}

type TextItem struct {
	Span source.ByteSpan
}

func (*TextItem) isInlineItem() {}

type DelimiterItem struct {
	Span      source.ByteSpan
	Delimiter byte
}

func (*DelimiterItem) isInlineItem() {}

type NodeItem struct {
	Span source.ByteSpan
	Node ast.Inline
}

func (*NodeItem) isInlineItem() {}

type ConsumedItem struct {
	Span source.ByteSpan
}

func (*ConsumedItem) isInlineItem() {}
