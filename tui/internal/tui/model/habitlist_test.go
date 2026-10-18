package model

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/commands"
	"crona/tui/internal/tui/scope"
	"crona/tui/internal/tui/views/habitlist"
	types "crona/tui/internal/tui/views/types"
)

func TestCreateHabitDefaultsToBrowseScope(t *testing.T) {
	for _, test := range []struct {
		scope        scope.Scope
		repo, stream string
	}{
		{scope.Scope{}, "", ""}, {scope.Scope{RepoID: 1}, "crona", ""}, {scope.Scope{RepoID: 1, StreamID: 4}, "crona", "daemon"},
	} {
		m := scopeModel(0)
		m.view = ViewHabitList
		m.pane = PaneHabits
		m.browseScope = test.scope
		next, _ := press(m, "a")
		if next.dialog != "create_habit" || next.dialogInputs[0].Value() != test.repo || next.dialogInputs[1].Value() != test.stream {
			t.Fatalf("scope %+v: %s %q %q", test.scope, next.dialog, next.dialogInputs[0].Value(), next.dialogInputs[1].Value())
		}
	}
}
func TestHabitListKeysActOnScopedHabit(t *testing.T) {
	m := scopeModel(1)
	m.view = ViewHabitList
	m.pane = PaneHabits
	m.allHabits = []api.HabitWithMeta{{Habit: api.Habit{ID: 1, StreamID: 9, Name: "Other"}, RepoID: 3}, {Habit: api.Habit{ID: 2, StreamID: 4, Name: "Walk"}, RepoID: 1}}
	next, _ := press(m, "e")
	if next.dialog != "edit_habit" || next.dialogHabitID != 2 {
		t.Fatalf("edit = %s habit %d", next.dialog, next.dialogHabitID)
	}
	m.timer = &api.TimerState{State: "running"}
	next, _ = press(m, "m")
	if next.dialog != "" {
		t.Fatalf("running timer opened %s", next.dialog)
	}
	next, _ = press(m, "D")
	if next.dialog != "" {
		t.Fatalf("running timer opened %s", next.dialog)
	}
	next, _ = press(m, "e")
	if next.dialog != "edit_habit" {
		t.Fatalf("edit unavailable: %s", next.dialog)
	}
}
func TestNoMetaView(t *testing.T) {
	m := scopeModel(0)
	m.view = ViewHabitList
	m.pane = PaneHabits
	if len(m.layoutState().PaneActions) != 0 {
		t.Fatal("Habits bottom row has pane actions")
	}
	if terminalViewTitle(ViewHabitList) != "Habits" {
		t.Fatal("incorrect terminal title")
	}
	jump := m.openViewJumpDialog().dialogState()
	if strings.Contains(strings.Join(jump.ChoiceItems, " "), "Meta") || !strings.Contains(strings.Join(jump.ChoiceItems, " "), "Habits") {
		t.Fatalf("view jump choices = %v, want Habits without Meta", jump.ChoiceItems)
	}
	next := m.openHelpDialog()
	if strings.Contains(next.dialogViewBody, "Meta") {
		t.Fatal(next.dialogViewBody)
	}
}

func TestHabitScopeChangeClampsCursor(t *testing.T) {
	m := scopeModel(1)
	m.view = ViewHabitList
	m.pane = PaneScope
	m.cursor[PaneHabits] = 2
	m.allHabits = []api.HabitWithMeta{{Habit: api.Habit{ID: 1, StreamID: 2, Name: "Walk"}, RepoID: 1}}
	next, _ := press(m, "down")
	if next.cursor[PaneHabits] != 0 {
		t.Fatalf("habit cursor = %d", next.cursor[PaneHabits])
	}
}

func TestHabitSummaryUsesTodayDespiteHistoricalDailyDate(t *testing.T) {
	m := scopeModel(1)
	m.view = ViewDaily
	m.pane = PaneHabits
	m.date = "2020-01-01"
	habit := api.HabitWithMeta{Habit: api.Habit{ID: 2, StreamID: 2, Name: "Walk", Active: true}, RepoID: 1}
	m.allHabits = []api.HabitWithMeta{habit}
	m.dueHabits = []api.HabitDailyItem{{HabitWithMeta: habit, Status: "failed"}}
	loaded, _ := m.update(commands.TodayHabitsLoadedMsg{Date: m.currentLogicalDate(), Habits: []api.HabitDailyItem{{HabitWithMeta: habit, Status: "completed", Completed: true}}})
	m = loaded.(Model)
	m.view = ViewHabitList
	state := m.layoutState().ContentState
	rows := habitlist.StripRows(types.Theme{}, state, 100)
	if !strings.Contains(rows[1], "Done today 1  failed 0") {
		t.Fatal(rows)
	}
	m.view = ViewDaily
	if due := m.layoutState().ContentState.DueHabits; len(due) != 1 || due[0].Status != "failed" || m.currentDate() != "2020-01-01" {
		t.Fatalf("Daily changed: %+v %s", due, m.currentDate())
	}
}

func TestHabitLogDefaultsToTodayWithoutChangingDailyDate(t *testing.T) {
	m := scopeModel(1)
	m.view = ViewHabitList
	m.pane = PaneHabits
	m.date = "2020-01-01"
	m.allHabits = []api.HabitWithMeta{{Habit: api.Habit{ID: 2, StreamID: 2, Name: "Walk"}, RepoID: 1}}
	next, _ := press(m, "m")
	if next.dialog != "complete_habit" || next.dialogCheckInDate != m.currentLogicalDate() {
		t.Fatalf("habit log = %s date %s, want today %s", next.dialog, next.dialogCheckInDate, m.currentLogicalDate())
	}
}

func TestReloadScopedHabitsClampsSelection(t *testing.T) {
	m := scopeModel(1)
	m.view = ViewHabitList
	m.pane = PaneHabits
	m.cursor[PaneHabits] = 2
	loaded, _ := m.update(commands.AllHabitsLoadedMsg{Habits: []api.HabitWithMeta{{Habit: api.Habit{ID: 1, StreamID: 2, Name: "Walk"}, RepoID: 1}}})
	if got := loaded.(Model).cursor[PaneHabits]; got != 0 {
		t.Fatalf("reloaded habit cursor = %d, want 0", got)
	}
}
