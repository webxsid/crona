package habitlist

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func TestHabitStripTwoRows(t *testing.T) {
	state := types.ContentState{}
	for i := range 6 {
		state.ScopedHabits = append(
			state.ScopedHabits,
			api.HabitWithMeta{Habit: api.Habit{ID: int64(i), Active: i < 5}},
		)
	}
	for i := range 3 {
		state.DueHabits = append(
			state.DueHabits,
			api.HabitDailyItem{HabitWithMeta: state.ScopedHabits[i], Completed: i < 2},
		)
	}
	state.MomentumCards = []api.MomentumCard{
		{
			Definition: api.HabitStreakDefinition{
				Enabled:    true,
				TargetKind: "habit",
				Period:     "day",
				HabitIDs:   []int64{1},
			},
			Longest: 12,
		},
	}
	for _, width := range []int{60, 100, 160} {
		rows := StripRows(types.Theme{}, state, width)
		if len(rows) != 2 {
			t.Fatal(rows)
		}
		for _, row := range rows {
			if ansi.StringWidth(row) != width {
				t.Fatal(row)
			}
		}
		if !strings.Contains(rows[0], "Active 5  paused 1  due today 3") ||
			!strings.Contains(rows[1], "Done today 2  failed 0  best streak 12d") {
			t.Fatal(rows)
		}
	}
}
func TestHabitListCardsAndActions(t *testing.T) {
	target := 30
	state := types.ContentState{
		View:   "habit_list",
		Pane:   "habits",
		Width:  120,
		Height: 20,
		ScopedHabits: []api.HabitWithMeta{
			{
				Habit: api.Habit{
					ID:            1,
					Name:          "Walk",
					ScheduleType:  "weekdays",
					TargetMinutes: &target,
				},
				RepoName:   "crona",
				StreamName: "tui",
			},
		},
	}
	rendered := ansi.Strip(Render(types.Theme{}, state))
	for _, want := range []string{"[a] new habit", "Walk", "[enter] open  [e] edit  [D] delete  [m] log  [+] tags"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q: %s", want, rendered)
		}
	}
	state.Pane = "scope"
	rendered = ansi.Strip(Render(types.Theme{}, state))
	if !strings.Contains(rendered, "crona › tui | weekdays | 30m") ||
		strings.Contains(rendered, "[a] new habit") {
		t.Fatal(rendered)
	}
	state.Pane = "habits"
	state.Timer = &api.TimerState{State: "running"}
	rendered = Render(types.Theme{}, state)
	if strings.Contains(rendered, "[D]") || strings.Contains(rendered, "[m]") {
		t.Fatal(rendered)
	}
}

func TestBestStreakRespectsScopeAndPeriod(t *testing.T) {
	state := types.ContentState{ScopedHabits: []api.HabitWithMeta{{Habit: api.Habit{ID: 1}}}}
	state.MomentumCards = []api.MomentumCard{
		{
			Definition: api.HabitStreakDefinition{
				Enabled:    true,
				TargetKind: "habit",
				Period:     "day",
				HabitIDs:   []int64{1},
			},
			Longest: 12,
		},
		{
			Definition: api.HabitStreakDefinition{
				Enabled:    true,
				TargetKind: "habit",
				Period:     "day",
				HabitIDs:   []int64{1, 2},
			},
			Longest: 90,
		},
		{
			Definition: api.HabitStreakDefinition{
				Enabled:    true,
				TargetKind: "habit",
				Period:     "week",
				HabitIDs:   []int64{1},
			},
			Longest: 50,
		},
		{
			Definition: api.HabitStreakDefinition{
				TargetKind: "habit",
				Period:     "day",
				HabitIDs:   []int64{1},
			},
			Longest: 99,
		},
	}
	if row := StripRows(types.Theme{}, state, 100)[1]; !strings.Contains(row, "best streak 12d") {
		t.Fatal(row)
	}
	state.MomentumCards = state.MomentumCards[1:]
	if row := StripRows(types.Theme{}, state, 100)[1]; !strings.Contains(row, "best streak —") {
		t.Fatal(row)
	}
}

func TestHabitLayoutDimensions(t *testing.T) {
	for _, width := range []int{60, 99, 100, 159, 160, 200} {
		state := types.ContentState{View: "habit_list", Pane: "habits", Width: width, Height: 20}
		for i := range 20 {
			state.ScopedHabits = append(
				state.ScopedHabits,
				api.HabitWithMeta{Habit: api.Habit{ID: int64(i), Name: "Walk"}},
			)
		}
		rows := strings.Split(Render(types.Theme{}, state), "\n")
		if len(rows) != state.Height {
			t.Fatalf("width %d: height %d", width, len(rows))
		}
		for _, row := range rows {
			if ansi.StringWidth(row) != width {
				t.Fatalf("width %d: got %d in %q", width, ansi.StringWidth(row), row)
			}
		}
	}
}

func TestHabitPagingKeepsBothIndicators(t *testing.T) {
	state := types.ContentState{Pane: "habits", Cursors: map[string]int{"habits": 9}}
	for i := range 20 {
		state.ScopedHabits = append(
			state.ScopedHabits,
			api.HabitWithMeta{Habit: api.Habit{ID: int64(i), Name: "Walk"}},
		)
	}
	rendered := renderHabits(types.Theme{}, state, 100, 20)
	if !strings.Contains(rendered, "↑") || !strings.Contains(rendered, "↓") ||
		strings.Contains(rendered, "↓ more") {
		t.Fatalf("paged cards lost indicator: %s", rendered)
	}
}
