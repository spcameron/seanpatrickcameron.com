package testkit

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/html"

func HTMLFragmentNode(children ...html.Node) html.Fragment {
	if children == nil {
		children = []html.Node{}
	}

	return html.Fragment{
		Children: children,
	}
}

func HTMLTextNode(s string) html.Text {
	return html.Text{Value: s}
}

func HTMLRawNode(s string) html.Raw {
	return html.Raw{Value: s}
}

func HTMLVoidNode(tag string, attr html.Attributes) html.VoidElement {
	if attr == nil {
		attr = html.Attributes{}
	}

	return html.VoidElement{
		Tag:  tag,
		Attr: attr,
	}
}

func HTMLElemNode(tag string, attr html.Attributes, children ...html.Node) html.Element {
	if attr == nil {
		attr = html.Attributes{}
	}
	if children == nil {
		children = []html.Node{}
	}

	return html.Element{
		Tag:      tag,
		Attr:     attr,
		Children: children,
	}
}
