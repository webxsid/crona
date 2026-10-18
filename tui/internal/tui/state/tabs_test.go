package state

import (
	"slices"
	"testing"
)

func TestEveryViewBelongsToExactlyOneTab(t *testing.T) {
	all := []View{
		ViewAway, ViewDefault, ViewDaily, ViewReview, ViewMomentum, ViewHabitList,
		ViewSessionHistory, ViewHabitHistory, ViewSessionActive, ViewOps,
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
	wantInsights := []View{ViewReview, ViewMomentum}
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

func TestTodayTabHasNoSummary(t *testing.T) {
	got := TabViews(TabToday)
	want := []View{ViewDaily, ViewSessionActive, ViewAway}
	if len(got) != len(want) {
		t.Fatalf("Today views = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Today views = %v, want %v", got, want)
		}
	}
}

func TestIssuesTabSubTabs(t *testing.T) {
	if got := TabViews(TabIssues); !slices.Equal(got, []View{ViewDefault, View("habit_list")}) {
		t.Fatalf("Issues tabs = %v", got)
	}
	if ViewLabel(View("habit_list")) != "Habits" {
		t.Fatal("missing Habits label")
	}
	if len(ViewPanes(View("meta"))) > 0 {
		t.Fatal("Meta still registered")
	}
}
