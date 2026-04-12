package block

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type Cursor struct {
	Source       *source.Source
	Rules        []BuildRule
	Lines        []Line
	Metadata     *BuildMetadata
	Index        int
	BaselineCols int
}

func NewCursor(src *source.Source, rules []BuildRule, lines []Line, baselineCols int, state *BuildMetadata) *Cursor {
	return &Cursor{
		Source:       src,
		Rules:        rules,
		Lines:        lines,
		Metadata:     state,
		Index:        0,
		BaselineCols: 0,
	}
}

func (c *Cursor) Peek() (Line, bool) {
	if c.EOF() {
		return Line{}, false
	}

	return c.Lines[c.Index], true
}

func (c *Cursor) Next() (Line, bool) {
	if c.EOF() {
		return Line{}, false
	}

	out := c.Lines[c.Index]
	c.Index++
	return out, true
}

func (c *Cursor) MustNext() Line {
	line, ok := c.Next()
	if !ok {
		panic("block cursor invariant violated: Next() returned false after Peek() succeeded")
	}
	return line
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

func (c *Cursor) AbsBlockIndent(line Line) (absCols int, indentBytes int) {
	return line.BlockIndent(c.Source)
}

func (c *Cursor) RelBlockIndent(line Line) (relCols int, indentBytes int, ok bool) {
	absCols, indentBytes := line.BlockIndent(c.Source)
	relCols = absCols - c.BaselineCols
	if relCols < 0 {
		return 0, indentBytes, false
	}

	return relCols, indentBytes, true
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

func (c *Cursor) StartsParagraphInterruptingBlock() (bool, error) {
	m := c.Mark()
	for _, rule := range c.Rules {
		if _, ok := rule.(ParagraphTransparentRuleMarker); ok {
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
