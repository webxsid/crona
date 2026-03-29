package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"crona/kernel/internal/core"

	"github.com/google/uuid"

	sharedtypes "crona/shared/types"
)

func CreateIssue(ctx context.Context, c *core.Context, input struct {
	StreamID        int64
	Title           string
	Description     *string
	EstimateMinutes *int
	Notes           *string
	TodoForDate     *string
}) (sharedtypes.Issue, error) {
	if strings.TrimSpace(input.Title) == "" {
		return sharedtypes.Issue{}, errors.New("issue title cannot be empty")
	}
	if input.EstimateMinutes != nil && *input.EstimateMinutes < 0 {
		return sharedtypes.Issue{}, errors.New("estimate must be >= 0")
	}
	nextID, err := c.Issues.NextID(ctx)
	if err != nil {
		return sharedtypes.Issue{}, err
	}
	issue := sharedtypes.Issue{
		ID:              nextID,
		StreamID:        input.StreamID,
		Title:           strings.TrimSpace(input.Title),
		Description:     normalizeOptionalString(input.Description),
		Status:          sharedtypes.IssueStatusBacklog,
		EstimateMinutes: input.EstimateMinutes,
		Notes:           input.Notes,
		TodoForDate:     input.TodoForDate,
	}
	if input.TodoForDate != nil && strings.TrimSpace(*input.TodoForDate) != "" {
		issue.Status = sharedtypes.IssueStatusPlanned
	}
	now := c.Now()
	if err := finalizeExpiredDailyPlanFailures(ctx, c, now); err != nil {
		return sharedtypes.Issue{}, err
	}
	created, err := c.Issues.Create(ctx, issue, c.UserID, now)
	if err != nil {
		return sharedtypes.Issue{}, err
	}
	if created.TodoForDate != nil && strings.TrimSpace(*created.TodoForDate) != "" {
		if err := commitIssueToDailyPlan(ctx, c, created.ID, strings.TrimSpace(*created.TodoForDate), now); err != nil {
			return sharedtypes.Issue{}, err
		}
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityIssue,
		EntityID:  fmt.Sprintf("%d", created.ID),
		Action:    sharedtypes.OpActionCreate,
		Payload:   created,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return sharedtypes.Issue{}, err
	}
	emit(c, sharedtypes.EventTypeIssueCreated, created)
	return created, nil
}

func UpdateIssue(ctx context.Context, c *core.Context, issueID int64, updates struct {
	Title           sharedtypes.Patch[string]
	Description     sharedtypes.Patch[string]
	EstimateMinutes sharedtypes.Patch[int]
	Notes           sharedtypes.Patch[string]
}) (*sharedtypes.Issue, error) {
	if updates.Title.Set && updates.Title.Value != nil && strings.TrimSpace(*updates.Title.Value) == "" {
		return nil, errors.New("issue title cannot be empty")
	}
	if updates.Title.Set && updates.Title.Value != nil {
		trimmed := strings.TrimSpace(*updates.Title.Value)
		updates.Title.Value = &trimmed
	}
	if updates.EstimateMinutes.Set && updates.EstimateMinutes.Value != nil && *updates.EstimateMinutes.Value < 0 {
		return nil, errors.New("estimate must be >= 0")
	}
	if updates.Description.Set {
		updates.Description.Value = normalizeOptionalString(updates.Description.Value)
	}
	now := c.Now()
	updated, err := c.Issues.Update(ctx, issueID, c.UserID, now, struct {
		Title           sharedtypes.Patch[string]
		Description     sharedtypes.Patch[string]
		Status          sharedtypes.Patch[sharedtypes.IssueStatus]
		EstimateMinutes sharedtypes.Patch[int]
		Notes           sharedtypes.Patch[string]
		TodoForDate     sharedtypes.Patch[string]
		CompletedAt     sharedtypes.Patch[string]
		AbandonedAt     sharedtypes.Patch[string]
	}{
		Title:           updates.Title,
		Description:     updates.Description,
		EstimateMinutes: updates.EstimateMinutes,
		Notes:           updates.Notes,
	})
	if err != nil {
		return nil, err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityIssue,
		EntityID:  fmt.Sprintf("%d", issueID),
		Action:    sharedtypes.OpActionUpdate,
		Payload:   updates,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return nil, err
	}
	if updated != nil {
		emit(c, sharedtypes.EventTypeIssueUpdated, updated)
	}
	return updated, nil
}

func ChangeIssueStatus(ctx context.Context, c *core.Context, issueID int64, nextStatus sharedtypes.IssueStatus, note *string) (*sharedtypes.Issue, error) {
	return changeIssueStatus(ctx, c, issueID, nextStatus, note, false)
}

func changeIssueStatus(ctx context.Context, c *core.Context, issueID int64, nextStatus sharedtypes.IssueStatus, note *string, allowWhenActive bool) (*sharedtypes.Issue, error) {
	issue, err := c.Issues.GetByID(ctx, issueID, c.UserID)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, errors.New("issue not found")
	}
	activeSession, err := c.Sessions.GetActiveSession(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	if !allowWhenActive && activeSession != nil && activeSession.IssueID == issueID {
		return nil, errors.New("cannot change issue status while a focus session is active")
	}
	currentStatus := sharedtypes.NormalizeIssueStatus(issue.Status)
	nextStatus = sharedtypes.NormalizeIssueStatus(nextStatus)
	if !sharedtypes.IsValidIssueTransition(currentStatus, nextStatus) {
		return nil, errors.New("invalid status transition")
	}
	if nextStatus == sharedtypes.IssueStatusBlocked && (note == nil || strings.TrimSpace(*note) == "") {
		return nil, errors.New("blocked status requires a blocker note")
	}
	if nextStatus == sharedtypes.IssueStatusAbandoned && (note == nil || strings.TrimSpace(*note) == "") {
		return nil, errors.New("abandoned status requires a reason")
	}

	now := c.Now()
	notes := appendIssueStatusNote(issue.Notes, now, currentStatus, nextStatus, note)
	var completedAt *string
	var abandonedAt *string
	switch nextStatus {
	case sharedtypes.IssueStatusDone:
		completedAt = &now
	case sharedtypes.IssueStatusAbandoned:
		abandonedAt = &now
	}
	if nextStatus != sharedtypes.IssueStatusDone {
		completedAt = nil
	}
	if nextStatus != sharedtypes.IssueStatusAbandoned {
		abandonedAt = nil
	}

	updated, err := c.Issues.Update(ctx, issueID, c.UserID, now, struct {
		Title           sharedtypes.Patch[string]
		Description     sharedtypes.Patch[string]
		Status          sharedtypes.Patch[sharedtypes.IssueStatus]
		EstimateMinutes sharedtypes.Patch[int]
		Notes           sharedtypes.Patch[string]
		TodoForDate     sharedtypes.Patch[string]
		CompletedAt     sharedtypes.Patch[string]
		AbandonedAt     sharedtypes.Patch[string]
	}{
		Status:      sharedtypes.Patch[sharedtypes.IssueStatus]{Set: true, Value: &nextStatus},
		Notes:       sharedtypes.Patch[string]{Set: notes != issue.Notes, Value: notes},
		CompletedAt: sharedtypes.Patch[string]{Set: true, Value: completedAt},
		AbandonedAt: sharedtypes.Patch[string]{Set: true, Value: abandonedAt},
	})
	if err != nil {
		return nil, err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityIssue,
		EntityID:  fmt.Sprintf("%d", issueID),
		Action:    sharedtypes.OpActionUpdate,
		Payload:   map[string]any{"from": currentStatus, "to": nextStatus, "note": note},
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return nil, err
	}
	if updated != nil {
		resolveDate := entryResolveDate(updated, now)
		if resolveDate != "" {
			switch nextStatus {
			case sharedtypes.IssueStatusDone:
				if err := resolveDailyPlanEntry(ctx, c, issueID, resolveDate, now, sharedtypes.DailyPlanEntryStatusCompleted, sharedtypes.DailyPlanEventCompleted, nil); err != nil {
					return nil, err
				}
			case sharedtypes.IssueStatusAbandoned:
				reason := sharedtypes.DailyPlanFailureReasonAbandoned
				if err := resolveDailyPlanEntry(ctx, c, issueID, resolveDate, now, sharedtypes.DailyPlanEntryStatusAbandoned, sharedtypes.DailyPlanEventAbandoned, &reason); err != nil {
					return nil, err
				}
			}
		}
		emit(c, sharedtypes.EventTypeIssueUpdated, updated)
	}
	return updated, nil
}

func appendIssueStatusNote(existing *string, now string, from, to sharedtypes.IssueStatus, note *string) *string {
	trimmedNote := ""
	if note != nil {
		trimmedNote = strings.TrimSpace(*note)
	}
	if trimmedNote == "" && to != sharedtypes.IssueStatusBlocked && to != sharedtypes.IssueStatusInReview {
		return existing
	}

	entry := fmt.Sprintf("[%s] %s -> %s", now, from, to)
	if trimmedNote != "" {
		entry += ": " + trimmedNote
	}

	if existing == nil || strings.TrimSpace(*existing) == "" {
		return &entry
	}
	merged := strings.TrimSpace(*existing) + "\n" + entry
	return &merged
}

func DeleteIssue(ctx context.Context, c *core.Context, issueID int64) error {
	now := c.Now()
	if err := c.Issues.SoftDelete(ctx, issueID, c.UserID, now); err != nil {
		return err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityIssue,
		EntityID:  fmt.Sprintf("%d", issueID),
		Action:    sharedtypes.OpActionDelete,
		Payload:   nil,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return err
	}
	emit(c, sharedtypes.EventTypeIssueDeleted, sharedtypes.IDEventPayload{ID: issueID})
	return nil
}

func RestoreIssue(ctx context.Context, c *core.Context, issueID int64) error {
	now := c.Now()
	if err := c.Issues.RestoreDeletedByID(ctx, issueID, c.UserID, now); err != nil {
		return err
	}
	return c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityIssue,
		EntityID:  fmt.Sprintf("%d", issueID),
		Action:    sharedtypes.OpActionRestore,
		Payload:   nil,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	})
}

func ListIssuesByStream(ctx context.Context, c *core.Context, streamID int64) ([]sharedtypes.Issue, error) {
	issues, err := c.Issues.ListByStream(ctx, streamID, c.UserID)
	if err != nil {
		return nil, err
	}
	sortIssues(issues, loadListSortSettings(ctx, c).issueSort)
	return issues, nil
}

func ListAllIssues(ctx context.Context, c *core.Context) ([]sharedtypes.IssueWithMeta, error) {
	issues, err := c.Issues.ListAll(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	sortIssuesWithMeta(issues, loadListSortSettings(ctx, c).issueSort)
	return issues, nil
}

func MarkIssueTodoForDate(ctx context.Context, c *core.Context, issueID int64, todoForDate string) (*sharedtypes.Issue, error) {
	if err := ensureDailyPlanDate(todoForDate); err != nil {
		return nil, err
	}
	issue, err := c.Issues.GetByID(ctx, issueID, c.UserID)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, errors.New("issue not found")
	}
	now := c.Now()
	if err := finalizeExpiredDailyPlanFailures(ctx, c, now); err != nil {
		return nil, err
	}
	previousDate := issueCommittedDate(issue)
	settings, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	nextStatus := sharedtypes.AutoStatusOnTodoAssigned(issue.Status)
	updated, err := c.Issues.Update(ctx, issueID, c.UserID, now, struct {
		Title           sharedtypes.Patch[string]
		Description     sharedtypes.Patch[string]
		Status          sharedtypes.Patch[sharedtypes.IssueStatus]
		EstimateMinutes sharedtypes.Patch[int]
		Notes           sharedtypes.Patch[string]
		TodoForDate     sharedtypes.Patch[string]
		CompletedAt     sharedtypes.Patch[string]
		AbandonedAt     sharedtypes.Patch[string]
	}{
		Status:      sharedtypes.Patch[sharedtypes.IssueStatus]{Set: nextStatus != sharedtypes.NormalizeIssueStatus(issue.Status), Value: &nextStatus},
		TodoForDate: sharedtypes.Patch[string]{Set: true, Value: &todoForDate},
	})
	if err != nil {
		return nil, err
	}
	var (
		baselineDate       = todoForDate
		currentPlannedDate = todoForDate
		postponeCount      = 0
		maxDelayedDays     = 0
	)
	if previousDate != "" && previousDate != todoForDate {
		previousEntry, err := c.DailyPlans.GetEntry(ctx, c.UserID, previousDate, issueID)
		if err != nil {
			return nil, err
		}
		baselineDate, currentPlannedDate, postponeCount, maxDelayedDays = nextDailyPlanChainState(previousEntry, previousDate, todoForDate, settings)
		if previousEntry != nil {
			if err := c.DailyPlans.UpdateChainState(ctx, previousEntry.ID, now, baselineDate, currentPlannedDate, postponeCount, maxDelayedDays); err != nil {
				return nil, err
			}
		}
		if err := markDailyPlanPendingFailure(ctx, c, issueID, previousDate, now, sharedtypes.DailyPlanFailureReasonMoved, sharedtypes.DailyPlanEventRescheduled); err != nil {
			return nil, err
		}
	}
	if err := commitIssueToDailyPlanWithChain(ctx, c, issueID, todoForDate, now, baselineDate, currentPlannedDate, postponeCount, maxDelayedDays); err != nil {
		return nil, err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityIssue,
		EntityID:  fmt.Sprintf("%d", issueID),
		Action:    sharedtypes.OpActionUpdate,
		Payload:   map[string]any{"todoForDate": todoForDate, "status": nextStatus},
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return nil, err
	}
	if updated != nil {
		emit(c, sharedtypes.EventTypeIssueUpdated, updated)
	}
	return updated, nil
}

func MarkIssueTodoForToday(ctx context.Context, c *core.Context, issueID int64) (*sharedtypes.Issue, error) {
	today := strings.Split(c.Now(), "T")[0]
	if today == "" {
		return nil, errors.New("invalid date")
	}
	return MarkIssueTodoForDate(ctx, c, issueID, today)
}

func ClearIssueTodoForDate(ctx context.Context, c *core.Context, issueID int64) (*sharedtypes.Issue, error) {
	issue, err := c.Issues.GetByID(ctx, issueID, c.UserID)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, errors.New("issue not found")
	}
	now := c.Now()
	if err := finalizeExpiredDailyPlanFailures(ctx, c, now); err != nil {
		return nil, err
	}
	updated, err := c.Issues.Update(ctx, issueID, c.UserID, now, struct {
		Title           sharedtypes.Patch[string]
		Description     sharedtypes.Patch[string]
		Status          sharedtypes.Patch[sharedtypes.IssueStatus]
		EstimateMinutes sharedtypes.Patch[int]
		Notes           sharedtypes.Patch[string]
		TodoForDate     sharedtypes.Patch[string]
		CompletedAt     sharedtypes.Patch[string]
		AbandonedAt     sharedtypes.Patch[string]
	}{
		TodoForDate: sharedtypes.Patch[string]{Set: true, Value: nil},
	})
	if err != nil {
		return nil, err
	}
	if previousDate := issueCommittedDate(issue); previousDate != "" {
		if err := markDailyPlanPendingFailure(ctx, c, issueID, previousDate, now, sharedtypes.DailyPlanFailureReasonCleared, sharedtypes.DailyPlanEventCleared); err != nil {
			return nil, err
		}
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityIssue,
		EntityID:  fmt.Sprintf("%d", issueID),
		Action:    sharedtypes.OpActionUpdate,
		Payload:   map[string]any{"todoForDate": nil},
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return nil, err
	}
	if updated != nil {
		updated.TodoForDate = nil
		emit(c, sharedtypes.EventTypeIssueUpdated, updated)
	}
	return updated, nil
}

func ClearTodayTodos(ctx context.Context, c *core.Context) error {
	today := strings.Split(c.Now(), "T")[0]
	if today == "" {
		return errors.New("invalid date")
	}
	issues, err := c.Issues.ListByTodoForDate(ctx, today, c.UserID)
	if err != nil {
		return err
	}
	for _, issue := range issues {
		if _, err := ClearIssueTodoForDate(ctx, c, issue.ID); err != nil {
			return err
		}
	}
	return nil
}

func ComputeDailyIssueSummaryForDate(ctx context.Context, c *core.Context, date string) (sharedtypes.DailyIssueSummary, error) {
	if date == "" {
		return sharedtypes.DailyIssueSummary{}, errors.New("invalid date")
	}
	issues, err := c.Issues.ListByTodoForDate(ctx, date, c.UserID)
	if err != nil {
		return sharedtypes.DailyIssueSummary{}, err
	}
	sortIssues(issues, loadListSortSettings(ctx, c).issueSort)
	totalEstimatedMinutes := 0
	completedIssues := 0
	abandonedIssues := 0
	issueIDs := map[int64]bool{}
	for _, issue := range issues {
		totalEstimatedMinutes += derefIssueEstimate(issue.EstimateMinutes)
		if issue.CompletedAt != nil && strings.HasPrefix(*issue.CompletedAt, date) {
			completedIssues++
		}
		if issue.AbandonedAt != nil && strings.HasPrefix(*issue.AbandonedAt, date) {
			abandonedIssues++
		}
		issueIDs[issue.ID] = true
	}
	dayStart := date + "T00:00:00.000Z"
	dayEnd := date + "T23:59:59.999Z"
	endedSessions, err := c.Sessions.ListEnded(ctx, struct {
		UserID   string
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
		Since    *string
		Until    *string
		Limit    *int
		Offset   *int
	}{
		UserID: c.UserID,
		Since:  &dayStart,
		Until:  &dayEnd,
	})
	if err != nil {
		return sharedtypes.DailyIssueSummary{}, err
	}
	workedSeconds := 0
	for _, session := range endedSessions {
		if issueIDs[session.IssueID] {
			workedSeconds += derefIssueEstimate(session.DurationSeconds)
		}
	}
	return sharedtypes.DailyIssueSummary{
		Date:                  date,
		TotalIssues:           len(issues),
		Issues:                issues,
		TotalEstimatedMinutes: totalEstimatedMinutes,
		CompletedIssues:       completedIssues,
		AbandonedIssues:       abandonedIssues,
		WorkedSeconds:         workedSeconds,
	}, nil
}

func ComputeDailyIssueSummaryForToday(ctx context.Context, c *core.Context) (sharedtypes.DailyIssueSummary, error) {
	today := strings.Split(c.Now(), "T")[0]
	if today == "" {
		return sharedtypes.DailyIssueSummary{}, errors.New("invalid date")
	}
	return ComputeDailyIssueSummaryForDate(ctx, c, today)
}

func derefIssueEstimate(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
