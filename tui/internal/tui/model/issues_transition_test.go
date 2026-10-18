package model

import (
	"fmt"
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/scope"
	selectionpkg "crona/tui/internal/tui/selection"
)

func TestRegressionEntryClampsScopeList(t *testing.T) {
	m := scopeModel(0)
	m.view, m.pane = ViewDaily, PaneIssues
	m.cursor[PaneIssues] = 2
	m.allIssues = []api.IssueWithMeta{{Issue: api.Issue{ID: 1, StreamID: 2, Title: "Only issue", Status: "ready"}, RepoID: 1}}
	got, _ := press(m, "2")
	if issue, ok := selectionpkg.SelectedIssue(got.selectionSnapshot()); !ok {
		t.Fatalf("entry left cursor %d with scope %+v; selected %+v", got.cursor[PaneIssues], got.browseScope, issue)
	}
}

func TestRegressionScopeMoveKeepsOtherListInRange(t *testing.T) {
	m := scopeModel(3)
	m.view, m.pane = ViewHabitList, PaneScope
	m.browseScope = scope.Scope{}
	m.cursor[PaneIssues] = 2
	m.issues = []api.Issue{{ID: 1}, {ID: 2}, {ID: 3}}
	m.allIssues = []api.IssueWithMeta{{Issue: api.Issue{ID: 4, StreamID: 9, Title: "Only issue", Status: "ready"}, RepoID: 3}}
	got, _ := press(m, "down")
	got, _ = press(got, "[")
	if issue, ok := selectionpkg.SelectedIssue(got.selectionSnapshot()); !ok {
		t.Fatalf("sub-tab left issue cursor %d in scope %+v; selected %+v", got.cursor[PaneIssues], got.browseScope, issue)
	}
}

func TestRegressionViewJumpSeedsScope(t *testing.T) {
	for _, key := range []string{"i", "m"} {
		m := scopeModel(4)
		m.view, m.pane = ViewDaily, PaneIssues
		m.browseScope = scope.Scope{RepoID: 3}
		got, _ := press(m, "v")
		got, cmd := press(got, key)
		if got.browseScope != (scope.Scope{RepoID: 1, StreamID: 2}) || got.cursor[PaneScope] != 2 || cmd == nil {
			t.Errorf("jump %s: view %s scope %+v cursor %d cmd nil %t", key, got.view, got.browseScope, got.cursor[PaneScope], cmd == nil)
		}
	}
}

func TestIssuesTransitionsPreserveScopeAndClampLists(t *testing.T) {
	for _, source := range []View{ViewDefault, ViewHabitList} {
		for _, empty := range []bool{false, true} {
			for _, cursor := range []int{-2, 2} {
				t.Run(string(source)+fmt.Sprint(empty, cursor), func(t *testing.T) {
					m := scopeModel(3)
					m.view = source
					m.pane = PaneScope
					m.browseScope = scope.Scope{}
					m.cursor[PaneIssues] = cursor
					m.cursor[PaneHabits] = cursor
					m.issues = []api.Issue{{ID: 1}, {ID: 2}, {ID: 3}}
					if !empty {
						m.allIssues = []api.IssueWithMeta{{Issue: api.Issue{ID: 4, StreamID: 9, Title: "Match", Status: "done"}, RepoID: 3}}
						m.allHabits = []api.HabitWithMeta{{Habit: api.Habit{ID: 5, StreamID: 9, Name: "Match"}, RepoID: 3}}
					}
					m.defaultIssueSection = DefaultIssueSectionCompleted
					m.filters[PaneIssues] = "Match"
					m.filters[PaneHabits] = "Match"
					got, _ := press(m, "down")
					destination := "m"
					if source == ViewHabitList {
						destination = "i"
					}
					got, _ = press(got, "v")
					got, _ = press(got, destination)
					if got.browseScope != (scope.Scope{RepoID: 3}) {
						t.Fatalf("scope changed %+v", got.browseScope)
					}
					for _, pane := range []Pane{PaneIssues, PaneHabits} {
						if got.cursor[pane] != 0 {
							t.Fatalf("%s cursor %d", pane, got.cursor[pane])
						}
					}
					snapshot := got.selectionSnapshot()
					snapshot.View = ViewDefault
					snapshot.Pane = PaneIssues
					issue, ok := selectionpkg.SelectedIssue(snapshot)
					if ok == empty || !empty && issue.ID != 4 {
						t.Fatalf("issue %+v found %t empty %t", issue, ok, empty)
					}
					snapshot.View = ViewHabitList
					snapshot.Pane = PaneHabits
					habit, ok := selectionpkg.SelectedHabit(snapshot)
					if ok == empty || !empty && habit.ID != 5 {
						t.Fatalf("habit %+v found %t empty %t", habit, ok, empty)
					}
				})
			}
		}
	}
}

func TestEntryRoutesClampFilteredLists(t *testing.T) {
	for _, route := range []string{"2", "i", "m"} {
		for _, empty := range []bool{false, true} {
			t.Run(route+fmt.Sprint(empty), func(t *testing.T) {
				m := scopeModel(4)
				m.view = ViewDaily
				m.pane = PaneIssues
				m.cursor[PaneIssues] = 2
				m.cursor[PaneHabits] = 2
				m.filters[PaneIssues] = "Match"
				m.filters[PaneHabits] = "Match"
				if !empty {
					m.allIssues = []api.IssueWithMeta{{Issue: api.Issue{ID: 1, StreamID: 2, Title: "Match", Status: "ready"}, RepoID: 1}}
					m.allHabits = []api.HabitWithMeta{{Habit: api.Habit{ID: 2, StreamID: 2, Name: "Match"}, RepoID: 1}}
				}
				if route != "2" {
					m, _ = press(m, "v")
				}
				got, cmd := press(m, route)
				if cmd == nil {
					t.Fatal("entry must refresh today")
				}
				if got.cursor[PaneIssues] != 0 || got.cursor[PaneHabits] != 0 || got.browseScope != (scope.Scope{RepoID: 1, StreamID: 2}) {
					t.Fatalf("scope %+v cursors %v", got.browseScope, got.cursor)
				}
			})
		}
	}
}
