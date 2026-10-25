package commands

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"crona/kernel/internal/core"
	"crona/kernel/internal/store/repositories"
	sharedtypes "crona/shared/types"

	"github.com/google/uuid"
)

// EnsureSystemTags puts the shipped system tags in the vocabulary.
func EnsureSystemTags(ctx context.Context, c *core.Context) error {
	return c.Tags.EnsureSystemTags(ctx, c.UserID, c.Now())
}

// ListTags returns the tag vocabulary with usage counts.
func ListTags(ctx context.Context, c *core.Context) ([]sharedtypes.Tag, error) {
	if err := EnsureSystemTags(ctx, c); err != nil {
		return nil, err
	}
	return c.Tags.List(ctx, c.UserID)
}

// SetIssueTags replaces an issue's tags.
func SetIssueTags(ctx context.Context, c *core.Context, issueID int64, names []string) (*sharedtypes.Issue, error) {
	issue, err := c.Issues.GetByID(ctx, issueID, c.UserID)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, errors.New("issue not found")
	}
	if err := setItemTags(ctx, c, repositories.TagKindIssue, sharedtypes.OpEntityIssue, issueID, names); err != nil {
		return nil, err
	}
	updated, err := c.Issues.GetByID(ctx, issueID, c.UserID)
	if err != nil {
		return nil, err
	}
	emit(c, sharedtypes.EventTypeIssueUpdated, updated)
	return updated, nil
}

// UpdateIssueTags adds and removes tags across several issues at once.
func UpdateIssueTags(ctx context.Context, c *core.Context, issueIDs []int64, add, remove []string) ([]sharedtypes.Issue, error) {
	out := make([]sharedtypes.Issue, 0, len(issueIDs))
	for _, id := range issueIDs {
		current, err := c.Tags.ItemTags(ctx, repositories.TagKindIssue, id)
		if err != nil {
			return nil, err
		}
		next, err := editTagSet(current, add, remove)
		if err != nil {
			return nil, err
		}
		updated, err := SetIssueTags(ctx, c, id, next)
		if err != nil {
			return nil, err
		}
		out = append(out, *updated)
	}
	return out, nil
}

// SetHabitTags replaces a habit's tags.
func SetHabitTags(ctx context.Context, c *core.Context, habitID int64, names []string) (*sharedtypes.Habit, error) {
	habit, err := c.Habits.GetByID(ctx, habitID, c.UserID)
	if err != nil {
		return nil, err
	}
	if habit == nil {
		return nil, errors.New("habit not found")
	}
	if err := setItemTags(ctx, c, repositories.TagKindHabit, sharedtypes.OpEntityHabit, habitID, names); err != nil {
		return nil, err
	}
	updated, err := c.Habits.GetByID(ctx, habitID, c.UserID)
	if err != nil {
		return nil, err
	}
	emit(c, sharedtypes.EventTypeHabitUpdated, updated)
	return updated, nil
}

// UpdateHabitTags adds and removes tags across several habits at once.
func UpdateHabitTags(ctx context.Context, c *core.Context, habitIDs []int64, add, remove []string) ([]sharedtypes.Habit, error) {
	out := make([]sharedtypes.Habit, 0, len(habitIDs))
	for _, id := range habitIDs {
		current, err := c.Tags.ItemTags(ctx, repositories.TagKindHabit, id)
		if err != nil {
			return nil, err
		}
		next, err := editTagSet(current, add, remove)
		if err != nil {
			return nil, err
		}
		updated, err := SetHabitTags(ctx, c, id, next)
		if err != nil {
			return nil, err
		}
		out = append(out, *updated)
	}
	return out, nil
}

// RenameTag renames a tag. Renaming onto an existing name merges the two, which the caller
// must confirm with merge. System tags can be neither renamed nor merged into.
func RenameTag(ctx context.Context, c *core.Context, from, to string, merge bool) error {
	from, err := sharedtypes.NormalizeTagName(from)
	if err != nil {
		return err
	}
	to, err = sharedtypes.NormalizeTagName(to)
	if err != nil {
		return err
	}
	if from == to {
		return nil
	}
	if sharedtypes.IsSystemTag(from) {
		return fmt.Errorf("%q is a system tag and cannot be renamed", from)
	}
	if sharedtypes.IsSystemTag(to) {
		return fmt.Errorf("%q is a system tag; other tags cannot be renamed or merged onto it", to)
	}
	existing, err := c.Tags.Get(ctx, c.UserID, to)
	if err != nil {
		return err
	}
	if existing != nil && !merge {
		return fmt.Errorf("tag %q already exists; confirm to merge %q into it", to, from)
	}
	now := c.Now()
	if err := c.Tags.Rename(ctx, c.UserID, from, to, now); err != nil {
		return err
	}
	return recordTagChange(ctx, c, from, now, sharedtypes.OpActionUpdate, map[string]any{"renamedTo": to, "merged": existing != nil})
}

// SetTagColor sets a tag's palette color, or clears it when color is nil or empty.
func SetTagColor(ctx context.Context, c *core.Context, name string, color *string) error {
	name, err := sharedtypes.NormalizeTagName(name)
	if err != nil {
		return err
	}
	if color != nil && *color == "" {
		color = nil
	}
	if color != nil && !sharedtypes.IsTagColor(*color) {
		return fmt.Errorf("unknown tag color %q", *color)
	}
	existing, err := c.Tags.Get(ctx, c.UserID, name)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("tag not found")
	}
	now := c.Now()
	if err := c.Tags.SetColor(ctx, c.UserID, name, color, now); err != nil {
		return err
	}
	return recordTagChange(ctx, c, name, now, sharedtypes.OpActionUpdate, map[string]any{"color": color})
}

// DeleteTag removes a user tag from every issue and habit and from the vocabulary.
func DeleteTag(ctx context.Context, c *core.Context, name string) error {
	name, err := sharedtypes.NormalizeTagName(name)
	if err != nil {
		return err
	}
	if sharedtypes.IsSystemTag(name) {
		return fmt.Errorf("%q is a system tag and cannot be deleted", name)
	}
	now := c.Now()
	if err := c.Tags.Delete(ctx, c.UserID, name); err != nil {
		return err
	}
	return recordTagChange(ctx, c, name, now, sharedtypes.OpActionDelete, map[string]any{"deleted": true})
}

// RemoveUnusedTags deletes user tags no issue or habit uses and returns their names.
func RemoveUnusedTags(ctx context.Context, c *core.Context) ([]string, error) {
	removed, err := c.Tags.DeleteUnused(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	now := c.Now()
	for _, name := range removed {
		if err := recordTagChange(ctx, c, name, now, sharedtypes.OpActionDelete, map[string]any{"deleted": true}); err != nil {
			return nil, err
		}
	}
	return removed, nil
}

func setItemTags(
	ctx context.Context,
	c *core.Context,
	kind repositories.TagKind,
	entity sharedtypes.OpEntity,
	itemID int64,
	names []string,
) error {
	normalized, err := sharedtypes.NormalizeTagNames(names)
	if err != nil {
		return err
	}
	now := c.Now()
	if err := c.Tags.SetItemTags(ctx, kind, c.UserID, itemID, normalized, now); err != nil {
		return err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    entity,
		EntityID:  fmt.Sprintf("%d", itemID),
		Action:    sharedtypes.OpActionUpdate,
		Payload:   map[string]any{"tags": normalized},
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return err
	}
	emit(c, sharedtypes.EventTypeTagsChanged, map[string]any{})
	return nil
}

// editTagSet applies add and remove to current, normalizing the names.
func editTagSet(current, add, remove []string) ([]string, error) {
	adds, err := sharedtypes.NormalizeTagNames(add)
	if err != nil {
		return nil, err
	}
	removes, err := sharedtypes.NormalizeTagNames(remove)
	if err != nil {
		return nil, err
	}
	next := []string{}
	for _, name := range append(slices.Clone(current), adds...) {
		if !slices.Contains(removes, name) && !slices.Contains(next, name) {
			next = append(next, name)
		}
	}
	slices.Sort(next)
	return next, nil
}

func recordTagChange(
	ctx context.Context,
	c *core.Context,
	name, now string,
	action sharedtypes.OpAction,
	payload map[string]any,
) error {
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityTag,
		EntityID:  name,
		Action:    action,
		Payload:   payload,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return err
	}
	emit(c, sharedtypes.EventTypeTagsChanged, payload)
	return nil
}
