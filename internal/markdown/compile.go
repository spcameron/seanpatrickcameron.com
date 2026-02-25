package markdown

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/block"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/codegen"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/html"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/lower"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

func Compile(md string) (html.Node, error) {
	src := source.NewSource(md)

	irDoc, err := block.Parse(src)
	if err != nil {
		return nil, err
	}

	astDoc, err := lower.Document(irDoc)
	if err != nil {
		return nil, err
	}

	htmlTree, err := codegen.HTML(astDoc)
	if err != nil {
		return nil, err
	}

	return htmlTree, nil
}

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
