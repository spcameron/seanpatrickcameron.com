package render

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/html"
)

func HTML(doc ast.Document) (html.Node, error) {
	rootNode := html.Fragment{
		Children: make([]html.Node, 0, len(doc.Blocks)),
	}

	for _, v := range doc.Blocks {
		node, err := renderBlock(v)
		if err != nil {
			return nil, err
		}

		rootNode.Children = append(rootNode.Children, node)
	}

	return rootNode, nil
}

func renderBlock(block ast.Block) (html.Node, error) {
	switch v := block.(type) {
	case ast.Paragraph:
		return renderParagraph(v)
	case ast.Header:
		return renderHeader(v)
	default:
		return nil, fmt.Errorf("unrecognized block type: %T", block)
	}
}

func appendChild(children []html.Node, child html.Node) []html.Node {
	// drop empty text nodes
	if t, ok := child.(html.Text); ok && t.Value == "" {
		return children
	}

	// merge adjacent text nodes
	if t, ok := child.(html.Text); ok && len(children) > 0 {
		if last, ok := children[len(children)-1].(html.Text); ok {
			last.Value += t.Value
			children[len(children)-1] = last
			return children
		}
	}

	return append(children, child)
}

func renderHeader(v ast.Header) (html.Node, error) {
	node := html.Element{
		Tag:      fmt.Sprintf("h%d", v.Level),
		Attr:     html.Attributes{},
		Children: make([]html.Node, 0, len(v.Inlines)),
	}

	for _, inl := range v.Inlines {
		child, err := renderInline(inl)
		if err != nil {
			return nil, err
		}

		node.Children = appendChild(node.Children, child)
	}

	return node, nil
}

func renderParagraph(v ast.Paragraph) (html.Node, error) {
	node := html.Element{
		Tag:      "p",
		Attr:     html.Attributes{},
		Children: make([]html.Node, 0, len(v.Inlines)),
	}

	for _, inl := range v.Inlines {
		child, err := renderInline(inl)
		if err != nil {
			return nil, err
		}

		node.Children = appendChild(node.Children, child)
	}

	return node, nil
}

func renderInline(inline ast.Inline) (html.Node, error) {
	switch v := inline.(type) {
	case ast.Text:
		return renderText(v)
	case ast.SoftBreak:
		return renderSoftBreak(v)
	case ast.HardBreak:
		return renderHardBreak(v)
	default:
		return nil, fmt.Errorf("unrecognized inline type: %T", inline)
	}
}

func renderText(v ast.Text) (html.Node, error) {
	node := html.Text{
		Value: v.Value,
	}

	return node, nil
}

func renderSoftBreak(_ ast.SoftBreak) (html.Node, error) {
	node := html.Text{
		Value: " ",
	}

	return node, nil
}

func renderHardBreak(_ ast.HardBreak) (html.Node, error) {
	node := html.VoidElement{
		Tag:  "br",
		Attr: html.Attributes{},
	}

	return node, nil
}
