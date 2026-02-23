package inline

import (
	"errors"
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
)

var (
	ErrRuleAdvancedOnDecline = errors.New("inline rule advanced cursor but declined")
	ErrNoEventConsumed       = errors.New("inline rule accepted but did not advance cursor")
	ErrNoRuleMatched         = errors.New("no inline rule could be applied")
)

func Build(events []Event) ([]ast.Inline, error) {
	out := []ast.Inline{}
	c := NewCursor(events)

	rules := []InlineRule{
		TextRule{},
	}

	for {
		if c.EOF() {
			break
		}

		matched := false

		for _, rule := range rules {
			beforeRule := c.Index

			applied, ok, err := rule.Apply(c)
			if err != nil {
				return nil, err
			}
			if !ok {
				if c.Index != beforeRule {
					return nil, fmt.Errorf("%w: (index %d)", ErrRuleAdvancedOnDecline, beforeRule)
				}
				continue
			}

			if c.Index == beforeRule {
				return nil, fmt.Errorf("%w (index %d)", ErrNoEventConsumed, c.Index)
			}

			matched = true
			out = append(out, applied)
			break
		}

		if !matched {
			return nil, fmt.Errorf("%w: (index %d)", ErrNoRuleMatched, c.Index)
		}
	}

	return out, nil
}
