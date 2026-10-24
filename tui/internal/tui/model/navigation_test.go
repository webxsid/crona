package model

import (
	"strings"
	"testing"
	"time"

	"crona/tui/internal/api"
	uistate "crona/tui/internal/tui/state"
	viewruntime "crona/tui/internal/tui/views/runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
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
		t.Fatalf(
			"expected dialog meta to include worked/estimate summary, got %q",
			next.dialogViewMeta,
		)
	}
	if strings.Contains(next.dialogViewBody, "Worked / est.") {
		t.Fatalf(
			"expected worked/estimate only in the meta, not repeated in the body, got %q",
			next.dialogViewBody,
		)
	}
}

func TestIssueDetailsDialogSplitsNotesAndHistory(t *testing.T) {
	notes := "Keep the window open.\n[2026-10-14T09:00:00Z] planned -> ready: Keep the window open."
	issue := api.Issue{ID: 1, Title: "Window", Status: "ready", Notes: &notes}
	m := Model{
		width:     120,
		view:      ViewDefault,
		pane:      PaneIssues,
		cursor:    map[Pane]int{PaneIssues: 0},
		issues:    []api.Issue{issue},
		allIssues: []api.IssueWithMeta{{Issue: issue, RepoName: "Core", StreamName: "TUI"}},
	}
	next, ok := m.openSelectedViewDialog()
	if !ok {
		t.Fatal("expected issue details dialog to open")
	}
	body := ansi.Strip(next.dialogViewBody)
	for _, want := range []string{"Notes\nKeep the window open.", "History\n", "planned → ready"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected %q in the dialog body, got %q", want, body)
		}
	}
	if strings.Contains(body, "[2026-10-14") || strings.Count(body, "Keep the window open.") != 1 {
		t.Fatalf("expected the raw entry replaced and its repeated note hidden, got %q", body)
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
		t.Fatalf(
			"expected settings view to remain visible during configured rest, got %q",
			state.View,
		)
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
	if active, away, detail := viewruntime.ProtectedRestMode(m.settings, restDate); !active ||
		away ||
		detail == "" {
		t.Fatalf(
			"expected configured rest to be date-based, got active=%t away=%t detail=%q",
			active,
			away,
			detail,
		)
	}
}

func TestAvailableViewsFollowSidebarOrder(t *testing.T) {
	views := (Model{}).availableViews()
	want := []View{
		ViewDaily,
		ViewReview,
		ViewMomentum,
		ViewDefault,
		ViewHabitList,
		ViewTags,
		ViewOps,
		ViewReports,
		ViewConfig,
		ViewSettings,
		ViewAlerts,
		ViewUpdates,
		ViewSupport,
		ViewSessionHistory,
		ViewHabitHistory,
	}
	if len(views) != len(want) {
		t.Fatalf("unexpected view count: got %d want %d (%#v)", len(views), len(want), views)
	}
	for i := range want {
		if views[i] != want[i] {
			t.Fatalf("view order mismatch at %d: got %q want %q (%#v)", i, views[i], want[i], views)
		}
	}
	if views[0] != uistate.ViewDaily {
		t.Fatalf("expected daily to be first sidebar navigation target, got %q", views[0])
	}
}

func daySummaryModel(width int) Model {
	date := "2026-05-27"
	issue := api.Issue{ID: 1, Title: "Today", Status: "planned", TodoForDate: &date}
	return Model{
		view: ViewDaily, pane: PaneIssues, width: width, height: 40, date: date,
		cursor:  map[Pane]int{PaneIssues: 0, PaneHabits: 0},
		filters: map[Pane]string{},
		allIssues: []api.IssueWithMeta{
			{Issue: issue},
			{Issue: api.Issue{ID: 2, Title: "Other", Status: "planned", TodoForDate: &date}},
		},
		summarySnapshot: &api.SummarySnapshot{
			Date:    date,
			Focus:   &api.FocusScoreSummary{Score: 72},
			Metrics: &api.DailyMetricsDay{Date: date},
		},
	}
}

func pressKey(m Model, key rune) Model {
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
	return next.(Model)
}

func TestDaySummaryKeyTogglesOnWideAndOpensDialogWhenNarrow(t *testing.T) {
	wide := pressKey(daySummaryModel(200), 'S')
	if !wide.dailyDetailDay || wide.dialog != "" {
		t.Fatalf(
			"expected S to show the Day summary in the pane, got day=%v dialog=%q",
			wide.dailyDetailDay,
			wide.dialog,
		)
	}
	if back := pressKey(wide, 'S'); back.dailyDetailDay {
		t.Fatal("expected a second S to return to the selection")
	}
	narrow := pressKey(daySummaryModel(120), 'S')
	if narrow.dailyDetailDay || narrow.dialog != "view_entity" || narrow.dialogViewTitle != "Day" {
		t.Fatalf(
			"expected a Day dialog when narrow, got day=%v dialog=%q title=%q",
			narrow.dailyDetailDay,
			narrow.dialog,
			narrow.dialogViewTitle,
		)
	}
	if body := ansi.Strip(narrow.dialogViewBody); !strings.Contains(
		strings.Join(strings.Fields(body), " "),
		"Focus 72/100",
	) {
		t.Fatalf("expected the Day summary in the dialog, got %q", body)
	}
}

func TestCursorMoveReturnsDetailToSelection(t *testing.T) {
	m := pressKey(daySummaryModel(200), 'S')
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if moved := next.(Model); moved.dailyDetailDay || moved.cursor[PaneIssues] != 1 {
		t.Fatal("expected moving the cursor to return the pane to the selection")
	}
}
