package inline

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
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

	// token dispatch
	switch b {
	case '[':
		return s.emitSingle(EventOpenBracket)

	case ']':
		return s.emitSingle(EventCloseBracket)

	case '(':
		return s.emitSingle(EventOpenParen)

	case ')':
		return s.emitSingle(EventCloseParen)

	case '\n':
		return s.emitSingle(EventIllegalNewline)

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

	default:
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

		span := s.eventSpan(start, end)

		return Event{
			Kind: EventText,
			Span: span,
		}, true
	}
}

func (s *Scanner) eventSpan(start, end int) source.ByteSpan {
	return source.ByteSpan{
		Start: s.Base + source.BytePos(start),
		End:   s.Base + source.BytePos(end),
	}
}

func (s *Scanner) emitSingle(kind EventKind) (Event, bool) {
	start := s.Position
	s.Position++

	return Event{
		Kind: kind,
		Span: s.eventSpan(start, s.Position),
	}, true
}

func terminatesText(b byte) bool {
	switch b {
	case '\n', '*', '[', ']', '(', ')':
		return true

	default:
		return false
	}
}
