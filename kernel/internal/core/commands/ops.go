package commands

import (
	"context"

	"crona/kernel/internal/core"

	sharedtypes "crona/shared/types"
)

func ListLatestOps(ctx context.Context, c *core.Context, limit int) ([]sharedtypes.Op, error) {
	if limit <= 0 {
		limit = 50
	}
	return c.Ops.Latest(ctx, limit)
}

func ListOpsSince(ctx context.Context, c *core.Context, since string) ([]sharedtypes.Op, error) {
	return c.Ops.ListSince(ctx, c.UserID, since)
}

func ListOpsByEntity(
	ctx context.Context,
	c *core.Context,
	entity sharedtypes.OpEntity,
	entityID string,
	limit int,
) ([]sharedtypes.Op, error) {
	return c.Ops.ListByEntity(ctx, entity, entityID, c.UserID, limit)
}
