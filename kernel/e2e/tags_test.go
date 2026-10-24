//go:build e2e

package e2e

import (
	"slices"
	"testing"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestTagsOverIPC(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)

	repo := createRepo(t, kernel, "Work")
	stream := createStream(t, kernel, repo.ID, "app")

	var tagged sharedtypes.Issue
	kernel.call(t, protocol.MethodIssueCreate, shareddto.CreateIssueRequest{
		StreamID: stream.ID, Title: "Fix login", Tags: []string{"Urgent", "bug"},
	}, &tagged)
	if !slices.Equal(tagged.Tags, []string{"bug", "urgent"}) {
		t.Fatalf("expected create to apply tags, got %v", tagged.Tags)
	}
	plain := createIssue(t, kernel, stream.ID, "Write docs", nil)

	var filtered []sharedtypes.IssueWithMeta
	kernel.call(t, protocol.MethodIssueListAll, shareddto.ListAllQuery{
		Tags: &sharedtypes.TagFilter{Any: []string{"urgent"}},
	}, &filtered)
	if len(filtered) != 1 || filtered[0].ID != tagged.ID {
		t.Fatalf("expected only the urgent issue, got %+v", filtered)
	}
	kernel.call(t, protocol.MethodIssueListAll, shareddto.ListAllQuery{
		Tags: &sharedtypes.TagFilter{Any: []string{sharedtypes.TagUntagged}},
	}, &filtered)
	if len(filtered) != 1 || filtered[0].ID != plain.ID {
		t.Fatalf("expected only the untagged issue, got %+v", filtered)
	}

	var updated sharedtypes.Issue
	kernel.call(t, protocol.MethodIssueUpdate, map[string]any{"id": plain.ID, "tags": []string{"quick"}}, &updated)
	if !slices.Equal(updated.Tags, []string{"quick"}) {
		t.Fatalf("expected update to set tags, got %v", updated.Tags)
	}

	var tags []sharedtypes.Tag
	kernel.call(t, protocol.MethodTagList, nil, &tags)
	names := []string{}
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	for _, want := range append([]string{"bug"}, sharedtypes.SystemTags...) {
		if !slices.Contains(names, want) {
			t.Fatalf("expected tag %q in %v", want, names)
		}
	}
}
