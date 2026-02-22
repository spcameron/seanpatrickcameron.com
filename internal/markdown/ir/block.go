package ir

type Block interface {
	isBlock()
}

type Document struct {
	Blocks []Block
}

func (Document) isBlock() {}

type Paragraph struct {
	Text string
	Span LineSpan
}

func (Paragraph) isBlock() {}
