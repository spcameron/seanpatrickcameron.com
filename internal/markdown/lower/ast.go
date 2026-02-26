package lower

import (
	"fmt"
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/inline"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

func Document(irDoc ir.Document) (ast.Document, error) {
	astDoc := ast.Document{
		Source: irDoc.Source,
		Blocks: make([]ast.Block, 0, len(irDoc.Blocks)),
	}

	for _, v := range irDoc.Blocks {
		block, err := buildBlock(astDoc.Source, v)
		if err != nil {
			return ast.Document{}, err
		}

		astDoc.Blocks = append(astDoc.Blocks, block)
	}

	return astDoc, nil
}

func buildBlock(src *source.Source, block ir.Block) (ast.Block, error) {
	switch v := block.(type) {
	case ir.Header:
		return buildHeader(src, v)
	case ir.ThematicBreak:
		return buildThematicBreak(v)
	case ir.Paragraph:
		return buildParagraph(src, v)
	default:
		return nil, fmt.Errorf("unrecognized block type: %T", block)
	}
}

func buildHeader(src *source.Source, h ir.Header) (ast.Block, error) {
	inlines, err := inline.Parse(src, h.ContentSpan)
	if err != nil {
		return nil, err
	}

	block := ast.Header{
		Span:    h.Span,
		Level:   h.Level,
		Inlines: inlines,
	}

	return block, nil
}

func buildThematicBreak(tb ir.ThematicBreak) (ast.Block, error) {
	block := ast.ThematicBreak{
		Span: tb.Span,
	}

	return block, nil
}

func buildParagraph(src *source.Source, p ir.Paragraph) (ast.Block, error) {
	inlines := []ast.Inline{}
	last := len(p.Lines) - 1

	for i, ls := range p.Lines {
		ps := ls
		hardBreak := false

		if i < last {
			s := src.Slice(ls)
			if strings.HasSuffix(s, "  ") {
				hardBreak = true
				ps.End = ps.End - 2
			} else if strings.HasSuffix(s, "\\") {
				hardBreak = true
				ps.End = ps.End - 1
			}
		}

		lineInlines, err := inline.Parse(src, ps)
		if err != nil {
			return nil, err
		}

		inlines = append(inlines, lineInlines...)

		if i < last {
			anchor := source.ByteSpan{
				Start: ls.End,
				End:   ls.End,
			}

			if hardBreak {
				inlines = append(inlines, ast.HardBreak{Span: anchor})
			} else {
				inlines = append(inlines, ast.SoftBreak{Span: anchor})
			}
		}

	}

	block := ast.Paragraph{
		Span:    p.Span,
		Inlines: inlines,
	}

	return block, nil
}
