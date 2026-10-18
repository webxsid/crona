package scope

import (
	"slices"
	"testing"

	"crona/tui/internal/api"
)

func treeFixture() ([]api.Repo, []api.Stream) {
	return []api.Repo{{ID: 1, Name: "crona"}, {ID: 3, Name: "dotfiles"}},
		[]api.Stream{{ID: 2, RepoID: 1, Name: "tui"}, {ID: 4, RepoID: 1, Name: "daemon"}}
}

func TestNodesNestStreamsUnderRepos(t *testing.T) {
	nodes := Nodes(treeFixture())
	names, depths := []string{}, []int{}
	for _, node := range nodes {
		names, depths = append(names, node.Name), append(depths, node.Depth)
	}
	if !slices.Equal(names, []string{"All", "crona", "tui", "daemon", "dotfiles"}) || !slices.Equal(depths, []int{0, 0, 1, 1, 0}) {
		t.Fatalf("got names %v depths %v", names, depths)
	}
	if nodes[2].Scope != (Scope{RepoID: 1, StreamID: 2}) || !nodes[0].Scope.IsAll() {
		t.Fatalf("unexpected scopes %+v", nodes)
	}
}

func TestIndexOfFallsBackToRepoThenAll(t *testing.T) {
	nodes := Nodes(treeFixture())
	if got := IndexOf(nodes, Scope{RepoID: 1, StreamID: 4}); got != 3 {
		t.Fatalf("exact match: got %d", got)
	}
	if got := IndexOf(nodes, Scope{RepoID: 1, StreamID: 9}); got != 1 {
		t.Fatalf("missing stream: got %d, want the repo", got)
	}
	if got := IndexOf(nodes, Scope{RepoID: 7}); got != 0 {
		t.Fatalf("missing repo: got %d, want All", got)
	}
}

func TestScopeContains(t *testing.T) {
	if !(Scope{}).Contains(3, 4) {
		t.Fatal("All contains everything")
	}
	if repo := (Scope{RepoID: 1}); !repo.Contains(1, 2) || repo.Contains(3, 4) {
		t.Fatal("repo scope contains its streams only")
	}
	if stream := (Scope{RepoID: 1, StreamID: 2}); !stream.Contains(1, 2) || stream.Contains(1, 4) {
		t.Fatal("stream scope contains only itself")
	}
}

func TestFromContext(t *testing.T) {
	repo, stream := int64(1), int64(2)
	if got := FromContext(&api.ActiveContext{RepoID: &repo, StreamID: &stream}); got != (Scope{RepoID: 1, StreamID: 2}) {
		t.Fatalf("got %+v", got)
	}
	if !FromContext(nil).IsAll() {
		t.Fatal("nil context is All")
	}
}
