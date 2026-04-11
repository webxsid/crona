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

type coreSettingQueryKind int

const (
	coreSettingQueryString coreSettingQueryKind = iota
	coreSettingQueryBool
	coreSettingQueryInt
)

type coreSettingMeta struct {
	column    string
	queryKind coreSettingQueryKind
}

var coreSettingMetas = map[sharedtypes.CoreSettingsKey]coreSettingMeta{
	sharedtypes.CoreSettingsKeyTimerMode:             {column: "timer_mode", queryKind: coreSettingQueryString},
	sharedtypes.CoreSettingsKeyBreaksEnabled:         {column: "breaks_enabled", queryKind: coreSettingQueryBool},
	sharedtypes.CoreSettingsKeyWorkDurationMinutes:   {column: "work_duration_minutes", queryKind: coreSettingQueryInt},
	sharedtypes.CoreSettingsKeyShortBreakMinutes:     {column: "short_break_minutes", queryKind: coreSettingQueryInt},
	sharedtypes.CoreSettingsKeyLongBreakMinutes:      {column: "long_break_minutes", queryKind: coreSettingQueryInt},
	sharedtypes.CoreSettingsKeyLongBreakEnabled:      {column: "long_break_enabled", queryKind: coreSettingQueryBool},
	sharedtypes.CoreSettingsKeyCyclesBeforeLongBreak: {column: "cycles_before_long_break", queryKind: coreSettingQueryInt},
	sharedtypes.CoreSettingsKeyAutoStartBreaks:       {column: "auto_start_breaks", queryKind: coreSettingQueryBool},
	sharedtypes.CoreSettingsKeyAutoStartWork:         {column: "auto_start_work", queryKind: coreSettingQueryBool},
	sharedtypes.CoreSettingsKeyBoundaryNotifications: {column: "boundary_notifications_enabled", queryKind: coreSettingQueryBool},
	sharedtypes.CoreSettingsKeyBoundarySound:         {column: "boundary_sound_enabled", queryKind: coreSettingQueryBool},
	sharedtypes.CoreSettingsKeyAlertSoundPreset:      {column: "alert_sound_preset", queryKind: coreSettingQueryString},
	sharedtypes.CoreSettingsKeyAlertUrgency:          {column: "alert_urgency", queryKind: coreSettingQueryString},
	sharedtypes.CoreSettingsKeyAlertIconEnabled:      {column: "alert_icon_enabled", queryKind: coreSettingQueryBool},
	sharedtypes.CoreSettingsKeyInactivityAlerts:      {column: "inactivity_alerts_enabled", queryKind: coreSettingQueryBool},
	sharedtypes.CoreSettingsKeyInactivityThreshold:   {column: "inactivity_threshold_minutes", queryKind: coreSettingQueryInt},
	sharedtypes.CoreSettingsKeyInactivityRepeat:      {column: "inactivity_repeat_minutes", queryKind: coreSettingQueryInt},
	sharedtypes.CoreSettingsKeyUpdateChecksEnabled:   {column: "update_checks_enabled", queryKind: coreSettingQueryBool},
	sharedtypes.CoreSettingsKeyUpdatePromptEnabled:   {column: "update_prompt_enabled", queryKind: coreSettingQueryBool},
	sharedtypes.CoreSettingsKeyUpdateChannel:         {column: "update_channel", queryKind: coreSettingQueryString},
	sharedtypes.CoreSettingsKeyRepoSort:              {column: "repo_sort", queryKind: coreSettingQueryString},
	sharedtypes.CoreSettingsKeyStreamSort:            {column: "stream_sort", queryKind: coreSettingQueryString},
	sharedtypes.CoreSettingsKeyIssueSort:             {column: "issue_sort", queryKind: coreSettingQueryString},
	sharedtypes.CoreSettingsKeyHabitSort:             {column: "habit_sort", queryKind: coreSettingQueryString},
	sharedtypes.CoreSettingsKeyAwayModeEnabled:       {column: "away_mode_enabled", queryKind: coreSettingQueryBool},
	sharedtypes.CoreSettingsKeyFrozenStreakKinds:     {column: "frozen_streak_kinds", queryKind: coreSettingQueryString},
	sharedtypes.CoreSettingsKeyRestWeekdays:          {column: "rest_weekdays", queryKind: coreSettingQueryString},
	sharedtypes.CoreSettingsKeyRestSpecificDates:     {column: "rest_specific_dates", queryKind: coreSettingQueryString},
	sharedtypes.CoreSettingsKeyDailyPlanRollbackMins: {column: "daily_plan_rollback_minutes", queryKind: coreSettingQueryInt},
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
	meta, ok := coreSettingMetas[key]
	if !ok {
		return nil, nil
	}
	switch meta.queryKind {
	case coreSettingQueryString:
		var value string
		err := r.db.NewSelect().
			Model((*storemodels.CoreSettingsModel)(nil)).
			Column(meta.column).
			Where("user_id = ?", userID).
			Limit(1).
			Scan(ctx, &value)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		return coreSettingsValueFromColumn(key, value), nil
	case coreSettingQueryBool:
		var value bool
		err := r.db.NewSelect().
			Model((*storemodels.CoreSettingsModel)(nil)).
			Column(meta.column).
			Where("user_id = ?", userID).
			Limit(1).
			Scan(ctx, &value)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		return value, nil
	default:
		var value int
		err := r.db.NewSelect().
			Model((*storemodels.CoreSettingsModel)(nil)).
			Column(meta.column).
			Where("user_id = ?", userID).
			Limit(1).
			Scan(ctx, &value)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		return coreSettingsValueFromColumn(key, value), nil
	}
}

func (r *CoreSettingsRepository) SetSetting(ctx context.Context, userID string, key sharedtypes.CoreSettingsKey, value any) error {
	q := r.db.NewUpdate().Model((*storemodels.CoreSettingsModel)(nil)).Where("user_id = ?", userID).Set("updated_at = ?", strconv.FormatInt(time.Now().UnixMilli(), 10))
	if meta, ok := coreSettingMetas[key]; ok {
		dbValue, err := coreSettingsDBValue(key, value)
		if err != nil {
			return err
		}
		q = q.Set(meta.column+" = ?", dbValue)
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
		AlertSoundPreset:      sharedconstants.DefaultCoreSettings["alertSoundPreset"].(string),
		AlertUrgency:          sharedconstants.DefaultCoreSettings["alertUrgency"].(string),
		AlertIconEnabled:      sharedconstants.DefaultCoreSettings["alertIconEnabled"].(bool),
		InactivityAlerts:      sharedconstants.DefaultCoreSettings["inactivityAlertsEnabled"].(bool),
		InactivityThreshold:   sharedconstants.DefaultCoreSettings["inactivityThresholdMinutes"].(int),
		InactivityRepeat:      sharedconstants.DefaultCoreSettings["inactivityRepeatMinutes"].(int),
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
	case sharedtypes.CoreSettingsKeyAlertSoundPreset:
		return sharedtypes.NormalizeAlertSoundPreset(sharedtypes.AlertSoundPreset(row.AlertSoundPreset))
	case sharedtypes.CoreSettingsKeyAlertUrgency:
		return sharedtypes.NormalizeAlertUrgency(sharedtypes.AlertUrgency(row.AlertUrgency))
	case sharedtypes.CoreSettingsKeyAlertIconEnabled:
		return row.AlertIconEnabled
	case sharedtypes.CoreSettingsKeyInactivityAlerts:
		return row.InactivityAlerts
	case sharedtypes.CoreSettingsKeyInactivityThreshold:
		return row.InactivityThreshold
	case sharedtypes.CoreSettingsKeyInactivityRepeat:
		return row.InactivityRepeat
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

func coreSettingsColumnName(key sharedtypes.CoreSettingsKey) string {
	if meta, ok := coreSettingMetas[key]; ok {
		return meta.column
	}
	return ""
}

func coreSettingsValueFromColumn(key sharedtypes.CoreSettingsKey, value any) any {
	switch key {
	case sharedtypes.CoreSettingsKeyTimerMode:
		return sharedtypes.TimerMode(toString(value))
	case sharedtypes.CoreSettingsKeyUpdateChannel:
		return sharedtypes.NormalizeUpdateChannel(sharedtypes.UpdateChannel(toString(value)))
	case sharedtypes.CoreSettingsKeyAlertSoundPreset:
		return sharedtypes.NormalizeAlertSoundPreset(sharedtypes.AlertSoundPreset(toString(value)))
	case sharedtypes.CoreSettingsKeyAlertUrgency:
		return sharedtypes.NormalizeAlertUrgency(sharedtypes.AlertUrgency(toString(value)))
	case sharedtypes.CoreSettingsKeyRepoSort:
		return sharedtypes.NormalizeRepoSort(sharedtypes.RepoSort(toString(value)))
	case sharedtypes.CoreSettingsKeyStreamSort:
		return sharedtypes.NormalizeStreamSort(sharedtypes.StreamSort(toString(value)))
	case sharedtypes.CoreSettingsKeyIssueSort:
		return sharedtypes.NormalizeIssueSort(sharedtypes.IssueSort(toString(value)))
	case sharedtypes.CoreSettingsKeyHabitSort:
		return sharedtypes.NormalizeHabitSort(sharedtypes.HabitSort(toString(value)))
	case sharedtypes.CoreSettingsKeyFrozenStreakKinds:
		return parseStreakKinds(toString(value))
	case sharedtypes.CoreSettingsKeyRestWeekdays:
		return parseIntSlice(toString(value))
	case sharedtypes.CoreSettingsKeyRestSpecificDates:
		return parseStringSlice(toString(value))
	case sharedtypes.CoreSettingsKeyDailyPlanRollbackMins:
		return clampRollbackMinutes(value)
	case sharedtypes.CoreSettingsKeyInactivityThreshold, sharedtypes.CoreSettingsKeyInactivityRepeat:
		return clampInactivityMinutes(value)
	default:
		return value
	}
}

func coreSettingsDBValue(key sharedtypes.CoreSettingsKey, value any) (any, error) {
	switch key {
	case sharedtypes.CoreSettingsKeyUpdateChannel:
		return string(sharedtypes.NormalizeUpdateChannel(sharedtypes.UpdateChannel(toString(value)))), nil
	case sharedtypes.CoreSettingsKeyAlertSoundPreset:
		return string(sharedtypes.NormalizeAlertSoundPreset(sharedtypes.AlertSoundPreset(toString(value)))), nil
	case sharedtypes.CoreSettingsKeyAlertUrgency:
		return string(sharedtypes.NormalizeAlertUrgency(sharedtypes.AlertUrgency(toString(value)))), nil
	case sharedtypes.CoreSettingsKeyRepoSort:
		return string(sharedtypes.NormalizeRepoSort(sharedtypes.RepoSort(toString(value)))), nil
	case sharedtypes.CoreSettingsKeyStreamSort:
		return string(sharedtypes.NormalizeStreamSort(sharedtypes.StreamSort(toString(value)))), nil
	case sharedtypes.CoreSettingsKeyIssueSort:
		return string(sharedtypes.NormalizeIssueSort(sharedtypes.IssueSort(toString(value)))), nil
	case sharedtypes.CoreSettingsKeyHabitSort:
		return string(sharedtypes.NormalizeHabitSort(sharedtypes.HabitSort(toString(value)))), nil
	case sharedtypes.CoreSettingsKeyFrozenStreakKinds:
		return streakKindsJSON(value)
	case sharedtypes.CoreSettingsKeyRestWeekdays:
		return intSliceJSON(value)
	case sharedtypes.CoreSettingsKeyRestSpecificDates:
		return stringSliceJSON(value)
	case sharedtypes.CoreSettingsKeyDailyPlanRollbackMins:
		return clampRollbackMinutes(value), nil
	case sharedtypes.CoreSettingsKeyInactivityThreshold, sharedtypes.CoreSettingsKeyInactivityRepeat:
		return clampInactivityMinutes(value), nil
	default:
		return value, nil
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
		AlertSoundPreset:      sharedtypes.NormalizeAlertSoundPreset(sharedtypes.AlertSoundPreset(row.AlertSoundPreset)),
		AlertUrgency:          sharedtypes.NormalizeAlertUrgency(sharedtypes.AlertUrgency(row.AlertUrgency)),
		AlertIconEnabled:      row.AlertIconEnabled,
		InactivityAlerts:      row.InactivityAlerts,
		InactivityThreshold:   clampInactivityMinutes(row.InactivityThreshold),
		InactivityRepeat:      clampInactivityMinutes(row.InactivityRepeat),
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

func clampInactivityMinutes(value any) int {
	minutes := 60
	switch typed := value.(type) {
	case int:
		minutes = typed
	case int64:
		minutes = int(typed)
	case float64:
		minutes = int(typed)
	}
	if minutes < 15 {
		return 15
	}
	if minutes > 720 {
		return 720
	}
	return minutes
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
