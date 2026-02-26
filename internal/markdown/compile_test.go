package markdown_test

import (
	"strings"
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestCompile(t *testing.T) {
	testCases := []struct {
		name    string
		md      string
		html    string
		wantErr error
	}{
		{
			name:    "plain text: one paragraph",
			md:      "hello",
			html:    `<p>hello</p>`,
			wantErr: nil,
		},
		{
			name:    "plain text: soft break paragraph",
			md:      "a\nb",
			html:    `<p>a b</p>`,
			wantErr: nil,
		},
		{
			name:    "plain text: hard break paragraph (spaces)",
			md:      "a  \nb",
			html:    `<p>a<br>b</p>`,
			wantErr: nil,
		},
		{
			name:    "plain text: hard break paragraph (backslash)",
			md:      "a\\\nb",
			html:    `<p>a<br>b</p>`,
			wantErr: nil,
		},
		{
			name:    "plain text: blank line splits paragraphs",
			md:      "a\n\nb",
			html:    `<p>a</p><p>b</p>`,
			wantErr: nil,
		},
		{
			name:    "header: level 1",
			md:      "# header",
			html:    `<h1>header</h1>`,
			wantErr: nil,
		},
		{
			name:    "header: level 6",
			md:      "###### header",
			html:    `<h6>header</h6>`,
			wantErr: nil,
		},
		{
			name: "header then paragraph",
			md: strings.Join([]string{
				"# h",
				"a",
			}, "\n"),
			html:    `<h1>h</h1><p>a</p>`,
			wantErr: nil,
		},
		{
			name: "paragraph then header",
			md: strings.Join([]string{
				"a",
				"# h",
			}, "\n"),
			html:    `<p>a</p><h1>h</h1>`,
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := markdown.HTML(tc.md)

			assert.Equal(t, got, tc.html)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
