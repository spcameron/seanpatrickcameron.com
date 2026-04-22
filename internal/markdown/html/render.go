package html

import (
	"io"
	"strings"
)

// Writable is implemented by values that can serialize themselves to an io.Writer.
type Writable interface {
	Write(io.Writer) error
}

// Render writes w to a string and returns the result.
func Render(w Writable) (string, error) {
	var sb strings.Builder
	if err := w.Write(&sb); err != nil {
		return "", err
	}
	return sb.String(), nil
}
