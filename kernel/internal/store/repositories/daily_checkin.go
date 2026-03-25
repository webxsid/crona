package repositories

import (
	"context"
	"database/sql"
	"errors"

	sharedtypes "crona/shared/types"

	storemodels "crona/kernel/internal/store/models"
	"github.com/uptrace/bun"
)

type DailyCheckInRepository struct {
	db *bun.DB
}

func NewDailyCheckInRepository(db *bun.DB) *DailyCheckInRepository {
	return &DailyCheckInRepository{db: db}
}

func (r *DailyCheckInRepository) GetByDate(ctx context.Context, userID string, date string) (*sharedtypes.DailyCheckIn, error) {
	var model storemodels.DailyCheckInModel
	err := r.db.NewSelect().
		Model(&model).
		Where("user_id = ?", userID).
		Where("date = ?", date).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return dailyCheckInFromModel(model), nil
}

func (r *DailyCheckInRepository) Upsert(ctx context.Context, checkIn sharedtypes.DailyCheckIn, userID string, deviceID string, now string) (*sharedtypes.DailyCheckIn, error) {
	model := storemodels.DailyCheckInModel{
		UserID:            userID,
		DeviceID:          deviceID,
		Date:              checkIn.Date,
		Mood:              checkIn.Mood,
		Energy:            checkIn.Energy,
		SleepHours:        checkIn.SleepHours,
		SleepScore:        checkIn.SleepScore,
		ScreenTimeMinutes: checkIn.ScreenTimeMinutes,
		Notes:             checkIn.Notes,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	_, err := r.db.NewInsert().
		Model(&model).
		On("CONFLICT (user_id, date) DO UPDATE").
		Set("device_id = EXCLUDED.device_id").
		Set("mood = EXCLUDED.mood").
		Set("energy = EXCLUDED.energy").
		Set("sleep_hours = EXCLUDED.sleep_hours").
		Set("sleep_score = EXCLUDED.sleep_score").
		Set("screen_time_minutes = EXCLUDED.screen_time_minutes").
		Set("notes = EXCLUDED.notes").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return r.GetByDate(ctx, userID, checkIn.Date)
}

func (r *DailyCheckInRepository) DeleteByDate(ctx context.Context, userID string, date string) error {
	res, err := r.db.NewDelete().
		Model((*storemodels.DailyCheckInModel)(nil)).
		Where("user_id = ?", userID).
		Where("date = ?", date).
		Exec(ctx)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *DailyCheckInRepository) ListRange(ctx context.Context, userID string, start string, end string) ([]sharedtypes.DailyCheckIn, error) {
	var models []storemodels.DailyCheckInModel
	if err := r.db.NewSelect().
		Model(&models).
		Where("user_id = ?", userID).
		Where("date >= ?", start).
		Where("date <= ?", end).
		OrderExpr("date ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	out := make([]sharedtypes.DailyCheckIn, 0, len(models))
	for _, model := range models {
		out = append(out, *dailyCheckInFromModel(model))
	}
	return out, nil
}

func dailyCheckInFromModel(model storemodels.DailyCheckInModel) *sharedtypes.DailyCheckIn {
	return &sharedtypes.DailyCheckIn{
		Date:              model.Date,
		Mood:              model.Mood,
		Energy:            model.Energy,
		SleepHours:        model.SleepHours,
		SleepScore:        model.SleepScore,
		ScreenTimeMinutes: model.ScreenTimeMinutes,
		Notes:             model.Notes,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}
}
