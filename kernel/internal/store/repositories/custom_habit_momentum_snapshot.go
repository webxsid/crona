package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	storemodels "crona/kernel/internal/store/models"

	"github.com/uptrace/bun"
)

type CustomHabitMomentumSnapshotRepository struct {
	db *bun.DB
}

type CustomHabitMomentumSnapshotRow struct {
	Date        string
	SummaryJSON string
	StateJSON   string
	CreatedAt   string
	UpdatedAt   string
}

func NewCustomHabitMomentumSnapshotRepository(db *bun.DB) *CustomHabitMomentumSnapshotRepository {
	return &CustomHabitMomentumSnapshotRepository{db: db}
}

func (r *CustomHabitMomentumSnapshotRepository) IsEmpty(
	ctx context.Context,
	userID string,
) (bool, error) {
	var count int
	if err := r.db.NewSelect().
		Model((*storemodels.CustomHabitMomentumSnapshotModel)(nil)).
		ColumnExpr("COUNT(*)").
		Where("user_id = ?", userID).
		Scan(ctx, &count); err != nil {
		return false, err
	}
	return count == 0, nil
}

func (r *CustomHabitMomentumSnapshotRepository) GetByDate(
	ctx context.Context,
	userID string,
	date string,
) (*CustomHabitMomentumSnapshotRow, error) {
	return r.selectOne(ctx, userID, "date = ?", date)
}

func (r *CustomHabitMomentumSnapshotRepository) GetLatestBeforeDate(
	ctx context.Context,
	userID string,
	date string,
) (*CustomHabitMomentumSnapshotRow, error) {
	return r.selectOne(ctx, userID, "date < ?", date)
}

func (r *CustomHabitMomentumSnapshotRepository) Upsert(
	ctx context.Context,
	userID string,
	date string,
	summaryJSON string,
	stateJSON string,
	now string,
) error {
	row := storemodels.CustomHabitMomentumSnapshotModel{
		ID:          fmt.Sprintf("momentum-snapshot:%s:%s", userID, date),
		UserID:      userID,
		Date:        date,
		SummaryJSON: summaryJSON,
		StateJSON:   stateJSON,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err := r.db.NewInsert().
		Model(&row).
		On("CONFLICT(user_id, date) DO UPDATE").
		Set("summary_json = excluded.summary_json").
		Set("state_json = excluded.state_json").
		Set("updated_at = excluded.updated_at").
		Exec(ctx)
	return err
}

func (r *CustomHabitMomentumSnapshotRepository) DeleteFromDate(
	ctx context.Context,
	userID string,
	date string,
) error {
	_, err := r.db.NewDelete().
		Model((*storemodels.CustomHabitMomentumSnapshotModel)(nil)).
		Where("user_id = ?", userID).
		Where("date >= ?", date).
		Exec(ctx)
	return err
}

func (r *CustomHabitMomentumSnapshotRepository) selectOne(
	ctx context.Context,
	userID string,
	whereExpr string,
	args ...any,
) (*CustomHabitMomentumSnapshotRow, error) {
	type row struct {
		Date        string `bun:"date"`
		SummaryJSON string `bun:"summary_json"`
		StateJSON   string `bun:"state_json"`
		CreatedAt   string `bun:"created_at"`
		UpdatedAt   string `bun:"updated_at"`
	}
	var item row
	q := r.db.NewSelect().
		TableExpr("custom_habit_momentum_snapshots").
		ColumnExpr("date").
		ColumnExpr("summary_json").
		ColumnExpr("state_json").
		ColumnExpr("created_at").
		ColumnExpr("updated_at").
		Where("user_id = ?", userID).
		OrderExpr("date DESC").
		Limit(1)
	if whereExpr != "" {
		q = q.Where(whereExpr, args...)
	}
	if err := q.Scan(ctx, &item); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &CustomHabitMomentumSnapshotRow{
		Date:        item.Date,
		SummaryJSON: item.SummaryJSON,
		StateJSON:   item.StateJSON,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}, nil
}
