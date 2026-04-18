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

func NewSource(s string) *Source {
	r := normalizeText(s)
	ls := computeLineStarts(r)

	return &Source{
		Raw:        r,
		LineStarts: ls,
	}
}

func (src *Source) Slice(span ByteSpan) string {
	if !src.validateSpan(span) {
		return ""
	}

	return src.Raw[int(span.Start):int(span.End)]
}

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
