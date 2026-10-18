package dialogs

import (
	"regexp"
	"strings"
	"testing"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/x/ansi"
)

type formCase struct {
	name       string
	state      controllerpkg.State
	title      string
	context    string
	focusCount int
}

var bracketKey = regexp.MustCompile(`\[[a-z+/←→↑↓]+\]`)

func sized(state controllerpkg.State, width, height int) controllerpkg.State {
	state.Width, state.Height = width, height
	return state
}

func assertFormChrome(t *testing.T, c formCase) {
	t.Helper()
	out := ansi.Strip(Render(testTheme(), sized(c.state, 120, 36)))
	lines := strings.Split(out, "\n")
	if !strings.HasPrefix(lines[0], "╭─ "+c.title) {
		t.Errorf("%s: top border %q lacks title %q", c.name, lines[0], c.title)
	}
	if c.context != "" && !strings.Contains(lines[0], c.context) {
		t.Errorf("%s: top border %q lacks context %q", c.name, lines[0], c.context)
	}
	if strings.Contains(out, "(Optional)") || bracketKey.MatchString(out) {
		t.Errorf("%s: old chrome remains:\n%s", c.name, out)
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
}

func assertHeightStable(t *testing.T, c formCase) {
	t.Helper()
	height := -1
	for focus := 0; focus < c.focusCount; focus++ {
		state := sized(c.state, 120, 36)
		state.FocusIdx = focus
		state = controllerpkg.SyncDialogFocus(state)
		got := len(strings.Split(Render(testTheme(), state), "\n"))
		if height >= 0 && got != height {
			t.Errorf("%s: focus %d height %d, want %d", c.name, focus, got, height)
		}
		height = got
	}
}

func TestRepoStreamForms(t *testing.T) {
	cases := []formCase{
		{"create_repo", controllerpkg.OpenCreateRepo(controllerpkg.State{}), "New Repo", "", 2},
		{"edit_repo", controllerpkg.OpenEditRepo(controllerpkg.State{}, 1, "OSS", nil), "Edit Repo", "", 2},
		{"create_stream", controllerpkg.OpenCreateStream(controllerpkg.State{}, 1, "OSS"), "New Stream", "OSS", 2},
		{"edit_stream", controllerpkg.OpenEditStream(controllerpkg.State{}, 2, 1, "cli", "OSS", nil), "Edit Stream", "OSS", 2},
	}
	for _, c := range cases {
		assertFormChrome(t, c)
		assertHeightStable(t, c)
	}
}

func habitEditState() controllerpkg.State {
	target := 20
	return controllerpkg.OpenEditHabit(controllerpkg.State{}, 3, 2, "Read", nil, "daily", &target, true)
}

func TestHabitAndEditIssueForms(t *testing.T) {
	create := controllerpkg.OpenCreateHabit(controllerpkg.State{})
	create.RepoName, create.StreamName = "OSS", "cli"
	cases := []formCase{
		{"create_habit", create, "New Habit", "OSS › cli", 6},
		{"edit_habit", habitEditState(), "Edit Habit", "", 4},
		{"edit_issue", controllerpkg.OpenEditIssue(controllerpkg.State{}, 1, 2, "Fix", nil, nil, nil), "Edit Issue", "", 4},
	}
	for _, c := range cases {
		assertFormChrome(t, c)
		assertHeightStable(t, c)
	}
}

func TestEditIssueDueKeepsTodayHint(t *testing.T) {
	state := sized(controllerpkg.OpenEditIssue(controllerpkg.State{}, 1, 2, "Fix", nil, nil, nil), 120, 36)
	state.FocusIdx = 3
	out := ansi.Strip(Render(testTheme(), controllerpkg.SyncDialogFocus(state)))
	if !strings.Contains(out, "g today") {
		t.Fatalf("expected g today on Due:\n%s", out)
	}
}

func TestNewHabitShortTerminal(t *testing.T) {
	state := sized(controllerpkg.OpenCreateHabit(controllerpkg.State{}), 60, 20)
	state.FocusIdx = 5
	out := ansi.Strip(Render(testTheme(), controllerpkg.SyncDialogFocus(state)))
	if !strings.Contains(out, "Target") || !strings.Contains(out, "esc cancel") {
		t.Fatalf("expected Target and esc cancel visible:\n%s", out)
	}
}
