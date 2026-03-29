package state

import navigationutil "crona/tui/internal/tui/navigationutil"

type View string

const (
	ViewDefault        View = "default"
	ViewDaily          View = "daily"
	ViewMeta           View = "meta"
	ViewSessionHistory View = "session_history"
	ViewSessionActive  View = "session_active"
	ViewScratch        View = "scratchpads"
	ViewOps            View = "ops"
	ViewWellbeing      View = "wellbeing"
	ViewReports        View = "reports"
	ViewConfig         View = "config"
	ViewSettings       View = "settings"
	ViewUpdates        View = "updates"
)

type Pane string

const (
	PaneRepos         Pane = "repos"
	PaneStreams       Pane = "streams"
	PaneIssues        Pane = "issues"
	PaneHabits        Pane = "habits"
	PaneSessions      Pane = "sessions"
	PaneScratchpads   Pane = "scratchpads"
	PaneOps           Pane = "ops"
	PaneExportReports Pane = "export_reports"
	PaneConfig        Pane = "config"
	PaneSettings      Pane = "settings"
)

type DefaultIssueSection string

const (
	DefaultIssueSectionOpen      DefaultIssueSection = "open"
	DefaultIssueSectionCompleted DefaultIssueSection = "completed"
)

var viewOrder = []View{ViewSessionHistory, ViewDaily, ViewWellbeing, ViewReports, ViewDefault, ViewMeta, ViewScratch, ViewOps, ViewConfig, ViewSettings, ViewUpdates}

var viewPanes = map[View][]Pane{
	ViewDefault:        {PaneIssues},
	ViewDaily:          {PaneIssues, PaneHabits},
	ViewMeta:           {PaneRepos, PaneStreams, PaneIssues, PaneHabits},
	ViewSessionHistory: {PaneSessions},
	ViewSessionActive:  {},
	ViewScratch:        {PaneScratchpads},
	ViewOps:            {PaneOps},
	ViewWellbeing:      {},
	ViewReports:        {PaneExportReports},
	ViewConfig:         {PaneConfig},
	ViewSettings:       {PaneSettings},
	ViewUpdates:        {},
}

var viewDefaultPane = map[View]Pane{
	ViewDefault:        PaneIssues,
	ViewDaily:          PaneIssues,
	ViewMeta:           PaneRepos,
	ViewSessionHistory: PaneSessions,
	ViewSessionActive:  PaneIssues,
	ViewScratch:        PaneScratchpads,
	ViewOps:            PaneOps,
	ViewWellbeing:      PaneIssues,
	ViewReports:        PaneExportReports,
	ViewConfig:         PaneConfig,
	ViewSettings:       PaneSettings,
	ViewUpdates:        PaneIssues,
}

func ViewOrder() []View {
	out := make([]View, len(viewOrder))
	copy(out, viewOrder)
	return out
}

func ViewPanes(view View) []Pane {
	panes := viewPanes[view]
	out := make([]Pane, len(panes))
	copy(out, panes)
	return out
}

func DefaultPane(view View) Pane {
	return viewDefaultPane[view]
}

func NextPane(view View, current Pane, dir int) Pane {
	return navigationutil.NextPane(viewPanes[view], current, dir)
}
