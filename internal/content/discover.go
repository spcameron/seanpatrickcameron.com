package content

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func DiscoverPosts(contentRoot string) ([]string, error) {
	contentRoot = filepath.Clean(contentRoot)

	info, err := os.Stat(contentRoot)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", contentRoot, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", contentRoot)
	}

	postsDir := filepath.Join(contentRoot, "posts")

	postsInfo, err := os.Stat(postsDir)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", postsDir, err)
	}
	if !postsInfo.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", postsDir)
	}

	var candidates []string

	walkErr := filepath.WalkDir(postsDir, func(path string, d fs.DirEntry, err error) error {
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

		if strings.EqualFold(name, "index.md") {
			if filepath.Clean(filepath.Dir(path)) == filepath.Clean(postsDir) {
				return nil
			}
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
