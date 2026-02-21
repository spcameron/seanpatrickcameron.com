package html

import (
	"html"
	"io"
)

type Node interface {
	isNode()
	Write(io.Writer) error
}

type Raw struct {
	Value string
}

func RawNode(s string) Node {
	return Raw{Value: s}
}

func (Raw) isNode() {}

func (r Raw) Write(w io.Writer) error {
	_, err := io.WriteString(w, r.Value)
	return err
}

type Text struct {
	Value string
}

func TextNode(s string) Node {
	return Text{Value: s}
}

func (Text) isNode() {}

func (t Text) Write(w io.Writer) error {
	s := html.EscapeString(t.Value)
	_, err := io.WriteString(w, s)
	return err
}

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

func (VoidElement) isNode() {}

func (ve VoidElement) Write(w io.Writer) error {
	if _, err := io.WriteString(w, "<"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, ve.Tag); err != nil {
		return err
	}
	if err := ve.Attr.Write(w); err != nil {
		return err
	}
	if _, err := io.WriteString(w, ">"); err != nil {
		return err
	}

	return nil
}
