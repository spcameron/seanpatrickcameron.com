package source

import "fmt"

type BytePos int

type ByteSpan struct {
	Start BytePos
	End   BytePos
}

func (s ByteSpan) String() string {
	return fmt.Sprintf("[%d:%d]", s.Start, s.End)
}

func (s ByteSpan) Width() int {
	return int(s.End - s.Start)
}
