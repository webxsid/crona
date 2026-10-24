package issues

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	chromepkg "crona/tui/internal/tui/chrome"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func issuesState(width, height int) types.ContentState {
	repo, stream := int64(1), int64(2)
	yesterday, today := "2000-01-01", "2999-01-01"
	issue := func(id int64, title, status string, due *string) api.IssueWithMeta {
		return api.IssueWithMeta{
			Issue: api.Issue{
				ID:              id,
				Title:           title,
				Status:          sharedtypes.IssueStatus(status),
				StreamID:        2,
				TodoForDate:     due,
				EstimateMinutes: new(60),
			},
			RepoID: 1, RepoName: "crona", StreamName: "tui",
		}
	}
	issues := []api.IssueWithMeta{
		issue(1, "Fix sync conflict", "in_progress", &yesterday),
		issue(2, "Write release notes", "ready", &yesterday),
		issue(3, "Plan errands", "planned", &today),
		issue(4, "Old cleanup", "blocked", nil),
		issue(5, "Shipped thing", "done", nil),
	}
	return types.ContentState{
		View: "default", Pane: "issues", Width: width, Height: height,
		DefaultIssueSection: "open",
		Cursors:             map[string]int{"issues": 0, "scope": 2},
		Filters:             map[string]string{"issues": ""},
		Repos:               []api.Repo{{ID: 1, Name: "crona"}},
		AllStreams:          []api.Stream{{ID: 2, RepoID: 1, Name: "tui"}},
		Context:             &api.ActiveContext{RepoID: &repo, StreamID: &stream},
		DefaultIssues:       issues,
		AllIssues:           issues,
	}
}

func plain(s string) []string { return strings.Split(ansi.Strip(s), "\n") }

func squash(s string) string { return strings.Join(strings.Fields(s), " ") }

func TestIssuesStripTwoRows(t *testing.T) {
	for _, width := range []int{60, 100, 160} {
		rows := StripRows(types.Theme{}, issuesState(width, 40), width)
		if len(rows) != 2 {
			t.Fatalf("width %d: expected 2 strip rows, got %d", width, len(rows))
		}
		for i, row := range rows {
			if w := ansi.StringWidth(row); w != width {
				t.Fatalf("width %d: row %d is %d cells: %q", width, i, w, ansi.Strip(row))
			}
		}
	}
	rows := StripRows(types.Theme{}, issuesState(160, 40), 160)
	top, bottom := squash(ansi.Strip(rows[0])), squash(ansi.Strip(rows[1]))
	for _, want := range []string{"Open 4", "in progress 1", "blocked 1", "Due today 0", "overdue 2"} {
		if !strings.Contains(top, want) {
			t.Fatalf("expected %q in %q", want, top)
		}
	}
	for _, want := range []string{"Closed 1", "done 1", "abandoned 0", "Est. 4h over 4 issues"} {
		if !strings.Contains(bottom, want) {
			t.Fatalf("expected %q in %q", want, bottom)
		}
	}
	if narrow := ansi.Strip(strings.Join(StripRows(types.Theme{}, issuesState(60, 40), 60), "\n")); strings.Contains(
		narrow,
		"Due",
	) {
		t.Fatalf("expected the Due column dropped at 60, got %q", narrow)
	}
}

func TestIssuesListSelectorAndCreateLine(t *testing.T) {
	rows := plain(Render(types.Theme{}, issuesState(120, 40)))
	joined := strings.Join(rows, "\n")
	if !strings.Contains(joined, "Open Done") && !strings.Contains(squash(joined), "Open Done") {
		t.Fatalf("expected the Open/Done selector, got %q", joined)
	}
	if !strings.Contains(squash(joined), "[a] new issue [A] ad hoc") {
		t.Fatalf("expected the create line, got %q", joined)
	}
	running := issuesState(120, 40)
	running.Timer = &api.TimerState{State: "running"}
	if strings.Contains(ansi.Strip(Render(types.Theme{}, running)), "[A] ad hoc") {
		t.Fatal("expected [A] hidden while a timer runs")
	}
}

func cardSecondLine(t *testing.T, rendered, title string) string {
	t.Helper()
	rows := plain(rendered)
	for i, row := range rows {
		if strings.Contains(row, title) {
			return rows[i+1]
		}
	}
	t.Fatalf("%q not found in %q", title, rows)
	return ""
}

func TestIssuesCardActionLine(t *testing.T) {
	focused := Render(types.Theme{}, issuesState(120, 40))
	line := cardSecondLine(t, focused, "Fix sync conflict")
	for _, want := range []string{"[s] status", "[d] due", "[e] edit"} {
		if !strings.Contains(line, want) {
			t.Fatalf("expected %q on the selected card, got %q", want, line)
		}
	}
	if !strings.Contains(cardSecondLine(t, focused, "Write release notes"), "crona") {
		t.Fatal("expected other cards to keep their meta")
	}
	unfocused := issuesState(120, 40)
	unfocused.Pane = "scope"
	if line := cardSecondLine(t, Render(types.Theme{}, unfocused), "Fix sync conflict"); strings.Contains(
		line,
		"[s]",
	) {
		t.Fatalf("expected meta when unfocused, got %q", line)
	}
	running := issuesState(120, 40)
	running.Timer = &api.TimerState{State: "running"}
	line = cardSecondLine(t, Render(types.Theme{}, running), "Fix sync conflict")
	for _, hidden := range []string{"[f]", "[m]", "[e]", "[D]"} {
		if strings.Contains(line, hidden) {
			t.Fatalf("expected %q hidden while a timer runs, got %q", hidden, line)
		}
	}
}

func TestIssuesLayoutByWidth(t *testing.T) {
	wide := strings.Join(plain(Render(types.Theme{}, issuesState(200, 40))), "\n")
	if !strings.Contains(wide, "─ Scope") || !strings.Contains(wide, "─ Issues") ||
		!strings.Contains(wide, "─ Fix sync conflict") {
		t.Fatalf("expected scope, list and detail at 200, got %q", wide)
	}
	medium := strings.Join(plain(Render(types.Theme{}, issuesState(120, 40))), "\n")
	if !strings.Contains(medium, "─ Scope") || strings.Contains(medium, "─ Fix sync conflict") {
		t.Fatalf("expected scope and list only at 120, got %q", medium)
	}
	narrow := strings.Join(plain(Render(types.Theme{}, issuesState(80, 40))), "\n")
	if strings.Contains(narrow, "─ Scope") || !strings.Contains(narrow, "Scope ‹ crona › tui ›") {
		t.Fatalf("expected the one-row scope at 80, got %q", narrow)
	}
	for _, width := range []int{60, 80, 120, 200} {
		state := issuesState(width, 30)
		lines := plain(Render(types.Theme{}, state))
		if len(lines) != 30 {
			t.Fatalf("width %d: expected 30 rows, got %d", width, len(lines))
		}
		for i, line := range lines {
			if w := ansi.StringWidth(line); w != width {
				t.Fatalf("width %d: row %d is %d cells: %q", width, i, w, line)
			}
		}
	}
}

func TestDoneSectionListsOnlyClosed(t *testing.T) {
	state := issuesState(120, 40)
	state.DefaultIssueSection = "completed"
	rendered := ansi.Strip(Render(types.Theme{}, state))
	if !strings.Contains(rendered, "Shipped thing") ||
		strings.Contains(rendered, "Fix sync conflict") {
		t.Fatalf("expected only closed issues on Done, got %q", rendered)
	}
}

func TestIssuesCardActionsWithAppStyles(t *testing.T) {
	theme := types.Theme{StyleHeader: chromepkg.StyleHeader, StyleDim: chromepkg.StyleDim}
	line := cardSecondLine(t, Render(theme, issuesState(80, 24)), "Fix sync conflict")
	if !strings.Contains(line, "[enter] open") || !strings.Contains(line, "[s] status") {
		t.Fatalf("app styles: action row %q", line)
	}
}

func TestIssuesListHintsRequireListFocus(t *testing.T) {
	for _, height := range []int{6, 40} {
		state := issuesState(120, height)
		state.Pane = "scope"
		rendered := ansi.Strip(renderIssues(types.Theme{}, state, 93, height))
		for _, hint := range []string{"[a]", "[A]", "[←/→]"} {
			if strings.Contains(rendered, hint) {
				t.Errorf("unfocused list height %d shows %s: %q", height, hint, rendered)
			}
		}
		if !strings.Contains(squash(rendered), "Open Done") {
			t.Errorf("unfocused selector missing at height %d: %q", height, rendered)
		}
	}
}
