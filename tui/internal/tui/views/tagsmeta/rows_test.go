package tagsmeta

import (
	"testing"

	sharedtypes "crona/shared/types"
)

func TestSelectedFollowsTheFilterAndUnusedSkipsSystemTags(t *testing.T) {
	tags := []sharedtypes.Tag{
		{Name: "urgent", System: true},
		{Name: "release", IssueCount: 1},
		{Name: "old"},
	}
	if got := Unused(tags); got != 1 {
		t.Fatalf("expected 1 unused tag, got %d", got)
	}
	tag, ok := Selected(tags, "rel", 0)
	if !ok || tag.Name != "release" || Usage(tag) != "1 issue, 0 habits" {
		t.Fatalf("unexpected selection %+v", tag)
	}
}

func TestFilteredIndicesPutSystemTagsFirst(t *testing.T) {
	tags := []sharedtypes.Tag{{Name: "alpha"}, {Name: "urgent", System: true}, {Name: "beta"}}
	got := FilteredIndices(tags, "")
	if len(got) != 3 || got[0] != 1 || got[1] != 0 || got[2] != 2 {
		t.Fatalf("unexpected order %v", got)
	}
}

func TestTaggedIssuesPutOpenFirstAndCountUntagged(t *testing.T) {
	issue := func(title string, status sharedtypes.IssueStatus, tags ...string) sharedtypes.IssueWithMeta {
		return sharedtypes.IssueWithMeta{Issue: sharedtypes.Issue{Title: title, Status: status, Tags: tags}}
	}
	issues := []sharedtypes.IssueWithMeta{
		issue("shipped", sharedtypes.IssueStatusDone, "release"),
		issue("draft", sharedtypes.IssueStatusPlanned, "release"),
		issue("loose", sharedtypes.IssueStatusPlanned),
	}
	got := TaggedIssues("release", issues)
	if len(got) != 2 || got[0].Title != "draft" || got[1].Title != "shipped" {
		t.Fatalf("unexpected tagged issues %+v", got)
	}
	if open, closed := OpenDone("release", issues); open != 1 || closed != 1 {
		t.Fatalf("unexpected counts %d open %d done", open, closed)
	}
	if openIssues, _ := Untagged(issues, nil); openIssues != 1 {
		t.Fatalf("expected 1 untagged open issue, got %d", openIssues)
	}
}
