package commands

import (
	"context"
	"strings"
	"testing"

	"crona/kernel/internal/sessionnotes"
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

func TestUpdateSessionHistoryPreservesNotesAndReplacesSegments(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, _, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, coreCtx, issue.ID)
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyManageHistory, true); err != nil {
		t.Fatalf("enable history management: %v", err)
	}
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyManageHistoryDays, 0); err != nil {
		t.Fatalf("set history window: %v", err)
	}
	commit, notes := "old commit", "original notes"
	session, err := LogManualSession(ctx, coreCtx, ManualSessionInput{
		IssueID: issue.ID, Date: "2026-05-24", WorkDurationSeconds: 1800,
		CommitMessage: &commit, Notes: &notes,
	})
	if err != nil {
		t.Fatalf("create manual session: %v", err)
	}
	segmentEnd := "2026-05-24T09:40:00Z"
	_, err = UpdateSessionHistory(ctx, coreCtx, shareddto.UpdateSessionHistoryRequest{
		ID: session.ID, StartTime: "2026-05-24T09:00:00Z", EndTime: "2026-05-24T09:40:00Z",
		EditableNotes: sharedtypes.ParsedSessionNotes{
			sharedtypes.SessionNoteSectionCommit: "updated commit",
			sharedtypes.SessionNoteSectionWork:   "updated work",
			sharedtypes.SessionNoteSectionNotes:  "updated notes",
		},
		Segments: []sharedtypes.SessionSegment{{SegmentType: sharedtypes.SessionSegmentWork,
			StartTime: "2026-05-24T09:00:00Z", EndTime: &segmentEnd}},
	})
	if err != nil {
		t.Fatalf("update session history: %v", err)
	}
	detail, err := GetSessionDetail(ctx, coreCtx, session.ID)
	if err != nil {
		t.Fatalf("get updated detail: %v", err)
	}
	if detail.ParsedNotes[sharedtypes.SessionNoteSectionContext] == "" {
		t.Errorf("expected generated context to survive, got %+v", detail.ParsedNotes)
	}
	for _, want := range []string{"::context", "::commit\nupdated commit", "::work\nupdated work", "::notes\nupdated notes"} {
		if detail.Notes == nil || !strings.Contains(*detail.Notes, want) {
			t.Errorf("updated notes lost %q: %v", want, detail.Notes)
		}
	}
	if len(detail.Segments) != 1 || detail.Segments[0].SegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected replacement work segment, got %+v", detail.Segments)
	}
	parsed := sessionnotes.Parse(detail.Notes)
	if parsed[sharedtypes.SessionNoteSectionCommit] != "updated commit" {
		t.Fatalf("expected edited commit section, got %+v", parsed)
	}
}

func TestSessionHistoryManagementRequiresEnabledAndWithinWindow(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, _, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, coreCtx, issue.ID)
	commit := "today"
	today, err := LogManualSession(ctx, coreCtx, ManualSessionInput{
		IssueID: issue.ID, Date: "2026-05-24", WorkDurationSeconds: 60, CommitMessage: &commit,
	})
	if err != nil {
		t.Fatalf("create today's session: %v", err)
	}
	request := shareddto.UpdateSessionHistoryRequest{ID: today.ID,
		StartTime: "2026-05-24T09:00:00Z", EndTime: "2026-05-24T09:01:00Z"}
	if _, err := UpdateSessionHistory(ctx, coreCtx, request); err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled history edit rejection, got %v", err)
	}
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyManageHistory, true); err != nil {
		t.Fatalf("enable history management: %v", err)
	}
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyManageHistoryDays, 0); err != nil {
		t.Fatalf("set zero day window: %v", err)
	}
	if _, err := UpdateSessionHistory(ctx, coreCtx, request); err != nil {
		t.Fatalf("expected today's edit inside zero-day window: %v", err)
	}
	oldCommit := "old"
	old, err := LogManualSession(ctx, coreCtx, ManualSessionInput{
		IssueID: issue.ID, Date: "2026-05-23", WorkDurationSeconds: 60, CommitMessage: &oldCommit,
	})
	if err != nil {
		t.Fatalf("create yesterday's session: %v", err)
	}
	if _, err := DeleteSessionHistory(ctx, coreCtx, old.ID); err == nil ||
		!strings.Contains(err.Error(), "History Edit Window starts on") ||
		!strings.Contains(err.Error(), "Increase it in Settings") {
		t.Fatalf("expected actionable out-of-window delete rejection, got %v", err)
	}
}

func TestUpdateSessionHistoryRejectsInvalidSessionAndSegmentRanges(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, _, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, coreCtx, issue.ID)
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyManageHistory, true); err != nil {
		t.Fatalf("enable history management: %v", err)
	}
	commit := "session"
	session, err := LogManualSession(ctx, coreCtx, ManualSessionInput{
		IssueID: issue.ID, Date: "2026-05-24", WorkDurationSeconds: 120,
		BreakDurationSeconds: 60, CommitMessage: &commit,
	})
	if err != nil {
		t.Fatalf("create manual session: %v", err)
	}
	detail, err := GetSessionDetail(ctx, coreCtx, session.ID)
	if err != nil {
		t.Fatalf("get session detail: %v", err)
	}
	base := shareddto.UpdateSessionHistoryRequest{
		ID: session.ID, StartTime: detail.StartTime, EndTime: *detail.EndTime,
		Segments: detail.Segments,
	}
	invalid := []struct {
		name   string
		change func(*shareddto.UpdateSessionHistoryRequest)
	}{
		{name: "malformed session time", change: func(request *shareddto.UpdateSessionHistoryRequest) {
			request.EndTime = "not-a-time"
		}},
		{name: "end before start", change: func(request *shareddto.UpdateSessionHistoryRequest) {
			request.EndTime = request.StartTime
		}},
		{name: "segment outside session", change: func(request *shareddto.UpdateSessionHistoryRequest) {
			request.Segments = []sharedtypes.SessionSegment{{SegmentType: sharedtypes.SessionSegmentWork,
				StartTime: "2026-05-24T08:00:00Z", EndTime: historyTestStringPointer("2026-05-24T08:01:00Z")}}
		}},
		{name: "overlapping segments", change: func(request *shareddto.UpdateSessionHistoryRequest) {
			request.Segments = append(request.Segments, request.Segments[0])
		}},
		{name: "invalid segment type", change: func(request *shareddto.UpdateSessionHistoryRequest) {
			request.Segments[0].SegmentType = "unknown"
		}},
	}
	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			request := base
			request.Segments = append([]sharedtypes.SessionSegment(nil), base.Segments...)
			test.change(&request)
			if _, err := UpdateSessionHistory(ctx, coreCtx, request); err == nil {
				t.Fatal("expected invalid history update to fail")
			}
		})
	}
}

func historyTestStringPointer(value string) *string { return &value }
