package inline

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"

func Parse(input string) ([]ast.Inline, error) {
	stream, err := Scan(input)
	if err != nil {
		return nil, err
	}

	out, err := Build(stream)
	if err != nil {
		return nil, err
	}

	return out, nil
}
