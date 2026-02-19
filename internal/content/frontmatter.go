package content

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

var (
	ErrMissingTitle       = errors.New("frontmatter is missing title")
	ErrMissingSlug        = errors.New("frontmatter is missing slug")
	ErrMissingDate        = errors.New("frontmatter is missing date")
	ErrInvalidDate        = errors.New("frontmatter contains an invalid date format")
	ErrInvalidFrontMatter = errors.New("frontmatter is malformed")
)

type FrontMatter struct {
	Title string
	Slug  string
	Date  time.Time
}

func DecodeFrontMatter(data []byte) (FrontMatter, error) {
	var raw struct {
		Title string `yaml:"title"`
		Slug  string `yaml:"slug"`
		Date  string `yaml:"date"`
	}

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	if err := dec.Decode(&raw); err != nil {
		return FrontMatter{}, fmt.Errorf("%w: %w", ErrInvalidFrontMatter, err)
	}

	if strings.TrimSpace(raw.Title) == "" {
		return FrontMatter{}, ErrMissingTitle
	}
	if strings.TrimSpace(raw.Slug) == "" {
		return FrontMatter{}, ErrMissingSlug
	}
	if strings.TrimSpace(raw.Date) == "" {
		return FrontMatter{}, ErrMissingDate
	}

	t, err := time.Parse("2006-01-02", raw.Date)
	if err != nil {
		return FrontMatter{}, fmt.Errorf("%w (expected YYYY-MM-DD): %q", ErrInvalidDate, raw.Date)
	}

	return FrontMatter{
		Title: raw.Title,
		Slug:  raw.Slug,
		Date:  t,
	}, nil
}
