package commands

import (
	"os"
	"path/filepath"
	"testing"

	"crona/shared/config"
)

func TestExpectedChecksum(t *testing.T) {
	checksums := []byte("abc123  install-crona-tui.sh\nzzz999  checksums.txt\n")
	if got := expectedChecksum("install-crona-tui.sh", checksums); got != "abc123" {
		t.Fatalf("expected checksum abc123, got %q", got)
	}
	if got := expectedChecksum("missing", checksums); got != "" {
		t.Fatalf("expected missing checksum to be empty, got %q", got)
	}
}

func TestNormalizedInstalledExecutableRequiresInstallDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CRONA_INSTALL_DIR", dir)
	binaryName := config.TUIBinaryName()
	path := filepath.Join(dir, binaryName)
	if got := normalizedInstalledExecutable(path, binaryName); got != path {
		t.Fatalf("expected install dir executable to be accepted, got %q", got)
	}

	otherPath := filepath.Join(t.TempDir(), binaryName)
	if got := normalizedInstalledExecutable(otherPath, binaryName); got != "" {
		t.Fatalf("expected non-install-dir executable to be rejected, got %q", got)
	}
}

func TestResolveRelaunchPathUsesInstallDirBinary(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CRONA_INSTALL_DIR", dir)
	binaryName := config.TUIBinaryName()
	target := filepath.Join(dir, binaryName)
	if err := os.WriteFile(target, []byte("stub"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := resolveRelaunchPath(filepath.Join(t.TempDir(), binaryName))
	if err != nil {
		t.Fatalf("resolveRelaunchPath: %v", err)
	}
	if got != target {
		t.Fatalf("expected relaunch path %q, got %q", target, got)
	}
}
