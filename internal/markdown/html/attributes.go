package html

import (
	"fmt"
	"html"
	"slices"
	"strings"
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

func (a Attributes) Render() string {
	keys := a.SortedKeys()

	var sb strings.Builder
	for _, k := range keys {
		v := html.EscapeString(a[k])
		fmt.Fprintf(&sb, " %s=\"%s\"", k, v)
	}

	return sb.String()
}

type HasAttrs interface {
	Attrs() Attributes
}
