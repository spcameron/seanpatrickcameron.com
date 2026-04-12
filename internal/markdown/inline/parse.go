package inline

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

func Parse(src *source.Source, defs map[string]ir.ReferenceDefinition, span source.ByteSpan) ([]ast.Inline, error) {
	tokens, err := Scan(src, span)
	if err != nil {
		return nil, err
	}

	out, err := Build(src, defs, span, tokens)
	if err != nil {
		return nil, err
	}

	return out, nil
}
