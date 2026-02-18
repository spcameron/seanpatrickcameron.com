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
	FrontMatter FrontMatter
	BodyMD      string
}

type PostSummary struct{}

func ParsePosts(paths []string) ([]Post, error) {
	var posts []Post
	for _, s := range paths {
		p, err := ParsePost(s)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

func ParsePost(path string) (Post, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Post{}, err
	}

	fmBytes, mdBytes, err := SplitPost(data)
	if err != nil {
		return Post{}, err
	}

	fm, err := DecodeFrontMatter(fmBytes)
	if err != nil {
		return Post{}, err
	}

	// TODO: convert markdown
	md := string(mdBytes)

	post := Post{
		FrontMatter: fm,
		BodyMD:      md,
	}

	return post, nil
}

func SplitPost(src []byte) (fmBytes, mdBytes []byte, err error) {
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
			fmBytes = src[openEnd:pos]
			mdBytes = src[nextPos:]
			return fmBytes, mdBytes, nil
		}

		pos = nextPos
	}

	return nil, nil, ErrMissingClosingFence
}
