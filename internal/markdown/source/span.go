package source

import "fmt"

// BytePos is a byte offset into a normalized Source buffer.
type BytePos int

// ByteSpan represents a half-open byte range [Start, End).
//
// Both Start and End are byte offsets into the same Source. A span is
// valid when 0 <= Start <= End.
type ByteSpan struct {
	Start BytePos
	End   BytePos
}

func (s ByteSpan) String() string {
	return fmt.Sprintf("[%d:%d]", s.Start, s.End)
}

// Width reports the number of bytes covered by the span.
func (s ByteSpan) Width() int {
	return int(s.End - s.Start)
}
