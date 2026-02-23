package html

import (
	"html"
	"io"
)

type Text struct {
	Value string
}

func (Text) isNode() {}

func (t Text) Write(w io.Writer) error {
	s := html.EscapeString(t.Value)
	_, err := io.WriteString(w, s)
	return err
}

func TextNode(s string) Text {
	return Text{Value: s}
}
