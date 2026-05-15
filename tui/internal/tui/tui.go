package tui

import (
	"crona/tui/internal/api"
	modelpkg "crona/tui/internal/tui/model"
	sharedposthog "crona/shared/posthog"
)

type Model = modelpkg.Model
type View = modelpkg.View
type Pane = modelpkg.Pane

const (
	ViewDefault        = modelpkg.ViewDefault
	ViewDaily          = modelpkg.ViewDaily
	ViewMeta           = modelpkg.ViewMeta
	ViewSessionHistory = modelpkg.ViewSessionHistory
	ViewSessionActive  = modelpkg.ViewSessionActive
	ViewScratch        = modelpkg.ViewScratch
	ViewOps            = modelpkg.ViewOps
	ViewSettings       = modelpkg.ViewSettings
	ViewAlerts         = modelpkg.ViewAlerts
	ViewSupport        = modelpkg.ViewSupport
)

const (
	PaneRepos       = modelpkg.PaneRepos
	PaneStreams     = modelpkg.PaneStreams
	PaneIssues      = modelpkg.PaneIssues
	PaneSessions    = modelpkg.PaneSessions
	PaneScratchpads = modelpkg.PaneScratchpads
	PaneOps         = modelpkg.PaneOps
	PaneSettings    = modelpkg.PaneSettings
	PaneAlerts      = modelpkg.PaneAlerts
)

func SetEventChannel(ch <-chan api.KernelEvent) {
	modelpkg.SetEventChannel(ch)
}

func New(transport, endpoint, scratchDir, env, executablePath string, done chan struct{}, telemetry sharedposthog.Client) Model {
	return modelpkg.New(transport, endpoint, scratchDir, env, executablePath, done, telemetry)
}
