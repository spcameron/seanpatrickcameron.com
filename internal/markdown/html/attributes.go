package html

import (
	"html"
	"io"
	"slices"
)

type Attributes map[string]string

func (a Attributes) SortedKeys() []string {
	keys := make([]string, len(a))

	var i int
	for k := range a {
		keys[i] = k
		i++
	}

	slices.Sort(keys)
	return keys
}

func (a Attributes) Write(w io.Writer) error {
	keys := a.SortedKeys()

	for _, k := range keys {
		v := html.EscapeString(a[k])

		if _, err := io.WriteString(w, " "); err != nil {
			return err
		}
		if _, err := io.WriteString(w, k); err != nil {
			return err
		}
		if _, err := io.WriteString(w, `="`); err != nil {
			return err
		}
		if _, err := io.WriteString(w, v); err != nil {
			return err
		}
		if _, err := io.WriteString(w, `"`); err != nil {
			return err
		}
	}

	return nil
}
