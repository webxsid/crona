package dispatch

import (
	"fmt"
	"os/exec"
	"testing"

	"crona/tui/internal/api"
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

func TestAllIssuesLoadedRestoresCreatedIssueInDefaultView(t *testing.T) {
	selectedID := int64(2)
	state, _, handled := HandleMessage(MessageState{
		View:    uistate.ViewDefault,
		Pane:    uistate.PaneIssues,
		Cursor:  map[uistate.Pane]int{uistate.PaneIssues: 0},
		Filters: map[uistate.Pane]string{uistate.PaneIssues: ""},
		AllIssues: []api.IssueWithMeta{
			{Issue: api.Issue{ID: 1, Title: "Alpha", Status: "planned"}},
			{Issue: api.Issue{ID: 2, Title: "Bravo", Status: "done"}},
		},
	}, commands.AllIssuesLoadedMsg{
		Issues: []api.IssueWithMeta{
			{Issue: api.Issue{ID: 1, Title: "Alpha", Status: "planned"}},
			{Issue: api.Issue{ID: 2, Title: "Bravo", Status: "done"}},
		},
		SelectedIssueID: &selectedID,
	}, MessageDeps{
		ClampFiltered: func(state *MessageState, pane uistate.Pane) {
			state.Cursor[pane] = 0
		},
	})
	if !handled {
		t.Fatalf("expected all issues load to be handled")
	}
	if got := state.Cursor[uistate.PaneIssues]; got != 1 {
		t.Fatalf("expected created issue to remain selected at cursor 1, got %d", got)
	}
}

func TestIssuesLoadedRestoresSelectedIssueInDailyView(t *testing.T) {
	selectedID := int64(20)
	state, _, handled := HandleMessage(MessageState{
		View:          uistate.ViewDaily,
		Pane:          uistate.PaneIssues,
		DashboardDate: "2026-05-06",
		Cursor:        map[uistate.Pane]int{uistate.PaneIssues: 0},
		Filters:       map[uistate.Pane]string{uistate.PaneIssues: ""},
		Issues: []api.Issue{
			{ID: 10, Title: "Alpha", Status: "planned", TodoForDate: strPtr("2026-05-06")},
			{ID: 20, Title: "Bravo", Status: "in_progress", TodoForDate: strPtr("2026-05-06")},
		},
		AllIssues: []api.IssueWithMeta{
			{Issue: api.Issue{ID: 10, StreamID: 1, Title: "Alpha", Status: "planned", TodoForDate: strPtr("2026-05-06")}, RepoID: 1, RepoName: "Work", StreamName: "app"},
			{Issue: api.Issue{ID: 20, StreamID: 1, Title: "Bravo", Status: "in_progress", TodoForDate: strPtr("2026-05-06")}, RepoID: 1, RepoName: "Work", StreamName: "app"},
		},
	}, commands.IssuesLoadedMsg{
		StreamID: 1,
		Issues: []api.Issue{
			{ID: 10, Title: "Alpha", Status: "planned", TodoForDate: strPtr("2026-05-06")},
			{ID: 20, Title: "Bravo", Status: "in_progress", TodoForDate: strPtr("2026-05-06")},
		},
		SelectedIssueID: &selectedID,
	}, MessageDeps{
		ClampFiltered: func(state *MessageState, pane uistate.Pane) {
			state.Cursor[pane] = 0
		},
	})
	if !handled {
		t.Fatalf("expected issues load to be handled")
	}
	if got := state.Cursor[uistate.PaneIssues]; got != 1 {
		t.Fatalf("expected selected issue to remain on cursor 1, got %d", got)
	}
}

func TestIssuesLoadedClampsWhenSelectedIssueMissing(t *testing.T) {
	state, _, handled := HandleMessage(MessageState{
		View:    uistate.ViewDaily,
		Pane:    uistate.PaneIssues,
		Cursor:  map[uistate.Pane]int{uistate.PaneIssues: 1},
		Filters: map[uistate.Pane]string{uistate.PaneIssues: ""},
		Issues: []api.Issue{
			{ID: 10, Title: "Alpha", Status: "planned"},
			{ID: 20, Title: "Bravo", Status: "in_progress"},
		},
	}, commands.IssuesLoadedMsg{
		StreamID: 1,
		Issues: []api.Issue{
			{ID: 10, Title: "Alpha", Status: "planned"},
		},
	}, MessageDeps{
		ClampFiltered: func(state *MessageState, pane uistate.Pane) {
			state.Cursor[pane] = 0
		},
	})
	if !handled {
		t.Fatalf("expected issues load to be handled")
	}
	if got := state.Cursor[uistate.PaneIssues]; got != 0 {
		t.Fatalf("expected cursor to clamp to the remaining issue, got %d", got)
	}
}

func TestIssuePinnedDailyChangedSetsStatusMessage(t *testing.T) {
	state, _, handled := HandleMessage(MessageState{}, commands.IssuePinnedDailyChangedMsg{
		IssueID: 42,
		Pinned:  true,
	}, MessageDeps{
		SetStatus: func(state *MessageState, message string, isErr bool) tea.Cmd {
			state.StatusMsg = message
			state.StatusErr = isErr
			return nil
		},
	})
	if !handled {
		t.Fatalf("expected pin change message to be handled")
	}
	if got := state.StatusMsg; got != "Issue pinned" {
		t.Fatalf("expected status message to reflect pinning, got %q", got)
	}
	if state.StatusErr {
		t.Fatalf("expected pin status to be informational")
	}

	state, _, handled = HandleMessage(MessageState{}, commands.IssuePinnedDailyChangedMsg{
		IssueID: 42,
		Pinned:  false,
	}, MessageDeps{
		SetStatus: func(state *MessageState, message string, isErr bool) tea.Cmd {
			state.StatusMsg = message
			state.StatusErr = isErr
			return nil
		},
	})
	if !handled {
		t.Fatalf("expected unpin change message to be handled")
	}
	if got := state.StatusMsg; got != "Issue unpinned" {
		t.Fatalf("expected status message to reflect unpinning, got %q", got)
	}
}

func TestSettingsLoadedOpensOnboardingWhenIncomplete(t *testing.T) {
	opened := false
	state, _, handled := HandleMessage(MessageState{}, commands.SettingsLoadedMsg{
		Settings: &api.CoreSettings{
			OnboardingCompleted: false,
		},
	}, MessageDeps{
		ClampFiltered: func(*MessageState, uistate.Pane) {},
		OpenOnboardingDialog: func(state *MessageState) {
			opened = true
			state.Dialog = "onboarding"
		},
	})
	if !handled {
		t.Fatalf("expected settings load to be handled")
	}
	if !opened {
		t.Fatalf("expected onboarding dialog to open")
	}
	if state.Dialog != "onboarding" {
		t.Fatalf("expected onboarding dialog state, got %q", state.Dialog)
	}
}

func TestSettingsLoadedSkipsOnboardingWhenCompleted(t *testing.T) {
	opened := false
	state, _, handled := HandleMessage(MessageState{}, commands.SettingsLoadedMsg{
		Settings: &api.CoreSettings{
			OnboardingCompleted: true,
		},
	}, MessageDeps{
		ClampFiltered: func(*MessageState, uistate.Pane) {},
		OpenOnboardingDialog: func(state *MessageState) {
			opened = true
			state.Dialog = "onboarding"
		},
	})
	if !handled {
		t.Fatalf("expected settings load to be handled")
	}
	if opened {
		t.Fatalf("did not expect onboarding dialog to open")
	}
	if state.Dialog != "" {
		t.Fatalf("expected no dialog, got %q", state.Dialog)
	}
}

func strPtr(value string) *string {
	return &value
}
