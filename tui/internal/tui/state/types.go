package state

import navigationutil "crona/tui/internal/tui/navigationutil"

type View string

const (
	ViewAway           View = "away"
	ViewDefault        View = "default"
	ViewDaily          View = "daily"
	ViewReview         View = "review"
	ViewMomentum       View = "momentum"
	ViewHabitList      View = "habit_list"
	ViewSessionHistory View = "session_history"
	ViewHabitHistory   View = "habit_history"
	ViewSessionActive  View = "session_active"
	ViewOps            View = "ops"
	ViewReports        View = "reports"
	ViewConfig         View = "config"
	ViewSettings       View = "settings"
	ViewAlerts         View = "alerts"
	ViewUpdates        View = "updates"
	ViewSupport        View = "support"
)

type Pane string

const (
	PaneRepos         Pane = "repos"
	PaneStreams       Pane = "streams"
	PaneScope         Pane = "scope"
	PaneIssues        Pane = "issues"
	PaneHabits        Pane = "habits"
	PaneReviewDays    Pane = "review_days"
	PaneSessions      Pane = "sessions"
	PaneHabitHistory  Pane = "habit_history"
	PaneMomentumCards Pane = "momentum_cards"
	PaneOps           Pane = "ops"
	PaneExportReports Pane = "export_reports"
	PaneConfig        Pane = "config"
	PaneSettings      Pane = "settings"
	PaneAlerts        Pane = "alerts"
)

type DefaultIssueSection string

const (
	DefaultIssueSectionOpen      DefaultIssueSection = "open"
	DefaultIssueSectionCompleted DefaultIssueSection = "completed"
)

type DailyTaskSection string

const (
	DailyTaskSectionPlanned  DailyTaskSection = "planned"
	DailyTaskSectionOverdue  DailyTaskSection = "overdue"
	DailyTaskSectionUpcoming DailyTaskSection = "upcoming"
)

var viewOrder = []View{
	ViewDaily,
	ViewReview,
	ViewMomentum,
	ViewDefault,
	ViewHabitList,
	ViewOps,
	ViewReports,
	ViewConfig,
	ViewSettings,
	ViewAlerts,
	ViewUpdates,
	ViewSupport,
	ViewSessionHistory,
	ViewHabitHistory,
}

var viewPanes = map[View][]Pane{
	ViewAway:           {},
	ViewDefault:        {PaneScope, PaneIssues},
	ViewDaily:          {PaneIssues, PaneHabits},
	ViewReview:         {PaneReviewDays},
	ViewMomentum:       {PaneMomentumCards},
	ViewHabitList:      {PaneScope, PaneHabits},
	ViewSessionHistory: {PaneSessions},
	ViewHabitHistory:   {PaneHabitHistory},
	ViewSessionActive:  {},
	ViewOps:            {PaneOps},
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
	ViewReview:         PaneReviewDays,
	ViewMomentum:       PaneMomentumCards,
	ViewHabitList:      PaneHabits,
	ViewSessionHistory: PaneSessions,
	ViewHabitHistory:   PaneHabitHistory,
	ViewSessionActive:  PaneIssues,
	ViewOps:            PaneOps,
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
