package scopepane

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/scope"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func treeState(cursor int, pane string) types.ContentState {
	repo, stream := int64(1), int64(2)
	return types.ContentState{
		Pane:       pane,
		Cursors:    map[string]int{"scope": cursor},
		Repos:      []api.Repo{{ID: 1, Name: "crona"}, {ID: 3, Name: "dotfiles"}},
		AllStreams: []api.Stream{{ID: 2, RepoID: 1, Name: "tui"}, {ID: 4, RepoID: 1, Name: "daemon"}},
		Context:    &api.ActiveContext{RepoID: &repo, StreamID: &stream},
	}
}

var treeCounts = map[scope.Scope]int{{}: 24, {RepoID: 1}: 18, {RepoID: 1, StreamID: 2}: 9, {RepoID: 1, StreamID: 4}: 6, {RepoID: 3}: 6}

func plainRows(s string) []string {
	rows := strings.Split(ansi.Strip(s), "\n")
	for i, row := range rows {
		rows[i] = strings.Join(strings.Fields(row), " ")
	}
	return rows
}

func TestScopeTreeRows(t *testing.T) {
	rows := plainRows(Render(types.Theme{}, treeState(1, "issues"), treeCounts, 24, 12))
	want := []string{"All 24", "crona 18 ●", "tui 9 ●", "daemon 6", "dotfiles 6"}
	for i, w := range want {
		if rows[i+1] != w {
			t.Fatalf("row %d: want %q, got %q (all %q)", i+1, w, rows[i+1], rows)
		}
	}
	for _, line := range strings.Split(ansi.Strip(Render(types.Theme{}, treeState(1, "issues"), treeCounts, 24, 12)), "\n") {
		if ansi.StringWidth(line) != 24 {
			t.Fatalf("expected 24-cell rows, got %q", line)
		}
	}
}

func TestScopeActionLineOnlyWhenFocused(t *testing.T) {
	focused := plainRows(Render(types.Theme{}, treeState(1, "scope"), treeCounts, 60, 14))
	if focused[3] != "[c] check out [e] edit [D] delete [enter] open" {
		t.Fatalf("expected the action line under crona, got %q", focused)
	}
	if strings.Contains(strings.Join(plainRows(Render(types.Theme{}, treeState(1, "issues"), treeCounts, 60, 14)), "\n"), "[c]") {
		t.Fatal("expected no action line while unfocused")
	}
	if strings.Contains(strings.Join(plainRows(Render(types.Theme{}, treeState(0, "scope"), treeCounts, 60, 14)), "\n"), "[c]") {
		t.Fatal("expected no action line on All")
	}
}

func TestScopeCreateLine(t *testing.T) {
	for cursor, want := range map[int]string{0: "[a] new repo", 1: "[a] new stream", 2: "[a] new stream", 4: "[a] new stream"} {
		joined := strings.Join(plainRows(Render(types.Theme{}, treeState(cursor, "scope"), treeCounts, 30, 14)), "\n")
		if !strings.Contains(joined, want) {
			t.Fatalf("cursor %d: expected %q, got %q", cursor, want, joined)
		}
	}
}

func TestScopeRowNarrow(t *testing.T) {
	row := ansi.Strip(RenderRow(types.Theme{}, treeState(2, "issues"), 80))
	if !strings.Contains(row, "Scope ‹ crona › tui ›") || !strings.Contains(row, "●") || ansi.StringWidth(row) != 80 {
		t.Fatalf("got %q", row)
	}
}

func TestIssueCountsOpenPerNode(t *testing.T) {
	counts := IssueCounts([]api.IssueWithMeta{
		{Issue: api.Issue{StreamID: 2, Status: "ready"}, RepoID: 1},
		{Issue: api.Issue{StreamID: 2, Status: "done"}, RepoID: 1},
		{Issue: api.Issue{StreamID: 5, Status: "planned"}, RepoID: 3},
	})
	if counts[scope.Scope{}] != 2 || counts[scope.Scope{RepoID: 1}] != 1 || counts[scope.Scope{RepoID: 1, StreamID: 2}] != 1 || counts[scope.Scope{RepoID: 3, StreamID: 5}] != 1 {
		t.Fatalf("got %+v", counts)
	}
}

func TestScopeHintsHonorFocusAndTimer(t *testing.T) {
	state := treeState(1, "issues")
	if strings.Contains(Render(types.Theme{}, state, treeCounts, 80, 14), "[a]") {
		t.Fatal("unfocused create hint")
	}
	state.Pane = "scope"
	state.Timer = &api.TimerState{State: "running"}
	if strings.Contains(Render(types.Theme{}, state, treeCounts, 80, 14), "[D]") {
		t.Fatal("timer guarded delete hint")
	}
}
