package repositories

import (
	"context"
	"database/sql"
	"errors"

	sharedtypes "crona/shared/types"

	storemodels "crona/kernel/internal/store/models"
	"github.com/uptrace/bun"
)

type StashRepository struct {
	db *bun.DB
}

func NewStashRepository(db *bun.DB) *StashRepository {
	return &StashRepository{db: db}
}

func (r *StashRepository) List(ctx context.Context, userID string) ([]sharedtypes.Stash, error) {
	return r.listWithQuery(ctx, r.baseQuery().Where("stash.user_id = ?", userID).OrderExpr("stash.created_at DESC"))
}

func (r *StashRepository) ListByIssue(ctx context.Context, issueID int64, userID string) ([]sharedtypes.Stash, error) {
	return r.listWithQuery(ctx, r.baseQuery().
		Where("stash.user_id = ?", userID).
		Where("issues.public_id = ?", issueID).
		OrderExpr("stash.created_at DESC"))
}

func (r *StashRepository) listWithQuery(ctx context.Context, query *bun.SelectQuery) ([]sharedtypes.Stash, error) {
	type row struct {
		ID             string  `bun:"id"`
		UserID         string  `bun:"user_id"`
		DeviceID       string  `bun:"device_id"`
		RepoPublicID   *int64  `bun:"repo_public_id"`
		StreamPublicID *int64  `bun:"stream_public_id"`
		IssuePublicID  *int64  `bun:"issue_public_id"`
		SessionID      *string `bun:"session_id"`
		SegmentType    *string `bun:"segment_type"`
		ElapsedSeconds *int    `bun:"elapsed_seconds"`
		Note           *string `bun:"note"`
		CreatedAt      string  `bun:"created_at"`
		UpdatedAt      string  `bun:"updated_at"`
	}
	var rows []row
	if err := query.Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.Stash, 0, len(rows))
	for _, row := range rows {
		var segmentType *sharedtypes.SessionSegmentType
		if row.SegmentType != nil {
			value := sharedtypes.SessionSegmentType(*row.SegmentType)
			segmentType = &value
		}
		out = append(out, sharedtypes.Stash{
			ID:                row.ID,
			UserID:            row.UserID,
			DeviceID:          row.DeviceID,
			RepoID:            row.RepoPublicID,
			StreamID:          row.StreamPublicID,
			IssueID:           row.IssuePublicID,
			SessionID:         row.SessionID,
			PausedSegmentType: segmentType,
			ElapsedSeconds:    row.ElapsedSeconds,
			Note:              row.Note,
			CreatedAt:         row.CreatedAt,
			UpdatedAt:         row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *StashRepository) Get(ctx context.Context, id string, userID string) (*sharedtypes.Stash, error) {
	type row struct {
		ID             string  `bun:"id"`
		UserID         string  `bun:"user_id"`
		DeviceID       string  `bun:"device_id"`
		RepoPublicID   *int64  `bun:"repo_public_id"`
		StreamPublicID *int64  `bun:"stream_public_id"`
		IssuePublicID  *int64  `bun:"issue_public_id"`
		SessionID      *string `bun:"session_id"`
		SegmentType    *string `bun:"segment_type"`
		ElapsedSeconds *int    `bun:"elapsed_seconds"`
		Note           *string `bun:"note"`
		CreatedAt      string  `bun:"created_at"`
		UpdatedAt      string  `bun:"updated_at"`
	}
	var item row
	err := r.baseQuery().Where("stash.id = ?", id).Where("stash.user_id = ?", userID).Limit(1).Scan(ctx, &item)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var segmentType *sharedtypes.SessionSegmentType
	if item.SegmentType != nil {
		value := sharedtypes.SessionSegmentType(*item.SegmentType)
		segmentType = &value
	}
	out := sharedtypes.Stash{
		ID:                item.ID,
		UserID:            item.UserID,
		DeviceID:          item.DeviceID,
		RepoID:            item.RepoPublicID,
		StreamID:          item.StreamPublicID,
		IssueID:           item.IssuePublicID,
		SessionID:         item.SessionID,
		PausedSegmentType: segmentType,
		ElapsedSeconds:    item.ElapsedSeconds,
		Note:              item.Note,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
	return &out, nil
}

func (r *StashRepository) Save(ctx context.Context, stash sharedtypes.Stash) error {
	var (
		repoInternalID   *string
		streamInternalID *string
		issueInternalID  *string
	)
	if stash.RepoID != nil {
		resolved, err := resolveRepoInternalID(ctx, r.db, *stash.RepoID, stash.UserID)
		if err != nil {
			return err
		}
		repoInternalID = &resolved
	}
	if stash.StreamID != nil {
		resolved, err := resolveStreamInternalID(ctx, r.db, *stash.StreamID, stash.UserID)
		if err != nil {
			return err
		}
		streamInternalID = &resolved
	}
	if stash.IssueID != nil {
		resolved, err := resolveIssueInternalID(ctx, r.db, *stash.IssueID, stash.UserID)
		if err != nil {
			return err
		}
		issueInternalID = &resolved
	}

	var segmentType *string
	if stash.PausedSegmentType != nil {
		value := string(*stash.PausedSegmentType)
		segmentType = &value
	}
	_, err := r.db.NewInsert().Model(&storemodels.StashModel{
		ID:             stash.ID,
		UserID:         stash.UserID,
		DeviceID:       stash.DeviceID,
		RepoID:         repoInternalID,
		StreamID:       streamInternalID,
		IssueID:        issueInternalID,
		SessionID:      stash.SessionID,
		SegmentType:    segmentType,
		ElapsedSeconds: stash.ElapsedSeconds,
		Note:           stash.Note,
		CreatedAt:      stash.CreatedAt,
		UpdatedAt:      stash.UpdatedAt,
	}).Exec(ctx)
	return err
}

func (r *StashRepository) Delete(ctx context.Context, id string, userID string) error {
	_, err := r.db.NewDelete().Model((*storemodels.StashModel)(nil)).Where("id = ?", id).Where("user_id = ?", userID).Exec(ctx)
	return err
}

func (r *StashRepository) baseQuery() *bun.SelectQuery {
	return r.db.NewSelect().
		TableExpr("stash").
		Join("LEFT JOIN repos ON repos.id = stash.repo_id").
		Join("LEFT JOIN streams ON streams.id = stash.stream_id").
		Join("LEFT JOIN issues ON issues.id = stash.issue_id").
		ColumnExpr("stash.id").
		ColumnExpr("stash.user_id").
		ColumnExpr("stash.device_id").
		ColumnExpr("repos.public_id AS repo_public_id").
		ColumnExpr("streams.public_id AS stream_public_id").
		ColumnExpr("issues.public_id AS issue_public_id").
		ColumnExpr("stash.session_id").
		ColumnExpr("stash.segment_type").
		ColumnExpr("stash.elapsed_seconds").
		ColumnExpr("stash.note").
		ColumnExpr("stash.created_at").
		ColumnExpr("stash.updated_at")
}
