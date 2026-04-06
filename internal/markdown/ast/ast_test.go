package ast

var (
	_ Block = Paragraph{}
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
