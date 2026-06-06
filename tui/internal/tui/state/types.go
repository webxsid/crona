package state

import navigationutil "crona/tui/internal/tui/navigationutil"

type View string

const (
	ViewAway           View = "away"
	ViewDefault        View = "default"
	ViewDaily          View = "daily"
	ViewRollup         View = "rollup"
	ViewMomentum       View = "momentum"
	ViewMeta           View = "meta"
	ViewSessionHistory View = "session_history"
	ViewHabitHistory   View = "habit_history"
	ViewSessionActive  View = "session_active"
	ViewOps            View = "ops"
	ViewWellbeing      View = "wellbeing"
	ViewReports        View = "reports"
	ViewConfig         View = "config"
	ViewSettings       View = "settings"
	ViewAlerts         View = "alerts"
	ViewUpdates        View = "updates"
	ViewSupport        View = "support"
)

type Pane string

const (
	PaneRepos            Pane = "repos"
	PaneStreams          Pane = "streams"
	PaneIssues           Pane = "issues"
	PaneHabits           Pane = "habits"
	PaneRollupDays       Pane = "rollup_days"
	PaneSessions         Pane = "sessions"
	PaneHabitHistory     Pane = "habit_history"
	PaneMomentumCards    Pane = "momentum_cards"
	PaneOps              Pane = "ops"
	PaneExportReports    Pane = "export_reports"
	PaneConfig           Pane = "config"
	PaneSettings         Pane = "settings"
	PaneAlerts           Pane = "alerts"
	PaneWellbeingSummary Pane = "wellbeing_summary"
	PaneWellbeingTrends  Pane = "wellbeing_trends"
)

type DefaultIssueSection string

const (
	DefaultIssueSectionOpen      DefaultIssueSection = "open"
	DefaultIssueSectionCompleted DefaultIssueSection = "completed"
)

type DailyTaskSection string

const (
	DailyTaskSectionPlanned DailyTaskSection = "planned"
	DailyTaskSectionPinned  DailyTaskSection = "pinned"
	DailyTaskSectionOverdue DailyTaskSection = "overdue"
)

var viewOrder = []View{
	ViewSessionHistory,
	ViewHabitHistory,
	ViewDaily,
	ViewRollup,
	ViewMomentum,
	ViewWellbeing,
	ViewReports,
	ViewConfig,
	ViewDefault,
	ViewMeta,
	ViewOps,
	ViewSettings,
	ViewAlerts,
	ViewUpdates,
	ViewSupport,
}

var viewPanes = map[View][]Pane{
	ViewAway:           {},
	ViewDefault:        {PaneIssues},
	ViewDaily:          {PaneIssues, PaneHabits},
	ViewRollup:         {PaneRollupDays},
	ViewMomentum:       {PaneMomentumCards},
	ViewMeta:           {PaneRepos, PaneStreams, PaneIssues, PaneHabits},
	ViewSessionHistory: {PaneSessions},
	ViewHabitHistory:   {PaneHabitHistory},
	ViewSessionActive:  {},
	ViewOps:            {PaneOps},
	ViewWellbeing:      {PaneWellbeingSummary, PaneWellbeingTrends},
	ViewReports:        {PaneExportReports},
	ViewConfig:         {PaneConfig},
	ViewSettings:       {PaneSettings},
	ViewAlerts:         {PaneAlerts},
	ViewUpdates:        {},
	ViewSupport:        {},
}

var viewDefaultPane = map[View]Pane{
	ViewAway:           PaneIssues,
	ViewDefault:        PaneIssues,
	ViewDaily:          PaneIssues,
	ViewRollup:         PaneRollupDays,
	ViewMomentum:       PaneMomentumCards,
	ViewMeta:           PaneRepos,
	ViewSessionHistory: PaneSessions,
	ViewHabitHistory:   PaneHabitHistory,
	ViewSessionActive:  PaneIssues,
	ViewOps:            PaneOps,
	ViewWellbeing:      PaneWellbeingSummary,
	ViewReports:        PaneExportReports,
	ViewConfig:         PaneConfig,
	ViewSettings:       PaneSettings,
	ViewAlerts:         PaneAlerts,
	ViewUpdates:        PaneIssues,
	ViewSupport:        PaneIssues,
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
