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

type RepoRepository struct {
	db *bun.DB
}

func NewRepoRepository(db *bun.DB) *RepoRepository {
	return &RepoRepository{db: db}
}

func (r *RepoRepository) Create(ctx context.Context, repo sharedtypes.Repo, userID string, now string) (sharedtypes.Repo, error) {
	model := storemodels.RepoModel{
		InternalID:  repoInternalID(repo.ID),
		PublicID:    repo.ID,
		Name:        repo.Name,
		Description: repo.Description,
		Color:       repo.Color,
		UserID:      userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
		return sharedtypes.Repo{}, err
	}
	return repo, nil
}

func (r *RepoRepository) NextID(ctx context.Context) (int64, error) {
	return nextPublicID(ctx, r.db, "repos")
}

func (r *RepoRepository) GetByID(ctx context.Context, repoID int64, userID string) (*sharedtypes.Repo, error) {
	var model storemodels.RepoModel
	err := r.db.NewSelect().
		Model(&model).
		Column("public_id", "name", "description", "color").
		Where("public_id = ?", repoID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sharedtypes.Repo{ID: model.PublicID, Name: model.Name, Description: model.Description, Color: model.Color}, nil
}

func (r *RepoRepository) List(ctx context.Context, userID string) ([]sharedtypes.Repo, error) {
	var models []storemodels.RepoModel
	if err := r.db.NewSelect().
		Model(&models).
		Column("public_id", "name", "description", "color").
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Order("created_at ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	out := make([]sharedtypes.Repo, 0, len(models))
	for _, model := range models {
		out = append(out, sharedtypes.Repo{ID: model.PublicID, Name: model.Name, Description: model.Description, Color: model.Color})
	}
	return out, nil
}

func (r *RepoRepository) ListDeleted(ctx context.Context, userID string) ([]sharedtypes.Repo, error) {
	var models []storemodels.RepoModel
	if err := r.db.NewSelect().
		Model(&models).
		Column("public_id", "name", "description", "color").
		Where("user_id = ?", userID).
		Where("deleted_at IS NOT NULL").
		Order("created_at ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	out := make([]sharedtypes.Repo, 0, len(models))
	for _, model := range models {
		out = append(out, sharedtypes.Repo{ID: model.PublicID, Name: model.Name, Description: model.Description, Color: model.Color})
	}
	return out, nil
}

func (r *RepoRepository) Update(ctx context.Context, repoID int64, userID string, now string, updates struct {
	Name        Patch[string]
	Description Patch[string]
	Color       Patch[string]
},
) (*sharedtypes.Repo, error) {
	q := r.db.NewUpdate().
		Model((*storemodels.RepoModel)(nil)).
		Where("public_id = ?", repoID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("updated_at = ?", now)

	if updates.Name.Set && updates.Name.Value != nil {
		q = q.Set("name = ?", *updates.Name.Value)
	}
	if updates.Description.Set {
		if updates.Description.Value == nil {
			q = q.Set("description = NULL")
		} else {
			q = q.Set("description = ?", *updates.Description.Value)
		}
	}
	if updates.Color.Set {
		if updates.Color.Value == nil {
			q = q.Set("color = NULL")
		} else {
			q = q.Set("color = ?", *updates.Color.Value)
		}
	}

	res, err := q.Exec(ctx)
	if err != nil {
		return nil, err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, errors.New("repo not found or already deleted")
	}
	return r.GetByID(ctx, repoID, userID)
}

func (r *RepoRepository) SoftDelete(ctx context.Context, repoID int64, userID string, now string) error {
	res, err := r.db.NewUpdate().
		Model((*storemodels.RepoModel)(nil)).
		Where("public_id = ?", repoID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("repo not found or already deleted")
	}
	return nil
}

func (r *RepoRepository) Restore(ctx context.Context, repoID int64, userID string, now string) error {
	res, err := r.db.NewUpdate().
		Model((*storemodels.RepoModel)(nil)).
		Where("public_id = ?", repoID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NOT NULL").
		Set("deleted_at = NULL").
		Set("updated_at = ?", now).
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("repo not found or not deleted")
	}
	return nil
}

func repoInternalID(publicID int64) string {
	return fmt.Sprintf("repo-%d", publicID)
}
