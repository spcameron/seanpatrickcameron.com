package site

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spcameron/seanpatrickcameron.com/templates"
)

func BuildSite(out string) ([]string, error) {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir: %s: %w", out, err)
	}

	indexPath := filepath.Join(out, "index.html")
	f, err := os.Create(indexPath)
	if err != nil {
		return nil, fmt.Errorf("create: %s: %w", indexPath, err)
	}

	if err := templates.Home().Render(context.Background(), f); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("render %s: %w", indexPath, err)
	}

	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("close %s: %w", indexPath, err)
	}

	return []string{indexPath}, nil
}
