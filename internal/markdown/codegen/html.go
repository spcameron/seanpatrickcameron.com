package codegen

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/html"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

func HTML(doc ast.Document) (html.Node, error) {
	rootNode := html.Fragment{
		Children: make([]html.Node, 0, len(doc.Blocks)),
	}

	for _, v := range doc.Blocks {
		node, err := renderBlock(doc.Source, v)
		if err != nil {
			return nil, err
		}

		rootNode.Children = append(rootNode.Children, node)
	}

	return rootNode, nil
}

func renderBlock(src *source.Source, block ast.Block) (html.Node, error) {
	switch v := block.(type) {
	case ast.Header:
		return renderHeader(src, v)
	case ast.ThematicBreak:
		return renderThematicBreak()
	case ast.Paragraph:
		return renderParagraph(src, v)
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

func renderHeader(src *source.Source, h ast.Header) (html.Node, error) {
	node := html.Element{
		Tag:      fmt.Sprintf("h%d", h.Level),
		Attr:     html.Attributes{},
		Children: make([]html.Node, 0, len(h.Inlines)),
	}

	for _, inl := range h.Inlines {
		child, err := renderInline(src, inl)
		if err != nil {
			return nil, err
		}

		node.Children = appendChild(node.Children, child)
	}

	return node, nil
}

func renderThematicBreak() (html.Node, error) {
	node := html.VoidElement{
		Tag:  "hr",
		Attr: html.Attributes{},
	}

	return node, nil
}

func renderParagraph(src *source.Source, p ast.Paragraph) (html.Node, error) {
	node := html.Element{
		Tag:      "p",
		Attr:     html.Attributes{},
		Children: make([]html.Node, 0, len(p.Inlines)),
	}

	for _, inl := range p.Inlines {
		child, err := renderInline(src, inl)
		if err != nil {
			return nil, err
		}

		node.Children = appendChild(node.Children, child)
	}

	return node, nil
}

func renderInline(src *source.Source, inl ast.Inline) (html.Node, error) {
	switch v := inl.(type) {
	case ast.Text:
		return renderText(src, v)
	case ast.SoftBreak:
		return renderSoftBreak(v)
	case ast.HardBreak:
		return renderHardBreak(v)
	default:
		return nil, fmt.Errorf("unrecognized inline type: %T", inl)
	}
}

func renderText(src *source.Source, t ast.Text) (html.Node, error) {
	node := html.Text{
		Value: src.Slice(t.Span),
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
