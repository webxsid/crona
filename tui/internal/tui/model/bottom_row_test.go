package model

import (
	"strings"
	"testing"

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
	for _, want := range []string{"Navigation", "[1-5] tabs", "[[/]] sub-tabs", "[,/.] date", "[g] today", "[S] day", "[C] calendar"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected the help sheet to list %q, got %q", want, body)
		}
	}
}
