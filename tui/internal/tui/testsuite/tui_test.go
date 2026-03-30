package testsuite

import (
	"fmt"
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	dialogs "crona/tui/internal/tui/dialogs"
	"crona/tui/internal/tui/testsuite/support"
	"crona/tui/internal/tui/views"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestPaneActionLineWrapsInsteadOfDroppingActions(t *testing.T) {
	rendered := views.RenderPaneActionLine(
		support.Theme(),
		"",
		20,
		[]string{"[enter] view", "[a] new", "[c] context"},
	)
	lines := strings.Split(rendered, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected wrapped action lines, got %q", rendered)
	}
	if !strings.Contains(rendered, "[c] context") {
		t.Fatalf("expected final action to be preserved, got %q", rendered)
	}
	for _, line := range lines {
		if got := lipgloss.Width(line); got > 20 {
			t.Fatalf("line width %d exceeds max width 20: %q", got, line)
		}
	}
}

func TestPaneBoxClipsOverflowingContent(t *testing.T) {
	rendered := support.RenderPaneBox(support.Theme(), true, 40, 8, strings.Join([]string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
		"line6",
		"line7",
		"line8",
	}, "\n"))
	if got := lipgloss.Height(rendered); got > 8 {
		t.Fatalf("pane box height %d exceeds allotted height 8", got)
	}
}

func TestDailyViewStacksOnNarrowWidths(t *testing.T) {
	state := views.ContentState{
		View:   "daily",
		Pane:   "issues",
		Width:  32,
		Height: 22,
		Cursors: map[string]int{
			"issues": 0,
			"habits": 0,
		},
		Filters: map[string]string{
			"issues": "",
			"habits": "",
		},
	}

	rendered := support.RenderDaily(state)
	for _, line := range strings.Split(rendered, "\n") {
		if got := lipgloss.Width(line); got > state.Width {
			t.Fatalf("daily view line width %d exceeds content width %d: %q", got, state.Width, line)
		}
	}
}

func TestDailyViewDoesNotExceedAllocatedHeight(t *testing.T) {
	state := views.ContentState{
		View:   "daily",
		Pane:   "issues",
		Width:  80,
		Height: 14,
		Cursors: map[string]int{
			"issues": 0,
			"habits": 0,
		},
		Filters: map[string]string{
			"issues": "",
			"habits": "",
		},
	}

	rendered := support.RenderDaily(state)
	if got := lipgloss.Height(rendered); got > state.Height {
		t.Fatalf("daily view height %d exceeds allocated height %d", got, state.Height)
	}
}

func TestDailyViewReportedHeightRangeFitsAllocation(t *testing.T) {
	estimate := 60
	target := 15
	state := views.ContentState{
		View:   "daily",
		Pane:   "issues",
		Width:  70,
		Height: 43,
		Cursors: map[string]int{
			"issues": 0,
			"habits": 0,
		},
		Filters: map[string]string{
			"issues": "",
			"habits": "",
		},
		DailySummary: &api.DailyIssueSummary{
			Date: "2026-03-19",
			Issues: []api.Issue{
				{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
			},
		},
		DailyIssues: []api.Issue{
			{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
		},
		AllIssues: []api.IssueWithMeta{
			{Issue: api.Issue{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate}, RepoName: "Work", StreamName: "app"},
		},
		DueHabits: []api.HabitDailyItem{
			{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{Name: "Inbox Zero Sweep", TargetMinutes: &target}}, Status: "pending"},
		},
		Context: &api.ActiveContext{
			RepoName:   strPtr("Work"),
			StreamName: strPtr("app"),
		},
	}

	rendered := support.RenderDaily(state)
	if got := lipgloss.Height(rendered); got > state.Height {
		t.Fatalf("daily view height %d exceeds allocated height %d", got, state.Height)
	}
}

func TestAwayViewShowsProtectedModeBlankState(t *testing.T) {
	state := views.ContentState{
		View:            "away",
		Pane:            "",
		Width:           80,
		Height:          20,
		RestModeActive:  true,
		AwayModeActive:  true,
		RestModeMessage: "Enjoy your break",
		RestModeDetail:  "Away mode is active.",
	}

	rendered := views.RenderContent(support.Theme(), state)
	for _, want := range []string{"Away", "Enjoy your break", "Away mode is active.", "[w]", "disable away"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected away view to contain %q, got %q", want, rendered)
		}
	}
	for _, unwanted := range []string{"Issues", "Habits", "Daily Dashboard", "Wellbeing"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("expected away view to hide %q, got %q", unwanted, rendered)
		}
	}
}

func TestAwayViewHidesDisableActionOnConfiguredRestDay(t *testing.T) {
	state := views.ContentState{
		View:            "away",
		Pane:            "",
		Width:           80,
		Height:          20,
		RestModeActive:  true,
		AwayModeActive:  false,
		RestModeMessage: "Take the day lightly",
		RestModeDetail:  "This is one of your configured rest days.",
	}

	rendered := views.RenderContent(support.Theme(), state)
	for _, want := range []string{"Take the day lightly", "configured rest days"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected away view to contain %q, got %q", want, rendered)
		}
	}
	for _, unwanted := range []string{"Metrics Window", "Mood", "Energy", "[w]"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("expected away view to hide %q, got %q", unwanted, rendered)
		}
	}
}

func TestDailyViewDoesNotExceedTerminalHeightInReportedRange(t *testing.T) {
	for height := 46; height <= 54; height++ {
		model := support.NewDailyModel(92, height)
		if got, want := model.BodyHeight(), model.ContentHeight(); got > want {
			t.Fatalf("daily body height %d exceeds content height %d at terminal height %d", got, want, height)
		}
		rendered := model.RenderString()
		if got := lipgloss.Height(rendered); got > height {
			t.Fatalf("daily view height %d exceeds terminal height %d", got, height)
		}
	}
}

func TestDailySummaryUsesCompactInlineModeBelowHeight55(t *testing.T) {
	estimate := 60
	target := 15
	state := views.ContentState{
		View:   "daily",
		Pane:   "issues",
		Width:  70,
		Height: 54,
		Cursors: map[string]int{
			"issues": 0,
			"habits": 0,
		},
		Filters: map[string]string{
			"issues": "",
			"habits": "",
		},
		DailySummary: &api.DailyIssueSummary{
			Date: "2026-03-19",
			Issues: []api.Issue{
				{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
			},
		},
		DailyIssues: []api.Issue{
			{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
		},
		DueHabits: []api.HabitDailyItem{
			{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{Name: "Inbox Zero Sweep", TargetMinutes: &target}}, Status: "pending"},
		},
		Context: &api.ActiveContext{
			RepoName:   strPtr("Work"),
			StreamName: strPtr("app"),
		},
	}

	rendered := support.RenderDaily(state)
	if !strings.Contains(rendered, "Issues  0/1 resolved") {
		t.Fatalf("expected compact inline issues row below height 55")
	}
	if !strings.Contains(rendered, "Habits  0/1 completed") {
		t.Fatalf("expected compact inline habits row below height 55")
	}
	if !strings.Contains(rendered, "planned 1") {
		t.Fatalf("expected compact legend text below height 55")
	}
	if !strings.Contains(rendered, "logged 0m / target 15m") {
		t.Fatalf("expected compact habit meta below height 55")
	}
	if !strings.Contains(rendered, "█") {
		t.Fatalf("expected inline bars below height 55")
	}
}

func TestResolvedIssuesShowResolutionDateInsteadOfDueLabel(t *testing.T) {
	estimate1, estimate2, estimate3 := 60, 35, 25
	due := "2026-03-19"
	completedAt := "2026-03-20T09:30:00Z"
	abandonedAt := "2026-03-21T18:45:00Z"
	state := views.ContentState{
		View:   "default",
		Pane:   "issues",
		Width:  100,
		Height: 32,
		Cursors: map[string]int{
			"issues": 0,
		},
		Filters: map[string]string{
			"issues": "",
		},
		DefaultIssueSection: "completed",
		DefaultIssues: []api.IssueWithMeta{
			{Issue: api.Issue{ID: 1, Title: "Open planned issue", Status: "planned", EstimateMinutes: &estimate1, TodoForDate: &due}, RepoName: "Work", StreamName: "app"},
			{Issue: api.Issue{ID: 2, Title: "Completed shipped fix", Status: "done", EstimateMinutes: &estimate2, TodoForDate: &due, CompletedAt: &completedAt}, RepoName: "Work", StreamName: "app"},
			{Issue: api.Issue{ID: 3, Title: "Abandoned old task", Status: "abandoned", EstimateMinutes: &estimate3, TodoForDate: &due, AbandonedAt: &abandonedAt}, RepoName: "Personal", StreamName: "home"},
		},
		Context: &api.ActiveContext{},
	}

	rendered := ansi.Strip(support.RenderDefault(state))
	if !strings.Contains(rendered, "Completed shipped fix  [on 2026-03-20]") {
		t.Fatalf("expected completed issue to show resolution date, got %q", rendered)
	}
	if !strings.Contains(rendered, "Abandoned old task  [on 2026-03-21]") {
		t.Fatalf("expected abandoned issue to show resolution date, got %q", rendered)
	}
	if strings.Contains(rendered, "Completed shipped fix  [due ") || strings.Contains(rendered, "Completed shipped fix  [overdue ") {
		t.Fatalf("completed issue should not show due or overdue label, got %q", rendered)
	}
	if strings.Contains(rendered, "Abandoned old task  [due ") || strings.Contains(rendered, "Abandoned old task  [overdue ") {
		t.Fatalf("abandoned issue should not show due or overdue label, got %q", rendered)
	}
}

func TestExportDialogListsPhase3ReportChoices(t *testing.T) {
	repos := []api.Repo{{ID: 1, Name: "Work"}, {ID: 2, Name: "OSS"}}
	checkedRepoID := int64(2)
	state := dialogs.OpenExportDaily(dialogs.State{}, "2026-03-19", true, repos, &checkedRepoID)
	if state.Kind != "export_report" {
		t.Fatalf("expected export_report dialog kind, got %q", state.Kind)
	}
	if state.RepoID != checkedRepoID || state.RepoName != "OSS" {
		t.Fatalf("expected checked-out repo to be selected, got id=%d name=%q", state.RepoID, state.RepoName)
	}
	joined := strings.Join(state.ChoiceItems, "\n")
	for _, want := range []string{
		"Daily report: write Markdown file",
		"Weekly summary: write Markdown file",
		"Repo report: write Markdown file",
		"Stream report: write Markdown file",
		"Issue rollup: write Markdown file",
		"CSV session export: write file",
		"Calendar export: write ICS file",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected export choice %q in dialog", want)
		}
	}
}

func TestExportDialogDefaultsCalendarRepoToFirstRepoWhenNoContextRepo(t *testing.T) {
	repos := []api.Repo{{ID: 5, Name: "Work"}, {ID: 9, Name: "Personal"}}
	state := dialogs.OpenExportDaily(dialogs.State{}, "2026-03-19", false, repos, nil)
	if state.RepoID != 5 || state.RepoName != "Work" || state.RepoIndex != 0 {
		t.Fatalf("expected first repo selected by default, got id=%d name=%q index=%d", state.RepoID, state.RepoName, state.RepoIndex)
	}
}

func TestExportDialogCalendarChoiceOpensRepoPicker(t *testing.T) {
	repos := []api.Repo{{ID: 5, Name: "Work"}, {ID: 9, Name: "Personal"}}
	state := dialogs.OpenExportDaily(dialogs.State{}, "2026-03-19", false, repos, nil)
	for i, item := range state.ChoiceItems {
		if item == "Calendar export: write ICS file" {
			state.ChoiceCursor = i
			break
		}
	}
	next, action, status := dialogs.Update(state, dialogs.UpdateContext{}, "2026-03-19", tea.KeyMsg{Type: tea.KeyEnter})
	if status != "" {
		t.Fatalf("unexpected status: %s", status)
	}
	if action != nil {
		t.Fatalf("expected repo picker step before export action")
	}
	if next.Kind != "export_calendar_repo" {
		t.Fatalf("expected export_calendar_repo dialog, got %q", next.Kind)
	}
	if len(next.ChoiceItems) != 2 || next.ChoiceItems[0] != "Work" || next.ChoiceItems[1] != "Personal" {
		t.Fatalf("unexpected repo picker options: %#v", next.ChoiceItems)
	}
}

func TestSettingsViewShowsBoundaryNotificationToggles(t *testing.T) {
	state := views.ContentState{
		View:   "settings",
		Pane:   "settings",
		Width:  70,
		Height: 32,
		Cursors: map[string]int{
			"settings": 0,
		},
		Filters: map[string]string{
			"settings": "",
		},
		Settings: &api.CoreSettings{
			TimerMode:             sharedtypes.TimerModeStructured,
			BreaksEnabled:         true,
			WorkDurationMinutes:   25,
			ShortBreakMinutes:     5,
			LongBreakMinutes:      15,
			LongBreakEnabled:      true,
			CyclesBeforeLongBreak: 4,
			AutoStartBreaks:       true,
			AutoStartWork:         false,
			BoundaryNotifications: true,
			BoundarySound:         true,
			RepoSort:              sharedtypes.RepoSortChronologicalAsc,
			StreamSort:            sharedtypes.StreamSortChronologicalAsc,
			IssueSort:             sharedtypes.IssueSortPriority,
			HabitSort:             sharedtypes.HabitSortTargetMinutesDesc,
		},
	}

	rendered := support.RenderSettings(state)
	for _, want := range []string{"Boundary Notifications", "Boundary Sound", "Habit Sort", "Target longest", "Away Mode"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected settings view to contain %q, got %q", want, rendered)
		}
	}

	state.Cursors["settings"] = 18
	rendered = support.RenderSettings(state)
	for _, want := range []string{"Rest & Streak Protection", "All streaks"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected settings view to contain %q, got %q", want, rendered)
		}
	}
}

func TestReportsViewActionsExposeEditOpenDeleteSeparately(t *testing.T) {
	actions := views.ContextualActions(support.Theme(), views.ActionsState{
		View: "reports",
		Pane: "export_reports",
	})
	joined := strings.Join(actions, " ")
	for _, want := range []string{"[e]", "edit", "[o]", "open", "[d]", "delete", "[enter]", "details"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected reports actions to contain %q, got %q", want, joined)
		}
	}
}

func TestGlobalActionsExposeUpdatesShortcutWhenVisible(t *testing.T) {
	actions := views.GlobalActions(support.Theme(), views.ActionsState{
		View:          "daily",
		Pane:          "issues",
		UpdateVisible: true,
	})
	joined := strings.Join(actions, " ")
	for _, want := range []string{"[u]", "updates"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected update actions to contain %q, got %q", want, joined)
		}
	}
}

func TestUpdatesViewActionsExposeCheckOpenInstallDismiss(t *testing.T) {
	actions := views.ContextualActions(support.Theme(), views.ActionsState{
		View:                   "updates",
		UpdateInstallAvailable: true,
	})
	joined := strings.Join(actions, " ")
	for _, want := range []string{"[r]", "check now", "[o]", "open release", "[i]", "install", "[U]", "dismiss"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected updates actions to contain %q, got %q", want, joined)
		}
	}
}

func TestUpdatesViewShowsInstallUnavailableReason(t *testing.T) {
	rendered := support.RenderUpdates(views.ContentState{
		View:   "updates",
		Pane:   "issues",
		Width:  100,
		Height: 24,
		UpdateStatus: &api.UpdateStatus{
			Enabled:                  true,
			PromptEnabled:            true,
			UpdateAvailable:          true,
			LatestVersion:            "0.3.0",
			InstallAvailable:         false,
			InstallUnavailableReason: "Release is missing the checksums.txt asset.",
		},
	})
	for _, want := range []string{"[i] install unavailable", "Release is missing the checksums.txt asset."} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected updates view to contain %q, got %q", want, rendered)
		}
	}
}

func TestConfigViewShowsSeparateICSExportDirectory(t *testing.T) {
	state := views.ContentState{
		View:   "config",
		Pane:   "config",
		Width:  90,
		Height: 18,
		Cursors: map[string]int{
			"config": 0,
		},
		Filters: map[string]string{
			"config": "",
		},
		ExportAssets: &api.ExportAssetStatus{
			ReportsDir: "/tmp/reports",
			ICSDir:     "/tmp/calendar",
		},
	}
	rendered := support.RenderConfig(state)
	for _, want := range []string{"Reports directory", "ICS export directory", "/tmp/calendar"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected config view to contain %q, got %q", want, rendered)
		}
	}
}

func TestExportReportsViewShowsReportKindsAndScopeLabels(t *testing.T) {
	state := views.ContentState{
		View:   "reports",
		Pane:   "export_reports",
		Width:  90,
		Height: 16,
		Cursors: map[string]int{
			"export_reports": 0,
		},
		Filters: map[string]string{
			"export_reports": "",
		},
		ExportAssets: &api.ExportAssetStatus{ReportsDir: "/tmp/reports"},
		ExportReports: []api.ExportReportFile{
			{
				Name:       "weekly-2026-03-17-to-2026-03-23.md",
				Path:       "/tmp/reports/weekly-2026-03-17-to-2026-03-23.md",
				Kind:       sharedtypes.ExportReportKindWeekly,
				ScopeLabel: "Work / app",
				DateLabel:  "2026-03-17 to 2026-03-23",
				Format:     string(sharedtypes.ExportFormatMarkdown),
				SizeBytes:  2048,
			},
		},
	}

	rendered := support.RenderReports(state)
	for _, want := range []string{"Reports", "[weekly]", "Work / app", "2026-03-17 to 2026-03-23"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected export reports view to contain %q, got %q", want, rendered)
		}
	}
}

func TestDailySummaryShowsBarsAtHeight55AndAbove(t *testing.T) {
	estimate := 60
	target := 15
	state := views.ContentState{
		View:   "daily",
		Pane:   "issues",
		Width:  70,
		Height: 55,
		Cursors: map[string]int{
			"issues": 0,
			"habits": 0,
		},
		Filters: map[string]string{
			"issues": "",
			"habits": "",
		},
		DailySummary: &api.DailyIssueSummary{
			Date: "2026-03-19",
			Issues: []api.Issue{
				{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
			},
		},
		DailyIssues: []api.Issue{
			{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
		},
		DueHabits: []api.HabitDailyItem{
			{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{Name: "Inbox Zero Sweep", TargetMinutes: &target}}, Status: "pending"},
		},
		Context: &api.ActiveContext{
			RepoName:   strPtr("Work"),
			StreamName: strPtr("app"),
		},
	}

	rendered := support.RenderDaily(state)
	if !strings.Contains(rendered, "████") {
		t.Fatalf("expected bars to remain visible at height 55 and above")
	}
}

func TestDailySummaryUsesUltraCompactModeBelowHeight48(t *testing.T) {
	estimate := 60
	target := 15
	state := views.ContentState{
		View:   "daily",
		Pane:   "issues",
		Width:  70,
		Height: 46,
		Cursors: map[string]int{
			"issues": 0,
			"habits": 0,
		},
		Filters: map[string]string{
			"issues": "",
			"habits": "",
		},
		DailySummary: &api.DailyIssueSummary{
			Date: "2026-03-19",
			Issues: []api.Issue{
				{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
			},
		},
		DailyIssues: []api.Issue{
			{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
		},
		DueHabits: []api.HabitDailyItem{
			{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{Name: "Inbox Zero Sweep", TargetMinutes: &target}}, Status: "pending"},
		},
		Context: &api.ActiveContext{
			RepoName:   strPtr("Work"),
			StreamName: strPtr("app"),
		},
	}

	rendered := support.RenderDaily(state)
	if !strings.Contains(rendered, "Issues  0/1 resolved") || !strings.Contains(rendered, "Habits  0/1 completed") {
		t.Fatalf("expected ultra-compact rows for both issues and habits")
	}
	if strings.Contains(rendered, "planned 1") {
		t.Fatalf("expected issue legend row to be omitted below height 48")
	}
	if strings.Contains(rendered, "failed 0   remaining 1") {
		t.Fatalf("expected habit meta row to be omitted below height 48")
	}
	if !strings.Contains(rendered, "█") {
		t.Fatalf("expected inline bars to remain in ultra-compact mode")
	}
}

func TestDailySummaryUsesTinyHeightModeAt36(t *testing.T) {
	estimate := 60
	target := 15
	state := views.ContentState{
		View:   "daily",
		Pane:   "issues",
		Width:  70,
		Height: 36,
		Cursors: map[string]int{
			"issues": 0,
			"habits": 0,
		},
		Filters: map[string]string{
			"issues": "",
			"habits": "",
		},
		DailySummary: &api.DailyIssueSummary{
			Date: "2026-03-19",
			Issues: []api.Issue{
				{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
			},
		},
		DailyIssues: []api.Issue{
			{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
		},
		DueHabits: []api.HabitDailyItem{
			{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{Name: "Inbox Zero Sweep", TargetMinutes: &target}}, Status: "pending"},
		},
		Context: &api.ActiveContext{
			RepoName:   strPtr("Work"),
			StreamName: strPtr("app"),
		},
	}

	rendered := support.RenderDaily(state)
	assertTinySummary(t, rendered)
}

func TestDailySummaryUsesTinyHeightModeAt30(t *testing.T) {
	estimate := 60
	target := 15
	state := views.ContentState{
		View:   "daily",
		Pane:   "issues",
		Width:  70,
		Height: 30,
		Cursors: map[string]int{
			"issues": 0,
			"habits": 0,
		},
		Filters: map[string]string{
			"issues": "",
			"habits": "",
		},
		DailySummary: &api.DailyIssueSummary{
			Date: "2026-03-19",
			Issues: []api.Issue{
				{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
			},
		},
		DailyIssues: []api.Issue{
			{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate},
		},
		DueHabits: []api.HabitDailyItem{
			{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{Name: "Inbox Zero Sweep", TargetMinutes: &target}}, Status: "pending"},
		},
		Context: &api.ActiveContext{
			RepoName:   strPtr("Work"),
			StreamName: strPtr("app"),
		},
	}

	rendered := support.RenderDaily(state)
	assertTinySummary(t, rendered)
}

func TestDefaultViewUsesCompactModeAt36(t *testing.T) {
	state := compactDefaultState(36)
	rendered := support.RenderDefault(state)
	assertCompactDefault(t, rendered, state.Height)
}

func TestDefaultViewUsesCompactModeAt30(t *testing.T) {
	state := compactDefaultState(30)
	rendered := support.RenderDefault(state)
	assertCompactDefault(t, rendered, state.Height)
}

func TestWellbeingViewUsesCompactModeAt36(t *testing.T) {
	state := compactWellbeingState(36)
	rendered := support.RenderWellbeing(state)
	assertCompactWellbeing(t, rendered, state.Height)
}

func TestWellbeingViewUsesCompactModeAt30(t *testing.T) {
	state := compactWellbeingState(30)
	rendered := support.RenderWellbeing(state)
	assertCompactWellbeing(t, rendered, state.Height)
}

func TestUndersizedWidthShowsMinimumSizeWarning(t *testing.T) {
	minWidth, minHeight := support.MinimumSize()
	model := support.NewDailyModel(minWidth-1, minHeight)
	rendered := model.RenderString()
	assertMinimumSizeWarning(t, rendered, minWidth-1, minHeight, minWidth, minHeight)
}

func TestUndersizedHeightShowsMinimumSizeWarning(t *testing.T) {
	minWidth, minHeight := support.MinimumSize()
	model := support.NewDailyModel(minWidth, minHeight-1)
	rendered := model.RenderString()
	assertMinimumSizeWarning(t, rendered, minWidth, minHeight-1, minWidth, minHeight)
}

func TestUndersizedBothDimensionsShowMinimumSizeWarning(t *testing.T) {
	minWidth, minHeight := support.MinimumSize()
	model := support.NewDailyModel(minWidth-5, minHeight-2)
	rendered := model.RenderString()
	assertMinimumSizeWarning(t, rendered, minWidth-5, minHeight-2, minWidth, minHeight)
}

func TestMinimumSizeThresholdRendersNormalUI(t *testing.T) {
	minWidth, minHeight := support.MinimumSize()
	model := support.NewDailyModel(minWidth, minHeight)
	rendered := model.RenderString()
	if strings.Contains(rendered, "Terminal Too Small") {
		t.Fatalf("expected normal UI at minimum size")
	}
	if !strings.Contains(rendered, "Daily Dashboard") {
		t.Fatalf("expected daily UI at minimum size")
	}
	if got := lipgloss.Height(rendered); got > minHeight {
		t.Fatalf("rendered height %d exceeds terminal height %d", got, minHeight)
	}
}

func TestAboveMinimumSizeRendersNormalUI(t *testing.T) {
	minWidth, minHeight := support.MinimumSize()
	model := support.NewDailyModel(minWidth+1, minHeight+1)
	rendered := model.RenderString()
	if strings.Contains(rendered, "Terminal Too Small") {
		t.Fatalf("expected normal UI above minimum size")
	}
	if !strings.Contains(rendered, "Daily Dashboard") {
		t.Fatalf("expected daily UI above minimum size")
	}
}

func assertMinimumSizeWarning(t *testing.T, rendered string, currentWidth, currentHeight, minWidth, minHeight int) {
	t.Helper()
	if !strings.Contains(rendered, "Terminal Too Small") {
		t.Fatalf("expected undersized warning, got %q", rendered)
	}
	if !strings.Contains(rendered, fmt.Sprintf("Current: %dx%d", currentWidth, currentHeight)) {
		t.Fatalf("expected current dimensions in warning")
	}
	if !strings.Contains(rendered, fmt.Sprintf("Required: %dx%d", minWidth, minHeight)) {
		t.Fatalf("expected required dimensions in warning")
	}
	if !strings.Contains(rendered, "Resize the terminal to continue.") {
		t.Fatalf("expected resize instruction in warning")
	}
	if got := lipgloss.Width(rendered); got > currentWidth {
		t.Fatalf("warning width %d exceeds viewport width %d", got, currentWidth)
	}
	if got := lipgloss.Height(rendered); got > currentHeight {
		t.Fatalf("warning height %d exceeds viewport height %d", got, currentHeight)
	}
	if strings.Contains(rendered, "Daily Dashboard") {
		t.Fatalf("expected normal UI to be suppressed while undersized")
	}
}

func strPtr(v string) *string { return &v }

func assertTinySummary(t *testing.T, rendered string) {
	t.Helper()
	if !strings.Contains(rendered, "Daily Dashboard") {
		t.Fatalf("expected dashboard title in tiny summary")
	}
	if !strings.Contains(rendered, "2026-03-19") {
		t.Fatalf("expected date in tiny summary")
	}
	if !strings.Contains(rendered, "Scope: Work > app") {
		t.Fatalf("expected scope in tiny summary")
	}
	if !strings.Contains(rendered, "[,] [.] [g]") {
		t.Fatalf("expected compact date hints in tiny summary")
	}
	if !strings.Contains(rendered, "Issues  0/1") || !strings.Contains(rendered, "Habits  0/1") {
		t.Fatalf("expected both issue and habit summary rows in tiny summary")
	}
	if !strings.Contains(rendered, "p1") {
		t.Fatalf("expected abbreviated issue legend in tiny summary")
	}
	if !strings.Contains(rendered, "f0 r1") {
		t.Fatalf("expected abbreviated habit tail in tiny summary")
	}
	if !strings.Contains(rendered, "█") {
		t.Fatalf("expected micro-bars in tiny summary")
	}
}

func compactDefaultState(height int) views.ContentState {
	estimate1, estimate2, estimate3 := 60, 35, 25
	today := "2026-03-19"
	return views.ContentState{
		View:   "default",
		Pane:   "issues",
		Width:  92,
		Height: height,
		Cursors: map[string]int{
			"issues": 0,
		},
		Filters: map[string]string{
			"issues": "",
		},
		DefaultIssueSection: "open",
		DefaultIssues: []api.IssueWithMeta{
			{Issue: api.Issue{ID: 1, Title: "Add keyboard-first command palette", Status: "planned", EstimateMinutes: &estimate1, TodoForDate: &today}, RepoName: "Work", StreamName: "app"},
			{Issue: api.Issue{ID: 2, Title: "Improve install docs for Linux", Status: "planned", EstimateMinutes: &estimate2, TodoForDate: &today}, RepoName: "OSS", StreamName: "cli"},
			{Issue: api.Issue{ID: 3, Title: "Research standing desk options", Status: "abandoned", EstimateMinutes: &estimate3}, RepoName: "Personal", StreamName: "home"},
		},
		Context: &api.ActiveContext{},
	}
}

func compactWellbeingState(height int) views.ContentState {
	avgMood, avgEnergy := 4.0, 3.7
	return views.ContentState{
		View:          "wellbeing",
		Pane:          "issues",
		Width:         92,
		Height:        height,
		WellbeingDate: "2026-03-19",
		DailyPlan: &api.DailyPlan{
			Date: "2026-03-19",
			Summary: api.DailyPlanAccountabilitySummary{
				PlannedCount:         3,
				CompletedCount:       1,
				FailedCount:          1,
				PendingRollbackCount: 1,
				AccountabilityScore:  2.6,
				DelayedIssueCount:    1,
			},
			Entries: []api.DailyPlanEntry{
				{ID: "1", Date: "2026-03-19", IssueID: 1, Status: "completed", CommittedAt: "2026-03-19T08:00:00Z"},
				{ID: "2", Date: "2026-03-19", IssueID: 2, Status: "failed", CommittedAt: "2026-03-19T08:00:00Z", FailureReason: dailyPlanFailureReasonPtr("moved")},
				{ID: "3", Date: "2026-03-19", IssueID: 3, Status: "planned", CommittedAt: "2026-03-19T08:00:00Z", PendingFailureAt: strPtr("2026-03-19T10:00:00Z")},
			},
		},
		MetricsRollup: &api.MetricsRollup{
			Days:          7,
			CheckInDays:   6,
			FocusDays:     1,
			WorkedSeconds: 4956,
			RestSeconds:   2,
			AverageMood:   &avgMood,
			AverageEnergy: &avgEnergy,
			LatestBurnout: &api.BurnoutIndicator{
				Level:   "low",
				Score:   31,
				Factors: map[string]float64{"breakCompliance": 0.99},
			},
		},
		Streaks: &api.StreakSummary{
			CurrentCheckInDays: 0,
			LongestCheckInDays: 6,
			CurrentFocusDays:   0,
			LongestFocusDays:   1,
		},
	}
}

func assertCompactDefault(t *testing.T, rendered string, height int) {
	t.Helper()
	plain := ansi.Strip(rendered)
	if !strings.Contains(plain, "Due 2") || !strings.Contains(plain, "Open 2") || !strings.Contains(plain, "Closed 1") {
		t.Fatalf("expected compact stats header in default view")
	}
	if !strings.Contains(plain, "Active Issues [1]") {
		t.Fatalf("expected primary issue list in compact default view")
	}
	if !strings.Contains(plain, "Closed") {
		t.Fatalf("expected compact completed footer in default view")
	}
	if !strings.Contains(plain, "Add keyboard-first command palette") {
		t.Fatalf("expected open issue rows in compact default view")
	}
	if got := lipgloss.Height(rendered); got > height {
		t.Fatalf("default compact view height %d exceeds allocated height %d", got, height)
	}
}

func assertCompactWellbeing(t *testing.T, rendered string, height int) {
	t.Helper()
	plain := ansi.Strip(rendered)
	if !strings.Contains(plain, "Wellbeing") || !strings.Contains(plain, "2026-03-19") {
		t.Fatalf("expected compact wellbeing header")
	}
	if !strings.Contains(plain, "[,/.]") && !strings.Contains(plain, "[a/e]") {
		t.Fatalf("expected action hints in compact wellbeing view")
	}
	if !strings.Contains(plain, "No check-in recorded for this date") {
		t.Fatalf("expected current day summary in compact wellbeing view")
	}
	if !strings.Contains(plain, "Burnout") || !strings.Contains(plain, "31 LOW") {
		t.Fatalf("expected burnout summary in compact wellbeing view")
	}
	if !strings.Contains(plain, "Metrics Window") || !strings.Contains(plain, "Days  7") {
		t.Fatalf("expected compact metrics block in wellbeing view")
	}
	if !strings.Contains(plain, "Accountability") || !strings.Contains(plain, "p3 c1 f1 r1 s2.6 d1") {
		t.Fatalf("expected compact accountability summary in wellbeing view")
	}
	if got := lipgloss.Height(rendered); got > height {
		t.Fatalf("wellbeing compact view height %d exceeds allocated height %d", got, height)
	}
}

func dailyPlanFailureReasonPtr(value string) *api.DailyPlanFailureReason {
	reason := api.DailyPlanFailureReason(value)
	return &reason
}

func TestDailyHabitDeleteDialogUsesDailySelection(t *testing.T) {
	model := support.NewDailyHabitDeleteModel([]api.HabitDailyItem{
		{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: 42, StreamID: 7, Name: "Inbox Zero"}}},
	})

	next, ok := support.OpenSelectedDeleteDialog(model)
	if !ok {
		t.Fatalf("expected delete dialog to open for daily habit")
	}
	if next.DialogDeleteKind() != "habit" || next.DialogDeleteID() != "42" {
		t.Fatalf("expected habit delete dialog, got kind=%q id=%q", next.DialogDeleteKind(), next.DialogDeleteID())
	}
	if next.DialogStreamID() != 7 {
		t.Fatalf("expected dialog stream id 7, got %d", next.DialogStreamID())
	}
}

func TestDefaultStreamOptionsIncludeExistingStreamsWithoutContext(t *testing.T) {
	repoInput := textinput.New()
	repoInput.SetValue("Work")
	streamInput := textinput.New()
	streamInput.SetValue(" app ")

	options := support.DefaultStreamOptions(
		[]textinput.Model{repoInput, streamInput},
		0,
		[]api.Repo{{ID: 1, Name: "Work"}},
		nil,
		[]api.Stream{{ID: 9, RepoID: 1, Name: "app"}},
		nil,
	)

	if len(options) == 0 || options[0].ID != "9" || options[0].Label != "app" {
		t.Fatalf("expected existing stream option first, got %+v", options)
	}
}

func TestMatchStreamSelectionNormalizesWhitespaceAndCase(t *testing.T) {
	streamID, streamName := support.MatchStreamSelection(
		"  APP  ",
		1,
		"Work",
		0,
		[]api.Repo{{ID: 1, Name: "Work"}},
		nil,
		[]api.Stream{{ID: 9, RepoID: 1, Name: "app"}},
		nil,
	)

	if streamID != 9 || streamName != "app" {
		t.Fatalf("expected normalized existing stream match, got %d %q", streamID, streamName)
	}
}
