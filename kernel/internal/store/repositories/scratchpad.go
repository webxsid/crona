package repositories

import (
	"context"
	"database/sql"
	"errors"

	sharedtypes "crona/shared/types"

	storemodels "crona/kernel/internal/store/models"
	"github.com/uptrace/bun"
)

type ScratchPadRepository struct {
	db *bun.DB
}

func NewScratchPadRepository(db *bun.DB) *ScratchPadRepository {
	return &ScratchPadRepository{db: db}
}

func (r *ScratchPadRepository) Upsert(ctx context.Context, meta sharedtypes.ScratchPadMeta, userID string, deviceID string) error {
	_, err := r.db.NewInsert().
		Model(&storemodels.ScratchPadMetaModel{
			ID:           meta.ID,
			UserID:       userID,
			DeviceID:     deviceID,
			Name:         meta.Name,
			Path:         meta.Path,
			LastOpenedAt: meta.LastOpenedAt,
			Pinned:       meta.Pinned,
		}).
		On("CONFLICT (path) DO UPDATE").
		Set("name = EXCLUDED.name").
		Set("last_opened_at = EXCLUDED.last_opened_at").
		Set("pinned = EXCLUDED.pinned").
		Where("user_id = ?", userID).
		Exec(ctx)
	return err
}

func (r *ScratchPadRepository) List(ctx context.Context, userID string, deviceID string, pinnedOnly bool) ([]sharedtypes.ScratchPadMeta, error) {
	var models []storemodels.ScratchPadMetaModel
	q := r.db.NewSelect().Model(&models).Where("user_id = ?", userID).Where("device_id = ?", deviceID)
	if pinnedOnly {
		q = q.Where("pinned = ?", true)
	}
	if err := q.Order("last_opened_at DESC").Scan(ctx); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.ScratchPadMeta, 0, len(models))
	for _, model := range models {
		out = append(out, sharedtypes.ScratchPadMeta{
			ID:           model.ID,
			Name:         model.Name,
			Path:         model.Path,
			LastOpenedAt: model.LastOpenedAt,
			Pinned:       model.Pinned,
		})
	}
	return out, nil
}

func (r *ScratchPadRepository) Get(ctx context.Context, path string, userID string, deviceID string) (*sharedtypes.ScratchPadMeta, error) {
	var model storemodels.ScratchPadMetaModel
	err := r.db.NewSelect().Model(&model).
		Where("path = ?", path).
		Where("user_id = ?", userID).
		Where("device_id = ?", deviceID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sharedtypes.ScratchPadMeta{
		ID:           model.ID,
		Name:         model.Name,
		Path:         model.Path,
		LastOpenedAt: model.LastOpenedAt,
		Pinned:       model.Pinned,
	}, nil
}

func (r *ScratchPadRepository) GetByID(ctx context.Context, id string, userID string, deviceID string) (*sharedtypes.ScratchPadMeta, error) {
	var model storemodels.ScratchPadMetaModel
	err := r.db.NewSelect().Model(&model).
		Where("id = ?", id).
		Where("user_id = ?", userID).
		Where("device_id = ?", deviceID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sharedtypes.ScratchPadMeta{
		ID:           model.ID,
		Name:         model.Name,
		Path:         model.Path,
		LastOpenedAt: model.LastOpenedAt,
		Pinned:       model.Pinned,
	}, nil
}

func (r *ScratchPadRepository) Remove(ctx context.Context, path string, userID string, deviceID string) error {
	_, err := r.db.NewDelete().Model((*storemodels.ScratchPadMetaModel)(nil)).
		Where("path = ?", path).
		Where("user_id = ?", userID).
		Where("device_id = ?", deviceID).
		Exec(ctx)
	return err
}

func (r *ScratchPadRepository) RemoveByID(ctx context.Context, id string, userID string, deviceID string) error {
	_, err := r.db.NewDelete().Model((*storemodels.ScratchPadMetaModel)(nil)).
		Where("id = ?", id).
		Where("user_id = ?", userID).
		Where("device_id = ?", deviceID).
		Exec(ctx)
	return err
}
