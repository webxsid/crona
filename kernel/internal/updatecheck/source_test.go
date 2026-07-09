package updatecheck

import (
	"testing"

	sharedtypes "crona/shared/types"
	versionpkg "crona/shared/version"
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
			name: "scoop",
			path: `C:\Users\alice\scoop\apps\crona\current\crona.exe`,
			want: sharedtypes.InstallSourceScoop,
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
	prev := versionpkg.Version
	versionpkg.Version = "1.6.1"
	defer func() { versionpkg.Version = prev }()

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
			want: "brew upgrade crona",
		},
		{
			name: "scoop stable",
			status: sharedtypes.UpdateStatus{
				InstallSource:  sharedtypes.InstallSourceScoop,
				ReleaseChannel: sharedtypes.UpdateChannelStable,
			},
			want: "scoop update crona",
		},
		{
			name: "scoop beta",
			status: sharedtypes.UpdateStatus{
				InstallSource:  sharedtypes.InstallSourceScoop,
				ReleaseChannel: sharedtypes.UpdateChannelBeta,
			},
			want: "scoop update crona-beta",
		},
		{
			name: "brew migration",
			status: sharedtypes.UpdateStatus{
				InstallSource: sharedtypes.InstallSourceBrew,
				BrewFormula:   "crona-beta",
			},
			want: "brew uninstall crona-beta && brew install webxsid/tap/crona",
		},
		{
			name: "script",
			status: sharedtypes.UpdateStatus{
				InstallSource: sharedtypes.InstallSourceScript,
			},
			want: "https://crona.work/migration",
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

func TestDefaultReleaseChannel(t *testing.T) {
	tests := []struct {
		name    string
		source  sharedtypes.InstallSource
		version string
		want    sharedtypes.UpdateChannel
	}{
		{
			name:    "stable scoop",
			source:  sharedtypes.InstallSourceScoop,
			version: "1.6.1",
			want:    sharedtypes.UpdateChannelStable,
		},
		{
			name:    "beta release scoop",
			source:  sharedtypes.InstallSourceScoop,
			version: "1.6.1-beta.1",
			want:    sharedtypes.UpdateChannelBeta,
		},
		{
			name:    "beta scoop path",
			source:  sharedtypes.InstallSourceScoop,
			version: "1.6.1",
			want:    sharedtypes.UpdateChannelBeta,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prev := versionpkg.Version
			versionpkg.Version = tc.version
			defer func() { versionpkg.Version = prev }()
			path := ""
			if tc.name == "beta scoop path" {
				path = `C:\Users\alice\scoop\apps\crona-beta\current\crona.exe`
			}
			if got := defaultReleaseChannel(tc.source, path); got != tc.want {
				t.Fatalf("defaultReleaseChannel() = %q, want %q", got, tc.want)
			}
		})
	}
}
