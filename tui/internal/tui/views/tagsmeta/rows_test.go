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
