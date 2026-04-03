package commands

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"crona/tui/internal/api"
	"crona/tui/internal/logger"
	helperpkg "crona/tui/internal/tui/helpers"

	tea "github.com/charmbracelet/bubbletea"
)

func GenerateSupportBundle(c *api.Client, input helperpkg.SupportDiagnosticsInput, window time.Duration) tea.Cmd {
	return func() tea.Msg {
		now := time.Now().UTC()
		since := now.Add(-window)
		collectionErrors := []string{}

		ops, err := c.ListOpsSince(since.Format(time.RFC3339))
		if err != nil {
			logger.Errorf("GenerateSupportBundle ListOpsSince: %v", err)
			collectionErrors = append(collectionErrors, "ops: "+err.Error())
		}

		baseDir := helperpkg.SupportRuntimeBaseDir(input.KernelInfo)
		tuiErrors, kernelErrors, logErrs := helperpkg.ReadRecentSupportErrorEntries(baseDir, since, now)
		collectionErrors = append(collectionErrors, logErrs...)

		path, sizeBytes, err := helperpkg.GenerateSupportBundle(baseDir, now, window, input, ops, tuiErrors, kernelErrors, collectionErrors)
		if err != nil {
			logger.Errorf("GenerateSupportBundle write: %v", err)
			if strings.TrimSpace(baseDir) == "" {
				return ErrMsg{Err: fmt.Errorf("runtime base directory unavailable for support bundle")}
			}
			return ErrMsg{Err: err}
		}
		return SupportBundleGeneratedMsg{Path: path, SizeBytes: sizeBytes, WindowLabel: helperpkg.SupportRecentWindowLabel(window)}
	}
}

func OpenExternalPath(path string) tea.Cmd {
	return func() tea.Msg {
		target := strings.TrimSpace(path)
		if target == "" {
			return ErrMsg{Err: fmt.Errorf("path is unavailable")}
		}
		target = filepath.Dir(target)
		cmd, err := externalOpenCommand(target)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			if err != nil {
				return ErrMsg{Err: err}
			}
			return nil
		})()
	}
}
