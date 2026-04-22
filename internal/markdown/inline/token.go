package inline

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

// TokenKind identifies the kind of inline token produced by scanning.
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

// Token represents a scanned inline token with its source span.
type Token struct {
	Span source.ByteSpan
	Kind TokenKind
}

// EscapeBehavior describes how a token should be handled when preceded
// by a backslash escape.
type EscapeBehavior int

const (
	_ EscapeBehavior = iota
	EscapeNone
	EscapeLiteralize
	EscapeLiteralizeLeadingByte
	EscapeDecompose
)

// classifyEscapeTarget determines how tok should be treated when escaped.
func classifyEscapeTarget(tok Token, src *source.Source) EscapeBehavior {
	switch tok.Kind {
	case TokenEOF:
		return EscapeNone

	case TokenImageOpenBracket:
		return EscapeDecompose

	case TokenText:
		if tok.Span.Start >= tok.Span.End {
			return EscapeNone
		}

		s := src.Slice(tok.Span)
		if source.IsEscapablePunctuation(s[0]) {
			return EscapeLiteralizeLeadingByte
		}

		return EscapeNone

	default:
		return EscapeLiteralize
	}
}
