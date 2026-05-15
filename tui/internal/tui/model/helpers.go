package model

import (
	"strings"

	"crona/tui/internal/api"
	commands "crona/tui/internal/tui/commands"
	helperpkg "crona/tui/internal/tui/helpers"

	tea "github.com/charmbracelet/bubbletea"
)

func habitHistoryScopeLabel(ctx *api.ActiveContext) string {
	if ctx == nil {
		return "Recent habit activity across the workspace"
	}
	repoName := ""
	if ctx.RepoName != nil {
		repoName = strings.TrimSpace(*ctx.RepoName)
	}
	streamName := ""
	if ctx.StreamName != nil {
		streamName = strings.TrimSpace(*ctx.StreamName)
	}
	switch {
	case repoName != "" && streamName != "":
		return "Recent habit activity in " + repoName + " > " + streamName
	case repoName != "":
		return "Recent habit activity in " + repoName
	case streamName != "":
		return "Recent habit activity in " + streamName
	default:
		return "Recent habit activity across the workspace"
	}
}

func loadSessionHistoryForModel(m Model, limit int) tea.Cmd {
	return commands.LoadSessionHistory(m.client, helperpkg.SessionHistoryScopeIssueID(m.timer), limit)
}
