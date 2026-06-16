package runtime

import (
	"path/filepath"
	"testing"

	sharedtypes "crona/shared/types"
)

func TestInstallSourceRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "install.json")

	if err := WriteInstallSource(path, sharedtypes.InstallSourceBrew); err != nil {
		t.Fatalf("WriteInstallSource: %v", err)
	}

	got, err := LoadInstallSource(path)
	if err != nil {
		t.Fatalf("LoadInstallSource: %v", err)
	}
	if got != sharedtypes.InstallSourceBrew {
		t.Fatalf("expected brew install source, got %s", got)
	}
}

func TestInstallSourceRoundTripWinget(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "install.json")

	if err := WriteInstallSource(path, sharedtypes.InstallSourceWinget); err != nil {
		t.Fatalf("WriteInstallSource: %v", err)
	}

	got, err := LoadInstallSource(path)
	if err != nil {
		t.Fatalf("LoadInstallSource: %v", err)
	}
	if got != sharedtypes.InstallSourceWinget {
		t.Fatalf("expected winget install source, got %s", got)
	}
}
