package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureAllWritesEmbeddedAssets(t *testing.T) {
	root := t.TempDir()

	if err := EnsureAll(root); err != nil {
		t.Fatalf("ensure all: %v", err)
	}

	checks := []string{
		filepath.Join(root, "export", "daily", "report.default.hbs"),
		filepath.Join(root, "alerts", "sounds", "chime.mp3"),
		filepath.Join(root, "alerts", "sounds", "soft-bell.mp3"),
		filepath.Join(root, "alerts", "sounds", "notification-ping.mp3"),
		filepath.Join(root, "alerts", "sounds", "focus-gong.mp3"),
		filepath.Join(root, "alerts", "sounds", "minimal-click.mp3"),
	}
	for _, path := range checks {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if info.IsDir() {
			t.Fatalf("expected file at %s", path)
		}
	}
}

func TestEnsureDoesNotOverwriteExistingFiles(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "export", "daily", "report.default.hbs")
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(target, []byte("sentinel"), 0o600); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if err := Ensure(root, "export/daily/report.default.hbs"); err != nil {
		t.Fatalf("ensure file: %v", err)
	}

	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if strings.TrimSpace(string(body)) != "sentinel" {
		t.Fatalf("expected existing file to remain untouched, got %q", string(body))
	}
}
