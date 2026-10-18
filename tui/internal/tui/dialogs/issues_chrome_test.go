package dialogs

import (
	"strings"
	"testing"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/x/ansi"
)

func issueState(kind string, focus, width, height int) controllerpkg.State {
	inputs := make([]textinput.Model, 6)
	for i := range inputs {
		inputs[i] = textinput.New()
	}
	return controllerpkg.State{
		Kind: kind, Width: width, Height: height, FocusIdx: focus,
		Inputs: inputs, Description: textarea.New(),
		RepoName: "OSS", StreamName: "cli",
		RepoSelectorLabel: "OSS", StreamSelectorLabel: "cli",
		RepoOptionLabels: []string{"OSS", "Work", "Home"}, StreamOptionLabels: []string{"cli"},
	}
}

func issueText(state controllerpkg.State) string {
	return ansi.Strip(renderIssueDialog(testTheme(), state))
}

func TestNewIssueChrome(t *testing.T) {
	text := issueText(issueState("create_issue_default", 2, 120, 40))
	for _, unwanted := range []string{"[ > ", "⌕", "Description (Optional)"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("expected no %q, got\n%s", unwanted, text)
		}
	}
	if !strings.Contains(text, "╭─ New Issue") || !strings.Contains(text, "ctrl+s save") || !strings.Contains(text, "Title") {
		t.Fatalf("expected the new chrome, got\n%s", text)
	}
	paired := false
	for _, line := range strings.Split(text, "\n") {
		paired = paired || (strings.Contains(line, "Estimate") && strings.Contains(line, "Due"))
	}
	if !paired {
		t.Fatalf("expected Estimate and Due on one line, got\n%s", text)
	}
}

func TestNewIssuePickersAtBottom(t *testing.T) {
	text := issueText(issueState("create_issue_default", 0, 120, 40))
	estimate, repo := strings.Index(text, "Estimate"), strings.Index(text, "‹ OSS ›")
	if estimate < 0 || repo < estimate {
		t.Fatalf("expected the repo picker below Estimate, got\n%s", text)
	}
}

func TestNewIssueHeightIsStable(t *testing.T) {
	height := -1
	for focus := 0; focus <= 5; focus++ {
		n := len(strings.Split(issueText(issueState("create_issue_default", focus, 120, 40)), "\n"))
		if height >= 0 && n != height {
			t.Fatalf("focus %d: height %d, expected %d", focus, n, height)
		}
		height = n
	}
}

func TestAdhocIssueKeepsSummary(t *testing.T) {
	state := issueState("create_issue_default", 2, 120, 40)
	state.AdhocTimer, state.ViewName = true, "Ad hoc session"
	text := issueText(state)
	if !strings.Contains(text, "Create Issue & Finish Session") || !strings.Contains(text, "Summary") || !strings.Contains(text, "Ad hoc session") {
		t.Fatalf("expected the ad hoc variant, got\n%s", text)
	}
}

func TestNewIssueShortTerminal(t *testing.T) {
	state := issueState("create_issue_default", 5, 120, 20)
	text := issueText(state)
	if !strings.Contains(text, "YYYY-MM-DD") || !strings.Contains(text, "esc cancel") {
		t.Fatalf("expected the focused Due field and footer, got\n%s", text)
	}
	if n := len(strings.Split(text, "\n")); n > 16 {
		t.Fatalf("expected the dialog within the screen, got %d lines", n)
	}
}

func TestNewIssuePairFitsTypedEstimate(t *testing.T) {
	state := issueState("create_issue_default", 4, 120, 40)
	state.Inputs[3].SetValue("1h30m")
	state.Inputs[3].Focus()
	for _, line := range strings.Split(issueText(state), "\n") {
		if strings.Contains(line, "Estimate") && (strings.Contains(line, "…") || !strings.Contains(line, "1h30m")) {
			t.Fatalf("expected the estimate whole, got %q", line)
		}
	}
}

func TestNewIssueStackedPairAligns(t *testing.T) {
	text := issueText(issueState("create_issue_default", 2, 60, 40))
	var estimate, due int
	for _, line := range strings.Split(text, "\n") {
		if i := strings.Index(line, "45m"); i >= 0 {
			estimate = i
		}
		if i := strings.Index(line, "YYYY-MM-DD"); i >= 0 {
			due = i
		}
	}
	if estimate == 0 || estimate != due {
		t.Fatalf("expected stacked values aligned, got %d and %d in\n%s", estimate, due, text)
	}
}
