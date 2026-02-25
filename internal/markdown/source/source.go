package source

import (
	"sort"
	"strings"
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

// TODO: test coverage
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
