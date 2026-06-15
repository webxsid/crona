package updatecheck

import (
	"testing"

	sharedtypes "crona/shared/types"
)

func TestSourceFromExecutablePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want sharedtypes.InstallSource
	}{
		{
			name: "homebrew",
			path: "/opt/homebrew/bin/crona",
			want: sharedtypes.InstallSourceBrew,
		},
		{
			name: "go bin",
			path: "/Users/alice/go/bin/crona",
			want: sharedtypes.InstallSourceGo,
		},
		{
			name: "unknown",
			path: "/tmp/crona",
			want: sharedtypes.InstallSourceUnknown,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := sourceFromExecutablePath(tc.path); got != tc.want {
				t.Fatalf("sourceFromExecutablePath(%q) = %s, want %s", tc.path, got, tc.want)
			}
		})
	}
}

func TestUpdateCommandForStatus(t *testing.T) {
	tests := []struct {
		name   string
		status sharedtypes.UpdateStatus
		want   string
	}{
		{
			name: "brew",
			status: sharedtypes.UpdateStatus{
				InstallSource: sharedtypes.InstallSourceBrew,
			},
			want: "brew upgrade crona-beta",
		},
		{
			name: "brew migration",
			status: sharedtypes.UpdateStatus{
				InstallSource: sharedtypes.InstallSourceBrew,
				BrewFormula:   "crona",
			},
			want: "brew uninstall crona && brew install webxsid/tap/crona-beta",
		},
		{
			name: "script",
			status: sharedtypes.UpdateStatus{
				InstallSource: sharedtypes.InstallSourceScript,
			},
			want: "curl -fsSL https://crona.work/install.sh | sh",
		},
		{
			name: "go",
			status: sharedtypes.UpdateStatus{
				InstallSource: sharedtypes.InstallSourceGo,
			},
			want: "go install github.com/webxsid/crona/...@latest",
		},
		{
			name: "manual release url",
			status: sharedtypes.UpdateStatus{
				ReleaseURL: "https://example.com/releases/v1.0.0",
			},
			want: "https://example.com/releases/v1.0.0",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := updateCommandForStatus(tc.status); got != tc.want {
				t.Fatalf("updateCommandForStatus() = %q, want %q", got, tc.want)
			}
		})
	}
}
