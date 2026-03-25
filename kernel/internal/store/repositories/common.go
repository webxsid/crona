package repositories

import (
	"context"

	sharedtypes "crona/shared/types"
	sharedutils "crona/shared/utils"

	"github.com/uptrace/bun"
)

type Patch[T any] = sharedtypes.Patch[T]

func nextPublicID(ctx context.Context, db *bun.DB, table string) (int64, error) {
	return sharedutils.NextPublicID(ctx, db, table)
}

func resolveRepoInternalID(ctx context.Context, db *bun.DB, repoID int64, userID string) (string, error) {
	return sharedutils.ResolveRepoInternalID(ctx, db, repoID, userID)
}

func resolveStreamInternalID(ctx context.Context, db *bun.DB, streamID int64, userID string) (string, error) {
	return sharedutils.ResolveStreamInternalID(ctx, db, streamID, userID)
}

func resolveIssueInternalID(ctx context.Context, db *bun.DB, issueID int64, userID string) (string, error) {
	return sharedutils.ResolveIssueInternalID(ctx, db, issueID, userID)
}

func resolveHabitInternalID(ctx context.Context, db *bun.DB, habitID int64, userID string) (string, error) {
	return sharedutils.ResolveHabitInternalID(ctx, db, habitID, userID)
}
