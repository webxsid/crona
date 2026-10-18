package model

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

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

func pressReview(m Model, key tea.KeyMsg) Model {
	next, _ := m.Update(key)
	return next.(Model)
}

func TestNarrowReviewEnterOpensDayDialog(t *testing.T) {
	m := reviewModel()
	m.width = 80
	m = pressReview(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.dialogState().Kind != "view_entity" || !strings.Contains(m.dialogState().ViewTitle, "Day") {
		t.Fatalf("expected the day dialog, got %q %q", m.dialogState().Kind, m.dialogState().ViewTitle)
	}
	wide := pressReview(reviewModel(), tea.KeyMsg{Type: tea.KeyEnter})
	if wide.dialogState().Kind != "" {
		t.Fatalf("expected no dialog on enter when the detail is visible, got %q", wide.dialogState().Kind)
	}
}

func TestNarrowReviewSOpensTimeDialog(t *testing.T) {
	m := reviewModel()
	m.width = 80
	m = pressReview(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	if m.dialogState().Kind != "view_entity" || !strings.Contains(m.dialogState().ViewTitle, "Time") || m.reviewShowTime {
		t.Fatalf("expected the Time dialog without toggling, got %q %q show=%t", m.dialogState().Kind, m.dialogState().ViewTitle, m.reviewShowTime)
	}
}

func TestReviewDialogSurvivesResize(t *testing.T) {
	m := reviewModel()
	m.width = 80
	m = pressReview(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 40})
	m = next.(Model)
	if lines := strings.Split(m.View(), "\n"); len(lines) > 40 {
		t.Fatalf("expected the screen to fit after resize, got %d lines", len(lines))
	}
	m = pressReview(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.dialogState().Kind != "" {
		t.Fatalf("expected esc to close the dialog after resize")
	}
}

func TestHelpListsReviewKeys(t *testing.T) {
	body := ansi.Strip(reviewModel().openHelpDialog().dialogState().ViewBody)
	for _, want := range []string{"[S] trends/time", "[w] check-in", "[+/-/0] window", "[↑/↓] day"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected the help sheet to list %q", want)
		}
	}
}
