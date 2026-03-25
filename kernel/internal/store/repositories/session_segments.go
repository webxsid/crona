package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	storemodels "crona/kernel/internal/store/models"
	sharedtypes "crona/shared/types"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type SessionSegmentRepository struct {
	db *bun.DB
}

func NewSessionSegmentRepository(db *bun.DB) *SessionSegmentRepository {
	return &SessionSegmentRepository{db: db}
}

func (r *SessionSegmentRepository) GetActive(ctx context.Context, userID string, deviceID string, sessionID string) (*sharedtypes.SessionSegment, error) {
	var model storemodels.SessionSegmentModel
	err := r.db.NewSelect().
		Model(&model).
		Where("user_id = ?", userID).
		Where("device_id = ?", deviceID).
		Where("session_id = ?", sessionID).
		Where("end_time IS NULL").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return segmentFromModel(model), nil
}

func (r *SessionSegmentRepository) StartSegment(ctx context.Context, userID string, deviceID string, sessionID string, segmentType sharedtypes.SessionSegmentType) (*sharedtypes.SessionSegment, error) {
	if err := r.EndActiveSegment(ctx, userID, deviceID, sessionID); err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	model := storemodels.SessionSegmentModel{
		ID:          uuid.NewString(),
		UserID:      userID,
		DeviceID:    deviceID,
		SessionID:   sessionID,
		SegmentType: string(segmentType),
		StartTime:   now,
		CreatedAt:   now,
	}
	if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
		return nil, err
	}
	return segmentFromModel(model), nil
}

func (r *SessionSegmentRepository) EndActiveSegment(ctx context.Context, userID string, deviceID string, sessionID string) error {
	_, err := r.db.NewUpdate().
		Model((*storemodels.SessionSegmentModel)(nil)).
		Where("user_id = ?", userID).
		Where("device_id = ?", deviceID).
		Where("session_id = ?", sessionID).
		Where("end_time IS NULL").
		Set("end_time = ?", time.Now().UTC().Format(time.RFC3339)).
		Exec(ctx)
	return err
}

func (r *SessionSegmentRepository) ListBySession(ctx context.Context, sessionID string) ([]sharedtypes.SessionSegment, error) {
	var models []storemodels.SessionSegmentModel
	if err := r.db.NewSelect().Model(&models).Where("session_id = ?", sessionID).Order("start_time ASC").Scan(ctx); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.SessionSegment, 0, len(models))
	for _, model := range models {
		out = append(out, *segmentFromModel(model))
	}
	return out, nil
}

func (r *SessionSegmentRepository) ListEndedInRange(ctx context.Context, userID string, since string, until string) ([]sharedtypes.SessionSegment, error) {
	var models []storemodels.SessionSegmentModel
	if err := r.db.NewSelect().
		Model(&models).
		Where("user_id = ?", userID).
		Where("end_time IS NOT NULL").
		Where("start_time >= ?", since).
		Where("start_time <= ?", until).
		OrderExpr("start_time ASC").
		Scan(ctx); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.SessionSegment, 0, len(models))
	for _, model := range models {
		out = append(out, *segmentFromModel(model))
	}
	return out, nil
}

func (r *SessionSegmentRepository) ApplyElapsedOffset(ctx context.Context, sessionID string, offsetSeconds int) error {
	if offsetSeconds <= 0 {
		return nil
	}
	_, err := r.db.NewUpdate().
		Model((*storemodels.SessionSegmentModel)(nil)).
		Where("session_id = ?", sessionID).
		Where("end_time IS NULL").
		Set("elapsed_offset_seconds = ?", offsetSeconds).
		Exec(ctx)
	return err
}

func (r *SessionSegmentRepository) CountWorkSegments(ctx context.Context, sessionID string) (int, error) {
	count, err := r.db.NewSelect().
		Model((*storemodels.SessionSegmentModel)(nil)).
		Where("session_id = ?", sessionID).
		Where("segment_type = ?", string(sharedtypes.SessionSegmentWork)).
		Where("end_time IS NOT NULL").
		Count(ctx)
	return count, err
}

func segmentFromModel(model storemodels.SessionSegmentModel) *sharedtypes.SessionSegment {
	return &sharedtypes.SessionSegment{
		ID:                   model.ID,
		UserID:               model.UserID,
		DeviceID:             model.DeviceID,
		SessionID:            model.SessionID,
		SegmentType:          sharedtypes.SessionSegmentType(model.SegmentType),
		StartTime:            model.StartTime,
		EndTime:              model.EndTime,
		ElapsedOffsetSeconds: model.ElapsedOffsetSeconds,
		CreatedAt:            model.CreatedAt,
	}
}
