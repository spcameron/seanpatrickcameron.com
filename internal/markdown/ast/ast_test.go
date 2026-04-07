package ast

var (
	_ Block = Paragraph{}
	_ Block = BlockQuote{}
	_ Block = Header{}
	_ Block = ThematicBreak{}
	_ Block = OrderedList{}
	_ Block = UnorderedList{}
	_ Block = ListItem{}
	_ Block = CodeBlock{}
	_ Block = HTMLBlock{}
)

var (
	_ Inline = CodeSpan{}
	_ Inline = Link{}
	_ Inline = Image{}
	_ Inline = Emph{}
	_ Inline = Strong{}
	_ Inline = Text{}
	_ Inline = RawText{}
	_ Inline = HardBreak{}
	_ Inline = SoftBreak{}
	_ Inline = Newline{}
)
