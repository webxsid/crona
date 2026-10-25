package testsuite

import (
	"context"
	"slices"
	"testing"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	sharedtypes "crona/shared/types"
)

func newTagFixture(t *testing.T) (context.Context, *core.Context, sharedtypes.Issue, sharedtypes.Issue, sharedtypes.Habit) {
	t.Helper()
	ctx, coreCtx, _, first := newCarryFixture(t)
	second := mustCreateCommandIssue(t, ctx, coreCtx, first.StreamID, "Second task", nil)
	habit, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: first.StreamID, Name: "Walk", ScheduleType: "daily"})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}
	return ctx, coreCtx, first, second, habit
}

func tagNamed(tags []sharedtypes.Tag, name string) *sharedtypes.Tag {
	for i := range tags {
		if tags[i].Name == name {
			return &tags[i]
		}
	}
	return nil
}

func TestSystemTagsAreSeededAndProtected(t *testing.T) {
	ctx, coreCtx, _, _, _ := newTagFixture(t)
	tags, err := corecommands.ListTags(ctx, coreCtx)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	for _, name := range sharedtypes.SystemTags {
		if tag := tagNamed(tags, name); tag == nil || !tag.System {
			t.Fatalf("expected system tag %q, got %+v", name, tags)
		}
	}
	if err := corecommands.DeleteTag(ctx, coreCtx, "urgent"); err == nil {
		t.Fatal("expected deleting a system tag to fail")
	}
	if err := corecommands.RenameTag(ctx, coreCtx, "urgent", "asap", false); err == nil {
		t.Fatal("expected renaming a system tag to fail")
	}
	if err := corecommands.SetTagColor(ctx, coreCtx, "urgent", new("red")); err != nil {
		t.Fatalf("expected a system tag's color to change: %v", err)
	}
	if removed, _ := corecommands.RemoveUnusedTags(ctx, coreCtx); len(removed) != 0 {
		t.Fatalf("expected unused system tags to stay, removed %v", removed)
	}
}

func TestIssueAndHabitTagsRoundTrip(t *testing.T) {
	ctx, coreCtx, first, _, habit := newTagFixture(t)
	issue, err := corecommands.SetIssueTags(ctx, coreCtx, first.ID, []string{"Quick", "+client acme", "quick"})
	if err != nil {
		t.Fatalf("set issue tags: %v", err)
	}
	if !slices.Equal(issue.Tags, []string{"client-acme", "quick"}) {
		t.Fatalf("expected normalized tags, got %v", issue.Tags)
	}
	all, _ := corecommands.ListAllIssues(ctx, coreCtx)
	for _, item := range all {
		if item.ID == first.ID && !slices.Equal(item.Tags, issue.Tags) {
			t.Fatalf("expected list_all to carry tags, got %v", item.Tags)
		}
	}
	tagged, err := corecommands.SetHabitTags(ctx, coreCtx, habit.ID, []string{"health"})
	if err != nil || !slices.Equal(tagged.Tags, []string{"health"}) {
		t.Fatalf("set habit tags: %v %+v", err, tagged)
	}
	if _, err := corecommands.SetIssueTags(ctx, coreCtx, first.ID, []string{"bad tag!"}); err == nil {
		t.Fatal("expected an invalid tag name to be rejected")
	}
	tags, _ := corecommands.ListTags(ctx, coreCtx)
	if tag := tagNamed(tags, "quick"); tag == nil || tag.IssueCount != 1 {
		t.Fatalf("expected quick used once, got %+v", tag)
	}
	if tag := tagNamed(tags, "health"); tag == nil || tag.HabitCount != 1 {
		t.Fatalf("expected health on one habit, got %+v", tag)
	}
}

func TestBulkUpdateIssueTags(t *testing.T) {
	ctx, coreCtx, first, second, _ := newTagFixture(t)
	if _, err := corecommands.SetIssueTags(ctx, coreCtx, first.ID, []string{"waiting"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	updated, err := corecommands.UpdateIssueTags(ctx, coreCtx, []int64{first.ID, second.ID}, []string{"urgent"}, []string{"waiting"})
	if err != nil {
		t.Fatalf("bulk update: %v", err)
	}
	for _, issue := range updated {
		if !slices.Equal(issue.Tags, []string{"urgent"}) {
			t.Fatalf("expected only urgent on issue %d, got %v", issue.ID, issue.Tags)
		}
	}
}

func TestRenameMergeAndDeleteTags(t *testing.T) {
	ctx, coreCtx, first, second, _ := newTagFixture(t)
	if _, err := corecommands.SetIssueTags(ctx, coreCtx, first.ID, []string{"bug"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, err := corecommands.SetIssueTags(ctx, coreCtx, second.ID, []string{"defect", "bug"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := corecommands.RenameTag(ctx, coreCtx, "defect", "bug", false); err == nil {
		t.Fatal("expected renaming onto an existing tag to need merge")
	}
	if err := corecommands.RenameTag(ctx, coreCtx, "defect", "bug", true); err != nil {
		t.Fatalf("merge: %v", err)
	}
	issue, _ := coreCtx.Issues.GetByID(ctx, second.ID, coreCtx.UserID)
	if !slices.Equal(issue.Tags, []string{"bug"}) {
		t.Fatalf("expected merged tags, got %v", issue.Tags)
	}
	if err := corecommands.RenameTag(ctx, coreCtx, "bug", "defect", false); err != nil {
		t.Fatalf("rename: %v", err)
	}
	issue, _ = coreCtx.Issues.GetByID(ctx, first.ID, coreCtx.UserID)
	if !slices.Equal(issue.Tags, []string{"defect"}) {
		t.Fatalf("expected renamed tag, got %v", issue.Tags)
	}
	if err := corecommands.DeleteTag(ctx, coreCtx, "defect"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	issue, _ = coreCtx.Issues.GetByID(ctx, first.ID, coreCtx.UserID)
	if len(issue.Tags) != 0 {
		t.Fatalf("expected tag removed from issue, got %v", issue.Tags)
	}
	if _, err := corecommands.SetIssueTags(ctx, coreCtx, first.ID, []string{"temp"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, err := corecommands.SetIssueTags(ctx, coreCtx, first.ID, nil); err != nil {
		t.Fatalf("clear: %v", err)
	}
	removed, err := corecommands.RemoveUnusedTags(ctx, coreCtx)
	if err != nil || !slices.Equal(removed, []string{"temp"}) {
		t.Fatalf("expected temp removed, got %v %v", removed, err)
	}
}
