package ast

var (
	_ Block = Paragraph{}
)

var (
	_ Inline = Text{}
	_ Inline = HardBreak{}
	_ Inline = SoftBreak{}
)
