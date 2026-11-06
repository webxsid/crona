package daily

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func detailState() (state types.ContentState) {
	state = dailyState(200, 40)
	state.DailyTaskSection = "planned"
	description := "Make the sync resolve conflicts by keeping both edits"
	state.DailyIssues = []api.Issue{
		{ID: 1, Title: "First task", Status: "ready", TodoForDate: new(state.DashboardDate)},
		{
			ID:              2,
			Title:           "Fix sync conflict",
			Status:          "in_progress",
			TodoForDate:     new(state.DashboardDate),
			EstimateMinutes: new(60),
			WorkedSeconds:   1800,
			Description:     &description,
			RepoName:        "crona",
			StreamName:      "core",
		},
	}
	state.Cursors["issues"] = 1
	return state
}

func TestDetailShowsSelectedTask(t *testing.T) {
	rendered := ansi.Strip(renderDetail(testTheme(), focusDetail(detailState()), 70, 30))
	for _, want := range []string{"Fix sync conflict", "in progress", "crona › core", "30m", "Effort/1h", "keeping both edits"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected detail to contain %q, got %q", want, rendered)
		}
	}
	if strings.Contains(rendered, "First task") {
		t.Fatalf("expected only the selected task, got %q", rendered)
	}
}

func TestDetailFollowsHabitsFocus(t *testing.T) {
	state := detailState()
	state.Pane = "habits"
	state.Cursors["habits"] = 1
	target := 20
	state.DueHabits[1].TargetMinutes = &target
	rendered := strings.Join(
		strings.Fields(ansi.Strip(renderDetail(testTheme(), focusDetail(state), 70, 30))),
		" ",
	)
	if !strings.Contains(rendered, "Read") {
		t.Fatalf("expected selected habit details, got %q", rendered)
	}
	if !strings.Contains(rendered, "Today/20m") {
		t.Fatalf("expected the Today arc against its 20m target, got %q", rendered)
	}
	if strings.Contains(rendered, "Fix sync conflict") {
		t.Fatalf("expected habit focus to replace task detail, got %q", rendered)
	}
}

func TestDetailEmptySelection(t *testing.T) {
	state := detailState()
	state.DailyIssues = nil
	rendered := strings.Join(
		strings.Fields(ansi.Strip(renderDetail(testTheme(), focusDetail(state), 70, 30))),
		" ",
	)
	if !strings.Contains(rendered, "Nothing selected") {
		t.Fatalf("expected empty-selection hint, got %q", rendered)
	}
}

func TestDetailHasNoCalendar(t *testing.T) {
	if rendered := ansi.Strip(renderDetail(testTheme(), focusDetail(detailState()), 70, 40)); strings.Contains(
		rendered,
		"Mo Tu We",
	) {
		t.Fatalf("expected the calendar to live in the strip, got %q", rendered)
	}
}

func TestDetailFollowsModelFilteredCursor(t *testing.T) {
	state := detailState()
	state.Filters["issues"] = "sync"
	state.DailyIssueIndices = []int{1}
	state.Cursors["issues"] = 0
	detail := ansi.Strip(renderDetail(testTheme(), focusDetail(state), 70, 30))
	if !strings.Contains(detail, "Fix sync conflict") || strings.Contains(detail, "First task") {
		t.Fatalf("expected the filtered selection in detail, got %q", detail)
	}
	list := ansi.Strip(renderIssues(testTheme(), state, 130, 20))
	assertContainsLine(t, strings.Split(list, "\n"), "▶", "Fix sync conflict")
}

func TestDetailFollowsModelFilteredHabitCursor(t *testing.T) {
	state := detailState()
	state.Pane = "habits"
	state.DailyHabitIndices = []int{1}
	state.Cursors["habits"] = 0
	detail := ansi.Strip(renderDetail(testTheme(), focusDetail(state), 70, 30))
	if !strings.Contains(detail, "Read") || strings.Contains(detail, "Walk") {
		t.Fatalf("expected the filtered habit in detail, got %q", detail)
	}
}

func TestDailyTrustsModelTaskSection(t *testing.T) {
	state := dailyState(130, 30)
	state.DailyTaskSection = "planned"
	yesterday := "2026-05-26"
	state.DailyIssues = []api.Issue{
		{
			ID:          1,
			Title:       "Closed pinned task",
			Status:      "done",
			PinnedDaily: true,
			TodoForDate: &yesterday,
		},
		{ID: 2, Title: "Open pinned task", Status: "ready", PinnedDaily: true},
	}
	state.Cursors["issues"] = 0
	if detail := ansi.Strip(renderDetail(testTheme(), focusDetail(state), 70, 30)); !strings.Contains(
		detail,
		"Closed pinned task",
	) {
		t.Fatalf("expected the model's first pinned task selected, got %q", detail)
	}
}

func TestDetailHabitListsTheMomentumsItFeeds(t *testing.T) {
	state := detailState()
	state.Pane = "habits"
	state.Cursors["habits"] = 1 // Read, habit ID 2
	state.MomentumCards = []api.MomentumCard{
		{
			Definition: api.HabitStreakDefinition{
				ID:            "fit",
				Name:          "Fitness",
				Enabled:       true,
				Period:        "week",
				RequiredCount: 3,
				HabitIDs:      []int64{1, 2},
			},
			Current: 4,
			Longest: 9,
			Series:  []api.MomentumSeriesPoint{{Count: 1, Target: 3}, {Count: 2, Target: 3}},
		},
		{
			Definition: api.HabitStreakDefinition{
				ID:       "other",
				Name:     "Other habit",
				Enabled:  true,
				HabitIDs: []int64{1},
			},
		},
		{
			Definition: api.HabitStreakDefinition{
				ID:       "off",
				Name:     "Paused",
				Enabled:  false,
				HabitIDs: []int64{2},
			},
		},
	}
	rendered := strings.Join(
		strings.Fields(ansi.Strip(renderDetail(testTheme(), focusDetail(state), 70, 30))),
		" ",
	)
	lines := strings.Split(rendered, "\n")
	assertContainsLine(t, lines, "Fitness", "weekly")
	assertContainsLine(t, lines, "Fitness", "2/3")
	assertContainsLine(t, lines, "Fitness", "4w best 9w")
	for _, absent := range []string{"Other habit", "Paused"} {
		if strings.Contains(rendered, absent) {
			t.Fatalf("expected %q left out of the detail, got %q", absent, rendered)
		}
	}
	if !strings.Contains(rendered, "Momentum") {
		t.Fatalf("expected a Momentum heading, got %q", rendered)
	}
}

func TestDetailTitleIsTheSelection(t *testing.T) {
	lines := strings.Split(ansi.Strip(renderDetail(testTheme(), focusDetail(detailState()), 70, 30)), "\n")
	if !strings.HasPrefix(lines[0], "─ Fix sync conflict ") ||
		strings.Contains(strings.Join(lines, "\n"), "Selected") {
		t.Fatalf("expected the selected task as the pane title, got %q", lines[0])
	}
	if strings.Contains(strings.Join(lines[1:], "\n"), "Fix sync conflict") {
		t.Fatalf("expected the title only in the header, got %q", lines)
	}
}

func TestDetailDescriptionComesFirst(t *testing.T) {
	rendered := ansi.Strip(renderDetail(testTheme(), focusDetail(detailState()), 70, 30))
	if strings.Index(rendered, "keeping both edits") > strings.Index(rendered, "Effort") {
		t.Fatalf("expected the description above the fields, got %q", rendered)
	}
}

func TestDetailFactsSitBesideTheEffortArc(t *testing.T) {
	wide := strings.Split(ansi.Strip(renderDetail(testTheme(), focusDetail(detailState()), 70, 30)), "\n")
	assertContainsLine(t, wide, "⣤", "● in progress")
	narrow := ansi.Strip(renderDetail(testTheme(), focusDetail(detailState()), 32, 30))
	for _, line := range strings.Split(narrow, "\n") {
		if strings.Contains(line, "⣤") && strings.Contains(line, "in progress") {
			t.Fatalf("expected the facts under the arc when narrow, got %q", line)
		}
	}
}

func TestDetailIssueListsTheContextMomentumsItFeeds(t *testing.T) {
	state := detailState()
	state.AllIssues = []api.IssueWithMeta{
		{Issue: state.DailyIssues[1], RepoID: 1, RepoName: "crona", StreamName: "core"},
	}
	state.DailyIssues[1].StreamID = 10
	otherStream := int64(11)
	sameStream := int64(10)
	context := func(name string, contexts ...sharedtypes.MomentumContext) api.MomentumCard {
		return api.MomentumCard{
			Definition: api.HabitStreakDefinition{
				ID:            name,
				Name:          name,
				Enabled:       true,
				Period:        "day",
				RequiredCount: 3600,
				Contexts:      contexts,
			},
			Current: 2,
			Longest: 6,
			Series:  []api.MomentumSeriesPoint{{Count: 2700, Target: 3600}},
		}
	}
	state.MomentumCards = []api.MomentumCard{
		context("Deep work", sharedtypes.MomentumContext{RepoID: 1}),
		context("Core focus", sharedtypes.MomentumContext{RepoID: 1, StreamID: &sameStream}),
		context("Docs push", sharedtypes.MomentumContext{RepoID: 1, StreamID: &otherStream}),
		context("Other repo", sharedtypes.MomentumContext{RepoID: 2}),
		{
			Definition: api.HabitStreakDefinition{
				ID:       "h",
				Name:     "Habit only",
				Enabled:  true,
				HabitIDs: []int64{2},
			},
		},
	}
	rendered := ansi.Strip(renderDetail(testTheme(), focusDetail(state), 80, 30))
	lines := strings.Split(rendered, "\n")
	assertContainsLine(t, lines, "Deep work", "45m/1h")
	assertContainsLine(t, lines, "Core focus", "2d best 6d")
	for _, absent := range []string{"Docs push", "Other repo", "Habit only"} {
		if strings.Contains(rendered, absent) {
			t.Fatalf("expected %q left out, got %q", absent, rendered)
		}
	}
}

func TestDetailShowsDaySummaryUntilFocused(t *testing.T) {
	state := detailState()
	state.DailyDetailFocus = false
	state.SummaryDate = "2026-05-27"
	state.SummarySnapshot = &api.SummarySnapshot{
		Date:    "2026-05-27",
		Focus:   &api.FocusScoreSummary{Score: 72},
		Metrics: &api.DailyMetricsDay{Date: "2026-05-27", WorkedSeconds: 3600},
	}
	rendered := strings.Join(
		strings.Fields(ansi.Strip(renderDetail(testTheme(), state, 70, 30))),
		" ",
	)
	for _, want := range []string{"─ Day", "72", "Accountability", "Wellbeing"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected the Day summary to contain %q, got %q", want, rendered)
		}
	}
	if strings.Contains(rendered, "Fix sync conflict") {
		t.Fatalf("expected the selection replaced by the Day summary, got %q", rendered)
	}
}

func TestDefinitionHabitStatus(t *testing.T) {
	for _, tc := range []struct {
		name   string
		active bool
		status string
		want   string
	}{
		{"paused", false, "", "paused"}, {"off schedule", true, "", "not due today"},
		{"due", true, "due", "due"}, {"paused completed today", false, "completed", "completed"}, {"completed", true, "completed", "completed"}, {"failed", true, "failed", "failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			habit := api.HabitWithMeta{Habit: api.Habit{ID: 10, Active: tc.active}}
			state := types.ContentState{}
			if tc.status != "" {
				state.DueHabits = []api.HabitDailyItem{
					{HabitWithMeta: habit, Status: sharedtypes.HabitCompletionStatus(tc.status)},
				}
			}
			got := ansi.Strip(strings.Join(HabitDetailLines(testTheme(), state, habit, 80), "\n"))
			if !strings.Contains(got, "● "+tc.want) {
				t.Fatalf("wanted %q: %s", tc.want, got)
			}
		})
	}
}

func TestDailyHabitZeroStatusRemainsDue(t *testing.T) {
	lines := habitDetailLines(testTheme(), types.ContentState{}, api.HabitDailyItem{}, 80)
	if got := ansi.Strip(strings.Join(lines, "\n")); !strings.Contains(got, "● due") {
		t.Fatal(got)
	}
}

func TestNarrowDailyShowsTheDayLineOverTheStrip(t *testing.T) {
	state := detailState()
	state.SummaryDate = "2026-05-27"
	state.SummarySnapshot = &api.SummarySnapshot{
		Date:    "2026-05-27",
		Focus:   &api.FocusScoreSummary{Score: 72, TargetWorkedSeconds: 7200},
		Metrics: &api.DailyMetricsDay{Date: "2026-05-27", WorkedSeconds: 3600},
	}
	state.Width, state.Height = 80, 30
	narrow := strings.Split(ansi.Strip(renderView(testTheme(), state)), "\n")
	if !strings.Contains(narrow[0], "Focus 72") || !strings.Contains(narrow[0], "Worked 1h/2h") {
		t.Fatalf("expected the day line first on a narrow screen, got %q", narrow[0])
	}
	if len(narrow) != 30 {
		t.Fatalf("expected the view to keep its height, got %d rows", len(narrow))
	}
	state.Width = 200
	if wide := strings.Split(ansi.Strip(renderView(testTheme(), state)), "\n"); strings.Contains(wide[0], "Focus 72") {
		t.Fatalf("expected no day line beside the detail pane, got %q", wide[0])
	}
}

// focusDetail focuses the detail pane on the selection, as enter does.
func focusDetail(state types.ContentState) types.ContentState {
	state.DailyDetailFocus = true
	return state
}

func TestDetailListsTheIssuesSessionsNewestFirst(t *testing.T) {
	state := focusDetail(detailState())
	long, short := 3600, 900
	end := "2026-05-27T10:00:00Z"
	state.DetailSessionsIssueID = 2
	state.DetailSessions = []api.SessionHistoryEntry{
		{Session: api.Session{IssueID: 2, StartTime: "2026-05-26T09:00:00Z", EndTime: &end, DurationSeconds: &short}},
		{
			Session:     api.Session{IssueID: 2, StartTime: "2026-05-27T09:00:00Z", EndTime: &end, DurationSeconds: &long},
			ParsedNotes: sharedtypes.ParsedSessionNotes{sharedtypes.SessionNoteSectionCommit: "Fix sync merge\n\nlonger body"},
		},
	}
	rendered := ansi.Strip(renderDetail(testTheme(), state, 90, 40))
	if !strings.Contains(rendered, "1h  Fix sync merge") || strings.Contains(rendered, "longer body") {
		t.Fatalf("expected the commit's first line beside its session, got %q", rendered)
	}
	if !strings.Contains(rendered, "Sessions  2 · 1h15m total") ||
		strings.Index(rendered, "1h") > strings.Index(rendered, "15m ") {
		t.Fatalf("expected both sessions, newest first, got %q", rendered)
	}
	state.DetailSessions = nil
	if rendered := ansi.Strip(renderDetail(testTheme(), state, 70, 40)); !strings.Contains(rendered, "no work logged yet") {
		t.Fatalf("expected the empty sessions note, got %q", rendered)
	}
}

func TestHabitDetailShowsHitRateTrailAndRecent(t *testing.T) {
	state := types.ContentState{DashboardDate: "2026-05-27"}
	ten, five := 10, 5
	habit := api.HabitDailyItem{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: 7, Name: "Read", TargetMinutes: &ten}}}
	state.HabitHistory = []api.HabitCompletion{
		{HabitID: 7, Date: "2026-05-26", Status: "completed", DurationMinutes: &ten},
		{HabitID: 7, Date: "2026-05-25", Status: "failed", DurationMinutes: &five},
		{HabitID: 8, Date: "2026-05-26", Status: "completed"},
	}
	got := strings.Join(strings.Fields(ansi.Strip(strings.Join(habitDetailLines(testTheme(), state, habit, 80), "\n"))), " ")
	for _, want := range []string{"50%", "Hit/30d", "Last 14d", "✗ ● ·", "Recent", "Tue 26 May", "10m ✓", "5m ✗"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}
}
