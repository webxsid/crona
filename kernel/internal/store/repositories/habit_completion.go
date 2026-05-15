package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sharedtypes "crona/shared/types"

	storemodels "crona/kernel/internal/store/models"
	"github.com/uptrace/bun"
)

type HabitCompletionRepository struct {
	db *bun.DB
}

func NewHabitCompletionRepository(db *bun.DB) *HabitCompletionRepository {
	return &HabitCompletionRepository{db: db}
}

func (r *HabitCompletionRepository) NextID(ctx context.Context) (int64, error) {
	return nextPublicID(ctx, r.db, "habit_completions")
}

func (r *HabitCompletionRepository) Upsert(ctx context.Context, completion sharedtypes.HabitCompletion, userID string, now string) (sharedtypes.HabitCompletion, error) {
	habitInternalID, err := resolveHabitInternalID(ctx, r.db, completion.HabitID, userID)
	if err != nil {
		return sharedtypes.HabitCompletion{}, err
	}
	if habitInternalID == "" {
		return sharedtypes.HabitCompletion{}, errors.New("habit not found")
	}
	existing, err := r.GetByHabitAndDate(ctx, completion.HabitID, completion.Date, userID)
	if err != nil {
		return sharedtypes.HabitCompletion{}, err
	}
	if existing != nil {
		snapshotSchedule := ""
		if completion.SnapshotType != nil {
			snapshotSchedule = string(*completion.SnapshotType)
		}
		snapshotDays, err := weekdaysJSON(completion.SnapshotDays)
		if err != nil {
			return sharedtypes.HabitCompletion{}, err
		}
		res, err := r.db.NewUpdate().
			Model((*storemodels.HabitCompletionModel)(nil)).
			Where("public_id = ?", existing.ID).
			Where("user_id = ?", userID).
			Set("kind = ?", string(sharedtypes.NormalizeHabitHistoryKind(sharedtypes.HabitHistoryKindCompletion))).
			Set("started_at = NULL").
			Set("ended_at = NULL").
			Set("status = ?", completion.Status).
			Set("duration_minutes = ?", completion.DurationMinutes).
			Set("notes = ?", completion.Notes).
			Set("snapshot_name = ?", completion.SnapshotName).
			Set("snapshot_description = ?", completion.SnapshotDesc).
			Set("snapshot_schedule_type = ?", nullableString(snapshotSchedule)).
			Set("snapshot_weekdays = ?", snapshotDays).
			Set("snapshot_target_minutes = ?", completion.SnapshotTarget).
			Set("updated_at = ?", now).
			Set("deleted_at = NULL").
			Exec(ctx)
		if err != nil {
			return sharedtypes.HabitCompletion{}, err
		}
		if rows, _ := res.RowsAffected(); rows == 0 {
			return sharedtypes.HabitCompletion{}, errors.New("habit completion not found")
		}
		existing.Status = completion.Status
		existing.Kind = sharedtypes.NormalizeHabitHistoryKind(sharedtypes.HabitHistoryKindCompletion)
		existing.StartedAt = nil
		existing.EndedAt = nil
		existing.DurationMinutes = completion.DurationMinutes
		existing.Notes = completion.Notes
		existing.SnapshotName = completion.SnapshotName
		existing.SnapshotDesc = completion.SnapshotDesc
		existing.SnapshotType = completion.SnapshotType
		existing.SnapshotDays = append([]int(nil), completion.SnapshotDays...)
		existing.SnapshotTarget = completion.SnapshotTarget
		existing.UpdatedAt = now
		return *existing, nil
	}
	snapshotSchedule := ""
	if completion.SnapshotType != nil {
		snapshotSchedule = string(*completion.SnapshotType)
	}
	snapshotDays, err := weekdaysJSON(completion.SnapshotDays)
	if err != nil {
		return sharedtypes.HabitCompletion{}, err
	}
	model := storemodels.HabitCompletionModel{
		InternalID:      habitCompletionInternalID(completion.ID),
		PublicID:        completion.ID,
		HabitID:         habitInternalID,
		Kind:            string(sharedtypes.NormalizeHabitHistoryKind(sharedtypes.HabitHistoryKindCompletion)),
		Date:            completion.Date,
		Status:          string(completion.Status),
		StartedAt:       nil,
		EndedAt:         nil,
		DurationMinutes: completion.DurationMinutes,
		Notes:           completion.Notes,
		SnapshotName:    completion.SnapshotName,
		SnapshotDesc:    completion.SnapshotDesc,
		SnapshotType:    nullableString(snapshotSchedule),
		SnapshotDays:    snapshotDays,
		SnapshotTarget:  completion.SnapshotTarget,
		UserID:          userID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
		return sharedtypes.HabitCompletion{}, err
	}
	completion.CreatedAt = now
	completion.UpdatedAt = now
	completion.Kind = sharedtypes.NormalizeHabitHistoryKind(sharedtypes.HabitHistoryKindCompletion)
	completion.StartedAt = nil
	completion.EndedAt = nil
	return completion, nil
}

func (r *HabitCompletionRepository) GetByHabitAndDate(ctx context.Context, habitID int64, date, userID string) (*sharedtypes.HabitCompletion, error) {
	type row struct {
		PublicID        int64   `bun:"public_id"`
		HabitPublicID   int64   `bun:"habit_public_id"`
		Kind            string  `bun:"kind"`
		Date            string  `bun:"date"`
		Status          string  `bun:"status"`
		StartedAt       *string `bun:"started_at"`
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
	var item row
	err := r.db.NewSelect().
		TableExpr("habit_completions").
		Join("INNER JOIN habits ON habits.id = habit_completions.habit_id").
		ColumnExpr("habit_completions.public_id").
		ColumnExpr("habits.public_id AS habit_public_id").
		ColumnExpr("habit_completions.kind").
		ColumnExpr("habit_completions.date").
		ColumnExpr("habit_completions.status").
		ColumnExpr("habit_completions.started_at").
		ColumnExpr("habit_completions.ended_at").
		ColumnExpr("habit_completions.duration_minutes").
		ColumnExpr("habit_completions.notes").
		ColumnExpr("habit_completions.snapshot_name").
		ColumnExpr("habit_completions.snapshot_description").
		ColumnExpr("habit_completions.snapshot_schedule_type").
		ColumnExpr("habit_completions.snapshot_weekdays").
		ColumnExpr("habit_completions.snapshot_target_minutes").
		ColumnExpr("habit_completions.created_at").
		ColumnExpr("habit_completions.updated_at").
		Where("habits.public_id = ?", habitID).
		Where("habit_completions.date = ?", date).
		Where("habit_completions.user_id = ?", userID).
		Where("habit_completions.deleted_at IS NULL").
		Limit(1).
		Scan(ctx, &item)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var snapshotType *sharedtypes.HabitScheduleType
	if item.SnapshotType != nil && *item.SnapshotType != "" {
		value := sharedtypes.NormalizeHabitScheduleType(sharedtypes.HabitScheduleType(*item.SnapshotType))
		snapshotType = &value
	}
	return &sharedtypes.HabitCompletion{
		ID:              item.PublicID,
		HabitID:         item.HabitPublicID,
		Kind:            sharedtypes.NormalizeHabitHistoryKind(sharedtypes.HabitHistoryKind(item.Kind)),
		Date:            item.Date,
		Status:          sharedtypes.NormalizeHabitCompletionStatus(sharedtypes.HabitCompletionStatus(item.Status)),
		StartedAt:       item.StartedAt,
		EndedAt:         item.EndedAt,
		DurationMinutes: item.DurationMinutes,
		Notes:           item.Notes,
		SnapshotName:    item.SnapshotName,
		SnapshotDesc:    item.SnapshotDesc,
		SnapshotType:    snapshotType,
		SnapshotDays:    parseWeekdays(item.SnapshotDays),
		SnapshotTarget:  item.SnapshotTarget,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}, nil
}

func (r *HabitCompletionRepository) ListByHabit(ctx context.Context, habitID int64, userID string) ([]sharedtypes.HabitCompletion, error) {
	type row struct {
		PublicID        int64   `bun:"public_id"`
		HabitPublicID   int64   `bun:"habit_public_id"`
		Kind            string  `bun:"kind"`
		Date            string  `bun:"date"`
		Status          string  `bun:"status"`
		StartedAt       *string `bun:"started_at"`
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
	var rows []row
	if err := r.db.NewSelect().
		TableExpr("habit_completions").
		Join("INNER JOIN habits ON habits.id = habit_completions.habit_id").
		ColumnExpr("habit_completions.public_id").
		ColumnExpr("habits.public_id AS habit_public_id").
		ColumnExpr("habit_completions.kind").
		ColumnExpr("habit_completions.date").
		ColumnExpr("habit_completions.status").
		ColumnExpr("habit_completions.started_at").
		ColumnExpr("habit_completions.ended_at").
		ColumnExpr("habit_completions.duration_minutes").
		ColumnExpr("habit_completions.notes").
		ColumnExpr("habit_completions.snapshot_name").
		ColumnExpr("habit_completions.snapshot_description").
		ColumnExpr("habit_completions.snapshot_schedule_type").
		ColumnExpr("habit_completions.snapshot_weekdays").
		ColumnExpr("habit_completions.snapshot_target_minutes").
		ColumnExpr("habit_completions.created_at").
		ColumnExpr("habit_completions.updated_at").
		Where("habits.public_id = ?", habitID).
		Where("habit_completions.user_id = ?", userID).
		Where("habit_completions.deleted_at IS NULL").
		OrderExpr("habit_completions.date DESC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.HabitCompletion, 0, len(rows))
	for _, row := range rows {
		var snapshotType *sharedtypes.HabitScheduleType
		if row.SnapshotType != nil && *row.SnapshotType != "" {
			value := sharedtypes.NormalizeHabitScheduleType(sharedtypes.HabitScheduleType(*row.SnapshotType))
			snapshotType = &value
		}
		out = append(out, sharedtypes.HabitCompletion{
			ID:              row.PublicID,
			HabitID:         row.HabitPublicID,
			Kind:            sharedtypes.NormalizeHabitHistoryKind(sharedtypes.HabitHistoryKind(row.Kind)),
			Date:            row.Date,
			Status:          sharedtypes.NormalizeHabitCompletionStatus(sharedtypes.HabitCompletionStatus(row.Status)),
			StartedAt:       row.StartedAt,
			EndedAt:         row.EndedAt,
			DurationMinutes: row.DurationMinutes,
			Notes:           row.Notes,
			SnapshotName:    row.SnapshotName,
			SnapshotDesc:    row.SnapshotDesc,
			SnapshotType:    snapshotType,
			SnapshotDays:    parseWeekdays(row.SnapshotDays),
			SnapshotTarget:  row.SnapshotTarget,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *HabitCompletionRepository) ListHistory(ctx context.Context, userID string, repoID, streamID *int64) ([]sharedtypes.HabitCompletion, error) {
	type row struct {
		PublicID        int64   `bun:"public_id"`
		HabitPublicID   int64   `bun:"habit_public_id"`
		HabitName       string  `bun:"habit_name"`
		RepoName        string  `bun:"repo_name"`
		StreamName      string  `bun:"stream_name"`
		Kind            string  `bun:"kind"`
		Date            string  `bun:"date"`
		Status          string  `bun:"status"`
		StartedAt       *string `bun:"started_at"`
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
	q := r.db.NewSelect().
		TableExpr("habit_completions").
		Join("INNER JOIN habits ON habits.id = habit_completions.habit_id").
		Join("INNER JOIN streams ON streams.id = habits.stream_id").
		Join("INNER JOIN repos ON repos.id = streams.repo_id").
		ColumnExpr("habit_completions.public_id").
		ColumnExpr("habits.public_id AS habit_public_id").
		ColumnExpr("habits.name AS habit_name").
		ColumnExpr("repos.name AS repo_name").
		ColumnExpr("streams.name AS stream_name").
		ColumnExpr("habit_completions.kind").
		ColumnExpr("habit_completions.date").
		ColumnExpr("habit_completions.status").
		ColumnExpr("habit_completions.started_at").
		ColumnExpr("habit_completions.ended_at").
		ColumnExpr("habit_completions.duration_minutes").
		ColumnExpr("habit_completions.notes").
		ColumnExpr("habit_completions.snapshot_name").
		ColumnExpr("habit_completions.snapshot_description").
		ColumnExpr("habit_completions.snapshot_schedule_type").
		ColumnExpr("habit_completions.snapshot_weekdays").
		ColumnExpr("habit_completions.snapshot_target_minutes").
		ColumnExpr("habit_completions.created_at").
		ColumnExpr("habit_completions.updated_at").
		Where("habit_completions.user_id = ?", userID).
		Where("habit_completions.deleted_at IS NULL").
		OrderExpr("habit_completions.date DESC, repos.name ASC, streams.name ASC, habits.name ASC, habit_completions.created_at DESC")
	if repoID != nil {
		q = q.Where("repos.public_id = ?", *repoID)
	}
	if streamID != nil {
		q = q.Where("streams.public_id = ?", *streamID)
	}
	var rows []row
	if err := q.Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.HabitCompletion, 0, len(rows))
	for _, row := range rows {
		var snapshotType *sharedtypes.HabitScheduleType
		if row.SnapshotType != nil && *row.SnapshotType != "" {
			value := sharedtypes.NormalizeHabitScheduleType(sharedtypes.HabitScheduleType(*row.SnapshotType))
			snapshotType = &value
		}
		out = append(out, sharedtypes.HabitCompletion{
			ID:              row.PublicID,
			HabitID:         row.HabitPublicID,
			HabitName:       row.HabitName,
			RepoName:        row.RepoName,
			StreamName:      row.StreamName,
			Kind:            sharedtypes.NormalizeHabitHistoryKind(sharedtypes.HabitHistoryKind(row.Kind)),
			Date:            row.Date,
			Status:          sharedtypes.NormalizeHabitCompletionStatus(sharedtypes.HabitCompletionStatus(row.Status)),
			StartedAt:       row.StartedAt,
			EndedAt:         row.EndedAt,
			DurationMinutes: row.DurationMinutes,
			Notes:           row.Notes,
			SnapshotName:    row.SnapshotName,
			SnapshotDesc:    row.SnapshotDesc,
			SnapshotType:    snapshotType,
			SnapshotDays:    parseWeekdays(row.SnapshotDays),
			SnapshotTarget:  row.SnapshotTarget,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *HabitCompletionRepository) EarliestDate(ctx context.Context, userID string, throughDate string) (*string, error) {
	type row struct {
		Date string `bun:"date"`
	}
	var item row
	err := r.db.NewSelect().
		TableExpr("habit_completions").
		ColumnExpr("date").
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Where("date <= ?", throughDate).
		OrderExpr("date ASC").
		Limit(1).
		Scan(ctx, &item)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if item.Date == "" {
		return nil, nil
	}
	return &item.Date, nil
}

func (r *HabitCompletionRepository) ListForDate(ctx context.Context, date, userID string) ([]sharedtypes.HabitCompletion, error) {
	type row struct {
		PublicID        int64   `bun:"public_id"`
		HabitPublicID   int64   `bun:"habit_public_id"`
		Kind            string  `bun:"kind"`
		Date            string  `bun:"date"`
		Status          string  `bun:"status"`
		StartedAt       *string `bun:"started_at"`
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
	var rows []row
	if err := r.db.NewSelect().
		TableExpr("habit_completions").
		Join("INNER JOIN habits ON habits.id = habit_completions.habit_id").
		ColumnExpr("habit_completions.public_id").
		ColumnExpr("habits.public_id AS habit_public_id").
		ColumnExpr("habit_completions.kind").
		ColumnExpr("habit_completions.date").
		ColumnExpr("habit_completions.status").
		ColumnExpr("habit_completions.started_at").
		ColumnExpr("habit_completions.ended_at").
		ColumnExpr("habit_completions.duration_minutes").
		ColumnExpr("habit_completions.notes").
		ColumnExpr("habit_completions.snapshot_name").
		ColumnExpr("habit_completions.snapshot_description").
		ColumnExpr("habit_completions.snapshot_schedule_type").
		ColumnExpr("habit_completions.snapshot_weekdays").
		ColumnExpr("habit_completions.snapshot_target_minutes").
		ColumnExpr("habit_completions.created_at").
		ColumnExpr("habit_completions.updated_at").
		Where("habit_completions.date = ?", date).
		Where("habit_completions.user_id = ?", userID).
		Where("habit_completions.deleted_at IS NULL").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.HabitCompletion, 0, len(rows))
	for _, row := range rows {
		var snapshotType *sharedtypes.HabitScheduleType
		if row.SnapshotType != nil && *row.SnapshotType != "" {
			value := sharedtypes.NormalizeHabitScheduleType(sharedtypes.HabitScheduleType(*row.SnapshotType))
			snapshotType = &value
		}
		out = append(out, sharedtypes.HabitCompletion{
			ID:              row.PublicID,
			HabitID:         row.HabitPublicID,
			Kind:            sharedtypes.NormalizeHabitHistoryKind(sharedtypes.HabitHistoryKind(row.Kind)),
			Date:            row.Date,
			Status:          sharedtypes.NormalizeHabitCompletionStatus(sharedtypes.HabitCompletionStatus(row.Status)),
			StartedAt:       row.StartedAt,
			EndedAt:         row.EndedAt,
			DurationMinutes: row.DurationMinutes,
			Notes:           row.Notes,
			SnapshotName:    row.SnapshotName,
			SnapshotDesc:    row.SnapshotDesc,
			SnapshotType:    snapshotType,
			SnapshotDays:    parseWeekdays(row.SnapshotDays),
			SnapshotTarget:  row.SnapshotTarget,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *HabitCompletionRepository) DeleteByHabitAndDate(ctx context.Context, habitID int64, date, userID, now string) error {
	habitInternalID, err := resolveHabitInternalID(ctx, r.db, habitID, userID)
	if err != nil {
		return err
	}
	if habitInternalID == "" {
		return errors.New("habit not found")
	}
	res, err := r.db.NewUpdate().
		Model((*storemodels.HabitCompletionModel)(nil)).
		Where("habit_id = ?", habitInternalID).
		Where("date = ?", date).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("habit completion not found")
	}
	return nil
}

func habitCompletionInternalID(publicID int64) string {
	return fmt.Sprintf("habit-completion-%d", publicID)
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
