package types

import (
	"errors"
	"slices"
	"strings"
)

// TagUntagged is the filter value that matches items with no tags. It cannot be a tag name.
const TagUntagged = "untagged"

// MaxTagNameLength caps a tag name, in characters.
const MaxTagNameLength = 32

// SystemTags ship with Crona. They are always in the vocabulary and cannot be deleted or
// renamed; only their color changes.
var SystemTags = []string{"urgent", "waiting", "quick", "review", "someday"}

// TagColors are the palette names a tag may be colored with.
var TagColors = []string{"blue", "cyan", "green", "magenta", "yellow", "red", "orange", "lavender"}

// Tag is one entry in the tag vocabulary, with how many items use it.
type Tag struct {
	Name       string  `json:"name"`
	Color      *string `json:"color,omitempty"`
	System     bool    `json:"system"`
	IssueCount int     `json:"issueCount"`
	HabitCount int     `json:"habitCount"`
}

// TagFilter keeps items that have at least one Any tag (when Any is set) and none of the None
// tags. TagUntagged in either list stands for "no tags at all".
type TagFilter struct {
	Any  []string `json:"any,omitempty"`
	None []string `json:"none,omitempty"`
}

// Empty reports whether the filter keeps everything.
func (f TagFilter) Empty() bool {
	return len(f.Any) == 0 && len(f.None) == 0
}

// Matches reports whether an item with tags passes the filter.
func (f TagFilter) Matches(tags []string) bool {
	has := func(name string) bool {
		if name == TagUntagged {
			return len(tags) == 0
		}
		return slices.Contains(tags, name)
	}
	if len(f.Any) > 0 && !slices.ContainsFunc(f.Any, has) {
		return false
	}
	return !slices.ContainsFunc(f.None, has)
}

// IsSystemTag reports whether name is one of the shipped system tags.
func IsSystemTag(name string) bool {
	return slices.Contains(SystemTags, name)
}

// NormalizeTagName lowercases name, turns spaces into dashes and strips a leading "+", then
// checks it: 1–32 characters of letters, digits, "-", "_" and "/", starting with a letter or
// digit, and not the reserved "untagged".
func NormalizeTagName(name string) (string, error) {
	name = strings.TrimPrefix(strings.TrimSpace(name), "+")
	name = strings.Join(strings.Fields(strings.ToLower(name)), "-")
	if name == "" {
		return "", errors.New("tag name is empty")
	}
	if len([]rune(name)) > MaxTagNameLength {
		return "", errors.New("tag name is longer than 32 characters")
	}
	for i, r := range name {
		alnum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if i == 0 && !alnum {
			return "", errors.New("tag name must start with a letter or digit")
		}
		if !alnum && r != '-' && r != '_' && r != '/' {
			return "", errors.New("tag name may only use letters, digits, -, _ and /")
		}
	}
	if name == TagUntagged {
		return "", errors.New(`"untagged" is reserved for filtering`)
	}
	return name, nil
}

// NormalizeTagNames normalizes every name, dropping duplicates and sorting the result.
func NormalizeTagNames(names []string) ([]string, error) {
	out := make([]string, 0, len(names))
	for _, name := range names {
		normalized, err := NormalizeTagName(name)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(out, normalized) {
			out = append(out, normalized)
		}
	}
	slices.Sort(out)
	return out, nil
}

// IsTagColor reports whether color is one of the palette names.
func IsTagColor(color string) bool {
	return slices.Contains(TagColors, color)
}
