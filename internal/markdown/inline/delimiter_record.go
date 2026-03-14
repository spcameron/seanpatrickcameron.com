package inline

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"

type DelimiterRecord struct {
	Span         source.ByteSpan
	Delimiter    byte
	OriginalRun  int
	RemainingRun int
	CanOpen      bool
	CanClose     bool
	ItemIndex    int
}
