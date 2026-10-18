package model

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/input"
	layoutpkg "crona/tui/internal/tui/layout"
	supportview "crona/tui/internal/tui/views/support"

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
	for _, view := range []View{ViewDaily, ViewReview, ViewMomentum} {
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

func TestIssuesBottomRowIsGlobalsOnly(t *testing.T) {
	t.Parallel()
	testIssuesTabBottomRow(t, ViewDefault, PaneIssues)
}

func TestHabitsBottomRowIsGlobalsOnly(t *testing.T) {
	t.Parallel()
	testIssuesTabBottomRow(t, ViewHabitList, PaneHabits)
}

func testIssuesTabBottomRow(t *testing.T, view View, list Pane) {
	t.Helper()
	for _, pane := range []Pane{PaneScope, list} {
		for _, width := range []int{60, 120, 200} {
			m := Model{view: view, pane: pane, width: width, height: 36, cursor: map[Pane]int{}, filters: map[Pane]string{}}
			row := bottomRow(m)
			if pane == PaneScope && strings.Contains(row, "[c] context") {
				t.Errorf("bottomRow(%s, scope, %d) = %q, want context hidden", view, width, row)
			}
			if !strings.Contains(row, "[tab] next pane") {
				t.Errorf("bottomRow(%s, %s, %d) = %q, want next-pane hint", view, pane, width, row)
			}
			if !strings.Contains(row, "[?] all keys") {
				t.Errorf("bottomRow(%s, %s, %d) = %q, want all keys hint", view, pane, width, row)
			}
			for _, unwanted := range []string{"[s]", "[c] check out", "[a]", "[S]", "open/resolved"} {
				if strings.Contains(row, unwanted) {
					t.Errorf("bottomRow(%s, %s, %d) = %q, want no %q", view, pane, width, row, unwanted)
				}
			}
		}
	}
}

func TestHelpSheetListsScopeKeys(t *testing.T) {
	t.Parallel()
	for _, view := range []View{ViewDefault, ViewHabitList} {
		list := PaneIssues
		if view == ViewHabitList {
			list = PaneHabits
		}
		for _, pane := range []Pane{PaneScope, list} {
			m := Model{view: view, pane: pane, cursor: map[Pane]int{}, filters: map[Pane]string{}}
			body := ansi.Strip(m.openHelpDialog().dialogState().ViewBody)
			for _, want := range []string{"Current View", "Scope", "[c] check out", "[a] new repo", "[a] new stream", "[enter] open", "[e] edit", "[m] log"} {
				if !strings.Contains(body, want) {
					t.Errorf("openHelpDialog(%s, %s) = %q, want %q", view, pane, body, want)
				}
			}
		}
	}
}

func TestMomentumBottomRowIsGlobalsOnly(t *testing.T) {
	row := bottomRow(Model{view: ViewMomentum, pane: PaneMomentumCards, date: "2026-10-18", width: 200, height: 36, cursor: map[Pane]int{}, filters: map[Pane]string{}})
	for _, unwanted := range []string{"[e] edit", "[enter]", "[D] delete", "[x] toggle"} {
		if strings.Contains(row, unwanted) {
			t.Fatalf("expected %q off the bottom row, got %q", unwanted, row)
		}
	}
	if !strings.Contains(row, "[?] all keys") {
		t.Fatalf("expected the help key on the bottom row, got %q", row)
	}
}

func TestLogBottomRowsGlobalsOnly(t *testing.T) {
	for _, tc := range []struct {
		view View
		pane Pane
	}{{ViewSessionHistory, PaneSessions}, {ViewHabitHistory, PaneHabitHistory}, {ViewOps, PaneOps}} {
		m := Model{view: tc.view, pane: tc.pane, width: 200, height: 36, cursor: map[Pane]int{}, filters: map[Pane]string{}}
		row := bottomRow(m)
		if !strings.Contains(row, "[?] all keys") {
			t.Fatalf("%s: expected [?] all keys on the bottom row, got %q", tc.view, row)
		}
		for _, unwanted := range []string{"[enter]", "[+]", "[-]", "[+/-]"} {
			if strings.Contains(row, unwanted) {
				t.Fatalf("%s: expected %q off the bottom row, got %q", tc.view, unwanted, row)
			}
		}
	}
}

func TestSupportActionOrderMatchesInput(t *testing.T) {
	keys := []string{}
	for _, action := range supportview.Actions() {
		keys = append(keys, action.Key)
	}
	if strings.Join(keys, ",") != strings.Join(input.SupportActionKeys(), ",") {
		t.Fatalf("expected the view and input to list the same actions, got %v vs %v", keys, input.SupportActionKeys())
	}
}
