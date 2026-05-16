package model

import (
	"fmt"
	"slices"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	navigationutil "crona/tui/internal/tui/navigationutil"
	selectionpkg "crona/tui/internal/tui/selection"
	uistate "crona/tui/internal/tui/state"
	issuecore "crona/tui/internal/tui/views/issuecore"
	viewruntime "crona/tui/internal/tui/views/runtime"
)

func (m Model) availableViews() []View {
	if protected, _, _ := viewruntime.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02")); protected {
		return []View{ViewAway, ViewReports, ViewSessionHistory, ViewHabitHistory}
	}
	ordered := uistate.ViewOrder()
	views := make([]View, 0, len(ordered))
	views = append(views, ordered...)
	if len(views) == 0 {
		return []View{ViewDaily}
	}
	return views
}

func (m Model) jumpAvailableViews() []View {
	if protected, _, _ := viewruntime.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02")); protected {
		return []View{ViewAway, ViewReports, ViewSessionHistory, ViewHabitHistory}
	}
	if m.timer != nil && m.timer.State != "idle" {
		return []View{ViewSessionActive, ViewSessionHistory, ViewScratch}
	}
	return m.availableViews()
}

func (m Model) canJumpToView(target View) bool {
	canJump := slices.Contains(m.jumpAvailableViews(), target)

	return canJump
}

func (m Model) nextWorkspaceView(dir int) View {
	views := m.availableViews()
	for _, candidate := range views {
		if candidate == m.view {
			return navigationutil.NextView(views, candidate, dir)
		}
	}
	if len(views) > 0 {
		return views[0]
	}
	return ViewDaily
}

func (m Model) nextActiveSessionView(dir int) View {
	views := []View{ViewSessionActive, ViewSessionHistory, ViewScratch}
	current := ViewSessionActive
	for _, candidate := range views {
		if m.view == candidate {
			current = candidate
			break
		}
	}
	for i, candidate := range views {
		if candidate == current {
			return views[(i+dir+len(views))%len(views)]
		}
	}
	return ViewSessionActive
}

func nextPane(view View, current Pane, dir int) Pane {
	return uistate.NextPane(view, current, dir)
}

func (m *Model) setDefaultIssueSection(section DefaultIssueSection) {
	if section != DefaultIssueSectionOpen && section != DefaultIssueSectionCompleted {
		section = DefaultIssueSectionOpen
	}
	m.defaultIssueSection = section
	if m.view != ViewDefault || m.pane != PaneIssues {
		return
	}
	openIndices, completedIndices := issuecore.SplitDefaultIssueIndices(m.allIssues, m.filters[PaneIssues], m.settings)
	switch section {
	case DefaultIssueSectionOpen:
		if len(openIndices) > 0 {
			m.cursor[PaneIssues] = 0
		}
	case DefaultIssueSectionCompleted:
		if len(completedIndices) > 0 {
			m.cursor[PaneIssues] = len(openIndices)
		}
	}
	snapshot := m.selectionSnapshot()
	indices := selectionpkg.FilteredIndices(snapshot, PaneIssues)
	m.clamp(PaneIssues, len(indices))
}

func (m *Model) setDailyTaskSection(section DailyTaskSection) {
	if section != DailyTaskSectionPlanned && section != DailyTaskSectionPinned && section != DailyTaskSectionOverdue {
		section = DailyTaskSectionPlanned
	}
	m.dailyTaskSection = section
	if m.view != ViewDaily || m.pane != PaneIssues {
		return
	}
	snapshot := m.selectionSnapshot()
	indices := selectionpkg.FilteredIndices(snapshot, PaneIssues)
	m.clamp(PaneIssues, len(indices))
}

func (m Model) selectedMetaRepo() (int64, string, bool) {
	return selectionpkg.SelectedMetaRepo(m.selectionSnapshot())
}

func (m Model) selectedMetaStream() (int64, string, string, bool) {
	return selectionpkg.SelectedMetaStream(m.selectionSnapshot())
}

func (m Model) selectedIssueRecord() (*api.Issue, bool) {
	snapshot := m.selectionSnapshot()
	if m.timer != nil && m.timer.State != "idle" && (m.view == ViewSessionActive || m.view == ViewScratch) {
		if issue := selectionpkg.ActiveIssue(snapshot); issue != nil {
			copy := issue.Issue
			return &copy, true
		}
	}
	if m.pane != PaneIssues {
		return nil, false
	}
	return selectionpkg.SelectedIssue(snapshot)
}

func (m Model) selectedHabitRecord() (*api.Habit, bool) {
	return selectionpkg.SelectedHabit(m.selectionSnapshot())
}

func (m Model) selectedDailyHabitRecord() (*api.HabitDailyItem, bool) {
	return selectionpkg.SelectedDailyHabit(m.selectionSnapshot())
}

func (m Model) selectedHabitForAction() (*api.Habit, bool) {
	if m.view == ViewDaily {
		if habit, ok := m.selectedDailyHabitRecord(); ok {
			copy := habit.Habit
			return &copy, true
		}
		return nil, false
	}
	return m.selectedHabitRecord()
}

func (m Model) selectedSessionHistoryEntry() (*api.SessionHistoryEntry, bool) {
	return selectionpkg.SelectedSessionHistoryEntry(m.selectionSnapshot())
}

func (m Model) selectedHabitHistoryEntry() (*api.HabitCompletion, bool) {
	return selectionpkg.SelectedHabitHistoryEntry(m.selectionSnapshot())
}

func (m Model) selectedRollupDay() (*api.DashboardWindowDay, bool) {
	if m.view != ViewRollup || m.pane != PaneRollupDays || m.dashboardWindow == nil {
		return nil, false
	}
	cur := m.cursor[PaneRollupDays]
	if cur < 0 || cur >= len(m.dashboardWindow.Days) {
		return nil, false
	}
	return &m.dashboardWindow.Days[cur], true
}

func (m Model) openSelectedEditDialog() (Model, bool) {
	snapshot := m.selectionSnapshot()
	switch m.pane {
	case PaneRepos:
		if repoID, repoName, ok := m.selectedMetaRepo(); ok {
			return m.openEditRepoDialog(repoID, repoName), true
		}
	case PaneStreams:
		rawIdx := selectionpkg.FilteredIndexAtCursor(snapshot, PaneStreams)
		if rawIdx >= 0 && rawIdx < len(m.streams) {
			stream := m.streams[rawIdx]
			return m.openEditStreamDialog(stream.ID, stream.RepoID, stream.Name, m.repoNameByID(stream.RepoID)), true
		}
	case PaneIssues:
		if issue, ok := m.selectedIssueRecord(); ok {
			return m.openEditIssueDialog(issue.ID, issue.StreamID, issue.Title, issue.Description, issue.EstimateMinutes, issue.TodoForDate), true
		}
	case PaneHabits:
		if m.view == ViewDaily {
			if habit, ok := m.selectedDailyHabitRecord(); ok {
				return m.openEditHabitDialog(habit.ID, habit.StreamID, habit.Name, habit.Description, string(habit.ScheduleType), habit.Weekdays, habit.TargetMinutes, habit.Active), true
			}
		}
		if habit, ok := m.selectedHabitRecord(); ok {
			return m.openEditHabitDialog(habit.ID, habit.StreamID, habit.Name, habit.Description, string(habit.ScheduleType), habit.Weekdays, habit.TargetMinutes, habit.Active), true
		}
	}
	return m, false
}

func (m Model) openSelectedViewDialog() (Model, bool) {
	snapshot := m.selectionSnapshot()
	switch m.pane {
	case PaneRollupDays:
		if day, ok := m.selectedRollupDay(); ok {
			meta := strings.Join([]string{
				fmt.Sprintf("Status %s", strings.ReplaceAll(string(day.Status), "_", " ")),
				fmt.Sprintf("Planned %d", day.PlannedCount),
				fmt.Sprintf("Done %d", day.CompletedCount),
				fmt.Sprintf("Failed %d", day.FailedCount),
				fmt.Sprintf("Carry %d", day.CarryOverCount),
			}, "   ")
			body := strings.Join([]string{
				"Rollup Day",
				day.Date,
				"",
				"Accountability",
				fmt.Sprintf("%.1f", day.AccountabilityScore),
				"",
				"Summary",
				rollupDayNarrative(*day),
			}, "\n")
			return m.openViewEntityDialog("Rollup Day", day.Date, meta, body), true
		}
		return m, false
	case PaneRepos:
		if m.view != ViewMeta {
			return m, false
		}
		rawIdx := selectionpkg.FilteredIndexAtCursor(snapshot, PaneRepos)
		if rawIdx >= 0 && rawIdx < len(m.repos) {
			repo := m.repos[rawIdx]
			body := strings.Join([]string{
				"Description",
				optionalText(repo.Description),
			}, "\n")
			return m.openViewEntityDialog("Repo", repo.Name, fmt.Sprintf("ID %d", repo.ID), body), true
		}
	case PaneStreams:
		if m.view != ViewMeta {
			return m, false
		}
		rawIdx := selectionpkg.FilteredIndexAtCursor(snapshot, PaneStreams)
		if rawIdx >= 0 && rawIdx < len(m.streams) {
			stream := m.streams[rawIdx]
			meta := strings.Join([]string{
				fmt.Sprintf("Repo %s", m.repoNameByID(stream.RepoID)),
				fmt.Sprintf("Visibility %s", stream.Visibility),
				fmt.Sprintf("ID %d", stream.ID),
			}, "   ")
			body := strings.Join([]string{
				"Description",
				optionalText(stream.Description),
			}, "\n")
			return m.openViewEntityDialog("Stream", stream.Name, meta, body), true
		}
	case PaneIssues:
		issue, ok := m.selectedIssueRecord()
		if !ok {
			return m, false
		}
		repoName, streamName := "-", "-"
		if meta := selectionpkg.IssueMetaByID(snapshot, issue.ID); meta != nil {
			repoName = meta.RepoName
			streamName = meta.StreamName
		}
		estimate := "-"
		if issue.EstimateMinutes != nil {
			estimate = helperpkg.FormatCompactDurationMinutes(*issue.EstimateMinutes)
		}
		due := "-"
		if issue.TodoForDate != nil && strings.TrimSpace(*issue.TodoForDate) != "" {
			due = strings.TrimSpace(*issue.TodoForDate)
		}
		metaBits := []string{
			fmt.Sprintf("Repo %s", repoName),
			fmt.Sprintf("Stream %s", streamName),
			fmt.Sprintf("Status %s", issue.Status),
			fmt.Sprintf("Estimate %s", estimate),
			fmt.Sprintf("Due %s", due),
			fmt.Sprintf("ID %d", issue.ID),
		}
		body := []string{
			"Description",
			optionalText(issue.Description),
		}
		if issue.Notes != nil && strings.TrimSpace(*issue.Notes) != "" {
			body = append(body, "", "Notes", strings.TrimSpace(*issue.Notes))
		}
		return m.openViewEntityDialog("Issue", issue.Title, strings.Join(metaBits, "   "), strings.Join(body, "\n")), true
	case PaneHabits:
		if m.view == ViewMeta {
			if habit, ok := m.selectedHabitRecord(); ok {
				meta := []string{
					fmt.Sprintf("Schedule %s", formatHabitSchedule(habit.ScheduleType, habit.Weekdays)),
					fmt.Sprintf("Target %s", formatHabitTarget(habit.TargetMinutes)),
					fmt.Sprintf("Active %t", habit.Active),
					fmt.Sprintf("ID %d", habit.ID),
				}
				body := strings.Join([]string{"Description", optionalText(habit.Description)}, "\n")
				return m.openViewEntityDialog("Habit", habit.Name, strings.Join(meta, "   "), body), true
			}
		}
		if m.view == ViewDaily {
			if habit, ok := m.selectedDailyHabitRecord(); ok {
				habitStatus := string(habit.Status)
				if strings.TrimSpace(habitStatus) == "" {
					habitStatus = "pending"
				}
				meta := []string{
					fmt.Sprintf("Repo %s", habit.RepoName),
					fmt.Sprintf("Stream %s", habit.StreamName),
					fmt.Sprintf("Schedule %s", formatHabitSchedule(habit.ScheduleType, habit.Weekdays)),
					fmt.Sprintf("Status %s", habitStatus),
					fmt.Sprintf("Duration %s", formatHabitTarget(habit.DurationMinutes)),
				}
				body := []string{"Description", optionalText(habit.Description)}
				if habit.Notes != nil && strings.TrimSpace(*habit.Notes) != "" {
					body = append(body, "", "Notes", strings.TrimSpace(*habit.Notes))
				}
				return m.openViewEntityDialog("Habit", habit.Name, strings.Join(meta, "   "), strings.Join(body, "\n")), true
			}
		}
	}
	return m, false
}

func rollupDayNarrative(day api.DashboardWindowDay) string {
	switch day.Status {
	case "done":
		return "Planned work was completed inside this day."
	case "missed":
		return "Planned work fell through and ended in failure or miss."
	case "carry_over":
		return "Some planned work rolled forward from an earlier day."
	case "mixed":
		return "This day had a mix of completed, missed, or carried work."
	case "planned":
		return "Planned work existed, but it was not completed in this window."
	default:
		return "No tracked plan activity for this day."
	}
}

func optionalText(text *string) string {
	if text == nil || strings.TrimSpace(*text) == "" {
		return "-"
	}
	return strings.TrimSpace(*text)
}

func formatHabitSchedule(scheduleType sharedtypes.HabitScheduleType, weekdays []int) string {
	switch scheduleType {
	case sharedtypes.HabitScheduleWeekdays:
		return "weekdays"
	case sharedtypes.HabitScheduleWeekly:
		return formatHabitWeekdays(weekdays)
	default:
		return "daily"
	}
}

func formatHabitWeekdays(weekdays []int) string {
	if len(weekdays) == 0 {
		return "-"
	}
	names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	out := make([]string, 0, len(weekdays))
	for _, day := range weekdays {
		if day >= 0 && day < len(names) {
			out = append(out, names[day])
		}
	}
	return strings.Join(out, ", ")
}

func formatHabitTarget(target *int) string {
	if target == nil {
		return "-"
	}
	return helperpkg.FormatCompactDurationMinutes(*target)
}

func (m Model) openSelectedDeleteDialog() (Model, bool) {
	if m.timer != nil && m.timer.State != "idle" {
		return m.withStatus("Stop the active session before deleting work items", true), true
	}
	snapshot := m.selectionSnapshot()
	switch m.pane {
	case PaneRepos:
		if repoID, repoName, ok := m.selectedMetaRepo(); ok {
			return m.openConfirmDeleteEntity("repo", fmt.Sprintf("%d", repoID), repoName), true
		}
	case PaneStreams:
		rawIdx := selectionpkg.FilteredIndexAtCursor(snapshot, PaneStreams)
		if rawIdx >= 0 && rawIdx < len(m.streams) {
			stream := m.streams[rawIdx]
			next := m.openConfirmDeleteEntity("stream", fmt.Sprintf("%d", stream.ID), stream.Name)
			next.dialogRepoID = stream.RepoID
			return next, true
		}
	case PaneIssues:
		if issue, ok := m.selectedIssueRecord(); ok {
			next := m.openConfirmDeleteEntity("issue", fmt.Sprintf("%d", issue.ID), issue.Title)
			next.dialogStreamID = issue.StreamID
			return next, true
		}
	case PaneHabits:
		if m.view == ViewDaily {
			if habit, ok := m.selectedDailyHabitRecord(); ok {
				next := m.openConfirmDeleteEntity("habit", fmt.Sprintf("%d", habit.ID), habit.Name)
				next.dialogStreamID = habit.StreamID
				return next, true
			}
		}
		if habit, ok := m.selectedHabitRecord(); ok {
			next := m.openConfirmDeleteEntity("habit", fmt.Sprintf("%d", habit.ID), habit.Name)
			next.dialogStreamID = habit.StreamID
			return next, true
		}
	}
	return m, false
}

func (m Model) repoNameByID(repoID int64) string {
	for _, repo := range m.repos {
		if repo.ID == repoID {
			return repo.Name
		}
	}
	return ""
}

func (m Model) repoDescriptionByID(repoID int64) *string {
	for _, repo := range m.repos {
		if repo.ID == repoID {
			return repo.Description
		}
	}
	return nil
}

func (m Model) streamDescriptionByID(streamID int64) *string {
	for _, stream := range m.streams {
		if stream.ID == streamID {
			return stream.Description
		}
	}
	return nil
}

func (m Model) currentDashboardDate() string {
	if m.dashboardDate != "" {
		return m.dashboardDate
	}
	return time.Now().Format("2006-01-02")
}

func (m Model) currentRollupEndDate() string {
	if m.rollupEndDate != "" {
		return m.rollupEndDate
	}
	return time.Now().Format("2006-01-02")
}

func (m Model) currentRollupStartDate() string {
	if m.rollupStartDate != "" {
		return m.rollupStartDate
	}
	return shiftISODate(m.currentRollupEndDate(), -6)
}

func (m Model) currentWellbeingDate() string {
	if m.wellbeingDate != "" {
		return m.wellbeingDate
	}
	return time.Now().Format("2006-01-02")
}

func (m Model) currentWellbeingWindowDays() int {
	if m.wellbeingWindowDays < 1 {
		return 7
	}
	return m.wellbeingWindowDays
}

func shiftISODate(date string, days int) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}
