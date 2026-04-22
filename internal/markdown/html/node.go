package html

import (
	"html"
	"io"
)

// Node represents a renderable HTML node.
type Node interface {
	isNode()
	Write(io.Writer) error
}

// Fragment is a container node that renders its children in sequence
// without introducing additional markup.
type Fragment struct {
	Children []Node
}

func (Fragment) isNode() {}

func (f Fragment) Write(w io.Writer) error {
	for _, c := range f.Children {
		if err := c.Write(w); err != nil {
			return err
		}
	}

	return nil
}

// Text represents escaped text content.
type Text struct {
	Value string
}

func (Text) isNode() {}

func (t Text) Write(w io.Writer) error {
	s := html.EscapeString(t.Value)
	_, err := io.WriteString(w, s)
	return err
}

// Raw represents unescaped HTML content.
type Raw struct {
	Value string
}

func (Raw) isNode() {}

func (r Raw) Write(w io.Writer) error {
	_, err := io.WriteString(w, r.Value)
	return err
}

type VoidElement struct {
	Tag  string
	Attr Attributes
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
