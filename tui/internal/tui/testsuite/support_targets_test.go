package testsuite

import (
	"net/url"
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
)

func TestSupportBugReportURLIncludesPrefillAndBundleGuidance(t *testing.T) {
	input := helperpkg.SupportDiagnosticsInput{
		View: "support",
		Timer: &api.TimerState{
			State: "running",
		},
		UpdateStatus: &api.UpdateStatus{
			CurrentVersion: "0.4.0-beta.2",
			RunningChannel: sharedtypes.UpdateChannelBeta,
			RunningIsBeta:  true,
			Channel:        sharedtypes.UpdateChannelBeta,
		},
	}

	raw := helperpkg.SupportBugReportURL(input, "/tmp/support-bundle-123.zip")
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("url.Parse: %v", err)
	}
	query := parsed.Query()
	if got := query.Get("title"); got != "bug: " {
		t.Fatalf("unexpected title %q", got)
	}
	body := query.Get("body")
	for _, want := range []string{
		"## Summary",
		"Version: v0.4.0-beta.2",
		"Running channel: beta",
		"Update channel: beta",
		"Attach a support bundle zip",
		"support-bundle-123.zip",
		"Timer state: running",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected prefill body to contain %q, got %q", want, body)
		}
	}
}

func TestSupportPublicURLs(t *testing.T) {
	tests := map[string]string{
		"issues":      helperpkg.SupportIssuesURL(),
		"discussions": helperpkg.SupportDiscussionsURL(),
		"releases":    helperpkg.SupportReleasesURL(),
		"roadmap":     helperpkg.SupportRoadmapURL(),
	}
	for name, raw := range tests {
		parsed, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("%s url.Parse: %v", name, err)
		}
		if parsed.Scheme != "https" || parsed.Host != "github.com" {
			t.Fatalf("%s expected github https url, got %q", name, raw)
		}
	}
}
