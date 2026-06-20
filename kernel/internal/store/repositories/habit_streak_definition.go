package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	storemodels "crona/kernel/internal/store/models"
	sharedtypes "crona/shared/types"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type HabitStreakDefinitionRepository struct {
	db *bun.DB
}

type habitStreakDefinitionRow struct {
	ID            string  `bun:"id"`
	Name          string  `bun:"name"`
	Description   *string `bun:"description"`
	Enabled       bool    `bun:"enabled"`
	TargetKind    string  `bun:"target_kind"`
	MatchMode     string  `bun:"match_mode"`
	RepoID        *int64  `bun:"repo_id"`
	StreamID      *int64  `bun:"stream_id"`
	Contexts      string  `bun:"contexts"`
	Period        string  `bun:"period"`
	RequiredCount int     `bun:"required_count"`
	HabitID       *int64  `bun:"habit_id"`
}

func NewHabitStreakDefinitionRepository(db *bun.DB) *HabitStreakDefinitionRepository {
	return &HabitStreakDefinitionRepository{db: db}
}

func (r *HabitStreakDefinitionRepository) List(
	ctx context.Context,
	userID string,
) ([]sharedtypes.HabitStreakDefinition, error) {
	rows, err := r.selectDefinitionRows(ctx, userID, "")
	if err != nil {
		return nil, err
	}
	return assembleHabitStreakDefinitions(rows), nil
}

func (r *HabitStreakDefinitionRepository) GetByID(
	ctx context.Context,
	userID string,
	id string,
) (*sharedtypes.HabitStreakDefinition, error) {
	rows, err := r.selectDefinitionRows(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	defs := assembleHabitStreakDefinitions(rows)
	if len(defs) == 0 {
		return nil, nil
	}
	return &defs[0], nil
}

func (r *HabitStreakDefinitionRepository) selectDefinitionRows(
	ctx context.Context,
	userID string,
	id string,
) ([]habitStreakDefinitionRow, error) {
	var rows []habitStreakDefinitionRow
	query := r.db.NewSelect().
		TableExpr("momentums AS m").
		ColumnExpr("m.id").
		ColumnExpr("m.name").
		ColumnExpr("m.description").
		ColumnExpr("m.enabled").
		ColumnExpr("m.target_kind").
		ColumnExpr("m.match_mode").
		ColumnExpr("m.repo_id").
		ColumnExpr("m.stream_id").
		ColumnExpr("m.contexts").
		ColumnExpr("m.period").
		ColumnExpr("m.required_count").
		ColumnExpr("mh.habit_id").
		Join("LEFT JOIN momentum_habits AS mh ON mh.momentum_id = m.id AND mh.user_id = m.user_id").
		Where("m.user_id = ?", userID).
		Where("m.deleted_at IS NULL")
	if strings.TrimSpace(id) != "" {
		query = query.Where("m.id = ?", id).OrderExpr("mh.habit_id ASC")
	} else {
		query = query.OrderExpr("m.created_at ASC, mh.habit_id ASC")
	}
	if err := query.Scan(ctx, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func assembleHabitStreakDefinitions(rows []habitStreakDefinitionRow) []sharedtypes.HabitStreakDefinition {
	defsByID := make(map[string]*sharedtypes.HabitStreakDefinition, len(rows))
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		def, ok := defsByID[row.ID]
		if !ok {
			item := sharedtypes.HabitStreakDefinition{
				ID:            row.ID,
				Name:          row.Name,
				Description:   row.Description,
				Enabled:       row.Enabled,
				TargetKind:    sharedtypes.MomentumTargetKind(row.TargetKind),
				MatchMode:     sharedtypes.MomentumMatchMode(row.MatchMode),
				Contexts:      parseMomentumContexts(row.Contexts, row.RepoID, row.StreamID),
				Period:        sharedtypes.HabitStreakPeriod(row.Period),
				RequiredCount: row.RequiredCount,
			}
			def = &item
			defsByID[row.ID] = def
			order = append(order, row.ID)
		}
		if row.HabitID != nil {
			def.HabitIDs = append(def.HabitIDs, *row.HabitID)
		}
	}
	defs := make([]sharedtypes.HabitStreakDefinition, 0, len(order))
	for _, rowID := range order {
		defs = append(defs, sharedtypes.NormalizeHabitStreakDefinition(*defsByID[rowID]))
	}
	return sharedtypes.NormalizeHabitStreakDefinitions(defs)
}

func (r *HabitStreakDefinitionRepository) Create(
	ctx context.Context,
	userID string,
	now string,
	def sharedtypes.HabitStreakDefinition,
) (sharedtypes.HabitStreakDefinition, error) {
	def = normalizeStoredHabitStreakDefinition(def)
	if def.ID == "" {
		def.ID = uuid.NewString()
	}
	model := storemodels.HabitStreakDefinitionModel{
		ID:            def.ID,
		UserID:        userID,
		Name:          def.Name,
		Description:   def.Description,
		Enabled:       def.Enabled,
		TargetKind:    string(def.TargetKind),
		MatchMode:     string(def.MatchMode),
		Contexts:      mustJSONContexts(def.Contexts),
		Period:        string(def.Period),
		RequiredCount: def.RequiredCount,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := insertHabitStreakDefinitionRow(ctx, tx, model); err != nil {
			return err
		}
		if sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) != sharedtypes.MomentumTargetKindHabit {
			return clearHabitStreakHabitLinks(ctx, tx, userID, def.ID)
		}
		return replaceHabitStreakHabitLinks(ctx, tx, userID, def.ID, def.HabitIDs, now)
	}); err != nil {
		return sharedtypes.HabitStreakDefinition{}, err
	}
	return def, nil
}

func (r *HabitStreakDefinitionRepository) Update(
	ctx context.Context,
	userID string,
	now string,
	def sharedtypes.HabitStreakDefinition,
) (*sharedtypes.HabitStreakDefinition, error) {
	def = normalizeStoredHabitStreakDefinition(def)
	if strings.TrimSpace(def.ID) == "" {
		return nil, errors.New("habit streak id is required")
	}
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		query := tx.NewUpdate().
			Model((*storemodels.HabitStreakDefinitionModel)(nil)).
			Where("id = ?", def.ID).
			Where("user_id = ?", userID).
			Where("deleted_at IS NULL").
			Set("name = ?", def.Name).
			Set("enabled = ?", def.Enabled).
			Set("target_kind = ?", string(def.TargetKind)).
			Set("match_mode = ?", string(def.MatchMode)).
			Set("contexts = ?", mustJSONContexts(def.Contexts)).
			Set("period = ?", string(def.Period)).
			Set("required_count = ?", def.RequiredCount).
			Set("updated_at = ?", now)
		if def.Description == nil {
			query = query.Set("description = NULL")
		} else {
			query = query.Set("description = ?", *def.Description)
		}
		res, err := query.Exec(ctx)
		if err != nil {
			return err
		}
		if rows, _ := res.RowsAffected(); rows == 0 {
			return errors.New("habit streak not found")
		}
		if sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) != sharedtypes.MomentumTargetKindHabit {
			return clearHabitStreakHabitLinks(ctx, tx, userID, def.ID)
		}
		return replaceHabitStreakHabitLinks(ctx, tx, userID, def.ID, def.HabitIDs, now)
	})
	if err != nil {
		return nil, err
	}
	return &def, nil
}

func (r *HabitStreakDefinitionRepository) Delete(
	ctx context.Context,
	userID string,
	id string,
	now string,
) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		res, err := tx.NewUpdate().
			Model((*storemodels.HabitStreakDefinitionModel)(nil)).
			Where("id = ?", id).
			Where("user_id = ?", userID).
			Where("deleted_at IS NULL").
			Set("deleted_at = ?", now).
			Set("updated_at = ?", now).
			Exec(ctx)
		if err != nil {
			return err
		}
		if rows, _ := res.RowsAffected(); rows == 0 {
			return errors.New("habit streak not found")
		}
		_, err = tx.NewDelete().
			Model((*storemodels.HabitStreakHabitModel)(nil)).
			Where("momentum_id = ?", id).
			Where("user_id = ?", userID).
			Exec(ctx)
		return err
	})
}

func (r *HabitStreakDefinitionRepository) ReplaceAll(
	ctx context.Context,
	userID string,
	now string,
	defs []sharedtypes.HabitStreakDefinition,
) error {
	defs = sharedtypes.NormalizeHabitStreakDefinitions(defs)
	existing, err := r.List(ctx, userID)
	if err != nil {
		return err
	}
	existingByID := make(map[string]sharedtypes.HabitStreakDefinition, len(existing))
	for _, def := range existing {
		existingByID[def.ID] = def
	}
	nextByID := make(map[string]sharedtypes.HabitStreakDefinition, len(defs))
	for i, def := range defs {
		def = normalizeStoredHabitStreakDefinition(def)
		if def.ID == "" {
			def.ID = uuid.NewString()
		}
		defs[i] = def
		nextByID[def.ID] = def
	}

	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, def := range defs {
			if _, ok := existingByID[def.ID]; ok {
				query := tx.NewUpdate().
					Model((*storemodels.HabitStreakDefinitionModel)(nil)).
					Where("id = ?", def.ID).
					Where("user_id = ?", userID).
					Set("name = ?", def.Name).
					Set("enabled = ?", def.Enabled).
					Set("target_kind = ?", string(def.TargetKind)).
					Set("match_mode = ?", string(def.MatchMode)).
					Set("contexts = ?", mustJSONContexts(def.Contexts)).
					Set("period = ?", string(def.Period)).
					Set("required_count = ?", def.RequiredCount).
					Set("updated_at = ?", now).
					Set("deleted_at = NULL")
				if def.Description == nil {
					query = query.Set("description = NULL")
				} else {
					query = query.Set("description = ?", *def.Description)
				}
				res, err := query.Exec(ctx)
				if err != nil {
					return err
				}
				if rows, _ := res.RowsAffected(); rows == 0 {
					return errors.New("habit streak not found")
				}
			} else {
				model := storemodels.HabitStreakDefinitionModel{
					ID:            def.ID,
					UserID:        userID,
					Name:          def.Name,
					Description:   def.Description,
					TargetKind:    string(def.TargetKind),
					MatchMode:     string(def.MatchMode),
					Contexts:      mustJSONContexts(def.Contexts),
					Enabled:       def.Enabled,
					Period:        string(def.Period),
					RequiredCount: def.RequiredCount,
					CreatedAt:     now,
					UpdatedAt:     now,
				}
				if err := insertHabitStreakDefinitionRow(ctx, tx, model); err != nil {
					return err
				}
			}
			if sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) != sharedtypes.MomentumTargetKindHabit {
				if err := clearHabitStreakHabitLinks(ctx, tx, userID, def.ID); err != nil {
					return err
				}
				continue
			}
			if err := replaceHabitStreakHabitLinks(ctx, tx, userID, def.ID, def.HabitIDs, now); err != nil {
				return err
			}
		}
		for _, def := range existing {
			if _, ok := nextByID[def.ID]; ok {
				continue
			}
			if _, err := tx.NewUpdate().
				Model((*storemodels.HabitStreakDefinitionModel)(nil)).
				Where("id = ?", def.ID).
				Where("user_id = ?", userID).
				Where("deleted_at IS NULL").
				Set("deleted_at = ?", now).
				Set("updated_at = ?", now).
				Exec(ctx); err != nil {
				return err
			}
			if _, err := tx.NewDelete().
				Model((*storemodels.HabitStreakHabitModel)(nil)).
				Where("momentum_id = ?", def.ID).
				Where("user_id = ?", userID).
				Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

func HasHabitStreakDefinitions(ctx context.Context, db bun.IDB, userID string) (bool, error) {
	var count int
	if err := db.NewSelect().
		Model((*storemodels.HabitStreakDefinitionModel)(nil)).
		ColumnExpr("COUNT(*)").
		Where("user_id = ?", userID).
		Scan(ctx, &count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func normalizeStoredHabitStreakDefinition(
	def sharedtypes.HabitStreakDefinition,
) sharedtypes.HabitStreakDefinition {
	def = sharedtypes.NormalizeHabitStreakDefinition(def)
	switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		def.HabitIDs = nil
	default:
		def.TargetKind = sharedtypes.MomentumTargetKindHabit
		def.Contexts = nil
	}
	def.Name = strings.TrimSpace(def.Name)
	return def
}

func replaceHabitStreakHabitLinks(
	ctx context.Context,
	db bun.IDB,
	userID string,
	momentumID string,
	habitIDs []int64,
	now string,
) error {
	if _, err := db.NewDelete().
		Model((*storemodels.HabitStreakHabitModel)(nil)).
		Where("momentum_id = ?", momentumID).
		Where("user_id = ?", userID).
		Exec(ctx); err != nil {
		return err
	}
	habitIDs = sharedtypes.NormalizeHabitStreakDefinition(
		sharedtypes.HabitStreakDefinition{HabitIDs: habitIDs},
	).HabitIDs
	if len(habitIDs) == 0 {
		return nil
	}
	rows := make([]storemodels.HabitStreakHabitModel, 0, len(habitIDs))
	for _, habitID := range habitIDs {
		rows = append(rows, storemodels.HabitStreakHabitModel{
			MomentumID: momentumID,
			HabitID:    habitID,
			UserID:     userID,
			CreatedAt:  now,
		})
	}
	_, err := db.NewInsert().Model(&rows).Exec(ctx)
	return err
}

func clearHabitStreakHabitLinks(
	ctx context.Context,
	db bun.IDB,
	userID string,
	momentumID string,
) error {
	_, err := db.NewDelete().
		Model((*storemodels.HabitStreakHabitModel)(nil)).
		Where("momentum_id = ?", momentumID).
		Where("user_id = ?", userID).
		Exec(ctx)
	return err
}

func insertHabitStreakDefinitionRow(
	ctx context.Context,
	db bun.IDB,
	model storemodels.HabitStreakDefinitionModel,
) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO momentums
			(id, user_id, name, description, enabled, target_kind, match_mode, contexts, period, required_count, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		model.ID,
		model.UserID,
		model.Name,
		model.Description,
		model.Enabled,
		model.TargetKind,
		model.MatchMode,
		model.Contexts,
		model.Period,
		model.RequiredCount,
		model.CreatedAt,
		model.UpdatedAt,
	)
	return err
}

func mustJSONContexts(values []sharedtypes.MomentumContext) string {
	payload, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func parseMomentumContexts(raw string, repoID, streamID *int64) []sharedtypes.MomentumContext {
	raw = strings.TrimSpace(raw)
	if raw != "" {
		var contexts []sharedtypes.MomentumContext
		if err := json.Unmarshal([]byte(raw), &contexts); err == nil {
			return sharedtypes.NormalizeHabitStreakDefinition(sharedtypes.HabitStreakDefinition{Contexts: contexts}).Contexts
		}
	}
	if repoID != nil && *repoID > 0 {
		var streamPtr *int64
		if streamID != nil && *streamID > 0 {
			value := *streamID
			streamPtr = &value
		}
		return []sharedtypes.MomentumContext{{RepoID: *repoID, StreamID: streamPtr}}
	}
	return nil
}
