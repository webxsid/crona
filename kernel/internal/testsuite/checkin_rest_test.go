package testsuite

import (
	"context"
	"fmt"
	"testing"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	sharedtypes "crona/shared/types"
)

func TestSpecificRestDatePreservesFocusAndCheckInStreaks(t *testing.T) {
	ctx := context.Background()
	coreCtx, _ := newTestCoreContext(t, func() string { return "2026-03-29T12:00:00Z" })
	now := "2026-03-29T12:00:00Z"

	repo := mustCreateRepo(t, ctx, coreCtx.Repos, coreCtx.UserID, now, 1, "Work")
	stream := mustCreateStream(t, ctx, coreCtx.Streams, coreCtx.UserID, now, 1, repo.ID, "app")

	seedFocusDay(t, ctx, coreCtx, stream.ID, 101, "2026-03-27")
	seedFocusDay(t, ctx, coreCtx, stream.ID, 102, "2026-03-28")
	seedCheckInDay(t, ctx, coreCtx, "2026-03-27")
	seedCheckInDay(t, ctx, coreCtx, "2026-03-28")

	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyRestSpecificDates, []string{"2026-03-29"}); err != nil {
		t.Fatalf("set rest specific dates: %v", err)
	}

	streaks, err := corecommands.ComputeMetricsStreaks(ctx, coreCtx, "2026-03-27", "2026-03-29")
	if err != nil {
		t.Fatalf("compute streaks: %v", err)
	}
	if streaks.CurrentFocusDays != 2 || streaks.LongestFocusDays != 2 {
		t.Fatalf("unexpected focus streaks: %+v", streaks)
	}
	if streaks.CurrentCheckInDays != 2 || streaks.LongestCheckInDays != 2 {
		t.Fatalf("unexpected check-in streaks: %+v", streaks)
	}
}

func TestAwayModeOnlyFreezesSelectedStreakKinds(t *testing.T) {
	ctx := context.Background()
	coreCtx, _ := newTestCoreContext(t, func() string { return "2026-03-29T12:00:00Z" })
	now := "2026-03-29T12:00:00Z"

	repo := mustCreateRepo(t, ctx, coreCtx.Repos, coreCtx.UserID, now, 1, "Work")
	stream := mustCreateStream(t, ctx, coreCtx.Streams, coreCtx.UserID, now, 1, repo.ID, "app")

	seedFocusDay(t, ctx, coreCtx, stream.ID, 201, "2026-03-27")
	seedFocusDay(t, ctx, coreCtx, stream.ID, 202, "2026-03-28")
	seedCheckInDay(t, ctx, coreCtx, "2026-03-27")
	seedCheckInDay(t, ctx, coreCtx, "2026-03-28")

	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyAwayModeEnabled, true); err != nil {
		t.Fatalf("set away mode: %v", err)
	}
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyFrozenStreakKinds, []string{string(sharedtypes.StreakKindCheckInDays)}); err != nil {
		t.Fatalf("set frozen streak kinds: %v", err)
	}

	streaks, err := corecommands.ComputeMetricsStreaks(ctx, coreCtx, "2026-03-27", "2026-03-29")
	if err != nil {
		t.Fatalf("compute streaks: %v", err)
	}
	if streaks.CurrentFocusDays != 0 || streaks.LongestFocusDays != 2 {
		t.Fatalf("expected focus streak to break on away day when not selected, got %+v", streaks)
	}
	if streaks.CurrentCheckInDays != 2 || streaks.LongestCheckInDays != 2 {
		t.Fatalf("expected check-in streak to stay protected, got %+v", streaks)
	}
}

func seedFocusDay(t *testing.T, ctx context.Context, coreCtx *core.Context, streamID, issueID int64, date string) {
	t.Helper()

	issue, err := coreCtx.Issues.Create(ctx, sharedtypes.Issue{
		ID:          issueID,
		StreamID:    streamID,
		Title:       fmt.Sprintf("Issue %d", issueID),
		Status:      sharedtypes.IssueStatusPlanned,
		TodoForDate: stringPtr(date),
	}, coreCtx.UserID, date+"T08:00:00Z")
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	sessionID := fmt.Sprintf("session-%d", issueID)
	if _, err := coreCtx.Sessions.Start(ctx, sharedtypes.Session{
		ID:        sessionID,
		IssueID:   issue.ID,
		StartTime: date + "T09:00:00Z",
	}, coreCtx.UserID, coreCtx.DeviceID, date+"T09:00:00Z"); err != nil {
		t.Fatalf("start session: %v", err)
	}
	if _, err := coreCtx.Sessions.Stop(ctx, sessionID, struct {
		EndTime         string
		DurationSeconds int
		Notes           *string
	}{
		EndTime:         date + "T10:00:00Z",
		DurationSeconds: 3600,
	}, coreCtx.UserID, coreCtx.DeviceID, date+"T10:00:00Z"); err != nil {
		t.Fatalf("stop session: %v", err)
	}
}

func seedCheckInDay(t *testing.T, ctx context.Context, coreCtx *core.Context, date string) {
	t.Helper()

	if _, err := coreCtx.DailyCheckIns.Upsert(ctx, sharedtypes.DailyCheckIn{
		Date:      date,
		Mood:      3,
		Energy:    3,
		CreatedAt: date + "T08:30:00Z",
		UpdatedAt: date + "T08:30:00Z",
	}, coreCtx.UserID, coreCtx.DeviceID, date+"T08:30:00Z"); err != nil {
		t.Fatalf("upsert check-in: %v", err)
	}
}

func stringPtr(value string) *string {
	return &value
}
