package store

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	sharedtypes "crona/shared/types"

	"github.com/uptrace/bun"
)

type CoreSettingsRepository struct {
	db *bun.DB
}

func NewCoreSettingsRepository(db *bun.DB) *CoreSettingsRepository {
	return &CoreSettingsRepository{db: db}
}

func (r *CoreSettingsRepository) Get(ctx context.Context, userID string) (*sharedtypes.CoreSettings, error) {
	var model CoreSettingsModel
	err := r.db.NewSelect().
		Model(&model).
		Where("user_id = ?", userID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	settings := coreSettingsFromModel(model)
	return &settings, nil
}

func (r *CoreSettingsRepository) GetSetting(ctx context.Context, userID string, key sharedtypes.CoreSettingsKey) (any, error) {
	var model CoreSettingsModel
	err := r.db.NewSelect().
		Model(&model).
		Where("user_id = ?", userID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return coreSettingsValue(model, key), nil
}

func (r *CoreSettingsRepository) SetSetting(ctx context.Context, userID string, key sharedtypes.CoreSettingsKey, value any) error {
	q := r.db.NewUpdate().Model((*CoreSettingsModel)(nil)).Where("user_id = ?", userID).Set("updated_at = ?", strconv.FormatInt(time.Now().UnixMilli(), 10))
	switch key {
	case sharedtypes.CoreSettingsKeyTimerMode:
		q = q.Set("timer_mode = ?", value)
	case sharedtypes.CoreSettingsKeyBreaksEnabled:
		q = q.Set("breaks_enabled = ?", value)
	case sharedtypes.CoreSettingsKeyWorkDurationMinutes:
		q = q.Set("work_duration_minutes = ?", value)
	case sharedtypes.CoreSettingsKeyShortBreakMinutes:
		q = q.Set("short_break_minutes = ?", value)
	case sharedtypes.CoreSettingsKeyLongBreakMinutes:
		q = q.Set("long_break_minutes = ?", value)
	case sharedtypes.CoreSettingsKeyLongBreakEnabled:
		q = q.Set("long_break_enabled = ?", value)
	case sharedtypes.CoreSettingsKeyCyclesBeforeLongBreak:
		q = q.Set("cycles_before_long_break = ?", value)
	case sharedtypes.CoreSettingsKeyAutoStartBreaks:
		q = q.Set("auto_start_breaks = ?", value)
	case sharedtypes.CoreSettingsKeyAutoStartWork:
		q = q.Set("auto_start_work = ?", value)
	case sharedtypes.CoreSettingsKeyRepoSort:
		q = q.Set("repo_sort = ?", string(sharedtypes.NormalizeRepoSort(sharedtypes.RepoSort(toString(value)))))
	case sharedtypes.CoreSettingsKeyStreamSort:
		q = q.Set("stream_sort = ?", string(sharedtypes.NormalizeStreamSort(sharedtypes.StreamSort(toString(value)))))
	case sharedtypes.CoreSettingsKeyIssueSort:
		q = q.Set("issue_sort = ?", string(sharedtypes.NormalizeIssueSort(sharedtypes.IssueSort(toString(value)))))
	}
	_, err := q.Exec(ctx)
	return err
}

func (r *CoreSettingsRepository) GetAllSettings(ctx context.Context) (map[string]any, error) {
	var rows []CoreSettingsModel
	if err := r.db.NewSelect().Model(&rows).Scan(ctx); err != nil {
		return nil, err
	}
	result := map[string]any{}
	for _, row := range rows {
		result[row.UserID] = row
	}
	return result, nil
}

func (r *CoreSettingsRepository) InitializeDefaults(ctx context.Context, userID string, deviceID string) error {
	var exists int
	err := r.db.NewSelect().Model((*CoreSettingsModel)(nil)).ColumnExpr("1").Where("user_id = ?", userID).Limit(1).Scan(ctx, &exists)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	now := strconv.FormatInt(time.Now().UnixMilli(), 10)
	_, err = r.db.NewInsert().Model(&CoreSettingsModel{
		UserID:                userID,
		DeviceID:              deviceID,
		TimerMode:             DefaultCoreSettings["timerMode"].(string),
		BreaksEnabled:         DefaultCoreSettings["breaksEnabled"].(bool),
		WorkDurationMinutes:   DefaultCoreSettings["workDurationMinutes"].(int),
		ShortBreakMinutes:     DefaultCoreSettings["shortBreakMinutes"].(int),
		LongBreakMinutes:      DefaultCoreSettings["longBreakMinutes"].(int),
		LongBreakEnabled:      DefaultCoreSettings["longBreakEnabled"].(bool),
		CyclesBeforeLongBreak: DefaultCoreSettings["cyclesBeforeLongBreak"].(int),
		AutoStartBreaks:       DefaultCoreSettings["autoStartBreaks"].(bool),
		AutoStartWork:         DefaultCoreSettings["autoStartWork"].(bool),
		RepoSort:              DefaultCoreSettings["repoSort"].(string),
		StreamSort:            DefaultCoreSettings["streamSort"].(string),
		IssueSort:             DefaultCoreSettings["issueSort"].(string),
		CreatedAt:             now,
		UpdatedAt:             now,
	}).Exec(ctx)
	return err
}

func coreSettingsValue(row CoreSettingsModel, key sharedtypes.CoreSettingsKey) any {
	switch key {
	case sharedtypes.CoreSettingsKeyTimerMode:
		return row.TimerMode
	case sharedtypes.CoreSettingsKeyBreaksEnabled:
		return row.BreaksEnabled
	case sharedtypes.CoreSettingsKeyWorkDurationMinutes:
		return row.WorkDurationMinutes
	case sharedtypes.CoreSettingsKeyShortBreakMinutes:
		return row.ShortBreakMinutes
	case sharedtypes.CoreSettingsKeyLongBreakMinutes:
		return row.LongBreakMinutes
	case sharedtypes.CoreSettingsKeyLongBreakEnabled:
		return row.LongBreakEnabled
	case sharedtypes.CoreSettingsKeyCyclesBeforeLongBreak:
		return row.CyclesBeforeLongBreak
	case sharedtypes.CoreSettingsKeyAutoStartBreaks:
		return row.AutoStartBreaks
	case sharedtypes.CoreSettingsKeyAutoStartWork:
		return row.AutoStartWork
	case sharedtypes.CoreSettingsKeyRepoSort:
		return sharedtypes.NormalizeRepoSort(sharedtypes.RepoSort(row.RepoSort))
	case sharedtypes.CoreSettingsKeyStreamSort:
		return sharedtypes.NormalizeStreamSort(sharedtypes.StreamSort(row.StreamSort))
	case sharedtypes.CoreSettingsKeyIssueSort:
		return sharedtypes.NormalizeIssueSort(sharedtypes.IssueSort(row.IssueSort))
	default:
		return nil
	}
}

func coreSettingsFromModel(row CoreSettingsModel) sharedtypes.CoreSettings {
	return sharedtypes.CoreSettings{
		UserID:                row.UserID,
		DeviceID:              row.DeviceID,
		TimerMode:             sharedtypes.TimerMode(row.TimerMode),
		BreaksEnabled:         row.BreaksEnabled,
		WorkDurationMinutes:   row.WorkDurationMinutes,
		ShortBreakMinutes:     row.ShortBreakMinutes,
		LongBreakMinutes:      row.LongBreakMinutes,
		LongBreakEnabled:      row.LongBreakEnabled,
		CyclesBeforeLongBreak: row.CyclesBeforeLongBreak,
		AutoStartBreaks:       row.AutoStartBreaks,
		AutoStartWork:         row.AutoStartWork,
		RepoSort:              sharedtypes.NormalizeRepoSort(sharedtypes.RepoSort(row.RepoSort)),
		StreamSort:            sharedtypes.NormalizeStreamSort(sharedtypes.StreamSort(row.StreamSort)),
		IssueSort:             sharedtypes.NormalizeIssueSort(sharedtypes.IssueSort(row.IssueSort)),
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}

func toString(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}
