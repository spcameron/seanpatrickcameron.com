package html

import "io"

type Element struct {
	Tag      string
	Attr     Attributes
	Children []Node
}

func (Element) isNode() {}

func (e Element) Write(w io.Writer) error {
	if _, err := io.WriteString(w, "<"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, e.Tag); err != nil {
		return err
	}
	if err := e.Attr.Write(w); err != nil {
		return err
	}
	if _, err := io.WriteString(w, ">"); err != nil {
		return err
	}

	for _, c := range e.Children {
		if err := c.Write(w); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, "</"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, e.Tag); err != nil {
		return err
	}
	if _, err := io.WriteString(w, ">"); err != nil {
		return err
	}

	return nil
}

func ElemNode(tag string, attr Attributes, children ...Node) Element {
	if attr == nil {
		attr = Attributes{}
	}
	if children == nil {
		children = []Node{}
	}

	return Element{
		Tag:      tag,
		Attr:     attr,
		Children: children,
	}
}
