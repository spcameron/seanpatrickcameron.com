package content

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func DiscoverPosts(dir string) ([]string, error) {
	dir = filepath.Clean(dir)

	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", dir)
	}

	var candidates []string

	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk %s: %w", path, err)
		}

		name := d.Name()

		if d.IsDir() {
			if shouldSkipDir(name) {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.Type().IsRegular() {
			return nil
		}

		if strings.EqualFold(filepath.Ext(name), ".md") {
			candidates = append(candidates, path)
		}

		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	sort.Strings(candidates)
	return candidates, nil
}

func shouldSkipDir(name string) bool {
	if name == "" {
		return false
	}

	if strings.HasPrefix(name, ".") {
		return true
	}

	switch name {
	case "node_modules", "vendor", "tmp", "dist", "build":
		return true
	default:
		return false
	}
}
