package helpers

import (
	"slices"
	"strings"

	sharedtypes "crona/shared/types"
)

// ItemFilter is a pane filter split into its free text and its tag part. In the filter text,
// "+tag" keeps items with that tag (several are OR), "-tag" drops items with it, and the
// remaining words match the item's text.
type ItemFilter struct {
	Text string
	Tags sharedtypes.TagFilter
}

// ParseItemFilter splits a filter query into text and tags.
func ParseItemFilter(query string) ItemFilter {
	var filter ItemFilter
	words := []string{}
	for _, field := range strings.Fields(strings.ToLower(query)) {
		switch {
		case len(field) > 1 && field[0] == '+':
			filter.Tags.Any = append(filter.Tags.Any, field[1:])
		case len(field) > 1 && field[0] == '-':
			filter.Tags.None = append(filter.Tags.None, field[1:])
		default:
			words = append(words, field)
		}
	}
	filter.Text = strings.Join(words, " ")
	return filter
}

// Match reports whether an item with text and tags passes the filter.
func (f ItemFilter) Match(text string, tags []string) bool {
	if f.Text != "" && !strings.Contains(strings.ToLower(text), f.Text) {
		return false
	}
	return f.Tags.Matches(tags)
}

// TagFilterQuery is the tag part of a filter query, as "+a -b" tokens; it is what persists.
func TagFilterQuery(query string) string {
	tags := ParseItemFilter(query).Tags
	tokens := make([]string, 0, len(tags.Any)+len(tags.None))
	for _, name := range tags.Any {
		tokens = append(tokens, "+"+name)
	}
	for _, name := range tags.None {
		tokens = append(tokens, "-"+name)
	}
	return strings.Join(tokens, " ")
}

// PruneTagFilter drops the tag tokens in query that name none of the known tags.
func PruneTagFilter(query string, known []string) string {
	kept := []string{}
	for _, field := range strings.Fields(query) {
		if len(field) > 1 && (field[0] == '+' || field[0] == '-') &&
			field[1:] != sharedtypes.TagUntagged && !slices.Contains(known, strings.ToLower(field[1:])) {
			continue
		}
		kept = append(kept, field)
	}
	return strings.Join(kept, " ")
}

// TagChips renders tags as "+a +b" for compact rows.
func TagChips(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return "+" + strings.Join(tags, " +")
}
