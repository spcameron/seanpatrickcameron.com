package inline

import (
	"errors"
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

var (
	ErrRuleAdvancedOnDecline = errors.New("inline rule advanced cursor but declined")
	ErrNoEventConsumed       = errors.New("inline rule accepted but did not advance cursor")
	ErrNoRuleMatched         = errors.New("no inline rule could be applied")
)

func Build(src *source.Source, events []Event) ([]ast.Inline, error) {
	inl := []ast.Inline{}

	rules := defaultRules()

	c := NewCursor(src, rules, events)

	for {
		if c.EOF() {
			break
		}

		matched := false

		for _, rule := range c.Rules {
			applied, ok, err := c.TryApply(rule)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}

			matched = true
			inl = append(inl, applied)
			break
		}

		if !matched {
			return nil, fmt.Errorf("%w: (index %d)", ErrNoRuleMatched, c.Index)
		}
	}

	return inl, nil
}

func defaultRules() []InlineRule {
	return []InlineRule{
		IllegalEventRule{},
		TextRule{},
	}
}
