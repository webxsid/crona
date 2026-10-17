package state

import (
	"slices"
	"testing"
)

func TestEveryViewBelongsToExactlyOneTab(t *testing.T) {
	all := []View{
		ViewAway, ViewSummary, ViewDefault, ViewDaily, ViewRollup, ViewMomentum, ViewMeta,
		ViewSessionHistory, ViewHabitHistory, ViewSessionActive, ViewOps, ViewWellbeing,
		ViewReports, ViewConfig, ViewSettings, ViewAlerts, ViewUpdates, ViewSupport,
	}
	seen := map[View]Tab{}
	for _, tab := range TabOrder() {
		for _, view := range TabViews(tab) {
			if prev, dup := seen[view]; dup {
				t.Fatalf("view %q is in both %q and %q", view, prev, tab)
			}
			seen[view] = tab
		}
	}
	for _, view := range all {
		if seen[view] == "" {
			t.Fatalf("view %q has no tab", view)
		}
		if got := TabForView(view); got != seen[view] {
			t.Fatalf("TabForView(%q) = %q, want %q", view, got, seen[view])
		}
		if ViewLabel(view) == "" {
			t.Fatalf("view %q has no label", view)
		}
	}
}

func TestTabOrderAndInsightsViews(t *testing.T) {
	wantOrder := []Tab{TabToday, TabIssues, TabInsights, TabLog, TabSystem}
	if got := TabOrder(); !slices.Equal(got, wantOrder) {
		t.Fatalf("TabOrder() = %v, want %v", got, wantOrder)
	}
	wantInsights := []View{ViewRollup, ViewMomentum, ViewWellbeing}
	if got := TabViews(TabInsights); !slices.Equal(got, wantInsights) {
		t.Fatalf("TabViews(insights) = %v, want %v", got, wantInsights)
	}
	if got := TabLabel(TabLog); got != "Log" {
		t.Fatalf("TabLabel(log) = %q, want Log", got)
	}
	if got := ViewLabel(ViewSessionHistory); got != "Sessions" {
		t.Fatalf("ViewLabel(session_history) = %q, want Sessions", got)
	}
}
