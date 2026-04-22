package ast

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

// Inline is the marker interface implemented by all inline AST nodes.
type Inline interface {
	isInline()
}

type CodeSpan struct {
	Span source.ByteSpan
}

func (CodeSpan) isInline() {}

func (c CodeSpan) String() string {
	return "CodeSpan"
}

// Link represents an inline link or autolink.
//
// Label, Destination, and Title refer to source spans in the original input.
// Children holds the parsed inline label content. MailTo reports whether the
// rendered destination should be treated as a mailto link.
type Link struct {
	Span        source.ByteSpan
	Label       source.ByteSpan
	Destination source.ByteSpan
	Title       source.ByteSpan
	MailTo      bool
	Children    []Inline
}

func (Link) isInline() {}

func (l Link) String() string {
	return fmt.Sprintf(
		"Link(mailto=%t,children=%s)",
		l.MailTo,
		summarizeInlines(l.Children),
	)
}

// Image represents an inline image.
//
// Destination and Title refer to the image destination and optional title.
// Children holds the parsed inline content of the image label, which is used
// as alt text during rendering.
type Image struct {
	Span        source.ByteSpan
	Destination source.ByteSpan
	Title       source.ByteSpan
	Children    []Inline
}

func (i Image) isInline() {}

func (i Image) String() string {
	return fmt.Sprintf("Image(children=%s)", summarizeInlines(i.Children))
}

type Emph struct {
	Span     source.ByteSpan
	Children []Inline
}

func (Emph) isInline() {}

func (e Emph) String() string {
	return fmt.Sprintf("Emphasis(children=%s)", summarizeInlines(e.Children))
}

type Strong struct {
	Span     source.ByteSpan
	Children []Inline
}

func (Strong) isInline() {}

func (s Strong) String() string {
	return fmt.Sprintf("Strong(children=%s)", summarizeInlines(s.Children))
}

type Text struct {
	Span source.ByteSpan
}

func (Text) isInline() {}

func (Text) String() string {
	return "Text"
}

// RawText represents inline content that should be emitted without normal
// text escaping rules applied to Text nodes.
type RawText struct {
	Span source.ByteSpan
}

func (RawText) isInline() {}

func (RawText) String() string {
	return "RawText"
}

type HardBreak struct {
	Span source.ByteSpan
}

func (HardBreak) isInline() {}

func (HardBreak) String() string {
	return "HardBreak"
}

type SoftBreak struct {
	Span source.ByteSpan
}

func (SoftBreak) isInline() {}

func (SoftBreak) String() string {
	return "SoftBreak"
}

// Newline represents a literal newline retained in the inline AST.
type Newline struct {
	Span source.ByteSpan
}

func (Newline) isInline() {}

func (Newline) String() string {
	return "Newline"
}
