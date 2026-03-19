package dialogs

import (
	"crona/tui/internal/api"

	"github.com/charmbracelet/bubbles/textinput"
)

func DefaultStreamOptionsForTesting(inputs []textinput.Model, repoIndex int, repos []api.Repo, allIssues []api.IssueWithMeta, streams []api.Stream, context *api.ActiveContext) []SelectorOption {
	return DefaultStreamOptions(inputs, repoIndex, repos, allIssues, streams, context)
}

func MatchStreamSelectionForTesting(raw string, repoID int64, repoName string, streamIndex int, repos []api.Repo, allIssues []api.IssueWithMeta, streams []api.Stream, context *api.ActiveContext) (int64, string) {
	return matchStreamSelection(raw, repoID, repoName, streamIndex, repos, allIssues, streams, context)
}
