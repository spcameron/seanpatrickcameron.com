package block

import (
	"errors"
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

// ErrRuleAdvancedOnDecline reports that a build rule moved the cursor even
// though it declined to match.
var ErrRuleAdvancedOnDecline = errors.New("build rule advanced cursor but declined")

// ErrNoLineConsumed reports that a build rule accepted input without
// advancing the cursor.
var ErrNoLineConsumed = errors.New("build rule accepted but did not advance cursor")

// ErrNoRuleMatched reports that no build rule could be applied at the
// current cursor position.
var ErrNoRuleMatched = errors.New("no build rule could be applied")

// BuildMetadata carries auxiliary state accumulated during block building.
type BuildMetadata struct {
	Definitions map[string]ir.ReferenceDefinition
}

// Build constructs the block-level IR document for src.
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

// buildBlocks applies block rules to lines within the current baseline
// indentation scope and returns the resulting IR blocks.
func buildBlocks(src *source.Source, rules []BuildRule, lines []Line, baselineCols int, state *BuildMetadata) ([]ir.Block, error) {
	c := NewCursor(src, state, rules, lines, baselineCols)
	blocks := []ir.Block{}

	for {
		c.SkipBlankLines()

		if c.EOF() {
			break
		}

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

// defaultRules returns the block build rules in precedence order.
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
