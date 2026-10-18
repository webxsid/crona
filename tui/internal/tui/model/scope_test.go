package model

import (
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/commands"
	"crona/tui/internal/tui/scope"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAllStreamsLoadedStored(t *testing.T) {
	m := Model{view: ViewDaily, cursor: map[Pane]int{}, filters: map[Pane]string{}}
	next, _ := m.update(commands.AllStreamsLoadedMsg{Streams: []api.Stream{{ID: 2, RepoID: 1}, {ID: 4, RepoID: 1}}})
	if got := next.(Model).allStreams; len(got) != 2 {
		t.Fatalf("expected all streams stored, got %+v", got)
	}
}

func TestEnteringIssuesTabResetsBrowseScope(t *testing.T) {
	repo, stream := int64(1), int64(2)
	m := Model{
		view: ViewDaily, pane: PaneIssues, width: 120, height: 36,
		cursor: map[Pane]int{}, filters: map[Pane]string{},
		context:     &api.ActiveContext{RepoID: &repo, StreamID: &stream},
		browseScope: scope.Scope{RepoID: 3},
	}
	next, _ := m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if got := next.(Model).browseScope; got != (scope.Scope{RepoID: 1, StreamID: 2}) {
		t.Fatalf("expected the browse scope reset to the checked-out stream, got %+v", got)
	}
}
