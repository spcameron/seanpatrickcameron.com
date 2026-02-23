package testkit

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"

func ASTDoc(blocks ...ast.Block) ast.Document {
	return ast.Document{
		Blocks: blocks,
	}
}

func ASTPara(inlines ...ast.Inline) ast.Paragraph {
	return ast.Paragraph{
		Inlines: inlines,
	}
}

func ASTText(value string) ast.Text {
	return ast.Text{
		Value: value,
	}
}
