package inline

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"

type BracketRecord struct {
	Span      source.ByteSpan
	ItemIndex int
	Active    bool
}
