package html_test

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/html"
	tk "github.com/spcameron/seanpatrickcameron.com/internal/markdown/testkit"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestRender(t *testing.T) {
	testCases := []struct {
		name    string
		node    html.Node
		want    string
		wantErr error
	}{
		// Text nodes
		{
			name:    "text: plain text",
			node:    tk.HTMLTextNode("test text"),
			want:    "test text",
			wantErr: nil,
		},
		{
			name:    "text: empty text",
			node:    tk.HTMLTextNode(""),
			want:    "",
			wantErr: nil,
		},
		{
			name:    "text: preserves newline",
			node:    tk.HTMLTextNode("line one\nline two"),
			want:    "line one\nline two",
			wantErr: nil,
		},
		{
			name:    "text: escapes lt, gt, amp",
			node:    tk.HTMLTextNode("Hello <world> & friends"),
			want:    "Hello &lt;world&gt; &amp; friends",
			wantErr: nil,
		},
		{
			name:    "text: escapes double quotes",
			node:    tk.HTMLTextNode("\"Good morning, Sean.\""),
			want:    "&#34;Good morning, Sean.&#34;",
			wantErr: nil,
		},
		{
			name:    "text: preserves unicode",
			node:    tk.HTMLTextNode("café — π ≈ 3.14159 — 你好"),
			want:    "café — π ≈ 3.14159 — 你好",
			wantErr: nil,
		},

		// Raw nodes
		{
			name:    "raw: renders without escaping",
			node:    tk.HTMLRawNode("<span>raw & literal</span>"),
			want:    "<span>raw & literal</span>",
			wantErr: nil,
		},
		{
			name:    "raw: empty raw",
			node:    tk.HTMLRawNode(""),
			want:    "",
			wantErr: nil,
		},

		// Fragment nodes
		{
			name:    "fragment: empty fragment",
			node:    tk.HTMLFragmentNode(),
			want:    "",
			wantErr: nil,
		},
		{
			name: "fragment: multiple children",
			node: tk.HTMLFragmentNode(
				tk.HTMLTextNode("a"),
				tk.HTMLElemNode("span", nil, tk.HTMLTextNode("b")),
				tk.HTMLTextNode("c"),
			),
			want:    "a<span>b</span>c",
			wantErr: nil,
		},
		{
			name: "fragment: nested elements",
			node: tk.HTMLFragmentNode(
				tk.HTMLElemNode("p", nil, tk.HTMLTextNode("first")),
				tk.HTMLElemNode("p", nil, tk.HTMLTextNode("second")),
			),
			want:    "<p>first</p><p>second</p>",
			wantErr: nil,
		},

		// Element nodes
		{
			name:    "element: no attributes, no children",
			node:    tk.HTMLElemNode("span", nil),
			want:    "<span></span>",
			wantErr: nil,
		},
		{
			name:    "element: one text child",
			node:    tk.HTMLElemNode("p", nil, tk.HTMLTextNode("test text")),
			want:    "<p>test text</p>",
			wantErr: nil,
		},
		{
			name: "element: multiple children, text and elements",
			node: tk.HTMLElemNode(
				"header",
				nil,
				tk.HTMLElemNode("span", nil),
				tk.HTMLTextNode("test text"),
				tk.HTMLElemNode("span", nil),
			),
			want:    "<header><span></span>test text<span></span></header>",
			wantErr: nil,
		},
		{
			name: "element: nested elements deep",
			node: tk.HTMLElemNode(
				"main",
				nil,
				tk.HTMLElemNode(
					"ul",
					nil,
					tk.HTMLElemNode(
						"li",
						nil,
						tk.HTMLElemNode(
							"p",
							nil,
							tk.HTMLTextNode("first list item"),
						),
					),
					tk.HTMLElemNode(
						"li",
						nil,
						tk.HTMLElemNode(
							"p",
							nil,
							tk.HTMLTextNode("second list item"),
						),
					),
				),
			),
			want:    "<main><ul><li><p>first list item</p></li><li><p>second list item</p></li></ul></main>",
			wantErr: nil,
		},
		{
			name: "element: single attribute",
			node: tk.HTMLElemNode(
				"a",
				html.Attributes{"href": "https://www.google.com"},
				tk.HTMLTextNode("click me"),
			),
			want:    `<a href="https://www.google.com">click me</a>`,
			wantErr: nil,
		},
		{
			name: "element: multiple attributes, sorted",
			node: tk.HTMLElemNode(
				"div",
				html.Attributes{
					"src": "/static/images/foo.png",
					"alt": "foo_picture",
				},
			),
			want:    `<div alt="foo_picture" src="/static/images/foo.png"></div>`,
			wantErr: nil,
		},
		{
			name: "element: escapes special characters in attributes",
			node: tk.HTMLElemNode(
				"div",
				html.Attributes{
					"src": "/static/images/img.png",
					"alt": "Hello <world> & \"friends\"",
				},
			),
			want:    `<div alt="Hello &lt;world&gt; &amp; &#34;friends&#34;" src="/static/images/img.png"></div>`,
			wantErr: nil,
		},
		{
			name: "element: preserves unicode in attributes",
			node: tk.HTMLElemNode(
				"div",
				html.Attributes{
					"title": "café — 你好",
				},
			),
			want:    `<div title="café — 你好"></div>`,
			wantErr: nil,
		},

		// Void elements
		{
			name:    "void element: no attributes",
			node:    tk.HTMLVoidNode("br", nil),
			want:    "<br>",
			wantErr: nil,
		},
		{
			name: "void element: multiple attributes",
			node: tk.HTMLVoidNode(
				"img",
				html.Attributes{
					"src": "/static/images/img.png",
					"alt": "an image",
				},
			),
			want:    `<img alt="an image" src="/static/images/img.png">`,
			wantErr: nil,
		},
		{
			name: "void element: attribute with empty value",
			node: tk.HTMLVoidNode(
				"img",
				html.Attributes{
					"alt": "",
				},
			),
			want:    `<img alt="">`,
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := html.Render(tc.node)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
