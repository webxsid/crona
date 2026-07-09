package runtime

import (
	"os"
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

func TestInstallSourceRoundTripScoop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "install.json")

	if err := WriteInstallSourceDetails(
		path,
		sharedtypes.InstallSourceScoop,
		"",
		sharedtypes.UpdateChannelBeta,
	); err != nil {
		t.Fatalf("WriteInstallSource: %v", err)
	}

	got, err := LoadInstallSource(path)
	if err != nil {
		t.Fatalf("LoadInstallSource: %v", err)
	}
	if got != sharedtypes.InstallSourceScoop {
		t.Fatalf("expected scoop install source, got %s", got)
	}
	file, err := LoadInstallSourceFile(path)
	if err != nil {
		t.Fatalf("LoadInstallSourceFile: %v", err)
	}
	if file.ReleaseChannel != sharedtypes.UpdateChannelBeta {
		t.Fatalf(
			"expected release channel to round-trip, got %q",
			file.ReleaseChannel,
		)
	}
}

func TestInstallSourceWriteNoopsForUnknownSourceAndEmptyPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "install.json")

	if err := WriteInstallSource("", sharedtypes.InstallSourceBrew); err != nil {
		t.Fatalf("WriteInstallSource with empty path: %v", err)
	}
	if err := WriteInstallSourceDetails(path, sharedtypes.InstallSourceUnknown, "", ""); err != nil {
		t.Fatalf("WriteInstallSourceDetails with unknown source: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected no install source file to be written, got err=%v", err)
	}
}
