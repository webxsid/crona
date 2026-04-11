package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/export"
	runtimepkg "crona/kernel/internal/runtime"
	"crona/kernel/internal/scratchfile"
	"crona/kernel/internal/store"
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

func (h *Handler) clearDevData(ctx context.Context) error {
	h.timer.ClearBoundary()

	if err := store.ClearAllData(ctx, h.core.Store.DB()); err != nil {
		return fmt.Errorf("clear sqlite data: %w", err)
	}
	if err := h.resetRuntimeFiles(ctx); err != nil {
		return err
	}

	payload, _ := json.Marshal(sharedtypes.TimerState{State: "idle"})
	h.core.Events.Emit(sharedtypes.KernelEvent{
		Type:    sharedtypes.EventTypeTimerState,
		Payload: payload,
	})
	return nil
}

func (h *Handler) wipeRuntimeData(ctx context.Context) error {
	h.timer.ClearBoundary()
	if err := store.ClearAllData(ctx, h.core.Store.DB()); err != nil {
		return fmt.Errorf("clear sqlite data: %w", err)
	}
	return h.resetRuntimeFiles(ctx)
}

func (h *Handler) resetRuntimeFiles(ctx context.Context) error {
	if err := runtimepkg.ResetManagedData(h.paths); err != nil {
		return fmt.Errorf("reset runtime files: %w", err)
	}
	if err := runtimepkg.EnsurePaths(h.paths); err != nil {
		return fmt.Errorf("ensure runtime paths: %w", err)
	}
	if err := h.core.InitDefaults(ctx); err != nil {
		return fmt.Errorf("reinitialize defaults: %w", err)
	}
	if _, err := export.EnsureAssets(h.paths); err != nil {
		return fmt.Errorf("ensure export assets: %w", err)
	}
	if err := scratchfile.ClearAll(h.core.ScratchDir); err != nil {
		return fmt.Errorf("clear scratch files: %w", err)
	}
	return nil
}

func (h *Handler) seedDevData(ctx context.Context) error {
	if err := h.clearDevData(ctx); err != nil {
		return err
	}

	baseNow := time.Now().UTC()
	today := baseNow.Format("2006-01-02")
	yesterday := baseNow.Add(-24 * time.Hour).Format("2006-01-02")
	dateAt := func(offset int) string {
		return baseNow.AddDate(0, 0, offset).Format("2006-01-02")
	}

	withSeedNow := func(ts time.Time, fn func() error) error {
		prev := h.core.Now
		h.core.Now = func() string { return ts.UTC().Format(time.RFC3339) }
		defer func() { h.core.Now = prev }()
		return fn()
	}
	seedAt := func(dayOffset, hour, minute int, fn func() error) error {
		ts := time.Date(baseNow.Year(), baseNow.Month(), baseNow.Day(), hour, minute, 0, 0, time.UTC).AddDate(0, 0, dayOffset)
		return withSeedNow(ts, fn)
	}
	createRepo := func(name, description, color string) (sharedtypes.Repo, error) {
		var repo sharedtypes.Repo
		err := seedAt(-6, 8, 0, func() error {
			created, err := corecommands.CreateRepo(ctx, h.core, struct {
				Name        string
				Description *string
				Color       *string
			}{
				Name:        name,
				Description: devStringPtr(description),
				Color:       devStringPtr(color),
			})
			if err != nil {
				return err
			}
			repo = created
			return nil
		})
		return repo, err
	}
	createStream := func(repoID int64, name, description string) (sharedtypes.Stream, error) {
		var stream sharedtypes.Stream
		err := seedAt(-6, 8, 15, func() error {
			created, err := corecommands.CreateStream(ctx, h.core, struct {
				RepoID      int64
				Name        string
				Description *string
				Visibility  *sharedtypes.StreamVisibility
			}{
				RepoID:      repoID,
				Name:        name,
				Description: devStringPtr(description),
			})
			if err != nil {
				return err
			}
			stream = created
			return nil
		})
		return stream, err
	}
	createIssue := func(dayOffset int, streamID int64, title, description string, estimate int, due *string, transitions []sharedtypes.IssueStatus, notes ...string) (sharedtypes.Issue, error) {
		var issue sharedtypes.Issue
		err := seedAt(dayOffset, 9, 0, func() error {
			var notePtr *string
			if len(notes) > 0 && notes[0] != "" {
				notePtr = devStringPtr(notes[0])
			}
			created, err := corecommands.CreateIssue(ctx, h.core, struct {
				StreamID        int64
				Title           string
				Description     *string
				EstimateMinutes *int
				Notes           *string
				TodoForDate     *string
			}{
				StreamID:        streamID,
				Title:           title,
				Description:     devStringPtr(description),
				EstimateMinutes: devIntPtr(estimate),
				Notes:           notePtr,
				TodoForDate:     due,
			})
			if err != nil {
				return err
			}
			issue = created
			currentStatus := issue.Status
			for _, nextStatus := range transitions {
				if !sharedtypes.IsValidIssueTransition(currentStatus, nextStatus) {
					return fmt.Errorf("invalid seeded issue transition for %q: %s -> %s", title, currentStatus, nextStatus)
				}
				if _, err := corecommands.ChangeIssueStatus(ctx, h.core, issue.ID, nextStatus, notePtr); err != nil {
					return err
				}
				currentStatus = nextStatus
			}
			return nil
		})
		return issue, err
	}
	createHabit := func(dayOffset int, streamID int64, name, description, schedule string, weekdays []int, targetMinutes *int) (sharedtypes.Habit, error) {
		var habit sharedtypes.Habit
		err := seedAt(dayOffset, 8, 30, func() error {
			created, err := corecommands.CreateHabit(ctx, h.core, struct {
				StreamID      int64
				Name          string
				Description   *string
				ScheduleType  string
				Weekdays      []int
				TargetMinutes *int
			}{
				StreamID:      streamID,
				Name:          name,
				Description:   devStringPtr(description),
				ScheduleType:  schedule,
				Weekdays:      weekdays,
				TargetMinutes: targetMinutes,
			})
			if err != nil {
				return err
			}
			habit = created
			return nil
		})
		return habit, err
	}
	seedHabitStatus := func(dayOffset int, habitID int64, date string, status sharedtypes.HabitCompletionStatus, durationMinutes *int, notes *string) error {
		return seedAt(dayOffset, 20, 15, func() error {
			_, err := corecommands.CompleteHabit(ctx, h.core, habitID, date, status, durationMinutes, notes)
			return err
		})
	}
	seedCheckIn := func(dayOffset int, date string, mood, energy int, sleepHours float64, sleepScore, screenMinutes int, notes string) error {
		return seedAt(dayOffset, 7, 45, func() error {
			_, err := corecommands.UpsertDailyCheckIn(ctx, h.core, shareddto.DailyCheckInUpsertRequest{
				Date:              date,
				Mood:              mood,
				Energy:            energy,
				SleepHours:        devFloatPtr(sleepHours),
				SleepScore:        devIntPtr(sleepScore),
				ScreenTimeMinutes: devIntPtr(screenMinutes),
				Notes:             devStringPtr(notes),
			})
			return err
		})
	}
	seedSession := func(dayOffset int, issueID int64, startHour, startMinute, durationMinutes int, input corecommands.SessionEndInput) error {
		start := time.Date(baseNow.Year(), baseNow.Month(), baseNow.Day(), startHour, startMinute, 0, 0, time.UTC).AddDate(0, 0, dayOffset)
		if err := withSeedNow(start, func() error {
			_, err := corecommands.StartSession(ctx, h.core, issueID)
			return err
		}); err != nil {
			return err
		}
		return withSeedNow(start.Add(time.Duration(durationMinutes)*time.Minute), func() error {
			_, err := corecommands.StopSession(ctx, h.core, input)
			return err
		})
	}

	workRepo, err := createRepo("Work", "Team delivery, product design, and release operations.", "blue")
	if err != nil {
		return err
	}
	personalRepo, err := createRepo("Personal", "Life admin, routines, and home maintenance.", "green")
	if err != nil {
		return err
	}
	ossRepo, err := createRepo("OSS", "Open-source tooling and documentation experiments.", "magenta")
	if err != nil {
		return err
	}

	appStream, err := createStream(workRepo.ID, "app", "Primary product and UX work.")
	if err != nil {
		return err
	}
	infraStream, err := createStream(workRepo.ID, "infra", "Release, packaging, and deployment plumbing.")
	if err != nil {
		return err
	}
	platformStream, err := createStream(workRepo.ID, "platform", "Shared APIs, reliability, and observability.")
	if err != nil {
		return err
	}
	homeStream, err := createStream(personalRepo.ID, "home", "Household admin and chores.")
	if err != nil {
		return err
	}
	wellbeingStream, err := createStream(personalRepo.ID, "wellbeing", "Health, fitness, and routines.")
	if err != nil {
		return err
	}
	cliStream, err := createStream(ossRepo.ID, "cli", "CLI polish and contributor onboarding.")
	if err != nil {
		return err
	}

	focusIssue, err := createIssue(-6, appStream.ID, "Port dev tooling to Go", "Validate the IPC-first workflow and replace the remaining shell scripts.", 90, devDatePtr(dateAt(-1)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusReady, sharedtypes.IssueStatusInProgress}, "Drive the migration end-to-end and keep dev experience tight.")
	if err != nil {
		return err
	}
	underEstimateIssue, err := createIssue(-5, appStream.ID, "Add keyboard-first command palette", "Speed up common view and dialog actions from one launcher.", 35, devDatePtr(dateAt(-3)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusReady, sharedtypes.IssueStatusInProgress}, "Intentionally underestimated for seed accuracy drift.")
	if err != nil {
		return err
	}
	overEstimateIssue, err := createIssue(-4, appStream.ID, "Review lifecycle UX copy", "Tighten action labels and empty-state wording across the dashboard.", 95, devDatePtr(dateAt(-2)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusInProgress, sharedtypes.IssueStatusInReview}, "Intentionally overestimated for rollup estimate-bias testing.")
	if err != nil {
		return err
	}
	readyIssue, err := createIssue(-6, infraStream.ID, "Prepare rollout checklist", "Capture deploy sequencing, smoke tests, and rollback checkpoints.", 45, devDatePtr(dateAt(-4)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusReady}, "Waiting only on final go/no-go review.")
	if err != nil {
		return err
	}
	inProgressIssue, err := createIssue(-6, infraStream.ID, "Wire release packaging checks", "Verify built artifacts, checksums, and update install docs.", 75, devDatePtr(dateAt(-5)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusInProgress}, "Partially implemented; needs artifact verification.")
	if err != nil {
		return err
	}
	blockedIssue, err := createIssue(-6, infraStream.ID, "Provision CI signing secrets", "Unblock package signing in CI and production release automation.", 30, devDatePtr(dateAt(-4)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusBlocked}, "Awaiting access to the signing account.")
	if err != nil {
		return err
	}
	if _, err := createIssue(-6, platformStream.ID, "Backfill metrics range tests", "Cover streaks, burnout rollups, and historical summaries.", 55, devDatePtr(dateAt(-2)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusInProgress, sharedtypes.IssueStatusDone}); err != nil {
		return err
	}
	if _, err := createIssue(-6, platformStream.ID, "Reduce websocket reconnect churn", "Stabilize event fanout under network flaps.", 50, devDatePtr(dateAt(-1)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusAbandoned}, "Superseded by the unix socket migration."); err != nil {
		return err
	}
	docsIssue, err := createIssue(-3, cliStream.ID, "Improve install docs for Linux", "Add shell completion, troubleshooting, and verification notes.", 35, devDatePtr(dateAt(1)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusReady}, "Future planned work to keep the window from being all closed.")
	if err != nil {
		return err
	}
	if _, err := createIssue(-3, cliStream.ID, "Publish example config bundle", "Ship realistic examples for repo, stream, and timer setup.", 25, devDatePtr(dateAt(0)), nil); err != nil {
		return err
	}
	if _, err := createIssue(-2, homeStream.ID, "Plan weekend errands", "Group shopping, laundry, and pickup tasks into one route.", 30, devDatePtr(dateAt(0)), nil); err != nil {
		return err
	}
	if _, err := createIssue(-6, homeStream.ID, "Research standing desk options", "Compare dimensions and price points for the office nook.", 25, devDatePtr(dateAt(2)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusAbandoned}, "Deferred until the room reorganization is done."); err != nil {
		return err
	}
	if _, err := createIssue(-6, platformStream.ID, "Audit event replay gaps", "Investigate dropped events and stale refresh paths.", 50, devDatePtr(dateAt(-6)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusReady}, "Left incomplete on purpose so the oldest day shows planned-only work."); err != nil {
		return err
	}
	if _, err := createIssue(-5, platformStream.ID, "Reconcile update relaunch UX", "Smooth out install handoff and watchdog reporting.", 40, devDatePtr(dateAt(-5)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusReady}, "Left untouched so one day can register as missed."); err != nil {
		return err
	}
	if err := seedAt(-3, 17, 0, func() error {
		_, err := corecommands.MarkIssueTodoForDate(ctx, h.core, readyIssue.ID, dateAt(-2))
		return err
	}); err != nil {
		return err
	}
	if err := seedAt(-4, 17, 30, func() error {
		_, err := corecommands.MarkIssueTodoForDate(ctx, h.core, blockedIssue.ID, dateAt(-3))
		return err
	}); err != nil {
		return err
	}

	workoutHabit, err := createHabit(-6, wellbeingStream.ID, "Strength Training", "Three lifting sessions each week with basic progression.", "weekly", []int{1, 3, 5}, devIntPtr(60))
	if err != nil {
		return err
	}
	walkHabit, err := createHabit(-6, wellbeingStream.ID, "Morning Walk", "Get outside before work for a short walk.", "daily", nil, devIntPtr(25))
	if err != nil {
		return err
	}
	journalHabit, err := createHabit(-6, wellbeingStream.ID, "Journal", "Capture a quick end-of-day reflection.", "daily", nil, devIntPtr(10))
	if err != nil {
		return err
	}
	inboxHabit, err := createHabit(-6, appStream.ID, "Inbox Zero Sweep", "Clear triage inboxes before the afternoon block.", "weekdays", nil, devIntPtr(15))
	if err != nil {
		return err
	}
	if _, err := createHabit(-6, homeStream.ID, "Laundry Reset", "Run and fold one load every Sunday.", "weekly", []int{0}, devIntPtr(45)); err != nil {
		return err
	}
	if _, err := createHabit(-6, cliStream.ID, "Read Release Notes", "Scan upstream release notes for dependency changes.", "weekly", []int{2, 4}, devIntPtr(20)); err != nil {
		return err
	}

	workoutDate, workoutOffset := mostRecentMatchingWeekday(baseNow, []int{1, 3, 5})
	if err := seedHabitStatus(workoutOffset, workoutHabit.ID, workoutDate, sharedtypes.HabitCompletionStatusCompleted, devIntPtr(58), devStringPtr("Main lifts plus accessory work.")); err != nil {
		return err
	}
	if err := seedHabitStatus(0, walkHabit.ID, today, sharedtypes.HabitCompletionStatusFailed, nil, devStringPtr("Weather broke the morning routine.")); err != nil {
		return err
	}
	inboxDate, inboxOffset := mostRecentMatchingWeekday(baseNow, []int{1, 2, 3, 4, 5})
	if err := seedHabitStatus(inboxOffset, inboxHabit.ID, inboxDate, sharedtypes.HabitCompletionStatusCompleted, devIntPtr(12), devStringPtr("Inbox and notifications cleared before lunch.")); err != nil {
		return err
	}
	if err := seedHabitStatus(-1, journalHabit.ID, yesterday, sharedtypes.HabitCompletionStatusCompleted, devIntPtr(9), devStringPtr("Quick reflection before bed.")); err != nil {
		return err
	}

	checkInNotes := []string{
		"Low-energy restart day with too much catch-up.",
		"Better morning structure but still some drag.",
		"Strong focus and decent recovery.",
		"Heavy day with visible fatigue.",
		"Recovered well after a lighter evening.",
		"Productive, but sleep debt still showed up.",
		"Stable finish with good control over context switching.",
	}
	moods := []int{2, 3, 4, 2, 4, 3, 4}
	energies := []int{2, 3, 4, 2, 4, 3, 4}
	sleepHours := []float64{5.9, 6.5, 7.6, 5.8, 8.0, 6.7, 7.3}
	sleepScores := []int{58, 68, 84, 55, 88, 71, 79}
	screenMinutes := []int{255, 225, 178, 290, 165, 214, 186}
	for i := 6; i >= 0; i-- {
		date := baseNow.AddDate(0, 0, -i).Format("2006-01-02")
		idx := 6 - i
		if err := seedCheckIn(-i, date, moods[idx], energies[idx], sleepHours[idx], sleepScores[idx], screenMinutes[idx], checkInNotes[idx]); err != nil {
			return err
		}
	}

	sessionPlan := []struct {
		dayOffset       int
		issueID         int64
		startHour       int
		startMinute     int
		durationMinutes int
		input           corecommands.SessionEndInput
	}{
		{-6, focusIssue.ID, 9, 15, 42, corecommands.SessionEndInput{CommitMessage: devStringPtr("refactor: bootstrap kernel repos"), WorkedOn: devStringPtr("repo and stream initialization"), Outcome: devStringPtr("baseline workspace created")}},
		{-5, inProgressIssue.ID, 9, 40, 72, corecommands.SessionEndInput{CommitMessage: devStringPtr("chore: verify packaging outputs"), WorkedOn: devStringPtr("artifact checks and hashes"), Outcome: devStringPtr("release checks mostly wired")}},
		{-4, readyIssue.ID, 10, 0, 38, corecommands.SessionEndInput{CommitMessage: devStringPtr("docs: shape rollout checklist"), WorkedOn: devStringPtr("deploy sequencing and rollback notes"), Outcome: devStringPtr("carry-over item clarified")}},
		{-3, underEstimateIssue.ID, 9, 30, 84, corecommands.SessionEndInput{CommitMessage: devStringPtr("feat: add command palette navigation"), WorkedOn: devStringPtr("palette actions and filtering"), Outcome: devStringPtr("underestimate case seeded")}},
		{-3, underEstimateIssue.ID, 14, 10, 27, corecommands.SessionEndInput{CommitMessage: devStringPtr("fix: polish command palette hints"), WorkedOn: devStringPtr("labels and keyboard copy"), Outcome: devStringPtr("issue should exceed estimate cleanly")}},
		{-2, readyIssue.ID, 13, 30, 31, corecommands.SessionEndInput{CommitMessage: devStringPtr("chore: finish rollout checklist"), WorkedOn: devStringPtr("go/no-go notes"), Outcome: devStringPtr("carried item completed")}},
		{-2, overEstimateIssue.ID, 16, 15, 24, corecommands.SessionEndInput{CommitMessage: devStringPtr("copy: revise lifecycle prompts"), WorkedOn: devStringPtr("short UX copy pass"), Outcome: devStringPtr("overestimate case remains under target")}},
		{-1, focusIssue.ID, 9, 45, 58, corecommands.SessionEndInput{CommitMessage: devStringPtr("feat: refresh rollup dashboard"), WorkedOn: devStringPtr("range summaries and details"), Outcome: devStringPtr("focus issue moved forward")}},
		{-1, focusIssue.ID, 15, 55, 22, corecommands.SessionEndInput{CommitMessage: devStringPtr("fix: small-screen rollup layout"), WorkedOn: devStringPtr("compact dashboard polish"), Outcome: devStringPtr("session mix for a strong recent day")}},
		{0, docsIssue.ID, 11, 0, 19, corecommands.SessionEndInput{CommitMessage: devStringPtr("docs: note linux install caveat"), WorkedOn: devStringPtr("shell completion docs"), Outcome: devStringPtr("future work has early progress")}},
	}
	for _, item := range sessionPlan {
		if err := seedSession(item.dayOffset, item.issueID, item.startHour, item.startMinute, item.durationMinutes, item.input); err != nil {
			return err
		}
	}
	if err := seedAt(-3, 18, 0, func() error {
		_, err := corecommands.ChangeIssueStatus(ctx, h.core, underEstimateIssue.ID, sharedtypes.IssueStatusDone, devStringPtr("Completed after running long relative to the estimate."))
		return err
	}); err != nil {
		return err
	}
	if err := seedAt(-2, 18, 15, func() error {
		_, err := corecommands.ChangeIssueStatus(ctx, h.core, overEstimateIssue.ID, sharedtypes.IssueStatusDone, devStringPtr("Wrapped quickly; estimate was padded."))
		return err
	}); err != nil {
		return err
	}

	scratchPath, err := corecommands.RegisterScratchpad(ctx, h.core, sharedtypes.ScratchPadMeta{
		Name: "Daily Notes",
		Path: "dev/[[date]]-notes.md",
	})
	if err != nil {
		return err
	}
	if _, err := scratchfile.Create(h.core.ScratchDir, scratchPath, "Daily Notes"); err != nil {
		return err
	}
	designPath, err := corecommands.RegisterScratchpad(ctx, h.core, sharedtypes.ScratchPadMeta{
		Name: "Launch Checklist",
		Path: "dev/launch-checklist.md",
	})
	if err != nil {
		return err
	}
	if _, err := scratchfile.Create(h.core.ScratchDir, designPath, "Launch Checklist"); err != nil {
		return err
	}

	if _, err := corecommands.SetContext(ctx, h.core, corecommands.ContextPatch{
		RepoSet:   true,
		RepoID:    &workRepo.ID,
		StreamSet: true,
		StreamID:  &appStream.ID,
		IssueSet:  true,
		IssueID:   &focusIssue.ID,
	}); err != nil {
		return err
	}

	return nil
}

func devStringPtr(value string) *string {
	return &value
}

func devIntPtr(value int) *int {
	return &value
}

func devFloatPtr(value float64) *float64 {
	return &value
}

func devDatePtr(value string) *string {
	return &value
}

func mostRecentMatchingWeekday(base time.Time, weekdays []int) (string, int) {
	for offset := 0; offset >= -6; offset-- {
		candidate := base.AddDate(0, 0, offset)
		weekday := int(candidate.Weekday())
		for _, day := range weekdays {
			if day == weekday {
				return candidate.Format("2006-01-02"), offset
			}
		}
	}
	return base.Format("2006-01-02"), 0
}
