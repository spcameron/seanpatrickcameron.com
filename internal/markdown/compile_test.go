package markdown_test

import (
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := markdown.CompileAndRender(tc.md)

			assert.Equal(t, got, tc.html)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
