package ast

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"

type Inline interface {
	isInline()
}

type Em struct {
	Span     source.ByteSpan
	Children []Inline
}

func (Em) isInline() {}

type Strong struct {
	Span     source.ByteSpan
	Children []Inline
}

func (Strong) isInline() {}

type Text struct {
	Span source.ByteSpan
}

func (Text) isInline() {}

type RawText struct {
	Span source.ByteSpan
}

func (RawText) isInline() {}

type HardBreak struct {
	Span source.ByteSpan
}

func (HardBreak) isInline() {}

type SoftBreak struct {
	Span source.ByteSpan
}

func (SoftBreak) isInline() {}

type Newline struct {
	Span source.ByteSpan
}

func (Newline) isInline() {}
