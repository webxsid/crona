package flags

import (
	"flag"
	"strings"

	sharedtypes "crona/shared/types"
)

// TagList collects a repeatable, comma-separated tag flag such as `--tag urgent,quick`.
type TagList []string

func (l *TagList) String() string { return strings.Join(*l, ",") }

func (l *TagList) Set(value string) error {
	for _, part := range strings.Split(value, ",") {
		if part = strings.TrimSpace(part); part != "" {
			*l = append(*l, part)
		}
	}
	return nil
}

// TagFilterFlags registers --tag and --not-tag on fs. Call the returned function after
// parsing to get the filter, or nil when neither flag was given.
func TagFilterFlags(fs *flag.FlagSet) func() (*sharedtypes.TagFilter, error) {
	var anyOf, none TagList
	fs.Var(&anyOf, "tag", "")
	fs.Var(&none, "not-tag", "")
	return func() (*sharedtypes.TagFilter, error) {
		filter := &sharedtypes.TagFilter{}
		var err error
		if filter.Any, err = normalizeFilterNames(anyOf); err != nil {
			return nil, err
		}
		if filter.None, err = normalizeFilterNames(none); err != nil {
			return nil, err
		}
		if filter.Empty() {
			return nil, nil
		}
		return filter, nil
	}
}

func normalizeFilterNames(names []string) ([]string, error) {
	out := []string{}
	for _, name := range names {
		if strings.EqualFold(strings.TrimPrefix(name, "+"), sharedtypes.TagUntagged) {
			out = append(out, sharedtypes.TagUntagged)
			continue
		}
		normalized, err := sharedtypes.NormalizeTagName(name)
		if err != nil {
			return nil, err
		}
		out = append(out, normalized)
	}
	return out, nil
}
