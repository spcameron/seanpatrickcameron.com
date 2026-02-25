package block

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

func Parse(src *source.Source) (ir.Document, error) {
	lines, err := Scan(src)
	if err != nil {
		return ir.Document{}, err
	}

	out, err := Build(src, lines)
	if err != nil {
		return ir.Document{}, err
	}

	return out, nil
}
