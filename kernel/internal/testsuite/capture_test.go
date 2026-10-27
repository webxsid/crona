package testsuite

import (
	"slices"
	"testing"

	corecommands "crona/kernel/internal/core/commands"
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

func TestIssueCaptureParsesAndImports(t *testing.T) {
	ctx, coreCtx, _, existing := newCarryFixture(t)
	proposals, err := corecommands.ParseIssueCapture(ctx, coreCtx, shareddto.IssueCaptureParseRequest{
		Text: "- fix login #app +urgent for 45 min today\n  after the SSO change\n- call bank\n- broken #nope",
	})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(proposals) != 3 {
		t.Fatalf("want 3 proposals, got %+v", proposals)
	}
	if proposals[0].StreamID != existing.StreamID || len(proposals[0].Flags) != 0 {
		t.Fatalf("first line should target the app stream unflagged: %+v", proposals[0])
	}
	if !proposals[1].Inbox {
		t.Fatalf("untargeted line should go to the Inbox: %+v", proposals[1])
	}
	if len(proposals[2].Flags) == 0 {
		t.Fatalf("unknown target should be flagged: %+v", proposals[2])
	}

	// A row with no stream fails the whole import before anything is created.
	before, _ := corecommands.ListAllIssues(ctx, coreCtx)
	if _, err := corecommands.ImportIssueCapture(ctx, coreCtx, shareddto.IssueCaptureImportRequest{Proposals: proposals}); err == nil {
		t.Fatal("expected an import with an unresolved row to fail")
	}
	if after, _ := corecommands.ListAllIssues(ctx, coreCtx); len(after) != len(before) {
		t.Fatalf("failed import created issues: %d -> %d", len(before), len(after))
	}

	result, err := corecommands.ImportIssueCapture(ctx, coreCtx, shareddto.IssueCaptureImportRequest{Proposals: proposals[:2]})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	first, second := result.Issues[0], result.Issues[1]
	if first.Title != "fix login" || first.EstimateMinutes == nil || *first.EstimateMinutes != 45 ||
		first.TodoForDate == nil || *first.TodoForDate != "2026-03-24" || !slices.Equal(first.Tags, []string{"urgent"}) ||
		first.Description == nil || *first.Description != "after the SSO change" {
		t.Fatalf("unexpected first issue %+v", first)
	}

	// The Inbox repo and stream are created once and reused.
	repos, _ := corecommands.ListRepos(ctx, coreCtx)
	inboxRepos := 0
	for _, repo := range repos {
		if repo.Name == "Inbox" {
			inboxRepos++
		}
	}
	if inboxRepos != 1 || second.StreamID == existing.StreamID {
		t.Fatalf("expected one Inbox repo holding the second issue, repos %+v issue %+v", repos, second)
	}
	again, err := corecommands.ImportIssueCapture(ctx, coreCtx, shareddto.IssueCaptureImportRequest{Proposals: proposals[1:2]})
	if err != nil || again.Issues[0].StreamID != second.StreamID {
		t.Fatalf("second Inbox import should reuse the stream: %v %+v", err, again)
	}
}

func TestInboxIsSystemAndIssuesMoveOut(t *testing.T) {
	ctx, coreCtx, _, existing := newCarryFixture(t)
	result, err := corecommands.ImportIssueCapture(ctx, coreCtx, shareddto.IssueCaptureImportRequest{
		Proposals: []shareddto.IssueCaptureProposal{{Line: 1, Title: "call bank", Inbox: true, Tags: []string{"home"}}},
	})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	inboxIssue := result.Issues[0]
	stream, _ := coreCtx.Streams.GetByID(ctx, inboxIssue.StreamID, coreCtx.UserID)
	repo, _ := coreCtx.Repos.GetByID(ctx, stream.RepoID, coreCtx.UserID)
	if !stream.System || !repo.System {
		t.Fatalf("Inbox should be system: repo %+v stream %+v", repo, stream)
	}
	name := "Later"
	if _, err := corecommands.UpdateStream(ctx, coreCtx, stream.ID, struct {
		Name        *string
		Description sharedtypes.Patch[string]
		Visibility  *sharedtypes.StreamVisibility
	}{Name: &name}); err == nil {
		t.Fatal("renaming the Inbox stream should be refused")
	}
	if err := corecommands.DeleteRepo(ctx, coreCtx, repo.ID); err == nil {
		t.Fatal("deleting the Inbox repo should be refused")
	}

	// Only Inbox issues move, and never into a system stream.
	if _, err := corecommands.MoveIssue(ctx, coreCtx, shareddto.MoveIssueRequest{ID: inboxIssue.ID, StreamID: stream.ID, Title: "x"}); err == nil {
		t.Fatal("moving into the Inbox should be refused")
	}
	moved, err := corecommands.MoveIssue(ctx, coreCtx, shareddto.MoveIssueRequest{
		ID: inboxIssue.ID, StreamID: existing.StreamID, Title: "call the bank", Tags: &[]string{"money"},
	})
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	if moved.StreamID != existing.StreamID || moved.Title != "call the bank" || !slices.Equal(moved.Tags, []string{"money"}) {
		t.Fatalf("unexpected moved issue %+v", moved)
	}
	if _, err := corecommands.MoveIssue(ctx, coreCtx, shareddto.MoveIssueRequest{ID: inboxIssue.ID, StreamID: existing.StreamID, Title: "x"}); err == nil {
		t.Fatal("an issue outside the Inbox should not move")
	}
}
