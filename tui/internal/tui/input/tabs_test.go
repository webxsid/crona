package input

import (
	"testing"

	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func runeKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func TestDigitSwitchesTab(t *testing.T) {
	var requested uistate.Tab
	deps := Deps{
		DefaultPane: uistate.DefaultPane,
		TabTargetView: func(_ State, tab uistate.Tab) (uistate.View, bool) {
			requested = tab
			return uistate.ViewReview, true
		},
	}
	next, _ := Handle(
		State{ActiveView: uistate.ViewDaily, ActivePane: uistate.PaneIssues},
		runeKey('3'),
		deps,
	)
	if requested != uistate.TabInsights {
		t.Fatalf("expected 3 to request insights, got %q", requested)
	}
	if next.ActiveView != uistate.ViewReview || next.ActivePane != uistate.PaneReviewDays {
		t.Fatalf("expected rollup/rollup_days, got %q/%q", next.ActiveView, next.ActivePane)
	}
}

func TestDigitForHiddenTabIsNoop(t *testing.T) {
	deps := Deps{
		DefaultPane: uistate.DefaultPane,
		TabTargetView: func(State, uistate.Tab) (uistate.View, bool) {
			return "", false
		},
	}
	start := State{ActiveView: uistate.ViewSessionActive, ActivePane: uistate.PaneIssues}
	next, _ := Handle(start, runeKey('3'), deps)
	if next.ActiveView != uistate.ViewSessionActive {
		t.Fatalf("expected hidden tab to keep view, got %q", next.ActiveView)
	}
}

func TestBracketCyclesSubTab(t *testing.T) {
	var gotDir int
	deps := Deps{
		DefaultPane: uistate.DefaultPane,
		NextSubTabView: func(_ State, dir int) uistate.View {
			gotDir = dir
			return uistate.ViewMomentum
		},
	}
	next, _ := Handle(
		State{ActiveView: uistate.ViewReview, ActivePane: uistate.PaneReviewDays},
		runeKey(']'),
		deps,
	)
	if gotDir != 1 || next.ActiveView != uistate.ViewMomentum {
		t.Fatalf(
			"expected ] to move to momentum with dir 1, got %q dir %d",
			next.ActiveView,
			gotDir,
		)
	}
}

func TestIssuesDigitsNoLongerSwitchSection(t *testing.T) {
	var requested uistate.Tab
	sectionChanged := false
	deps := Deps{
		DefaultPane: uistate.DefaultPane,
		TabTargetView: func(_ State, tab uistate.Tab) (uistate.View, bool) {
			requested = tab
			return uistate.ViewDefault, true
		},
		SetDefaultIssueSection: func(*State, uistate.DefaultIssueSection) { sectionChanged = true },
	}
	start := State{
		ActiveView:          uistate.ViewDefault,
		ActivePane:          uistate.PaneIssues,
		DefaultIssueSection: uistate.DefaultIssueSectionOpen,
	}
	_, _ = Handle(start, runeKey('2'), deps)
	if requested != uistate.TabIssues {
		t.Fatalf("expected 2 to request issues tab, got %q", requested)
	}
	if sectionChanged {
		t.Fatal("expected 2 not to switch issue section")
	}
}
