package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	storemodels "crona/kernel/internal/store/models"
	sharedtypes "crona/shared/types"

	"github.com/uptrace/bun"
)

type StreamRepository struct {
	db *bun.DB
}

func NewStreamRepository(db *bun.DB) *StreamRepository {
	return &StreamRepository{db: db}
}

func (r *StreamRepository) NextID(ctx context.Context) (int64, error) {
	return nextPublicID(ctx, r.db, "streams")
}

func (r *StreamRepository) Create(ctx context.Context, stream sharedtypes.Stream, userID string, now string) (sharedtypes.Stream, error) {
	repoInternalID, err := resolveRepoInternalID(ctx, r.db, stream.RepoID, userID)
	if err != nil {
		return sharedtypes.Stream{}, err
	}
	if repoInternalID == "" {
		return sharedtypes.Stream{}, errors.New("repo not found")
	}

	model := storemodels.StreamModel{
		InternalID:  streamInternalID(stream.ID),
		PublicID:    stream.ID,
		RepoID:      repoInternalID,
		Name:        stream.Name,
		Description: stream.Description,
		Visibility:  string(stream.Visibility),
		UserID:      userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
		return sharedtypes.Stream{}, err
	}
	return stream, nil
}

func (r *StreamRepository) ListByRepo(ctx context.Context, repoID int64, userID string) ([]sharedtypes.Stream, error) {
	type row struct {
		PublicID    int64   `bun:"public_id"`
		RepoPublic  int64   `bun:"repo_public_id"`
		Name        string  `bun:"name"`
		Description *string `bun:"description"`
		Visibility  string  `bun:"visibility"`
	}

	var rows []row
	if err := r.db.NewSelect().
		TableExpr("streams").
		Join("INNER JOIN repos ON repos.id = streams.repo_id").
		ColumnExpr("streams.public_id").
		ColumnExpr("repos.public_id AS repo_public_id").
		ColumnExpr("streams.name").
		ColumnExpr("streams.description").
		ColumnExpr("streams.visibility").
		Where("repos.public_id = ?", repoID).
		Where("streams.user_id = ?", userID).
		Where("streams.deleted_at IS NULL").
		Where("repos.deleted_at IS NULL").
		OrderExpr("streams.created_at ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}

	out := make([]sharedtypes.Stream, 0, len(rows))
	for _, row := range rows {
		out = append(out, sharedtypes.Stream{
			ID:          row.PublicID,
			RepoID:      row.RepoPublic,
			Name:        row.Name,
			Description: row.Description,
			Visibility:  sharedtypes.StreamVisibility(row.Visibility),
		})
	}
	return out, nil
}

func (r *StreamRepository) GetByID(ctx context.Context, streamID int64, userID string) (*sharedtypes.Stream, error) {
	type row struct {
		PublicID    int64   `bun:"public_id"`
		RepoPublic  int64   `bun:"repo_public_id"`
		Name        string  `bun:"name"`
		Description *string `bun:"description"`
		Visibility  string  `bun:"visibility"`
	}

	var item row
	err := r.db.NewSelect().
		TableExpr("streams").
		Join("INNER JOIN repos ON repos.id = streams.repo_id").
		ColumnExpr("streams.public_id").
		ColumnExpr("repos.public_id AS repo_public_id").
		ColumnExpr("streams.name").
		ColumnExpr("streams.description").
		ColumnExpr("streams.visibility").
		Where("streams.public_id = ?", streamID).
		Where("streams.user_id = ?", userID).
		Where("streams.deleted_at IS NULL").
		Where("repos.deleted_at IS NULL").
		Limit(1).
		Scan(ctx, &item)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &sharedtypes.Stream{
		ID:          item.PublicID,
		RepoID:      item.RepoPublic,
		Name:        item.Name,
		Description: item.Description,
		Visibility:  sharedtypes.StreamVisibility(item.Visibility),
	}, nil
}

func (r *StreamRepository) ListDeletedByRepo(ctx context.Context, repoID int64, userID string) ([]sharedtypes.Stream, error) {
	type row struct {
		PublicID    int64   `bun:"public_id"`
		RepoPublic  int64   `bun:"repo_public_id"`
		Name        string  `bun:"name"`
		Description *string `bun:"description"`
		Visibility  string  `bun:"visibility"`
	}

	var rows []row
	if err := r.db.NewSelect().
		TableExpr("streams").
		Join("INNER JOIN repos ON repos.id = streams.repo_id").
		ColumnExpr("streams.public_id").
		ColumnExpr("repos.public_id AS repo_public_id").
		ColumnExpr("streams.name").
		ColumnExpr("streams.description").
		ColumnExpr("streams.visibility").
		Where("repos.public_id = ?", repoID).
		Where("streams.user_id = ?", userID).
		Where("streams.deleted_at IS NOT NULL").
		Where("repos.deleted_at IS NULL").
		OrderExpr("streams.created_at ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}

	out := make([]sharedtypes.Stream, 0, len(rows))
	for _, row := range rows {
		out = append(out, sharedtypes.Stream{
			ID:          row.PublicID,
			RepoID:      row.RepoPublic,
			Name:        row.Name,
			Description: row.Description,
			Visibility:  sharedtypes.StreamVisibility(row.Visibility),
		})
	}
	return out, nil
}

func (r *StreamRepository) Update(ctx context.Context, streamID int64, userID string, now string, updates struct {
	Name        *string
	Description sharedtypes.Patch[string]
	Visibility  *sharedtypes.StreamVisibility
},
) (*sharedtypes.Stream, error) {
	q := r.db.NewUpdate().
		Model((*storemodels.StreamModel)(nil)).
		Where("public_id = ?", streamID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("updated_at = ?", now)
	if updates.Name != nil {
		q = q.Set("name = ?", *updates.Name)
	}
	if updates.Description.Set {
		if updates.Description.Value == nil {
			q = q.Set("description = NULL")
		} else {
			q = q.Set("description = ?", *updates.Description.Value)
		}
	}
	if updates.Visibility != nil {
		q = q.Set("visibility = ?", string(*updates.Visibility))
	}
	res, err := q.Exec(ctx)
	if err != nil {
		return nil, err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, errors.New("stream not found or already deleted")
	}
	return r.GetByID(ctx, streamID, userID)
}

func (r *StreamRepository) SoftDeleteByRepo(ctx context.Context, repoID int64, userID string, now string) error {
	repoInternalID, err := resolveRepoInternalID(ctx, r.db, repoID, userID)
	if err != nil || repoInternalID == "" {
		return err
	}
	_, err = r.db.NewUpdate().
		Model((*storemodels.StreamModel)(nil)).
		Where("repo_id = ?", repoInternalID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Exec(ctx)
	return err
}

func (r *StreamRepository) SoftDelete(ctx context.Context, streamID int64, userID string, now string) error {
	res, err := r.db.NewUpdate().
		Model((*storemodels.StreamModel)(nil)).
		Where("public_id = ?", streamID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("stream not found or already deleted")
	}
	return nil
}

func (r *StreamRepository) Restore(ctx context.Context, streamID int64, userID string, now string) error {
	res, err := r.db.NewUpdate().
		Model((*storemodels.StreamModel)(nil)).
		Where("public_id = ?", streamID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NOT NULL").
		Set("deleted_at = NULL").
		Set("updated_at = ?", now).
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("stream not found or not deleted")
	}
	return nil
}

func (r *StreamRepository) RestoreByRepo(ctx context.Context, repoID int64, userID string, now string) error {
	repoInternalID, err := resolveRepoInternalID(ctx, r.db, repoID, userID)
	if err != nil || repoInternalID == "" {
		return err
	}
	_, err = r.db.NewUpdate().
		Model((*storemodels.StreamModel)(nil)).
		Where("repo_id = ?", repoInternalID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NOT NULL").
		Set("deleted_at = NULL").
		Set("updated_at = ?", now).
		Exec(ctx)
	return err
}

func streamInternalID(publicID int64) string {
	return fmt.Sprintf("stream-%d", publicID)
}
