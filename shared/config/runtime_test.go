package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultRuntimeBaseDirForModeOnOS(t *testing.T) {
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
			want: filepath.Join(home, "Library", "Application Support", "Crona"),
		},
		{
			name: "darwin dev",
			mode: ModeDev,
			goos: "darwin",
			want: filepath.Join(home, "Library", "Application Support", "Crona Dev"),
		},
		{
			name:        "linux prod xdg",
			mode:        ModeProd,
			goos:        "linux",
			xdgDataHome: filepath.Join(home, ".data"),
			want:        filepath.Join(home, ".data", "crona"),
		},
		{
			name: "linux dev default",
			mode: ModeDev,
			goos: "linux",
			want: filepath.Join(home, ".local", "share", "crona-dev"),
		},
		{
			name:         "windows prod",
			mode:         ModeProd,
			goos:         "windows",
			localAppData: `C:\Users\alice\AppData\Local`,
			want:         filepath.Join(`C:\Users\alice\AppData\Local`, "Crona"),
		},
		{
			name: "windows dev fallback",
			mode: ModeDev,
			goos: "windows",
			want: filepath.Join(home, "AppData", "Local", "Crona Dev"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultRuntimeBaseDirForModeOnOS(tt.mode, tt.goos, home, tt.xdgDataHome, tt.localAppData)
			if err != nil {
				t.Fatalf("DefaultRuntimeBaseDirForModeOnOS: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestLegacyRuntimeBaseDirForModeOnOS(t *testing.T) {
	home := "/home/alice"

	if got := LegacyRuntimeBaseDirForModeOnOS(ModeProd, "darwin", home); got != filepath.Join(home, ".crona") {
		t.Fatalf("expected prod legacy dir, got %q", got)
	}
	if got := LegacyRuntimeBaseDirForModeOnOS(ModeDev, "linux", home); got != filepath.Join(home, ".crona-dev") {
		t.Fatalf("expected dev legacy dir, got %q", got)
	}
	if got := LegacyRuntimeBaseDirForModeOnOS(ModeProd, "windows", home); got != filepath.Join(home, "Crona") {
		t.Fatalf("expected windows runtime dir, got %q", got)
	}
}

func TestRuntimeBaseDirUsesOverride(t *testing.T) {
	t.Setenv(EnvVarRuntimeDir, "/tmp/custom-crona")

	got, err := RuntimeBaseDirForMode(ModeProd)
	if err != nil {
		t.Fatalf("RuntimeBaseDirForMode: %v", err)
	}
	if got != "/tmp/custom-crona" {
		t.Fatalf("expected override path, got %q", got)
	}
}
