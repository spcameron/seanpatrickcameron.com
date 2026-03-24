package inline

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestOpenTag(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "simple tag name",
			input: "<a>",
			want:  true,
		},
		{
			name:  "tag name uppercase",
			input: "<A>",
			want:  true,
		},
		{
			name:  "tag name with hyphen",
			input: "<a-b>",
			want:  true,
		},
		{
			name:  "tag name with numeral",
			input: "<a0>",
			want:  true,
		},
		{
			name:  "longer tag name",
			input: "<blockquote>",
			want:  true,
		},
		{
			name:  "trailing spaces before close",
			input: "<a   >",
			want:  true,
		},
		{
			name:  "trailing tabs before close",
			input: "<a\t\t>",
			want:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := tryHTMLOpenTag(tc.input)
			assert.Equal(t, ok, tc.want)
		})
	}
}
