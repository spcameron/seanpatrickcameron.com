package block

import (
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

// Scan splits source into logical lines for block parsing.
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

// Line represents a scanned source line as a span into the normalized input.
type Line struct {
	Span source.ByteSpan
}

// IsBlankLine reports whether the line contains only whitespace.
func (l Line) IsBlankLine(src *source.Source) bool {
	s := src.Slice(l.Span)
	return strings.TrimSpace(s) == ""
}

// IsPhysicalLineStart reports whether the line begins at the start of the
// source or immediately after a newline byte.
func (l Line) IsPhysicalLineStart(src *source.Source) bool {
	start := l.Span.Start
	if start == 0 {
		return true
	}
	return src.Raw[start-1] == '\n'
}

// BlockIndent reports the line's leading indentation in visual columns and bytes.
//
// A space advances one column. A tab advances to the next multiple of
// source.TabWidth. Only leading spaces and tabs are considered.
func (l Line) BlockIndent(src *source.Source) (indentCols int, indentBytes int) {
	s := src.Slice(l.Span)

	col := 0
	pos := 0

	for pos < len(s) {
		b := s[pos]

		if b == ' ' {
			col++
			pos++
			continue
		}

		if b == '\t' {
			col += source.TabWidth - (col % source.TabWidth)
			pos++
			continue
		}

		break
	}

	return col, pos
}

// TrimIndentToCols returns a line rebased by trimming up to baselineCols
// of leading indentation.
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

// Scanner incrementally scans normalized source text into lines.
type Scanner struct {
	Input             string
	Position          source.BytePos
	pendingFinalEmpty bool
}

// NewScanner constructs a scanner over normalized source text.
func NewScanner(input string) *Scanner {
	return &Scanner{
		Input:    input,
		Position: 0,
	}
}

func (s *Scanner) EOF() bool {
	return int(s.Position) >= len(s.Input)
}

// Next returns the next scanned line.
//
// If the input ends with a newline, Next emits a final empty line before
// reporting EOF.
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
