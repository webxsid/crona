package model

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	layoutpkg "crona/tui/internal/tui/layout"

	"github.com/charmbracelet/x/ansi"
)

func bottomRow(m Model) string {
	lines := strings.Split(ansi.Strip(layoutpkg.Render(m.layoutState())), "\n")
	return lines[len(lines)-1]
}

func TestDailyBottomRowIsGlobalsOnly(t *testing.T) {
	for _, pane := range []Pane{PaneIssues, PaneHabits} {
		m := Model{view: ViewDaily, pane: pane, width: 200, height: 36, cursor: map[Pane]int{}, filters: map[Pane]string{}}
		row := bottomRow(m)
		for _, want := range []string{"[tab] next pane", "[c] context", "[w] check-in", "[W] away", "[?] all keys"} {
			if !strings.Contains(row, want) {
				t.Fatalf("%s: expected %q on the bottom row, got %q", pane, want, row)
			}
		}
		for _, unwanted := range []string{"[s]", "[a]", "[x]", "[,/.]", "[1-5]", "[[/]]", "[S]", "[C]"} {
			if strings.Contains(row, unwanted) {
				t.Fatalf("%s: expected %q off the bottom row, got %q", pane, unwanted, row)
			}
		}
	}
}

func TestRollupBottomRowDropsDateKeys(t *testing.T) {
	m := Model{view: ViewRollup, pane: PaneRollupDays, width: 200, height: 36, cursor: map[Pane]int{}, filters: map[Pane]string{}}
	row := bottomRow(m)
	if !strings.Contains(row, "[enter]") {
		t.Fatalf("expected Rollup's own keys kept, got %q", row)
	}
	for _, unwanted := range []string{"[,/.]", "[g]", "[+/-/0]", "[E] calendar", "[1-5]"} {
		if strings.Contains(row, unwanted) {
			t.Fatalf("expected %q off the bottom row, got %q", unwanted, row)
		}
	}
}

func TestHelpSheetListsNavigationKeys(t *testing.T) {
	m := Model{view: ViewDaily, pane: PaneIssues, cursor: map[Pane]int{}, filters: map[Pane]string{}}
	body := ansi.Strip(m.openHelpDialog().dialogState().ViewBody)
	for _, want := range []string{"Navigation", "[1-5] tabs", "[[/]] sub-tabs", "[,/.] date", "[g] today", "[S] day summary", "[C] calendar"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected the help sheet to list %q, got %q", want, body)
		}
	}
}

func TestModelContentHeightCountsDateRow(t *testing.T) {
	for _, view := range []View{ViewDaily, ViewRollup, ViewMomentum, ViewWellbeing} {
		m := Model{view: view, width: 120, height: 36, cursor: map[Pane]int{}, filters: map[Pane]string{}}
		state := m.layoutState()
		want := layoutpkg.ContentHeight(state)
		if state.ContentState.Height != want || m.contentHeight() != want {
			t.Fatalf("%s: views sized %d, model %d, shell shows %d", view, state.ContentState.Height, m.contentHeight(), want)
		}
	}
}

func TestNarrowDailyBottomRowOffersDayKey(t *testing.T) {
	m := Model{view: ViewDaily, pane: PaneIssues, width: 120, height: 36, cursor: map[Pane]int{}, filters: map[Pane]string{}}
	if row := bottomRow(m); !strings.Contains(row, "[S] day summary") {
		t.Fatalf("expected [S] day summary on the bottom row without a detail pane, got %q", row)
	}
	m.width = 200
	if row := bottomRow(m); strings.Contains(row, "[S]") {
		t.Fatalf("expected [S] left to the detail pane when it shows, got %q", row)
	}
	m.width, m.date = 120, "2026-05-01"
	m.settings = &api.CoreSettings{AwayDates: []string{"2026-05-01"}}
	if m.layoutState().ContentState.HistoricalAwayDate == "" {
		t.Fatal("expected a historical away date in this fixture")
	}
	if row := bottomRow(m); strings.Contains(row, "[S]") {
		t.Fatalf("expected no [S] on a historical away date, got %q", row)
	}
}
