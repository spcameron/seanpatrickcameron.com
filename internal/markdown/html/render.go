package html

import "strings"

func Render(node Node) (string, error) {
	var sb strings.Builder
	err := node.Write(&sb)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}
