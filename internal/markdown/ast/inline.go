package ast

type Inline interface {
	isInline()
}

type Text struct {
	Value string
}

func (Text) isInline() {}
