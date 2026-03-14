package inline

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type WorkingItem interface {
	isWorkingItem()
}

type TextItem struct {
	Span source.ByteSpan
}

func (*TextItem) isWorkingItem() {}

type DelimiterItem struct {
	Span      source.ByteSpan
	Delimiter byte
}

func (*DelimiterItem) isWorkingItem() {}

type NodeItem struct {
	Node ast.Inline
}

func (*NodeItem) isWorkingItem() {}
