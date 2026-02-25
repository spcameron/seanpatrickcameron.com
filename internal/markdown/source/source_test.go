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
			name:  "one internal newline emits two line starts",
			input: "a\nb",
			want:  []BytePos{0, 2},
		},
		{
			name:  "trailing newline includes empty last line",
			input: "a\n",
			want:  []BytePos{0, 2},
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
			name:  "empty span (start == end)",
			input: "hello world",
			span:  ByteSpan{2, 2},
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := NewSource(tc.input)
			got := src.Slice(tc.span)

			assert.Equal(t, got, tc.want)
		})
	}
}
