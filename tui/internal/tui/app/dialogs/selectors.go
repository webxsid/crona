package dialogs

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
	query := strings.TrimSpace(strings.ToLower(inputs[0].Value()))
	options := make([]SelectorOption, 0, len(repos)+1)
	for _, repo := range repos {
		if query != "" && !strings.Contains(strings.ToLower(repo.Name), query) {
			continue
		}
		options = append(options, SelectorOption{ID: strconv.FormatInt(repo.ID, 10), Label: repo.Name})
	}
	if raw := strings.TrimSpace(inputs[0].Value()); raw != "" {
		options = append(options, SelectorOption{ID: "__new__", Label: "Create New Repo: " + raw})
	}
	return options
}

func CheckoutRepoOptions(inputs []textinput.Model, repos []api.Repo) []SelectorOption {
	return DefaultRepoOptions(inputs, repos)
}

func DefaultStreamOptions(inputs []textinput.Model, repoIndex int, repos []api.Repo, allIssues []api.IssueWithMeta, streams []api.Stream, context *api.ActiveContext) []SelectorOption {
	query := strings.TrimSpace(strings.ToLower(inputs[1].Value()))
	repoOptions := DefaultRepoOptions(inputs, repos)
	if len(repoOptions) == 0 {
		return optionsForNewStream(inputs[1].Value())
	}
	repoOpt := repoOptions[minInt(repoIndex, len(repoOptions)-1)]
	if repoOpt.ID == "__new__" {
		return optionsForNewStream(inputs[1].Value())
	}

	seen := map[string]bool{}
	options := []SelectorOption{}
	for _, issue := range allIssues {
		if strconv.FormatInt(issue.RepoID, 10) != repoOpt.ID || seen[strconv.FormatInt(issue.StreamID, 10)] {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(issue.StreamName), query) {
			continue
		}
		seen[strconv.FormatInt(issue.StreamID, 10)] = true
		options = append(options, SelectorOption{ID: strconv.FormatInt(issue.StreamID, 10), Label: issue.StreamName})
	}
	if context != nil && context.RepoID != nil && strconv.FormatInt(*context.RepoID, 10) == repoOpt.ID {
		for _, stream := range streams {
			streamKey := strconv.FormatInt(stream.ID, 10)
			if seen[streamKey] {
				continue
			}
			if query != "" && !strings.Contains(strings.ToLower(stream.Name), query) {
				continue
			}
			seen[streamKey] = true
			options = append(options, SelectorOption{ID: streamKey, Label: stream.Name})
		}
	}
	if raw := strings.TrimSpace(inputs[1].Value()); raw != "" {
		options = append(options, SelectorOption{ID: "__new__", Label: "Create New Stream: " + raw})
	}
	return options
}

func CheckoutStreamOptions(inputs []textinput.Model, repoIndex int, repos []api.Repo, allIssues []api.IssueWithMeta, streams []api.Stream, context *api.ActiveContext) []SelectorOption {
	return DefaultStreamOptions(inputs, repoIndex, repos, allIssues, streams, context)
}

func CheckoutDialogLabels(inputs []textinput.Model, repoIndex, streamIndex int, repos []api.Repo, allIssues []api.IssueWithMeta, streams []api.Stream, context *api.ActiveContext) (string, string) {
	repoOptions := CheckoutRepoOptions(inputs, repos)
	streamOptions := CheckoutStreamOptions(inputs, repoIndex, repos, allIssues, streams, context)
	if len(repoOptions) == 0 {
		return "Type to search", "Select a repo first"
	}
	if len(streamOptions) == 0 {
		return repoOptions[minInt(repoIndex, len(repoOptions)-1)].Label, "Type to search or create"
	}
	return repoOptions[minInt(repoIndex, len(repoOptions)-1)].Label, streamOptions[minInt(streamIndex, len(streamOptions)-1)].Label
}

func CheckoutDialogSelection(inputs []textinput.Model, repoIndex, streamIndex int, repos []api.Repo, allIssues []api.IssueWithMeta, streams []api.Stream, context *api.ActiveContext) (int64, string, *int64, string) {
	repoRaw := strings.TrimSpace(inputs[0].Value())
	streamRaw := strings.TrimSpace(inputs[1].Value())
	if repoRaw == "" && streamRaw == "" {
		return 0, "", nil, ""
	}

	repoID, repoName := matchRepoSelection(repoRaw, repoIndex, repos)
	if repoName == "" {
		return 0, "", nil, ""
	}
	if streamRaw == "" {
		return repoID, repoName, nil, ""
	}

	streamID, streamName := matchStreamSelection(streamRaw, repoID, repoName, streamIndex, repos, allIssues, streams, context)
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
		if strings.EqualFold(strings.TrimSpace(repo.Name), raw) {
			return repo.ID, repo.Name
		}
	}
	options := DefaultRepoOptions([]textinput.Model{valueInput(raw)}, repos)
	if len(options) == 0 {
		return 0, raw
	}
	selected := options[minInt(repoIndex, len(options)-1)]
	if selected.ID == "__new__" {
		return 0, raw
	}
	id, err := strconv.ParseInt(selected.ID, 10, 64)
	if err != nil {
		return 0, raw
	}
	return id, selected.Label
}

func matchStreamSelection(raw string, repoID int64, repoName string, streamIndex int, repos []api.Repo, allIssues []api.IssueWithMeta, streams []api.Stream, context *api.ActiveContext) (int64, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, ""
	}
	for _, stream := range streams {
		if stream.RepoID == repoID && strings.EqualFold(strings.TrimSpace(stream.Name), raw) {
			return stream.ID, stream.Name
		}
	}
	inputs := []textinput.Model{valueInput(repoName), valueInput(raw)}
	options := DefaultStreamOptions(inputs, 0, repos, allIssues, streams, context)
	if len(options) == 0 {
		return 0, raw
	}
	selected := options[minInt(streamIndex, len(options)-1)]
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

func DefaultIssueDialogLabels(inputs []textinput.Model, repoIndex, streamIndex int, repos []api.Repo, allIssues []api.IssueWithMeta, streams []api.Stream, context *api.ActiveContext) (string, string) {
	repoOptions := DefaultRepoOptions(inputs, repos)
	streamOptions := DefaultStreamOptions(inputs, repoIndex, repos, allIssues, streams, context)
	if len(repoOptions) == 0 {
		return "Type to search or create", "Select a repo first"
	}
	if len(streamOptions) == 0 {
		return repoOptions[minInt(repoIndex, len(repoOptions)-1)].Label, "Type to search or create"
	}
	return repoOptions[minInt(repoIndex, len(repoOptions)-1)].Label, streamOptions[minInt(streamIndex, len(streamOptions)-1)].Label
}

func DefaultIssueDialogNames(inputs []textinput.Model, repoIndex, streamIndex int, repos []api.Repo, allIssues []api.IssueWithMeta, streams []api.Stream, context *api.ActiveContext) (string, string) {
	repoOptions := DefaultRepoOptions(inputs, repos)
	streamOptions := DefaultStreamOptions(inputs, repoIndex, repos, allIssues, streams, context)
	if len(repoOptions) == 0 || len(streamOptions) == 0 {
		return "", ""
	}
	repo := repoOptions[minInt(repoIndex, len(repoOptions)-1)]
	stream := streamOptions[minInt(streamIndex, len(streamOptions)-1)]
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
	return (current + dir + total) % total
}

func optionsForNewStream(raw string) []SelectorOption {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []SelectorOption{}
	}
	return []SelectorOption{{ID: "__new__", Label: "Create New Stream: " + raw}}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
