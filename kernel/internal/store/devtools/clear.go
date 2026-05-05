package devtools

import (
	"context"

	"github.com/uptrace/bun"
)

func ClearAllData(ctx context.Context, db *bun.DB) error {
	tables := []string{
		"session_segments",
		"sessions",
		"stash",
		"ops",
		"scratch_pad_meta",
		"daily_plan_events",
		"daily_plan_entries",
		"daily_plans",
		"daily_checkins",
		"habit_completions",
		"habit_focus_sessions",
		"habits",
		"active_context",
		"issues",
		"streams",
		"repos",
		"core_settings",
	}

	for _, table := range tables {
		if _, err := db.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			return err
		}
	}

	return nil
}
