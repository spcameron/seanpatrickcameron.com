package testkit

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"

func Span(start, end int) source.ByteSpan {
	return source.ByteSpan{
		Start: source.BytePos(start),
		End:   source.BytePos(end),
	}
}

func SpanPtr(start, end int) *source.ByteSpan {
	s := Span(start, end)
	return &s
}
