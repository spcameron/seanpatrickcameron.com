package markdown_test

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/testhtml"
)

// Based on CommonMark Version 0.31.2 (2024-01-28).
// These cases use CommonMark examples as a compliance reference while also
// documenting intentional Scribe divergences where design choices differ.

type compareMode int

const (
	_ compareMode = iota
	compareExact
	compareStructural
	compareDocumentedDivergence
)

func TestCommonMarkSpec(t *testing.T) {
	testCases := []struct {
		name   string
		md     string
		cm     string
		scribe string
		mode   compareMode
		reason string
	}{
		// Section 2 - Preliminaries
		//
		// 2.2 - Tabs
		{
			name:   "1: tabs define block structure",
			md:     "\tfoo\tbaz\t\tbim",
			cm:     "<pre><code>foo\tbaz\t\tbim\n</code></pre>",
			scribe: "<pre><code>foo\tbaz\t\tbim</code></pre>",
			mode:   compareDocumentedDivergence,
			reason: "final code block newline not preserved",
		},
		{
			name:   "2: tabs define block structure",
			md:     "  \tfoo\tbaz\t\tbim",
			cm:     "<pre><code>foo\tbaz\t\tbim\n</code></pre>",
			scribe: "<pre><code>foo\tbaz\t\tbim</code></pre>",
			mode:   compareDocumentedDivergence,
			reason: "final code block newline not preserved",
		},
		{
			name:   "3: tabs define block structure",
			md:     "    a→a\n    ὐ→a",
			cm:     "<pre><code>a→a\nὐ→a\n</code></pre>",
			scribe: "<pre><code>a→a\nὐ→a</code></pre>",
			mode:   compareDocumentedDivergence,
			reason: "final code block newline not preserved",
		},
		{
			name:   "4: paragraph continuation of a list item",
			md:     "  - foo\n\n\tbar",
			cm:     "<ul>\n<li>\n<p>foo</p>\n<p>bar</p>\n</li>\n</ul>",
			mode:   compareStructural,
			reason: "HTML serialization formatting differs; structure matches CommonMark",
		},
		{
			name:   "5: code block continuation of a list item",
			md:     "- foo\n\n\t\tbar",
			cm:     "<ul>\n<li>\n<p>foo</p>\n<pre><code>  bar\n</code></pre>\n</li>\n</ul>",
			scribe: "<ul><li><p>foo</p><pre><code>bar</code></pre></li></ul>",
			mode:   compareDocumentedDivergence,
			reason: "tab-based indentation is trimmed by whole-byte span cuts, so CommonMark's partial-tab space preservation is not reproduced",
		},
		{
			name:   "6: block quote marker followed by tabs",
			md:     ">\t\tfoo",
			cm:     "<blockquote>\n<pre><code>  foo\n</code></pre>\n</blockquote>",
			scribe: "<blockquote><pre><code>foo</code></pre></blockquote>",
			mode:   compareDocumentedDivergence,
			reason: "tab-based indentation is trimmed by whole-byte span cuts, so CommonMark's partial-tab space preservation is not reproduced",
		},
		{
			name:   "7: list item followed by tabs",
			md:     "-\t\tfoo",
			cm:     "<ul>\n<li>\n<pre><code>  foo\n</code></pre>\n</li>\n</ul>",
			scribe: "<ul><li>foo</li></ul>",
			mode:   compareDocumentedDivergence,
			reason: "list marker parsing consumes the full post-marker tab run as delimiter whitespace, so no indentation remains to form a code block",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cm == "" {
				t.Fatalf("CM HTML must be defined")
			}

			got, err := markdown.HTML(tc.md)
			assert.NoError(t, err)

			switch tc.mode {
			case compareExact:
				assert.Equal(t, got, tc.cm)

			case compareStructural:
				wantTree, gotTree, err := testhtml.ParseAndNormalizePair(tc.cm, got)
				assert.NoError(t, err)
				assert.Equal(t, gotTree, wantTree)

			case compareDocumentedDivergence:
				if tc.scribe == "" {
					t.Fatalf("scribe HTML must be defined for documented divergences")
				}

				assert.Equal(t, got, tc.scribe)

				if got == tc.cm {
					t.Fatalf("expected divergence from CommonMark, but output matched exactly")
				}

			default:
				t.Fatalf("unknown comparison mode: %v", tc.mode)
			}
		})
	}
}
