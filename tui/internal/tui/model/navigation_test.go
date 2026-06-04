package model

import (
	"strings"
	"testing"
	"time"

	"crona/tui/internal/api"
	viewruntime "crona/tui/internal/tui/views/runtime"
)

func TestOpenSelectedViewDialogIncludesWorkedEstimate(t *testing.T) {
	issue := api.Issue{
		ID:            1,
		Title:         "Investigate timer display",
		Status:        "in_progress",
		WorkedSeconds: 0,
	}
	m := Model{
		view:   ViewDefault,
		pane:   PaneIssues,
		cursor: map[Pane]int{PaneIssues: 0},
		issues: []api.Issue{issue},
		allIssues: []api.IssueWithMeta{{
			Issue: api.Issue{
				ID:            1,
				Title:         "Investigate timer display",
				Status:        "in_progress",
				WorkedSeconds: 4500,
			},
			RepoName:   "Core",
			StreamName: "TUI",
		}},
	}

	next, ok := m.openSelectedViewDialog()
	if !ok {
		t.Fatal("expected issue details dialog to open")
	}
	if next.dialog != "view_entity" {
		t.Fatalf("expected view entity dialog, got %q", next.dialog)
	}
	if !strings.Contains(next.dialogViewMeta, "worked 1h15m / est. -") {
		t.Fatalf("expected dialog meta to include worked/estimate summary, got %q", next.dialogViewMeta)
	}
	if !strings.Contains(next.dialogViewBody, "Worked / est.") ||
		!strings.Contains(next.dialogViewBody, "1h15m / est. -") {
		t.Fatalf("expected dialog body to include worked/estimate summary, got %q", next.dialogViewBody)
	}
}

func TestProtectedRestKeepsSettingsReachable(t *testing.T) {
	restDate := time.Now().Format("2006-01-02")
	m := Model{
		view: ViewSettings,
		settings: &api.CoreSettings{
			RestSpecificDates: []string{restDate},
		},
	}

	views := m.availableViews()
	foundSettings := false
	for _, view := range views {
		if view == ViewSettings {
			foundSettings = true
			break
		}
	}
	if !foundSettings {
		t.Fatalf("expected protected rest to keep settings in available views, got %#v", views)
	}
	if got := m.jumpAvailableViews(); len(got) == 0 || got[len(got)-1] != ViewSettings {
		t.Fatalf("expected protected rest jump views to include settings, got %#v", got)
	}
	state := m.layoutState()
	if state.View != ViewSettings {
		t.Fatalf("expected settings view to remain visible during configured rest, got %q", state.View)
	}
	if !state.ContentState.RestModeActive {
		t.Fatal("expected settings view to keep rest mode active")
	}
	if state.ContentState.AwayModeActive {
		t.Fatal("expected date-based rest not to report manual away mode")
	}
	inputState := m.inputState()
	if inputState.ActiveView != ViewSettings {
		t.Fatalf("expected settings input view to remain reachable, got %q", inputState.ActiveView)
	}
	if active, away, detail := viewruntime.ProtectedRestMode(m.settings, restDate); !active || away || detail == "" {
		t.Fatalf("expected configured rest to be date-based, got active=%t away=%t detail=%q", active, away, detail)
	}
}
