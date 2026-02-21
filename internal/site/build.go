package site

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/content"
	"github.com/spcameron/seanpatrickcameron.com/templates"
)

type BuildContext struct {
	OutDir string
	Posts  []content.Post
}

func BuildSite(out string) ([]string, error) {
	out = filepath.Clean(out)

	if err := os.MkdirAll(out, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir: %s: %w", out, err)
	}

	candidates, err := content.DiscoverPosts("content")
	if err != nil {
		return nil, err
	}

	posts, err := content.ParsePosts(candidates)
	if err != nil {
		return nil, err
	}
	if err := validatePosts(posts); err != nil {
		return nil, err
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].FrontMatter.Date.After(posts[j].FrontMatter.Date)
	})

	ctx := BuildContext{
		OutDir: out,
		Posts:  posts,
	}

	var written []string

	if w, err := buildHome(ctx); err != nil {
		return nil, err
	} else {
		written = append(written, w...)
	}

	if w, err := buildBlogIndex(ctx); err != nil {
		return nil, err
	} else {
		written = append(written, w...)
	}

	if w, err := buildBlogPosts(ctx); err != nil {
		return nil, err
	} else {
		written = append(written, w...)
	}

	return written, nil

}

func buildHome(ctx BuildContext) ([]string, error) {
	path := filepath.Join(ctx.OutDir, "index.html")
	if err := writeRendered(path, func(w io.Writer) error {
		return templates.Home().Render(context.Background(), w)
	}); err != nil {
		return nil, err
	}

	return []string{path}, nil
}

func buildBlogIndex(ctx BuildContext) ([]string, error) {
	path := blogIndexPath(ctx.OutDir)
	if err := writeRendered(path, func(w io.Writer) error {
		return templates.BlogIndex(ctx.Posts).Render(context.Background(), w)
	}); err != nil {
		return nil, err
	}

	return []string{path}, nil
}

func buildBlogPosts(ctx BuildContext) ([]string, error) {
	var written []string
	for _, p := range ctx.Posts {
		path := blogPostPath(ctx.OutDir, p.FrontMatter.Slug)
		if err := writeRendered(path, func(w io.Writer) error {
			return templates.BlogPost(p).Render(context.Background(), w)
		}); err != nil {
			return nil, err
		}

		written = append(written, path)

		srcMedia := filepath.Join(p.SourceDir, "media")
		dstMedia := filepath.Join(ctx.OutDir, "blog", p.FrontMatter.Slug, "media")

		copied, err := copyDirIfExists(srcMedia, dstMedia)
		if err != nil {
			return nil, err
		}

		written = append(written, copied...)
	}

	return written, nil
}

func blogIndexPath(out string) string {
	return filepath.Join(out, "blog", "index.html")
}

func blogPostPath(out, slug string) string {
	return filepath.Join(out, "blog", slug, "index.html")
}

func writeRendered(path string, render func(io.Writer) error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir: %s: %w", filepath.Dir(path), err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create: %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	if err := render(f); err != nil {
		return fmt.Errorf("render: %s: %w", path, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close: %s: %w", path, err)
	}

	return nil
}

func validatePosts(posts []content.Post) error {
	seen := make(map[string]struct{}, len(posts))
	for _, p := range posts {
		s := strings.TrimSpace(p.FrontMatter.Slug)
		if s == "" {
			return content.ErrMissingSlug
		}
		if _, ok := seen[s]; ok {
			return fmt.Errorf("duplicate slug: %q", s)
		}
		seen[s] = struct{}{}
	}

	return nil
}

func copyDirIfExists(srcDir, dstDir string) ([]string, error) {
	info, err := os.Stat(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat: %s: %w", srcDir, err)
	}
	if !info.IsDir() {
		return nil, nil
	}

	var written []string

	err = filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk %s: %w", path, err)
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("rel: %s: %w", path, err)
		}
		dstPath := filepath.Join(dstDir, rel)

		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return fmt.Errorf("mkdir: %s: %w", filepath.Dir(dstPath), err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read: %s: %w", path, err)
		}
		if err := os.WriteFile(dstPath, data, 0o644); err != nil {
			return fmt.Errorf("write: %s: %w", dstPath, err)
		}

		written = append(written, dstPath)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(written)
	return written, nil
}
