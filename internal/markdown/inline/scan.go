package inline

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type EventKind int

const (
	_ EventKind = iota
	EventText
	EventIllegalNewline
)

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

type Event struct {
	Kind EventKind
	Span source.ByteSpan
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
		if s.Input[end] == '\n' {
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
