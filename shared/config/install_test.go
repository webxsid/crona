package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultInstallDirForOS(t *testing.T) {
	home := "/home/alice"

	tests := []struct {
		name         string
		goos         string
		localAppData string
		want         string
	}{
		{
			name: "darwin",
			goos: "darwin",
			want: filepath.Join(home, ".local", "bin"),
		},
		{
			name: "linux",
			goos: "linux",
			want: filepath.Join(home, ".local", "bin"),
		},
		{
			name:         "windows with local app data",
			goos:         "windows",
			localAppData: `C:\Users\alice\AppData\Local`,
			want:         filepath.Join(`C:\Users\alice\AppData\Local`, "Programs", "Crona", "bin"),
		},
		{
			name: "windows fallback",
			goos: "windows",
			want: filepath.Join(home, "AppData", "Local", "Programs", "Crona", "bin"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultInstallDirForOS(tt.goos, home, tt.localAppData); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestInstallerAssetNameForGOOS(t *testing.T) {
	if got := InstallerAssetNameForGOOS("windows"); got != "install-crona-tui.ps1" {
		t.Fatalf("expected windows installer asset, got %q", got)
	}
	if got := InstallerAssetNameForGOOS("darwin"); got != "install-crona-tui.sh" {
		t.Fatalf("expected unix installer asset, got %q", got)
	}
}
