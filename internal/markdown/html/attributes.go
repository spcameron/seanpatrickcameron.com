package html

import (
	"html"
	"io"
	"slices"
)

// Attributes represents a set of HTML element attributes.
type Attributes map[string]string

// SortedKeys returns the attribute keys in deterministic order.
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

// Write serializes attributes to w in key-sorted order.
//
// Values are HTML-escaped before being written.
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
