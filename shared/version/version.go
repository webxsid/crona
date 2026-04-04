package version

import (
	"strings"

	sharedtypes "crona/shared/types"
)

var Version = "0.4.0-beta.3"

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

func IsBetaVersion(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Contains(value, "-beta")
}

func IsBetaRelease() bool {
	return IsBetaVersion(Current())
}

func RunningChannel() sharedtypes.UpdateChannel {
	if IsBetaRelease() {
		return sharedtypes.UpdateChannelBeta
	}
	return sharedtypes.UpdateChannelStable
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
