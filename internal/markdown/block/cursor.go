package block

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type Cursor struct {
	Source *source.Source
	Rules  []BuildRule
	Lines  []String
	Index  int
}

func NewCursor(src *source.Source, rules []BuildRule, lines []String) *Cursor {
	return &Cursor{
		Source: src,
		Rules:  rules,
		Lines:  lines,
		Index:  0,
	}
}

func (c *Cursor) Peek() (String, bool) {
	if c.EOF() {
		return String{}, false
	}

	return c.Lines[c.Index], true
}

func (c *Cursor) Next() (String, bool) {
	if c.EOF() {
		return String{}, false
	}

	out := c.Lines[c.Index]
	c.Index++
	return out, true
}

func (c *Cursor) Mark() int {
	return c.Index
}

func (c *Cursor) Reset(i int) {
	c.Index = i
}

func (c *Cursor) EOF() bool {
	return c.Index >= len(c.Lines)
}

func (c *Cursor) SkipBlankLines() {
	for {
		line, ok := c.Peek()
		if !ok {
			return
		}
		if !line.IsBlankLine(c.Source) {
			return
		}

		c.Next()
	}
}

func (c *Cursor) TryApply(rule BuildRule) (ir.Block, bool, error) {
	m := c.Mark()

	applied, ok, err := rule.Apply(c)
	if err != nil {
		c.Reset(m)
		return nil, false, err
	}
	if !ok {
		if c.Index != m {
			c.Reset(m)
			return nil, false, ruleError(rule, m, ErrRuleAdvancedOnDecline)
		}
		return nil, false, nil
	}

	if c.Index == m {
		c.Reset(m)
		return nil, false, ruleError(rule, m, ErrNoLineConsumed)
	}

	return applied, true, nil
}

func (c *Cursor) StartsNonParagraphBlock() (bool, error) {
	m := c.Mark()
	for _, rule := range c.Rules {
		if _, ok := rule.(ParagraphRuleMarker); ok {
			continue
		}

		_, ok, err := c.TryApply(rule)
		c.Reset(m)

		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		}
	}

	return false, nil
}

func ruleError(rule BuildRule, index int, err error) error {
	return fmt.Errorf("rule %T at index %d: %w", rule, index, err)
}
