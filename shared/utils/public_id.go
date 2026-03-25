package utils

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"
)

// NextPublicID returns the next available public_id for a table.
func NextPublicID(ctx context.Context, db *bun.DB, table string) (int64, error) {
	var maxID sql.NullInt64
	if err := db.NewSelect().
		TableExpr(table).
		ColumnExpr("MAX(public_id)").
		Scan(ctx, &maxID); err != nil {
		return 0, err
	}
	if !maxID.Valid {
		return 1, nil
	}
	return maxID.Int64 + 1, nil
}

// ResolveRepoInternalID resolves a repo public_id to its internal id.
func ResolveRepoInternalID(ctx context.Context, db *bun.DB, repoID int64, userID string) (string, error) {
	var internalID string
	err := db.NewSelect().
		TableExpr("repos").
		ColumnExpr("id").
		Where("public_id = ?", repoID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(ctx, &internalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return internalID, nil
}

// ResolveStreamInternalID resolves a stream public_id to its internal id.
func ResolveStreamInternalID(ctx context.Context, db *bun.DB, streamID int64, userID string) (string, error) {
	var internalID string
	err := db.NewSelect().
		TableExpr("streams").
		ColumnExpr("id").
		Where("public_id = ?", streamID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(ctx, &internalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return internalID, nil
}

// ResolveIssueInternalID resolves an issue public_id to its internal id.
func ResolveIssueInternalID(ctx context.Context, db *bun.DB, issueID int64, userID string) (string, error) {
	var internalID string
	err := db.NewSelect().
		TableExpr("issues").
		ColumnExpr("id").
		Where("public_id = ?", issueID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(ctx, &internalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return internalID, nil
}

// ResolveHabitInternalID resolves a habit public_id to its internal id.
func ResolveHabitInternalID(ctx context.Context, db *bun.DB, habitID int64, userID string) (string, error) {
	var internalID string
	err := db.NewSelect().
		TableExpr("habits").
		ColumnExpr("id").
		Where("public_id = ?", habitID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(ctx, &internalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return internalID, nil
}
