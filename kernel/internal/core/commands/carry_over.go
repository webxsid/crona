package commands

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"crona/kernel/internal/core"
	sharedtypes "crona/shared/types"

	"github.com/google/uuid"
)

const carryOverSource = "carry_over"

// CarryOverPastDays carries every planned, unfinished issue from a day before today that got at
// least the configured amount of work that day. It is idempotent: carried entries are no longer
// planned, so a second run finds nothing.
func CarryOverPastDays(ctx context.Context, c *core.Context) error {
	settings, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil {
		return err
	}
	if settings == nil || !settings.CarryOverEnabled {
		return nil
	}
	if err := finalizeExpiredDailyPlanFailures(ctx, c, c.Now()); err != nil {
		return err
	}
	today := c.CurrentDate()
	active, err := c.DailyPlans.ListActiveEntries(ctx, c.UserID)
	if err != nil {
		return err
	}
	sort.SliceStable(active, func(i, j int) bool { return active[i].Date < active[j].Date })
	worked := map[string]map[int64]int{}
	for _, entry := range active {
		if entry.Date >= today || entry.PendingFailureAt != nil {
			continue
		}
		if _, ok := worked[entry.Date]; !ok {
			byIssue, err := workedSecondsByIssueOnDate(ctx, c, entry.Date)
			if err != nil {
				return err
			}
			worked[entry.Date] = byIssue
		}
		seconds := worked[entry.Date][entry.IssueID]
		if !meetsCarryThreshold(seconds, settings.CarryOverMinMinutes) {
			continue
		}
		issue, err := c.Issues.GetByID(ctx, entry.IssueID, c.UserID)
		if err != nil {
			return err
		}
		if issue == nil || !sharedtypes.CanCarryOver(issue.Status) {
			continue
		}
		target := nextCarryDate(entry.Date, today, settings)
		if _, err := carryEntry(ctx, c, issue, entry, target, seconds, false); err != nil {
			return err
		}
	}
	return nil
}

// CarryIssue carries an issue from its current plan date to the next working day (or targetDate
// when given), whatever it was worked. Choosing to carry is the signal, so neither the threshold
// nor the automatic status rule applies: any open issue can be carried by hand.
func CarryIssue(
	ctx context.Context,
	c *core.Context,
	issueID int64,
	targetDate string,
) (*sharedtypes.Issue, error) {
	issue, err := c.Issues.GetByID(ctx, issueID, c.UserID)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, errors.New("issue not found")
	}
	if issue.Status == sharedtypes.IssueStatusDone || issue.Status == sharedtypes.IssueStatusAbandoned {
		return nil, errors.New("only open issues can be carried over")
	}
	fromDate := issueCommittedDate(issue)
	if fromDate == "" {
		return nil, errors.New("issue is not planned for a day")
	}
	if err := finalizeExpiredDailyPlanFailures(ctx, c, c.Now()); err != nil {
		return nil, err
	}
	entry, err := c.DailyPlans.GetEntry(ctx, c.UserID, fromDate, issueID)
	if err != nil {
		return nil, err
	}
	if entry == nil || entry.Status != sharedtypes.DailyPlanEntryStatusPlanned {
		return nil, errors.New("issue has no open plan entry to carry")
	}
	settings, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	target := targetDate
	if target == "" {
		target = nextCarryDate(fromDate, c.CurrentDate(), settings)
	}
	if err := ensureDailyPlanDate(target); err != nil {
		return nil, err
	}
	if target <= fromDate {
		return nil, errors.New("carry-over target must be after the planned date")
	}
	byIssue, err := workedSecondsByIssueOnDate(ctx, c, fromDate)
	if err != nil {
		return nil, err
	}
	return carryEntry(ctx, c, issue, *entry, target, byIssue[issueID], true)
}

// UndoCarry reverses the issue's latest carry-over while it is inside the rollback window:
// the carried entry is planned again, the entry the carry created is removed, and the issue
// moves back to the earlier date.
func UndoCarry(ctx context.Context, c *core.Context, issueID int64) (*sharedtypes.Issue, error) {
	issue, err := c.Issues.GetByID(ctx, issueID, c.UserID)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, errors.New("issue not found")
	}
	current, err := c.DailyPlans.GetEntry(ctx, c.UserID, issueCommittedDate(issue), issueID)
	if err != nil {
		return nil, err
	}
	if current == nil || current.CarriedFromEntryID == nil {
		return nil, errors.New("issue has no carry-over to undo")
	}
	previous, err := c.DailyPlans.GetEntryByID(ctx, c.UserID, *current.CarriedFromEntryID)
	if err != nil {
		return nil, err
	}
	if previous == nil || previous.Status != sharedtypes.DailyPlanEntryStatusCarried {
		return nil, errors.New("issue has no carry-over to undo")
	}
	settings, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	now := c.Now()
	if !withinRollbackWindow(previous.ResolvedAt, now, settings) {
		return nil, errors.New("the carry-over can no longer be undone")
	}
	if err := c.DailyPlans.ReopenEntry(ctx, previous.ID, now); err != nil {
		return nil, err
	}
	if err := c.DailyPlans.SetCarryLinks(ctx, previous.ID, now, previous.CarriedFromEntryID, nil, previous.CarryCount); err != nil {
		return nil, err
	}
	if err := c.DailyPlans.DeleteEntry(ctx, current.ID); err != nil {
		return nil, err
	}
	updated, err := setIssueTodoDate(ctx, c, issueID, previous.Date, now)
	if err != nil {
		return nil, err
	}
	if err := c.DailyPlans.AppendEvent(ctx, previous.ID, c.UserID, c.DeviceID, now, sharedtypes.DailyPlanEventCarryUndone, nil, map[string]any{
		"issueId": issueID,
		"date":    previous.Date,
		"toDate":  current.Date,
	}); err != nil {
		return nil, err
	}
	return updated, recordCarryChange(ctx, c, updated, now, map[string]any{
		"todoForDate": previous.Date,
		"carryUndone": current.Date,
	})
}

// rescheduleAsCarry turns a reschedule into carry-over work when that is what it means: moving
// an issue worked on today to a later day carries it, and moving a carried-in issue back to the
// day it came from, inside the rollback window, undoes the carry.
func rescheduleAsCarry(
	ctx context.Context,
	c *core.Context,
	issue *sharedtypes.Issue,
	previousDate, nextDate string,
	settings *sharedtypes.CoreSettings,
) (bool, *sharedtypes.Issue, error) {
	if previousDate == "" || previousDate == nextDate || settings == nil || !settings.CarryOverEnabled {
		return false, nil, nil
	}
	entry, err := c.DailyPlans.GetEntry(ctx, c.UserID, previousDate, issue.ID)
	if err != nil || entry == nil || entry.Status != sharedtypes.DailyPlanEntryStatusPlanned {
		return false, nil, err
	}
	if nextDate < previousDate {
		if entry.CarriedFromEntryID == nil {
			return false, nil, nil
		}
		from, err := c.DailyPlans.GetEntryByID(ctx, c.UserID, *entry.CarriedFromEntryID)
		if err != nil || from == nil || from.Date != nextDate || !withinRollbackWindow(from.ResolvedAt, c.Now(), settings) {
			return false, nil, err
		}
		undone, err := UndoCarry(ctx, c, issue.ID)
		return true, undone, err
	}
	if previousDate != c.CurrentDate() {
		return false, nil, nil
	}
	if !sharedtypes.CanCarryOver(issue.Status) {
		return false, nil, nil
	}
	worked, err := workedSecondsByIssueOnDate(ctx, c, previousDate)
	if err != nil || worked[issue.ID] <= 0 {
		return false, nil, err
	}
	carried, err := carryEntry(ctx, c, issue, *entry, nextDate, worked[issue.ID], true)
	return true, carried, err
}

// carryEntry resolves entry as carried and opens the next entry in the chain on target. The new
// entry starts a fresh delay baseline: the carried day was planned work, not a slip.
func carryEntry(
	ctx context.Context,
	c *core.Context,
	issue *sharedtypes.Issue,
	entry sharedtypes.DailyPlanEntry,
	target string,
	workedSeconds int,
	manual bool,
) (*sharedtypes.Issue, error) {
	now := c.Now()
	updated, err := setIssueTodoDate(ctx, c, issue.ID, target, now)
	if err != nil {
		return nil, err
	}
	if err := c.DailyPlans.ResolveEntry(ctx, entry.ID, now, sharedtypes.DailyPlanEntryStatusCarried, nil); err != nil {
		return nil, err
	}
	next, _, err := c.DailyPlans.UpsertCommittedEntryWithChain(ctx, c.UserID, target, carryOverSource, now, issue.ID, target, target, 0, 0)
	if err != nil {
		return nil, err
	}
	if next == nil {
		return nil, errors.New("carry-over entry was not created")
	}
	if err := c.DailyPlans.SetCarryLinks(ctx, entry.ID, now, entry.CarriedFromEntryID, &next.ID, entry.CarryCount); err != nil {
		return nil, err
	}
	if err := c.DailyPlans.SetCarryLinks(ctx, next.ID, now, &entry.ID, nil, entry.CarryCount+1); err != nil {
		return nil, err
	}
	if err := c.DailyPlans.AppendEvent(ctx, entry.ID, c.UserID, c.DeviceID, now, sharedtypes.DailyPlanEventCarriedOver, nil, map[string]any{
		"issueId":       issue.ID,
		"date":          entry.Date,
		"toDate":        target,
		"workedSeconds": workedSeconds,
		"manual":        manual,
	}); err != nil {
		return nil, err
	}
	if err := c.DailyPlans.AppendEvent(ctx, next.ID, c.UserID, c.DeviceID, now, sharedtypes.DailyPlanEventCommitted, nil, map[string]any{
		"issueId":     issue.ID,
		"date":        target,
		"carriedFrom": entry.Date,
	}); err != nil {
		return nil, err
	}
	return updated, recordCarryChange(ctx, c, updated, now, map[string]any{
		"todoForDate": target,
		"carriedFrom": entry.Date,
	})
}

// recordCarryChange logs the issue's moved plan date in the op log and tells clients.
func recordCarryChange(
	ctx context.Context,
	c *core.Context,
	issue *sharedtypes.Issue,
	now string,
	payload map[string]any,
) error {
	if issue == nil {
		return nil
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityIssue,
		EntityID:  fmt.Sprintf("%d", issue.ID),
		Action:    sharedtypes.OpActionUpdate,
		Payload:   payload,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return err
	}
	emit(c, sharedtypes.EventTypeIssueUpdated, issue)
	return nil
}

func setIssueTodoDate(ctx context.Context, c *core.Context, issueID int64, date, now string) (*sharedtypes.Issue, error) {
	return c.Issues.Update(ctx, issueID, c.UserID, now, struct {
		Title           sharedtypes.Patch[string]
		Description     sharedtypes.Patch[string]
		Status          sharedtypes.Patch[sharedtypes.IssueStatus]
		EstimateMinutes sharedtypes.Patch[int]
		Notes           sharedtypes.Patch[string]
		PinnedDaily     sharedtypes.Patch[bool]
		TodoForDate     sharedtypes.Patch[string]
		CompletedAt     sharedtypes.Patch[string]
		AbandonedAt     sharedtypes.Patch[string]
	}{
		TodoForDate: sharedtypes.Patch[string]{Set: true, Value: &date},
	})
}

// workedSecondsByIssueOnDate totals ended sessions that started on date, per issue. It uses the
// same day window as daily reports.
func workedSecondsByIssueOnDate(ctx context.Context, c *core.Context, date string) (map[int64]int, error) {
	start := date + "T00:00:00.000Z"
	end := date + "T23:59:59.999Z"
	history, err := c.Sessions.ListEnded(ctx, struct {
		UserID   string
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
		Since    *string
		Until    *string
		Limit    *int
		Offset   *int
	}{UserID: c.UserID, Since: &start, Until: &end})
	if err != nil {
		return nil, err
	}
	out := map[int64]int{}
	for _, entry := range history {
		if entry.DurationSeconds != nil {
			out[entry.IssueID] += *entry.DurationSeconds
		}
	}
	return out, nil
}

// meetsCarryThreshold reports whether a day's work counts as progress. A zero threshold accepts
// any session, however short.
func meetsCarryThreshold(workedSeconds, minMinutes int) bool {
	if minMinutes <= 0 {
		return workedSeconds > 0
	}
	return workedSeconds >= minMinutes*60
}

// nextCarryDate is the first working day after from, never before today: away and rest days are
// skipped, and catching up after downtime lands on today rather than a past day.
func nextCarryDate(from, today string, settings *sharedtypes.CoreSettings) string {
	start, err := time.Parse("2006-01-02", from)
	if err != nil {
		return today
	}
	candidate := start.AddDate(0, 0, 1)
	if floor, err := time.Parse("2006-01-02", today); err == nil && candidate.Before(floor) {
		candidate = floor
	}
	for range 30 {
		date := candidate.Format("2006-01-02")
		if settings == nil || (!settings.IsAwayDate(date) && !settings.IsConfiguredRestDate(date)) {
			return date
		}
		candidate = candidate.AddDate(0, 0, 1)
	}
	return candidate.Format("2006-01-02")
}

func withinRollbackWindow(resolvedAt *string, now string, settings *sharedtypes.CoreSettings) bool {
	if resolvedAt == nil {
		return false
	}
	resolved, err := time.Parse(time.RFC3339, *resolvedAt)
	if err != nil {
		return false
	}
	current, err := time.Parse(time.RFC3339, now)
	if err != nil {
		return false
	}
	return current.Sub(resolved) <= dailyPlanRollbackWindow(settings)
}

func carryOverStallAfter(settings *sharedtypes.CoreSettings) int {
	if settings == nil || settings.CarryOverStallAfter <= 0 {
		return sharedtypes.DefaultCarryOverStallAfter
	}
	return settings.CarryOverStallAfter
}

// dailyPlanStallComponent grows slowly with each carry past the stall limit; it is zero on the
// carry that first makes the issue stalled.
func dailyPlanStallComponent(entry sharedtypes.DailyPlanEntry, settings *sharedtypes.CoreSettings) float64 {
	return 0.5 * math.Log2(1+float64(max(0, entry.CarryCount-carryOverStallAfter(settings))))
}
