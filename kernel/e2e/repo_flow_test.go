//go:build e2e

package e2e

import (
	"testing"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestRepoStreamIssueLifecycleOverIPC(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)

	repoA := createRepo(t, kernel, "Office")
	repoB := createRepo(t, kernel, "Home")

	if repoA.ID != 1 {
		t.Fatalf("expected first repo public ID to be 1, got %d", repoA.ID)
	}
	if repoB.ID != 2 {
		t.Fatalf("expected second repo public ID to be 2, got %d", repoB.ID)
	}

	var repos []sharedtypes.Repo
	kernel.call(t, protocol.MethodRepoList, nil, &repos)
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}

	var updated sharedtypes.Repo
	kernel.call(t, protocol.MethodRepoUpdate, map[string]any{
		"id":   repoA.ID,
		"name": "Office HQ",
	}, &updated)
	if updated.Name != "Office HQ" {
		t.Fatalf("expected updated repo name, got %q", updated.Name)
	}

	stream := createStream(t, kernel, repoA.ID, "main")
	if stream.ID != 1 {
		t.Fatalf("expected first stream public ID to be 1, got %d", stream.ID)
	}

	estimate := 30
	issue := createIssue(t, kernel, stream.ID, "Ship release", &estimate)
	if issue.ID != 1 {
		t.Fatalf("expected first issue public ID to be 1, got %d", issue.ID)
	}

	var issues []sharedtypes.Issue
	kernel.call(t, protocol.MethodIssueList, shareddto.ListIssuesQuery{StreamID: stream.ID}, &issues)
	if len(issues) != 1 || issues[0].Title != "Ship release" {
		t.Fatalf("unexpected issue list: %+v", issues)
	}

	var deleted shareddto.OKResponse
	kernel.call(t, protocol.MethodRepoDelete, shareddto.NumericIDRequest{ID: repoB.ID}, &deleted)
	if !deleted.OK {
		t.Fatalf("expected repo delete ok response")
	}

	kernel.call(t, protocol.MethodRepoList, nil, &repos)
	if len(repos) != 1 || repos[0].ID != repoA.ID {
		t.Fatalf("unexpected repos after delete: %+v", repos)
	}
}
