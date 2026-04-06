package contextmeta

import (
	"strings"

	"crona/tui/internal/api"
)

func DefaultScopeLabel(ctx *api.ActiveContext) string {
	if ctx == nil {
		return "Scope: All"
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
		return "Scope: " + repoName + " > " + streamName
	case repoName != "":
		return "Scope: " + repoName
	case streamName != "":
		return "Scope: " + streamName
	}
	return "Scope: All"
}
