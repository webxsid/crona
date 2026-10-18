package model

import (
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/commands"
	"crona/tui/internal/tui/scope"
	selectionpkg "crona/tui/internal/tui/selection"

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

func scopeModel(cursor int) Model {
	repo, stream := int64(1), int64(2)
	return Model{
		view: ViewDefault, pane: PaneScope, width: 200, height: 40,
		cursor: map[Pane]int{PaneScope: cursor}, filters: map[Pane]string{},
		context:     &api.ActiveContext{RepoID: &repo, StreamID: &stream},
		repos:       []api.Repo{{ID: 1, Name: "crona"}, {ID: 3, Name: "dotfiles"}},
		allStreams:  []api.Stream{{ID: 2, RepoID: 1, Name: "tui"}, {ID: 4, RepoID: 1, Name: "daemon"}},
		browseScope: scope.Scope{RepoID: 1},
	}
}

func press(m Model, key string) (Model, tea.Cmd) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	if key == "down" {
		msg = tea.KeyMsg{Type: tea.KeyDown}
	}
	next, cmd := m.update(msg)
	return next.(Model), cmd
}

func TestScopeCursorMovesBrowseScope(t *testing.T) {
	m, _ := press(scopeModel(1), "down")
	if m.browseScope != (scope.Scope{RepoID: 1, StreamID: 2}) || m.cursor[PaneScope] != 2 {
		t.Fatalf("expected the browse scope to follow the cursor to tui, got %+v at %d", m.browseScope, m.cursor[PaneScope])
	}
	if *m.context.StreamID != 2 || *m.context.RepoID != 1 {
		t.Fatal("expected the checked-out context unchanged")
	}
}

func TestScopeCheckoutUsesNode(t *testing.T) {
	if _, cmd := press(scopeModel(3), "c"); cmd == nil {
		t.Fatal("expected c on a stream node to check it out")
	}
	if _, cmd := press(scopeModel(0), "c"); cmd != nil {
		t.Fatal("expected c on All to do nothing")
	}
}

func TestScopeCreateOpensRepoOrStream(t *testing.T) {
	if m, _ := press(scopeModel(0), "a"); m.dialog != "create_repo" {
		t.Fatalf("expected a on All to create a repo, got %q", m.dialog)
	}
	if m, _ := press(scopeModel(2), "a"); m.dialog != "create_stream" || m.dialogRepoID != 1 {
		t.Fatalf("expected a on tui to create a stream in crona, got %q repo %d", m.dialog, m.dialogRepoID)
	}
}

func TestScopeEditAndDeleteUseNode(t *testing.T) {
	if m, _ := press(scopeModel(3), "e"); m.dialog != "edit_stream" {
		t.Fatalf("expected e on daemon to edit the stream, got %q", m.dialog)
	}
	if m, _ := press(scopeModel(1), "D"); m.dialog != "confirm_delete" {
		t.Fatalf("expected D on crona to confirm a delete, got %q", m.dialog)
	}
}

func TestScopeCursorFollowsScopeWhenStreamsLoad(t *testing.T) {
	m := scopeModel(0)
	m.allStreams = nil
	m.browseScope = scope.Scope{RepoID: 1, StreamID: 4}
	next, _ := m.update(commands.AllStreamsLoadedMsg{Streams: []api.Stream{{ID: 2, RepoID: 1, Name: "tui"}, {ID: 4, RepoID: 1, Name: "daemon"}}})
	got := next.(Model)
	if got.browseScope != (scope.Scope{RepoID: 1, StreamID: 4}) || got.cursor[PaneScope] != 3 {
		t.Fatalf("expected the cursor on daemon once streams load, got %+v at %d", got.browseScope, got.cursor[PaneScope])
	}
}

func TestDeletedBrowsedStreamFallsBackToAll(t *testing.T) {
	m := scopeModel(3)
	m.browseScope = scope.Scope{RepoID: 1, StreamID: 4}
	next, _ := m.update(commands.AllStreamsLoadedMsg{Streams: []api.Stream{{ID: 2, RepoID: 1, Name: "tui"}}})
	got := next.(Model)
	if !got.browseScope.IsAll() || got.cursor[PaneScope] != 0 {
		t.Fatalf("expected All after the browsed stream went away, got %+v at %d", got.browseScope, got.cursor[PaneScope])
	}
}

func TestEnteringIssuesTabSeedsScopeCursor(t *testing.T) {
	m := scopeModel(0)
	m.view, m.pane = ViewDaily, PaneIssues
	got, _ := press(m, "2")
	if got.cursor[PaneScope] != 2 {
		t.Fatalf("expected the scope cursor on tui, got %d", got.cursor[PaneScope])
	}
}

func TestScopeArrowsStepNodes(t *testing.T) {
	next, _ := scopeModel(1).update(tea.KeyMsg{Type: tea.KeyRight})
	if got := next.(Model); got.cursor[PaneScope] != 2 || got.browseScope != (scope.Scope{RepoID: 1, StreamID: 2}) {
		t.Fatalf("expected → to step to tui, got %d %+v", got.cursor[PaneScope], got.browseScope)
	}
	next, _ = scopeModel(1).update(tea.KeyMsg{Type: tea.KeyLeft})
	if got := next.(Model); got.cursor[PaneScope] != 0 || !got.browseScope.IsAll() {
		t.Fatalf("expected ← to step to All, got %d", got.cursor[PaneScope])
	}
}

func TestArrowsSwitchOpenDoneAndTabCyclesPanes(t *testing.T) {
	m := scopeModel(1)
	m.pane = PaneIssues
	m.cursor[PaneIssues] = 2
	next, _ := m.update(tea.KeyMsg{Type: tea.KeyRight})
	got := next.(Model)
	if got.defaultIssueSection != DefaultIssueSectionCompleted || got.cursor[PaneIssues] != 0 {
		t.Fatalf("right: section %q cursor %d", got.defaultIssueSection, got.cursor[PaneIssues])
	}
	next, _ = got.update(tea.KeyMsg{Type: tea.KeyTab})
	if got := next.(Model); got.pane != PaneScope {
		t.Fatalf("tab: pane %q", got.pane)
	}
}

func TestDoneSectionCursorInRange(t *testing.T) {
	m := scopeModel(1)
	m.pane = PaneIssues
	m.allIssues = []api.IssueWithMeta{{Issue: api.Issue{ID: 1, StreamID: 2, Status: "ready"}, RepoID: 1}}
	m.cursor[PaneIssues] = 3
	m.setDefaultIssueSection(DefaultIssueSectionCompleted)
	if m.cursor[PaneIssues] != 0 || len(selectionpkg.FilteredIndices(m.selectionSnapshot(), PaneIssues)) != 0 {
		t.Fatalf("empty Done: cursor %d indices %v", m.cursor[PaneIssues], selectionpkg.FilteredIndices(m.selectionSnapshot(), PaneIssues))
	}
}

func TestCreateIssueDefaultsToBrowseScope(t *testing.T) {
	m := scopeModel(3)
	m.pane = PaneIssues
	m.browseScope = scope.Scope{RepoID: 1, StreamID: 4}
	got, _ := press(m, "a")
	if got.dialog != "create_issue_default" {
		t.Fatalf("dialog %q", got.dialog)
	}
	if got.dialogInputs[0].Value() != "crona" || got.dialogInputs[1].Value() != "daemon" {
		t.Fatalf("defaults %q / %q", got.dialogInputs[0].Value(), got.dialogInputs[1].Value())
	}
}

func TestDailyCreateIssueDefaultsToCheckout(t *testing.T) {
	for _, browse := range []scope.Scope{{}, {RepoID: 1, StreamID: 4}} {
		m := scopeModel(3)
		m.view, m.pane = ViewDaily, PaneIssues
		repoName, streamName := "crona", "tui"
		m.context.RepoName, m.context.StreamName = &repoName, &streamName
		m.browseScope = browse
		got, _ := press(m, "a")
		if got.dialog != "create_issue_default" {
			t.Fatalf("browse %+v: dialog %q", browse, got.dialog)
		}
		if got.dialogInputs[0].Value() != "crona" || got.dialogInputs[1].Value() != "tui" {
			t.Errorf("browse %+v: Daily defaults %q / %q, want crona / tui", browse, got.dialogInputs[0].Value(), got.dialogInputs[1].Value())
		}
	}
}
