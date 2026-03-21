package inline

import (
	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

func Scan(src *source.Source, span source.ByteSpan) ([]Token, error) {
	input := src.Slice(span)
	scanner := NewScanner(input, span.Start)

	tokens := []Token{}
	for {
		// repeatedly call Next to emit tokens
		token, ok := scanner.Next()
		if !ok {
			// if EOF, append TokenEOF and break
			anchor := source.ByteSpan{
				Start: scanner.Base + source.BytePos(scanner.Position),
				End:   scanner.Base + source.BytePos(scanner.Position),
			}

			tokens = append(tokens, Token{
				Span: anchor,
				Kind: TokenEOF,
			})

			break
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

type Scanner struct {
	Input    string
	Position int
	Base     source.BytePos
}

func NewScanner(input string, base source.BytePos) *Scanner {
	return &Scanner{
		Input:    input,
		Position: 0,
		Base:     base,
	}
}

func (s *Scanner) EOF() bool {
	return s.Position >= len(s.Input)
}

func (s *Scanner) Current() (byte, bool) {
	if s.EOF() {
		return 0, false
	}

	return s.Input[s.Position], true
}

func (s *Scanner) Peek() (byte, bool) {
	next := s.Position + 1
	if next >= len(s.Input) {
		return 0, false
	}

	return s.Input[next], true
}

func (s *Scanner) Next() (Token, bool) {
	if s.EOF() {
		return Token{}, false
	}

	kind, ok := s.CurrentKind()
	if ok {
		return s.token(kind), true
	}

	start := s.Position
	for !s.EOF() {
		if _, ok := s.CurrentKind(); ok {
			break
		}
		s.Position++
	}
	end := s.Position

	if start == end {
		panic("inline scanner made no progress")
	}

	return Token{
		Span: s.span(start, end),
		Kind: TokenText,
	}, true
}

func (s *Scanner) CurrentKind() (TokenKind, bool) {
	b, ok := s.Current()
	if !ok {
		return 0, false
	}

	switch b {
	case '*':
		return TokenStarDelimiter, true

	case '_':
		return TokenUnderscoreDelimiter, true

	case '`':
		return TokenBacktick, true

	case '[':
		return TokenOpenBracket, true

	case ']':
		return TokenCloseBracket, true

	case '(':
		return TokenOpenParen, true

	case ')':
		return TokenCloseParen, true

	case '<':
		return TokenOpenAngle, true

	case '>':
		return TokenCloseAngle, true

	case '!':
		if next, ok := s.Peek(); ok && next == '[' {
			return TokenImageOpenBracket, true
		}
		return TokenBang, true

	case '\n':
		panic("illegal newline character encountered during inline parsing")

	default:
		return 0, false
	}

}

func (s *Scanner) token(kind TokenKind) Token {
	start := s.Position
	s.Position++
	if kind == TokenImageOpenBracket {
		s.Position++
	}

	return Token{
		Span: s.span(start, s.Position),
		Kind: kind,
	}
}

func (s *Scanner) span(start, end int) source.ByteSpan {
	return source.ByteSpan{
		Start: s.Base + source.BytePos(start),
		End:   s.Base + source.BytePos(end),
	}
}
