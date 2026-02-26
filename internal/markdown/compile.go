package markdown

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/block"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/codegen"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/html"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/lower"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

func Tree(md string) (html.Node, error) {
	src := source.NewSource(md)

	irDoc, err := block.Parse(src)
	if err != nil {
		return nil, err
	}

	astDoc, err := lower.Document(irDoc)
	if err != nil {
		return nil, err
	}

	tree, err := codegen.HTML(astDoc)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func HTML(md string) (string, error) {
	tree, err := Tree(md)
	if err != nil {
		return "", err
	}

	htmlStr, err := html.Render(tree)
	if err != nil {
		return "", err
	}

	return htmlStr, nil
}
