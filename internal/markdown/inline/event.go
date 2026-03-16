package inline

import (
	"fmt"
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type EventKind int

const (
	_ EventKind = iota
	EventText
	EventDelimiterRun
	EventOpenBracket
	EventCloseBracket
	EventOpenParen
	EventCloseParen
	EventIllegalNewline
)

func (ek EventKind) String() string {
	switch ek {
	case EventText:
		return "text"

	case EventDelimiterRun:
		return "delimiter_run"

	case EventOpenBracket:
		return "open_bracket"

	case EventCloseBracket:
		return "close_bracket"

	case EventOpenParen:
		return "open_paren"

	case EventCloseParen:
		return "close_paren"

	case EventIllegalNewline:
		return "illegal_newline"

	default:
		return fmt.Sprintf("unknown_event_kind(%d)", ek)
	}
}

type Event struct {
	Kind      EventKind
	Span      source.ByteSpan
	Delimiter byte
	RunLength int
}

func (e Event) String() string {
	switch e.Kind {
	case EventText:
		return fmt.Sprintf("text(%s)", e.Span)

	case EventDelimiterRun:
		return fmt.Sprintf(
			"delimiter_run(%q, %s)",
			strings.Repeat(string(e.Delimiter), e.RunLength),
			e.Span,
		)

	case EventOpenBracket:
		return fmt.Sprintf("open_bracket(%s)", e.Span)

	case EventCloseBracket:
		return fmt.Sprintf("close_bracket(%s)", e.Span)

	case EventOpenParen:
		return fmt.Sprintf("open_paren(%s)", e.Span)

	case EventCloseParen:
		return fmt.Sprintf("close_paren(%s)", e.Span)

	case EventIllegalNewline:
		return fmt.Sprintf("illegal_newline(%s)", e.Span)

	default:
		return fmt.Sprintf("unknown_event(%d, %s)", e.Kind, e.Span)
	}
}
