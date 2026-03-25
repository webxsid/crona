package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/uptrace/bun"
)

func ensurePublicIDColumn(ctx context.Context, db *bun.DB, table string) error {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info('%s')", table))
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
		if strings.EqualFold(name, "public_id") {
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
		return nil
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN public_id integer", table))
	return err
}

func backfillPublicIDs(ctx context.Context, db *bun.DB, table string) error {
	type row struct {
		InternalID string `bun:"id"`
	}

	var rows []row
	if err := db.NewSelect().
		TableExpr(table).
		ColumnExpr("id").
		Where("public_id IS NULL OR public_id = 0").
		OrderExpr("created_at ASC, id ASC").
		Scan(ctx, &rows); err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	var maxID sql.NullInt64
	if err := db.NewSelect().
		TableExpr(table).
		ColumnExpr("MAX(public_id)").
		Scan(ctx, &maxID); err != nil {
		return err
	}
	next := int64(1)
	if maxID.Valid && maxID.Int64 >= next {
		next = maxID.Int64 + 1
	}

	for _, row := range rows {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET public_id = ? WHERE id = ?", table), next, row.InternalID); err != nil {
			return err
		}
		next++
	}
	return nil
}
