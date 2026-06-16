package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultBackupBaseDirForModeOnOS(t *testing.T) {
	home := "/home/alice"

	tests := []struct {
		name         string
		mode         string
		goos         string
		xdgDataHome  string
		localAppData string
		want         string
	}{
		{
			name: "darwin prod",
			mode: ModeProd,
			goos: "darwin",
			want: filepath.Join(home, "Library", "Application Support", "Crona Backups"),
		},
		{
			name: "darwin dev",
			mode: ModeDev,
			goos: "darwin",
			want: filepath.Join(home, "Library", "Application Support", "Crona Dev Backups"),
		},
		{
			name:        "linux prod xdg",
			mode:        ModeProd,
			goos:        "linux",
			xdgDataHome: filepath.Join(home, ".data"),
			want:        filepath.Join(home, ".data", "crona-backups"),
		},
		{
			name: "linux dev default",
			mode: ModeDev,
			goos: "linux",
			want: filepath.Join(home, ".local", "share", "crona-dev-backups"),
		},
		{
			name:         "windows prod",
			mode:         ModeProd,
			goos:         "windows",
			localAppData: `C:\Users\alice\AppData\Local`,
			want:         filepath.Join(`C:\Users\alice\AppData\Local`, "Crona Backups"),
		},
		{
			name: "windows dev fallback",
			mode: ModeDev,
			goos: "windows",
			want: filepath.Join(home, "AppData", "Local", "Crona Dev Backups"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultBackupBaseDirForModeOnOS(
				tt.mode,
				tt.goos,
				home,
				tt.xdgDataHome,
				tt.localAppData,
			)
			if err != nil {
				t.Fatalf("DefaultBackupBaseDirForModeOnOS: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
