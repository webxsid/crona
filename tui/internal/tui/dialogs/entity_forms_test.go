package dialogs

import (
	"regexp"
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
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
		{
			"edit_repo",
			controllerpkg.OpenEditRepo(controllerpkg.State{}, 1, "OSS", nil),
			"Edit Repo",
			"",
			2,
		},
		{
			"create_stream",
			controllerpkg.OpenCreateStream(controllerpkg.State{}, 1, "OSS"),
			"New Stream",
			"OSS",
			2,
		},
		{
			"edit_stream",
			controllerpkg.OpenEditStream(controllerpkg.State{}, 2, 1, "cli", "OSS", nil),
			"Edit Stream",
			"OSS",
			2,
		},
	}
	for _, c := range cases {
		assertFormChrome(t, c)
		assertHeightStable(t, c)
	}
}

func habitEditState() controllerpkg.State {
	target := 20
	return controllerpkg.OpenEditHabit(
		controllerpkg.State{},
		3,
		2,
		"Read",
		nil,
		"daily",
		&target,
		true,
	)
}

func TestHabitAndEditIssueForms(t *testing.T) {
	create := controllerpkg.OpenCreateHabit(controllerpkg.State{})
	create.RepoName, create.StreamName = "OSS", "cli"
	cases := []formCase{
		{"create_habit", create, "New Habit", "OSS › cli", 7},
		{"edit_habit", habitEditState(), "Edit Habit", "", 5},
		{
			"edit_issue",
			controllerpkg.OpenEditIssue(controllerpkg.State{}, 1, 2, "Fix", nil, nil, nil),
			"Edit Issue",
			"",
			4,
		},
	}
	for _, c := range cases {
		assertFormChrome(t, c)
		assertHeightStable(t, c)
	}
}

func TestEditIssueDueKeepsTodayHint(t *testing.T) {
	state := sized(
		controllerpkg.OpenEditIssue(controllerpkg.State{}, 1, 2, "Fix", nil, nil, nil),
		120,
		36,
	)
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

func TestRemainingEntityForms(t *testing.T) {
	note := controllerpkg.OpenIssueStatusNote(
		controllerpkg.State{},
		1,
		2,
		"blocked",
		"Why is it blocked?",
		true,
	)
	note.ViewName = "Fix alerts"
	cases := []formCase{
		{"issue_status_note", note, "Block Issue", "Fix alerts", 1},
		{
			"create_checkin",
			controllerpkg.OpenCreateCheckIn(controllerpkg.State{}, "2026-10-18"),
			"New Check-In",
			"",
			6,
		},
		{
			"edit_checkin",
			controllerpkg.OpenEditCheckIn(controllerpkg.State{}, nil, "2026-10-18"),
			"Edit Check-In",
			"",
			6,
		},
		{
			"complete_habit",
			controllerpkg.OpenHabitCompletion(controllerpkg.State{}, 1, "2026-10-18", nil, nil),
			"Habit Log",
			"2026-10-18",
			2,
		},
		{
			"create_alert_reminder",
			controllerpkg.OpenCreateAlertReminder(controllerpkg.State{}, ""),
			"Add Check-In Reminder",
			"",
			3,
		},
		{
			"edit_alert_reminder",
			controllerpkg.OpenEditAlertReminder(
				controllerpkg.State{},
				sharedtypes.AlertReminder{TimeHHMM: "20:00"},
			),
			"Edit Check-In Reminder",
			"",
			3,
		},
	}
	for _, c := range cases {
		assertFormChrome(t, c)
		assertHeightStable(t, c)
	}
}

func TestCheckInShortTerminal(t *testing.T) {
	state := sized(controllerpkg.OpenCreateCheckIn(controllerpkg.State{}, "2026-10-18"), 60, 20)
	state.FocusIdx = 5
	out := ansi.Strip(Render(testTheme(), controllerpkg.SyncDialogFocus(state)))
	if !strings.Contains(out, "Notes") || !strings.Contains(out, "esc cancel") {
		t.Fatalf("expected Notes and esc cancel visible:\n%s", out)
	}
}

func TestEntityFormErrorVisible(t *testing.T) {
	state := sized(controllerpkg.OpenCreateCheckIn(controllerpkg.State{}, "2026-10-18"), 60, 20)
	state.ErrorMessage = "Mood must be 1-5"
	if out := ansi.Strip(Render(testTheme(), state)); !strings.Contains(out, "Mood must be 1-5") {
		t.Fatalf("expected the error visible:\n%s", out)
	}
}

func TestFormPairsFitTheirPlaceholders(t *testing.T) {
	habit := ansi.Strip(
		Render(testTheme(), sized(controllerpkg.OpenCreateHabit(controllerpkg.State{}), 120, 36)),
	)
	if strings.Contains(habit, "…") {
		t.Errorf("schedule placeholder truncated:\n%s", habit)
	}
	reminder := ansi.Strip(
		Render(
			testTheme(),
			sized(controllerpkg.OpenCreateAlertReminder(controllerpkg.State{}, ""), 120, 36),
		),
	)
	if !strings.Contains(reminder, "Time     20:00") {
		t.Errorf("time label not aligned with Cadence:\n%s", reminder)
	}
}

func TestCadenceRows(t *testing.T) {
	daily := ansi.Strip(
		Render(testTheme(), sized(controllerpkg.OpenCreateHabit(controllerpkg.State{}), 120, 36)),
	)
	for _, want := range []string{"Cadence", "‹ Daily ›", "Days", "Mo  Tu  We"} {
		if !strings.Contains(daily, want) {
			t.Errorf("new habit lacks %q:\n%s", want, daily)
		}
	}
	if strings.Contains(daily, "Schedule") || strings.Contains(daily, "filter:") {
		t.Errorf("new habit keeps the typed schedule:\n%s", daily)
	}
	custom := ansi.Strip(
		Render(
			testTheme(),
			sized(
				controllerpkg.OpenEditHabit(
					controllerpkg.State{},
					3,
					2,
					"Read",
					nil,
					"tue,thu",
					nil,
					true,
				),
				120,
				36,
			),
		),
	)
	if !strings.Contains(custom, "‹ Custom days ›") || !strings.Contains(custom, "[Tu]") ||
		!strings.Contains(custom, "[Th]") {
		t.Errorf("custom habit should tick Tu and Th:\n%s", custom)
	}
	reminder := ansi.Strip(
		Render(
			testTheme(),
			sized(controllerpkg.OpenCreateAlertReminder(controllerpkg.State{}, ""), 120, 36),
		),
	)
	if !strings.Contains(reminder, "‹ Daily ›") || strings.Contains(reminder, "weekday lists") {
		t.Errorf("reminder should use the cadence picker without the typing help:\n%s", reminder)
	}
}

func TestCadenceRowsStackAboveTarget(t *testing.T) {
	states := []controllerpkg.State{
		controllerpkg.OpenCreateHabit(controllerpkg.State{}),
		controllerpkg.OpenEditHabit(controllerpkg.State{}, 3, 2, "Read", nil, "daily", nil, true),
		controllerpkg.OpenCreateAlertReminder(controllerpkg.State{}, ""),
	}
	for _, state := range states {
		lines := strings.Split(ansi.Strip(Render(testTheme(), sized(state, 120, 36))), "\n")
		row := func(label string) int {
			for i, line := range lines {
				if strings.Contains(line, "  "+label+" ") {
					return i
				}
			}
			return -1
		}
		last := "Target"
		if state.Kind == "create_alert_reminder" {
			last = "Time"
		}
		cadence, days, end := row("Cadence"), row("Days"), row(last)
		if cadence < 0 || days != cadence+1 || end != days+1 {
			t.Errorf(
				"%s: Cadence %d, Days %d, %s %d; want consecutive rows\n%s",
				state.Kind,
				cadence,
				days,
				last,
				end,
				strings.Join(lines, "\n"),
			)
		}
	}
}
