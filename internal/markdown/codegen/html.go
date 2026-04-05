package codegen

import (
	"fmt"
	"strconv"

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

	case ast.OrderedList:
		return renderOrderedList(src, v)

	case ast.UnorderedList:
		return renderUnorderedList(src, v)

	case ast.CodeBlock:
		return renderCodeBlock(src, v)

	case ast.HTMLBlock:
		return renderHTMLBlock(src, v)

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

func appendChildren(dst []html.Node, src []html.Node) []html.Node {
	for _, n := range src {
		dst = appendChild(dst, n)
	}

	return dst
}

func renderBlockQuote(src *source.Source, block ast.BlockQuote) (html.Node, error) {
	node := html.Element{
		Tag:      "blockquote",
		Attr:     html.Attributes{},
		Children: make([]html.Node, 0, len(block.Children)),
	}

	for _, bqChild := range block.Children {
		htmlChild, err := renderBlock(src, bqChild)
		if err != nil {
			return nil, err
		}

		node.Children = appendChild(node.Children, htmlChild)
	}

	return node, nil
}

func renderHeader(src *source.Source, block ast.Header) (html.Node, error) {
	children, err := renderInlines(src, block.Inlines)
	if err != nil {
		return nil, err
	}

	node := html.Element{
		Tag:      fmt.Sprintf("h%d", block.Level),
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

func renderOrderedList(src *source.Source, block ast.OrderedList) (html.Node, error) {
	attr := html.Attributes{}
	if block.Start != 1 {
		attr["start"] = strconv.Itoa(block.Start)
	}

	node := html.Element{
		Tag:      "ol",
		Attr:     attr,
		Children: make([]html.Node, 0, len(block.Items)),
	}

	for _, olItem := range block.Items {
		liNode, err := renderListItem(src, olItem, block.Tight)
		if err != nil {
			return nil, err
		}

		node.Children = appendChild(node.Children, liNode)
	}

	return node, nil
}

func renderUnorderedList(src *source.Source, block ast.UnorderedList) (html.Node, error) {
	node := html.Element{
		Tag:      "ul",
		Attr:     html.Attributes{},
		Children: make([]html.Node, 0, len(block.Items)),
	}

	for _, ulItem := range block.Items {
		liNode, err := renderListItem(src, ulItem, block.Tight)
		if err != nil {
			return nil, err
		}

		node.Children = appendChild(node.Children, liNode)
	}

	return node, nil
}

// tight-list rendering: unwraps a single paragraph child
func renderListItem(src *source.Source, block ast.ListItem, tight bool) (html.Node, error) {
	node := html.Element{
		Tag:  "li",
		Attr: html.Attributes{},
	}

	if tight {
		for _, liChild := range block.Children {
			if p, ok := liChild.(ast.Paragraph); ok {
				inlines, err := renderInlines(src, p.Inlines)
				if err != nil {
					return nil, err
				}
				node.Children = appendChildren(node.Children, inlines)
			} else {
				htmlChild, err := renderBlock(src, liChild)
				if err != nil {
					return nil, err
				}
				node.Children = appendChild(node.Children, htmlChild)
			}
		}

		return node, nil
	}

	for _, liChild := range block.Children {
		htmlChild, err := renderBlock(src, liChild)
		if err != nil {
			return nil, err
		}

		node.Children = appendChild(node.Children, htmlChild)
	}

	return node, nil
}

func renderCodeBlock(src *source.Source, block ast.CodeBlock) (html.Node, error) {
	attr := html.Attributes{}

	languageString := src.Slice(block.LanguageTokenSpan)
	if languageString != "" {
		attr["class"] = fmt.Sprintf("language-%s", languageString)
	}
	payload, err := renderInlines(src, block.Payload)
	if err != nil {
		return nil, err
	}

	node := html.Element{
		Tag:  "pre",
		Attr: html.Attributes{},
		Children: []html.Node{
			html.Element{
				Tag:      "code",
				Attr:     attr,
				Children: payload,
			},
		},
	}

	return node, nil
}

func renderHTMLBlock(src *source.Source, block ast.HTMLBlock) (html.Node, error) {
	children, err := renderInlines(src, block.Payload)
	if err != nil {
		return nil, err
	}

	node := html.Fragment{
		Children: children,
	}

	return node, nil
}

func renderParagraph(src *source.Source, block ast.Paragraph) (html.Node, error) {
	children, err := renderInlines(src, block.Inlines)
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
	case ast.CodeSpan:
		return renderCodeSpan(src, v)

	case ast.Link:
		return renderLink(src, v)

	case ast.Emph:
		return renderEmphasis(src, v)

	case ast.Strong:
		return renderStrong(src, v)

	case ast.Text:
		return renderText(src, v)

	case ast.RawText:
		return renderRawText(src, v)

	case ast.SoftBreak:
		return renderSoftBreak()

	case ast.HardBreak:
		return renderHardBreak()

	case ast.Newline:
		return renderNewline()

	default:
		return nil, fmt.Errorf("unrecognized inline type: %T", inl)
	}
}

func renderCodeSpan(src *source.Source, inl ast.CodeSpan) (html.Node, error) {
	contentNode := html.Text{
		Value: src.Slice(inl.Span),
	}

	node := html.Element{
		Tag:      "code",
		Attr:     html.Attributes{},
		Children: []html.Node{contentNode},
	}

	return node, nil
}

func renderLink(src *source.Source, inl ast.Link) (html.Node, error) {
	inlines, err := renderInlines(src, inl.Children)
	if err != nil {
		return nil, err
	}

	href := src.Slice(inl.Destination)
	if inl.MailTo {
		href = "mailto:" + href
	}

	attr := html.Attributes{
		"href": href,
	}

	if inl.Title != (source.ByteSpan{}) {
		attr["title"] = src.Slice(inl.Title)
	}

	node := html.Element{
		Tag:      "a",
		Attr:     attr,
		Children: inlines,
	}

	return node, nil
}

func renderEmphasis(src *source.Source, inl ast.Emph) (html.Node, error) {
	inlines, err := renderInlines(src, inl.Children)
	if err != nil {
		return nil, err
	}

	node := html.Element{
		Tag:      "em",
		Attr:     html.Attributes{},
		Children: inlines,
	}

	return node, nil
}

func renderStrong(src *source.Source, inl ast.Strong) (html.Node, error) {
	inlines, err := renderInlines(src, inl.Children)
	if err != nil {
		return nil, err
	}

	node := html.Element{
		Tag:      "strong",
		Attr:     html.Attributes{},
		Children: inlines,
	}

	return node, nil
}

func renderText(src *source.Source, inl ast.Text) (html.Node, error) {
	node := html.Text{
		Value: src.Slice(inl.Span),
	}

	return node, nil
}

func renderRawText(src *source.Source, inl ast.RawText) (html.Node, error) {
	node := html.Raw{
		Value: src.Slice(inl.Span),
	}

	return node, nil
}

func renderSoftBreak() (html.Node, error) {
	node := html.Text{
		Value: " ",
	}

	return node, nil
}

func renderHardBreak() (html.Node, error) {
	node := html.VoidElement{
		Tag:  "br",
		Attr: html.Attributes{},
	}

	return node, nil
}

func renderNewline() (html.Node, error) {
	node := html.Text{
		Value: "\n",
	}

	return node, nil
}
