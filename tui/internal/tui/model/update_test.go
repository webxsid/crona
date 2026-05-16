package model

import (
	"testing"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDispatchMessageStatePreservesStashConflictDialogPayload(t *testing.T) {
	model := Model{width: 100}
	model = model.openStashConflictDialog(sharedtypes.StashConflict{
		IssueID: 42,
		Stashes: []sharedtypes.Stash{{
			ID:        "stash-1",
			CreatedAt: "2026-04-10T09:00:00Z",
		}},
	}, 7, 8, 42)

	roundTripped := model.applyDispatchMessageState(model.dispatchMessageState())
	if roundTripped.dialog != "stash_conflict" {
		t.Fatalf("expected stash conflict dialog, got %q", roundTripped.dialog)
	}
	if roundTripped.dialogDeleteID != "stash-1" {
		t.Fatalf("expected stash id to survive dispatch bridge, got %q", roundTripped.dialogDeleteID)
	}
	if roundTripped.dialogRepoID != 7 || roundTripped.dialogStreamID != 8 || roundTripped.dialogIssueID != 42 {
		t.Fatalf("expected issue path to survive dispatch bridge, got repo=%d stream=%d issue=%d", roundTripped.dialogRepoID, roundTripped.dialogStreamID, roundTripped.dialogIssueID)
	}
	if len(roundTripped.dialogChoiceValues) != 2 || roundTripped.dialogChoiceValues[0] != "resume" || roundTripped.dialogChoiceValues[1] != "continue" {
		t.Fatalf("expected choice values to survive dispatch bridge, got %#v", roundTripped.dialogChoiceValues)
	}
}

func TestTimerActivityTouchCmdOnlyForActiveTimerAndThrottles(t *testing.T) {
	now := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	model := Model{client: api.NewClient("unix", "/tmp/missing.sock", ""), timer: &api.TimerState{State: "idle"}}
	if cmd := model.timerActivityTouchCmd(now); cmd != nil {
		t.Fatalf("expected no touch command while idle")
	}

	model.timer = &api.TimerState{State: "running"}
	if cmd := model.timerActivityTouchCmd(now); cmd == nil {
		t.Fatalf("expected touch command for active timer")
	}
	if cmd := model.timerActivityTouchCmd(now.Add(30 * time.Second)); cmd != nil {
		t.Fatalf("expected touch command to be throttled")
	}
	if cmd := model.timerActivityTouchCmd(now.Add(61 * time.Second)); cmd == nil {
		t.Fatalf("expected touch command after throttle window")
	}
}

func TestNewStartsWellbeingCursorsAtScrollableBottom(t *testing.T) {
	model := New("unix", "/tmp/missing.sock", "", "dev", "/tmp/crona", make(chan struct{}), nil)
	if model.cursor[PaneWellbeingSummary] == 0 {
		t.Fatal("expected wellbeing summary cursor to start near bottom")
	}
	if model.cursor[PaneWellbeingTrends] == 0 {
		t.Fatal("expected wellbeing trends cursor to start near bottom")
	}
	if model.cursor[PaneWellbeingStreaks] == 0 {
		t.Fatal("expected wellbeing streaks cursor to start near bottom")
	}
}

func TestDialogModeQuitsOnQAndCtrlC(t *testing.T) {
	model := Model{}.withDialogState(dialogstate.State{Kind: "onboarding"})

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatalf("expected quit command for q")
	}
	if next.(Model).dialog != "onboarding" {
		t.Fatalf("expected dialog state to remain unchanged, got %q", next.(Model).dialog)
	}

	next, cmd = model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatalf("expected quit command for ctrl+c")
	}
	if next.(Model).dialog != "onboarding" {
		t.Fatalf("expected dialog state to remain unchanged, got %q", next.(Model).dialog)
	}
}

func TestDialogRuntimeDepsWireTelemetryHooks(t *testing.T) {
	model := Model{client: api.NewClient("unix", "/tmp/missing.sock", "")}
	deps := model.dialogRuntimeDeps()
	if deps.PatchTelemetrySettings == nil {
		t.Fatal("expected patch telemetry settings hook to be wired")
	}
	if deps.CompleteOnboarding == nil {
		t.Fatal("expected complete onboarding hook to be wired")
	}
}
