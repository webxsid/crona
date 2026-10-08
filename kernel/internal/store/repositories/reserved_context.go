package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	storemodels "crona/kernel/internal/store/models"
	sharedtypes "crona/shared/types"

	"github.com/uptrace/bun"
)

const (
	adhocStreamName        = "adhoc-sessions"
	reservedRepoNamePrefix = "crona-reserved-"
)

type ReservedContext struct {
	RepoID   int64
	StreamID int64
	IssueID  int64
}

type ReservedContextRepository struct {
	db *bun.DB
}

func NewReservedContextRepository(db *bun.DB) *ReservedContextRepository {
	return &ReservedContextRepository{db: db}
}

func (r *ReservedContextRepository) EnsureAdhocContext(
	ctx context.Context,
	userID string,
	now string,
) (ReservedContext, error) {
	if strings.TrimSpace(userID) == "" {
		return ReservedContext{}, errors.New("user id is required")
	}

	var out ReservedContext
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		repoName := reservedRepoName(userID)
		var repo storemodels.RepoModel
		err := tx.NewSelect().
			Model(&repo).
			Where("name = ?", repoName).
			Where("user_id = ?", userID).
			Limit(1).
			Scan(ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("find adhoc repo: %w", err)
		}

		if errors.Is(err, sql.ErrNoRows) {
			publicID, err := nextPublicIDTx(ctx, tx, "repos")
			if err != nil {
				return fmt.Errorf("allocate adhoc repo id: %w", err)
			}
			repo = storemodels.RepoModel{
				InternalID:  repoInternalID(publicID),
				PublicID:    publicID,
				Name:        repoName,
				Description: stringPointer("Reserved repo for adhoc sessions"),
				IsReserved:  true,
				UserID:      userID,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if _, err := tx.NewInsert().Model(&repo).Exec(ctx); err != nil {
				return fmt.Errorf("create adhoc repo: %w", err)
			}
		} else if !repo.IsReserved {
			return errors.New("adhoc repo name is already used by a normal repo")
		} else if repo.DeletedAt != nil {
			if _, err := tx.NewUpdate().Model((*storemodels.RepoModel)(nil)).
				Where("id = ?", repo.InternalID).
				Set("deleted_at = NULL").
				Set("updated_at = ?", now).
				Exec(ctx); err != nil {
				return fmt.Errorf("restore adhoc repo: %w", err)
			}
		}

		var stream storemodels.StreamModel
		err = tx.NewSelect().
			Model(&stream).
			Where("name = ?", adhocStreamName).
			Where("repo_id = ?", repo.InternalID).
			Where("user_id = ?", userID).
			Limit(1).
			Scan(ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("find adhoc stream: %w", err)
		}

		if errors.Is(err, sql.ErrNoRows) {
			publicID, err := nextPublicIDTx(ctx, tx, "streams")
			if err != nil {
				return fmt.Errorf("allocate adhoc stream id: %w", err)
			}
			stream = storemodels.StreamModel{
				InternalID:  streamInternalID(publicID),
				PublicID:    publicID,
				RepoID:      repo.InternalID,
				Name:        adhocStreamName,
				Description: stringPointer("Reserved stream for adhoc sessions"),
				Visibility:  string(sharedtypes.StreamVisibilityPersonal),
				IsReserved:  true,
				UserID:      userID,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if _, err := tx.NewInsert().Model(&stream).Exec(ctx); err != nil {
				return fmt.Errorf("create adhoc stream: %w", err)
			}
		} else if !stream.IsReserved {
			return errors.New("adhoc stream name is already used by a normal stream")
		} else if stream.DeletedAt != nil {
			if _, err := tx.NewUpdate().Model((*storemodels.StreamModel)(nil)).
				Where("id = ?", stream.InternalID).
				Set("deleted_at = NULL").
				Set("updated_at = ?", now).
				Exec(ctx); err != nil {
				return fmt.Errorf("restore adhoc stream: %w", err)
			}
		}

		var issue storemodels.IssueModel
		err = tx.NewSelect().Model(&issue).
			Where("stream_id = ?", stream.InternalID).
			Where("title = ?", "Ad hoc pending").
			Where("user_id = ?", userID).
			Where("is_reserved = 1").
			Limit(1).
			Scan(ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("find adhoc issue: %w", err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			publicID, err := nextPublicIDTx(ctx, tx, "issues")
			if err != nil {
				return fmt.Errorf("allocate adhoc issue id: %w", err)
			}
			issue = storemodels.IssueModel{
				InternalID: issueInternalID(publicID),
				PublicID:   publicID,
				StreamID:   stream.InternalID,
				Title:      "Ad hoc pending",
				Status:     string(sharedtypes.IssueStatusBacklog),
				IsReserved: true,
				UserID:     userID,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			if _, err := tx.NewInsert().Model(&issue).Exec(ctx); err != nil {
				return fmt.Errorf("create adhoc issue: %w", err)
			}
		} else if issue.DeletedAt != nil {
			if _, err := tx.NewUpdate().Model((*storemodels.IssueModel)(nil)).
				Where("id = ?", issue.InternalID).
				Set("deleted_at = NULL").
				Set("updated_at = ?", now).
				Exec(ctx); err != nil {
				return fmt.Errorf("restore adhoc issue: %w", err)
			}
		}

		out = ReservedContext{RepoID: repo.PublicID, StreamID: stream.PublicID, IssueID: issue.PublicID}
		return nil
	})
	if err != nil {
		return ReservedContext{}, err
	}
	return out, nil
}

func nextPublicIDTx(ctx context.Context, tx bun.Tx, table string) (int64, error) {
	var maxID *int64
	if err := tx.NewSelect().TableExpr(table).ColumnExpr("MAX(public_id)").Scan(ctx, &maxID); err != nil {
		return 0, err
	}
	if maxID == nil {
		return 1, nil
	}
	return *maxID + 1, nil
}

func reservedRepoName(userID string) string {
	return reservedRepoNamePrefix + userID
}

func stringPointer(value string) *string {
	return &value
}
