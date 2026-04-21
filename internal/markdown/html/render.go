package html

import (
	"io"
	"strings"
)

type Writable interface {
	Write(io.Writer) error
}

func Render(w Writable) (string, error) {
	var sb strings.Builder
	if err := w.Write(&sb); err != nil {
		return "", err
	}
	return sb.String(), nil
}
