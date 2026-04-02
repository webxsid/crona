package tui

import (
	"crona/tui/internal/api"
	"crona/tui/internal/tui/app"
)

type Model = app.Model
type View = app.View
type Pane = app.Pane

const (
	ViewDefault        = app.ViewDefault
	ViewDaily          = app.ViewDaily
	ViewMeta           = app.ViewMeta
	ViewSessionHistory = app.ViewSessionHistory
	ViewSessionActive  = app.ViewSessionActive
	ViewScratch        = app.ViewScratch
	ViewOps            = app.ViewOps
	ViewSettings       = app.ViewSettings
	ViewSupport        = app.ViewSupport
)

const (
	PaneRepos       = app.PaneRepos
	PaneStreams     = app.PaneStreams
	PaneIssues      = app.PaneIssues
	PaneSessions    = app.PaneSessions
	PaneScratchpads = app.PaneScratchpads
	PaneOps         = app.PaneOps
	PaneSettings    = app.PaneSettings
)

func SetEventChannel(ch <-chan api.KernelEvent) {
	app.SetEventChannel(ch)
}

func New(transport, endpoint, scratchDir, env, executablePath string, done chan struct{}) Model {
	return app.New(transport, endpoint, scratchDir, env, executablePath, done)
}
