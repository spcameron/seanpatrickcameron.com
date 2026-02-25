package inline

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

func Parse(src *source.Source, span source.ByteSpan) ([]ast.Inline, error) {
	events, err := Scan(src, span)
	if err != nil {
		return nil, err
	}

	out, err := Build(src, events)
	if err != nil {
		return nil, err
	}

	return out, nil
}
