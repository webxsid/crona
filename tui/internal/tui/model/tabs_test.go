package model

import (
	"slices"
	"strings"
	"testing"
	"time"

	versionpkg "crona/shared/version"
	"crona/tui/internal/api"
	layoutpkg "crona/tui/internal/tui/layout"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestVisibleTabsWhenIdle(t *testing.T) {
	m := Model{view: ViewDaily}
	if got := m.visibleTabs(); !slices.Equal(got, uistate.TabOrder()) {
		t.Fatalf("visibleTabs() = %v, want all tabs", got)
	}
	want := []View{ViewDaily}
	if got := m.visibleTabViews(uistate.TabToday); !slices.Equal(got, want) {
		t.Fatalf("visibleTabViews(today) = %v, want %v", got, want)
	}
}

func TestTabKeyIgnoresLockedTab(t *testing.T) {
	m := Model{view: ViewSessionActive, timer: &api.TimerState{State: "running"}}
	want := []uistate.Tab{uistate.TabToday, uistate.TabLog}
	if got := m.visibleTabs(); !slices.Equal(got, want) {
		t.Fatalf("visibleTabs() = %v, want %v", got, want)
	}
	if view, ok := m.tabTargetView(uistate.TabInsights); ok {
		t.Fatalf("expected insights to be locked during a timer, got %q", view)
	}
}

func TestProtectedRestTabViews(t *testing.T) {
	m := Model{
		view:     ViewReports,
		settings: &api.CoreSettings{RestSpecificDates: []string{time.Now().Format("2006-01-02")}},
	}
	wantTabs := []uistate.Tab{uistate.TabToday, uistate.TabLog, uistate.TabSystem}
	if got := m.visibleTabs(); !slices.Equal(got, wantTabs) {
		t.Fatalf("visibleTabs() = %v, want %v", got, wantTabs)
	}
	wantSystem := []View{ViewSettings, ViewReports}
	if got := m.visibleTabViews(uistate.TabSystem); !slices.Equal(got, wantSystem) {
		t.Fatalf("visibleTabViews(system) = %v, want %v", got, wantSystem)
	}
	if got := m.nextSubTabView(1); got != ViewSettings {
		t.Fatalf("nextSubTabView(1) from reports = %q, want settings", got)
	}
}

func TestNextSubTabViewWraps(t *testing.T) {
	m := Model{view: ViewMomentum}
	if got := m.nextSubTabView(1); got != ViewReview {
		t.Fatalf("nextSubTabView(1) = %q, want review", got)
	}
	if got := m.nextSubTabView(-1); got != ViewReview {
		t.Fatalf("nextSubTabView(-1) = %q, want review", got)
	}
}

func TestTabTargetViewKeepsCurrentViewInTab(t *testing.T) {
	m := Model{view: ViewMomentum}
	if got, ok := m.tabTargetView(uistate.TabInsights); !ok || got != ViewMomentum {
		t.Fatalf("tabTargetView(insights) = %q,%t, want momentum", got, ok)
	}
	if got, ok := m.tabTargetView(uistate.TabLog); !ok || got != ViewSessionHistory {
		t.Fatalf("tabTargetView(log) = %q,%t, want session_history", got, ok)
	}
}

func TestDigitsInDialogDoNotSwitchTabs(t *testing.T) {
	m := Model{
		view:    ViewDaily,
		pane:    PaneIssues,
		cursor:  map[Pane]int{},
		filters: map[Pane]string{},
	}
	m = m.withDialogState(m.dialogSnapshot().OpenCreateIssueDefault())
	next, _ := m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	got := next.(Model)
	if got.view != ViewDaily {
		t.Fatalf("expected digit in dialog to keep view daily, got %q", got.view)
	}
	if got.dialog == "" {
		t.Fatal("expected dialog to stay open")
	}
}

func TestHelpDialogShowsVersion(t *testing.T) {
	m := Model{view: ViewDaily, pane: PaneIssues, cursor: map[Pane]int{}, filters: map[Pane]string{}}
	m = m.openHelpDialog()
	if !strings.Contains(m.dialogState().ViewMeta, "Version "+versionpkg.Current()) {
		t.Fatalf("expected help meta to include version, got %q", m.dialogState().ViewMeta)
	}
}

func TestRestModeChromeUsesAwayView(t *testing.T) {
	m := Model{
		view:     ViewAlerts,
		pane:     PaneAlerts,
		width:    120,
		height:   36,
		cursor:   map[Pane]int{},
		filters:  map[Pane]string{},
		settings: &api.CoreSettings{RestSpecificDates: []string{time.Now().Format("2006-01-02")}},
	}
	state := m.layoutState()
	if state.View != ViewAway {
		t.Fatalf("expected away view during rest, got %q", state.View)
	}
	if slices.Contains(state.SubTabs, ViewSettings) {
		t.Fatalf("expected away sub-tabs, got system sub-tabs %v", state.SubTabs)
	}
	if got, want := m.contentHeight(), layoutpkg.ContentHeight(state); got != want {
		t.Fatalf("contentHeight() = %d, layout content height = %d", got, want)
	}
}

func TestDailyContentStateCarriesModelFilterIndices(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	m := Model{
		view:    ViewDaily,
		pane:    PaneIssues,
		width:   120,
		height:  36,
		cursor:  map[Pane]int{},
		filters: map[Pane]string{PaneIssues: "sync"},
		allIssues: []api.IssueWithMeta{
			{Issue: api.Issue{ID: 1, Title: "Write docs", Status: "ready", TodoForDate: &today}},
			{Issue: api.Issue{ID: 2, Title: "Fix sync bug", Status: "ready", TodoForDate: &today}},
		},
	}
	state := m.layoutState().ContentState
	if !slices.Equal(state.DailyIssueIndices, []int{1}) {
		t.Fatalf("expected the model's filtered task indices, got %v (issues %d)", state.DailyIssueIndices, len(state.DailyIssues))
	}
}

func TestContentStateSharesOneDate(t *testing.T) {
	for _, view := range []View{ViewDaily, ViewMomentum, ViewReview} {
		m := Model{view: view, width: 120, height: 36, cursor: map[Pane]int{}, filters: map[Pane]string{}, date: "2026-05-01"}
		state := m.layoutState().ContentState
		for name, got := range map[string]string{
			"DashboardDate": state.DashboardDate,
			"SummaryDate":   state.SummaryDate,
			"MomentumDate":  state.MomentumDate,
			"WellbeingDate": state.WellbeingDate,
			"RollupEndDate": state.RollupEndDate,
		} {
			if got != "2026-05-01" {
				t.Fatalf("%s: expected %s to be the shared date, got %q", view, name, got)
			}
		}
		if state.RollupStartDate != "2026-04-25" {
			t.Fatalf("%s: expected a 7-day Review window ending on the shared date, got start %q", view, state.RollupStartDate)
		}
	}
}

func TestContentStatePlannedIssuesIgnoreTab(t *testing.T) {
	date := "2026-05-01"
	past := "2026-04-20"
	issues := []api.IssueWithMeta{
		{Issue: api.Issue{ID: 1, Title: "Today", Status: "planned", TodoForDate: &date}},
		{Issue: api.Issue{ID: 2, Title: "Late", Status: "planned", TodoForDate: &past}},
	}
	m := Model{view: ViewDaily, pane: PaneIssues, width: 120, height: 36, cursor: map[Pane]int{}, filters: map[Pane]string{},
		date: date, allIssues: issues, dailyTaskSection: DailyTaskSectionOverdue}
	state := m.layoutState().ContentState
	if len(state.DailyIssues) != 1 || state.DailyIssues[0].ID != 2 {
		t.Fatalf("expected the Overdue tab's list, got %+v", state.DailyIssues)
	}
	if len(state.DailyPlannedIssues) != 1 || state.DailyPlannedIssues[0].ID != 1 {
		t.Fatalf("expected Planned issues kept for the strip, got %+v", state.DailyPlannedIssues)
	}
}
