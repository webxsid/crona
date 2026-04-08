package support

import (
	"crona/tui/internal/api"
	"crona/tui/internal/tui/app"
	"crona/tui/internal/tui/dialogs"
	layoutpkg "crona/tui/internal/tui/layout"
	alertsview "crona/tui/internal/tui/views/alerts"
	awayview "crona/tui/internal/tui/views/away"
	viewchrome "crona/tui/internal/tui/views/chrome"
	configview "crona/tui/internal/tui/views/config"
	dailyview "crona/tui/internal/tui/views/daily"
	issuesview "crona/tui/internal/tui/views/issues"
	reportsview "crona/tui/internal/tui/views/reports"
	rollupview "crona/tui/internal/tui/views/rollup"
	scratchpadsview "crona/tui/internal/tui/views/scratchpads"
	settingsview "crona/tui/internal/tui/views/settings"
	supportview "crona/tui/internal/tui/views/support"
	viewtypes "crona/tui/internal/tui/views/types"
	updatesview "crona/tui/internal/tui/views/updates"
	wellbeingview "crona/tui/internal/tui/views/wellbeing"

	"github.com/charmbracelet/bubbles/textinput"
)

func Theme() viewtypes.Theme { return layoutpkg.ViewTheme() }

func RenderDaily(state viewtypes.ContentState) string   { return dailyview.Render(Theme(), state) }
func RenderDefault(state viewtypes.ContentState) string { return issuesview.Render(Theme(), state) }
func RenderRollup(state viewtypes.ContentState) string  { return rollupview.Render(Theme(), state) }
func RenderWellbeing(state viewtypes.ContentState) string {
	return wellbeingview.Render(Theme(), state)
}
func RenderReports(state viewtypes.ContentState) string  { return reportsview.Render(Theme(), state) }
func RenderSettings(state viewtypes.ContentState) string { return settingsview.Render(Theme(), state) }
func RenderAlerts(state viewtypes.ContentState) string   { return alertsview.Render(Theme(), state) }
func RenderConfig(state viewtypes.ContentState) string   { return configview.Render(Theme(), state) }
func RenderSupport(state viewtypes.ContentState) string  { return supportview.Render(Theme(), state) }
func RenderPaneBox(theme viewtypes.Theme, active bool, width, height int, content string) string {
	return viewchrome.RenderPaneBox(theme, active, width, height, content)
}
func RenderUpdates(state viewtypes.ContentState) string {
	return updatesview.Render(Theme(), state)
}
func RenderAway(state viewtypes.ContentState) string { return awayview.Render(Theme(), state) }
func RenderScratchpads(state viewtypes.ContentState) string {
	return scratchpadsview.Render(Theme(), state)
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
