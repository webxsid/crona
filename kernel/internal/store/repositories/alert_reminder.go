package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	storemodels "crona/kernel/internal/store/models"
	sharedtypes "crona/shared/types"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AlertReminderRepository struct {
	db *bun.DB
}

func NewAlertReminderRepository(db *bun.DB) *AlertReminderRepository {
	return &AlertReminderRepository{db: db}
}

func (r *AlertReminderRepository) List(ctx context.Context, userID string) ([]sharedtypes.AlertReminder, error) {
	var models []storemodels.AlertReminderModel
	if err := r.db.NewSelect().
		Model(&models).
		Where("user_id = ?", userID).
		OrderExpr("time_hhmm ASC, created_at ASC").
		Scan(ctx); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.AlertReminder, 0, len(models))
	for _, model := range models {
		out = append(out, alertReminderFromModel(model))
	}
	return out, nil
}

func (r *AlertReminderRepository) GetByID(ctx context.Context, userID string, id string) (*sharedtypes.AlertReminder, error) {
	var model storemodels.AlertReminderModel
	if err := r.db.NewSelect().
		Model(&model).
		Where("user_id = ?", userID).
		Where("id = ?", id).
		Limit(1).
		Scan(ctx); err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	reminder := alertReminderFromModel(model)
	return &reminder, nil
}

func (r *AlertReminderRepository) Create(ctx context.Context, userID, deviceID string, reminder sharedtypes.AlertReminder, now string) (*sharedtypes.AlertReminder, error) {
	weekdays, err := encodeWeekdays(reminder.Weekdays)
	if err != nil {
		return nil, err
	}
	model := storemodels.AlertReminderModel{
		ID:           uuid.NewString(),
		UserID:       userID,
		DeviceID:     deviceID,
		Kind:         string(sharedtypes.NormalizeAlertReminderKind(reminder.Kind)),
		Enabled:      reminder.Enabled,
		ScheduleType: string(sharedtypes.NormalizeAlertReminderScheduleType(reminder.ScheduleType)),
		Weekdays:     weekdays,
		TimeHHMM:     strings.TrimSpace(reminder.TimeHHMM),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
		return nil, err
	}
	out := alertReminderFromModel(model)
	return &out, nil
}

func (r *AlertReminderRepository) Update(ctx context.Context, userID string, reminder sharedtypes.AlertReminder, now string) (*sharedtypes.AlertReminder, error) {
	weekdays, err := encodeWeekdays(reminder.Weekdays)
	if err != nil {
		return nil, err
	}
	res, err := r.db.NewUpdate().
		Model((*storemodels.AlertReminderModel)(nil)).
		Set("kind = ?", string(sharedtypes.NormalizeAlertReminderKind(reminder.Kind))).
		Set("enabled = ?", reminder.Enabled).
		Set("schedule_type = ?", string(sharedtypes.NormalizeAlertReminderScheduleType(reminder.ScheduleType))).
		Set("weekdays = ?", weekdays).
		Set("time_hhmm = ?", strings.TrimSpace(reminder.TimeHHMM)).
		Set("updated_at = ?", now).
		Where("user_id = ?", userID).
		Where("id = ?", reminder.ID).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return nil, nil
	}
	return r.GetByID(ctx, userID, reminder.ID)
}

func (r *AlertReminderRepository) SetEnabled(ctx context.Context, userID, id string, enabled bool, now string) (*sharedtypes.AlertReminder, error) {
	res, err := r.db.NewUpdate().
		Model((*storemodels.AlertReminderModel)(nil)).
		Set("enabled = ?", enabled).
		Set("updated_at = ?", now).
		Where("user_id = ?", userID).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return nil, nil
	}
	return r.GetByID(ctx, userID, id)
}

func (r *AlertReminderRepository) Delete(ctx context.Context, userID, id string) error {
	res, err := r.db.NewDelete().
		Model((*storemodels.AlertReminderModel)(nil)).
		Where("user_id = ?", userID).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("alert reminder not found")
	}
	return nil
}

func alertReminderFromModel(model storemodels.AlertReminderModel) sharedtypes.AlertReminder {
	return sharedtypes.AlertReminder{
		ID:           model.ID,
		Kind:         sharedtypes.NormalizeAlertReminderKind(sharedtypes.AlertReminderKind(model.Kind)),
		Enabled:      model.Enabled,
		ScheduleType: sharedtypes.NormalizeAlertReminderScheduleType(sharedtypes.AlertReminderScheduleType(model.ScheduleType)),
		Weekdays:     decodeWeekdays(model.Weekdays),
		TimeHHMM:     strings.TrimSpace(model.TimeHHMM),
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func encodeWeekdays(values []int) (string, error) {
	if len(values) == 0 {
		return "[]", nil
	}
	unique := make([]int, 0, len(values))
	for _, value := range values {
		if value < 0 || value > 6 || slices.Contains(unique, value) {
			continue
		}
		unique = append(unique, value)
	}
	slices.Sort(unique)
	raw, err := json.Marshal(unique)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func decodeWeekdays(raw string) []int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var out []int
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	filtered := make([]int, 0, len(out))
	for _, value := range out {
		if value >= 0 && value <= 6 && !slices.Contains(filtered, value) {
			filtered = append(filtered, value)
		}
	}
	slices.Sort(filtered)
	return filtered
}

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || strings.Contains(strings.ToLower(err.Error()), "no rows")
}
