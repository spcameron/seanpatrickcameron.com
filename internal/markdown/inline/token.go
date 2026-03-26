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
	TokenBackslash
	TokenEOF
)

type Token struct {
	Span source.ByteSpan
	Kind TokenKind
}

type EscapeBehavior int

const (
	_ EscapeBehavior = iota
	EscapeNone
	EscapeLiteralize
	EscapeDecompose
)

func classifyEscapeTarget(tok Token) EscapeBehavior {
	switch tok.Kind {
	case TokenText, TokenEOF:
		return EscapeNone

	case TokenImageOpenBracket:
		return EscapeDecompose

	default:
		return EscapeLiteralize

	}
}
