package commands

import (
	"context"
	"path/filepath"
	"testing"

	"crona/kernel/internal/core"
	"crona/kernel/internal/events"
	"crona/kernel/internal/store"
	"crona/shared/config"
	sharedtypes "crona/shared/types"
)

func TestTimerStructuredBoundaryPreparesBreakWhenAutoStartBreaksDisabled(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, false, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	state, err := service.applyBoundaryTransition(
		ctx,
		session.ID,
		sharedtypes.SessionSegmentWork,
		&boundaryResult{NextSegment: sharedtypes.SessionSegmentShortBreak},
	)
	if err != nil {
		t.Fatalf("apply boundary: %v", err)
	}
	if state.State != "ready" {
		t.Fatalf("expected ready state, got %q", state.State)
	}
	if state.ReadySegmentType == nil ||
		*state.ReadySegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected prepared short break, got %+v", state.ReadySegmentType)
	}
	if state.ElapsedSeconds != 0 {
		t.Fatalf("expected zero elapsed for prepared segment, got %d", state.ElapsedSeconds)
	}
	activeSegment, err := coreCtx.SessionSegments.GetActive(
		ctx,
		coreCtx.UserID,
		coreCtx.DeviceID,
		session.ID,
	)
	if err != nil {
		t.Fatalf("get active segment: %v", err)
	}
	if activeSegment != nil {
		t.Fatalf("expected no active segment for prepared break, got %+v", activeSegment)
	}
}

func TestTimerStructuredBoundaryAutoStartsBreakWhenEnabled(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, true, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	state, err := service.applyBoundaryTransition(
		ctx,
		session.ID,
		sharedtypes.SessionSegmentWork,
		&boundaryResult{NextSegment: sharedtypes.SessionSegmentShortBreak},
	)
	if err != nil {
		t.Fatalf("apply boundary: %v", err)
	}
	if state.State != "paused" {
		t.Fatalf("expected paused state for active break, got %q", state.State)
	}
	if state.SegmentType == nil || *state.SegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected active short break, got %+v", state.SegmentType)
	}
	if state.ReadySegmentType != nil {
		t.Fatalf("expected no prepared segment, got %+v", state.ReadySegmentType)
	}
}

func TestTimerStructuredBoundaryPreparesWorkWhenAutoStartWorkDisabled(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, true, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if err := PauseSession(ctx, coreCtx, sharedtypes.SessionSegmentShortBreak); err != nil {
		t.Fatalf("start short break: %v", err)
	}
	state, err := service.applyBoundaryTransition(
		ctx,
		session.ID,
		sharedtypes.SessionSegmentShortBreak,
		&boundaryResult{NextSegment: sharedtypes.SessionSegmentWork},
	)
	if err != nil {
		t.Fatalf("apply boundary: %v", err)
	}
	if state.State != "ready" {
		t.Fatalf("expected ready state, got %q", state.State)
	}
	if state.ReadySegmentType == nil || *state.ReadySegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected prepared work segment, got %+v", state.ReadySegmentType)
	}
	if state.ElapsedSeconds != 0 {
		t.Fatalf("expected zero elapsed for prepared work, got %d", state.ElapsedSeconds)
	}
}

func TestTimerStructuredBoundaryAutoStartsWorkWhenEnabled(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, true, true)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if err := PauseSession(ctx, coreCtx, sharedtypes.SessionSegmentShortBreak); err != nil {
		t.Fatalf("start short break: %v", err)
	}
	state, err := service.applyBoundaryTransition(
		ctx,
		session.ID,
		sharedtypes.SessionSegmentShortBreak,
		&boundaryResult{NextSegment: sharedtypes.SessionSegmentWork},
	)
	if err != nil {
		t.Fatalf("apply boundary: %v", err)
	}
	if state.State != "running" {
		t.Fatalf("expected running work state, got %q", state.State)
	}
	if state.SegmentType == nil || *state.SegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected active work segment, got %+v", state.SegmentType)
	}
}

func TestTimerStructuredBoundaryPreparesWorkAfterLongBreakWhenAutoStartWorkDisabled(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, true, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if err := PauseSession(ctx, coreCtx, sharedtypes.SessionSegmentLongBreak); err != nil {
		t.Fatalf("start long break: %v", err)
	}
	state, err := service.applyBoundaryTransition(
		ctx,
		session.ID,
		sharedtypes.SessionSegmentLongBreak,
		&boundaryResult{NextSegment: sharedtypes.SessionSegmentWork},
	)
	if err != nil {
		t.Fatalf("apply boundary: %v", err)
	}
	if state.State != "ready" {
		t.Fatalf("expected ready state, got %q", state.State)
	}
	if state.ReadySegmentType == nil || *state.ReadySegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected prepared work segment after long break, got %+v", state.ReadySegmentType)
	}
}

func TestTimerAdvanceStartsPreparedSegment(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, false, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if _, err := service.applyBoundaryTransition(ctx, session.ID, sharedtypes.SessionSegmentWork, &boundaryResult{NextSegment: sharedtypes.SessionSegmentShortBreak}); err != nil {
		t.Fatalf("prepare short break: %v", err)
	}
	state, err := service.Advance(ctx)
	if err != nil {
		t.Fatalf("advance timer: %v", err)
	}
	if state.State != "paused" {
		t.Fatalf("expected paused active break after advance, got %q", state.State)
	}
	if state.SegmentType == nil || *state.SegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected active short break after advance, got %+v", state.SegmentType)
	}
	if state.ReadySegmentType != nil {
		t.Fatalf(
			"expected prepared segment to clear after advance, got %+v",
			state.ReadySegmentType,
		)
	}
}

func TestTimerAdvanceFromActiveStructuredWorkStartsNextBreak(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, false, false)

	if _, err := StartSession(ctx, coreCtx, issue.ID); err != nil {
		t.Fatalf("start session: %v", err)
	}
	state, err := service.GetState(ctx)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.NextSegmentType == nil ||
		*state.NextSegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf(
			"expected active work to advertise short break next, got %+v",
			state.NextSegmentType,
		)
	}
	state, err = service.Advance(ctx)
	if err != nil {
		t.Fatalf("advance active work: %v", err)
	}
	if state.SegmentType == nil || *state.SegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected short break after manual advance, got %+v", state.SegmentType)
	}
}

func TestTimerAdvanceFromActiveStructuredBreakStartsWork(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, true, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if err := PauseSession(ctx, coreCtx, sharedtypes.SessionSegmentShortBreak); err != nil {
		t.Fatalf("start short break: %v", err)
	}
	state, err := service.GetState(ctx)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.NextSegmentType == nil || *state.NextSegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected active break to advertise work next, got %+v", state.NextSegmentType)
	}
	state, err = service.Advance(ctx)
	if err != nil {
		t.Fatalf("advance active break: %v", err)
	}
	if state.State != "running" {
		t.Fatalf("expected running work after advancing break, got %q", state.State)
	}
	if state.SegmentType == nil || *state.SegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected work after manual advance, got %+v", state.SegmentType)
	}
	activeSegment, err := coreCtx.SessionSegments.GetActive(
		ctx,
		coreCtx.UserID,
		coreCtx.DeviceID,
		session.ID,
	)
	if err != nil {
		t.Fatalf("get active segment: %v", err)
	}
	if activeSegment == nil || activeSegment.SegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected persisted active work segment, got %+v", activeSegment)
	}
}

func TestTimerRecoverBoundaryPreservesPreparedSegment(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, false, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if _, err := service.applyBoundaryTransition(ctx, session.ID, sharedtypes.SessionSegmentWork, &boundaryResult{NextSegment: sharedtypes.SessionSegmentShortBreak}); err != nil {
		t.Fatalf("prepare short break: %v", err)
	}
	if err := service.RecoverBoundary(ctx); err != nil {
		t.Fatalf("recover boundary: %v", err)
	}
	state, err := service.GetState(ctx)
	if err != nil {
		t.Fatalf("get timer state: %v", err)
	}
	if state.State != "ready" {
		t.Fatalf("expected prepared state after recovery, got %q", state.State)
	}
	if state.ReadySegmentType == nil ||
		*state.ReadySegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected prepared short break after recovery, got %+v", state.ReadySegmentType)
	}
	if state.ElapsedSeconds != 0 {
		t.Fatalf("expected zero elapsed after recovery, got %d", state.ElapsedSeconds)
	}
}

func newTimerTestContext(
	t *testing.T,
	now func() string,
) (*core.Context, *TimerService, sharedtypes.Issue) {
	t.Helper()

	base := t.TempDir()
	t.Setenv(config.EnvVarRuntimeDir, base)

	db, err := store.Open(filepath.Join(base, "crona.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if err := store.InitSchema(context.Background(), db.DB()); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	registry := store.NewRegistry(db.DB())
	coreCtx := core.NewContext(
		db,
		registry,
		"local",
		"test-device",
		filepath.Join(base, "scratch"),
		now,
		events.NewBus(),
	)
	if err := coreCtx.InitDefaults(context.Background()); err != nil {
		t.Fatalf("init defaults: %v", err)
	}

	repo, err := CreateRepo(context.Background(), coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{Name: "Work"})
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	stream, err := CreateStream(context.Background(), coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: repo.ID, Name: "app"})
	if err != nil {
		t.Fatalf("create stream: %v", err)
	}
	issue, err := CreateIssue(context.Background(), coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Structured timer test"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	return coreCtx, GetTimerService(coreCtx), issue
}

func configureStructuredTimer(
	t *testing.T,
	ctx context.Context,
	coreCtx *core.Context,
	autoStartBreaks bool,
	autoStartWork bool,
) {
	t.Helper()

	settings := map[sharedtypes.CoreSettingsKey]any{
		sharedtypes.CoreSettingsKeyTimerMode:             string(sharedtypes.TimerModeStructured),
		sharedtypes.CoreSettingsKeyBreaksEnabled:         true,
		sharedtypes.CoreSettingsKeyWorkDurationMinutes:   25,
		sharedtypes.CoreSettingsKeyShortBreakMinutes:     5,
		sharedtypes.CoreSettingsKeyLongBreakMinutes:      15,
		sharedtypes.CoreSettingsKeyLongBreakEnabled:      true,
		sharedtypes.CoreSettingsKeyCyclesBeforeLongBreak: 4,
		sharedtypes.CoreSettingsKeyAutoStartBreaks:       autoStartBreaks,
		sharedtypes.CoreSettingsKeyAutoStartWork:         autoStartWork,
	}
	for key, value := range settings {
		if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, key, value); err != nil {
			t.Fatalf("set %s: %v", key, err)
		}
	}
}
