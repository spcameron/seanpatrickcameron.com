package testkit

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"

func Span(start, end int) source.ByteSpan {
	return source.ByteSpan{
		Start: source.BytePos(start),
		End:   source.BytePos(end),
	}
}

// SpanPtr returns a pointer to a ByteSpan, for use in tests that require
// addressable values.
func SpanPtr(start, end int) *source.ByteSpan {
	s := Span(start, end)
	return &s
}
