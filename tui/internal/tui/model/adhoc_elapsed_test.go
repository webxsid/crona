package model

import (
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
	tea "github.com/charmbracelet/bubbletea"
	"testing"
	"time"
)

func TestAdhocAttributionUsesDerivedElapsedTime(t *testing.T) {
	start := time.Now().Add(-185 * time.Second).Format(time.RFC3339)
	for _, tc := range []struct {
		name    string
		start   *string
		elapsed int
		want    int
	}{
		{name: "session start", start: &start, want: 185},
		{name: "local ticks", elapsed: 185, want: 185},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := Model{
				timer:   &api.TimerState{State: "running", Adhoc: true, SessionStartTime: tc.start},
				elapsed: tc.elapsed,
			}
			state := m.dialogSnapshot().OpenAdhocResolution()
			if state.AdhocElapsedSeconds < tc.want || state.AdhocElapsedSeconds > tc.want+2 {
				t.Fatalf("elapsed = %d, want approximately %d", state.AdhocElapsedSeconds, tc.want)
			}
			m = m.withDialogState(state)
			state = m.dialogState()
			state.ChoiceCursor = 2
			state, _, _ = dialogstate.Update(
				state,
				dialogstate.UpdateContext{},
				"2026-10-08",
				tea.KeyMsg{Type: tea.KeyEnter},
			)
			m = m.withDialogState(state)
			if m.dialogState().AdhocResolution != sharedtypes.AdhocSessionResolutionKeep ||
				m.dialogState().AdhocElapsedSeconds < tc.want {
				t.Fatalf("keep confirmation lost elapsed: %+v", m.dialogState())
			}
		})
	}
}

func TestActiveAdhocTimerDoesNotMarkOrdinaryFormsAsAttribution(t *testing.T) {
	m := Model{timer: &api.TimerState{State: "running", Adhoc: true}, elapsed: 185}
	ordinary := m.dialogSnapshot().OpenCreateIssueDefault()
	if ordinary.AdhocTimer {
		t.Fatal("ordinary issue creation inherited ad hoc attribution")
	}
	conversion := m.dialogSnapshot().
		OpenPromoteAdhocIssue(api.IssueWithMeta{Issue: sharedtypes.Issue{ID: 1, Title: "Saved", Adhoc: true}})
	if conversion.AdhocTimer || !conversion.PromotingAdhoc {
		t.Fatal("conversion inherited timer attribution")
	}
	attribution := m.dialogSnapshot().OpenAdhocResolution()
	if !attribution.AdhocTimer || attribution.AdhocElapsedSeconds != 185 {
		t.Fatal("attribution lost marker or elapsed")
	}
}
