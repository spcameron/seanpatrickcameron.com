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
	case ast.BlockQuote:
		return renderBlockQuote(src, v)
	case ast.Header:
		return renderHeader(src, v)
	case ast.ThematicBreak:
		return renderThematicBreak()
	case ast.UnorderedList:
		return renderUnorderedList(src, v)
	case ast.ListItem:
		return renderListItem(src, v)
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

func renderBlockQuote(src *source.Source, bq ast.BlockQuote) (html.Node, error) {
	node := html.Element{
		Tag:      "blockquote",
		Attr:     html.Attributes{},
		Children: make([]html.Node, 0, len(bq.Children)),
	}

	for _, bqChild := range bq.Children {
		htmlChild, err := renderBlock(src, bqChild)
		if err != nil {
			return nil, err
		}

		node.Children = appendChild(node.Children, htmlChild)
	}

	return node, nil
}

func renderHeader(src *source.Source, h ast.Header) (html.Node, error) {
	children, err := renderInlines(src, h.Inlines)
	if err != nil {
		return nil, err
	}

	node := html.Element{
		Tag:      fmt.Sprintf("h%d", h.Level),
		Attr:     html.Attributes{},
		Children: children,
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

func renderUnorderedList(src *source.Source, ul ast.UnorderedList) (html.Node, error) {
	node := html.Element{
		Tag:      "ul",
		Attr:     html.Attributes{},
		Children: make([]html.Node, 0, len(ul.Items)),
	}

	for _, ulItem := range ul.Items {
		htmlChild, err := renderBlock(src, ulItem)
		if err != nil {
			return nil, err
		}

		node.Children = appendChild(node.Children, htmlChild)
	}

	return node, nil
}

// tight-list rendering: unwraps a single paragraph child
func renderListItem(src *source.Source, li ast.ListItem) (html.Node, error) {
	if len(li.Children) == 1 {
		if p, ok := li.Children[0].(ast.Paragraph); ok {
			children, err := renderInlines(src, p.Inlines)
			if err != nil {
				return nil, err
			}

			node := html.Element{
				Tag:      "li",
				Attr:     html.Attributes{},
				Children: children,
			}

			return node, nil
		}
	}

	node := html.Element{
		Tag:      "li",
		Attr:     html.Attributes{},
		Children: make([]html.Node, 0, len(li.Children)),
	}

	for _, liChild := range li.Children {
		htmlChild, err := renderBlock(src, liChild)
		if err != nil {
			return nil, err
		}

		node.Children = appendChild(node.Children, htmlChild)
	}

	return node, nil
}

func renderParagraph(src *source.Source, p ast.Paragraph) (html.Node, error) {
	children, err := renderInlines(src, p.Inlines)
	if err != nil {
		return nil, err
	}

	node := html.Element{
		Tag:      "p",
		Attr:     html.Attributes{},
		Children: children,
	}

	return node, nil
}

func renderInlines(src *source.Source, inlines []ast.Inline) ([]html.Node, error) {
	children := make([]html.Node, 0, len(inlines))

	for _, inl := range inlines {
		child, err := renderInline(src, inl)
		if err != nil {
			return nil, err
		}

		children = appendChild(children, child)
	}

	return children, nil
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
