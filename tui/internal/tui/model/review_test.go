package model

import (
	"strings"
	"testing"

	uistate "crona/tui/internal/tui/state"
	viewtypes "crona/tui/internal/tui/views/types"
)

func reviewModel() Model {
	return Model{view: ViewReview, pane: PaneReviewDays, date: "2026-10-14", width: 200, height: 40, cursor: map[Pane]int{}, filters: map[Pane]string{}}
}

func reviewContent(m Model) viewtypes.ContentState {
	snapshot := m.selectionSnapshot()
	return m.viewContentState(m.width, m.height, snapshot, nil)
}

func TestReviewShowTimeSurvivesTabSwitch(t *testing.T) {
	m := reviewModel()
	m.reviewShowTime = true
	s := m.inputState()
	s.ActiveView, s.ActivePane = uistate.ViewDaily, uistate.PaneIssues
	m = m.applyInputState(s)
	s = m.inputState()
	s.ActiveView, s.ActivePane = uistate.ViewReview, uistate.PaneReviewDays
	m = m.applyInputState(s)
	if !m.reviewShowTime || !reviewContent(m).ReviewShowTime {
		t.Fatalf("expected Time kept after leaving and re-entering Review")
	}
}

func TestReviewWindowCarriedWhenDateLeaves(t *testing.T) {
	m := reviewModel()
	m.reviewWindowEnd = "2026-10-14"
	s := m.inputState()
	s.Date = "2026-10-20"
	m = m.applyInputState(s)
	if m.reviewWindowEnd != "2026-10-20" {
		t.Fatalf("expected the window to end on the new date, got %q", m.reviewWindowEnd)
	}
	s = m.inputState()
	s.Date = "2026-10-19"
	m = m.applyInputState(s)
	if m.reviewWindowEnd != "2026-10-20" {
		t.Fatalf("expected a date inside the window to keep it, got %q", m.reviewWindowEnd)
	}
}

func TestReviewContentStateUsesOneRange(t *testing.T) {
	m := reviewModel()
	m.reviewWindowDays, m.reviewWindowEnd = 30, "2026-10-20"
	state := reviewContent(m)
	if state.ReviewWindowDays != 30 || state.ReviewWindowEnd != "2026-10-20" || state.RollupStartDate != "2026-09-21" || state.RollupEndDate != "2026-10-20" {
		t.Fatalf("expected one 30-day window ending 2026-10-20, got %d %s %s..%s", state.ReviewWindowDays, state.ReviewWindowEnd, state.RollupStartDate, state.RollupEndDate)
	}
}

func TestReviewBottomRowGlobalsOnly(t *testing.T) {
	row := bottomRow(reviewModel())
	for _, want := range []string{"[w] check-in", "[W] away", "[?] all keys"} {
		if !strings.Contains(row, want) {
			t.Fatalf("expected %q on the bottom row, got %q", want, row)
		}
	}
	for _, unwanted := range []string{"[enter]", "[,/.]", "[+/-/0]", "[S]"} {
		if strings.Contains(row, unwanted) {
			t.Fatalf("expected %q off the bottom row, got %q", unwanted, row)
		}
	}
}
