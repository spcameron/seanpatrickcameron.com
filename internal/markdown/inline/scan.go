package inline

import (
	"fmt"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type EventKind int

const (
	_ EventKind = iota
	EventText
	EventDelimiterRun
	EventIllegalNewline
)

func (ek EventKind) String() string {
	switch ek {
	case EventText:
		return "Event - Text"
	case EventDelimiterRun:
		return "Event - Delimiter Run"
	case EventIllegalNewline:
		return "Event - Illegal Newline"
	default:
		return fmt.Sprintf("Unrecognized EventKind %d", ek)
	}
}

type Event struct {
	Kind      EventKind
	Span      source.ByteSpan
	Delimiter byte
	RunLength int
}

func (e Event) String() string {
	if e.Kind == EventDelimiterRun {
		return fmt.Sprintf("[%s] - Delimiter (%s), Length (%d)", e.Kind, string(e.Delimiter), e.RunLength)
	}

	return fmt.Sprintf("[%s]", e.Kind)
}

func Scan(src *source.Source, span source.ByteSpan) ([]Event, error) {
	input := src.Slice(span)
	scanner := NewScanner(input, span.Start)

	events := []Event{}
	for {
		event, ok := scanner.Next()
		if !ok {
			break
		}

		events = append(events, event)
	}

	return events, nil
}

type Scanner struct {
	Input    string
	Position int
	Base     source.BytePos
}

func NewScanner(input string, base source.BytePos) *Scanner {
	return &Scanner{
		Input:    input,
		Position: 0,
		Base:     base,
	}
}

func (s *Scanner) EOF() bool {
	return s.Position >= len(s.Input)
}

func (s *Scanner) Next() (Event, bool) {
	if s.EOF() {
		return Event{}, false
	}

	start := s.Position
	b := s.Input[s.Position]

	// special token dispatch
	switch b {
	case '*':
		for s.Position < len(s.Input) && s.Input[s.Position] == '*' {
			s.Position++
		}

		end := s.Position
		length := end - start

		span := source.ByteSpan{
			Start: s.Base + source.BytePos(start),
			End:   s.Base + source.BytePos(end),
		}

		return Event{
			Kind:      EventDelimiterRun,
			Span:      span,
			Delimiter: '*',
			RunLength: length,
		}, true
	case '\n':
		s.Position++
		end := s.Position

		span := source.ByteSpan{
			Start: s.Base + source.BytePos(start),
			End:   s.Base + source.BytePos(end),
		}

		return Event{
			Kind: EventIllegalNewline,
			Span: span,
		}, true
	}

	// otherwise, scan maximum text run until next special token
	end := s.Position
	for end < len(s.Input) {
		if terminatesText(s.Input[end]) {
			break
		}

		end++
	}

	s.Position = end

	if end == start {
		panic("inline scanner made no progress")
	}

	span := source.ByteSpan{
		Start: s.Base + source.BytePos(start),
		End:   s.Base + source.BytePos(end),
	}

	return Event{
		Kind: EventText,
		Span: span,
	}, true
}

func terminatesText(b byte) bool {
	switch b {
	case '\n', '*':
		return true
	default:
		return false
	}
}
