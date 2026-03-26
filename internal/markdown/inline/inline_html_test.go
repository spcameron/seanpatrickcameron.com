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

func TestClosingTag(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty input",
			input: "",
			want:  false,
		},
		{
			name:  "simple closing tag",
			input: "</a>",
			want:  true,
		},
		{
			name:  "closing tag uppercase",
			input: "</A>",
			want:  true,
		},
		{
			name:  "closing tag with hyphen",
			input: "</a-b>",
			want:  true,
		},
		{
			name:  "closing tag with numeral",
			input: "</a0>",
			want:  true,
		},
		{
			name:  "longer closing tag name",
			input: "</blockquote>",
			want:  true,
		},
		{
			name:  "trailing spaces before close",
			input: "</a   >",
			want:  true,
		},
		{
			name:  "trailing tabs before close",
			input: "</a\t\t>",
			want:  true,
		},
		{
			name:  "missing opening angle bracket",
			input: "/a>",
			want:  false,
		},
		{
			name:  "missing slash",
			input: "<a>",
			want:  false,
		},
		{
			name:  "missing closing angle bracket",
			input: "</a",
			want:  false,
		},
		{
			name:  "empty closing tag candidate",
			input: "</>",
			want:  false,
		},
		{
			name:  "space after slash",
			input: "</ a>",
			want:  false,
		},
		{
			name:  "tab after slash",
			input: "</\ta>",
			want:  false,
		},
		{
			name:  "tag name begins with numeral",
			input: "</1a>",
			want:  false,
		},
		{
			name:  "tag name begins with hyphen",
			input: "</-a>",
			want:  false,
		},
		{
			name:  "invalid character in tag name",
			input: "</a*b>",
			want:  false,
		},
		{
			name:  "invalid punctuation after tag name",
			input: "</a!>",
			want:  false,
		},
		{
			name:  "junk after tag name",
			input: "</a x>",
			want:  false,
		},
		{
			name:  "attribute like tail is invalid",
			input: `</a x="y">`,
			want:  false,
		},
		{
			name:  "slash after tag name is invalid",
			input: "</a/>",
			want:  false,
		},
		{
			name:  "slash after tag name with whitespace is invalid",
			input: "</a / >",
			want:  false,
		},
		{
			name:  "extra closing angle bracket after valid closer",
			input: "</a>>",
			want:  true,
		},
		{
			name:  "extra text after valid closer",
			input: "</a>rest",
			want:  true,
		},
		{
			name:  "space before opening slash is invalid form",
			input: "< /a>",
			want:  false,
		},
		{
			name:  "space between name characters is invalid",
			input: "</a b>",
			want:  false,
		},
		{
			name:  "underscore not allowed in closing tag name",
			input: "</a_b>",
			want:  false,
		},
		{
			name:  "colon not allowed in closing tag name",
			input: "</a:b>",
			want:  false,
		},
		{
			name:  "period not allowed in closing tag name",
			input: "</a.b>",
			want:  false,
		},
		{
			name:  "double slash invalid",
			input: "</ /a>",
			want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := tryHTMLClosingTag(tc.input)
			assert.Equal(t, ok, tc.want)
		})
	}
}

func TestHTMLComment(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty input",
			input: "",
			want:  false,
		},
		{
			name:  "simple comment",
			input: "<!-- hello -->",
			want:  true,
		},
		{
			name:  "empty comment",
			input: "<!---->",
			want:  true,
		},
		{
			name:  "comment with overlapping terminator",
			input: "<!-->",
			want:  true,
		},
		{
			name:  "comment with double hyphen body",
			input: "<!-- -- -->",
			want:  true,
		},
		{
			name:  "comment with extra trailing text",
			input: "<!-- hello -->rest",
			want:  true,
		},
		{
			name:  "comment with extra trailing angle bracket",
			input: "<!-- hello -->>",
			want:  true,
		},
		{
			name:  "missing opener",
			input: "!-- hello -->",
			want:  false,
		},
		{
			name:  "not a comment opener",
			input: "<!- hello -->",
			want:  false,
		},
		{
			name:  "unterminated comment",
			input: "<!-- hello",
			want:  false,
		},
		{
			name:  "unterminated empty comment start",
			input: "<!--",
			want:  false,
		},
		{
			name:  "wrong terminator",
			input: "<!-- hello ->",
			want:  false,
		},
		{
			name:  "processing instruction opener is not comment",
			input: "<?x?>",
			want:  false,
		},
		{
			name:  "open tag opener is not comment",
			input: "<div>",
			want:  false,
		},
		{
			name:  "closing tag opener is not comment",
			input: "</div>",
			want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := tryHTMLComment(tc.input)
			assert.Equal(t, ok, tc.want)
		})
	}
}

func TestHTMLProcessingInstruction(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty input",
			input: "",
			want:  false,
		},
		{
			name:  "simple processing instruction",
			input: "<?x?>",
			want:  true,
		},
		{
			name:  "empty body processing instruction",
			input: "<?>",
			want:  true,
		},
		{
			name:  "processing instruction with spaces",
			input: "<?xml version='1.0'?>",
			want:  true,
		},
		{
			name:  "processing instruction with trailing text",
			input: "<?x?>rest",
			want:  true,
		},
		{
			name:  "processing instruction with nested question mark",
			input: "<?x?y?>",
			want:  true,
		},
		{
			name:  "missing opener",
			input: "?x?>",
			want:  false,
		},
		{
			name:  "not a processing instruction opener",
			input: "<!x?>",
			want:  false,
		},
		{
			name:  "unterminated processing instruction",
			input: "<?x",
			want:  false,
		},
		{
			name:  "missing final greater than",
			input: "<?x?",
			want:  false,
		},
		{
			name:  "open tag opener is not processing instruction",
			input: "<x>",
			want:  false,
		},
		{
			name:  "closing tag opener is not processing instruction",
			input: "</x>",
			want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := tryHTMLProcessingInstruction(tc.input)
			assert.Equal(t, ok, tc.want)
		})
	}
}

func TestHTMLDeclaration(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty input",
			input: "",
			want:  false,
		},
		{
			name:  "simple declaration",
			input: "<!DOCTYPE html>",
			want:  true,
		},
		{
			name:  "minimal declaration",
			input: "<!x>",
			want:  true,
		},
		{
			name:  "declaration with spaces",
			input: "<!ELEMENT note (to,from,heading,body)>",
			want:  true,
		},
		{
			name:  "declaration with trailing text",
			input: "<!DOCTYPE html>rest",
			want:  true,
		},
		{
			name:  "comment form also matches broad declaration helper",
			input: "<!-- x -->",
			want:  true,
		},
		{
			name:  "cdata form also matches broad declaration helper",
			input: "<![CDATA[x]]>",
			want:  true,
		},
		{
			name:  "missing opener",
			input: "!DOCTYPE html>",
			want:  false,
		},
		{
			name:  "bare opener without closer",
			input: "<!",
			want:  false,
		},
		{
			name:  "unterminated declaration",
			input: "<!DOCTYPE html",
			want:  false,
		},
		{
			name:  "processing instruction opener is not declaration",
			input: "<?x?>",
			want:  false,
		},
		{
			name:  "open tag opener is not declaration",
			input: "<div>",
			want:  false,
		},
		{
			name:  "closing tag opener is not declaration",
			input: "</div>",
			want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := tryHTMLDeclaration(tc.input)
			assert.Equal(t, ok, tc.want)
		})
	}
}

func TestHTMLCDATA(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty input",
			input: "",
			want:  false,
		},
		{
			name:  "simple cdata",
			input: "<![CDATA[x]]>",
			want:  true,
		},
		{
			name:  "empty cdata",
			input: "<![CDATA[]]>",
			want:  true,
		},
		{
			name:  "cdata with markup like content",
			input: "<![CDATA[<div>&stuff</div>]]>",
			want:  true,
		},
		{
			name:  "cdata with trailing text",
			input: "<![CDATA[x]]>rest",
			want:  true,
		},
		{
			name:  "missing opener",
			input: "![CDATA[x]]>",
			want:  false,
		},
		{
			name:  "wrong cdata opener spelling",
			input: "<![cdata[x]]>",
			want:  false,
		},
		{
			name:  "missing final terminator",
			input: "<![CDATA[x]]",
			want:  false,
		},
		{
			name:  "unterminated cdata",
			input: "<![CDATA[x",
			want:  false,
		},
		{
			name:  "processing instruction opener is not cdata",
			input: "<?x?>",
			want:  false,
		},
		{
			name:  "open tag opener is not cdata",
			input: "<div>",
			want:  false,
		},
		{
			name:  "closing tag opener is not cdata",
			input: "</div>",
			want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := tryHTMLCDATA(tc.input)
			assert.Equal(t, ok, tc.want)
		})
	}
}
