// Package toolcli contains repository discovery shared by the standalone tools.
package toolcli

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindRoot validates an explicitly requested repository root, or searches from
// start upward. Relative requested paths are resolved against start.
func FindRoot(start, requested string) (string, error) {
	if requested != "" {
		if !filepath.IsAbs(requested) {
			requested = filepath.Join(start, requested)
		}
		root, err := filepath.Abs(requested)
		if err != nil {
			return "", fmt.Errorf("resolve --root: %w", err)
		}
		if err := checkRoot(root); err != nil {
			return "", fmt.Errorf("invalid repository root %q: %w", root, err)
		}
		return root, nil
	}
	root, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve current directory: %w", err)
	}
	for {
		if checkRoot(root) == nil {
			return root, nil
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", fmt.Errorf("repository root not found above %q; use --root PATH (requires data/publications.json and data/settings.json)", start)
		}
		root = parent
	}
}

func checkRoot(root string) error {
	for _, name := range []string{"publications.json", "settings.json"} {
		path := filepath.Join(root, "data", name)
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s must be a file", path)
		}
	}
	return nil
}
