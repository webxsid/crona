package version

import (
	"strings"

	sharedtypes "crona.work/shared/types"
)

var Version = "1.10.0-beta.5"

var InstallSource = ""

// RepoOwner owns Crona's GitHub repository, the source of truth for releases. It is the one
// place the owner is written for the Go code: update checks, update hints and support links
// all build on it. The release scripts read their copy from scripts/project_meta.sh.
const (
	RepoOwner                 = "crona-app"
	RepoName                  = "crona"
	InstallScriptMigrationURL = "https://crona.work/migration"
)

// PackageOwner owns the Homebrew tap and Scoop bucket repositories. They stay on the
// original account so existing taps keep their names; set it to RepoOwner if they move.
const PackageOwner = "webxsid"

// RepoSlug is owner/name, as GitHub's API and URLs use it.
func RepoSlug() string { return RepoOwner + "/" + RepoName }

// RepoURL is the repository's web address.
func RepoURL() string { return "https://github.com/" + RepoSlug() }

// ModulePath is Crona's Go module path. It is served from crona.work rather than GitHub so
// it does not change when the repository moves.
const ModulePath = "crona.work"

// GoInstallCommand installs or updates Crona from source with go install: the crona CLI,
// crona-daemon and crona-tui.
func GoInstallCommand() string { return "go install " + ModulePath + "/...@latest" }

// BrewTapFormula is formula's install name in Crona's Homebrew tap, like owner/tap/crona.
func BrewTapFormula(formula string) string { return PackageOwner + "/tap/" + formula }

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
