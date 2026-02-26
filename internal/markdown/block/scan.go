package block

import (
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

func Scan(src *source.Source) ([]Line, error) {
	input := src.Raw
	scanner := NewScanner(input)

	lines := []Line{}
	for {
		line, ok := scanner.Next()
		if !ok {
			break
		}

		lines = append(lines, line)
	}

	return lines, nil
}

type Line struct {
	Span source.ByteSpan
}

func (l Line) IsBlankLine(src *source.Source) bool {
	s := src.Slice(l.Span)
	return strings.TrimSpace(s) == ""
}

func (l Line) BlockIndentSpaces(src *source.Source) int {
	s := src.Slice(l.Span)

	indent := 0
	for indent < len(s) && s[indent] == ' ' {
		indent++
	}

	return indent
}

type Scanner struct {
	Input             string
	Position          source.BytePos
	pendingFinalEmpty bool
}

func NewScanner(input string) *Scanner {
	return &Scanner{
		Input:    input,
		Position: 0,
	}
}

func (s *Scanner) EOF() bool {
	return int(s.Position) >= len(s.Input)
}

func (s *Scanner) Next() (Line, bool) {
	if s.pendingFinalEmpty {
		s.pendingFinalEmpty = false

		eof := source.BytePos(len(s.Input))
		span := source.ByteSpan{
			Start: eof,
			End:   eof,
		}

		return Line{
			Span: span,
		}, true
	}

	if s.EOF() {
		return Line{}, false
	}

	start := s.Position
	for int(s.Position) < len(s.Input) {
		i := int(s.Position)
		if s.Input[i] == '\n' {
			span := source.ByteSpan{
				Start: start,
				End:   s.Position,
			}
			s.Position++

			if s.EOF() {
				s.pendingFinalEmpty = true
			}

			return Line{
				Span: span,
			}, true
		}
		s.Position++
	}

	span := source.ByteSpan{
		Start: start,
		End:   s.Position,
	}

	return Line{
		Span: span,
	}, true
}
