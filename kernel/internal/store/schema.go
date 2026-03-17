package store

import (
	"context"
	"database/sql"
	"fmt"

	sharedtypes "crona/shared/types"

	"github.com/uptrace/bun"
)

func InitSchema(ctx context.Context, db *bun.DB) error {
	models := []any{
		(*RepoModel)(nil),
		(*StreamModel)(nil),
		(*IssueModel)(nil),
		(*HabitModel)(nil),
		(*HabitCompletionModel)(nil),
		(*SessionModel)(nil),
		(*StashModel)(nil),
		(*OpModel)(nil),
		(*CoreSettingsModel)(nil),
		(*SessionSegmentModel)(nil),
		(*ActiveContextModel)(nil),
		(*ScratchPadMetaModel)(nil),
		(*DailyCheckInModel)(nil),
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
	} {
		if err := ensureTextColumn(ctx, db, spec.table, spec.column); err != nil {
			return err
		}
	}
	for columnName, defaultValue := range map[string]string{
		"repo_sort":   "chronological_asc",
		"stream_sort": "chronological_asc",
		"issue_sort":  "priority",
	} {
		if err := ensureCoreSettingsColumn(ctx, db, columnName, defaultValue); err != nil {
			return err
		}
	}
	if err := ensureHabitCompletionStatusColumn(ctx, db); err != nil {
		return err
	}
	for _, table := range []string{"repos", "streams", "issues", "habits", "habit_completions"} {
		if err := ensurePublicIDColumn(ctx, db, table); err != nil {
			return err
		}
		if err := backfillPublicIDs(ctx, db, table); err != nil {
			return err
		}
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
		"CREATE INDEX IF NOT EXISTS idx_streams_repo_id ON streams (repo_id)",
		"CREATE INDEX IF NOT EXISTS idx_issues_stream_id ON issues (stream_id)",
		"CREATE INDEX IF NOT EXISTS idx_habits_stream_id ON habits (stream_id)",
		"CREATE INDEX IF NOT EXISTS idx_habit_completions_habit_id ON habit_completions (habit_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_habit_completions_habit_date ON habit_completions (habit_id, date, user_id) WHERE deleted_at IS NULL",
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
