package helpers

import (
	"net/url"
	"path/filepath"
	"runtime"
	"strings"

	"crona/tui/internal/api"
)

const (
	supportGitHubRepoBaseURL = "https://github.com/webxsid/crona"
	supportGitHubRoadmapURL  = "https://github.com/webxsid/crona/blob/main/ROADMAP.md"
)

func SupportRepoURL() string {
	return supportGitHubRepoBaseURL
}

func SupportIssuesURL() string {
	return supportGitHubRepoBaseURL + "/issues"
}

func SupportDiscussionsURL() string {
	return supportGitHubRepoBaseURL + "/discussions"
}

func SupportReleasesURL() string {
	return supportGitHubRepoBaseURL + "/releases"
}

func SupportRoadmapURL() string {
	return supportGitHubRoadmapURL
}

func SupportBugReportURL(input SupportDiagnosticsInput, bundlePath string) string {
	values := url.Values{}
	values.Set("title", "bug: ")
	values.Set("body", supportBugReportBody(input, bundlePath))
	return supportGitHubRepoBaseURL + "/issues/new?" + values.Encode()
}

func supportBugReportBody(input SupportDiagnosticsInput, bundlePath string) string {
	lines := []string{
		"## Summary",
		"Describe the bug briefly.",
		"",
		"## Reproduction",
		"1.",
		"2.",
		"3.",
		"",
		"## Expected",
		"",
		"## Actual",
		"",
		"## Environment",
		"- Version: " + supportVersionValue(input.UpdateStatus),
		"- Update channel: " + supportUpdateChannelValue(input.UpdateStatus),
		"- Platform: " + runtime.GOOS + "/" + runtime.GOARCH,
		"- View: " + fallbackSupport(input.View, "-"),
	}
	if input.Timer != nil {
		lines = append(lines, "- Timer state: "+fallbackSupport(strings.TrimSpace(input.Timer.State), "unknown"))
	}
	lines = append(lines,
		"",
		"## Diagnostics",
		"- Attach a support bundle zip from `Support -> Bundle` if available.",
	)
	if name := strings.TrimSpace(filepath.Base(bundlePath)); name != "" {
		lines = append(lines, "- Bundle ready locally: `"+name+"`")
	}
	lines = append(lines, "- If you cannot attach the bundle, paste the copied diagnostics summary instead.")
	return strings.Join(lines, "\n")
}

func supportVersionValue(updateStatus *api.UpdateStatus) string {
	if updateStatus == nil {
		return "unknown"
	}
	value := strings.TrimSpace(updateStatus.CurrentVersion)
	if value == "" {
		return "unknown"
	}
	return "v" + value
}

func supportUpdateChannelValue(updateStatus *api.UpdateStatus) string {
	if updateStatus == nil {
		return "stable"
	}
	value := strings.TrimSpace(string(updateStatus.Channel))
	if value == "" {
		return "stable"
	}
	return value
}
