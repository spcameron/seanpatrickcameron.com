package ast

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"

type Inline interface {
	isInline()
}

type Text struct {
	Span source.ByteSpan
}

func (Text) isInline() {}

type HardBreak struct {
	Span source.ByteSpan
}

func (HardBreak) isInline() {}

type SoftBreak struct {
	Span source.ByteSpan
}

func (SoftBreak) isInline() {}
