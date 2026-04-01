package devtools_test

import (
	"context"
	"path/filepath"
	"testing"

	"crona/kernel/internal/store"
	"crona/kernel/internal/store/devtools"
)

func TestClearAllDataRemovesDailyPlanTables(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(filepath.Join(t.TempDir(), "crona.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.InitSchema(ctx, db.DB()); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	if _, err := db.DB().ExecContext(ctx, "INSERT INTO daily_plans (id, user_id, date, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", "plan-1", "local", "2026-04-01", "2026-04-01T00:00:00Z", "2026-04-01T00:00:00Z"); err != nil {
		t.Fatalf("insert daily plan: %v", err)
	}
	if _, err := db.DB().ExecContext(ctx, "INSERT INTO daily_plan_entries (id, plan_id, issue_id, source, status, committed_at, baseline_date, current_planned_date, postpone_count, max_delayed_days, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", "entry-1", "plan-1", "issue-internal-1", "todo_for_date", "planned", "2026-04-01T00:00:00Z", "2026-04-01", "2026-04-06", 2, 5, "2026-04-01T00:00:00Z", "2026-04-01T00:00:00Z"); err != nil {
		t.Fatalf("insert daily plan entry: %v", err)
	}
	if _, err := db.DB().ExecContext(ctx, "INSERT INTO daily_plan_events (id, plan_entry_id, user_id, device_id, event_type, payload, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)", "event-1", "entry-1", "local", "test-device", "committed", `{}`, "2026-04-01T00:00:00Z"); err != nil {
		t.Fatalf("insert daily plan event: %v", err)
	}

	if err := devtools.ClearAllData(ctx, db.DB()); err != nil {
		t.Fatalf("clear all data: %v", err)
	}

	for _, table := range []string{"daily_plans", "daily_plan_entries", "daily_plan_events"} {
		var count int
		if err := db.DB().NewSelect().Table(table).ColumnExpr("COUNT(*)").Scan(ctx, &count); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("expected %s to be empty after clear, got %d rows", table, count)
		}
	}
}
