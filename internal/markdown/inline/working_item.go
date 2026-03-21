package inline

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type InlineItem interface {
	isInlineItem()
}

type TextItem struct {
	Span source.ByteSpan
}

func (*TextItem) isInlineItem() {}

type DelimiterItem struct {
	Span      source.ByteSpan
	Delimiter byte
}

func (*DelimiterItem) isInlineItem() {}

type NodeItem struct {
	Span source.ByteSpan
	Node ast.Inline
}

func (*NodeItem) isInlineItem() {}

type ConsumedItem struct {
	Span source.ByteSpan
}

func (*ConsumedItem) isInlineItem() {}

// type TokenKind int
//
// func (tk TokenKind) String() string {
// 	switch tk {
// 	case TokenOpenBracket:
// 		return "open_bracket"
//
// 	case TokenCloseBracket:
// 		return "close_bracket"
//
// 	case TokenOpenParen:
// 		return "open_paren"
//
// 	case TokenCloseParen:
// 		return "close_paren"
//
// 	default:
// 		return fmt.Sprintf("unknown_token_kind(%d)", tk)
// 	}
// }

// type TokenItem struct {
// 	Span source.ByteSpan
// 	Kind TokenKind
// }
//
// func (*TokenItem) isWorkingItem() {}
