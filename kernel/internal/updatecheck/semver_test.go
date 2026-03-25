package updatecheck

import (
	"testing"

	sharedtypes "crona/shared/types"
)

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{current: "0.2.1", latest: "0.2.2", want: true},
		{current: "0.2.1", latest: "0.2.1", want: false},
		{current: "v0.2.1", latest: "v0.3.0", want: true},
		{current: "0.2.1-beta.1", latest: "0.2.1", want: true},
		{current: "0.2.1", latest: "0.2.1-beta.1", want: false},
		{current: "0.2.1-beta.2", latest: "0.2.1-beta.10", want: true},
		{current: "0.2.1-rc.1", latest: "0.2.1-rc.1+build.4", want: false},
		{current: "0.2.1-alpha.beta", latest: "0.2.1-alpha.1", want: false},
		{current: "dev", latest: "0.2.2", want: false},
	}

	for _, tc := range tests {
		if got := isNewerVersion(tc.current, tc.latest); got != tc.want {
			t.Fatalf("isNewerVersion(%q, %q) = %v, want %v", tc.current, tc.latest, got, tc.want)
		}
	}
}

func TestClearReleaseLocked(t *testing.T) {
	service := &Service{
		status: sharedtypes.UpdateStatus{
			LatestVersion:            "0.3.0",
			ReleaseTag:               "v0.3.0",
			ReleaseName:              "Release",
			ReleaseNotes:             "Notes",
			ReleaseURL:               "https://example.com/release",
			InstallScriptURL:         "https://example.com/install",
			ChecksumsURL:             "https://example.com/checksums",
			PublishedAt:              "2026-03-25T00:00:00Z",
			UpdateAvailable:          true,
			InstallAvailable:         true,
			InstallUnavailableReason: "no reason",
		},
	}

	service.clearReleaseLocked()

	if service.status.LatestVersion != "" || service.status.ReleaseTag != "" || service.status.ReleaseName != "" ||
		service.status.ReleaseNotes != "" || service.status.ReleaseURL != "" || service.status.InstallScriptURL != "" ||
		service.status.ChecksumsURL != "" || service.status.PublishedAt != "" || service.status.UpdateAvailable ||
		service.status.InstallAvailable || service.status.InstallUnavailableReason != "" {
		t.Fatalf("expected release fields to be cleared, got %+v", service.status)
	}
}
