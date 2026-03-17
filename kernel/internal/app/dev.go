package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corecommands "crona/kernel/internal/core/commands"
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
	if err := scratchfile.ClearAll(h.core.ScratchDir); err != nil {
		return fmt.Errorf("clear scratch files: %w", err)
	}
	if err := h.core.InitDefaults(ctx); err != nil {
		return fmt.Errorf("reinitialize defaults: %w", err)
	}

	payload, _ := json.Marshal(sharedtypes.TimerState{State: "idle"})
	h.core.Events.Emit(sharedtypes.KernelEvent{
		Type:    sharedtypes.EventTypeTimerState,
		Payload: payload,
	})
	return nil
}

func (h *Handler) seedDevData(ctx context.Context) error {
	if err := h.clearDevData(ctx); err != nil {
		return err
	}

	baseNow := time.Now().UTC()
	today := baseNow.Format("2006-01-02")
	tomorrow := baseNow.Add(24 * time.Hour).Format("2006-01-02")
	yesterday := baseNow.Add(-24 * time.Hour).Format("2006-01-02")

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

	focusIssue, err := createIssue(-2, appStream.ID, "Port dev tooling to Go", "Validate the IPC-first workflow and replace the remaining shell scripts.", 90, &today, []sharedtypes.IssueStatus{sharedtypes.IssueStatusReady, sharedtypes.IssueStatusInProgress}, "Drive the migration end-to-end and keep dev experience tight.")
	if err != nil {
		return err
	}
	if _, err := createIssue(-1, appStream.ID, "Add keyboard-first command palette", "Speed up common view and dialog actions from one launcher.", 60, &tomorrow, nil); err != nil {
		return err
	}
	if _, err := createIssue(-4, appStream.ID, "Review lifecycle UX copy", "Tighten action labels and empty-state wording across the dashboard.", 40, &today, []sharedtypes.IssueStatus{sharedtypes.IssueStatusInProgress, sharedtypes.IssueStatusInReview}, "Ready for wording and interaction review."); err != nil {
		return err
	}
	readyIssue, err := createIssue(-3, infraStream.ID, "Prepare rollout checklist", "Capture deploy sequencing, smoke tests, and rollback checkpoints.", 45, &today, []sharedtypes.IssueStatus{sharedtypes.IssueStatusReady}, "Waiting only on final go/no-go review.")
	if err != nil {
		return err
	}
	inProgressIssue, err := createIssue(-5, infraStream.ID, "Wire release packaging checks", "Verify built artifacts, checksums, and update install docs.", 75, &yesterday, []sharedtypes.IssueStatus{sharedtypes.IssueStatusInProgress}, "Partially implemented; needs artifact verification.")
	if err != nil {
		return err
	}
	if _, err := createIssue(-5, infraStream.ID, "Provision CI signing secrets", "Unblock package signing in CI and production release automation.", 30, &yesterday, []sharedtypes.IssueStatus{sharedtypes.IssueStatusBlocked}, "Awaiting access to the signing account."); err != nil {
		return err
	}
	if _, err := createIssue(-6, platformStream.ID, "Backfill metrics range tests", "Cover streaks, burnout rollups, and historical summaries.", 55, devDatePtr(baseNow.AddDate(0, 0, -2).Format("2006-01-02")), []sharedtypes.IssueStatus{sharedtypes.IssueStatusInProgress, sharedtypes.IssueStatusDone}); err != nil {
		return err
	}
	if _, err := createIssue(-6, platformStream.ID, "Reduce websocket reconnect churn", "Stabilize event fanout under network flaps.", 50, devDatePtr(baseNow.AddDate(0, 0, -1).Format("2006-01-02")), []sharedtypes.IssueStatus{sharedtypes.IssueStatusAbandoned}, "Superseded by the unix socket migration."); err != nil {
		return err
	}
	if _, err := createIssue(-1, cliStream.ID, "Improve install docs for Linux", "Add shell completion, troubleshooting, and verification notes.", 35, &tomorrow, nil); err != nil {
		return err
	}
	if _, err := createIssue(-3, cliStream.ID, "Publish example config bundle", "Ship realistic examples for repo, stream, and timer setup.", 25, &today, nil); err != nil {
		return err
	}
	if _, err := createIssue(-2, homeStream.ID, "Plan weekend errands", "Group shopping, laundry, and pickup tasks into one route.", 30, &today, nil); err != nil {
		return err
	}
	if _, err := createIssue(-6, homeStream.ID, "Research standing desk options", "Compare dimensions and price points for the office nook.", 25, devDatePtr(baseNow.AddDate(0, 0, 2).Format("2006-01-02")), []sharedtypes.IssueStatus{sharedtypes.IssueStatusAbandoned}, "Deferred until the room reorganization is done."); err != nil {
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
	if err := seedHabitStatus(0, inboxHabit.ID, today, sharedtypes.HabitCompletionStatusCompleted, devIntPtr(12), devStringPtr("Inbox and notifications cleared before lunch.")); err != nil {
		return err
	}
	if err := seedHabitStatus(-1, journalHabit.ID, yesterday, sharedtypes.HabitCompletionStatusCompleted, devIntPtr(9), devStringPtr("Quick reflection before bed.")); err != nil {
		return err
	}

	checkInNotes := []string{
		"Settled back into the routine after the weekend.",
		"Energy improved once the morning block was protected.",
		"Solid focus day with fewer interruptions.",
		"Longer work block, but recovery still felt okay.",
		"Good pace and decent sleep coming into the day.",
		"Sharper than yesterday and less context switching.",
		"Ended the week feeling steady and in control.",
	}
	moods := []int{3, 4, 4, 3, 4, 5, 4}
	energies := []int{3, 3, 4, 3, 4, 4, 4}
	sleepHours := []float64{6.8, 7.1, 7.4, 6.9, 7.6, 7.2, 7.0}
	sleepScores := []int{72, 76, 81, 74, 84, 79, 77}
	screenMinutes := []int{210, 195, 188, 225, 176, 182, 190}
	for i := 6; i >= 0; i-- {
		date := baseNow.AddDate(0, 0, -i).Format("2006-01-02")
		idx := 6 - i
		if err := seedCheckIn(-i, date, moods[idx], energies[idx], sleepHours[idx], sleepScores[idx], screenMinutes[idx], checkInNotes[idx]); err != nil {
			return err
		}
	}

	sessionInputs := []corecommands.SessionEndInput{
		{CommitMessage: devStringPtr("refactor: move repo bootstrap into kernel"), WorkedOn: devStringPtr("dev seed and repo bootstrap"), Outcome: devStringPtr("multi-repo data shape in place")},
		{CommitMessage: devStringPtr("feat: add dashboard streak summaries"), WorkedOn: devStringPtr("metrics rollup and streak calculations"), Outcome: devStringPtr("7-day trend cards rendering cleanly")},
		{CommitMessage: devStringPtr("fix: stabilize session detail overlay"), WorkedOn: devStringPtr("overlay sizing and note parsing"), Outcome: devStringPtr("session detail works at smaller widths")},
		{CommitMessage: devStringPtr("feat: support habit failure status"), WorkedOn: devStringPtr("habit domain, TUI actions, and summary bars"), Outcome: devStringPtr("failed habits are visible and actionable")},
		{CommitMessage: devStringPtr("chore: tighten release packaging checks"), WorkedOn: devStringPtr("artifact verification and install script updates"), Outcome: devStringPtr("release checklist issue moved forward")},
		{CommitMessage: devStringPtr("docs: refresh install examples"), WorkedOn: devStringPtr("CLI docs and contributor examples"), Outcome: devStringPtr("examples reflect current transport and settings")},
		{CommitMessage: devStringPtr("feat: expand dev seed workspace"), WorkedOn: devStringPtr("repos, streams, sessions, and check-ins"), Outcome: devStringPtr("dev mode feels realistic for demos")},
	}
	sessionIssues := []int64{focusIssue.ID, readyIssue.ID, inProgressIssue.ID, focusIssue.ID, inProgressIssue.ID, readyIssue.ID, focusIssue.ID}
	sessionDurations := []int{52, 46, 64, 58, 49, 41, 55}
	for i := 6; i >= 0; i-- {
		idx := 6 - i
		if err := seedSession(-i, sessionIssues[idx], 9, 30, sessionDurations[idx], sessionInputs[idx]); err != nil {
			return err
		}
	}
	if err := seedSession(-2, readyIssue.ID, 14, 0, 33, corecommands.SessionEndInput{
		CommitMessage: devStringPtr("chore: polish rollout notes"),
		WorkedOn:      devStringPtr("release checklist edits"),
		Outcome:       devStringPtr("go/no-go doc tightened"),
	}); err != nil {
		return err
	}
	if err := seedSession(-1, focusIssue.ID, 16, 0, 27, corecommands.SessionEndInput{
		CommitMessage: devStringPtr("refactor: simplify daily summary wiring"),
		WorkedOn:      devStringPtr("daily dashboard rendering"),
		Outcome:       devStringPtr("summary view less noisy"),
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
