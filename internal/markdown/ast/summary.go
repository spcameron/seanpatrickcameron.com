package ast

import (
	"fmt"
	"strings"
)

func summarizeBlocks(blocks []Block) string {
	if len(blocks) == 0 {
		return "[]"
	}

	var b strings.Builder
	b.WriteString("[")

	for i, blk := range blocks {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(summarizeBlock(blk))
	}

	b.WriteString("]")
	return b.String()
}

func summarizeListItems(items []ListItem) string {
	if len(items) == 0 {
		return "[]"
	}

	var b strings.Builder
	b.WriteString("[")

	for i, item := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(item.String())
	}

	b.WriteString("]")
	return b.String()
}

func summarizeInlines(inlines []Inline) string {
	if len(inlines) == 0 {
		return "[]"
	}

	var b strings.Builder
	b.WriteString("[")

	for i, inl := range inlines {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(summarizeInline(inl))
	}

	b.WriteString("]")
	return b.String()
}

func summarizeBlock(b Block) string {
	switch v := b.(type) {
	case BlockQuote:
		return v.String()
	case Header:
		return v.String()
	case ThematicBreak:
		return v.String()
	case OrderedList:
		return v.String()
	case UnorderedList:
		return v.String()
	case ListItem:
		return v.String()
	case CodeBlock:
		return v.String()
	case HTMLBlock:
		return v.String()
	case Paragraph:
		return v.String()
	default:
		return fmt.Sprintf("%T", v)
	}
}

func summarizeInline(in Inline) string {
	switch v := in.(type) {
	case CodeSpan:
		return v.String()
	case Link:
		return v.String()
	case Image:
		return v.String()
	case Emph:
		return v.String()
	case Strong:
		return v.String()
	case Text:
		return v.String()
	case RawText:
		return v.String()
	case HardBreak:
		return v.String()
	case SoftBreak:
		return v.String()
	case Newline:
		return v.String()
	default:
		return fmt.Sprintf("%T", v)
	}
}
