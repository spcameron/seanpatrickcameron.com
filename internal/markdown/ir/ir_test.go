package ir

var (
	_ Block = BlockQuote{}
	_ Block = Header{}
	_ Block = ThematicBreak{}
	_ Block = OrderedList{}
	_ Block = UnorderedList{}
	_ Block = ListItem{}
	_ Block = IndentedCodeBlock{}
	_ Block = FencedCodeBlock{}
	_ Block = HTMLBlock{}
	_ Block = Paragraph{}
)
