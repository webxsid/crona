package dialogruntime

import (
	"os"
	"os/exec"
	"runtime"

	"crona/tui/internal/tui/commands"

	tea "github.com/charmbracelet/bubbletea"
)

func OpenEditor(filePath string, errMsg func(error) tea.Msg) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}
	c := exec.Command(editor, filePath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errMsg(err)
		}
		return commands.EditorDoneMsg{}
	})
}

func OpenDefaultViewer(filePath string, errMsg func(error) tea.Msg) tea.Cmd {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", filePath)
	case "linux":
		c = exec.Command("xdg-open", filePath)
	case "windows":
		c = exec.Command("cmd", "/c", "start", "", filePath)
	default:
		return func() tea.Msg { return errMsg(os.ErrInvalid) }
	}
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errMsg(err)
		}
		return nil
	})
}
