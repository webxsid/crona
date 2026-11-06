package helpers

import (
	"os"
	"path/filepath"
	"strings"
)

// ShortPath shows a path with the home directory as ~ and, past limit cells, its middle
// folded to …, keeping the root and the last two parts: ~/…/Crona Dev/reports.
func ShortPath(path string, limit int) string {
	path = strings.TrimSpace(path)
	if home, err := os.UserHomeDir(); err == nil && home != "" && strings.HasPrefix(path, home) {
		path = "~" + strings.TrimPrefix(path, home)
	}
	if limit <= 0 || len([]rune(path)) <= limit {
		return path
	}
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) > 3 {
		short := parts[0] + "/…/" + strings.Join(parts[len(parts)-2:], "/")
		if len([]rune(short)) <= limit {
			return short
		}
		path = "…/" + parts[len(parts)-1]
	}
	if runes := []rune(path); len(runes) > limit {
		return "…" + string(runes[len(runes)-limit+1:])
	}
	return path
}
