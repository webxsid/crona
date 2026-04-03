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

type IssueRepository struct {
	db *bun.DB
}

func NewIssueRepository(db *bun.DB) *IssueRepository {
	return &IssueRepository{db: db}
}

func (r *IssueRepository) NextID(ctx context.Context) (int64, error) {
	return nextPublicID(ctx, r.db, "issues")
}

func (r *IssueRepository) Create(ctx context.Context, issue sharedtypes.Issue, userID string, now string) (sharedtypes.Issue, error) {
	streamInternalID, err := resolveStreamInternalID(ctx, r.db, issue.StreamID, userID)
	if err != nil {
		return sharedtypes.Issue{}, err
	}
	if streamInternalID == "" {
		return sharedtypes.Issue{}, errors.New("stream not found")
	}

	model := storemodels.IssueModel{
		InternalID:      issueInternalID(issue.ID),
		PublicID:        issue.ID,
		StreamID:        streamInternalID,
		Title:           issue.Title,
		Description:     issue.Description,
		Status:          string(sharedtypes.NormalizeIssueStatus(issue.Status)),
		EstimateMinutes: issue.EstimateMinutes,
		Notes:           issue.Notes,
		TodoForDate:     issue.TodoForDate,
		CompletedAt:     issue.CompletedAt,
		AbandonedAt:     issue.AbandonedAt,
		UserID:          userID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
		return sharedtypes.Issue{}, err
	}
	return issue, nil
}

func (r *IssueRepository) ListByStream(ctx context.Context, streamID int64, userID string) ([]sharedtypes.Issue, error) {
	type row struct {
		PublicID        int64   `bun:"public_id"`
		StreamPublicID  int64   `bun:"stream_public_id"`
		Title           string  `bun:"title"`
		Description     *string `bun:"description"`
		Status          string  `bun:"status"`
		EstimateMinutes *int    `bun:"estimate_minutes"`
		Notes           *string `bun:"notes"`
		TodoForDate     *string `bun:"todo_for_date"`
		CompletedAt     *string `bun:"completed_at"`
		AbandonedAt     *string `bun:"abandoned_at"`
	}

	var rows []row
	if err := r.db.NewSelect().
		TableExpr("issues").
		Join("INNER JOIN streams ON streams.id = issues.stream_id").
		ColumnExpr("issues.public_id").
		ColumnExpr("streams.public_id AS stream_public_id").
		ColumnExpr("issues.title").
		ColumnExpr("issues.description").
		ColumnExpr("issues.status").
		ColumnExpr("issues.estimate_minutes").
		ColumnExpr("issues.notes").
		ColumnExpr("issues.todo_for_date").
		ColumnExpr("issues.completed_at").
		ColumnExpr("issues.abandoned_at").
		Where("streams.public_id = ?", streamID).
		Where("issues.user_id = ?", userID).
		Where("issues.deleted_at IS NULL").
		Where("streams.deleted_at IS NULL").
		OrderExpr("issues.created_at ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}

	out := make([]sharedtypes.Issue, 0, len(rows))
	for _, row := range rows {
		out = append(out, sharedtypes.Issue{
			ID:              row.PublicID,
			StreamID:        row.StreamPublicID,
			Title:           row.Title,
			Description:     row.Description,
			Status:          sharedtypes.NormalizeIssueStatus(sharedtypes.IssueStatus(row.Status)),
			EstimateMinutes: row.EstimateMinutes,
			Notes:           row.Notes,
			TodoForDate:     row.TodoForDate,
			CompletedAt:     row.CompletedAt,
			AbandonedAt:     row.AbandonedAt,
		})
	}
	return out, nil
}

func (r *IssueRepository) GetByID(ctx context.Context, issueID int64, userID string) (*sharedtypes.Issue, error) {
	type row struct {
		PublicID        int64   `bun:"public_id"`
		StreamPublicID  int64   `bun:"stream_public_id"`
		Title           string  `bun:"title"`
		Description     *string `bun:"description"`
		Status          string  `bun:"status"`
		EstimateMinutes *int    `bun:"estimate_minutes"`
		Notes           *string `bun:"notes"`
		TodoForDate     *string `bun:"todo_for_date"`
		CompletedAt     *string `bun:"completed_at"`
		AbandonedAt     *string `bun:"abandoned_at"`
	}

	var item row
	err := r.db.NewSelect().
		TableExpr("issues").
		Join("INNER JOIN streams ON streams.id = issues.stream_id").
		ColumnExpr("issues.public_id").
		ColumnExpr("streams.public_id AS stream_public_id").
		ColumnExpr("issues.title").
		ColumnExpr("issues.description").
		ColumnExpr("issues.status").
		ColumnExpr("issues.estimate_minutes").
		ColumnExpr("issues.notes").
		ColumnExpr("issues.todo_for_date").
		ColumnExpr("issues.completed_at").
		ColumnExpr("issues.abandoned_at").
		Where("issues.public_id = ?", issueID).
		Where("issues.user_id = ?", userID).
		Where("issues.deleted_at IS NULL").
		Where("streams.deleted_at IS NULL").
		Limit(1).
		Scan(ctx, &item)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sharedtypes.Issue{
		ID:              item.PublicID,
		StreamID:        item.StreamPublicID,
		Title:           item.Title,
		Description:     item.Description,
		Status:          sharedtypes.NormalizeIssueStatus(sharedtypes.IssueStatus(item.Status)),
		EstimateMinutes: item.EstimateMinutes,
		Notes:           item.Notes,
		TodoForDate:     item.TodoForDate,
		CompletedAt:     item.CompletedAt,
		AbandonedAt:     item.AbandonedAt,
	}, nil
}

func (r *IssueRepository) ResolveInternalID(ctx context.Context, issueID int64, userID string) (string, error) {
	return resolveIssueInternalID(ctx, r.db, issueID, userID)
}

func (r *IssueRepository) ResolvePublicID(ctx context.Context, issueInternalID string) (int64, error) {
	var publicID int64
	if err := r.db.NewSelect().
		TableExpr("issues").
		ColumnExpr("public_id").
		Where("id = ?", issueInternalID).
		Limit(1).
		Scan(ctx, &publicID); err != nil {
		return 0, err
	}
	return publicID, nil
}

func (r *IssueRepository) ListAll(ctx context.Context, userID string) ([]sharedtypes.IssueWithMeta, error) {
	type row struct {
		PublicID        int64   `bun:"public_id"`
		StreamPublicID  int64   `bun:"stream_public_id"`
		Title           string  `bun:"title"`
		Description     *string `bun:"description"`
		Status          string  `bun:"status"`
		EstimateMinutes *int    `bun:"estimate_minutes"`
		Notes           *string `bun:"notes"`
		TodoForDate     *string `bun:"todo_for_date"`
		CompletedAt     *string `bun:"completed_at"`
		AbandonedAt     *string `bun:"abandoned_at"`
		StreamName      string  `bun:"stream_name"`
		RepoPublicID    int64   `bun:"repo_public_id"`
		RepoName        string  `bun:"repo_name"`
	}
	var rows []row
	if err := r.db.NewSelect().
		TableExpr("issues").
		Join("INNER JOIN streams ON streams.id = issues.stream_id").
		Join("INNER JOIN repos ON repos.id = streams.repo_id").
		ColumnExpr("issues.public_id").
		ColumnExpr("streams.public_id AS stream_public_id").
		ColumnExpr("issues.title").
		ColumnExpr("issues.description").
		ColumnExpr("issues.status").
		ColumnExpr("issues.estimate_minutes").
		ColumnExpr("issues.notes").
		ColumnExpr("issues.todo_for_date").
		ColumnExpr("issues.completed_at").
		ColumnExpr("issues.abandoned_at").
		ColumnExpr("streams.name AS stream_name").
		ColumnExpr("repos.public_id AS repo_public_id").
		ColumnExpr("repos.name AS repo_name").
		Where("issues.user_id = ?", userID).
		Where("issues.deleted_at IS NULL").
		Where("streams.deleted_at IS NULL").
		Where("repos.deleted_at IS NULL").
		OrderExpr("issues.created_at ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.IssueWithMeta, 0, len(rows))
	for _, row := range rows {
		out = append(out, sharedtypes.IssueWithMeta{
			Issue: sharedtypes.Issue{
				ID:              row.PublicID,
				StreamID:        row.StreamPublicID,
				Title:           row.Title,
				Description:     row.Description,
				Status:          sharedtypes.NormalizeIssueStatus(sharedtypes.IssueStatus(row.Status)),
				EstimateMinutes: row.EstimateMinutes,
				Notes:           row.Notes,
				TodoForDate:     row.TodoForDate,
				CompletedAt:     row.CompletedAt,
				AbandonedAt:     row.AbandonedAt,
			},
			RepoID:     row.RepoPublicID,
			RepoName:   row.RepoName,
			StreamName: row.StreamName,
		})
	}
	return out, nil
}

func (r *IssueRepository) ListDeletedByStream(ctx context.Context, streamID int64, userID string) ([]sharedtypes.Issue, error) {
	type row struct {
		PublicID        int64   `bun:"public_id"`
		StreamPublicID  int64   `bun:"stream_public_id"`
		Title           string  `bun:"title"`
		Description     *string `bun:"description"`
		Status          string  `bun:"status"`
		EstimateMinutes *int    `bun:"estimate_minutes"`
		Notes           *string `bun:"notes"`
		TodoForDate     *string `bun:"todo_for_date"`
		CompletedAt     *string `bun:"completed_at"`
		AbandonedAt     *string `bun:"abandoned_at"`
	}
	var rows []row
	if err := r.db.NewSelect().
		TableExpr("issues").
		Join("INNER JOIN streams ON streams.id = issues.stream_id").
		ColumnExpr("issues.public_id").
		ColumnExpr("streams.public_id AS stream_public_id").
		ColumnExpr("issues.title").
		ColumnExpr("issues.description").
		ColumnExpr("issues.status").
		ColumnExpr("issues.estimate_minutes").
		ColumnExpr("issues.notes").
		ColumnExpr("issues.todo_for_date").
		ColumnExpr("issues.completed_at").
		ColumnExpr("issues.abandoned_at").
		Where("streams.public_id = ?", streamID).
		Where("issues.user_id = ?", userID).
		Where("issues.deleted_at IS NOT NULL").
		Where("streams.deleted_at IS NULL").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.Issue, 0, len(rows))
	for _, row := range rows {
		out = append(out, sharedtypes.Issue{
			ID:              row.PublicID,
			StreamID:        row.StreamPublicID,
			Title:           row.Title,
			Description:     row.Description,
			Status:          sharedtypes.NormalizeIssueStatus(sharedtypes.IssueStatus(row.Status)),
			EstimateMinutes: row.EstimateMinutes,
			Notes:           row.Notes,
			TodoForDate:     row.TodoForDate,
			CompletedAt:     row.CompletedAt,
			AbandonedAt:     row.AbandonedAt,
		})
	}
	return out, nil
}

func (r *IssueRepository) ListByTodoForDate(ctx context.Context, todoForDate string, userID string) ([]sharedtypes.Issue, error) {
	type row struct {
		PublicID        int64   `bun:"public_id"`
		StreamPublicID  int64   `bun:"stream_public_id"`
		Title           string  `bun:"title"`
		Description     *string `bun:"description"`
		Status          string  `bun:"status"`
		EstimateMinutes *int    `bun:"estimate_minutes"`
		Notes           *string `bun:"notes"`
		TodoForDate     *string `bun:"todo_for_date"`
		CompletedAt     *string `bun:"completed_at"`
		AbandonedAt     *string `bun:"abandoned_at"`
	}
	var rows []row
	if err := r.db.NewSelect().
		TableExpr("issues").
		Join("INNER JOIN streams ON streams.id = issues.stream_id").
		ColumnExpr("issues.public_id").
		ColumnExpr("streams.public_id AS stream_public_id").
		ColumnExpr("issues.title").
		ColumnExpr("issues.description").
		ColumnExpr("issues.status").
		ColumnExpr("issues.estimate_minutes").
		ColumnExpr("issues.notes").
		ColumnExpr("issues.todo_for_date").
		ColumnExpr("issues.completed_at").
		ColumnExpr("issues.abandoned_at").
		Where("issues.todo_for_date = ?", todoForDate).
		Where("issues.user_id = ?", userID).
		Where("issues.deleted_at IS NULL").
		OrderExpr("issues.created_at ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.Issue, 0, len(rows))
	for _, row := range rows {
		out = append(out, sharedtypes.Issue{
			ID:              row.PublicID,
			StreamID:        row.StreamPublicID,
			Title:           row.Title,
			Description:     row.Description,
			Status:          sharedtypes.NormalizeIssueStatus(sharedtypes.IssueStatus(row.Status)),
			EstimateMinutes: row.EstimateMinutes,
			Notes:           row.Notes,
			TodoForDate:     row.TodoForDate,
			CompletedAt:     row.CompletedAt,
			AbandonedAt:     row.AbandonedAt,
		})
	}
	return out, nil
}

func (r *IssueRepository) ListByTodoDateRange(ctx context.Context, startDate string, endDate string, userID string) ([]sharedtypes.Issue, error) {
	type row struct {
		PublicID        int64   `bun:"public_id"`
		StreamPublicID  int64   `bun:"stream_public_id"`
		Title           string  `bun:"title"`
		Description     *string `bun:"description"`
		Status          string  `bun:"status"`
		EstimateMinutes *int    `bun:"estimate_minutes"`
		Notes           *string `bun:"notes"`
		TodoForDate     *string `bun:"todo_for_date"`
		CompletedAt     *string `bun:"completed_at"`
		AbandonedAt     *string `bun:"abandoned_at"`
	}
	var rows []row
	if err := r.db.NewSelect().
		TableExpr("issues").
		Join("INNER JOIN streams ON streams.id = issues.stream_id").
		ColumnExpr("issues.public_id").
		ColumnExpr("streams.public_id AS stream_public_id").
		ColumnExpr("issues.title").
		ColumnExpr("issues.description").
		ColumnExpr("issues.status").
		ColumnExpr("issues.estimate_minutes").
		ColumnExpr("issues.notes").
		ColumnExpr("issues.todo_for_date").
		ColumnExpr("issues.completed_at").
		ColumnExpr("issues.abandoned_at").
		Where("issues.todo_for_date >= ?", startDate).
		Where("issues.todo_for_date <= ?", endDate).
		Where("issues.user_id = ?", userID).
		Where("issues.deleted_at IS NULL").
		OrderExpr("issues.todo_for_date ASC, issues.created_at ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.Issue, 0, len(rows))
	for _, row := range rows {
		out = append(out, sharedtypes.Issue{
			ID:              row.PublicID,
			StreamID:        row.StreamPublicID,
			Title:           row.Title,
			Description:     row.Description,
			Status:          sharedtypes.NormalizeIssueStatus(sharedtypes.IssueStatus(row.Status)),
			EstimateMinutes: row.EstimateMinutes,
			Notes:           row.Notes,
			TodoForDate:     row.TodoForDate,
			CompletedAt:     row.CompletedAt,
			AbandonedAt:     row.AbandonedAt,
		})
	}
	return out, nil
}

func (r *IssueRepository) Update(ctx context.Context, issueID int64, userID string, now string, updates struct {
	Title           Patch[string]
	Description     Patch[string]
	Status          Patch[sharedtypes.IssueStatus]
	EstimateMinutes Patch[int]
	Notes           Patch[string]
	TodoForDate     Patch[string]
	CompletedAt     Patch[string]
	AbandonedAt     Patch[string]
},
) (*sharedtypes.Issue, error) {
	q := r.db.NewUpdate().
		Model((*storemodels.IssueModel)(nil)).
		Where("public_id = ?", issueID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("updated_at = ?", now)
	if updates.Title.Set && updates.Title.Value != nil {
		q = q.Set("title = ?", *updates.Title.Value)
	}
	if updates.Description.Set {
		if updates.Description.Value == nil {
			q = q.Set("description = NULL")
		} else {
			q = q.Set("description = ?", *updates.Description.Value)
		}
	}
	if updates.Status.Set && updates.Status.Value != nil {
		q = q.Set("status = ?", string(sharedtypes.NormalizeIssueStatus(*updates.Status.Value)))
	}
	if updates.EstimateMinutes.Set {
		if updates.EstimateMinutes.Value == nil {
			q = q.Set("estimate_minutes = NULL")
		} else {
			q = q.Set("estimate_minutes = ?", *updates.EstimateMinutes.Value)
		}
	}
	if updates.Notes.Set {
		if updates.Notes.Value == nil {
			q = q.Set("notes = NULL")
		} else {
			q = q.Set("notes = ?", *updates.Notes.Value)
		}
	}
	if updates.TodoForDate.Set {
		if updates.TodoForDate.Value == nil {
			q = q.Set("todo_for_date = NULL")
		} else {
			q = q.Set("todo_for_date = ?", *updates.TodoForDate.Value)
		}
	}
	if updates.CompletedAt.Set {
		if updates.CompletedAt.Value == nil {
			q = q.Set("completed_at = NULL")
		} else {
			q = q.Set("completed_at = ?", *updates.CompletedAt.Value)
		}
	}
	if updates.AbandonedAt.Set {
		if updates.AbandonedAt.Value == nil {
			q = q.Set("abandoned_at = NULL")
		} else {
			q = q.Set("abandoned_at = ?", *updates.AbandonedAt.Value)
		}
	}
	res, err := q.Exec(ctx)
	if err != nil {
		return nil, err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, errors.New("issue not found or already deleted")
	}
	return r.GetByID(ctx, issueID, userID)
}

func (r *IssueRepository) SoftDelete(ctx context.Context, issueID int64, userID string, now string) error {
	res, err := r.db.NewUpdate().
		Model((*storemodels.IssueModel)(nil)).
		Where("public_id = ?", issueID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("issue not found or already deleted")
	}
	return nil
}

func (r *IssueRepository) CascadeSoftDeleteByStream(ctx context.Context, streamID int64, userID string, now string) error {
	streamInternalID, err := resolveStreamInternalID(ctx, r.db, streamID, userID)
	if err != nil || streamInternalID == "" {
		return err
	}
	_, err = r.db.NewUpdate().
		Model((*storemodels.IssueModel)(nil)).
		Where("stream_id = ?", streamInternalID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Exec(ctx)
	return err
}

func (r *IssueRepository) CascadeSoftDeleteByRepo(ctx context.Context, repoID int64, userID string, now string) error {
	repoInternalID, err := resolveRepoInternalID(ctx, r.db, repoID, userID)
	if err != nil || repoInternalID == "" {
		return err
	}
	_, err = r.db.NewUpdate().
		Model((*storemodels.IssueModel)(nil)).
		Where("stream_id IN (SELECT id FROM streams WHERE repo_id = ? AND user_id = ?)", repoInternalID, userID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Exec(ctx)
	return err
}

func (r *IssueRepository) RestoreDeletedByID(ctx context.Context, issueID int64, userID string, now string) error {
	res, err := r.db.NewUpdate().
		Model((*storemodels.IssueModel)(nil)).
		Where("public_id = ?", issueID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NOT NULL").
		Set("deleted_at = NULL").
		Set("updated_at = ?", now).
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("issue not found or not deleted")
	}
	return nil
}

func (r *IssueRepository) RestoreDeletedByStream(ctx context.Context, streamID int64, userID string, now string) error {
	streamInternalID, err := resolveStreamInternalID(ctx, r.db, streamID, userID)
	if err != nil || streamInternalID == "" {
		return err
	}
	_, err = r.db.NewUpdate().
		Model((*storemodels.IssueModel)(nil)).
		Where("stream_id = ?", streamInternalID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NOT NULL").
		Set("deleted_at = NULL").
		Set("updated_at = ?", now).
		Exec(ctx)
	return err
}

func (r *IssueRepository) RestoreDeletedByRepo(ctx context.Context, repoID int64, userID string, now string) error {
	repoInternalID, err := resolveRepoInternalID(ctx, r.db, repoID, userID)
	if err != nil || repoInternalID == "" {
		return err
	}
	_, err = r.db.NewUpdate().
		Model((*storemodels.IssueModel)(nil)).
		Where("stream_id IN (SELECT id FROM streams WHERE repo_id = ? AND user_id = ?)", repoInternalID, userID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NOT NULL").
		Set("deleted_at = NULL").
		Set("updated_at = ?", now).
		Exec(ctx)
	return err
}

func issueInternalID(publicID int64) string {
	return fmt.Sprintf("issue-%d", publicID)
}
