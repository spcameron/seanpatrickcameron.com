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

	kind, width, ok := s.Special()
	if ok {
		return s.token(kind, width), true
	}

	start := s.Position
	for !s.EOF() {
		if _, _, ok := s.Special(); ok {
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

func (s *Scanner) Special() (TokenKind, int, bool) {
	b, ok := s.Current()
	if !ok {
		return 0, 0, false
	}

	switch b {
	case '*':
		return TokenStarDelimiter, s.runLength(b), true

	case '_':
		return TokenUnderscoreDelimiter, s.runLength(b), true

	case '`':
		return TokenBacktick, s.runLength(b), true

	case '[':
		return TokenOpenBracket, 1, true

	case ']':
		return TokenCloseBracket, 1, true

	case '(':
		return TokenOpenParen, 1, true

	case ')':
		return TokenCloseParen, 1, true

	case '<':
		return TokenOpenAngle, 1, true

	case '>':
		return TokenCloseAngle, 1, true

	case '!':
		if next, ok := s.Peek(); ok && next == '[' {
			return TokenImageOpenBracket, 2, true
		}
		return TokenBang, 1, true

	case '\\':
		return TokenBackslash, 1, true

	case '\n':
		panic("illegal newline character encountered during inline parsing")

	default:
		return 0, 0, false
	}

}

func (s *Scanner) token(kind TokenKind, width int) Token {
	start := s.Position
	s.Position += width

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

func (s *Scanner) runLength(b byte) int {
	pos := s.Position
	for pos < len(s.Input) && s.Input[pos] == b {
		pos++
	}

	return pos - s.Position
}
