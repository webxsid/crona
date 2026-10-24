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
