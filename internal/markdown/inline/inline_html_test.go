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
		{
			name:  "self closing",
			input: "<br/>",
			want:  true,
		},
		{
			name:  "self closing with space before slash",
			input: "<br />",
			want:  true,
		},
		{
			name:  "self closing with trailing space after slash",
			input: "<br/   >",
			want:  true,
		},
		{
			name:  "missing opening angle bracket",
			input: "a>",
			want:  false,
		},
		{
			name:  "missing closing angle bracket",
			input: "<a",
			want:  false,
		},
		{
			name:  "empty tag candidate",
			input: "<>",
			want:  false,
		},
		{
			name:  "tag name begins with numeral",
			input: "<1a>",
			want:  false,
		},
		{
			name:  "tag name begins with hyphen",
			input: "<-a>",
			want:  false,
		},
		{
			name:  "space after opening angle bracket",
			input: "< a>",
			want:  false,
		},
		{
			name:  "invalid character in tag name",
			input: "<a*b>",
			want:  false,
		},
		{
			name:  "invalid punctuation after tag name",
			input: "<a!>",
			want:  false,
		},
		{
			name:  "double slash self closing tail",
			input: "<a//>",
			want:  false,
		},
		{
			name:  "slash followed by junk",
			input: "<a/x>",
			want:  false,
		},
		{
			name:  "slash separated by spaces then junk",
			input: "<a / x>",
			want:  false,
		},
		{
			name:  "attribute with name only (no value)",
			input: "<a b>",
			want:  true,
		},
		{
			name:  "multiple bare attributes",
			input: "<a b c>",
			want:  true,
		},
		{
			name:  "attribute with unquoted value",
			input: "<a b=c>",
			want:  true,
		},
		{
			name:  "attribute with double quoted value",
			input: `<a b="c">`,
			want:  true,
		},
		{
			name:  "attribute with single quoted value",
			input: "<a b='c'>",
			want:  true,
		},
		{
			name:  "multiple attributes with values",
			input: `<a b="c" d='e' f=g>`,
			want:  true,
		},
		{
			name:  "attributes with whitespace around equals",
			input: `<a b = "c">`,
			want:  true,
		},
		{
			name:  "attribute followed by self closing suffix",
			input: `<img src="x" />`,
			want:  true,
		},
		{
			name:  "adjacent attributes without separator",
			input: `<a b="c"d="e">`,
			want:  false,
		},
		{
			name:  "equals without attribute name",
			input: "< a =x>",
			want:  false,
		},
		{
			name:  "attribute value missing after equals",
			input: `<a b=>`,
			want:  false,
		},
		{
			name:  "unterminated double quoted value",
			input: `<a b="c>`,
			want:  false,
		},
		{
			name:  "unterminated single quoted value",
			input: "<a b='c>",
			want:  false,
		},
		{
			name:  "quoted greater than inside double quoted value",
			input: `<a b=">">`,
			want:  true,
		},
		{
			name:  "quoted greater than inside single quoted value",
			input: "<a b='>'>",
			want:  true,
		},
		{
			name:  "greater than in quoted value followed by more attributes",
			input: `<a b=">" c="d">`,
			want:  true,
		},
		{
			name:  "junk after quoted value without separator",
			input: `<a b="c"x>`,
			want:  false,
		},
		{
			name:  "trailing whitespace after attribute before close",
			input: `<a b="c"   >`,
			want:  true,
		},
		{
			name:  "trailing whitespace after attribute before self closing suffix",
			input: `<a b="c"   />`,
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
