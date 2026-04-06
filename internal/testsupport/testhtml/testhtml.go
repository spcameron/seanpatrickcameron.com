package testhtml

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type NodeKind int

const (
	_ NodeKind = iota
	KindElement
	KindText
	KindComment
)

type Node struct {
	Kind     NodeKind
	Data     string
	Attrs    []Attr
	Children []Node
}

type Attr struct {
	Key string
	Val string
}

// ParseAndNormalizePair parses two HTML fragments and returns canonical,
// normalized trees suitable for structural comparison in tests.
func ParseAndNormalizePair(wantHTML, gotHTML string) ([]Node, []Node, error) {
	want, err := ParseAndNormalizeFragment(wantHTML)
	if err != nil {
		return nil, nil, fmt.Errorf("parse want fragment: %w", err)
	}

	got, err := ParseAndNormalizeFragment(gotHTML)
	if err != nil {
		return nil, nil, fmt.Errorf("parse got fragment: %w", err)
	}

	return want, got, nil
}

// ParseAndNormalizeFragment parses an HTML fragment and returns a normalized
// canonical tree. Formatting-only whitespace nodes are dropped in safe contexts.
func ParseAndNormalizeFragment(s string) ([]Node, error) {
	// Use a benign block container so fragment parsing behaves predictably.
	ctx := &xhtml.Node{
		Type:     xhtml.ElementNode,
		Data:     "div",
		DataAtom: atom.Div,
	}

	parsed, err := xhtml.ParseFragment(strings.NewReader(s), ctx)
	if err != nil {
		return nil, err
	}

	out := make([]Node, 0, len(parsed))
	for _, n := range parsed {
		canon, ok := normalizeNode(n, nil)
		if !ok {
			continue
		}
		out = append(out, canon)
	}

	return out, nil
}

func normalizeNode(n *xhtml.Node, ancestors []string) (Node, bool) {
	switch n.Type {
	case xhtml.ElementNode:
		name := n.Data

		children := make([]Node, 0)
		nextAncestors := append(append([]string(nil), ancestors...), name)

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			canon, ok := normalizeNode(child, nextAncestors)
			if !ok {
				continue
			}
			children = append(children, canon)
		}

		return Node{
			Kind:     KindElement,
			Data:     name,
			Attrs:    normalizeAttrs(n.Attr),
			Children: children,
		}, true

	case xhtml.TextNode:
		// Drop whitespace-only text nodes in contexts where they are usually
		// serialization/pretty-printing noise, but preserve them in whitespace-
		// significant contexts like <pre> and <code>.
		if isAllHTMLSpace(n.Data) && !preserveWhitespaceText(ancestors) {
			return Node{}, false
		}

		return Node{
			Kind: KindText,
			Data: n.Data,
		}, true

	case xhtml.CommentNode:
		return Node{
			Kind: KindComment,
			Data: n.Data,
		}, true

	default:
		// Ignore doctype, document wrapper, etc. for fragment comparison.
		return Node{}, false
	}
}

func normalizeAttrs(attrs []xhtml.Attribute) []Attr {
	if len(attrs) == 0 {
		return nil
	}

	out := make([]Attr, 0, len(attrs))
	for _, a := range attrs {
		key := a.Key
		if a.Namespace != "" {
			key = a.Namespace + ":" + a.Key
		}

		out = append(out, Attr{
			Key: key,
			Val: a.Val,
		})
	}

	// Small stable sort for deterministic comparison output.
	for i := 0; i < len(out)-1; i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Key < out[i].Key || (out[j].Key == out[i].Key && out[j].Val < out[i].Val) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}

	return out
}

func preserveWhitespaceText(ancestors []string) bool {
	if len(ancestors) == 0 {
		return false
	}

	// Preserve whitespace-only text in contexts where it is meaningful.
	for i := len(ancestors) - 1; i >= 0; i-- {
		switch ancestors[i] {
		case "pre", "code", "textarea":
			return true
		}
	}

	return false
}

func isAllHTMLSpace(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !isHTMLSpace(r) {
			return false
		}
	}
	return true
}

func isHTMLSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	default:
		return false
	}
}

func writeNode(w io.Writer, n Node, depth int) {
	indent := strings.Repeat("  ", depth)

	switch n.Kind {
	case KindElement:
		fmt.Fprintf(w, "%s<%s", indent, n.Data)
		for _, a := range n.Attrs {
			fmt.Fprintf(w, ` %s=%q`, a.Key, a.Val)
		}
		fmt.Fprintln(w, ">")
		for _, c := range n.Children {
			writeNode(w, c, depth+1)
		}
		fmt.Fprintf(w, "%s</%s>\n", indent, n.Data)

	case KindText:
		quoted := quoteVisibleWhitespace(n.Data)
		fmt.Fprintf(w, "%s#text(%s)\n", indent, quoted)

	case KindComment:
		fmt.Fprintf(w, "%s#comment(%q)\n", indent, n.Data)
	}
}

func quoteVisibleWhitespace(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if unicode.IsPrint(r) {
				b.WriteRune(r)
			} else {
				fmt.Fprintf(&b, `\u%04X`, r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}
