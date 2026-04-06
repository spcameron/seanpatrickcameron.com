package markdown_test

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestCommonMarkSpec(t *testing.T) {
	testCases := []struct {
		name  string
		md    string
		html  string
		match bool
	}{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := markdown.HTML(tc.md)

			if tc.match {
				assert.Equal(t, got, tc.html)
				assert.NoError(t, err)
			} else if got == tc.html {
				t.Errorf("unexpected match with CommonMark output")
			}
		})
	}
}
