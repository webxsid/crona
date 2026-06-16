package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"crona/tui/internal/api"
	"crona/tui/internal/logger"

	tea "github.com/charmbracelet/bubbletea"
)

func CheckUpdateNow(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		status, err := c.CheckUpdateNow()
		if err != nil {
			logger.Errorf("CheckUpdateNow: %v", err)
			return ErrMsg{Err: err}
		}
		return UpdateStatusLoadedMsg{Status: status}
	}
}

func OpenExternalURL(rawURL string) tea.Cmd {
	return func() tea.Msg {
		target := strings.TrimSpace(rawURL)
		if target == "" {
			return ErrMsg{Err: fmt.Errorf("release URL is unavailable")}
		}
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

func externalOpenCommand(target string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", target), nil
	case "linux":
		return exec.Command("xdg-open", target), nil
	case "windows":
		return exec.Command("cmd", "/c", "start", "", target), nil
	default:
		return nil, os.ErrInvalid
	}
}
