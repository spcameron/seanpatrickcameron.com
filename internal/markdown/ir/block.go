package ir

type Block interface {
	isBlock()
}

type Document struct {
	Blocks []Block
}

func (Document) isBlock() {}

type Header struct {
	Level int
	Text  string
	Span  LineSpan
}

func (Header) isBlock() {}

type Paragraph struct {
	Text string
	Span LineSpan
}

func (Paragraph) isBlock() {}
