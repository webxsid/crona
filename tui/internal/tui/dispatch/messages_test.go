package dispatch

import (
	"fmt"
	"os/exec"
	"testing"

	"crona/tui/internal/tui/commands"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateInstallPreparedRunsTerminalProcess(t *testing.T) {
	stopped := false
	state, cmd, handled := HandleMessage(MessageState{}, commands.UpdateInstallPreparedMsg{Cmd: exec.Command("sh", "-c", "exit 0")}, MessageDeps{
		CloseEventStop: func() { stopped = true },
	})
	if !handled {
		t.Fatalf("expected prepared install to be handled")
	}
	if !stopped {
		t.Fatalf("expected event stream to stop when install starts")
	}
	if !state.UpdateInstalling {
		t.Fatalf("expected install mode to be active")
	}
	if state.View != uistate.ViewUpdates {
		t.Fatalf("expected updates view, got %q", state.View)
	}
	if cmd == nil {
		t.Fatalf("expected terminal install command")
	}
}

func TestTransportErrorsAreSuppressedDuringUpdateInstall(t *testing.T) {
	called := false
	state, cmd, handled := HandleMessage(MessageState{
		UpdateInstalling: true,
	}, commands.ErrMsg{Err: fmt.Errorf("dial unix /tmp/kernel.sock: connect: no such file or directory")}, MessageDeps{
		SetStatus: func(*MessageState, string, bool) tea.Cmd {
			called = true
			return nil
		},
	})
	if !handled {
		t.Fatalf("expected transport error to be handled")
	}
	if called {
		t.Fatalf("expected transport error to be suppressed instead of setting status")
	}
	if cmd != nil {
		t.Fatalf("expected no follow-up command when suppressing transport error")
	}
	if !state.UpdateInstalling {
		t.Fatalf("expected install state to remain active")
	}
}
