package build

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/inline"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
)

func AST(doc ir.Document) (ast.Document, error) {
	astDoc := ast.Document{
		Blocks: make([]ast.Block, 0, len(doc.Blocks)),
	}

	for _, v := range doc.Blocks {
		block, err := buildBlock(v)
		if err != nil {
			return ast.Document{}, err
		}

		astDoc.Blocks = append(astDoc.Blocks, block)
	}

	return astDoc, nil
}

func buildBlock(block ir.Block) (ast.Block, error) {
	switch v := block.(type) {
	case ir.Paragraph:
		return buildParagraph(v)
	case ir.Header:
		return buildHeader(v)
	default:
		return nil, fmt.Errorf("unrecognized block type: %T", block)
	}
}

func buildHeader(h ir.Header) (ast.Block, error) {
	inlines, err := inline.Parse(h.Text)
	if err != nil {
		return nil, err
	}

	block := ast.Header{
		Level:   h.Level,
		Inlines: inlines,
	}

	return block, nil
}

func buildParagraph(p ir.Paragraph) (ast.Block, error) {
	inlines, err := inline.Parse(p.Text)
	if err != nil {
		return nil, err
	}

	block := ast.Paragraph{
		Inlines: inlines,
	}

	return block, nil
}
