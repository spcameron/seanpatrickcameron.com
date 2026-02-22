package block

import "strings"

func Scan(input string) ([]Line, error) {
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
	Text string
}

func (l Line) IsBlankLine() bool {
	return strings.TrimSpace(l.Text) == ""
}

type Scanner struct {
	Input    string
	Position int
}

func NewScanner(input string) *Scanner {
	return &Scanner{
		Input:    input,
		Position: 0,
	}
}

func (s *Scanner) EOF() bool {
	return s.Position >= len(s.Input)
}

func (s *Scanner) Next() (Line, bool) {
	if s.EOF() {
		return Line{}, false
	}

	start := s.Position
	for s.Position < len(s.Input) {
		b := s.Input[s.Position]
		if b == '\n' {
			text := normalizeLineText(s.Input[start:s.Position])
			s.Position++

			if text == "" && s.EOF() && start == 0 {
				return Line{}, false
			}

			return Line{
				Text: text,
			}, true
		}
		s.Position++
	}

	text := normalizeLineText(s.Input[start:s.Position])
	return Line{
		Text: text,
	}, true
}

func normalizeLineText(input string) string {
	return strings.TrimRight(input, " \t\r")
}
