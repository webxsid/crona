package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
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
	case sharedtypes.CoreSettingsKeyUpdateChannel:
		q = q.Set("update_channel = ?", string(sharedtypes.NormalizeUpdateChannel(sharedtypes.UpdateChannel(toString(value)))))
	case sharedtypes.CoreSettingsKeyRepoSort:
		q = q.Set("repo_sort = ?", string(sharedtypes.NormalizeRepoSort(sharedtypes.RepoSort(toString(value)))))
	case sharedtypes.CoreSettingsKeyStreamSort:
		q = q.Set("stream_sort = ?", string(sharedtypes.NormalizeStreamSort(sharedtypes.StreamSort(toString(value)))))
	case sharedtypes.CoreSettingsKeyIssueSort:
		q = q.Set("issue_sort = ?", string(sharedtypes.NormalizeIssueSort(sharedtypes.IssueSort(toString(value)))))
	case sharedtypes.CoreSettingsKeyHabitSort:
		q = q.Set("habit_sort = ?", string(sharedtypes.NormalizeHabitSort(sharedtypes.HabitSort(toString(value)))))
	case sharedtypes.CoreSettingsKeyAwayModeEnabled:
		q = q.Set("away_mode_enabled = ?", value)
	case sharedtypes.CoreSettingsKeyFrozenStreakKinds:
		raw, err := streakKindsJSON(value)
		if err != nil {
			return err
		}
		q = q.Set("frozen_streak_kinds = ?", raw)
	case sharedtypes.CoreSettingsKeyRestWeekdays:
		raw, err := intSliceJSON(value)
		if err != nil {
			return err
		}
		q = q.Set("rest_weekdays = ?", raw)
	case sharedtypes.CoreSettingsKeyRestSpecificDates:
		raw, err := stringSliceJSON(value)
		if err != nil {
			return err
		}
		q = q.Set("rest_specific_dates = ?", raw)
	case sharedtypes.CoreSettingsKeyDailyPlanRollbackMins:
		q = q.Set("daily_plan_rollback_minutes = ?", clampRollbackMinutes(value))
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
		UpdateChannel:         sharedconstants.DefaultCoreSettings["updateChannel"].(string),
		RepoSort:              sharedconstants.DefaultCoreSettings["repoSort"].(string),
		StreamSort:            sharedconstants.DefaultCoreSettings["streamSort"].(string),
		IssueSort:             sharedconstants.DefaultCoreSettings["issueSort"].(string),
		HabitSort:             sharedconstants.DefaultCoreSettings["habitSort"].(string),
		AwayModeEnabled:       sharedconstants.DefaultCoreSettings["awayModeEnabled"].(bool),
		FrozenStreakKinds:     mustJSON(sharedconstants.DefaultCoreSettings["frozenStreakKinds"]),
		RestWeekdays:          mustJSON(sharedconstants.DefaultCoreSettings["restWeekdays"]),
		RestSpecificDates:     mustJSON(sharedconstants.DefaultCoreSettings["restSpecificDates"]),
		DailyPlanRollbackMins: sharedconstants.DefaultCoreSettings["dailyPlanRollbackMinutes"].(int),
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
	case sharedtypes.CoreSettingsKeyUpdateChannel:
		return sharedtypes.NormalizeUpdateChannel(sharedtypes.UpdateChannel(row.UpdateChannel))
	case sharedtypes.CoreSettingsKeyRepoSort:
		return sharedtypes.NormalizeRepoSort(sharedtypes.RepoSort(row.RepoSort))
	case sharedtypes.CoreSettingsKeyStreamSort:
		return sharedtypes.NormalizeStreamSort(sharedtypes.StreamSort(row.StreamSort))
	case sharedtypes.CoreSettingsKeyIssueSort:
		return sharedtypes.NormalizeIssueSort(sharedtypes.IssueSort(row.IssueSort))
	case sharedtypes.CoreSettingsKeyHabitSort:
		return sharedtypes.NormalizeHabitSort(sharedtypes.HabitSort(row.HabitSort))
	case sharedtypes.CoreSettingsKeyAwayModeEnabled:
		return row.AwayModeEnabled
	case sharedtypes.CoreSettingsKeyFrozenStreakKinds:
		return parseStreakKinds(row.FrozenStreakKinds)
	case sharedtypes.CoreSettingsKeyRestWeekdays:
		return parseIntSlice(row.RestWeekdays)
	case sharedtypes.CoreSettingsKeyRestSpecificDates:
		return parseStringSlice(row.RestSpecificDates)
	case sharedtypes.CoreSettingsKeyDailyPlanRollbackMins:
		return row.DailyPlanRollbackMins
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
		UpdateChannel:         sharedtypes.NormalizeUpdateChannel(sharedtypes.UpdateChannel(row.UpdateChannel)),
		RepoSort:              sharedtypes.NormalizeRepoSort(sharedtypes.RepoSort(row.RepoSort)),
		StreamSort:            sharedtypes.NormalizeStreamSort(sharedtypes.StreamSort(row.StreamSort)),
		IssueSort:             sharedtypes.NormalizeIssueSort(sharedtypes.IssueSort(row.IssueSort)),
		HabitSort:             sharedtypes.NormalizeHabitSort(sharedtypes.HabitSort(row.HabitSort)),
		AwayModeEnabled:       row.AwayModeEnabled,
		FrozenStreakKinds:     parseStreakKinds(row.FrozenStreakKinds),
		RestWeekdays:          parseIntSlice(row.RestWeekdays),
		RestSpecificDates:     parseStringSlice(row.RestSpecificDates),
		DailyPlanRollbackMins: row.DailyPlanRollbackMins,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}

func clampRollbackMinutes(value any) int {
	minutes := 5
	switch typed := value.(type) {
	case int:
		minutes = typed
	case float64:
		minutes = int(typed)
	}
	if minutes < 1 {
		return 1
	}
	if minutes > 120 {
		return 120
	}
	return minutes
}

func toString(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func stringSliceJSON(value any) (string, error) {
	items := []string{}
	switch typed := value.(type) {
	case []string:
		items = typed
	case []any:
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return "", errors.New("expected string slice")
			}
			items = append(items, strings.TrimSpace(text))
		}
	default:
		return "", errors.New("expected string slice")
	}
	normalized := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		normalized = append(normalized, item)
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func intSliceJSON(value any) (string, error) {
	items := []int{}
	switch typed := value.(type) {
	case []int:
		items = typed
	case []any:
		for _, item := range typed {
			number, ok := item.(float64)
			if !ok {
				return "", errors.New("expected integer slice")
			}
			items = append(items, int(number))
		}
	default:
		return "", errors.New("expected integer slice")
	}
	normalized := make([]int, 0, len(items))
	seen := map[int]struct{}{}
	for _, item := range items {
		if item < 0 || item > 6 {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		normalized = append(normalized, item)
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func streakKindsJSON(value any) (string, error) {
	items := []sharedtypes.StreakKind{}
	switch typed := value.(type) {
	case []sharedtypes.StreakKind:
		items = typed
	case []string:
		for _, item := range typed {
			items = append(items, sharedtypes.StreakKind(item))
		}
	case []any:
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return "", errors.New("expected streak kind slice")
			}
			items = append(items, sharedtypes.StreakKind(text))
		}
	default:
		return "", errors.New("expected streak kind slice")
	}
	normalized := normalizeStreakKinds(items)
	raw := make([]string, 0, len(normalized))
	for _, item := range normalized {
		raw = append(raw, string(item))
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseStringSlice(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	out := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func parseIntSlice(raw string) []int {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var items []int
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	out := make([]int, 0, len(items))
	seen := map[int]struct{}{}
	for _, item := range items {
		if item < 0 || item > 6 {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func parseStreakKinds(raw string) []sharedtypes.StreakKind {
	items := parseStringSlice(raw)
	kinds := make([]sharedtypes.StreakKind, 0, len(items))
	for _, item := range items {
		kinds = append(kinds, sharedtypes.StreakKind(item))
	}
	return normalizeStreakKinds(kinds)
}

func normalizeStreakKinds(values []sharedtypes.StreakKind) []sharedtypes.StreakKind {
	allowed := map[sharedtypes.StreakKind]struct{}{}
	for _, kind := range sharedtypes.AvailableStreakKinds() {
		allowed[kind] = struct{}{}
	}
	out := make([]sharedtypes.StreakKind, 0, len(values))
	seen := map[sharedtypes.StreakKind]struct{}{}
	for _, value := range values {
		value = sharedtypes.NormalizeStreakKind(value)
		if _, ok := allowed[value]; !ok {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func mustJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}
