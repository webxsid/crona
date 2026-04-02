package support

import (
	"crona/tui/internal/api"
	"crona/tui/internal/tui/app"
	"crona/tui/internal/tui/dialogs"
	"crona/tui/internal/tui/views"

	"github.com/charmbracelet/bubbles/textinput"
)

func Theme() views.Theme { return views.TestTheme() }

func RenderDaily(state views.ContentState) string     { return views.RenderDaily(Theme(), state) }
func RenderDefault(state views.ContentState) string   { return views.RenderDefault(Theme(), state) }
func RenderRollup(state views.ContentState) string    { return views.RenderRollup(Theme(), state) }
func RenderWellbeing(state views.ContentState) string { return views.RenderWellbeing(Theme(), state) }
func RenderReports(state views.ContentState) string   { return views.RenderReports(Theme(), state) }
func RenderSettings(state views.ContentState) string  { return views.RenderSettings(Theme(), state) }
func RenderConfig(state views.ContentState) string    { return views.RenderConfig(Theme(), state) }
func RenderSupport(state views.ContentState) string   { return views.RenderSupport(Theme(), state) }
func RenderUpdates(state views.ContentState) string   { return views.RenderContent(Theme(), state) }
func RenderPaneBox(theme views.Theme, active bool, width, height int, content string) string {
	return views.RenderPaneBox(theme, active, width, height, content)
}

func NewDailyModel(width, height int) app.Model { return app.NewDailyRenderModel(width, height) }
func NewDailyHabitDeleteModel(habits []api.HabitDailyItem) app.Model {
	return app.NewDailyHabitDeleteModel(habits)
}
func MinimumSize() (int, int) { return app.MinimumSize() }
func OpenSelectedDeleteDialog(m app.Model) (app.Model, bool) {
	return app.OpenSelectedDeleteDialog(m)
}

func DefaultStreamOptions(inputs []textinput.Model, repoIndex int, repos []api.Repo, allIssues []api.IssueWithMeta, streams []api.Stream, context *api.ActiveContext) []dialogs.SelectorOption {
	return dialogs.DefaultStreamOptions(inputs, repoIndex, repos, allIssues, streams, context)
}

func MatchStreamSelection(raw string, repoID int64, repoName string, streamIndex int, repos []api.Repo, allIssues []api.IssueWithMeta, streams []api.Stream, context *api.ActiveContext) (int64, string) {
	return dialogs.MatchStreamSelection(raw, repoID, repoName, streamIndex, repos, allIssues, streams, context)
}
