package html

import "io"

type Raw struct {
	Value string
}

func (Raw) isNode() {}

func (r Raw) Write(w io.Writer) error {
	_, err := io.WriteString(w, r.Value)
	return err
}

func RawNode(s string) Raw {
	return Raw{Value: s}
}
