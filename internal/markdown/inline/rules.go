package inline

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
)

type InlineRule interface {
	Apply(c *Cursor) (ast.Inline, bool, error)
}

type TextRule struct{}

func (r TextRule) Apply(c *Cursor) (ast.Inline, bool, error) {
	event, ok := c.Peek()
	if !ok {
		return nil, false, nil
	}
	if event.Kind != EventText {
		return nil, false, nil
	}

	event, _ = c.Next()

	applied := ast.Text{
		Span: event.Span,
	}

	return applied, true, nil
}

type IllegalEventRule struct{}

func (r IllegalEventRule) Apply(c *Cursor) (ast.Inline, bool, error) {
	event, ok := c.Peek()
	if !ok {
		return nil, false, nil
	}

	if event.Kind == EventIllegalNewline {
		panic("illegal newline encountered during inline scanner")
	}

	return nil, false, nil
}
