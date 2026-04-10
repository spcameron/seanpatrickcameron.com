package html

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestRender(t *testing.T) {
	testCases := []struct {
		name    string
		node    Node
		want    string
		wantErr error
	}{
		// Text nodes
		{
			name:    "text: plain text",
			node:    TextNode("test text"),
			want:    "test text",
			wantErr: nil,
		},
		{
			name:    "text: empty text",
			node:    TextNode(""),
			want:    "",
			wantErr: nil,
		},
		{
			name:    "text: preserves newline",
			node:    TextNode("line one\nline two"),
			want:    "line one\nline two",
			wantErr: nil,
		},
		{
			name:    "text: escapes lt, gt, amp",
			node:    TextNode("Hello <world> & friends"),
			want:    "Hello &lt;world&gt; &amp; friends",
			wantErr: nil,
		},
		{
			name:    "text: escapes double quotes",
			node:    TextNode("\"Good morning, Sean.\""),
			want:    "&#34;Good morning, Sean.&#34;",
			wantErr: nil,
		},
		{
			name:    "text: preserves unicode",
			node:    TextNode("café — π ≈ 3.14159 — 你好"),
			want:    "café — π ≈ 3.14159 — 你好",
			wantErr: nil,
		},

		// Raw nodes
		{
			name:    "raw: renders without escaping",
			node:    RawNode("<span>raw & literal</span>"),
			want:    "<span>raw & literal</span>",
			wantErr: nil,
		},
		{
			name:    "raw: empty raw",
			node:    RawNode(""),
			want:    "",
			wantErr: nil,
		},

		// Fragment nodes
		{
			name:    "fragment: empty fragment",
			node:    FragmentNode(),
			want:    "",
			wantErr: nil,
		},
		{
			name: "fragment: multiple children",
			node: FragmentNode(
				TextNode("a"),
				ElemNode("span", nil, TextNode("b")),
				TextNode("c"),
			),
			want:    "a<span>b</span>c",
			wantErr: nil,
		},
		{
			name: "fragment: nested elements",
			node: FragmentNode(
				ElemNode("p", nil, TextNode("first")),
				ElemNode("p", nil, TextNode("second")),
			),
			want:    "<p>first</p><p>second</p>",
			wantErr: nil,
		},

		// Element nodes
		{
			name:    "element: no attributes, no children",
			node:    ElemNode("span", nil),
			want:    "<span></span>",
			wantErr: nil,
		},
		{
			name:    "element: one text child",
			node:    ElemNode("p", nil, TextNode("test text")),
			want:    "<p>test text</p>",
			wantErr: nil,
		},
		{
			name: "element: multiple children, text and elements",
			node: ElemNode(
				"header",
				nil,
				ElemNode("span", nil),
				TextNode("test text"),
				ElemNode("span", nil),
			),
			want:    "<header><span></span>test text<span></span></header>",
			wantErr: nil,
		},
		{
			name: "element: nested elements deep",
			node: ElemNode(
				"main",
				nil,
				ElemNode(
					"ul",
					nil,
					ElemNode(
						"li",
						nil,
						ElemNode(
							"p",
							nil,
							TextNode("first list item"),
						),
					),
					ElemNode(
						"li",
						nil,
						ElemNode(
							"p",
							nil,
							TextNode("second list item"),
						),
					),
				),
			),
			want:    "<main><ul><li><p>first list item</p></li><li><p>second list item</p></li></ul></main>",
			wantErr: nil,
		},
		{
			name: "element: single attribute",
			node: ElemNode(
				"a",
				Attributes{"href": "https://www.google.com"},
				TextNode("click me"),
			),
			want:    `<a href="https://www.google.com">click me</a>`,
			wantErr: nil,
		},
		{
			name: "element: multiple attributes, sorted",
			node: ElemNode(
				"div",
				Attributes{
					"src": "/static/images/foo.png",
					"alt": "foo_picture",
				},
			),
			want:    `<div alt="foo_picture" src="/static/images/foo.png"></div>`,
			wantErr: nil,
		},
		{
			name: "element: escapes special characters in attributes",
			node: ElemNode(
				"div",
				Attributes{
					"src": "/static/images/img.png",
					"alt": "Hello <world> & \"friends\"",
				},
			),
			want:    `<div alt="Hello &lt;world&gt; &amp; &#34;friends&#34;" src="/static/images/img.png"></div>`,
			wantErr: nil,
		},
		{
			name: "element: preserves unicode in attributes",
			node: ElemNode(
				"div",
				Attributes{
					"title": "café — 你好",
				},
			),
			want:    `<div title="café — 你好"></div>`,
			wantErr: nil,
		},

		// Void elements
		{
			name:    "void element: no attributes",
			node:    VoidNode("br", nil),
			want:    "<br>",
			wantErr: nil,
		},
		{
			name: "void element: multiple attributes",
			node: VoidNode(
				"img",
				Attributes{
					"src": "/static/images/img.png",
					"alt": "an image",
				},
			),
			want:    `<img alt="an image" src="/static/images/img.png">`,
			wantErr: nil,
		},
		{
			name: "void element: attribute with empty value",
			node: VoidNode(
				"img",
				Attributes{
					"alt": "",
				},
			),
			want:    `<img alt="">`,
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Render(tc.node)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
