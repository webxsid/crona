package migrations

import (
	"context"
	"database/sql"
	"fmt"

	storemodels "crona/kernel/internal/store/models"
	sharedtypes "crona/shared/types"

	"github.com/uptrace/bun"
)

func InitSchema(ctx context.Context, db *bun.DB) error {
	models := []any{
		(*storemodels.RepoModel)(nil),
		(*storemodels.StreamModel)(nil),
		(*storemodels.IssueModel)(nil),
		(*storemodels.HabitModel)(nil),
		(*storemodels.HabitCompletionModel)(nil),
		(*storemodels.HabitFocusSessionModel)(nil),
		(*storemodels.SessionModel)(nil),
		(*storemodels.StashModel)(nil),
		(*storemodels.OpModel)(nil),
		(*storemodels.CoreSettingsModel)(nil),
		(*storemodels.AlertReminderModel)(nil),
		(*storemodels.SessionSegmentModel)(nil),
		(*storemodels.ActiveContextModel)(nil),
		(*storemodels.ScratchPadMetaModel)(nil),
		(*storemodels.DailyCheckInModel)(nil),
		(*storemodels.DailyPlanModel)(nil),
		(*storemodels.DailyPlanEntryModel)(nil),
		(*storemodels.DailyPlanEventModel)(nil),
	}

	for _, model := range models {
		if _, err := db.NewCreateTable().Model(model).IfNotExists().Exec(ctx); err != nil {
			return err
		}
	}

	if err := ensureIssueColumn(ctx, db, "completed_at"); err != nil {
		return err
	}
	if err := ensureIssueColumn(ctx, db, "abandoned_at"); err != nil {
		return err
	}
	for _, spec := range []struct {
		table  string
		column string
	}{
		{table: "repos", column: "description"},
		{table: "streams", column: "description"},
		{table: "issues", column: "description"},
		{table: "sessions", column: "source"},
		{table: "daily_plan_entries", column: "baseline_date"},
		{table: "daily_plan_entries", column: "current_planned_date"},
		{table: "habit_completions", column: "snapshot_name"},
		{table: "habit_completions", column: "snapshot_description"},
		{table: "habit_completions", column: "snapshot_schedule_type"},
		{table: "habit_completions", column: "snapshot_weekdays"},
		{table: "habit_completions", column: "kind"},
		{table: "habit_completions", column: "started_at"},
		{table: "habit_completions", column: "ended_at"},
		{table: "habit_focus_sessions", column: "kind"},
		{table: "habit_focus_sessions", column: "date"},
		{table: "habit_focus_sessions", column: "started_at"},
		{table: "habit_focus_sessions", column: "ended_at"},
		{table: "habit_focus_sessions", column: "snapshot_name"},
		{table: "habit_focus_sessions", column: "snapshot_description"},
		{table: "habit_focus_sessions", column: "snapshot_schedule_type"},
		{table: "habit_focus_sessions", column: "snapshot_weekdays"},
	} {
		if err := ensureTextColumn(ctx, db, spec.table, spec.column); err != nil {
			return err
		}
	}
	for _, spec := range []struct {
		table        string
		column       string
		defaultValue int
	}{
		{table: "issues", column: "pinned_daily", defaultValue: 0},
		{table: "daily_plan_entries", column: "postpone_count", defaultValue: 0},
		{table: "daily_plan_entries", column: "max_delayed_days", defaultValue: 0},
		{table: "core_settings", column: "daily_plan_rollback_minutes", defaultValue: 5},
	} {
		if err := ensureIntegerColumn(ctx, db, spec.table, spec.column, spec.defaultValue); err != nil {
			return err
		}
	}
	if err := ensureNullableIntegerColumn(ctx, db, "habit_completions", "snapshot_target_minutes"); err != nil {
		return err
	}
	if err := ensureNullableIntegerColumn(ctx, db, "habit_focus_sessions", "snapshot_target_minutes"); err != nil {
		return err
	}
	for columnName, defaultValue := range map[string]string{
		"repo_sort":                "chronological_asc",
		"stream_sort":              "chronological_asc",
		"issue_sort":               "priority",
		"habit_sort":               "schedule",
		"alert_sound_preset":       "chime",
		"alert_urgency":            "normal",
		"update_channel":           "stable",
		"date_display_preset":      "iso",
		"date_display_format":      "",
		"prompt_glyph_mode":        "emoji",
		"habit_streak_definitions": "[]",
		"frozen_streak_kinds":      "[]",
		"rest_weekdays":            "[]",
		"rest_specific_dates":      "[]",
		"rest_recurring_dates":     "[]",
	} {
		if err := ensureCoreSettingsColumn(ctx, db, columnName, defaultValue); err != nil {
			return err
		}
	}
	if err := backfillSessionSource(ctx, db); err != nil {
		return err
	}
	for columnName, defaultValue := range map[string]int{
		"boundary_notifications_enabled": 1,
		"boundary_sound_enabled":         1,
		"alert_icon_enabled":             1,
		"update_checks_enabled":          1,
		"update_prompt_enabled":          1,
		"away_mode_enabled":              0,
		"inactivity_alerts_enabled":      1,
		"inactivity_threshold_minutes":   60,
		"inactivity_repeat_minutes":      60,
		"onboarding_completed":           0,
		"usage_telemetry_enabled":        0,
		"error_reporting_enabled":        0,
	} {
		if err := ensureCoreSettingsBoolColumn(ctx, db, columnName, defaultValue); err != nil {
			return err
		}
	}
	if err := ensureHabitCompletionStatusColumn(ctx, db); err != nil {
		return err
	}
	for _, table := range []string{"repos", "streams", "issues", "habits", "habit_completions", "habit_focus_sessions"} {
		if err := ensurePublicIDColumn(ctx, db, table); err != nil {
			return err
		}
		if err := backfillPublicIDs(ctx, db, table); err != nil {
			return err
		}
	}
	if err := backfillHabitCompletionKinds(ctx, db); err != nil {
		return err
	}
	if err := migrateLegacyIssueStatuses(ctx, db); err != nil {
		return err
	}

	indexes := []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_repos_public_id ON repos (public_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_streams_public_id ON streams (public_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_issues_public_id ON issues (public_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_habits_public_id ON habits (public_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_habit_completions_public_id ON habit_completions (public_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_habit_focus_sessions_public_id ON habit_focus_sessions (public_id)",
		"CREATE INDEX IF NOT EXISTS idx_streams_repo_id ON streams (repo_id)",
		"CREATE INDEX IF NOT EXISTS idx_issues_stream_id ON issues (stream_id)",
		"CREATE INDEX IF NOT EXISTS idx_habits_stream_id ON habits (stream_id)",
		"CREATE INDEX IF NOT EXISTS idx_habit_completions_habit_id ON habit_completions (habit_id)",
		"CREATE INDEX IF NOT EXISTS idx_habit_focus_sessions_habit_id ON habit_focus_sessions (habit_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_habit_completions_habit_date ON habit_completions (habit_id, date, user_id) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_habit_focus_sessions_habit_date ON habit_focus_sessions (habit_id, date, user_id)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_issue_id ON sessions (issue_id)",
		"CREATE INDEX IF NOT EXISTS idx_stash_repo_id ON stash (repo_id)",
		"CREATE INDEX IF NOT EXISTS idx_stash_stream_id ON stash (stream_id)",
		"CREATE INDEX IF NOT EXISTS idx_stash_issue_id ON stash (issue_id)",
		"CREATE INDEX IF NOT EXISTS idx_ops_entity_entity_id ON ops (entity, entity_id)",
		"CREATE INDEX IF NOT EXISTS idx_repos_user_id ON repos (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_streams_user_id ON streams (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_issues_user_id ON issues (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_habits_user_id ON habits (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_habit_completions_user_id ON habit_completions (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_habit_focus_sessions_user_id ON habit_focus_sessions (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_stash_user_id ON stash (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_ops_user_id ON ops (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_session_segments_session_id ON session_segments (session_id)",
		"CREATE INDEX IF NOT EXISTS idx_session_segments_user_id ON session_segments (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_active_context_user_id ON active_context (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_active_context_device_id ON active_context (device_id)",
		"CREATE INDEX IF NOT EXISTS idx_scratch_pad_meta_user_id ON scratch_pad_meta (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_scratch_pad_meta_device_id ON scratch_pad_meta (device_id)",
		"CREATE INDEX IF NOT EXISTS idx_scratch_pad_meta_last_opened_at ON scratch_pad_meta (last_opened_at)",
		"CREATE INDEX IF NOT EXISTS idx_daily_checkins_device_id ON daily_checkins (device_id)",
		"CREATE INDEX IF NOT EXISTS idx_daily_checkins_updated_at ON daily_checkins (updated_at)",
		"CREATE INDEX IF NOT EXISTS idx_alert_reminders_user_id ON alert_reminders (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_alert_reminders_enabled_time ON alert_reminders (enabled, time_hhmm)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_plans_user_date ON daily_plans (user_id, date)",
		"CREATE INDEX IF NOT EXISTS idx_daily_plan_entries_plan_id ON daily_plan_entries (plan_id)",
		"CREATE INDEX IF NOT EXISTS idx_daily_plan_entries_issue_id ON daily_plan_entries (issue_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_plan_entries_plan_issue ON daily_plan_entries (plan_id, issue_id)",
		"CREATE INDEX IF NOT EXISTS idx_daily_plan_events_entry_id ON daily_plan_events (plan_entry_id)",
		"CREATE INDEX IF NOT EXISTS idx_daily_plan_events_user_id ON daily_plan_events (user_id)",
	}

	for _, stmt := range indexes {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}

	return nil
}

func migrateLegacyIssueStatuses(ctx context.Context, db *bun.DB) error {
	replacements := map[string]string{
		"todo":   string(sharedtypes.IssueStatusBacklog),
		"active": string(sharedtypes.IssueStatusInProgress),
	}
	for from, to := range replacements {
		if _, err := db.ExecContext(ctx, "UPDATE issues SET status = ? WHERE status = ?", to, from); err != nil {
			return err
		}
	}
	return nil
}

func ensureIssueColumn(ctx context.Context, db *bun.DB, columnName string) error {
	return ensureTextColumn(ctx, db, "issues", columnName)
}

func ensureTextColumn(ctx context.Context, db *bun.DB, tableName string, columnName string) error {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info('%s')", tableName))
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	var (
		cid       int
		name      string
		typ       string
		notnull   int
		dfltValue sql.NullString
		pk        int
	)
	for rows.Next() {
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == columnName {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s text", tableName, columnName))
	return err
}

func backfillSessionSource(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, "UPDATE sessions SET source = ? WHERE source IS NULL OR TRIM(source) = ''", string(sharedtypes.SessionSourceTracked)); err != nil {
		return err
	}
	return nil
}

func backfillHabitCompletionKinds(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, "UPDATE habit_completions SET kind = ? WHERE kind IS NULL OR TRIM(kind) = ''", string(sharedtypes.HabitHistoryKindCompletion)); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, "UPDATE habit_completions SET started_at = NULL WHERE started_at IS NULL OR TRIM(started_at) = ''"); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, "UPDATE habit_completions SET ended_at = NULL WHERE ended_at IS NULL OR TRIM(ended_at) = ''"); err != nil {
		return err
	}
	return nil
}

func ensureIntegerColumn(ctx context.Context, db *bun.DB, tableName string, columnName string, defaultValue int) error {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info('%s')", tableName))
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	var (
		cid       int
		name      string
		typ       string
		notnull   int
		dfltValue sql.NullString
		pk        int
	)
	for rows.Next() {
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == columnName {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s integer NOT NULL DEFAULT %d", tableName, columnName, defaultValue))
	return err
}

func ensureNullableIntegerColumn(ctx context.Context, db *bun.DB, tableName string, columnName string) error {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info('%s')", tableName))
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	var (
		cid       int
		name      string
		typ       string
		notnull   int
		dfltValue sql.NullString
		pk        int
	)
	for rows.Next() {
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == columnName {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s integer", tableName, columnName))
	return err
}

func ensureCoreSettingsColumn(ctx context.Context, db *bun.DB, columnName string, defaultValue string) error {
	rows, err := db.QueryContext(ctx, "PRAGMA table_info('core_settings')")
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	var (
		cid       int
		name      string
		typ       string
		notnull   int
		dfltValue sql.NullString
		pk        int
	)
	for rows.Next() {
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == columnName {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE core_settings ADD COLUMN %s text NOT NULL DEFAULT '%s'", columnName, defaultValue))
	return err
}

func ensureCoreSettingsBoolColumn(ctx context.Context, db *bun.DB, columnName string, defaultValue int) error {
	rows, err := db.QueryContext(ctx, "PRAGMA table_info('core_settings')")
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	var (
		cid       int
		name      string
		typ       string
		notnull   int
		dfltValue sql.NullString
		pk        int
	)
	for rows.Next() {
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == columnName {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE core_settings ADD COLUMN %s integer NOT NULL DEFAULT %d", columnName, defaultValue))
	return err
}

func ensureHabitCompletionStatusColumn(ctx context.Context, db *bun.DB) error {
	rows, err := db.QueryContext(ctx, "PRAGMA table_info('habit_completions')")
	if err != nil {
		return err
	}

	var (
		cid       int
		name      string
		typ       string
		notnull   int
		dfltValue sql.NullString
		pk        int
	)
	found := false
	for rows.Next() {
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			_ = rows.Close()
			return err
		}
		if name == "status" {
			found = true
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	if found {
		_, err := db.ExecContext(ctx, "UPDATE habit_completions SET status = 'completed' WHERE status IS NULL OR status = ''")
		return err
	}

	if _, err := db.ExecContext(ctx, "ALTER TABLE habit_completions ADD COLUMN status text NOT NULL DEFAULT 'completed'"); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "UPDATE habit_completions SET status = 'completed' WHERE status IS NULL OR status = ''")
	return err
}
