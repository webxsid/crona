package tagsmeta

import (
	"fmt"
	"slices"

	sharedtypes "crona/shared/types"
	viewhelpers "crona/tui/internal/tui/views/helpers"
)

// Usage says how many issues and habits carry a tag.
func Usage(tag sharedtypes.Tag) string {
	return plural(tag.IssueCount, "issue") + ", " + plural(tag.HabitCount, "habit")
}

// Unused counts the tags Remove unused would delete: unused and not system.
func Unused(tags []sharedtypes.Tag) int {
	count := 0
	for _, tag := range tags {
		if !tag.System && tag.IssueCount+tag.HabitCount == 0 {
			count++
		}
	}
	return count
}

// FilteredIndices are the tags whose names match the pane filter, system tags first.
func FilteredIndices(tags []sharedtypes.Tag, filter string) []int {
	indices := viewhelpers.FilteredStrings(Names(tags), filter)
	slices.SortStableFunc(indices, func(a, b int) int {
		if tags[a].System == tags[b].System {
			return 0
		}
		if tags[a].System {
			return -1
		}
		return 1
	})
	return indices
}

func Names(tags []sharedtypes.Tag) []string {
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	return names
}

// Selected is the tag under the cursor of the filtered list.
func Selected(tags []sharedtypes.Tag, filter string, cursor int) (sharedtypes.Tag, bool) {
	indices := FilteredIndices(tags, filter)
	if cursor < 0 || cursor >= len(indices) {
		return sharedtypes.Tag{}, false
	}
	return tags[indices[cursor]], true
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// TaggedIssues are the issues carrying a tag, open ones first.
func TaggedIssues(name string, issues []sharedtypes.IssueWithMeta) []sharedtypes.IssueWithMeta {
	out := []sharedtypes.IssueWithMeta{}
	for _, issue := range issues {
		if slices.Contains(issue.Tags, name) {
			out = append(out, issue)
		}
	}
	slices.SortStableFunc(out, func(a, b sharedtypes.IssueWithMeta) int {
		return boolOrder(IssueClosed(a.Status), IssueClosed(b.Status))
	})
	return out
}

// TaggedHabits are the habits carrying a tag, active ones first.
func TaggedHabits(name string, habits []sharedtypes.HabitWithMeta) []sharedtypes.HabitWithMeta {
	out := []sharedtypes.HabitWithMeta{}
	for _, habit := range habits {
		if slices.Contains(habit.Tags, name) {
			out = append(out, habit)
		}
	}
	slices.SortStableFunc(out, func(a, b sharedtypes.HabitWithMeta) int {
		return boolOrder(!a.Active, !b.Active)
	})
	return out
}

// OpenDone splits a tag's issues into open and done or abandoned.
func OpenDone(name string, issues []sharedtypes.IssueWithMeta) (open, closed int) {
	for _, issue := range issues {
		if !slices.Contains(issue.Tags, name) {
			continue
		}
		if IssueClosed(issue.Status) {
			closed++
		} else {
			open++
		}
	}
	return open, closed
}

func IssueClosed(status sharedtypes.IssueStatus) bool {
	return status == sharedtypes.IssueStatusDone || status == sharedtypes.IssueStatusAbandoned
}

// Untagged counts the open issues and active habits with no tags.
func Untagged(issues []sharedtypes.IssueWithMeta, habits []sharedtypes.HabitWithMeta) (openIssues, activeHabits int) {
	for _, issue := range issues {
		if len(issue.Tags) == 0 && !IssueClosed(issue.Status) {
			openIssues++
		}
	}
	for _, habit := range habits {
		if len(habit.Tags) == 0 && habit.Active {
			activeHabits++
		}
	}
	return openIssues, activeHabits
}

func boolOrder(a, b bool) int {
	switch {
	case a == b:
		return 0
	case a:
		return 1
	default:
		return -1
	}
}
