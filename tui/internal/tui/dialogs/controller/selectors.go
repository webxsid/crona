package controller

import (
	"strconv"
	"strings"

	"crona/tui/internal/api"

	"github.com/charmbracelet/bubbles/textinput"
)

type SelectorOption struct {
	ID    string
	Label string
}

func DefaultRepoOptions(inputs []textinput.Model, repos []api.Repo) []SelectorOption {
	query := normalizeSelectorName(inputs[0].Value())
	options := make([]SelectorOption, 0, len(repos)+1)
	for _, repo := range repos {
		if query != "" && !strings.Contains(normalizeSelectorName(repo.Name), query) {
			continue
		}
		options = append(
			options,
			SelectorOption{ID: strconv.FormatInt(repo.ID, 10), Label: repo.Name},
		)
	}
	if raw := strings.TrimSpace(inputs[0].Value()); raw != "" {
		options = append(options, SelectorOption{ID: "__new__", Label: "Create New Repo: " + raw})
	}
	return options
}

func CheckoutRepoOptions(inputs []textinput.Model, repos []api.Repo) []SelectorOption {
	return DefaultRepoOptions(inputs, repos)
}

func DefaultStreamOptions(
	inputs []textinput.Model,
	repoIndex int,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
	context *api.ActiveContext,
) []SelectorOption {
	_ = context
	query := normalizeSelectorName(inputs[1].Value())
	repoOptions := DefaultRepoOptions(inputs, repos)
	repoOpt, ok := selectedOption(repoOptions, repoIndex)
	if !ok {
		return []SelectorOption{}
	}
	if repoOpt.ID == "__new__" {
		return optionsForNewStream(inputs[1].Value())
	}

	seen := map[string]bool{}
	options := []SelectorOption{}
	for _, issue := range allIssues {
		if strconv.FormatInt(issue.RepoID, 10) != repoOpt.ID ||
			seen[strconv.FormatInt(issue.StreamID, 10)] {
			continue
		}
		if query != "" && !strings.Contains(normalizeSelectorName(issue.StreamName), query) {
			continue
		}
		seen[strconv.FormatInt(issue.StreamID, 10)] = true
		options = append(
			options,
			SelectorOption{ID: strconv.FormatInt(issue.StreamID, 10), Label: issue.StreamName},
		)
	}
	for _, stream := range streams {
		if strconv.FormatInt(stream.RepoID, 10) != repoOpt.ID {
			continue
		}
		streamKey := strconv.FormatInt(stream.ID, 10)
		if seen[streamKey] {
			continue
		}
		if query != "" && !strings.Contains(normalizeSelectorName(stream.Name), query) {
			continue
		}
		seen[streamKey] = true
		options = append(options, SelectorOption{ID: streamKey, Label: stream.Name})
	}
	if raw := strings.TrimSpace(inputs[1].Value()); raw != "" {
		options = append(options, SelectorOption{ID: "__new__", Label: "Create New Stream: " + raw})
	}
	return options
}

func CheckoutStreamOptions(
	inputs []textinput.Model,
	repoIndex int,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
	context *api.ActiveContext,
) []SelectorOption {
	return DefaultStreamOptions(inputs, repoIndex, repos, allIssues, streams, context)
}

func CheckoutDialogLabels(
	inputs []textinput.Model,
	repoIndex, streamIndex int,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
	context *api.ActiveContext,
) (string, string) {
	repoRaw := strings.TrimSpace(inputs[0].Value())
	streamRaw := strings.TrimSpace(inputs[1].Value())
	repoOptions := CheckoutRepoOptions(inputs, repos)
	streamOptions := CheckoutStreamOptions(inputs, repoIndex, repos, allIssues, streams, context)
	return labelOrPlaceholder(repoRaw, repoOptions, repoIndex, "Select a repo"),
		labelOrPlaceholder(streamRaw, streamOptions, streamIndex, "Select a stream")
}

func CheckoutDialogSelection(
	inputs []textinput.Model,
	repoIndex, streamIndex int,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
	context *api.ActiveContext,
) (int64, string, *int64, string) {
	repoRaw := strings.TrimSpace(inputs[0].Value())
	streamRaw := strings.TrimSpace(inputs[1].Value())
	repoOptions := CheckoutRepoOptions(inputs, repos)
	var repoID int64
	var repoName string
	if repoRaw == "" {
		option, ok := selectedOption(repoOptions, repoIndex)
		if !ok {
			return 0, "", nil, ""
		}
		if option.ID == "__new__" {
			return 0, "", nil, ""
		}
		id, err := strconv.ParseInt(option.ID, 10, 64)
		if err != nil {
			return 0, "", nil, ""
		}
		repoID = id
		repoName = option.Label
	} else {
		repoID, repoName = matchRepoSelection(repoRaw, repoIndex, repos)
		if repoName == "" {
			return 0, "", nil, ""
		}
	}
	if streamRaw == "" {
		streamOptions := CheckoutStreamOptions(inputs, repoIndex, repos, allIssues, streams, context)
		option, ok := selectedOption(streamOptions, streamIndex)
		if !ok {
			return repoID, repoName, nil, ""
		}
		if option.ID == "__new__" {
			return repoID, repoName, nil, option.Label
		}
		id, err := strconv.ParseInt(option.ID, 10, 64)
		if err != nil {
			return repoID, repoName, nil, option.Label
		}
		return repoID, repoName, &id, option.Label
	}

	streamID, streamName := MatchStreamSelection(
		streamRaw,
		repoID,
		repoName,
		streamIndex,
		repos,
		allIssues,
		streams,
		context,
	)
	if streamName == "" {
		return repoID, repoName, nil, ""
	}
	if streamID == 0 {
		return repoID, repoName, nil, streamName
	}
	return repoID, repoName, &streamID, streamName
}

func matchRepoSelection(raw string, repoIndex int, repos []api.Repo) (int64, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, ""
	}
	for _, repo := range repos {
		if normalizeSelectorName(repo.Name) == normalizeSelectorName(raw) {
			return repo.ID, repo.Name
		}
	}
	options := DefaultRepoOptions([]textinput.Model{valueInput(raw)}, repos)
	if len(options) == 0 {
		return 0, raw
	}
	selected := options[min(repoIndex, len(options)-1)]
	if selected.ID == "__new__" {
		return 0, raw
	}
	id, err := strconv.ParseInt(selected.ID, 10, 64)
	if err != nil {
		return 0, raw
	}
	return id, selected.Label
}

func MatchStreamSelection(
	raw string,
	repoID int64,
	repoName string,
	streamIndex int,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
	context *api.ActiveContext,
) (int64, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, ""
	}
	for _, stream := range streams {
		if stream.RepoID == repoID &&
			normalizeSelectorName(stream.Name) == normalizeSelectorName(raw) {
			return stream.ID, stream.Name
		}
	}
	inputs := []textinput.Model{valueInput(repoName), valueInput(raw)}
	options := DefaultStreamOptions(inputs, 0, repos, allIssues, streams, context)
	if len(options) == 0 {
		return 0, raw
	}
	selected := options[min(streamIndex, len(options)-1)]
	if selected.ID == "__new__" {
		return 0, raw
	}
	id, err := strconv.ParseInt(selected.ID, 10, 64)
	if err != nil {
		return 0, raw
	}
	return id, selected.Label
}

func valueInput(value string) textinput.Model {
	input := textinput.New()
	input.SetValue(value)
	return input
}

func SyncFocus(inputs []textinput.Model, focusIdx int) []textinput.Model {
	for i := range inputs {
		inputs[i].Blur()
	}
	if focusIdx >= 0 && focusIdx < len(inputs) {
		inputs[focusIdx].Focus()
	}
	return inputs
}

func DefaultIssueDialogLabels(
	inputs []textinput.Model,
	repoIndex, streamIndex int,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
	context *api.ActiveContext,
) (string, string) {
	repoOptions := DefaultRepoOptions(inputs, repos)
	streamOptions := DefaultStreamOptions(inputs, repoIndex, repos, allIssues, streams, context)
	if len(repoOptions) == 0 {
		return "Type to search or create", "Select a repo first"
	}
	if len(streamOptions) == 0 {
		return repoOptions[min(repoIndex, len(repoOptions)-1)].Label, "Type to search or create"
	}
	return repoOptions[min(repoIndex, len(repoOptions)-1)].Label, streamOptions[min(streamIndex, len(streamOptions)-1)].Label
}

func MomentumDialogLabels(
	inputs []textinput.Model,
	repoIndex, streamIndex int,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
	context *api.ActiveContext,
) (string, string) {
	repoRaw := ""
	if len(inputs) > 0 {
		repoRaw = strings.TrimSpace(inputs[0].Value())
	}
	streamRaw := ""
	if len(inputs) > 1 {
		streamRaw = strings.TrimSpace(inputs[1].Value())
	}
	repoOptions := stripCreateOptions(DefaultRepoOptions(inputs, repos))
	streamOptions := momentumStreamOptionsForLabels(inputs, repoIndex, repos, allIssues, streams, context)
	return labelOrPlaceholder(repoRaw, repoOptions, repoIndex, "Type to search"),
		momentumStreamLabel(streamRaw, streamOptions, streamIndex)
}

func momentumStreamOptionsForLabels(
	inputs []textinput.Model,
	repoIndex int,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
	context *api.ActiveContext,
) []SelectorOption {
	options := []SelectorOption{{ID: "__any__", Label: "Any Stream"}}
	options = append(options, DefaultStreamOptions(inputs, repoIndex, repos, allIssues, streams, context)...)
	return stripCreateOptions(options)
}

func DefaultIssueDialogNames(
	inputs []textinput.Model,
	repoIndex, streamIndex int,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
	context *api.ActiveContext,
) (string, string) {
	repoOptions := DefaultRepoOptions(inputs, repos)
	streamOptions := DefaultStreamOptions(inputs, repoIndex, repos, allIssues, streams, context)
	if len(repoOptions) == 0 || len(streamOptions) == 0 {
		return "", ""
	}
	repo := repoOptions[min(repoIndex, len(repoOptions)-1)]
	stream := streamOptions[min(streamIndex, len(streamOptions)-1)]
	repoName := repo.Label
	if repo.ID == "__new__" {
		repoName = strings.TrimSpace(inputs[0].Value())
	}
	streamName := stream.Label
	if stream.ID == "__new__" {
		streamName = strings.TrimSpace(inputs[1].Value())
	}
	return repoName, streamName
}

func ShiftSelection(current, total, dir int) int {
	if total == 0 {
		return current
	}
	if current < 0 {
		if dir < 0 {
			return total - 1
		}
		return 0
	}
	return (current + dir + total) % total
}

func selectedOption(options []SelectorOption, index int) (SelectorOption, bool) {
	if len(options) == 0 || index < 0 {
		return SelectorOption{}, false
	}
	if index >= len(options) {
		index = len(options) - 1
	}
	return options[index], true
}

func labelOrPlaceholder(raw string, options []SelectorOption, index int, placeholder string) string {
	if strings.TrimSpace(raw) == "" && index < 0 {
		return placeholder
	}
	option, ok := selectedOption(options, index)
	if !ok {
		return placeholder
	}
	return option.Label
}

func momentumStreamLabel(raw string, options []SelectorOption, index int) string {
	if strings.TrimSpace(raw) == "" && index < 0 {
		if len(options) > 0 {
			return options[0].Label
		}
		return "Any Stream"
	}
	if index < 0 {
		if len(options) > 0 {
			return options[0].Label
		}
		return "Any Stream"
	}
	if index+1 < len(options) {
		return options[index+1].Label
	}
	if len(options) > 0 {
		return options[len(options)-1].Label
	}
	return "Any Stream"
}

func optionsForNewStream(raw string) []SelectorOption {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []SelectorOption{}
	}
	return []SelectorOption{{ID: "__new__", Label: "Create New Stream: " + raw}}
}

func normalizeSelectorName(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}
