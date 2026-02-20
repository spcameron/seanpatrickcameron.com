package html

type Node interface {
	isNode()
}

// TODO: add Raw struct (HTML passthrough)

type Text struct {
	Value string
}

func TextNode(s string) Node {
	return Text{Value: s}
}

func (t Text) isNode() {}

type Element struct {
	Tag      string
	Attr     Attributes
	Children []Node
}

func ElemNode(tag string, attr Attributes, children ...Node) Node {
	return Element{
		Tag:      tag,
		Attr:     attr,
		Children: children,
	}
}

func (e Element) isNode() {}

func (e Element) Attrs() Attributes {
	return e.Attr
}

type VoidElement struct {
	Tag  string
	Attr Attributes
}

func VoidNode(tag string, attr Attributes) Node {
	return VoidElement{
		Tag:  tag,
		Attr: attr,
	}
}

func (ve VoidElement) isNode() {}

func (ve VoidElement) Attrs() Attributes {
	return ve.Attr
}
