package source

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestNormalizeText(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "normal newline unchanged",
			input: "a\n",
			want:  "a\n",
		},
		{
			name:  "CRLF normalized to newline",
			input: "a\r\n",
			want:  "a\n",
		},
		{
			name:  "stray carriage return normalized to newline",
			input: "a\r",
			want:  "a\n",
		},
		{
			name:  "mixed line endings normalized in one pass",
			input: "a\r\nb\rc\n",
			want:  "a\nb\nc\n",
		},
		{
			name:  "consecutive carriage returns each normalize to newline",
			input: "a\r\rb",
			want:  "a\n\nb",
		},
		{
			name:  "consecutive CRLF pairs each normalize to newline",
			input: "a\r\n\r\nb",
			want:  "a\n\nb",
		},
		{
			name:  "only line endings normalize correctly",
			input: "\r\r\n\n",
			want:  "\n\n\n",
		},
		{
			name:  "trailing mixed line endings normalize correctly",
			input: "a\r\nb\r",
			want:  "a\nb\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeText(tc.input)

			assert.Equal(t, got, tc.want)
		})
	}
}

func TestComputeLineStarts(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  []BytePos
	}{
		{
			name:  "empty input",
			input: "",
			want:  []BytePos{0},
		},
		{
			name:  "single line without newline",
			input: "abc",
			want:  []BytePos{0},
		},
		{
			name:  "one internal newline emits two line starts",
			input: "a\nb",
			want:  []BytePos{0, 2},
		},
		{
			name:  "trailing newline includes empty last line",
			input: "a\n",
			want:  []BytePos{0, 2},
		},
		{
			name:  "single newline creates empty first and second lines",
			input: "\n",
			want:  []BytePos{0, 1},
		},
		{
			name:  "multiple internal newlines",
			input: "a\nb\nc",
			want:  []BytePos{0, 2, 4},
		},
		{
			name:  "multiple trailing blank lines",
			input: "a\n\n",
			want:  []BytePos{0, 2, 3},
		},
		{
			name:  "empty first line followed by content",
			input: "\na",
			want:  []BytePos{0, 1},
		},
		{
			name:  "only blank lines",
			input: "\n\n",
			want:  []BytePos{0, 1, 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeLineStarts(tc.input)

			assert.Equal(t, got, tc.want)
		})
	}
}

func TestSlice(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		span  ByteSpan
		want  string
	}{
		{
			name:  "valid span",
			input: "hello world",
			span:  ByteSpan{3, 7},
			want:  "lo w",
		},
		{
			name:  "full span",
			input: "hello world",
			span:  ByteSpan{0, 11},
			want:  "hello world",
		},
		{
			name:  "slice from start",
			input: "hello world",
			span:  ByteSpan{0, 5},
			want:  "hello",
		},
		{
			name:  "slice to end",
			input: "hello world",
			span:  ByteSpan{6, 11},
			want:  "world",
		},
		{
			name:  "empty span (start == end)",
			input: "hello world",
			span:  ByteSpan{2, 2},
			want:  "",
		},
		{
			name:  "empty source with zero span",
			input: "",
			span:  ByteSpan{0, 0},
			want:  "",
		},
		{
			name:  "span exactly at EOF",
			input: "hello world",
			span:  ByteSpan{11, 11},
			want:  "",
		},
		{
			name:  "span start < zero returns empty string",
			input: "hello world",
			span:  ByteSpan{-1, 1},
			want:  "",
		},
		{
			name:  "span end > len(input) returns empty string",
			input: "hello world",
			span:  ByteSpan{1, 999},
			want:  "",
		},
		{
			name:  "span start > span end returns empty string",
			input: "hello world",
			span:  ByteSpan{7, 3},
			want:  "",
		},
		{
			name:  "slice observes normalized source content",
			input: "a\r\nb",
			span:  ByteSpan{1, 2},
			want:  "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := NewSource(tc.input)
			got := src.Slice(tc.span)

			assert.Equal(t, got, tc.want)
		})
	}
}

func TestLineColumn(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		pos      BytePos
		wantLine int
		wantCol  int
	}{
		{
			name:     "empty source at zero",
			input:    "",
			pos:      0,
			wantLine: 0,
			wantCol:  0,
		},
		{
			name:     "position zero on non-empty source",
			input:    "abc",
			pos:      0,
			wantLine: 0,
			wantCol:  0,
		},
		{
			name:     "middle of first line",
			input:    "abc",
			pos:      2,
			wantLine: 0,
			wantCol:  2,
		},
		{
			name:     "exact newline byte belongs to preceding line",
			input:    "a\nb",
			pos:      1,
			wantLine: 0,
			wantCol:  1,
		},
		{
			name:     "first byte after newline is next line column zero",
			input:    "a\nb",
			pos:      2,
			wantLine: 1,
			wantCol:  0,
		},
		{
			name:     "EOF on non-empty input clamps to end of last line",
			input:    "a\nb",
			pos:      3,
			wantLine: 1,
			wantCol:  1,
		},
		{
			name:     "negative position clamps to zero",
			input:    "a\nb",
			pos:      -5,
			wantLine: 0,
			wantCol:  0,
		},
		{
			name:     "position past EOF clamps to EOF",
			input:    "a\nb",
			pos:      999,
			wantLine: 1,
			wantCol:  1,
		},
		{
			name:     "EOF on trailing empty line",
			input:    "a\n",
			pos:      2,
			wantLine: 1,
			wantCol:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := NewSource(tc.input)
			gotLine, gotCol := src.LineColumn(tc.pos)

			assert.Equal(t, gotLine, tc.wantLine)
			assert.Equal(t, gotCol, tc.wantCol)
		})
	}
}

func TestLineSpan(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		line  int
		want  ByteSpan
	}{
		{
			name:  "empty source returns zero span",
			input: "",
			line:  0,
			want:  ByteSpan{0, 0},
		},
		{
			name:  "single line without newline",
			input: "abc",
			line:  0,
			want:  ByteSpan{0, 3},
		},
		{
			name:  "single line with trailing newline excludes newline byte",
			input: "abc\n",
			line:  0,
			want:  ByteSpan{0, 3},
		},
		{
			name:  "final empty line after trailing newline is zero width at EOF",
			input: "abc\n",
			line:  1,
			want:  ByteSpan{4, 4},
		},
		{
			name:  "middle line in multi-line input",
			input: "a\nbc\ndef",
			line:  1,
			want:  ByteSpan{2, 4},
		},
		{
			name:  "last line in multi-line input extends to EOF",
			input: "a\nbc\ndef",
			line:  2,
			want:  ByteSpan{5, 8},
		},
		{
			name:  "negative line index clamps to first line",
			input: "a\nb",
			line:  -1,
			want:  ByteSpan{0, 1},
		},
		{
			name:  "too-large line index clamps to last line",
			input: "a\nb",
			line:  99,
			want:  ByteSpan{2, 3},
		},
		{
			name:  "too-large line index clamps to final empty line when trailing newline present",
			input: "a\n",
			line:  99,
			want:  ByteSpan{2, 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := NewSource(tc.input)
			got := src.LineSpan(tc.line)

			assert.Equal(t, got, tc.want)
		})
	}
}
