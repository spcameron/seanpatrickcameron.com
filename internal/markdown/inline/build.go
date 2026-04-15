package inline

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

func Build(src *source.Source, defs map[string]ir.ReferenceDefinition, span source.ByteSpan, tokens []Token) ([]ast.Inline, error) {
	c := NewCursor(src, defs, span, tokens)

	inlines, err := c.Build()
	if err != nil {
		return []ast.Inline{}, err
	}

	return inlines, err
}
