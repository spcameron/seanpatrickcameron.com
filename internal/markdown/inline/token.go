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
	EscapeLiteralizeLeadingByte
	EscapeDecompose
)

func classifyEscapeTarget(tok Token, src *source.Source) EscapeBehavior {
	switch tok.Kind {
	case TokenText:
		if tok.Span.Start >= tok.Span.End {
			return EscapeNone
		}

		s := src.Slice(tok.Span)
		b := s[0]

		if isEscapablePunctuation(b) {
			return EscapeLiteralizeLeadingByte
		}

		return EscapeNone

	case TokenEOF:
		return EscapeNone

	case TokenImageOpenBracket:
		return EscapeDecompose

	default:
		return EscapeLiteralize

	}
}

func isEscapablePunctuation(b byte) bool {
	switch b {
	case '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
		':', ';', '<', '=', '>', '?', '@',
		'[', '\\', ']', '^', '_', '`',
		'{', '|', '}', '~':
		return true
	default:
		return false
	}
}
