package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	storemodels "crona/kernel/internal/store/models"
	sharedtypes "crona/shared/types"

	"github.com/uptrace/bun"
)

type DailyPlanRepository struct {
	db *bun.DB
}

func NewDailyPlanRepository(db *bun.DB) *DailyPlanRepository {
	return &DailyPlanRepository{db: db}
}

func (r *DailyPlanRepository) EnsurePlan(ctx context.Context, userID, date, now string) (*sharedtypes.DailyPlan, error) {
	var model storemodels.DailyPlanModel
	err := r.db.NewSelect().
		Model(&model).
		Where("user_id = ?", userID).
		Where("date = ?", date).
		Limit(1).
		Scan(ctx)
	if err == nil {
		return &sharedtypes.DailyPlan{
			ID:        model.ID,
			Date:      model.Date,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	model = storemodels.DailyPlanModel{
		ID:        fmt.Sprintf("daily-plan:%s:%s", userID, date),
		UserID:    userID,
		Date:      date,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
		return nil, err
	}
	return &sharedtypes.DailyPlan{
		ID:        model.ID,
		Date:      model.Date,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}

func (r *DailyPlanRepository) GetByDate(ctx context.Context, userID, date string) (*sharedtypes.DailyPlan, error) {
	type planRow struct {
		ID        string `bun:"id"`
		Date      string `bun:"date"`
		CreatedAt string `bun:"created_at"`
		UpdatedAt string `bun:"updated_at"`
	}
	var plan planRow
	err := r.db.NewSelect().
		TableExpr("daily_plans").
		Column("id", "date", "created_at", "updated_at").
		Where("user_id = ?", userID).
		Where("date = ?", date).
		Limit(1).
		Scan(ctx, &plan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	entries, err := r.listEntriesByPlanIDs(ctx, []string{plan.ID})
	if err != nil {
		return nil, err
	}
	return &sharedtypes.DailyPlan{
		ID:        plan.ID,
		Date:      plan.Date,
		CreatedAt: plan.CreatedAt,
		UpdatedAt: plan.UpdatedAt,
		Entries:   entries,
	}, nil
}

func (r *DailyPlanRepository) ListByDateRange(ctx context.Context, userID, startDate, endDate string) ([]sharedtypes.DailyPlan, error) {
	type planRow struct {
		ID        string `bun:"id"`
		Date      string `bun:"date"`
		CreatedAt string `bun:"created_at"`
		UpdatedAt string `bun:"updated_at"`
	}
	var plans []planRow
	if err := r.db.NewSelect().
		TableExpr("daily_plans").
		Column("id", "date", "created_at", "updated_at").
		Where("user_id = ?", userID).
		Where("date >= ?", startDate).
		Where("date <= ?", endDate).
		OrderExpr("date ASC").
		Scan(ctx, &plans); err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return nil, nil
	}
	planIDs := make([]string, 0, len(plans))
	for _, plan := range plans {
		planIDs = append(planIDs, plan.ID)
	}
	entriesByPlanID, err := r.listEntriesGroupedByPlanIDs(ctx, planIDs)
	if err != nil {
		return nil, err
	}
	out := make([]sharedtypes.DailyPlan, 0, len(plans))
	for _, plan := range plans {
		out = append(out, sharedtypes.DailyPlan{
			ID:        plan.ID,
			Date:      plan.Date,
			CreatedAt: plan.CreatedAt,
			UpdatedAt: plan.UpdatedAt,
			Entries:   entriesByPlanID[plan.ID],
		})
	}
	return out, nil
}

func (r *DailyPlanRepository) GetEntry(ctx context.Context, userID, date string, issueID int64) (*sharedtypes.DailyPlanEntry, error) {
	plan, err := r.GetByDate(ctx, userID, date)
	if err != nil || plan == nil {
		return nil, err
	}
	for _, entry := range plan.Entries {
		if entry.IssueID == issueID {
			copy := entry
			return &copy, nil
		}
	}
	return nil, nil
}

func (r *DailyPlanRepository) UpsertCommittedEntry(ctx context.Context, userID, date, source, now string, issueID int64) (*sharedtypes.DailyPlanEntry, string, error) {
	return r.UpsertCommittedEntryWithChain(ctx, userID, date, source, now, issueID, date, date, 0, 0)
}

func (r *DailyPlanRepository) UpsertCommittedEntryWithChain(ctx context.Context, userID, date, source, now string, issueID int64, baselineDate, currentPlannedDate string, postponeCount, maxDelayedDays int) (*sharedtypes.DailyPlanEntry, string, error) {
	plan, err := r.EnsurePlan(ctx, userID, date, now)
	if err != nil {
		return nil, "", err
	}
	issueInternalID, err := resolveIssueInternalID(ctx, r.db, issueID, userID)
	if err != nil {
		return nil, "", err
	}
	existing, err := r.GetEntry(ctx, userID, date, issueID)
	if err != nil {
		return nil, "", err
	}
	if existing == nil {
		model := storemodels.DailyPlanEntryModel{
			ID:                 fmt.Sprintf("daily-plan-entry:%s:%d", plan.ID, issueID),
			PlanID:             plan.ID,
			IssueID:            issueInternalID,
			Source:             source,
			Status:             string(sharedtypes.DailyPlanEntryStatusPlanned),
			CommittedAt:        now,
			BaselineDate:       baselineDate,
			CurrentPlannedDate: currentPlannedDate,
			PostponeCount:      postponeCount,
			MaxDelayedDays:     maxDelayedDays,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
			return nil, "", err
		}
		entry, err := r.GetEntry(ctx, userID, date, issueID)
		return entry, "created", err
	}
	q := r.db.NewUpdate().
		Model((*storemodels.DailyPlanEntryModel)(nil)).
		Where("id = ?", existing.ID).
		Set("source = ?", source).
		Set("baseline_date = ?", baselineDate).
		Set("current_planned_date = ?", currentPlannedDate).
		Set("postpone_count = ?", postponeCount).
		Set("max_delayed_days = ?", maxDelayedDays).
		Set("updated_at = ?", now)
	if existing.PendingFailureAt != nil {
		q = q.Set("status = ?", string(sharedtypes.DailyPlanEntryStatusPlanned)).
			Set("failure_reason = NULL").
			Set("pending_failure_reason = NULL").
			Set("pending_failure_at = NULL").
			Set("resolved_at = NULL")
	}
	if _, err := q.Exec(ctx); err != nil {
		return nil, "", err
	}
	entry, err := r.GetEntry(ctx, userID, date, issueID)
	if existing.PendingFailureAt != nil {
		return entry, "restored", err
	}
	return entry, "unchanged", err
}

func (r *DailyPlanRepository) MarkPendingFailure(ctx context.Context, userID, date, now string, issueID int64, reason sharedtypes.DailyPlanFailureReason) (*sharedtypes.DailyPlanEntry, error) {
	entry, err := r.GetEntry(ctx, userID, date, issueID)
	if err != nil || entry == nil {
		return entry, err
	}
	reasonText := string(reason)
	if _, err := r.db.NewUpdate().
		Model((*storemodels.DailyPlanEntryModel)(nil)).
		Where("id = ?", entry.ID).
		Set("pending_failure_reason = ?", reasonText).
		Set("pending_failure_at = ?", now).
		Set("updated_at = ?", now).
		Exec(ctx); err != nil {
		return nil, err
	}
	return r.GetEntry(ctx, userID, date, issueID)
}

func (r *DailyPlanRepository) ResolveEntry(ctx context.Context, entryID, now string, status sharedtypes.DailyPlanEntryStatus, reason *sharedtypes.DailyPlanFailureReason) error {
	q := r.db.NewUpdate().
		Model((*storemodels.DailyPlanEntryModel)(nil)).
		Where("id = ?", entryID).
		Set("status = ?", string(status)).
		Set("pending_failure_reason = NULL").
		Set("pending_failure_at = NULL").
		Set("updated_at = ?", now).
		Set("resolved_at = ?", now)
	if reason == nil {
		q = q.Set("failure_reason = NULL")
	} else {
		q = q.Set("failure_reason = ?", string(*reason))
	}
	_, err := q.Exec(ctx)
	return err
}

func (r *DailyPlanRepository) UpdateChainState(ctx context.Context, entryID, now, baselineDate, currentPlannedDate string, postponeCount, maxDelayedDays int) error {
	_, err := r.db.NewUpdate().
		Model((*storemodels.DailyPlanEntryModel)(nil)).
		Where("id = ?", entryID).
		Set("baseline_date = ?", baselineDate).
		Set("current_planned_date = ?", currentPlannedDate).
		Set("postpone_count = ?", postponeCount).
		Set("max_delayed_days = ?", maxDelayedDays).
		Set("updated_at = ?", now).
		Exec(ctx)
	return err
}

func (r *DailyPlanRepository) ListPendingFailuresBefore(ctx context.Context, userID, cutoff string) ([]sharedtypes.DailyPlanEntry, error) {
	type row struct {
		ID                   string  `bun:"id"`
		Date                 string  `bun:"date"`
		IssuePublicID        int64   `bun:"issue_public_id"`
		Source               string  `bun:"source"`
		Status               string  `bun:"status"`
		FailureReason        *string `bun:"failure_reason"`
		PendingFailureReason *string `bun:"pending_failure_reason"`
		CommittedAt          string  `bun:"committed_at"`
		PendingFailureAt     *string `bun:"pending_failure_at"`
		ResolvedAt           *string `bun:"resolved_at"`
		BaselineDate         string  `bun:"baseline_date"`
		CurrentPlannedDate   string  `bun:"current_planned_date"`
		PostponeCount        int     `bun:"postpone_count"`
		MaxDelayedDays       int     `bun:"max_delayed_days"`
	}
	var rows []row
	if err := r.db.NewSelect().
		TableExpr("daily_plan_entries AS e").
		Join("INNER JOIN daily_plans AS p ON p.id = e.plan_id").
		Join("INNER JOIN issues ON issues.id = e.issue_id").
		ColumnExpr("e.id").
		ColumnExpr("p.date").
		ColumnExpr("issues.public_id AS issue_public_id").
		ColumnExpr("e.source").
		ColumnExpr("e.status").
		ColumnExpr("e.failure_reason").
		ColumnExpr("e.pending_failure_reason").
		ColumnExpr("e.committed_at").
		ColumnExpr("e.pending_failure_at").
		ColumnExpr("e.resolved_at").
		ColumnExpr("COALESCE(NULLIF(e.baseline_date, ''), p.date) AS baseline_date").
		ColumnExpr("COALESCE(NULLIF(e.current_planned_date, ''), p.date) AS current_planned_date").
		ColumnExpr("COALESCE(e.postpone_count, 0) AS postpone_count").
		ColumnExpr("COALESCE(e.max_delayed_days, 0) AS max_delayed_days").
		Where("p.user_id = ?", userID).
		Where("e.pending_failure_at IS NOT NULL").
		Where("e.pending_failure_at <= ?", cutoff).
		Where("e.status = ?", string(sharedtypes.DailyPlanEntryStatusPlanned)).
		OrderExpr("e.pending_failure_at ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.DailyPlanEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapDailyPlanEntryRow(row.ID, row.Date, row.IssuePublicID, row.Source, row.Status, row.FailureReason, row.PendingFailureReason, row.CommittedAt, row.PendingFailureAt, row.ResolvedAt, row.BaselineDate, row.CurrentPlannedDate, row.PostponeCount, row.MaxDelayedDays))
	}
	return out, nil
}

func (r *DailyPlanRepository) AppendEvent(ctx context.Context, entryID, userID, deviceID, now string, eventType sharedtypes.DailyPlanEventType, reason *sharedtypes.DailyPlanFailureReason, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	var reasonText *string
	if reason != nil {
		text := string(*reason)
		reasonText = &text
	}
	_, err = r.db.NewInsert().Model(&storemodels.DailyPlanEventModel{
		ID:            fmt.Sprintf("daily-plan-event:%s:%d", entryID, time.Now().UTC().UnixNano()),
		PlanEntryID:   entryID,
		UserID:        userID,
		DeviceID:      deviceID,
		EventType:     string(eventType),
		FailureReason: reasonText,
		Payload:       string(body),
		Timestamp:     now,
	}).Exec(ctx)
	return err
}

func (r *DailyPlanRepository) listEntriesByPlanIDs(ctx context.Context, planIDs []string) ([]sharedtypes.DailyPlanEntry, error) {
	grouped, err := r.listEntriesGroupedByPlanIDs(ctx, planIDs)
	if err != nil {
		return nil, err
	}
	if len(planIDs) == 0 {
		return nil, nil
	}
	return grouped[planIDs[0]], nil
}

func (r *DailyPlanRepository) listEntriesGroupedByPlanIDs(ctx context.Context, planIDs []string) (map[string][]sharedtypes.DailyPlanEntry, error) {
	if len(planIDs) == 0 {
		return map[string][]sharedtypes.DailyPlanEntry{}, nil
	}
	type row struct {
		PlanID               string  `bun:"plan_id"`
		ID                   string  `bun:"id"`
		Date                 string  `bun:"date"`
		IssuePublicID        int64   `bun:"issue_public_id"`
		Source               string  `bun:"source"`
		Status               string  `bun:"status"`
		FailureReason        *string `bun:"failure_reason"`
		PendingFailureReason *string `bun:"pending_failure_reason"`
		CommittedAt          string  `bun:"committed_at"`
		PendingFailureAt     *string `bun:"pending_failure_at"`
		ResolvedAt           *string `bun:"resolved_at"`
		BaselineDate         string  `bun:"baseline_date"`
		CurrentPlannedDate   string  `bun:"current_planned_date"`
		PostponeCount        int     `bun:"postpone_count"`
		MaxDelayedDays       int     `bun:"max_delayed_days"`
	}
	var rows []row
	if err := r.db.NewSelect().
		TableExpr("daily_plan_entries AS e").
		Join("INNER JOIN daily_plans AS p ON p.id = e.plan_id").
		Join("INNER JOIN issues ON issues.id = e.issue_id").
		ColumnExpr("e.plan_id").
		ColumnExpr("e.id").
		ColumnExpr("p.date").
		ColumnExpr("issues.public_id AS issue_public_id").
		ColumnExpr("e.source").
		ColumnExpr("e.status").
		ColumnExpr("e.failure_reason").
		ColumnExpr("e.pending_failure_reason").
		ColumnExpr("e.committed_at").
		ColumnExpr("e.pending_failure_at").
		ColumnExpr("e.resolved_at").
		ColumnExpr("COALESCE(NULLIF(e.baseline_date, ''), p.date) AS baseline_date").
		ColumnExpr("COALESCE(NULLIF(e.current_planned_date, ''), p.date) AS current_planned_date").
		ColumnExpr("COALESCE(e.postpone_count, 0) AS postpone_count").
		ColumnExpr("COALESCE(e.max_delayed_days, 0) AS max_delayed_days").
		Where("e.plan_id IN (?)", bun.In(planIDs)).
		OrderExpr("p.date ASC, e.committed_at ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	events, err := r.listEventsByPlanIDs(ctx, planIDs)
	if err != nil {
		return nil, err
	}
	grouped := make(map[string][]sharedtypes.DailyPlanEntry, len(planIDs))
	for _, row := range rows {
		entry := mapDailyPlanEntryRow(row.ID, row.Date, row.IssuePublicID, row.Source, row.Status, row.FailureReason, row.PendingFailureReason, row.CommittedAt, row.PendingFailureAt, row.ResolvedAt, row.BaselineDate, row.CurrentPlannedDate, row.PostponeCount, row.MaxDelayedDays)
		entry.Events = events[row.ID]
		grouped[row.PlanID] = append(grouped[row.PlanID], entry)
	}
	return grouped, nil
}

func (r *DailyPlanRepository) listEventsByPlanIDs(ctx context.Context, planIDs []string) (map[string][]sharedtypes.DailyPlanEvent, error) {
	if len(planIDs) == 0 {
		return map[string][]sharedtypes.DailyPlanEvent{}, nil
	}
	type row struct {
		ID            string  `bun:"id"`
		PlanEntryID   string  `bun:"plan_entry_id"`
		EventType     string  `bun:"event_type"`
		FailureReason *string `bun:"failure_reason"`
		Timestamp     string  `bun:"timestamp"`
	}
	var rows []row
	if err := r.db.NewSelect().
		TableExpr("daily_plan_events AS e").
		Join("INNER JOIN daily_plan_entries AS dpe ON dpe.id = e.plan_entry_id").
		ColumnExpr("e.id").
		ColumnExpr("e.plan_entry_id").
		ColumnExpr("e.event_type").
		ColumnExpr("e.failure_reason").
		ColumnExpr("e.timestamp").
		Where("dpe.plan_id IN (?)", bun.In(planIDs)).
		OrderExpr("e.timestamp ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make(map[string][]sharedtypes.DailyPlanEvent, len(rows))
	for _, row := range rows {
		var reason *sharedtypes.DailyPlanFailureReason
		if row.FailureReason != nil {
			value := sharedtypes.DailyPlanFailureReason(*row.FailureReason)
			reason = &value
		}
		out[row.PlanEntryID] = append(out[row.PlanEntryID], sharedtypes.DailyPlanEvent{
			ID:            row.ID,
			EntryID:       row.PlanEntryID,
			Type:          sharedtypes.DailyPlanEventType(row.EventType),
			FailureReason: reason,
			Timestamp:     row.Timestamp,
		})
	}
	return out, nil
}

func (r *DailyPlanRepository) ListActiveEntries(ctx context.Context, userID string) ([]sharedtypes.DailyPlanEntry, error) {
	type row struct {
		ID                   string  `bun:"id"`
		Date                 string  `bun:"date"`
		IssuePublicID        int64   `bun:"issue_public_id"`
		Source               string  `bun:"source"`
		Status               string  `bun:"status"`
		FailureReason        *string `bun:"failure_reason"`
		PendingFailureReason *string `bun:"pending_failure_reason"`
		CommittedAt          string  `bun:"committed_at"`
		PendingFailureAt     *string `bun:"pending_failure_at"`
		ResolvedAt           *string `bun:"resolved_at"`
		BaselineDate         string  `bun:"baseline_date"`
		CurrentPlannedDate   string  `bun:"current_planned_date"`
		PostponeCount        int     `bun:"postpone_count"`
		MaxDelayedDays       int     `bun:"max_delayed_days"`
	}
	var rows []row
	if err := r.db.NewSelect().
		TableExpr("daily_plan_entries AS e").
		Join("INNER JOIN daily_plans AS p ON p.id = e.plan_id").
		Join("INNER JOIN issues ON issues.id = e.issue_id").
		ColumnExpr("e.id").
		ColumnExpr("p.date").
		ColumnExpr("issues.public_id AS issue_public_id").
		ColumnExpr("e.source").
		ColumnExpr("e.status").
		ColumnExpr("e.failure_reason").
		ColumnExpr("e.pending_failure_reason").
		ColumnExpr("e.committed_at").
		ColumnExpr("e.pending_failure_at").
		ColumnExpr("e.resolved_at").
		ColumnExpr("COALESCE(NULLIF(e.baseline_date, ''), p.date) AS baseline_date").
		ColumnExpr("COALESCE(NULLIF(e.current_planned_date, ''), p.date) AS current_planned_date").
		ColumnExpr("COALESCE(e.postpone_count, 0) AS postpone_count").
		ColumnExpr("COALESCE(e.max_delayed_days, 0) AS max_delayed_days").
		Where("p.user_id = ?", userID).
		Where("e.status = ?", string(sharedtypes.DailyPlanEntryStatusPlanned)).
		Where("issues.todo_for_date = p.date").
		Where("issues.status NOT IN (?, ?)", string(sharedtypes.IssueStatusDone), string(sharedtypes.IssueStatusAbandoned)).
		OrderExpr("p.date ASC, issues.public_id ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.DailyPlanEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapDailyPlanEntryRow(row.ID, row.Date, row.IssuePublicID, row.Source, row.Status, row.FailureReason, row.PendingFailureReason, row.CommittedAt, row.PendingFailureAt, row.ResolvedAt, row.BaselineDate, row.CurrentPlannedDate, row.PostponeCount, row.MaxDelayedDays))
	}
	return out, nil
}

func mapDailyPlanEntryRow(id, date string, issueID int64, source, status string, failureReason, pendingFailureReason *string, committedAt string, pendingFailureAt, resolvedAt *string, baselineDate, currentPlannedDate string, postponeCount, maxDelayedDays int) sharedtypes.DailyPlanEntry {
	var finalReason *sharedtypes.DailyPlanFailureReason
	if failureReason != nil {
		value := sharedtypes.DailyPlanFailureReason(*failureReason)
		finalReason = &value
	}
	var pendingReason *sharedtypes.DailyPlanFailureReason
	if pendingFailureReason != nil {
		value := sharedtypes.DailyPlanFailureReason(*pendingFailureReason)
		pendingReason = &value
	}
	return sharedtypes.DailyPlanEntry{
		ID:                   id,
		Date:                 date,
		IssueID:              issueID,
		Source:               source,
		Status:               sharedtypes.DailyPlanEntryStatus(status),
		FailureReason:        finalReason,
		PendingFailureReason: pendingReason,
		CommittedAt:          committedAt,
		PendingFailureAt:     pendingFailureAt,
		ResolvedAt:           resolvedAt,
		BaselineDate:         baselineDate,
		CurrentPlannedDate:   currentPlannedDate,
		PostponeCount:        postponeCount,
		MaxDelayedDays:       maxDelayedDays,
	}
}
