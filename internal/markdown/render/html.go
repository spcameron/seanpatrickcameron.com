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
	default:
		return nil, fmt.Errorf("unrecognized block type: %T", block)
	}
}

func renderParagraph(p ast.Paragraph) (html.Node, error) {
	node := html.Element{
		Tag:      "p",
		Attr:     html.Attributes{},
		Children: make([]html.Node, 0, len(p.Inlines)),
	}

	for _, v := range p.Inlines {
		child, err := renderInline(v)
		if err != nil {
			return nil, err
		}

		node.Children = append(node.Children, child)
	}

	return node, nil
}

func renderInline(inline ast.Inline) (html.Node, error) {
	switch v := inline.(type) {
	case ast.Text:
		return renderText(v)
	default:
		return nil, fmt.Errorf("unrecognized inline type: %T", inline)
	}
}

func renderText(t ast.Text) (html.Node, error) {
	node := html.Text{
		Value: t.Value,
	}

	return node, nil
}
