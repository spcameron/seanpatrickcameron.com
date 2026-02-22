package html

import "io"

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

func VoidNode(tag string, attr Attributes) VoidElement {
	if attr == nil {
		attr = Attributes{}
	}

	return VoidElement{
		Tag:  tag,
		Attr: attr,
	}
}
