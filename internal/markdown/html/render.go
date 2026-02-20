package html

import (
	"errors"
	"fmt"
	"html"
	"strings"
)

var (
	ErrUnknownNode = errors.New("unknown node type")
)

func Render(node Node) (string, error) {
	switch v := node.(type) {
	case Text:
		return renderText(v)
	case Element:
		return renderElement(v)
	case VoidElement:
		return renderVoidElement(v)
	default:
		return "", ErrUnknownNode
	}
}

func renderText(node Text) (string, error) {
	out := html.EscapeString(node.Value)
	return out, nil
}

func renderElement(node Element) (string, error) {
	var sb strings.Builder

	fmt.Fprintf(&sb, "<%s%s>", node.Tag, renderAttributes(node))
	for _, n := range node.Children {
		s, err := Render(n)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&sb, "%s", s)
	}
	fmt.Fprintf(&sb, "</%s>", node.Tag)

	return sb.String(), nil
}

func renderVoidElement(node VoidElement) (string, error) {
	var sb strings.Builder

	fmt.Fprintf(&sb, "<%s%s>", node.Tag, renderAttributes(node))

	return sb.String(), nil
}

func renderAttributes(node Node) string {
	ha, ok := node.(HasAttrs)
	if !ok {
		return ""
	}

	return ha.Attrs().Render()
}
