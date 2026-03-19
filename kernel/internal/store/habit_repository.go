package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sharedtypes "crona/shared/types"

	"github.com/uptrace/bun"
)

type HabitRepository struct {
	db *bun.DB
}

func NewHabitRepository(db *bun.DB) *HabitRepository {
	return &HabitRepository{db: db}
}

func (r *HabitRepository) NextID(ctx context.Context) (int64, error) {
	return nextPublicID(ctx, r.db, "habits")
}

func (r *HabitRepository) Create(ctx context.Context, habit sharedtypes.Habit, userID string, now string) (sharedtypes.Habit, error) {
	streamInternalID, err := resolveStreamInternalID(ctx, r.db, habit.StreamID, userID)
	if err != nil {
		return sharedtypes.Habit{}, err
	}
	if streamInternalID == "" {
		return sharedtypes.Habit{}, errors.New("stream not found")
	}
	weekdays, err := weekdaysJSON(habit.Weekdays)
	if err != nil {
		return sharedtypes.Habit{}, err
	}
	model := HabitModel{
		InternalID:    habitInternalID(habit.ID),
		PublicID:      habit.ID,
		StreamID:      streamInternalID,
		Name:          habit.Name,
		Description:   habit.Description,
		ScheduleType:  string(sharedtypes.NormalizeHabitScheduleType(habit.ScheduleType)),
		Weekdays:      weekdays,
		TargetMinutes: habit.TargetMinutes,
		Active:        habit.Active,
		UserID:        userID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
		return sharedtypes.Habit{}, err
	}
	return habit, nil
}

func (r *HabitRepository) ListByStream(ctx context.Context, streamID int64, userID string) ([]sharedtypes.Habit, error) {
	type row struct {
		PublicID       int64   `bun:"public_id"`
		StreamPublicID int64   `bun:"stream_public_id"`
		Name           string  `bun:"name"`
		Description    *string `bun:"description"`
		ScheduleType   string  `bun:"schedule_type"`
		Weekdays       *string `bun:"weekdays"`
		TargetMinutes  *int    `bun:"target_minutes"`
		Active         bool    `bun:"active"`
	}
	var rows []row
	if err := r.db.NewSelect().
		TableExpr("habits").
		Join("INNER JOIN streams ON streams.id = habits.stream_id").
		ColumnExpr("habits.public_id").
		ColumnExpr("streams.public_id AS stream_public_id").
		ColumnExpr("habits.name").
		ColumnExpr("habits.description").
		ColumnExpr("habits.schedule_type").
		ColumnExpr("habits.weekdays").
		ColumnExpr("habits.target_minutes").
		ColumnExpr("habits.active").
		Where("streams.public_id = ?", streamID).
		Where("habits.user_id = ?", userID).
		Where("habits.deleted_at IS NULL").
		Where("streams.deleted_at IS NULL").
		OrderExpr("habits.created_at ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.Habit, 0, len(rows))
	for _, row := range rows {
		out = append(out, sharedtypes.Habit{
			ID:            row.PublicID,
			StreamID:      row.StreamPublicID,
			Name:          row.Name,
			Description:   row.Description,
			ScheduleType:  sharedtypes.NormalizeHabitScheduleType(sharedtypes.HabitScheduleType(row.ScheduleType)),
			Weekdays:      parseWeekdays(row.Weekdays),
			TargetMinutes: row.TargetMinutes,
			Active:        row.Active,
		})
	}
	return out, nil
}

func (r *HabitRepository) GetByID(ctx context.Context, habitID int64, userID string) (*sharedtypes.Habit, error) {
	type row struct {
		PublicID       int64   `bun:"public_id"`
		StreamPublicID int64   `bun:"stream_public_id"`
		Name           string  `bun:"name"`
		Description    *string `bun:"description"`
		ScheduleType   string  `bun:"schedule_type"`
		Weekdays       *string `bun:"weekdays"`
		TargetMinutes  *int    `bun:"target_minutes"`
		Active         bool    `bun:"active"`
	}
	var item row
	err := r.db.NewSelect().
		TableExpr("habits").
		Join("INNER JOIN streams ON streams.id = habits.stream_id").
		ColumnExpr("habits.public_id").
		ColumnExpr("streams.public_id AS stream_public_id").
		ColumnExpr("habits.name").
		ColumnExpr("habits.description").
		ColumnExpr("habits.schedule_type").
		ColumnExpr("habits.weekdays").
		ColumnExpr("habits.target_minutes").
		ColumnExpr("habits.active").
		Where("habits.public_id = ?", habitID).
		Where("habits.user_id = ?", userID).
		Where("habits.deleted_at IS NULL").
		Where("streams.deleted_at IS NULL").
		Limit(1).
		Scan(ctx, &item)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sharedtypes.Habit{
		ID:            item.PublicID,
		StreamID:      item.StreamPublicID,
		Name:          item.Name,
		Description:   item.Description,
		ScheduleType:  sharedtypes.NormalizeHabitScheduleType(sharedtypes.HabitScheduleType(item.ScheduleType)),
		Weekdays:      parseWeekdays(item.Weekdays),
		TargetMinutes: item.TargetMinutes,
		Active:        item.Active,
	}, nil
}

func (r *HabitRepository) ListDueWithMeta(ctx context.Context, date string, userID string) ([]sharedtypes.HabitWithMeta, error) {
	type row struct {
		PublicID       int64   `bun:"public_id"`
		StreamPublicID int64   `bun:"stream_public_id"`
		Name           string  `bun:"name"`
		Description    *string `bun:"description"`
		ScheduleType   string  `bun:"schedule_type"`
		Weekdays       *string `bun:"weekdays"`
		TargetMinutes  *int    `bun:"target_minutes"`
		Active         bool    `bun:"active"`
		StreamName     string  `bun:"stream_name"`
		RepoPublicID   int64   `bun:"repo_public_id"`
		RepoName       string  `bun:"repo_name"`
	}
	var rows []row
	if err := r.db.NewSelect().
		TableExpr("habits").
		Join("INNER JOIN streams ON streams.id = habits.stream_id").
		Join("INNER JOIN repos ON repos.id = streams.repo_id").
		ColumnExpr("habits.public_id").
		ColumnExpr("streams.public_id AS stream_public_id").
		ColumnExpr("habits.name").
		ColumnExpr("habits.description").
		ColumnExpr("habits.schedule_type").
		ColumnExpr("habits.weekdays").
		ColumnExpr("habits.target_minutes").
		ColumnExpr("habits.active").
		ColumnExpr("streams.name AS stream_name").
		ColumnExpr("repos.public_id AS repo_public_id").
		ColumnExpr("repos.name AS repo_name").
		Where("habits.user_id = ?", userID).
		Where("habits.deleted_at IS NULL").
		Where("streams.deleted_at IS NULL").
		Where("repos.deleted_at IS NULL").
		OrderExpr("habits.created_at ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.HabitWithMeta, 0, len(rows))
	for _, row := range rows {
		habit := sharedtypes.HabitWithMeta{
			Habit: sharedtypes.Habit{
				ID:            row.PublicID,
				StreamID:      row.StreamPublicID,
				Name:          row.Name,
				Description:   row.Description,
				ScheduleType:  sharedtypes.NormalizeHabitScheduleType(sharedtypes.HabitScheduleType(row.ScheduleType)),
				Weekdays:      parseWeekdays(row.Weekdays),
				TargetMinutes: row.TargetMinutes,
				Active:        row.Active,
			},
			RepoID:     row.RepoPublicID,
			RepoName:   row.RepoName,
			StreamName: row.StreamName,
		}
		if habit.Active && HabitMatchesDate(habit.Habit, date) {
			out = append(out, habit)
		}
	}
	return out, nil
}

func (r *HabitRepository) Update(ctx context.Context, habitID int64, userID string, now string, updates struct {
	Name          Patch[string]
	Description   Patch[string]
	ScheduleType  Patch[string]
	Weekdays      []int
	WeekdaysSet   bool
	TargetMinutes Patch[int]
	Active        *bool
}) (*sharedtypes.Habit, error) {
	q := r.db.NewUpdate().
		Model((*HabitModel)(nil)).
		Where("public_id = ?", habitID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("updated_at = ?", now)
	if updates.Name.Set && updates.Name.Value != nil {
		q = q.Set("name = ?", *updates.Name.Value)
	}
	if updates.Description.Set {
		if updates.Description.Value == nil {
			q = q.Set("description = NULL")
		} else {
			q = q.Set("description = ?", *updates.Description.Value)
		}
	}
	if updates.ScheduleType.Set && updates.ScheduleType.Value != nil {
		q = q.Set("schedule_type = ?", *updates.ScheduleType.Value)
	}
	if updates.WeekdaysSet {
		raw, err := weekdaysJSON(updates.Weekdays)
		if err != nil {
			return nil, err
		}
		if raw == nil {
			q = q.Set("weekdays = NULL")
		} else {
			q = q.Set("weekdays = ?", *raw)
		}
	}
	if updates.TargetMinutes.Set {
		if updates.TargetMinutes.Value == nil {
			q = q.Set("target_minutes = NULL")
		} else {
			q = q.Set("target_minutes = ?", *updates.TargetMinutes.Value)
		}
	}
	if updates.Active != nil {
		q = q.Set("active = ?", *updates.Active)
	}
	res, err := q.Exec(ctx)
	if err != nil {
		return nil, err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, errors.New("habit not found or already deleted")
	}
	return r.GetByID(ctx, habitID, userID)
}

func (r *HabitRepository) SoftDelete(ctx context.Context, habitID int64, userID string, now string) error {
	res, err := r.db.NewUpdate().
		Model((*HabitModel)(nil)).
		Where("public_id = ?", habitID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("habit not found or already deleted")
	}
	return nil
}

func (r *HabitRepository) SoftDeleteByStream(ctx context.Context, streamID int64, userID string, now string) error {
	streamInternalID, err := resolveStreamInternalID(ctx, r.db, streamID, userID)
	if err != nil || streamInternalID == "" {
		return err
	}
	_, err = r.db.NewUpdate().
		Model((*HabitModel)(nil)).
		Where("stream_id = ?", streamInternalID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Exec(ctx)
	return err
}

func (r *HabitRepository) SoftDeleteByRepo(ctx context.Context, repoID int64, userID string, now string) error {
	repoInternalID, err := resolveRepoInternalID(ctx, r.db, repoID, userID)
	if err != nil || repoInternalID == "" {
		return err
	}
	_, err = r.db.NewUpdate().
		Model((*HabitModel)(nil)).
		Where("stream_id IN (SELECT id FROM streams WHERE repo_id = ? AND user_id = ?)", repoInternalID, userID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Exec(ctx)
	return err
}

func (r *HabitRepository) RestoreDeletedByStream(ctx context.Context, streamID int64, userID string, now string) error {
	streamInternalID, err := resolveStreamInternalID(ctx, r.db, streamID, userID)
	if err != nil || streamInternalID == "" {
		return err
	}
	_, err = r.db.NewUpdate().
		Model((*HabitModel)(nil)).
		Where("stream_id = ?", streamInternalID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NOT NULL").
		Set("deleted_at = NULL").
		Set("updated_at = ?", now).
		Exec(ctx)
	return err
}

func (r *HabitRepository) RestoreDeletedByRepo(ctx context.Context, repoID int64, userID string, now string) error {
	repoInternalID, err := resolveRepoInternalID(ctx, r.db, repoID, userID)
	if err != nil || repoInternalID == "" {
		return err
	}
	_, err = r.db.NewUpdate().
		Model((*HabitModel)(nil)).
		Where("stream_id IN (SELECT id FROM streams WHERE repo_id = ? AND user_id = ?)", repoInternalID, userID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NOT NULL").
		Set("deleted_at = NULL").
		Set("updated_at = ?", now).
		Exec(ctx)
	return err
}

func habitInternalID(publicID int64) string {
	return fmt.Sprintf("habit-%d", publicID)
}

func weekdaysJSON(days []int) (*string, error) {
	if len(days) == 0 {
		return nil, nil
	}
	raw, err := json.Marshal(days)
	if err != nil {
		return nil, err
	}
	value := string(raw)
	return &value, nil
}

func parseWeekdays(raw *string) []int {
	if raw == nil || *raw == "" {
		return nil
	}
	var out []int
	if err := json.Unmarshal([]byte(*raw), &out); err != nil {
		return nil
	}
	return out
}

func HabitMatchesDate(habit sharedtypes.Habit, date string) bool {
	if !habit.Active {
		return false
	}
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	switch sharedtypes.NormalizeHabitScheduleType(habit.ScheduleType) {
	case sharedtypes.HabitScheduleWeekdays:
		wd := int(parsed.Weekday())
		return wd >= 1 && wd <= 5
	case sharedtypes.HabitScheduleWeekly:
		if len(habit.Weekdays) == 0 {
			return false
		}
		wd := int(parsed.Weekday())
		for _, day := range habit.Weekdays {
			if day == wd {
				return true
			}
		}
		return false
	default:
		return true
	}
}
