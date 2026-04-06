package inline

import (
	"errors"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

var (
	ErrRuleAdvancedOnDecline = errors.New("inline rule advanced cursor but declined")
	ErrNoEventConsumed       = errors.New("inline rule accepted but did not advance cursor")
	ErrNoRuleMatched         = errors.New("no inline rule could be applied")
)

func Build(src *source.Source, span source.ByteSpan, tokens []Token) ([]ast.Inline, error) {
	c := NewCursor(src, span, tokens)

	inlines, err := c.Build()
	if err != nil {
		return []ast.Inline{}, err
	}

	return inlines, err
}
