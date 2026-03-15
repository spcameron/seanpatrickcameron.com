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

func (l Line) IsPhysicalLineStart(src *source.Source) bool {
	start := l.Span.Start
	if start == 0 {
		return true
	}
	return src.Raw[start-1] == '\n'
}

// BlockIndent computes the leading indentation of the line.
//
// Indentation is measure in visual columns. A space advances
// one column; a tab advances to the next multiple of tabWidth
// columns. Only leading ' ' and '\t' are considered.
//
// The returned values are:
//   - indentCols: indentation measured in columns
//   - indentBytes: number of leading bytes consumed
//
// Scanning stops at the first non-space, non-tab byte.
func (l Line) BlockIndent(src *source.Source) (indentCols int, indentBytes int) {
	s := src.Slice(l.Span)

	col := 0
	pos := 0

outer:
	for pos < len(s) {
		b := s[pos]
		switch b {
		case ' ':
			col++
			pos++

		case '\t':
			col += source.TabWidth - (col % source.TabWidth)
			pos++

		default:
			break outer
		}
	}

	return col, pos
}

func (l Line) TrimIndentToCols(src *source.Source, baselineCols int) Line {
	if baselineCols <= 0 {
		return l
	}

	s := src.Slice(l.Span)

	col := 0
	pos := 0

	for pos < len(s) && col < baselineCols {
		switch s[pos] {
		case ' ':
			col++
			pos++

		case '\t':
			col += source.TabWidth - (col % source.TabWidth)
			pos++

		default:
			return l
		}
	}

	return Line{
		Span: source.ByteSpan{
			Start: l.Span.Start + source.BytePos(pos),
			End:   l.Span.End,
		},
	}
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
