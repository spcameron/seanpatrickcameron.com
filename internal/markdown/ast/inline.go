package ast

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type Inline interface {
	isInline()
}

type Em struct {
	Span     source.ByteSpan
	Children []Inline
}

func (Em) isInline() {}

func (e Em) String() string {
	return fmt.Sprintf("[Emphasis] (Children = %d)", len(e.Children))
}

type Strong struct {
	Span     source.ByteSpan
	Children []Inline
}

func (Strong) isInline() {}

func (s Strong) String() string {
	return fmt.Sprintf("[Strong] (Children = %d)", len(s.Children))
}

type Text struct {
	Span source.ByteSpan
}

func (Text) isInline() {}

func (Text) String() string {
	return "[Text]"
}

type RawText struct {
	Span source.ByteSpan
}

func (RawText) isInline() {}

func (RawText) String() string {
	return "[RawText]"
}

type HardBreak struct {
	Span source.ByteSpan
}

func (HardBreak) isInline() {}

func (HardBreak) String() string {
	return "[HardBreak]"
}

type SoftBreak struct {
	Span source.ByteSpan
}

func (SoftBreak) isInline() {}

func (SoftBreak) String() string {
	return "[SoftBreak]"
}

type Newline struct {
	Span source.ByteSpan
}

func (Newline) isInline() {}

func (Newline) String() string {
	return "[Newline]"
}
