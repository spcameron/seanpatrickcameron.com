package block

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

// Cursor tracks block-parser progress through a slice of lines within a
// baseline indentation scope.
type Cursor struct {
	Source       *source.Source
	Metadata     *BuildMetadata
	Rules        []BuildRule
	Lines        []Line
	Index        int
	BaselineCols int
}

// NewCursor constructs a cursor over lines using the provided rule set and
// baseline indentation.
//
// BaselineCols defines the indentation scope for RelBlockIndent and must
// correspond to the coordinate system of the provided lines.
func NewCursor(src *source.Source, state *BuildMetadata, rules []BuildRule, lines []Line, baselineCols int) *Cursor {
	return &Cursor{
		Source:       src,
		Metadata:     state,
		Rules:        rules,
		Lines:        lines,
		Index:        0,
		BaselineCols: baselineCols,
	}
}

// Peek returns the current line without advancing.
func (c *Cursor) Peek() (Line, bool) {
	if c.EOF() {
		return Line{}, false
	}

	return c.Lines[c.Index], true
}

// Next returns the current line and advances the cursor.
func (c *Cursor) Next() (Line, bool) {
	if c.EOF() {
		return Line{}, false
	}

	out := c.Lines[c.Index]
	c.Index++
	return out, true
}

// MustNext returns the current line and advances the cursor, panicking if
// the cursor is at EOF.
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

// AbsBlockIndent reports the line's absolute block indentation in columns
// and bytes.
func (c *Cursor) AbsBlockIndent(line Line) (absCols int, indentBytes int) {
	return line.BlockIndent(c.Source)
}

// RelBlockIndent reports the line's indentation relative to the cursor's
// baseline. ok is false when the line falls outside the current scope.
func (c *Cursor) RelBlockIndent(line Line) (relCols int, indentBytes int, ok bool) {
	absCols, indentBytes := line.BlockIndent(c.Source)
	relCols = absCols - c.BaselineCols
	if relCols < 0 {
		return 0, indentBytes, false
	}

	return relCols, indentBytes, true
}

// SkipBlankLines advances past consecutive blank lines.
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

// TryApply applies rule at the current cursor position.
//
// It enforces the build-rule contract: declines rules must not advance the
// cursor, and accepting rules must consume at least one line.
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

// StartsParagraphInterruptingBlock reports whether the current position
// begins a block that would interrupt an in-progress paragraph.
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
