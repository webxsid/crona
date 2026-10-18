package dialogs

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	uistate "crona/tui/internal/tui/state"

	"github.com/charmbracelet/x/ansi"
)

// assertListChrome is assertFormChrome for dialogs whose lists may show [k] item keys:
// it checks the frame title, the context, and a `key action` footer ending esc cancel.
func assertListChrome(t *testing.T, c formCase) string {
	t.Helper()
	out := ansi.Strip(Render(testTheme(), sized(c.state, 120, 36)))
	lines := strings.Split(out, "\n")
	if !strings.HasPrefix(lines[0], "╭─ "+c.title) {
		t.Errorf("%s: top border %q lacks title %q", c.name, lines[0], c.title)
	}
	if c.context != "" && !strings.Contains(lines[0], c.context) {
		t.Errorf("%s: top border %q lacks context %q", c.name, lines[0], c.context)
	}
	for _, old := range []string{"[esc]", "[enter]", "[↑/↓]"} {
		if strings.Contains(out, old) {
			t.Errorf("%s: old footer %q remains:\n%s", c.name, old, out)
		}
	}
	last := ""
	for i := len(lines) - 2; i > 0; i-- {
		if inside := strings.TrimSpace(strings.Trim(lines[i], "│")); inside != "" {
			last = inside
			break
		}
	}
	if !strings.HasSuffix(last, "esc cancel") {
		t.Errorf("%s: footer ends %q, want esc cancel", c.name, last)
	}
	return out
}

func TestConfirmAndListDialogs(t *testing.T) {
	status := controllerpkg.OpenIssueStatus(controllerpkg.State{}, "in_progress", "Fix alerts", "OSS", "cli", nil, 0)
	cases := []formCase{
		{name: "confirm_delete", state: controllerpkg.OpenConfirmDelete(controllerpkg.State{}, "issue", "1", "Fix alerts", 1, 2), title: "Delete Issue"},
		{name: "confirm_wipe", state: controllerpkg.OpenConfirmWipeData(controllerpkg.State{}), title: "Wipe All Data"},
		{name: "issue_status", state: status, title: "Set Issue Status", context: "OSS › cli"},
		{name: "view_jump", state: controllerpkg.OpenViewJump(controllerpkg.State{}, []uistate.View{uistate.ViewDaily, uistate.ViewDefault}), title: "Jump To View"},
		{name: "beta_support", state: controllerpkg.OpenBetaSupport(controllerpkg.State{}), title: "Beta Support"},
	}
	for _, c := range cases {
		out := assertListChrome(t, c)
		if c.name == "confirm_delete" && !strings.Contains(out, "enter delete") {
			t.Errorf("confirm_delete footer lacks enter delete:\n%s", out)
		}
		if c.name == "issue_status" && !strings.Contains(out, "▶") {
			t.Errorf("issue_status lacks a cursor on its list:\n%s", out)
		}
	}
}

func TestDateGridDialogs(t *testing.T) {
	due := "2026-10-20"
	pick := controllerpkg.OpenDatePicker(controllerpkg.State{}, "create_issue_default", 0, 5, &due, "2026-10-18")
	pick = controllerpkg.PopulateDatePresentation(controllerpkg.Theme{}, pick, "2026-10-18")
	summary := controllerpkg.OpenSummaryCalendar(controllerpkg.State{}, "2026-10-18")
	summary.Width = 120
	summary = controllerpkg.PopulateSummaryCalendarPresentation(controllerpkg.Theme{}, summary, "2026-10-18")
	cases := []formCase{
		{name: "pick_date", state: pick, title: "Pick Due Date", context: "October 2026"},
		{name: "summary_calendar", state: summary, title: "Summary Calendar", context: "October 2026"},
	}
	for _, c := range cases {
		out := assertListChrome(t, c)
		for _, want := range []string{"Mo", "enter choose", ",/. month"} {
			if !strings.Contains(out, want) {
				t.Errorf("%s lacks %q:\n%s", c.name, want, out)
			}
		}
		if strings.Contains(out, "(Optional)") {
			t.Errorf("%s keeps (Optional) in its title:\n%s", c.name, out)
		}
	}
}

func adhocPicker(tab int) controllerpkg.State {
	today, past := "2026-10-18", "2026-10-10"
	issues := []api.IssueWithMeta{
		{Issue: api.Issue{ID: 1, Title: "Fix alerts", Status: sharedtypes.IssueStatusInProgress, TodoForDate: &today}, RepoName: "OSS", StreamName: "cli"},
		{Issue: api.Issue{ID: 2, Title: "Write docs", Status: sharedtypes.IssueStatusInProgress, TodoForDate: &today}, RepoName: "OSS", StreamName: "docs"},
		{Issue: api.Issue{ID: 3, Title: "Old bug", Status: sharedtypes.IssueStatusInProgress, TodoForDate: &past}, RepoName: "Work", StreamName: "api"},
	}
	state := controllerpkg.State{Kind: "adhoc_issue_picker", AdhocIssueTab: tab}
	return controllerpkg.OpenAdhocIssuePicker(state, issues, "2026-10-18")
}

func TestAdhocIssuePickerChrome(t *testing.T) {
	out := assertListChrome(t, formCase{name: "adhoc_issue_picker", state: adhocPicker(0), title: "Attach to Issue"})
	for _, want := range []string{"Due today 2", "Overdue 1", "▶ Fix alerts", "enter continue", "←/→ tab"} {
		if !strings.Contains(out, want) {
			t.Errorf("adhoc picker lacks %q:\n%s", want, out)
		}
	}
	height := -1
	for tab := 0; tab < 4; tab++ {
		got := len(strings.Split(Render(testTheme(), sized(adhocPicker(tab), 120, 36)), "\n"))
		if height >= 0 && got != height {
			t.Errorf("tab %d height %d, want %d", tab, got, height)
		}
		height = got
	}
}
