package types

import "strings"

type InstallSource string

const (
	InstallSourceUnknown InstallSource = "unknown"
	InstallSourceScript  InstallSource = "script"
	InstallSourceBrew    InstallSource = "brew"
	InstallSourceScoop   InstallSource = "scoop"
	InstallSourceGo      InstallSource = "go"
	InstallSourceManual  InstallSource = "manual"
)

func NormalizeInstallSource(value InstallSource) InstallSource {
	switch value {
	case InstallSourceScript, InstallSourceBrew, InstallSourceScoop, InstallSourceGo, InstallSourceManual:
		return value
	default:
		return InstallSourceUnknown
	}
}

func ParseInstallSource(value string) InstallSource {
	return NormalizeInstallSource(InstallSource(strings.ToLower(strings.TrimSpace(value))))
}
