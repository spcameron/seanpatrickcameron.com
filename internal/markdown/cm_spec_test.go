package markdown_test

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

// Based on Version 0.31.2 (2024-01-28)
// The following test cases are based on the CommonMark spec examples, and are used
// to help illustrate where Scribe diverges in implementation from the CM spec.

func TestCommonMarkSpec(t *testing.T) {
	testCases := []struct {
		name      string
		md        string
		cmHTML    string
		localHTML string
		matchCM   bool
		reason    string
	}{
		// Section 2 - Preliminaries
		//
		// 2.2 - Tabs
		{
			name:      "1: tabs define block structure; code block payload ends with newline in CM",
			md:        "\tfoo\tbaz\t\tbim",
			cmHTML:    "<pre><code>foo\tbaz\t\tbim\n</code></pre>",
			localHTML: "<pre><code>foo\tbaz\t\tbim</code></pre>",
			matchCM:   false,
			reason:    "final code block newline not preserved",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cmHTML == "" {
				t.Fatalf("cmHTML must be defined")
			}
			if !tc.matchCM && tc.localHTML == "" {
				t.Fatalf("localHTML must be defined when matchCM is false")
			}

			got, err := markdown.HTML(tc.md)
			assert.NoError(t, err)

			if tc.matchCM {
				assert.Equal(t, got, tc.cmHTML)
				return
			}

			assert.Equal(t, got, tc.localHTML)

			if got == tc.cmHTML {
				t.Errorf("unexpected match with CommonMark output: %s", tc.reason)
			}
		})
	}
}
