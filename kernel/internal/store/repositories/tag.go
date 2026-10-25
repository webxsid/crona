package repositories

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"

	storemodels "crona/kernel/internal/store/models"
	sharedtypes "crona/shared/types"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// TagRepository stores the tag vocabulary and which issues and habits carry each tag.
type TagRepository struct {
	db *bun.DB
}

func NewTagRepository(db *bun.DB) *TagRepository {
	return &TagRepository{db: db}
}

// TagKind picks the tagging table for an item kind.
type TagKind string

const (
	TagKindIssue TagKind = "issue"
	TagKindHabit TagKind = "habit"
)

func (k TagKind) table() string {
	if k == TagKindHabit {
		return "habit_tags"
	}
	return "issue_tags"
}

func (k TagKind) column() string {
	if k == TagKindHabit {
		return "habit_id"
	}
	return "issue_id"
}

func (k TagKind) internalID(publicID int64) string {
	if k == TagKindHabit {
		return habitInternalID(publicID)
	}
	return issueInternalID(publicID)
}

// tagNamesColumn selects an item's tag names, comma-joined, for list queries; splitTagNames
// turns the value back into a sorted list.
func tagNamesColumn(kind TagKind, itemColumn string) string {
	return "COALESCE((SELECT GROUP_CONCAT(tags.name, ',') FROM " + kind.table() +
		" AS it INNER JOIN tags ON tags.id = it.tag_id WHERE it." + kind.column() + " = " + itemColumn +
		"), '') AS tag_names"
}

func splitTagNames(joined string) []string {
	if joined == "" {
		return nil
	}
	names := strings.Split(joined, ",")
	slices.Sort(names)
	return names
}

// EnsureSystemTags adds any missing system tag, and marks an existing user tag of the same name
// as system in place.
func (r *TagRepository) EnsureSystemTags(ctx context.Context, userID, now string) error {
	for _, name := range sharedtypes.SystemTags {
		existing, err := r.get(ctx, r.db, userID, name)
		if err != nil {
			return err
		}
		if existing == nil {
			if _, err := r.db.NewInsert().Model(&storemodels.TagModel{
				ID: uuid.NewString(), UserID: userID, Name: name, System: true, CreatedAt: now, UpdatedAt: now,
			}).Exec(ctx); err != nil {
				return err
			}
			continue
		}
		if !existing.System {
			if _, err := r.db.NewUpdate().Model((*storemodels.TagModel)(nil)).
				Where("id = ?", existing.ID).Set("system = 1").Set("updated_at = ?", now).Exec(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// List returns the vocabulary in name order with how many live issues and habits use each tag.
func (r *TagRepository) List(ctx context.Context, userID string) ([]sharedtypes.Tag, error) {
	type row struct {
		Name       string  `bun:"name"`
		Color      *string `bun:"color"`
		System     bool    `bun:"system"`
		IssueCount int     `bun:"issue_count"`
		HabitCount int     `bun:"habit_count"`
	}
	var rows []row
	if err := r.db.NewSelect().
		TableExpr("tags").
		ColumnExpr("tags.name, tags.color, tags.system").
		ColumnExpr("(SELECT COUNT(*) FROM issue_tags AS it INNER JOIN issues ON issues.id = it.issue_id WHERE it.tag_id = tags.id AND issues.deleted_at IS NULL) AS issue_count").
		ColumnExpr("(SELECT COUNT(*) FROM habit_tags AS ht INNER JOIN habits ON habits.id = ht.habit_id WHERE ht.tag_id = tags.id AND habits.deleted_at IS NULL) AS habit_count").
		Where("tags.user_id = ?", userID).
		OrderExpr("tags.name ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]sharedtypes.Tag, 0, len(rows))
	for _, row := range rows {
		out = append(out, sharedtypes.Tag{
			Name: row.Name, Color: row.Color, System: row.System, IssueCount: row.IssueCount, HabitCount: row.HabitCount,
		})
	}
	return out, nil
}

// Get returns the tag named name, or nil.
func (r *TagRepository) Get(ctx context.Context, userID, name string) (*sharedtypes.Tag, error) {
	model, err := r.get(ctx, r.db, userID, name)
	if err != nil || model == nil {
		return nil, err
	}
	return &sharedtypes.Tag{Name: model.Name, Color: model.Color, System: model.System}, nil
}

func (r *TagRepository) get(ctx context.Context, db bun.IDB, userID, name string) (*storemodels.TagModel, error) {
	var model storemodels.TagModel
	err := db.NewSelect().Model(&model).Where("user_id = ?", userID).Where("name = ?", name).Limit(1).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// ensure returns the id of the tag named name, creating it first if needed.
func (r *TagRepository) ensure(ctx context.Context, db bun.IDB, userID, name, now string) (string, error) {
	var id string
	err := db.NewSelect().Model((*storemodels.TagModel)(nil)).Column("id").
		Where("user_id = ?", userID).Where("name = ?", name).Limit(1).Scan(ctx, &id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	id = uuid.NewString()
	_, err = db.NewInsert().Model(&storemodels.TagModel{
		ID: id, UserID: userID, Name: name, CreatedAt: now, UpdatedAt: now,
	}).Exec(ctx)
	return id, err
}

// ItemTags returns one item's tag names in name order.
func (r *TagRepository) ItemTags(ctx context.Context, kind TagKind, itemID int64) ([]string, error) {
	var names []string
	err := r.db.NewSelect().
		TableExpr(kind.table()+" AS it").
		Join("INNER JOIN tags ON tags.id = it.tag_id").
		ColumnExpr("tags.name").
		Where("it."+kind.column()+" = ?", kind.internalID(itemID)).
		OrderExpr("tags.name ASC").
		Scan(ctx, &names)
	return names, err
}

// SetItemTags replaces one item's tags with names (already normalized), creating new tags.
func (r *TagRepository) SetItemTags(ctx context.Context, kind TagKind, userID string, itemID int64, names []string, now string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		internalID := kind.internalID(itemID)
		if _, err := tx.NewDelete().TableExpr(kind.table()).Where(kind.column()+" = ?", internalID).Exec(ctx); err != nil {
			return err
		}
		for _, name := range names {
			tagID, err := r.ensure(ctx, tx, userID, name, now)
			if err != nil {
				return err
			}
			if _, err := tx.NewRaw(
				"INSERT OR IGNORE INTO "+kind.table()+" ("+kind.column()+", tag_id, user_id, created_at) VALUES (?, ?, ?, ?)",
				internalID, tagID, userID, now,
			).Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// Rename renames a tag. Renaming onto an existing tag merges the two: taggings move to the
// existing tag and the renamed one is removed.
func (r *TagRepository) Rename(ctx context.Context, userID, from, to, now string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		source, err := r.get(ctx, tx, userID, from)
		if err != nil {
			return err
		}
		if source == nil {
			return errors.New("tag not found")
		}
		var targetID string
		err = tx.NewSelect().Model((*storemodels.TagModel)(nil)).Column("id").
			Where("user_id = ?", userID).Where("name = ?", to).Limit(1).Scan(ctx, &targetID)
		if errors.Is(err, sql.ErrNoRows) {
			_, err = tx.NewUpdate().Model((*storemodels.TagModel)(nil)).Where("id = ?", source.ID).
				Set("name = ?", to).Set("updated_at = ?", now).Exec(ctx)
			return err
		}
		if err != nil {
			return err
		}
		for _, kind := range []TagKind{TagKindIssue, TagKindHabit} {
			if _, err := tx.NewRaw(
				"INSERT OR IGNORE INTO "+kind.table()+" ("+kind.column()+", tag_id, user_id, created_at) SELECT "+
					kind.column()+", ?, user_id, created_at FROM "+kind.table()+" WHERE tag_id = ?",
				targetID, source.ID,
			).Exec(ctx); err != nil {
				return err
			}
		}
		return r.delete(ctx, tx, source.ID)
	})
}

// SetColor sets or clears a tag's color.
func (r *TagRepository) SetColor(ctx context.Context, userID, name string, color *string, now string) error {
	_, err := r.db.NewUpdate().Model((*storemodels.TagModel)(nil)).
		Where("user_id = ?", userID).Where("name = ?", name).
		Set("color = ?", color).Set("updated_at = ?", now).Exec(ctx)
	return err
}

// Delete removes a tag and every tagging of it.
func (r *TagRepository) Delete(ctx context.Context, userID, name string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		model, err := r.get(ctx, tx, userID, name)
		if err != nil || model == nil {
			return err
		}
		return r.delete(ctx, tx, model.ID)
	})
}

func (r *TagRepository) delete(ctx context.Context, db bun.IDB, tagID string) error {
	for _, kind := range []TagKind{TagKindIssue, TagKindHabit} {
		if _, err := db.NewDelete().TableExpr(kind.table()).Where("tag_id = ?", tagID).Exec(ctx); err != nil {
			return err
		}
	}
	_, err := db.NewDelete().Model((*storemodels.TagModel)(nil)).Where("id = ?", tagID).Exec(ctx)
	return err
}

// DeleteUnused removes user tags that no live issue or habit uses, returning their names.
func (r *TagRepository) DeleteUnused(ctx context.Context, userID string) ([]string, error) {
	tags, err := r.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	removed := []string{}
	for _, tag := range tags {
		if tag.System || tag.IssueCount > 0 || tag.HabitCount > 0 {
			continue
		}
		if err := r.Delete(ctx, userID, tag.Name); err != nil {
			return nil, err
		}
		removed = append(removed, tag.Name)
	}
	return removed, nil
}
