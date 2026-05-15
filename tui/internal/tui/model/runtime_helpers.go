package model

import (
	"sync"

	"crona/shared/config"
	"crona/tui/internal/api"
	commands "crona/tui/internal/tui/commands"
	appruntime "crona/tui/internal/tui/runtime"
)

func (m Model) selfUpdateInstallAvailable() bool {
	return m.updateStatus != nil && m.updateStatus.InstallAvailable && m.selfUpdateUnsupportedReason() == ""
}

func (m Model) selfUpdateUnsupportedReason() string {
	if m.updateStatus != nil && commands.IsLocalLoopbackUpdateURL(m.updateStatus.InstallScriptURL) && m.isDevMode() {
		return ""
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
