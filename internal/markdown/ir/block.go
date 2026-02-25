package ir

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"

type Block interface {
	isBlock()
}

type Header struct {
	Level       int
	Span        source.ByteSpan
	ContentSpan source.ByteSpan
}

func (Header) isBlock() {}

type Paragraph struct {
	Span  source.ByteSpan
	Lines []source.ByteSpan
}

func (Paragraph) isBlock() {}
