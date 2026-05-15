package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"crona/kernel/internal/sessionnotes"
	sharedtypes "crona/shared/types"

	storemodels "crona/kernel/internal/store/models"
	"github.com/uptrace/bun"
)

type SessionRepository struct {
	db *bun.DB
}

func NewSessionRepository(db *bun.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Start(ctx context.Context, session sharedtypes.Session, userID string, deviceID string, now string) (sharedtypes.Session, error) {
	issueInternalID, err := resolveIssueInternalID(ctx, r.db, session.IssueID, userID)
	if err != nil {
		return sharedtypes.Session{}, err
	}
	if issueInternalID == "" {
		return sharedtypes.Session{}, errors.New("issue not found")
	}

	model := storemodels.SessionModel{
		ID:        session.ID,
		IssueID:   issueInternalID,
		Source:    string(sessionSourceOrDefault(session.Source)),
		StartTime: session.StartTime,
		Notes:     session.Notes,
		UserID:    userID,
		DeviceID:  deviceID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
		return sharedtypes.Session{}, err
	}
	session.Source = sessionSourceOrDefault(session.Source)
	return session, nil
}

func (r *SessionRepository) CreateCompleted(ctx context.Context, session sharedtypes.Session, userID string, deviceID string, now string) (sharedtypes.Session, error) {
	issueInternalID, err := resolveIssueInternalID(ctx, r.db, session.IssueID, userID)
	if err != nil {
		return sharedtypes.Session{}, err
	}
	if issueInternalID == "" {
		return sharedtypes.Session{}, errors.New("issue not found")
	}
	model := storemodels.SessionModel{
		ID:              session.ID,
		IssueID:         issueInternalID,
		Source:          string(sessionSourceOrDefault(session.Source)),
		StartTime:       session.StartTime,
		EndTime:         session.EndTime,
		DurationSeconds: session.DurationSeconds,
		Notes:           session.Notes,
		UserID:          userID,
		DeviceID:        deviceID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
		return sharedtypes.Session{}, err
	}
	session.Source = sessionSourceOrDefault(session.Source)
	return session, nil
}

func (r *SessionRepository) Stop(ctx context.Context, sessionID string, updates struct {
	EndTime         string
	DurationSeconds int
	Notes           *string
}, userID string, deviceID string, now string,
) (*sharedtypes.Session, error) {
	q := r.db.NewUpdate().
		Model((*storemodels.SessionModel)(nil)).
		Where("id = ?", sessionID).
		Where("user_id = ?", userID).
		Where("end_time IS NULL").
		Where("deleted_at IS NULL").
		Set("end_time = ?", updates.EndTime).
		Set("duration_seconds = ?", updates.DurationSeconds).
		Set("updated_at = ?", now).
		Set("device_id = ?", deviceID)
	if updates.Notes != nil {
		q = q.Set("notes = ?", *updates.Notes)
	} else {
		q = q.Set("notes = NULL")
	}
	res, err := q.Exec(ctx)
	if err != nil {
		return nil, err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, errors.New("active session not found")
	}
	return r.GetByID(ctx, sessionID, userID)
}

func (r *SessionRepository) GetActiveSession(ctx context.Context, userID string) (*sharedtypes.Session, error) {
	return r.selectOne(ctx, r.db.NewSelect().
		TableExpr("sessions").
		Join("INNER JOIN issues ON issues.id = sessions.issue_id").
		ColumnExpr("sessions.id").
		ColumnExpr("issues.public_id AS issue_public_id").
		ColumnExpr("sessions.source").
		ColumnExpr("sessions.start_time").
		ColumnExpr("sessions.notes").
		Where("sessions.user_id = ?", userID).
		Where("sessions.end_time IS NULL").
		Where("sessions.deleted_at IS NULL").
		OrderExpr("sessions.start_time DESC").
		Limit(1))
}

func (r *SessionRepository) GetByID(ctx context.Context, sessionID string, userID string) (*sharedtypes.Session, error) {
	return r.selectOne(ctx, r.db.NewSelect().
		TableExpr("sessions").
		Join("INNER JOIN issues ON issues.id = sessions.issue_id").
		ColumnExpr("sessions.id").
		ColumnExpr("issues.public_id AS issue_public_id").
		ColumnExpr("sessions.source").
		ColumnExpr("sessions.start_time").
		ColumnExpr("sessions.end_time").
		ColumnExpr("sessions.duration_seconds").
		ColumnExpr("sessions.notes").
		Where("sessions.id = ?", sessionID).
		Where("sessions.user_id = ?", userID).
		Where("sessions.deleted_at IS NULL").
		Limit(1))
}

func (r *SessionRepository) GetDetail(ctx context.Context, sessionID string, userID string) (*sharedtypes.SessionDetail, error) {
	type row struct {
		ID              string  `bun:"id"`
		IssuePublicID   int64   `bun:"issue_public_id"`
		Source          *string `bun:"source"`
		IssueTitle      string  `bun:"issue_title"`
		StreamPublicID  int64   `bun:"stream_public_id"`
		StreamName      string  `bun:"stream_name"`
		RepoPublicID    int64   `bun:"repo_public_id"`
		RepoName        string  `bun:"repo_name"`
		StartTime       string  `bun:"start_time"`
		EndTime         *string `bun:"end_time"`
		DurationSeconds *int    `bun:"duration_seconds"`
		Notes           *string `bun:"notes"`
	}

	var item row
	err := r.db.NewSelect().
		TableExpr("sessions").
		Join("INNER JOIN issues ON issues.id = sessions.issue_id").
		Join("INNER JOIN streams ON streams.id = issues.stream_id").
		Join("INNER JOIN repos ON repos.id = streams.repo_id").
		ColumnExpr("sessions.id").
		ColumnExpr("issues.public_id AS issue_public_id").
		ColumnExpr("sessions.source").
		ColumnExpr("issues.title AS issue_title").
		ColumnExpr("streams.public_id AS stream_public_id").
		ColumnExpr("streams.name AS stream_name").
		ColumnExpr("repos.public_id AS repo_public_id").
		ColumnExpr("repos.name AS repo_name").
		ColumnExpr("sessions.start_time").
		ColumnExpr("sessions.end_time").
		ColumnExpr("sessions.duration_seconds").
		ColumnExpr("sessions.notes").
		Where("sessions.id = ?", sessionID).
		Where("sessions.user_id = ?", userID).
		Where("sessions.deleted_at IS NULL").
		Limit(1).
		Scan(ctx, &item)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &sharedtypes.SessionDetail{
		SessionHistoryEntry: sharedtypes.SessionHistoryEntry{
			Session: sharedtypes.Session{
				ID:              item.ID,
				IssueID:         item.IssuePublicID,
				Source:          sessionSourceFromStore(item.Source),
				StartTime:       item.StartTime,
				EndTime:         item.EndTime,
				DurationSeconds: item.DurationSeconds,
				Notes:           item.Notes,
			},
		},
		RepoID:     item.RepoPublicID,
		RepoName:   item.RepoName,
		StreamID:   item.StreamPublicID,
		StreamName: item.StreamName,
		IssueTitle: item.IssueTitle,
	}, nil
}

func (r *SessionRepository) ListByIssue(ctx context.Context, issueID int64, userID string) ([]sharedtypes.Session, error) {
	type row struct {
		ID              string  `bun:"id"`
		IssuePublicID   int64   `bun:"issue_public_id"`
		Source          *string `bun:"source"`
		StartTime       string  `bun:"start_time"`
		EndTime         *string `bun:"end_time"`
		DurationSeconds *int    `bun:"duration_seconds"`
		Notes           *string `bun:"notes"`
	}

	var rows []row
	if err := r.db.NewSelect().
		TableExpr("sessions").
		Join("INNER JOIN issues ON issues.id = sessions.issue_id").
		ColumnExpr("sessions.id").
		ColumnExpr("issues.public_id AS issue_public_id").
		ColumnExpr("sessions.source").
		ColumnExpr("sessions.start_time").
		ColumnExpr("sessions.end_time").
		ColumnExpr("sessions.duration_seconds").
		ColumnExpr("sessions.notes").
		Where("issues.public_id = ?", issueID).
		Where("sessions.user_id = ?", userID).
		Where("sessions.deleted_at IS NULL").
		OrderExpr("sessions.start_time ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}

	out := make([]sharedtypes.Session, 0, len(rows))
	for _, row := range rows {
		out = append(out, sharedtypes.Session{
			ID:              row.ID,
			IssueID:         row.IssuePublicID,
			Source:          sessionSourceFromStore(row.Source),
			StartTime:       row.StartTime,
			EndTime:         row.EndTime,
			DurationSeconds: row.DurationSeconds,
			Notes:           row.Notes,
		})
	}
	return out, nil
}

func (r *SessionRepository) GetLastSessionForIssue(ctx context.Context, issueID int64, userID string) (*sharedtypes.Session, error) {
	return r.selectOne(ctx, r.db.NewSelect().
		TableExpr("sessions").
		Join("INNER JOIN issues ON issues.id = sessions.issue_id").
		ColumnExpr("sessions.id").
		ColumnExpr("issues.public_id AS issue_public_id").
		ColumnExpr("sessions.source").
		ColumnExpr("sessions.start_time").
		ColumnExpr("sessions.end_time").
		ColumnExpr("sessions.duration_seconds").
		ColumnExpr("sessions.notes").
		Where("issues.public_id = ?", issueID).
		Where("sessions.user_id = ?", userID).
		Where("sessions.deleted_at IS NULL").
		OrderExpr("sessions.start_time DESC").
		Limit(1))
}

func (r *SessionRepository) GetLastSessionForUser(ctx context.Context, userID string) (*sharedtypes.Session, error) {
	return r.selectOne(ctx, r.db.NewSelect().
		TableExpr("sessions").
		Join("INNER JOIN issues ON issues.id = sessions.issue_id").
		ColumnExpr("sessions.id").
		ColumnExpr("issues.public_id AS issue_public_id").
		ColumnExpr("sessions.source").
		ColumnExpr("sessions.start_time").
		ColumnExpr("sessions.end_time").
		ColumnExpr("sessions.duration_seconds").
		ColumnExpr("sessions.notes").
		Where("sessions.user_id = ?", userID).
		Where("sessions.deleted_at IS NULL").
		OrderExpr("sessions.start_time DESC").
		Limit(1))
}

func (r *SessionRepository) AmendSessionNotes(ctx context.Context, sessionID string, notes string, userID string, deviceID string, now string) (*sharedtypes.Session, error) {
	res, err := r.db.NewUpdate().
		Model((*storemodels.SessionModel)(nil)).
		Where("id = ?", sessionID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Set("notes = ?", notes).
		Set("updated_at = ?", now).
		Set("device_id = ?", deviceID).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, errors.New("session not found for ammending notes")
	}
	return r.GetByID(ctx, sessionID, userID)
}

func (r *SessionRepository) ListEnded(ctx context.Context, input struct {
	UserID   string
	RepoID   *int64
	StreamID *int64
	IssueID  *int64
	Since    *string
	Until    *string
	Limit    *int
	Offset   *int
},
) ([]sharedtypes.SessionHistoryEntry, error) {
	q := r.db.NewSelect().
		TableExpr("sessions").
		Join("INNER JOIN issues ON issues.id = sessions.issue_id").
		Join("INNER JOIN streams ON streams.id = issues.stream_id").
		Join("INNER JOIN repos ON repos.id = streams.repo_id").
		ColumnExpr("sessions.id").
		ColumnExpr("issues.public_id AS issue_public_id").
		ColumnExpr("sessions.source").
		ColumnExpr("sessions.start_time").
		ColumnExpr("sessions.end_time").
		ColumnExpr("sessions.duration_seconds").
		ColumnExpr("sessions.notes").
		Where("sessions.user_id = ?", input.UserID).
		Where("sessions.end_time IS NOT NULL").
		Where("sessions.deleted_at IS NULL").
		OrderExpr("sessions.start_time DESC")

	if input.RepoID != nil {
		q = q.Where("repos.public_id = ?", *input.RepoID)
	}
	if input.StreamID != nil {
		q = q.Where("streams.public_id = ?", *input.StreamID)
	}
	if input.IssueID != nil {
		q = q.Where("issues.public_id = ?", *input.IssueID)
	}
	if input.Since != nil {
		q = q.Where("sessions.start_time >= ?", *input.Since)
	}
	if input.Until != nil {
		q = q.Where("sessions.start_time <= ?", *input.Until)
	}
	if input.Limit != nil {
		q = q.Limit(*input.Limit)
	}
	if input.Offset != nil {
		q = q.Offset(*input.Offset)
	}

	type row struct {
		ID              string  `bun:"id"`
		IssuePublicID   int64   `bun:"issue_public_id"`
		Source          *string `bun:"source"`
		StartTime       string  `bun:"start_time"`
		EndTime         *string `bun:"end_time"`
		DurationSeconds *int    `bun:"duration_seconds"`
		Notes           *string `bun:"notes"`
	}
	var rows []row
	if err := q.Scan(ctx, &rows); err != nil {
		return nil, err
	}

	out := make([]sharedtypes.SessionHistoryEntry, 0, len(rows))
	for _, row := range rows {
		session := sharedtypes.Session{
			ID:              row.ID,
			IssueID:         row.IssuePublicID,
			Source:          sessionSourceFromStore(row.Source),
			StartTime:       row.StartTime,
			EndTime:         row.EndTime,
			DurationSeconds: row.DurationSeconds,
			Notes:           row.Notes,
		}
		out = append(out, sharedtypes.SessionHistoryEntry{
			Session:     session,
			ParsedNotes: sessionnotes.Parse(row.Notes),
		})
	}
	return out, nil
}

func (r *SessionRepository) EarliestEndedDate(ctx context.Context, userID string, throughDate string) (*string, error) {
	type row struct {
		Date string `bun:"date"`
	}
	var item row
	err := r.db.NewSelect().
		TableExpr("sessions").
		ColumnExpr("substr(start_time, 1, 10) AS date").
		Where("user_id = ?", userID).
		Where("end_time IS NOT NULL").
		Where("deleted_at IS NULL").
		Where("substr(start_time, 1, 10) <= ?", throughDate).
		OrderExpr("start_time ASC").
		Limit(1).
		Scan(ctx, &item)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if item.Date == "" {
		return nil, nil
	}
	return &item.Date, nil
}

func (r *SessionRepository) selectOne(ctx context.Context, q *bun.SelectQuery) (*sharedtypes.Session, error) {
	type row struct {
		ID              string  `bun:"id"`
		IssuePublicID   int64   `bun:"issue_public_id"`
		Source          *string `bun:"source"`
		StartTime       string  `bun:"start_time"`
		EndTime         *string `bun:"end_time"`
		DurationSeconds *int    `bun:"duration_seconds"`
		Notes           *string `bun:"notes"`
	}
	var item row
	if err := q.Scan(ctx, &item); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	session := sharedtypes.Session{
		ID:              item.ID,
		IssueID:         item.IssuePublicID,
		Source:          sessionSourceFromStore(item.Source),
		StartTime:       item.StartTime,
		EndTime:         item.EndTime,
		DurationSeconds: item.DurationSeconds,
		Notes:           item.Notes,
	}
	return &session, nil
}

func sessionSourceFromStore(raw *string) sharedtypes.SessionSource {
	if raw == nil {
		return sharedtypes.SessionSourceTracked
	}
	source := sharedtypes.SessionSource(strings.TrimSpace(*raw))
	if source == sharedtypes.SessionSourceManual {
		return source
	}
	return sharedtypes.SessionSourceTracked
}

func sessionSourceOrDefault(source sharedtypes.SessionSource) sharedtypes.SessionSource {
	if source == sharedtypes.SessionSourceManual {
		return source
	}
	return sharedtypes.SessionSourceTracked
}
