package inline

import (
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type EventKind int

const (
	_ EventKind = iota
	EventText
	EventSoftBreak
	EventHardBreak
)

func Scan(input string) ([]Event, error) {
	scanner := NewScanner(input)

	stream := []Event{}
	for {
		event, ok := scanner.Next()
		if !ok {
			break
		}

		stream = append(stream, event)
	}

	return stream, nil
}

type Event struct {
	Kind     EventKind
	Lexeme   string
	Position source.BytePos
}

type Scanner struct {
	Input            string
	Position         int
	PendingHardBreak bool
}

func NewScanner(input string) *Scanner {
	return &Scanner{
		Input:            input,
		Position:         0,
		PendingHardBreak: false,
	}
}

func (s *Scanner) EOF() bool {
	return s.Position >= len(s.Input)
}

func (s *Scanner) Next() (Event, bool) {
	for {
		if s.EOF() {
			return Event{}, false
		}

		start := s.Position
		b := s.Input[s.Position]

		// special token dispatch
		switch b {
		case '\n':
			kind := EventSoftBreak
			if s.PendingHardBreak {
				kind = EventHardBreak
				s.PendingHardBreak = false
			}

			s.Position++

			return Event{
				Kind:     kind,
				Lexeme:   "",
				Position: start,
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

		lex := s.Input[start:end]

		if end < len(s.Input) && s.Input[end] == '\n' {
			if trimmed, ok := strings.CutSuffix(lex, "  "); ok {
				lex = trimmed
				s.PendingHardBreak = true
			} else if trimmed, ok := strings.CutSuffix(lex, `\`); ok {
				lex = trimmed
				s.PendingHardBreak = true
			}
		}

		s.Position = end

		if lex == "" {
			continue
		}

		return Event{
			Kind:     EventText,
			Lexeme:   lex,
			Position: start,
		}, true
	}
}
