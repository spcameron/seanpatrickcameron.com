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
	c := NewCursor(lines)

	rules := []BuildRule{
		HeaderRule{},
		ParagraphRule{},
	}

	for {
		c.SkipBlankLines()

		if c.EOF() {
			break
		}

		matched := false

		for _, rule := range rules {
			beforeRule := c.Index

			applied, ok, err := rule.Apply(c)
			if err != nil {
				return ir.Document{}, err
			}
			if !ok {
				if c.Index != beforeRule {
					return ir.Document{}, fmt.Errorf("%w: (index %d)", ErrRuleAdvancedOnDecline, beforeRule)
				}
				continue
			}

			if c.Index == beforeRule {
				return ir.Document{}, fmt.Errorf("%w (index %d)", ErrNoLineConsumed, c.Index)
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
