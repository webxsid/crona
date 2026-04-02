package helpers

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
)

func SupportDiagnosticsSummary(info *api.KernelInfo, assets *api.ExportAssetStatus, update *api.UpdateStatus, health *api.Health, tuiPath, kernelPath string) string {
	lines := []string{}
	if update != nil {
		lines = append(lines, fmt.Sprintf("Version: v%s", fallbackSupport(strings.TrimSpace(update.CurrentVersion), "unknown")))
		lines = append(lines, fmt.Sprintf("Update channel: %s", fallbackSupport(strings.TrimSpace(string(update.Channel)), "stable")))
	}
	if info != nil {
		lines = append(lines, fmt.Sprintf("Environment: %s", fallbackSupport(strings.TrimSpace(info.Env), "unknown")))
		lines = append(lines, fmt.Sprintf("Kernel transport: %s", fallbackSupport(strings.TrimSpace(info.Transport), "-")))
		lines = append(lines, fmt.Sprintf("Kernel endpoint: %s", fallbackSupport(strings.TrimSpace(info.Endpoint), "-")))
		lines = append(lines, fmt.Sprintf("Scratch dir: %s", fallbackSupport(strings.TrimSpace(info.ScratchDir), "-")))
	}
	if health != nil {
		lines = append(lines, fmt.Sprintf("Health: %s (db=%t)", fallbackSupport(strings.TrimSpace(health.Status), "unknown"), health.DB))
	}
	if assets != nil {
		lines = append(lines, fmt.Sprintf("Reports dir: %s", fallbackSupport(strings.TrimSpace(assets.ReportsDir), "-")))
		lines = append(lines, fmt.Sprintf("ICS dir: %s", fallbackSupport(strings.TrimSpace(assets.ICSDir), "-")))
	}
	lines = append(lines, fmt.Sprintf("TUI path: %s", fallbackSupport(strings.TrimSpace(tuiPath), "-")))
	lines = append(lines, fmt.Sprintf("Kernel path: %s", fallbackSupport(strings.TrimSpace(kernelPath), "-")))
	return strings.Join(lines, "\n")
}

func fallbackSupport(value, alt string) string {
	if strings.TrimSpace(value) == "" {
		return alt
	}
	return strings.TrimSpace(value)
}
