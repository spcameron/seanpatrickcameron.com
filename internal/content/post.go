package content

import (
	"bytes"
	"fmt"
	"os"
)

var fence = []byte("---")

var (
	ErrEmptyFile                 = fmt.Errorf("empty file")
	ErrMissingOpeningFence       = fmt.Errorf("missing frontmatter opening fence")
	ErrMissingClosingFence       = fmt.Errorf("missing frontmatter closing fence")
	ErrOpeningFenceNotTerminated = fmt.Errorf("opening fence missing terminating newline")
)

type Post struct {
	Meta   FrontMatter
	BodyMD string
}

type PostSummary struct{}

func ReadPostFile(path string) (Post, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Post{}, err
	}

	fm, md, err := SplitPost(data)
	if err != nil {
		return Post{}, err
	}

	_ = fm
	_ = md

	return Post{}, nil
}

func SplitPost(src []byte) (fm, md []byte, err error) {
	if len(src) == 0 {
		return nil, nil, ErrEmptyFile
	}

	// Find first newline (end of opening fence line).
	i := bytes.IndexByte(src, '\n')
	if i == -1 {
		return nil, nil, ErrOpeningFenceNotTerminated
	}

	if !bytes.Equal(src[:i], fence) {
		return nil, nil, ErrMissingOpeningFence
	}

	openEnd := i + 1

	// Scan lines starting at openEnd for a closing fence.
	for pos := openEnd; pos < len(src); {
		rel := bytes.IndexByte(src[pos:], '\n')

		var line []byte
		var nextPos int
		if rel == -1 {
			line = src[pos:]
			nextPos = len(src)
		} else {
			j := pos + rel
			line = src[pos:j]
			nextPos = j + 1
		}

		if bytes.Equal(line, fence) {
			fm = src[openEnd:pos]
			md = src[nextPos:]
			return fm, md, nil
		}

		pos = nextPos
	}

	return nil, nil, ErrMissingClosingFence
}
