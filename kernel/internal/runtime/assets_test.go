package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureBundledAssetsWritesMissingFiles(t *testing.T) {
	base := t.TempDir()
	paths := Paths{
		BundledAssetsDir: filepath.Join(base, "assets", "bundled"),
	}

	if err := EnsureBundledAssets(paths); err != nil {
		t.Fatalf("ensure bundled assets: %v", err)
	}

	for _, path := range []string{
		filepath.Join(paths.BundledAssetsDir, "export", "daily", "report.default.hbs"),
		filepath.Join(paths.BundledAssetsDir, "alerts", "sounds", "soft-bell.mp3"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
	}
}
