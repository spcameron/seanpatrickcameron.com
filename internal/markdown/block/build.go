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

type BuildMetadata struct {
	Definitions map[string]ir.ReferenceDefinition
}

func Build(src *source.Source, lines []Line) (ir.Document, error) {
	metadata := &BuildMetadata{
		Definitions: map[string]ir.ReferenceDefinition{},
	}

	blocks, err := buildBlocks(src, defaultRules(), lines, 0, metadata)
	if err != nil {
		return ir.Document{}, err
	}

	irDoc := ir.Document{
		Source:      src,
		Blocks:      blocks,
		Definitions: metadata.Definitions,
	}

	return irDoc, nil
}

func buildBlocks(src *source.Source, rules []BuildRule, lines []Line, baselineCols int, state *BuildMetadata) ([]ir.Block, error) {
	c := NewCursor(src, rules, lines, baselineCols, state)
	blocks := []ir.Block{}

	for {
		c.SkipBlankLines()

		if c.EOF() {
			break
		}

		// baseline scope termination
		line, _ := c.Peek()
		if _, _, ok := c.RelBlockIndent(line); !ok {
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

			if applied != nil {
				blocks = append(blocks, applied)
			}

			break
		}

		if !matched {
			return nil, fmt.Errorf("%w: (index %d)", ErrNoRuleMatched, c.Index)
		}
	}

	return blocks, nil
}

func defaultRules() []BuildRule {
	return []BuildRule{
		BlockQuoteRule{},
		HeaderRule{},
		ThematicBreakRule{},
		OrderedListRule{},
		UnorderedListRule{},
		FencedCodeBlockRule{},
		IndentedCodeBlockRule{},
		HTMLBlockRule{},
		ReferenceDefinitionRule{},
		ParagraphRule{},
	}
}
