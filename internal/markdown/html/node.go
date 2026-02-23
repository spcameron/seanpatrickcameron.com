package html

import (
	"io"
)

type Node interface {
	isNode()
	Write(io.Writer) error
}
