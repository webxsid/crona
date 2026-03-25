package repositories

import (
	"context"
	"encoding/json"

	sharedtypes "crona/shared/types"

	storemodels "crona/kernel/internal/store/models"
	"github.com/uptrace/bun"
)

type OpRepository struct {
	db *bun.DB
}

func NewOpRepository(db *bun.DB) *OpRepository {
	return &OpRepository{db: db}
}

func (r *OpRepository) Append(ctx context.Context, op sharedtypes.Op) error {
	payload, err := json.Marshal(op.Payload)
	if err != nil {
		return err
	}
	_, err = r.db.NewInsert().Model(&storemodels.OpModel{
		ID:        op.ID,
		Entity:    string(op.Entity),
		EntityID:  op.EntityID,
		Action:    string(op.Action),
		Payload:   string(payload),
		Timestamp: op.Timestamp,
		UserID:    op.UserID,
		DeviceID:  op.DeviceID,
	}).Exec(ctx)
	return err
}

func (r *OpRepository) Latest(ctx context.Context, limit int) ([]sharedtypes.Op, error) {
	var rows []storemodels.OpModel
	if err := r.db.NewSelect().Model(&rows).Order("timestamp DESC").Limit(limit).Scan(ctx); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.Op, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		out = append(out, mapOp(rows[i]))
	}
	return out, nil
}

func (r *OpRepository) ListSince(ctx context.Context, userID string, sinceTimestamp string) ([]sharedtypes.Op, error) {
	var rows []storemodels.OpModel
	if err := r.db.NewSelect().Model(&rows).Where("user_id = ?", userID).Where("timestamp > ?", sinceTimestamp).Order("timestamp ASC").Scan(ctx); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.Op, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapOp(row))
	}
	return out, nil
}

func (r *OpRepository) ListByEntity(ctx context.Context, entity sharedtypes.OpEntity, entityID string, userID string, limit int) ([]sharedtypes.Op, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []storemodels.OpModel
	if err := r.db.NewSelect().Model(&rows).
		Where("entity = ?", string(entity)).
		Where("entity_id = ?", entityID).
		Where("user_id = ?", userID).
		Order("timestamp ASC").
		Limit(limit).
		Scan(ctx); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.Op, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapOp(row))
	}
	return out, nil
}

func mapOp(row storemodels.OpModel) sharedtypes.Op {
	var payload any
	_ = json.Unmarshal([]byte(row.Payload), &payload)
	return sharedtypes.Op{
		ID:        row.ID,
		Entity:    sharedtypes.OpEntity(row.Entity),
		EntityID:  row.EntityID,
		Action:    sharedtypes.OpAction(row.Action),
		Payload:   payload,
		Timestamp: row.Timestamp,
		UserID:    row.UserID,
		DeviceID:  row.DeviceID,
	}
}
