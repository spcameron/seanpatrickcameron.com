package reference

import (
	"strings"
	"unicode"
)

func ValidateLabel(s string) bool {
	if len(s) == 0 {
		return false
	}

	seenContent := false
	escaped := false
	charCount := 0

	for _, r := range s {
		charCount++
		if charCount > 999 {
			return false
		}

		if !isLabelWhitespace(r) {
			seenContent = true
		}

		if escaped {
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if r == '[' || r == ']' {
			return false
		}
	}

	return seenContent
}

func NormalizeLabel(s string) string {
	if len(s) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(len(s))

	seenContent := false
	escaped := false
	pendingSpace := false

	for _, r := range s {
		if escaped {
			// this rune is literal content
			if pendingSpace {
				sb.WriteByte(' ')
				pendingSpace = false
			}

			sb.WriteRune(unicode.ToLower(r))
			seenContent = true
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if isLabelWhitespace(r) {
			if seenContent {
				pendingSpace = true
			}
			continue
		}

		// normal content rune
		if pendingSpace {
			sb.WriteByte(' ')
			pendingSpace = false
		}

		sb.WriteRune(unicode.ToLower(r))
		seenContent = true
	}

	return sb.String()
}

func isLabelWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}
