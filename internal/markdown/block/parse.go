package block

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"

func Parse(input string) (ir.Document, error) {
	lines, err := Scan(input)
	if err != nil {
		return ir.Document{}, err
	}

	out, err := Build(lines)
	if err != nil {
		return ir.Document{}, err
	}

	return out, nil
}
