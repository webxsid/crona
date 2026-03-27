package version

import "strings"

var Version = "0.3.1"

const (
	RepoOwner = "webxsid"
	RepoName  = "crona"
)

func Current() string {
	return strings.TrimSpace(Version)
}

func IsDevBuild() bool {
	value := strings.ToLower(strings.TrimSpace(Version))
	return value == "" || value == "dev"
}

func ReleaseTag() string {
	value := strings.TrimSpace(Current())
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "v") {
		return value
	}
	return "v" + value
}
