package inline

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
)

type Cursor struct {
	Rules  []InlineRule
	Events []Event
	Index  int
}

func NewCursor(rules []InlineRule, events []Event) *Cursor {
	return &Cursor{
		Rules:  rules,
		Events: events,
		Index:  0,
	}
}

func (c *Cursor) Peek() (Event, bool) {
	if c.EOF() {
		return Event{}, false
	}

	return c.Events[c.Index], true
}

func (c *Cursor) Next() (Event, bool) {
	if c.EOF() {
		return Event{}, false
	}

	out := c.Events[c.Index]
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
	return c.Index >= len(c.Events)
}

func (c *Cursor) TryApply(rule InlineRule) (ast.Inline, bool, error) {
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
		return nil, false, ruleError(rule, m, ErrNoEventConsumed)
	}

	return applied, true, nil
}

func ruleError(rule InlineRule, index int, err error) error {
	return fmt.Errorf("rule %T at index %d: %w", rule, index, err)
}
