package source

import (
	"sort"
	"strings"
)

const (
	TabWidth = 4
)

type Source struct {
	Raw        string
	LineStarts []BytePos
}

// NewSource constructs a Source from raw input text.
//
// Input is normalized to the package's canonical newline representation,
// and line start offsets are computed for later span and line lookups.
func NewSource(s string) *Source {
	r := normalizeText(s)
	ls := computeLineStarts(r)

	return &Source{
		Raw:        r,
		LineStarts: ls,
	}
}

// Slice returns the substring covered by span.
//
// If span is invalid for this source, Slice returns the empty string.
func (src *Source) Slice(span ByteSpan) string {
	if !src.validateSpan(span) {
		return ""
	}

	return src.Raw[int(span.Start):int(span.End)]
}

// UnescapedSlice returns the substring covered by span with Markdown
// backslash escapes resolved.
//
// A backslash is removed only when it escapes escapable punctuation.
// All other bytes are preserved as written.
func (src *Source) UnescapedSlice(span ByteSpan) string {
	s := src.Slice(span)

	var sb strings.Builder
	sb.Grow(len(s))

	pos := 0
	for pos < len(s) {
		b := s[pos]

		if b != '\\' {
			sb.WriteByte(b)
			pos++
			continue
		}

		if pos+1 >= len(s) {
			sb.WriteByte('\\')
			pos++
			continue
		}

		next := s[pos+1]
		if IsEscapablePunctuation(next) {
			sb.WriteByte(next)
			pos += 2
			continue
		}

		sb.WriteByte('\\')
		pos++
	}

	return sb.String()
}

// LineColumn reports the zero-based line and column for pos.
//
// Positions outside the source are clamped into the range [0, EOF()].
// The reported column is a byte offset from the start of the line.
func (src *Source) LineColumn(pos BytePos) (line int, col int) {
	pos = max(0, pos)
	pos = min(pos, src.EOF())

	i := sort.Search(len(src.LineStarts), func(i int) bool {
		return src.LineStarts[i] > pos
	})

	idx := max(0, i-1)
	ls := src.LineStarts[idx]
	col = int(pos - ls)

	return idx, col
}

// LineSpan returns the span of the specified line.
//
// The returned span covers the line's content and excludes any trailing
// newline byte. Out-of-range line numbers are clamped.
func (src *Source) LineSpan(line int) ByteSpan {
	if len(src.LineStarts) == 0 {
		eof := src.EOF()
		return ByteSpan{
			Start: eof,
			End:   eof,
		}
	}

	line = max(0, line)
	line = min(line, len(src.LineStarts)-1)

	start := src.LineStarts[line]
	end := src.EOF()

	if next := line + 1; next < len(src.LineStarts) {
		end = src.LineStarts[next] - 1
	}

	end = max(start, end)
	end = min(end, src.EOF())

	return ByteSpan{
		Start: start,
		End:   end,
	}
}

// LineSpansWithin returns the non-empty per-line fragments of span.
//
// Each returned span is clipped to both the requested span and the
// boundaries of its containing line.
func (src *Source) LineSpansWithin(span ByteSpan) []ByteSpan {
	if span.Start >= span.End {
		return []ByteSpan{}
	}

	firstLine, _ := src.LineColumn(span.Start)
	lastLine, _ := src.LineColumn(span.End - 1)

	spans := make([]ByteSpan, 0, lastLine-firstLine+1)
	for i := firstLine; i <= lastLine; i++ {
		lineSpan := src.LineSpan(i)
		clamped := ByteSpan{
			Start: max(lineSpan.Start, span.Start),
			End:   min(lineSpan.End, span.End),
		}

		if clamped.Start < clamped.End {
			spans = append(spans, clamped)
		}
	}

	return spans
}

func (src *Source) validateSpan(span ByteSpan) bool {
	if 0 <= span.Start &&
		span.Start <= span.End &&
		span.End <= src.EOF() {
		return true
	}

	return false
}

func (src *Source) EOF() BytePos {
	return BytePos(len(src.Raw))
}

// IsEscapablePunctuation reports whether b may be escaped with a
// backslash under the package's Markdown rules.
func IsEscapablePunctuation(b byte) bool {
	switch b {
	case '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
		':', ';', '<', '=', '>', '?', '@',
		'[', '\\', ']', '^', '_', '`',
		'{', '|', '}', '~':
		return true
	default:
		return false
	}
}

func normalizeText(s string) string {
	out := strings.ReplaceAll(s, "\r\n", "\n")
	out = strings.ReplaceAll(out, "\r", "\n")
	return out
}

func computeLineStarts(s string) []BytePos {
	ls := []BytePos{0}
	for i := range s {
		if s[i] == '\n' {
			pos := BytePos(i + 1)
			ls = append(ls, pos)
		}
	}

	return ls
}
