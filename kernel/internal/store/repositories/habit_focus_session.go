package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"

	storemodels "crona/kernel/internal/store/models"
	"github.com/uptrace/bun"
)

type HabitFocusSessionRepository struct {
	db *bun.DB
}

func NewHabitFocusSessionRepository(db *bun.DB) *HabitFocusSessionRepository {
	return &HabitFocusSessionRepository{db: db}
}

func (r *HabitFocusSessionRepository) NextID(ctx context.Context) (int64, error) {
	return nextPublicID(ctx, r.db, "habit_focus_sessions")
}

func (r *HabitFocusSessionRepository) Create(ctx context.Context, entry sharedtypes.HabitCompletion, userID string, now string) (sharedtypes.HabitCompletion, error) {
	habitInternalID, err := resolveHabitInternalID(ctx, r.db, entry.HabitID, userID)
	if err != nil {
		return sharedtypes.HabitCompletion{}, err
	}
	if habitInternalID == "" {
		return sharedtypes.HabitCompletion{}, errors.New("habit not found")
	}
	model := storemodels.HabitFocusSessionModel{
		InternalID:      habitFocusInternalID(entry.ID),
		PublicID:        entry.ID,
		HabitID:         habitInternalID,
		Kind:            string(sharedtypes.NormalizeHabitHistoryKind(entry.Kind)),
		Date:            entry.Date,
		StartedAt:       startedAtOrNow(entry.StartedAt, now),
		EndedAt:         entry.EndedAt,
		DurationMinutes: entry.DurationMinutes,
		Notes:           entry.Notes,
		SnapshotName:    entry.SnapshotName,
		SnapshotDesc:    entry.SnapshotDesc,
		SnapshotType:    habitScheduleTypeText(entry.SnapshotType),
		SnapshotTarget:  entry.SnapshotTarget,
		UserID:          userID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	snapshotDays, err := weekdaysJSON(entry.SnapshotDays)
	if err != nil {
		return sharedtypes.HabitCompletion{}, err
	}
	model.SnapshotDays = snapshotDays
	if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
		return sharedtypes.HabitCompletion{}, err
	}
	entry.Kind = sharedtypes.NormalizeHabitHistoryKind(entry.Kind)
	entry.StartedAt = &model.StartedAt
	entry.EndedAt = model.EndedAt
	entry.CreatedAt = now
	entry.UpdatedAt = now
	return entry, nil
}

func (r *HabitFocusSessionRepository) ListByHabit(ctx context.Context, habitID int64, userID string) ([]sharedtypes.HabitCompletion, error) {
	rows, err := r.list(ctx, userID, []string{"habits.public_id = ?"}, []any{habitID}, "habit_focus_sessions.started_at DESC")
	if err != nil {
		return nil, err
	}
	out := make([]sharedtypes.HabitCompletion, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toCompletion())
	}
	return out, nil
}

func (r *HabitFocusSessionRepository) ListForDate(ctx context.Context, date, userID string) ([]sharedtypes.HabitCompletion, error) {
	rows, err := r.list(ctx, userID, []string{"habit_focus_sessions.date = ?"}, []any{date}, "habit_focus_sessions.started_at DESC")
	if err != nil {
		return nil, err
	}
	out := make([]sharedtypes.HabitCompletion, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toCompletion())
	}
	return out, nil
}

type habitFocusRow struct {
	PublicID        int64   `bun:"public_id"`
	HabitPublicID   int64   `bun:"habit_public_id"`
	Kind            string  `bun:"kind"`
	Date            string  `bun:"date"`
	StartedAt       string  `bun:"started_at"`
	EndedAt         *string `bun:"ended_at"`
	DurationMinutes *int    `bun:"duration_minutes"`
	Notes           *string `bun:"notes"`
	SnapshotName    *string `bun:"snapshot_name"`
	SnapshotDesc    *string `bun:"snapshot_description"`
	SnapshotType    *string `bun:"snapshot_schedule_type"`
	SnapshotDays    *string `bun:"snapshot_weekdays"`
	SnapshotTarget  *int    `bun:"snapshot_target_minutes"`
	CreatedAt       string  `bun:"created_at"`
	UpdatedAt       string  `bun:"updated_at"`
}

func (r *HabitFocusSessionRepository) list(ctx context.Context, userID string, where []string, args []any, orderExpr string) ([]habitFocusRow, error) {
	var rows []habitFocusRow
	q := r.db.NewSelect().
		TableExpr("habit_focus_sessions").
		Join("INNER JOIN habits ON habits.id = habit_focus_sessions.habit_id").
		ColumnExpr("habit_focus_sessions.public_id").
		ColumnExpr("habits.public_id AS habit_public_id").
		ColumnExpr("habit_focus_sessions.kind").
		ColumnExpr("habit_focus_sessions.date").
		ColumnExpr("habit_focus_sessions.started_at").
		ColumnExpr("habit_focus_sessions.ended_at").
		ColumnExpr("habit_focus_sessions.duration_minutes").
		ColumnExpr("habit_focus_sessions.notes").
		ColumnExpr("habit_focus_sessions.snapshot_name").
		ColumnExpr("habit_focus_sessions.snapshot_description").
		ColumnExpr("habit_focus_sessions.snapshot_schedule_type").
		ColumnExpr("habit_focus_sessions.snapshot_weekdays").
		ColumnExpr("habit_focus_sessions.snapshot_target_minutes").
		ColumnExpr("habit_focus_sessions.created_at").
		ColumnExpr("habit_focus_sessions.updated_at").
		Where("habit_focus_sessions.user_id = ?", userID).
		Where("habit_focus_sessions.deleted_at IS NULL")
	for i, clause := range where {
		if i < len(args) {
			q = q.Where(clause, args[i])
		} else {
			q = q.Where(clause)
		}
	}
	if orderExpr != "" {
		q = q.OrderExpr(orderExpr)
	}
	if err := q.Scan(ctx, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *HabitFocusSessionRepository) listByHabitID(ctx context.Context, habitID int64, userID string) ([]habitFocusRow, error) {
	return r.list(ctx, userID, []string{"habits.public_id = ?"}, []any{habitID}, "habit_focus_sessions.started_at DESC")
}

func (r *HabitFocusSessionRepository) listByDate(ctx context.Context, date, userID string) ([]habitFocusRow, error) {
	return r.list(ctx, userID, []string{"habit_focus_sessions.date = ?"}, []any{date}, "habit_focus_sessions.started_at DESC")
}

func (row habitFocusRow) toCompletion() sharedtypes.HabitCompletion {
	return sharedtypes.HabitCompletion{
		ID:              row.PublicID,
		HabitID:         row.HabitPublicID,
		Kind:            sharedtypes.NormalizeHabitHistoryKind(sharedtypes.HabitHistoryKind(row.Kind)),
		Date:            row.Date,
		Status:          sharedtypes.HabitCompletionStatusCompleted,
		StartedAt:       &row.StartedAt,
		EndedAt:         row.EndedAt,
		DurationMinutes: row.DurationMinutes,
		Notes:           row.Notes,
		SnapshotName:    row.SnapshotName,
		SnapshotDesc:    row.SnapshotDesc,
		SnapshotType:    habitFocusScheduleTypePtr(row.SnapshotType),
		SnapshotDays:    parseWeekdays(row.SnapshotDays),
		SnapshotTarget:  row.SnapshotTarget,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func habitFocusInternalID(publicID int64) string {
	return fmt.Sprintf("habit-focus-%d", publicID)
}

func startedAtOrNow(value *string, now string) string {
	if value != nil && strings.TrimSpace(*value) != "" {
		return strings.TrimSpace(*value)
	}
	if strings.TrimSpace(now) != "" {
		return strings.TrimSpace(now)
	}
	return timeNowFallback()
}

func habitScheduleTypeText(value *sharedtypes.HabitScheduleType) *string {
	if value == nil {
		return nil
	}
	normalized := sharedtypes.NormalizeHabitScheduleType(*value)
	return nullableString(string(normalized))
}

func habitFocusScheduleTypePtr(value *string) *sharedtypes.HabitScheduleType {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	normalized := sharedtypes.NormalizeHabitScheduleType(sharedtypes.HabitScheduleType(strings.TrimSpace(*value)))
	return &normalized
}

func timeNowFallback() string {
	return "1970-01-01T00:00:00Z"
}
