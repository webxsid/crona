package state

type Tab string

const (
	TabToday    Tab = "today"
	TabIssues   Tab = "issues"
	TabInsights Tab = "insights"
	TabLog      Tab = "log"
	TabSystem   Tab = "system"
)

var tabOrder = []Tab{TabToday, TabIssues, TabInsights, TabLog, TabSystem}

var tabViews = map[Tab][]View{
	TabToday:    {ViewDaily, ViewSessionActive, ViewAway},
	TabIssues:   {ViewDefault, ViewHabitList},
	TabInsights: {ViewRollup, ViewMomentum, ViewWellbeing},
	TabLog:      {ViewSessionHistory, ViewHabitHistory, ViewOps},
	TabSystem:   {ViewSettings, ViewAlerts, ViewReports, ViewConfig, ViewUpdates, ViewSupport},
}

var tabLabels = map[Tab]string{
	TabToday:    "Today",
	TabIssues:   "Issues",
	TabInsights: "Insights",
	TabLog:      "Log",
	TabSystem:   "System",
}

var viewLabels = map[View]string{
	ViewDaily:          "Daily",
	ViewSessionActive:  "Session",
	ViewAway:           "Away",
	ViewDefault:        "Issues",
	ViewHabitList:      "Habits",
	ViewRollup:         "Rollup",
	ViewMomentum:       "Momentum",
	ViewWellbeing:      "Wellbeing",
	ViewSessionHistory: "Sessions",
	ViewHabitHistory:   "Habits",
	ViewOps:            "Ops",
	ViewSettings:       "Settings",
	ViewAlerts:         "Alerts",
	ViewReports:        "Reports",
	ViewConfig:         "Config",
	ViewUpdates:        "Updates",
	ViewSupport:        "Support",
}

func TabOrder() []Tab {
	out := make([]Tab, len(tabOrder))
	copy(out, tabOrder)
	return out
}

func TabViews(tab Tab) []View {
	views := tabViews[tab]
	out := make([]View, len(views))
	copy(out, views)
	return out
}

func TabForView(view View) Tab {
	for _, tab := range tabOrder {
		for _, candidate := range tabViews[tab] {
			if candidate == view {
				return tab
			}
		}
	}
	return TabToday
}

func TabLabel(tab Tab) string {
	return tabLabels[tab]
}

func ViewLabel(view View) string {
	return viewLabels[view]
}
