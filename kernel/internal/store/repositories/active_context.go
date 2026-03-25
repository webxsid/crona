package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	storemodels "crona/kernel/internal/store/models"
	sharedtypes "crona/shared/types"
	utils "crona/shared/utils"

	"github.com/uptrace/bun"
)

type ActiveContextRepository struct {
	db *bun.DB
}

func NewActiveContextRepository(db *bun.DB) *ActiveContextRepository {
	return &ActiveContextRepository{db: db}
}

func (r *ActiveContextRepository) Get(ctx context.Context, userID string, deviceID string) (*sharedtypes.ActiveContext, error) {
	type row struct {
		UserID     string  `bun:"user_id"`
		DeviceID   string  `bun:"device_id"`
		RepoID     *int64  `bun:"repo_public_id"`
		RepoName   *string `bun:"repo_name"`
		StreamID   *int64  `bun:"stream_public_id"`
		StreamName *string `bun:"stream_name"`
		IssueID    *int64  `bun:"issue_public_id"`
		IssueTitle *string `bun:"issue_title"`
		UpdatedAt  string  `bun:"updated_at"`
	}

	var model row
	err := r.db.NewSelect().
		TableExpr("active_context").
		Join("LEFT JOIN repos ON repos.id = active_context.repo_id AND repos.deleted_at IS NULL").
		Join("LEFT JOIN streams ON streams.id = active_context.stream_id AND streams.deleted_at IS NULL").
		Join("LEFT JOIN issues ON issues.id = active_context.issue_id AND issues.deleted_at IS NULL").
		ColumnExpr("active_context.user_id").
		ColumnExpr("active_context.device_id").
		ColumnExpr("repos.public_id AS repo_public_id").
		ColumnExpr("repos.name AS repo_name").
		ColumnExpr("streams.public_id AS stream_public_id").
		ColumnExpr("streams.name AS stream_name").
		ColumnExpr("issues.public_id AS issue_public_id").
		ColumnExpr("issues.title AS issue_title").
		ColumnExpr("active_context.updated_at").
		Where("active_context.user_id = ?", userID).
		Where("active_context.device_id = ?", deviceID).
		Limit(1).
		Scan(ctx, &model)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sharedtypes.ActiveContext{
		UserID:     model.UserID,
		DeviceID:   model.DeviceID,
		RepoID:     model.RepoID,
		RepoName:   model.RepoName,
		StreamID:   model.StreamID,
		StreamName: model.StreamName,
		IssueID:    model.IssueID,
		IssueTitle: model.IssueTitle,
		UpdatedAt:  &model.UpdatedAt,
	}, nil
}

func (r *ActiveContextRepository) Set(ctx context.Context, userID string, deviceID string, value struct {
	RepoID   *int64
	StreamID *int64
	IssueID  *int64
},
) (*sharedtypes.ActiveContext, error) {
	var (
		repoInternalID   *string
		streamInternalID *string
		issueInternalID  *string
	)

	if value.RepoID != nil {
		resolved, err := utils.ResolveRepoInternalID(ctx, r.db, *value.RepoID, userID)
		if err != nil {
			return nil, err
		}
		if resolved == "" {
			return nil, errors.New("repo not found")
		}
		repoInternalID = &resolved
	}
	if value.StreamID != nil {
		resolved, err := utils.ResolveStreamInternalID(ctx, r.db, *value.StreamID, userID)
		if err != nil {
			return nil, err
		}
		if resolved == "" {
			return nil, errors.New("stream not found")
		}
		streamInternalID = &resolved
	}
	if value.IssueID != nil {
		resolved, err := utils.ResolveIssueInternalID(ctx, r.db, *value.IssueID, userID)
		if err != nil {
			return nil, err
		}
		if resolved == "" {
			return nil, errors.New("issue not found")
		}
		issueInternalID = &resolved
	}

	now := time.Now().UTC().Format(time.RFC3339)
	model := storemodels.ActiveContextModel{
		UserID:    userID,
		DeviceID:  deviceID,
		RepoID:    repoInternalID,
		StreamID:  streamInternalID,
		IssueID:   issueInternalID,
		UpdatedAt: now,
	}
	_, err := r.db.NewInsert().
		Model(&model).
		On("CONFLICT (user_id) DO UPDATE").
		Set("device_id = EXCLUDED.device_id").
		Set("repo_id = EXCLUDED.repo_id").
		Set("stream_id = EXCLUDED.stream_id").
		Set("issue_id = EXCLUDED.issue_id").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, userID, deviceID)
}

func (r *ActiveContextRepository) Clear(ctx context.Context, userID string, deviceID string) error {
	_, err := r.db.NewUpdate().
		Model((*storemodels.ActiveContextModel)(nil)).
		Where("user_id = ?", userID).
		Where("device_id = ?", deviceID).
		Set("repo_id = NULL").
		Set("stream_id = NULL").
		Set("issue_id = NULL").
		Set("updated_at = ?", time.Now().UTC().Format(time.RFC3339)).
		Exec(ctx)
	return err
}

func (r *ActiveContextRepository) InitializeDefaults(ctx context.Context, userID string, deviceID string) error {
	existing, err := r.Get(ctx, userID, deviceID)
	if err != nil || existing != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = r.db.NewInsert().
		Model(&storemodels.ActiveContextModel{
			UserID:    userID,
			DeviceID:  deviceID,
			UpdatedAt: now,
		}).
		Exec(ctx)
	return err
}
