package ast

type Block interface {
	isBlock()
}

type Paragraph struct {
	Inlines []Inline
}

func (Paragraph) isBlock() {}
