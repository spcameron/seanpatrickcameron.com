package inline

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
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

	// TODO: span preserved for inclusion in AST later
	_ = ir.ByteSpan{
		Start: event.Position,
		End:   event.Position + len(event.Lexeme),
	}

	applied := ast.Text{
		Value: event.Lexeme,
	}

	return applied, true, nil
}
