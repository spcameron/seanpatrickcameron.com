package inline

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type TokenKind int

const (
	_ TokenKind = iota
	TokenText
	TokenStarDelimiter
	TokenUnderscoreDelimiter
	TokenBacktick
	TokenOpenBracket
	TokenCloseBracket
	TokenOpenParen
	TokenCloseParen
	TokenOpenAngle
	TokenCloseAngle
	TokenBang
	TokenImageOpenBracket
)

type Token struct {
	Span source.ByteSpan
	Kind TokenKind
}
