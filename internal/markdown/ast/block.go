package ast

type Block interface {
	isBlock()
}

type Header struct {
	Level   int
	Inlines []Inline
}

func (Header) isBlock() {}

type Paragraph struct {
	Inlines []Inline
}

func (Paragraph) isBlock() {}
