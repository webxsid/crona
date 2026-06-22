package app

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/export"
	runtimepkg "crona/kernel/internal/runtime"
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
	return nil
}

func (h *Handler) seedDevData(ctx context.Context) error {
	if err := h.clearDevData(ctx); err != nil {
		return err
	}

	baseNow := time.Now().UTC()
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
		ts := time.Date(baseNow.Year(), baseNow.Month(), baseNow.Day(), hour, minute, 0, 0, time.UTC).
			AddDate(0, 0, dayOffset)
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
				Description: new(description),
				Color:       new(color),
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
				Description: new(description),
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
				notePtr = new(notes[0])
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
				Description:     new(description),
				EstimateMinutes: new(estimate),
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
					return fmt.Errorf(
						"invalid seeded issue transition for %q: %s -> %s",
						title,
						currentStatus,
						nextStatus,
					)
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
				Description:   new(description),
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
			_, err := corecommands.CompleteHabit(
				ctx,
				h.core,
				habitID,
				date,
				status,
				durationMinutes,
				notes,
			)
			return err
		})
	}
	seedCheckIn := func(dayOffset int, date string, mood, energy int, sleepHours float64, sleepScore, screenMinutes int, notes string) error {
		return seedAt(dayOffset, 7, 45, func() error {
			_, err := corecommands.UpsertDailyCheckIn(
				ctx,
				h.core,
				shareddto.DailyCheckInUpsertRequest{
					Date:              date,
					Mood:              mood,
					Energy:            energy,
					SleepHours:        new(sleepHours),
					SleepScore:        new(sleepScore),
					ScreenTimeMinutes: new(screenMinutes),
					Notes:             new(notes),
				},
			)
			return err
		})
	}
	seedSessionNotes := func(input corecommands.SessionEndInput) *string {
		lines := []string{}
		appendLine := func(label string, value *string) {
			if value == nil || strings.TrimSpace(*value) == "" {
				return
			}
			lines = append(lines, fmt.Sprintf("%s: %s", label, strings.TrimSpace(*value)))
		}
		appendLine("Worked on", input.WorkedOn)
		appendLine("Outcome", input.Outcome)
		appendLine("Next step", input.NextStep)
		appendLine("Blockers", input.Blockers)
		appendLine("Links", input.Links)
		if len(lines) == 0 {
			return nil
		}
		joined := strings.Join(lines, "\n")
		return &joined
	}
	seedSession := func(dayOffset int, issueID int64, startHour, startMinute, durationMinutes int, input corecommands.SessionEndInput) error {
		start := time.Date(baseNow.Year(), baseNow.Month(), baseNow.Day(), startHour, startMinute, 0, 0, time.UTC).
			AddDate(0, 0, dayOffset)
		end := start.Add(time.Duration(durationMinutes) * time.Minute)
		issue, err := h.core.Issues.GetByID(ctx, issueID, h.core.UserID)
		if err != nil {
			return err
		}
		if issue == nil {
			return fmt.Errorf("seed session issue not found: %d", issueID)
		}
		originalStatus := sharedtypes.NormalizeIssueStatus(issue.Status)
		restorationStatus := originalStatus
		needsRestore := !sharedtypes.CanStartFocus(originalStatus)
		if needsRestore && originalStatus != sharedtypes.IssueStatusPlanned {
			if _, err := corecommands.ChangeIssueStatus(ctx, h.core, issueID, sharedtypes.IssueStatusPlanned, nil); err != nil {
				return err
			}
		}
		return withSeedNow(end, func() error {
			_, err := corecommands.LogManualSession(ctx, h.core, corecommands.ManualSessionInput{
				IssueID:             issueID,
				Date:                start.Format("2006-01-02"),
				StartTime:           new(fmt.Sprintf("%02d:%02d", startHour, startMinute)),
				WorkDurationSeconds: durationMinutes * 60,
				CommitMessage:       input.CommitMessage,
				Notes:               seedSessionNotes(input),
			})
			if err != nil {
				return err
			}
			if !needsRestore {
				return nil
			}
			switch restorationStatus {
			case sharedtypes.IssueStatusBacklog:
				if _, err := corecommands.ChangeIssueStatus(ctx, h.core, issueID, sharedtypes.IssueStatusPlanned, nil); err != nil {
					return err
				}
				if _, err := corecommands.ChangeIssueStatus(ctx, h.core, issueID, sharedtypes.IssueStatusBacklog, nil); err != nil {
					return err
				}
			case sharedtypes.IssueStatusPlanned,
				sharedtypes.IssueStatusReady,
				sharedtypes.IssueStatusInProgress,
				sharedtypes.IssueStatusBlocked,
				sharedtypes.IssueStatusInReview,
				sharedtypes.IssueStatusDone,
				sharedtypes.IssueStatusAbandoned:
				if _, err := corecommands.ChangeIssueStatus(ctx, h.core, issueID, restorationStatus, nil); err != nil {
					return err
				}
			}
			return nil
		})
	}

	workRepo, err := createRepo(
		"Work",
		"Team delivery, product design, and release operations.",
		"blue",
	)
	if err != nil {
		return err
	}
	personalRepo, err := createRepo(
		"Personal",
		"Life admin, routines, and home maintenance.",
		"green",
	)
	if err != nil {
		return err
	}
	ossRepo, err := createRepo(
		"OSS",
		"Open-source tooling and documentation experiments.",
		"magenta",
	)
	if err != nil {
		return err
	}

	appStream, err := createStream(workRepo.ID, "app", "Primary product and UX work.")
	if err != nil {
		return err
	}
	infraStream, err := createStream(
		workRepo.ID,
		"infra",
		"Release, packaging, and deployment plumbing.",
	)
	if err != nil {
		return err
	}
	platformStream, err := createStream(
		workRepo.ID,
		"platform",
		"Shared APIs, reliability, and observability.",
	)
	if err != nil {
		return err
	}
	homeStream, err := createStream(personalRepo.ID, "home", "Household admin and chores.")
	if err != nil {
		return err
	}
	wellbeingStream, err := createStream(
		personalRepo.ID,
		"wellbeing",
		"Health, fitness, and routines.",
	)
	if err != nil {
		return err
	}
	cliStream, err := createStream(ossRepo.ID, "cli", "CLI polish and contributor onboarding.")
	if err != nil {
		return err
	}

	focusIssue, err := createIssue(
		-6,
		appStream.ID,
		"Port dev tooling to Go",
		"Validate the IPC-first workflow and replace the remaining shell scripts.",
		90,
		new(dateAt(-1)),
		[]sharedtypes.IssueStatus{sharedtypes.IssueStatusReady, sharedtypes.IssueStatusInProgress},
		"Drive the migration end-to-end and keep dev experience tight.",
	)
	if err != nil {
		return err
	}
	underEstimateIssue, err := createIssue(
		-5,
		appStream.ID,
		"Add keyboard-first command palette",
		"Speed up common view and dialog actions from one launcher.",
		35,
		new(dateAt(-3)),
		[]sharedtypes.IssueStatus{sharedtypes.IssueStatusReady, sharedtypes.IssueStatusInProgress},
		"Intentionally underestimated for seed accuracy drift.",
	)
	if err != nil {
		return err
	}
	overEstimateIssue, err := createIssue(
		-4,
		appStream.ID,
		"Review lifecycle UX copy",
		"Tighten action labels and empty-state wording across the dashboard.",
		95,
		new(dateAt(-2)),
		[]sharedtypes.IssueStatus{
			sharedtypes.IssueStatusInProgress,
			sharedtypes.IssueStatusInReview,
		},
		"Intentionally overestimated for rollup estimate-bias testing.",
	)
	if err != nil {
		return err
	}
	readyIssue, err := createIssue(
		-6,
		infraStream.ID,
		"Prepare rollout checklist",
		"Capture deploy sequencing, smoke tests, and rollback checkpoints.",
		45,
		new(dateAt(-4)),
		[]sharedtypes.IssueStatus{sharedtypes.IssueStatusReady},
		"Waiting only on final go/no-go review.",
	)
	if err != nil {
		return err
	}
	inProgressIssue, err := createIssue(
		-6,
		infraStream.ID,
		"Wire release packaging checks",
		"Verify built artifacts, checksums, and update install docs.",
		75,
		new(dateAt(-5)),
		[]sharedtypes.IssueStatus{sharedtypes.IssueStatusInProgress},
		"Partially implemented; needs artifact verification.",
	)
	if err != nil {
		return err
	}
	blockedIssue, err := createIssue(
		-6,
		infraStream.ID,
		"Provision CI signing secrets",
		"Unblock package signing in CI and production release automation.",
		30,
		new(dateAt(-4)),
		[]sharedtypes.IssueStatus{sharedtypes.IssueStatusBlocked},
		"Awaiting access to the signing account.",
	)
	if err != nil {
		return err
	}
	if _, err := createIssue(-4, infraStream.ID, "Review release support copy", "Check updater and support wording before the tester build.", 20, new(dateAt(-1)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusInProgress, sharedtypes.IssueStatusInReview}, "Held in review so dev data covers the full lifecycle."); err != nil {
		return err
	}
	if _, err := createIssue(-6, platformStream.ID, "Backfill metrics range tests", "Cover streaks, burnout rollups, and historical summaries.", 55, new(dateAt(-2)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusInProgress, sharedtypes.IssueStatusDone}); err != nil {
		return err
	}
	if _, err := createIssue(-6, platformStream.ID, "Reduce websocket reconnect churn", "Stabilize event fanout under network flaps.", 50, new(dateAt(-1)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusAbandoned}, "Superseded by the unix socket migration."); err != nil {
		return err
	}
	docsIssue, err := createIssue(
		-3,
		cliStream.ID,
		"Improve install docs for Linux",
		"Add shell completion, troubleshooting, and verification notes.",
		35,
		new(dateAt(1)),
		[]sharedtypes.IssueStatus{sharedtypes.IssueStatusReady},
		"Future planned work to keep the window from being all closed.",
	)
	if err != nil {
		return err
	}
	if _, err := createIssue(-3, cliStream.ID, "Publish example config bundle", "Ship realistic examples for repo, stream, and timer setup.", 25, new(dateAt(0)), nil); err != nil {
		return err
	}
	if _, err := createIssue(-2, homeStream.ID, "Plan weekend errands", "Group shopping, laundry, and pickup tasks into one route.", 30, new(dateAt(0)), nil); err != nil {
		return err
	}
	if _, err := createIssue(-6, homeStream.ID, "Research standing desk options", "Compare dimensions and price points for the office nook.", 25, new(dateAt(2)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusAbandoned}, "Deferred until the room reorganization is done."); err != nil {
		return err
	}
	if _, err := createIssue(-6, platformStream.ID, "Audit event replay gaps", "Investigate dropped events and stale refresh paths.", 50, new(dateAt(-6)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusReady}, "Left incomplete on purpose so the oldest day shows planned-only work."); err != nil {
		return err
	}
	if _, err := createIssue(-5, platformStream.ID, "Reconcile update relaunch UX", "Smooth out install handoff and watchdog reporting.", 40, new(dateAt(-5)), []sharedtypes.IssueStatus{sharedtypes.IssueStatusReady}, "Left untouched so one day can register as missed."); err != nil {
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

	workoutHabit, err := createHabit(
		-29,
		wellbeingStream.ID,
		"Strength Training",
		"Three lifting sessions each week with basic progression.",
		"weekly",
		[]int{1, 3, 5},
		new(60),
	)
	if err != nil {
		return err
	}
	walkHabit, err := createHabit(
		-29,
		wellbeingStream.ID,
		"Morning Walk",
		"Get outside before work for a short walk.",
		"daily",
		nil,
		new(25),
	)
	if err != nil {
		return err
	}
	journalHabit, err := createHabit(
		-29,
		wellbeingStream.ID,
		"Journal",
		"Capture a quick end-of-day reflection.",
		"daily",
		nil,
		new(10),
	)
	if err != nil {
		return err
	}
	inboxHabit, err := createHabit(
		-29,
		appStream.ID,
		"Inbox Zero Sweep",
		"Clear triage inboxes before the afternoon block.",
		"weekdays",
		nil,
		new(15),
	)
	if err != nil {
		return err
	}
	if _, err := createHabit(-29, homeStream.ID, "Laundry Reset", "Run and fold one load every Sunday.", "weekly", []int{0}, new(45)); err != nil {
		return err
	}
	if _, err := createHabit(-29, cliStream.ID, "Read Release Notes", "Scan upstream release notes for dependency changes.", "weekly", []int{2, 4}, new(20)); err != nil {
		return err
	}

	habitStreakDefs := []sharedtypes.HabitStreakDefinition{
		{
			ID:            "daily-reflection",
			Name:          "Daily Reflection",
			Enabled:       true,
			Period:        sharedtypes.HabitStreakPeriodDay,
			RequiredCount: 1,
			HabitIDs:      []int64{journalHabit.ID},
		},
		{
			ID:            "daily-habits-any",
			Name:          "Daily Habits Any",
			Enabled:       true,
			Period:        sharedtypes.HabitStreakPeriodDay,
			MatchMode:     sharedtypes.MomentumMatchModeAny,
			RequiredCount: 1,
			HabitIDs:      []int64{walkHabit.ID, journalHabit.ID, inboxHabit.ID},
		},
		{
			ID:            "daily-context-any",
			Name:          "Daily Context Any",
			Enabled:       true,
			TargetKind:    sharedtypes.MomentumTargetKindContext,
			Period:        sharedtypes.HabitStreakPeriodDay,
			MatchMode:     sharedtypes.MomentumMatchModeAny,
			RequiredCount: 3600,
			Contexts: []sharedtypes.MomentumContext{
				{RepoID: workRepo.ID, StreamID: &appStream.ID},
				{RepoID: workRepo.ID, StreamID: &infraStream.ID},
			},
		},
		{
			ID:            "daily-context-all",
			Name:          "Daily Context All",
			Enabled:       true,
			TargetKind:    sharedtypes.MomentumTargetKindContext,
			Period:        sharedtypes.HabitStreakPeriodDay,
			MatchMode:     sharedtypes.MomentumMatchModeAll,
			RequiredCount: 3600,
			Contexts: []sharedtypes.MomentumContext{
				{RepoID: workRepo.ID, StreamID: &appStream.ID},
				{RepoID: workRepo.ID, StreamID: &infraStream.ID},
			},
		},
		{
			ID:            "recovery-any",
			Name:          "Recovery Any",
			Enabled:       true,
			Period:        sharedtypes.HabitStreakPeriodWeek,
			MatchMode:     sharedtypes.MomentumMatchModeAny,
			RequiredCount: 2,
			HabitIDs:      []int64{walkHabit.ID, journalHabit.ID},
		},
		{
			ID:            "wellbeing-all",
			Name:          "Wellbeing All",
			Enabled:       true,
			Period:        sharedtypes.HabitStreakPeriodMonth,
			MatchMode:     sharedtypes.MomentumMatchModeAll,
			RequiredCount: 1,
			HabitIDs:      []int64{walkHabit.ID, journalHabit.ID, workoutHabit.ID},
		},
		{
			ID:            "work-context-any",
			Name:          "Work Context Any",
			Enabled:       true,
			TargetKind:    sharedtypes.MomentumTargetKindContext,
			Period:        sharedtypes.HabitStreakPeriodWeek,
			MatchMode:     sharedtypes.MomentumMatchModeAny,
			RequiredCount: 7200,
			Contexts: []sharedtypes.MomentumContext{
				{RepoID: workRepo.ID},
				{RepoID: ossRepo.ID, StreamID: &cliStream.ID},
			},
		},
		{
			ID:            "delivery-context-all",
			Name:          "Delivery Context All",
			Enabled:       true,
			TargetKind:    sharedtypes.MomentumTargetKindContext,
			Period:        sharedtypes.HabitStreakPeriodWeek,
			MatchMode:     sharedtypes.MomentumMatchModeAll,
			RequiredCount: 3600,
			Contexts: []sharedtypes.MomentumContext{
				{RepoID: workRepo.ID, StreamID: &appStream.ID},
				{RepoID: workRepo.ID, StreamID: &infraStream.ID},
			},
		},
	}
	if err := h.core.HabitStreakDefinitions.ReplaceAll(
		ctx,
		h.core.UserID,
		h.core.Now(),
		habitStreakDefs,
	); err != nil {
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
	for i := 29; i >= 0; i-- {
		date := baseNow.AddDate(0, 0, -i).Format("2006-01-02")
		idx := (29 - i) % len(checkInNotes)
		mood := 2 + ((29 - i + 1) % 4)
		energy := 2 + ((29 - i + 2) % 4)
		sleep := 5.8 + float64((29-i)%8)*0.3
		sleepScore := 58 + ((29 - i) % 7 * 6)
		screenTime := 165 + ((29 - i) % 6 * 24)
		if err := seedCheckIn(-i, date, mood, energy, sleep, sleepScore, screenTime, checkInNotes[idx]); err != nil {
			return err
		}
		weekday := int(baseNow.AddDate(0, 0, -i).Weekday())
		if i == 0 {
			if err := seedHabitStatus(-i, walkHabit.ID, date, sharedtypes.HabitCompletionStatusFailed, nil, new("Weather broke the morning routine.")); err != nil {
				return err
			}
		} else if i%9 != 0 {
			if err := seedHabitStatus(-i, walkHabit.ID, date, sharedtypes.HabitCompletionStatusCompleted, new(22+(i%6)), new("Morning walk logged for the 30-day seed window.")); err != nil {
				return err
			}
		}
		if i%11 != 0 {
			if err := seedHabitStatus(-i, journalHabit.ID, date, sharedtypes.HabitCompletionStatusCompleted, new(8+(i%5)), new("Short reflection for streak testing.")); err != nil {
				return err
			}
		}
		if weekday == 1 || weekday == 3 || weekday == 5 {
			if i%13 != 0 {
				if err := seedHabitStatus(-i, workoutHabit.ID, date, sharedtypes.HabitCompletionStatusCompleted, new(48+(i%5)*4), new("Strength session for weekly streak cadence.")); err != nil {
					return err
				}
			}
		}
		if weekday >= 1 && weekday <= 5 && i%10 != 0 {
			if err := seedHabitStatus(-i, inboxHabit.ID, date, sharedtypes.HabitCompletionStatusCompleted, new(10+(i%4)), new("Inbox and notifications cleared.")); err != nil {
				return err
			}
		}
		if weekday >= 1 && weekday <= 5 && i%17 == 0 {
			if err := seedHabitStatus(-i, inboxHabit.ID, date, sharedtypes.HabitCompletionStatusFailed, nil, new("Inbox sweep missed during a heavy day.")); err != nil {
				return err
			}
		}
	}
	olderSessionPatterns := []struct {
		issueID         int64
		startHour       int
		startMinute     int
		durationMinutes int
		input           corecommands.SessionEndInput
	}{
		{
			focusIssue.ID,
			9,
			15,
			42,
			corecommands.SessionEndInput{
				CommitMessage: new("refactor: bootstrap kernel repos"),
				WorkedOn:      new("repo and stream initialization"),
				Outcome:       new("baseline workspace created"),
			},
		},
		{
			inProgressIssue.ID,
			9,
			40,
			72,
			corecommands.SessionEndInput{
				CommitMessage: new("chore: verify packaging outputs"),
				WorkedOn:      new("artifact checks and hashes"),
				Outcome:       new("release checks mostly wired"),
			},
		},
		{
			readyIssue.ID,
			10,
			0,
			38,
			corecommands.SessionEndInput{
				CommitMessage: new("docs: shape rollout checklist"),
				WorkedOn:      new("deploy sequencing and rollback notes"),
				Outcome:       new("carry-over item clarified"),
			},
		},
		{
			underEstimateIssue.ID,
			9,
			30,
			84,
			corecommands.SessionEndInput{
				CommitMessage: new("feat: add command palette navigation"),
				WorkedOn:      new("palette actions and filtering"),
				Outcome:       new("underestimate case seeded"),
			},
		},
		{
			overEstimateIssue.ID,
			14,
			10,
			27,
			corecommands.SessionEndInput{
				CommitMessage: new("fix: polish command palette hints"),
				WorkedOn:      new("labels and keyboard copy"),
				Outcome:       new("issue should exceed estimate cleanly"),
			},
		},
		{
			focusIssue.ID,
			13,
			30,
			31,
			corecommands.SessionEndInput{
				CommitMessage: new("chore: finish rollout checklist"),
				WorkedOn:      new("go/no-go notes"),
				Outcome:       new("carried item completed"),
			},
		},
	}
	for dayOffset := -29; dayOffset <= -7; dayOffset++ {
		pattern := olderSessionPatterns[(dayOffset+29)%len(olderSessionPatterns)]
		if err := seedSession(dayOffset, pattern.issueID, pattern.startHour, pattern.startMinute, pattern.durationMinutes, pattern.input); err != nil {
			return err
		}
	}
	workoutDate, workoutOffset := mostRecentMatchingWeekday(baseNow, []int{1, 3, 5})
	if err := seedHabitStatus(workoutOffset, workoutHabit.ID, workoutDate, sharedtypes.HabitCompletionStatusCompleted, new(58), new("Main lifts plus accessory work.")); err != nil {
		return err
	}
	inboxDate, inboxOffset := mostRecentMatchingWeekday(baseNow, []int{1, 2, 3, 4, 5})
	if err := seedHabitStatus(inboxOffset, inboxHabit.ID, inboxDate, sharedtypes.HabitCompletionStatusCompleted, new(12), new("Inbox and notifications cleared before lunch.")); err != nil {
		return err
	}

	sessionPlan := []struct {
		dayOffset       int
		issueID         int64
		startHour       int
		startMinute     int
		durationMinutes int
		input           corecommands.SessionEndInput
	}{
		{
			-6,
			focusIssue.ID,
			9,
			15,
			42,
			corecommands.SessionEndInput{
				CommitMessage: new("refactor: bootstrap kernel repos"),
				WorkedOn:      new("repo and stream initialization"),
				Outcome:       new("baseline workspace created"),
			},
		},
		{
			-5,
			inProgressIssue.ID,
			9,
			40,
			72,
			corecommands.SessionEndInput{
				CommitMessage: new("chore: verify packaging outputs"),
				WorkedOn:      new("artifact checks and hashes"),
				Outcome:       new("release checks mostly wired"),
			},
		},
		{
			-4,
			readyIssue.ID,
			10,
			0,
			38,
			corecommands.SessionEndInput{
				CommitMessage: new("docs: shape rollout checklist"),
				WorkedOn:      new("deploy sequencing and rollback notes"),
				Outcome:       new("carry-over item clarified"),
			},
		},
		{
			-3,
			underEstimateIssue.ID,
			9,
			30,
			84,
			corecommands.SessionEndInput{
				CommitMessage: new("feat: add command palette navigation"),
				WorkedOn:      new("palette actions and filtering"),
				Outcome:       new("underestimate case seeded"),
			},
		},
		{
			-3,
			underEstimateIssue.ID,
			14,
			10,
			27,
			corecommands.SessionEndInput{
				CommitMessage: new("fix: polish command palette hints"),
				WorkedOn:      new("labels and keyboard copy"),
				Outcome:       new("issue should exceed estimate cleanly"),
			},
		},
		{
			-2,
			readyIssue.ID,
			13,
			30,
			31,
			corecommands.SessionEndInput{
				CommitMessage: new("chore: finish rollout checklist"),
				WorkedOn:      new("go/no-go notes"),
				Outcome:       new("carried item completed"),
			},
		},
		{
			-2,
			overEstimateIssue.ID,
			16,
			15,
			24,
			corecommands.SessionEndInput{
				CommitMessage: new("copy: revise lifecycle prompts"),
				WorkedOn:      new("short UX copy pass"),
				Outcome:       new("overestimate case remains under target"),
			},
		},
		{
			-1,
			focusIssue.ID,
			9,
			45,
			58,
			corecommands.SessionEndInput{
				CommitMessage: new("feat: refresh rollup dashboard"),
				WorkedOn:      new("range summaries and details"),
				Outcome:       new("focus issue moved forward"),
			},
		},
		{
			-1,
			focusIssue.ID,
			15,
			55,
			22,
			corecommands.SessionEndInput{
				CommitMessage: new("fix: small-screen rollup layout"),
				WorkedOn:      new("compact dashboard polish"),
				Outcome:       new("session mix for a strong recent day"),
			},
		},
		{
			0,
			docsIssue.ID,
			11,
			0,
			19,
			corecommands.SessionEndInput{
				CommitMessage: new("docs: note linux install caveat"),
				WorkedOn:      new("shell completion docs"),
				Outcome:       new("future work has early progress"),
			},
		},
	}
	for _, item := range sessionPlan {
		if err := seedSession(item.dayOffset, item.issueID, item.startHour, item.startMinute, item.durationMinutes, item.input); err != nil {
			return err
		}
	}
	if err := seedAt(-3, 18, 0, func() error {
		_, err := corecommands.ChangeIssueStatus(ctx, h.core, underEstimateIssue.ID, sharedtypes.IssueStatusDone, new("Completed after running long relative to the estimate."))
		return err
	}); err != nil {
		return err
	}
	if err := seedAt(-2, 18, 15, func() error {
		_, err := corecommands.ChangeIssueStatus(ctx, h.core, overEstimateIssue.ID, sharedtypes.IssueStatusDone, new("Wrapped quickly; estimate was padded."))
		return err
	}); err != nil {
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

func mostRecentMatchingWeekday(base time.Time, weekdays []int) (string, int) {
	for offset := 0; offset >= -6; offset-- {
		candidate := base.AddDate(0, 0, offset)
		weekday := int(candidate.Weekday())

		if slices.Contains(weekdays, weekday) {
			return candidate.Format("2006-01-02"), offset
		}

	}
	return base.Format("2006-01-02"), 0
}
