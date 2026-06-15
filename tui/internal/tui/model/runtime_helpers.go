package model

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"crona/shared/config"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	commands "crona/tui/internal/tui/commands"
	appruntime "crona/tui/internal/tui/runtime"
)

func (m Model) selfUpdateInstallAvailable() bool {
	return m.updateStatus != nil && m.updateStatus.InstallAvailable &&
		m.selfUpdateUnsupportedReason() == ""
}

func (m Model) selfUpdateUnsupportedReason() string {
	if m.updateStatus != nil &&
		commands.IsLocalLoopbackUpdateURL(m.updateStatus.InstallScriptURL) &&
		m.isDevMode() {
		return ""
	}
	if m.updateStatus != nil {
		if reason := strings.TrimSpace(m.updateStatus.InstallUnavailableReason); reason != "" {
			return reason
		}
	}
	switch m.effectiveInstallSource() {
	case sharedtypes.InstallSourceScript:
		return ""
	case sharedtypes.InstallSourceBrew:
		if m.updateStatus != nil {
			if command := strings.TrimSpace(m.updateStatus.UpdateCommand); command != "" {
				if strings.HasPrefix(command, "brew uninstall ") {
					return "Homebrew channel migration required. Use the migration command shown below."
				}
			}
		}
		return fmt.Sprintf("Installed via Homebrew. Use brew upgrade %s.", currentBrewFormula())
	case sharedtypes.InstallSourceGo:
		return "Installed via go install. Use go install github.com/webxsid/crona/...@latest."
	case sharedtypes.InstallSourceManual, sharedtypes.InstallSourceUnknown:
		if m.updateStatus != nil {
			if command := strings.TrimSpace(m.updateStatus.UpdateCommand); command != "" {
				return fmt.Sprintf("Installed manually. Use %s.", command)
			}
			if strings.TrimSpace(m.updateStatus.ReleaseURL) != "" {
				return "Installed manually. Open the GitHub release page to update."
			}
		}
		return "Installed manually. Please update manually."
	}
	if reason := appruntime.NonStandardRuntimeReason(m.currentExecutablePath, config.TUIBinaryNameForMode(kernelEnvMode(m.kernelInfo))); reason != "" {
		return reason
	}
	if m.kernelInfo == nil {
		return "Kernel info is unavailable. Please update manually."
	}
	if reason := appruntime.NonStandardRuntimeReason(m.kernelInfo.ExecutablePath, config.KernelBinaryNameForMode(kernelEnvMode(m.kernelInfo))); reason != "" {
		return reason
	}
	return ""
}

func (m Model) effectiveInstallSource() sharedtypes.InstallSource {
	if m.updateStatus != nil {
		if source := sharedtypes.NormalizeInstallSource(m.updateStatus.InstallSource); source != sharedtypes.InstallSourceUnknown {
			return source
		}
	}
	if source := installSourceFromPath(m.currentExecutablePath); source != sharedtypes.InstallSourceUnknown {
		return source
	}
	if source := installSourceFromPath(kernelExecutablePath(m.kernelInfo)); source != sharedtypes.InstallSourceUnknown {
		return source
	}
	return sharedtypes.InstallSourceUnknown
}

func installSourceFromPath(path string) sharedtypes.InstallSource {
	normalized := strings.ToLower(strings.TrimSpace(path))
	if normalized == "" {
		return sharedtypes.InstallSourceUnknown
	}
	if strings.Contains(normalized, "/opt/homebrew/") ||
		strings.Contains(normalized, "/usr/local/cellar/") ||
		strings.Contains(normalized, "/home/linuxbrew/.linuxbrew/") ||
		strings.Contains(normalized, "/homebrew/") {
		return sharedtypes.InstallSourceBrew
	}
	if strings.Contains(normalized, string(filepath.Separator)+"go"+string(filepath.Separator)+"bin"+string(filepath.Separator)) ||
		strings.Contains(normalized, "/gobin/") {
		return sharedtypes.InstallSourceGo
	}
	return sharedtypes.InstallSourceUnknown
}

func (m *Model) stopEventStream() {
	if m.eventStop == nil {
		return
	}
	m.eventStop.Stop()
}

type eventStreamStop struct {
	ch   chan struct{}
	once sync.Once
}

func newEventStreamStop(ch chan struct{}) *eventStreamStop {
	if ch == nil {
		return nil
	}
	return &eventStreamStop{ch: ch}
}

func (s *eventStreamStop) Stop() {
	if s == nil || s.ch == nil {
		return
	}
	s.once.Do(func() {
		close(s.ch)
	})
}

func kernelEnvMode(info *api.KernelInfo) string {
	if info == nil {
		return ""
	}
	return info.Env
}

func kernelExecutablePath(info *api.KernelInfo) string {
	if info == nil {
		return ""
	}
	return info.ExecutablePath
}
