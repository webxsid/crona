package assets

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed export alerts
var embeddedFS embed.FS

var embeddedRoots = []string{"export", "alerts"}

func Read(rel string) ([]byte, bool, error) {
	normalized, err := normalize(rel)
	if err != nil {
		return nil, false, err
	}
	body, err := embeddedFS.ReadFile(normalized)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return body, true, nil
}

func Ensure(root, rel string) error {
	normalized, err := normalize(rel)
	if err != nil {
		return err
	}
	body, ok, err := Read(normalized)
	if err != nil {
		return err
	}
	if !ok {
		return fs.ErrNotExist
	}
	return writeIfMissing(root, normalized, body)
}

func EnsureAll(root string) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return fmt.Errorf("asset root is empty")
	}
	for _, top := range embeddedRoots {
		if err := fs.WalkDir(embeddedFS, top, func(name string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			return Ensure(root, name)
		}); err != nil {
			return err
		}
	}
	return nil
}

func normalize(rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return "", fmt.Errorf("asset path is empty")
	}
	rel = path.Clean(strings.TrimPrefix(rel, "/"))
	if rel == "." || rel == "/" || strings.HasPrefix(rel, "../") || rel == ".." {
		return "", fmt.Errorf("invalid asset path: %s", rel)
	}
	return rel, nil
}

func writeIfMissing(root, rel string, body []byte) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return fmt.Errorf("asset root is empty")
	}
	dest := filepath.Join(root, filepath.FromSlash(rel))
	info, err := os.Stat(dest)
	switch {
	case err == nil:
		if info.IsDir() {
			return fmt.Errorf("asset path exists as a directory: %s", dest)
		}
		return nil
	case !errors.Is(err, os.ErrNotExist):
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return err
	}
	return os.WriteFile(dest, body, 0o600)
}
