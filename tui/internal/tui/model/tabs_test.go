package model

import (
	"slices"
	"strings"
	"testing"
	"time"

	versionpkg "crona/shared/version"
	"crona/tui/internal/api"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestVisibleTabsWhenIdle(t *testing.T) {
	m := Model{view: ViewDaily}
	if got := m.visibleTabs(); !slices.Equal(got, uistate.TabOrder()) {
		t.Fatalf("visibleTabs() = %v, want all tabs", got)
	}
	want := []View{ViewDaily, ViewSummary}
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
	m := Model{view: ViewWellbeing}
	if got := m.nextSubTabView(1); got != ViewRollup {
		t.Fatalf("nextSubTabView(1) = %q, want rollup", got)
	}
	if got := m.nextSubTabView(-1); got != ViewMomentum {
		t.Fatalf("nextSubTabView(-1) = %q, want momentum", got)
	}
}

func TestTabTargetViewKeepsCurrentViewInTab(t *testing.T) {
	m := Model{view: ViewWellbeing}
	if got, ok := m.tabTargetView(uistate.TabInsights); !ok || got != ViewWellbeing {
		t.Fatalf("tabTargetView(insights) = %q,%t, want wellbeing", got, ok)
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
