package ast

type Inline interface {
	isInline()
}

type Text struct {
	Value string
}

func (Text) isInline() {}

type HardBreak struct{}

func (HardBreak) isInline() {}

type SoftBreak struct{}

func (SoftBreak) isInline() {}
