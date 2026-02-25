package block

import (
	"errors"
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

var (
	ErrRuleAdvancedOnDecline = errors.New("build rule advanced cursor but declined")
	ErrNoLineConsumed        = errors.New("build rule accepted but did not advance cursor")
	ErrNoRuleMatched         = errors.New("no build rule could be applied")
)

func Build(src *source.Source, lines []String) (ir.Document, error) {
	doc := ir.Document{
		Source: src,
		Blocks: []ir.Block{},
	}

	rules := []BuildRule{
		HeaderRule{},
		ParagraphRule{},
	}

	c := NewCursor(src, rules, lines)

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
