package version

import (
	"strings"

	sharedtypes "crona/shared/types"
)

var Version = "1.6.1-beta.2"

var InstallSource = ""

const (
	RepoOwner                 = "webxsid"
	RepoName                  = "crona"
	InstallScriptMigrationURL = "https://crona.work/migration"
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

func IsInstallScriptDeprecationVersion(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "v")
	return strings.HasPrefix(value, "1.6.")
}

func InstallScriptDeprecationEnabled() bool {
	return IsInstallScriptDeprecationVersion(Current())
}

func InstallScriptDeprecationMessage() string {
	return "Moving forward, Crona will stop exposing the GitHub install script. Migrate to a managed package installer."
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
