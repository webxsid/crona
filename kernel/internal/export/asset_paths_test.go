package export

import (
	"path/filepath"
	"testing"

	"crona/kernel/internal/runtime"
	sharedtypes "crona/shared/types"
)

func TestDefaultAssetSourceFallsBackToEmbeddedAssets(t *testing.T) {
	base := t.TempDir()
	paths := runtime.Paths{
		BundledAssetsDir: filepath.Join(base, "assets", "bundled"),
		UserAssetsDir:    filepath.Join(base, "assets", "user"),
	}
	descriptor, ok := findAssetDescriptor(
		sharedtypes.ExportReportKindDaily,
		sharedtypes.ExportAssetKindTemplateMarkdown,
	)
	if !ok {
		t.Fatal("daily template descriptor not found")
	}

	body, bundledPath, err := defaultAssetSource(paths, descriptor)
	if err != nil {
		t.Fatalf("default asset source: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected embedded asset body")
	}
	if bundledPath == "" {
		t.Fatal("expected bundled path")
	}
	if filepath.Clean(bundledPath) == "" {
		t.Fatal("expected a resolved bundled path")
	}
}
