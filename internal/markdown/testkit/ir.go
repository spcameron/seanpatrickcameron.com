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

func IRHeader(level int, input ...string) ir.Header {
	return IRHeaderAt(0, level, input...)
}

func IRHeaderAt(start, level int, input ...string) ir.Header {
	text := strings.Join(input, "\n")
	span := ir.LineSpan{
		Start: start,
		End:   start + len(input),
	}

	return ir.Header{
		Level: level,
		Text:  text,
		Span:  span,
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
