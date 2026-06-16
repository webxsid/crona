package notify

import (
	"os"
	"path/filepath"
	"testing"

	runtimepkg "crona/kernel/internal/runtime"
	sharedtypes "crona/shared/types"
)

func TestAlertSoundPathFallsBackToEmbeddedAssets(t *testing.T) {
	base := t.TempDir()
	paths := runtimepkg.Paths{
		BundledAssetsDir: filepath.Join(base, "assets", "bundled"),
	}

	path, err := alertSoundPath(paths, sharedtypes.AlertSoundPresetChime)
	if err != nil {
		t.Fatalf("alert sound path: %v", err)
	}
	if path == "" {
		t.Fatal("expected alert sound path")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected alert sound on disk: %v", err)
	}
}
