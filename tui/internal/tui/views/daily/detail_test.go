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
	rendered := ansi.Strip(renderDetail(testTheme(), detailState(), 70, 30))
	for _, want := range []string{"Fix sync conflict", "in progress", "crona › core", "30m/1h", "keeping both edits"} {
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
	rendered := ansi.Strip(renderDetail(testTheme(), state, 70, 30))
	if !strings.Contains(rendered, "Read") {
		t.Fatalf("expected selected habit details, got %q", rendered)
	}
	assertContainsLine(t, strings.Split(rendered, "\n"), "target", "20m")
	if strings.Contains(rendered, "Fix sync conflict") {
		t.Fatalf("expected habit focus to replace task detail, got %q", rendered)
	}
}

func TestDetailEmptySelection(t *testing.T) {
	state := detailState()
	state.DailyIssues = nil
	rendered := ansi.Strip(renderDetail(testTheme(), state, 70, 30))
	if !strings.Contains(rendered, "Nothing selected") {
		t.Fatalf("expected empty-selection hint, got %q", rendered)
	}
}

func TestDetailHasNoCalendar(t *testing.T) {
	if rendered := ansi.Strip(renderDetail(testTheme(), detailState(), 70, 40)); strings.Contains(rendered, "Mo Tu We") {
		t.Fatalf("expected the calendar to live in the strip, got %q", rendered)
	}
}

func TestDetailFollowsModelFilteredCursor(t *testing.T) {
	state := detailState()
	state.Filters["issues"] = "sync"
	state.DailyIssueIndices = []int{1}
	state.Cursors["issues"] = 0
	detail := ansi.Strip(renderDetail(testTheme(), state, 70, 30))
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
	detail := ansi.Strip(renderDetail(testTheme(), state, 70, 30))
	if !strings.Contains(detail, "Read") || strings.Contains(detail, "Walk") {
		t.Fatalf("expected the filtered habit in detail, got %q", detail)
	}
}

func TestDailyTrustsModelTaskSection(t *testing.T) {
	state := dailyState(130, 30)
	state.DailyTaskSection = "planned"
	yesterday := "2026-05-26"
	state.DailyIssues = []api.Issue{
		{ID: 1, Title: "Closed pinned task", Status: "done", PinnedDaily: true, TodoForDate: &yesterday},
		{ID: 2, Title: "Open pinned task", Status: "ready", PinnedDaily: true},
	}
	state.Cursors["issues"] = 0
	if detail := ansi.Strip(renderDetail(testTheme(), state, 70, 30)); !strings.Contains(detail, "Closed pinned task") {
		t.Fatalf("expected the model's first pinned task selected, got %q", detail)
	}
}

func TestDetailHabitListsTheMomentumsItFeeds(t *testing.T) {
	state := detailState()
	state.Pane = "habits"
	state.Cursors["habits"] = 1 // Read, habit ID 2
	state.MomentumCards = []api.MomentumCard{
		{
			Definition: api.HabitStreakDefinition{ID: "fit", Name: "Fitness", Enabled: true, Period: "week", RequiredCount: 3, HabitIDs: []int64{1, 2}},
			Current:    4,
			Longest:    9,
			Series:     []api.MomentumSeriesPoint{{Count: 1, Target: 3}, {Count: 2, Target: 3}},
		},
		{Definition: api.HabitStreakDefinition{ID: "other", Name: "Other habit", Enabled: true, HabitIDs: []int64{1}}},
		{Definition: api.HabitStreakDefinition{ID: "off", Name: "Paused", Enabled: false, HabitIDs: []int64{2}}},
	}
	rendered := ansi.Strip(renderDetail(testTheme(), state, 70, 30))
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
	lines := strings.Split(ansi.Strip(renderDetail(testTheme(), detailState(), 70, 30)), "\n")
	if !strings.HasPrefix(lines[0], "─ Fix sync conflict ") || strings.Contains(strings.Join(lines, "\n"), "Selected") {
		t.Fatalf("expected the selected task as the pane title, got %q", lines[0])
	}
	if strings.Contains(strings.Join(lines[1:], "\n"), "Fix sync conflict") {
		t.Fatalf("expected the title only in the header, got %q", lines)
	}
}

func TestDetailDescriptionComesFirst(t *testing.T) {
	rendered := ansi.Strip(renderDetail(testTheme(), detailState(), 70, 30))
	if strings.Index(rendered, "keeping both edits") > strings.Index(rendered, "status") {
		t.Fatalf("expected the description above the fields, got %q", rendered)
	}
}

func TestDetailFieldsFlowIntoColumns(t *testing.T) {
	wide := strings.Split(ansi.Strip(renderDetail(testTheme(), detailState(), 70, 30)), "\n")
	assertContainsLine(t, wide, "status", "context")
	narrow := ansi.Strip(renderDetail(testTheme(), detailState(), 32, 30))
	for _, line := range strings.Split(narrow, "\n") {
		if strings.Contains(line, "status") && strings.Contains(line, "context") {
			t.Fatalf("expected one field per row when narrow, got %q", line)
		}
	}
}

func TestDetailIssueListsTheContextMomentumsItFeeds(t *testing.T) {
	state := detailState()
	state.AllIssues = []api.IssueWithMeta{{Issue: state.DailyIssues[1], RepoID: 1, RepoName: "crona", StreamName: "core"}}
	state.DailyIssues[1].StreamID = 10
	otherStream := int64(11)
	sameStream := int64(10)
	context := func(name string, contexts ...sharedtypes.MomentumContext) api.MomentumCard {
		return api.MomentumCard{
			Definition: api.HabitStreakDefinition{ID: name, Name: name, Enabled: true, Period: "day", RequiredCount: 3600, Contexts: contexts},
			Current:    2,
			Longest:    6,
			Series:     []api.MomentumSeriesPoint{{Count: 2700, Target: 3600}},
		}
	}
	state.MomentumCards = []api.MomentumCard{
		context("Deep work", sharedtypes.MomentumContext{RepoID: 1}),
		context("Core focus", sharedtypes.MomentumContext{RepoID: 1, StreamID: &sameStream}),
		context("Docs push", sharedtypes.MomentumContext{RepoID: 1, StreamID: &otherStream}),
		context("Other repo", sharedtypes.MomentumContext{RepoID: 2}),
		{Definition: api.HabitStreakDefinition{ID: "h", Name: "Habit only", Enabled: true, HabitIDs: []int64{2}}},
	}
	rendered := ansi.Strip(renderDetail(testTheme(), state, 80, 30))
	lines := strings.Split(rendered, "\n")
	assertContainsLine(t, lines, "Deep work", "45m/1h")
	assertContainsLine(t, lines, "Core focus", "2d best 6d")
	for _, absent := range []string{"Docs push", "Other repo", "Habit only"} {
		if strings.Contains(rendered, absent) {
			t.Fatalf("expected %q left out, got %q", absent, rendered)
		}
	}
}

func TestDetailShowsDaySummaryWhenToggled(t *testing.T) {
	state := detailState()
	state.DailyDetailDay = true
	state.SummaryDate = "2026-05-27"
	state.SummarySnapshot = &api.SummarySnapshot{
		Date:    "2026-05-27",
		Focus:   &api.FocusScoreSummary{Score: 72},
		Metrics: &api.DailyMetricsDay{Date: "2026-05-27", WorkedSeconds: 3600},
	}
	rendered := ansi.Strip(renderDetail(testTheme(), state, 70, 30))
	for _, want := range []string{"─ Day", "Focus 72/100", "Accountability", "Wellbeing"} {
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
				state.DueHabits = []api.HabitDailyItem{{HabitWithMeta: habit, Status: sharedtypes.HabitCompletionStatus(tc.status)}}
			}
			got := ansi.Strip(strings.Join(HabitDetailLines(testTheme(), state, habit, 80), "\n"))
			if status := strings.Join(strings.Fields(strings.Split(got, "\n")[0]), " "); status != "status "+tc.want {
				t.Fatalf("wanted %q: %s", tc.want, got)
			}
		})
	}
}

func TestDailyHabitZeroStatusRemainsDue(t *testing.T) {
	lines := habitDetailLines(testTheme(), types.ContentState{}, api.HabitDailyItem{}, 80)
	if got := strings.Join(strings.Fields(ansi.Strip(lines[0])), " "); got != "status due" {
		t.Fatal(got)
	}
}
