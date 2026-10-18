package selection

import (
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/scope"
	uistate "crona/tui/internal/tui/state"
)

func TestDefaultScopedIssuesFollowBrowseScope(t *testing.T) {
	stream := int64(2)
	repo := int64(1)
	s := Snapshot{
		View:        uistate.ViewDefault,
		Context:     &api.ActiveContext{RepoID: &repo, StreamID: &stream},
		BrowseScope: scope.Scope{RepoID: 1, StreamID: 4},
		AllIssues: []api.IssueWithMeta{
			{Issue: api.Issue{ID: 1, StreamID: 2}, RepoID: 1},
			{Issue: api.Issue{ID: 2, StreamID: 4}, RepoID: 1},
		},
	}
	got := DefaultScopedIssues(s)
	if len(got) != 1 || got[0].ID != 2 {
		t.Fatalf("expected only the browsed stream's issue, got %+v", got)
	}
}

func TestScopedHabitsFollowBrowseScope(t *testing.T) {
	s := Snapshot{
		BrowseScope: scope.Scope{RepoID: 3},
		AllHabits: []api.HabitWithMeta{
			{Habit: api.Habit{ID: 1, StreamID: 2}, RepoID: 1},
			{Habit: api.Habit{ID: 2, StreamID: 5}, RepoID: 3},
		},
	}
	got := ScopedHabits(s)
	if len(got) != 1 || got[0].ID != 2 {
		t.Fatalf("expected only the browsed repo's habit, got %+v", got)
	}
}
