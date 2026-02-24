package block

import (
	"errors"
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
)

var (
	ErrRuleAdvancedOnDecline = errors.New("build rule advanced cursor but declined")
	ErrNoLineConsumed        = errors.New("build rule accepted but did not advance cursor")
	ErrNoRuleMatched         = errors.New("no build rule could be applied")
)

func Build(lines []Line) (ir.Document, error) {
	doc := ir.Document{}

	rules := []BuildRule{
		HeaderRule{},
		ParagraphRule{},
	}

	c := NewCursor(rules, lines)

	for {
		c.SkipBlankLines()

		if c.EOF() {
			break
		}

		matched := false

		for _, rule := range c.Rules {
			applied, ok, err := c.TryApply(rule)
			if err != nil {
				return ir.Document{}, err
			}
			if !ok {
				continue
			}

			matched = true
			doc.Blocks = append(doc.Blocks, applied)
			break
		}

		if !matched {
			return ir.Document{}, fmt.Errorf("%w: (index %d)", ErrNoRuleMatched, c.Index)
		}

	}

	return doc, nil
}
