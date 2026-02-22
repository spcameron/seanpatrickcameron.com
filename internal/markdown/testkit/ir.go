package testkit

import (
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
)

func IRDoc(blocks ...ir.Block) ir.Document {
	return ir.Document{
		Blocks: blocks,
	}
}

func IRPara(input ...string) ir.Paragraph {
	return IRParaAt(0, input...)
}

func IRParaAt(start int, input ...string) ir.Paragraph {
	text := strings.Join(input, "\n")
	span := ir.LineSpan{
		Start: start,
		End:   start + len(input),
	}

	return ir.Paragraph{
		Text: text,
		Span: span,
	}
}
