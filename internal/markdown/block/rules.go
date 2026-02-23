package block

import (
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
)

type BuildRule interface {
	Apply(c *Cursor) (ir.Block, bool, error)
}

type ParagraphRule struct{}

func (r ParagraphRule) Apply(c *Cursor) (ir.Block, bool, error) {
	line, ok := c.Peek()
	if !ok {
		return nil, false, nil
	}
	if line.IsBlankLine() {
		return nil, false, nil
	}

	var text []string
	start := c.Index

	for {
		line, ok := c.Peek()
		if !ok {
			break
		}
		if line.IsBlankLine() {
			break
		}

		line, _ = c.Next()
		text = append(text, line.Text)
	}

	end := c.Index

	span := ir.LineSpan{
		Start: start,
		End:   end,
	}

	joinedText := strings.Join(text, "\n")

	applied := ir.Paragraph{
		Text: joinedText,
		Span: span,
	}

	return applied, true, nil
}
