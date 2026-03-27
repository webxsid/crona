package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"crona/shared/config"
	sharedconstants "crona/shared/constants"
	sharedtypes "crona/shared/types"

	storemodels "crona/kernel/internal/store/models"

	"github.com/uptrace/bun"
)

type CoreSettingsRepository struct {
	db *bun.DB
}

func NewCoreSettingsRepository(db *bun.DB) *CoreSettingsRepository {
	return &CoreSettingsRepository{db: db}
}

func (r *CoreSettingsRepository) Get(ctx context.Context, userID string) (*sharedtypes.CoreSettings, error) {
	var model storemodels.CoreSettingsModel
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
	var model storemodels.CoreSettingsModel
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
	q := r.db.NewUpdate().Model((*storemodels.CoreSettingsModel)(nil)).Where("user_id = ?", userID).Set("updated_at = ?", strconv.FormatInt(time.Now().UnixMilli(), 10))
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
	case sharedtypes.CoreSettingsKeyBoundaryNotifications:
		q = q.Set("boundary_notifications_enabled = ?", value)
	case sharedtypes.CoreSettingsKeyBoundarySound:
		q = q.Set("boundary_sound_enabled = ?", value)
	case sharedtypes.CoreSettingsKeyUpdateChecksEnabled:
		q = q.Set("update_checks_enabled = ?", value)
	case sharedtypes.CoreSettingsKeyUpdatePromptEnabled:
		q = q.Set("update_prompt_enabled = ?", value)
	case sharedtypes.CoreSettingsKeyRepoSort:
		q = q.Set("repo_sort = ?", string(sharedtypes.NormalizeRepoSort(sharedtypes.RepoSort(toString(value)))))
	case sharedtypes.CoreSettingsKeyStreamSort:
		q = q.Set("stream_sort = ?", string(sharedtypes.NormalizeStreamSort(sharedtypes.StreamSort(toString(value)))))
	case sharedtypes.CoreSettingsKeyIssueSort:
		q = q.Set("issue_sort = ?", string(sharedtypes.NormalizeIssueSort(sharedtypes.IssueSort(toString(value)))))
	case sharedtypes.CoreSettingsKeyHabitSort:
		q = q.Set("habit_sort = ?", string(sharedtypes.NormalizeHabitSort(sharedtypes.HabitSort(toString(value)))))
	}
	_, err := q.Exec(ctx)
	return err
}

func (r *CoreSettingsRepository) GetAllSettings(ctx context.Context) (map[string]any, error) {
	var rows []storemodels.CoreSettingsModel
	if err := r.db.NewSelect().Model(&rows).Scan(ctx); err != nil {
		return nil, err
	}
	result := map[string]any{}
	for _, row := range rows {
		result[row.UserID] = coreSettingsFromModel(row)
	}
	return result, nil
}

func (r *CoreSettingsRepository) InitializeDefaults(ctx context.Context, userID string, deviceID string) error {
	var exists int
	err := r.db.NewSelect().Model((*storemodels.CoreSettingsModel)(nil)).ColumnExpr("1").Where("user_id = ?", userID).Limit(1).Scan(ctx, &exists)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	now := strconv.FormatInt(time.Now().UnixMilli(), 10)
	updateChecksEnabled := sharedconstants.DefaultCoreSettings["updateChecksEnabled"].(bool)
	updatePromptEnabled := sharedconstants.DefaultCoreSettings["updatePromptEnabled"].(bool)
	if config.Current().IsDev() {
		updateChecksEnabled = false
		updatePromptEnabled = false
	}
	_, err = r.db.NewInsert().Model(&storemodels.CoreSettingsModel{
		UserID:                userID,
		DeviceID:              deviceID,
		TimerMode:             sharedconstants.DefaultCoreSettings["timerMode"].(string),
		BreaksEnabled:         sharedconstants.DefaultCoreSettings["breaksEnabled"].(bool),
		WorkDurationMinutes:   sharedconstants.DefaultCoreSettings["workDurationMinutes"].(int),
		ShortBreakMinutes:     sharedconstants.DefaultCoreSettings["shortBreakMinutes"].(int),
		LongBreakMinutes:      sharedconstants.DefaultCoreSettings["longBreakMinutes"].(int),
		LongBreakEnabled:      sharedconstants.DefaultCoreSettings["longBreakEnabled"].(bool),
		CyclesBeforeLongBreak: sharedconstants.DefaultCoreSettings["cyclesBeforeLongBreak"].(int),
		AutoStartBreaks:       sharedconstants.DefaultCoreSettings["autoStartBreaks"].(bool),
		AutoStartWork:         sharedconstants.DefaultCoreSettings["autoStartWork"].(bool),
		BoundaryNotifications: sharedconstants.DefaultCoreSettings["boundaryNotificationsEnabled"].(bool),
		BoundarySound:         sharedconstants.DefaultCoreSettings["boundarySoundEnabled"].(bool),
		UpdateChecksEnabled:   updateChecksEnabled,
		UpdatePromptEnabled:   updatePromptEnabled,
		RepoSort:              sharedconstants.DefaultCoreSettings["repoSort"].(string),
		StreamSort:            sharedconstants.DefaultCoreSettings["streamSort"].(string),
		IssueSort:             sharedconstants.DefaultCoreSettings["issueSort"].(string),
		HabitSort:             sharedconstants.DefaultCoreSettings["habitSort"].(string),
		CreatedAt:             now,
		UpdatedAt:             now,
	}).Exec(ctx)
	return err
}

func coreSettingsValue(row storemodels.CoreSettingsModel, key sharedtypes.CoreSettingsKey) any {
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
	case sharedtypes.CoreSettingsKeyBoundaryNotifications:
		return row.BoundaryNotifications
	case sharedtypes.CoreSettingsKeyBoundarySound:
		return row.BoundarySound
	case sharedtypes.CoreSettingsKeyUpdateChecksEnabled:
		return row.UpdateChecksEnabled
	case sharedtypes.CoreSettingsKeyUpdatePromptEnabled:
		return row.UpdatePromptEnabled
	case sharedtypes.CoreSettingsKeyRepoSort:
		return sharedtypes.NormalizeRepoSort(sharedtypes.RepoSort(row.RepoSort))
	case sharedtypes.CoreSettingsKeyStreamSort:
		return sharedtypes.NormalizeStreamSort(sharedtypes.StreamSort(row.StreamSort))
	case sharedtypes.CoreSettingsKeyIssueSort:
		return sharedtypes.NormalizeIssueSort(sharedtypes.IssueSort(row.IssueSort))
	case sharedtypes.CoreSettingsKeyHabitSort:
		return sharedtypes.NormalizeHabitSort(sharedtypes.HabitSort(row.HabitSort))
	default:
		return nil
	}
}

func coreSettingsFromModel(row storemodels.CoreSettingsModel) sharedtypes.CoreSettings {
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
		BoundaryNotifications: row.BoundaryNotifications,
		BoundarySound:         row.BoundarySound,
		UpdateChecksEnabled:   row.UpdateChecksEnabled,
		UpdatePromptEnabled:   row.UpdatePromptEnabled,
		RepoSort:              sharedtypes.NormalizeRepoSort(sharedtypes.RepoSort(row.RepoSort)),
		StreamSort:            sharedtypes.NormalizeStreamSort(sharedtypes.StreamSort(row.StreamSort)),
		IssueSort:             sharedtypes.NormalizeIssueSort(sharedtypes.IssueSort(row.IssueSort)),
		HabitSort:             sharedtypes.NormalizeHabitSort(sharedtypes.HabitSort(row.HabitSort)),
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
