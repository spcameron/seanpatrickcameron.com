package markdown

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/block"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/build"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/html"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/render"
)

func CompileAndRender(md string) (string, error) {
	htmlTree, err := Compile(md)
	if err != nil {
		return "", err
	}

	htmlOutput, err := html.Render(htmlTree)
	if err != nil {
		return "", err
	}

	return htmlOutput, nil
}

func Compile(md string) (html.Node, error) {
	irDoc, err := block.Parse(md)
	if err != nil {
		return nil, err
	}

	astDoc, err := build.AST(irDoc)
	if err != nil {
		return nil, err
	}

	htmlTree, err := render.HTML(astDoc)
	if err != nil {
		return nil, err
	}

	return htmlTree, nil
}
