package types

import "strings"

type InstallSource string

const (
	InstallSourceUnknown InstallSource = "unknown"
	InstallSourceScript  InstallSource = "script"
	InstallSourceBrew    InstallSource = "brew"
	InstallSourceWinget  InstallSource = "winget"
	InstallSourceGo      InstallSource = "go"
	InstallSourceManual  InstallSource = "manual"
)

func NormalizeInstallSource(value InstallSource) InstallSource {
	switch value {
	case InstallSourceScript, InstallSourceBrew, InstallSourceWinget, InstallSourceGo, InstallSourceManual:
		return value
	default:
		return InstallSourceUnknown
	}
}

func ParseInstallSource(value string) InstallSource {
	return NormalizeInstallSource(InstallSource(strings.ToLower(strings.TrimSpace(value))))
}
