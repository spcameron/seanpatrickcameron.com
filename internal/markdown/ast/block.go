package ast

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"

type Block interface {
	isBlock()
}

type Header struct {
	Span    source.ByteSpan
	Level   int
	Inlines []Inline
}

func (Header) isBlock() {}

type Paragraph struct {
	Span    source.ByteSpan
	Inlines []Inline
}

func (Paragraph) isBlock() {}
