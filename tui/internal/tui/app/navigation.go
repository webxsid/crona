package app

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	commands "crona/tui/internal/tui/commands"
	dialogruntime "crona/tui/internal/tui/dialog_runtime"
	helperpkg "crona/tui/internal/tui/helpers"
	navigationutil "crona/tui/internal/tui/navigationutil"
	selectionpkg "crona/tui/internal/tui/selection"
	uistate "crona/tui/internal/tui/state"
	"crona/tui/internal/tui/views"

	tea "github.com/charmbracelet/bubbletea"
)

func nextView(current View, dir int) View {
	return navigationutil.NextView(uistate.ViewOrder(), current, dir)
}

func (m Model) availableViews() []View {
	if protected, _, _ := views.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02")); protected {
		return []View{ViewAway, ViewReports, ViewSessionHistory}
	}
	ordered := uistate.ViewOrder()
	views := make([]View, 0, len(ordered))
	for _, view := range ordered {
		views = append(views, view)
	}
	if len(views) == 0 {
		return []View{ViewDaily}
	}
	return views
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
	openIndices, completedIndices := views.SplitDefaultIssueIndices(m.allIssues, m.filters[PaneIssues], m.settings)
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
	indices := selectionpkg.FilteredIndices(m.selectionSnapshot(), PaneIssues)
	m.clamp(PaneIssues, len(indices))
}

func (m *Model) cycleDefaultIssueSection(dir int) {
	if dir >= 0 {
		if m.defaultIssueSection == DefaultIssueSectionOpen {
			m.setDefaultIssueSection(DefaultIssueSectionCompleted)
			return
		}
		m.setDefaultIssueSection(DefaultIssueSectionOpen)
		return
	}
	if m.defaultIssueSection == DefaultIssueSectionCompleted {
		m.setDefaultIssueSection(DefaultIssueSectionOpen)
		return
	}
	m.setDefaultIssueSection(DefaultIssueSectionCompleted)
}

func (m Model) selectedMetaRepo() (int64, string, bool) {
	return selectionpkg.SelectedMetaRepo(m.selectionSnapshot())
}

func (m Model) selectedMetaStream() (int64, string, string, bool) {
	return selectionpkg.SelectedMetaStream(m.selectionSnapshot())
}

func (m Model) selectedIssueRecord() (*api.Issue, bool) {
	if m.timer != nil && m.timer.State != "idle" && (m.view == ViewSessionActive || m.view == ViewScratch) {
		if issue := selectionpkg.ActiveIssue(m.selectionSnapshot()); issue != nil {
			copy := issue.Issue
			return &copy, true
		}
	}
	if m.pane != PaneIssues {
		return nil, false
	}
	return selectionpkg.SelectedIssue(m.selectionSnapshot())
}

func (m Model) selectedHabitRecord() (*api.Habit, bool) {
	return selectionpkg.SelectedHabit(m.selectionSnapshot())
}

func (m Model) selectedDailyHabitRecord() (*api.HabitDailyItem, bool) {
	return selectionpkg.SelectedDailyHabit(m.selectionSnapshot())
}

func (m Model) selectedSessionHistoryEntry() (*api.SessionHistoryEntry, bool) {
	return selectionpkg.SelectedSessionHistoryEntry(m.selectionSnapshot())
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
	switch m.pane {
	case PaneRepos:
		if repoID, repoName, ok := m.selectedMetaRepo(); ok {
			return m.openEditRepoDialog(repoID, repoName), true
		}
	case PaneStreams:
		rawIdx := selectionpkg.FilteredIndexAtCursor(m.selectionSnapshot(), PaneStreams)
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
		rawIdx := selectionpkg.FilteredIndexAtCursor(m.selectionSnapshot(), PaneRepos)
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
		rawIdx := selectionpkg.FilteredIndexAtCursor(m.selectionSnapshot(), PaneStreams)
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
		if meta := selectionpkg.IssueMetaByID(m.selectionSnapshot(), issue.ID); meta != nil {
			repoName = meta.RepoName
			streamName = meta.StreamName
		}
		estimate := "-"
		if issue.EstimateMinutes != nil {
			estimate = fmt.Sprintf("%dm", *issue.EstimateMinutes)
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
	return fmt.Sprintf("%dm", *target)
}

func (m Model) openSelectedDeleteDialog() (Model, bool) {
	if m.timer != nil && m.timer.State != "idle" {
		return m.withStatus("Stop the active session before deleting work items", true), true
	}
	switch m.pane {
	case PaneRepos:
		if repoID, repoName, ok := m.selectedMetaRepo(); ok {
			return m.openConfirmDeleteEntity("repo", fmt.Sprintf("%d", repoID), repoName), true
		}
	case PaneStreams:
		rawIdx := selectionpkg.FilteredIndexAtCursor(m.selectionSnapshot(), PaneStreams)
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

func shiftISODate(date string, days int) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}

func (m Model) checkout() (Model, tea.Cmd) {
	switch m.pane {
	case PaneRepos:
		rawIdx := selectionpkg.FilteredIndexAtCursor(m.selectionSnapshot(), PaneRepos)
		if rawIdx >= 0 && rawIdx < len(m.repos) {
			repo := m.repos[rawIdx]
			return m, commands.CheckoutRepo(m.client, repo.ID)
		}
	case PaneStreams:
		rawIdx := selectionpkg.FilteredIndexAtCursor(m.selectionSnapshot(), PaneStreams)
		if rawIdx >= 0 && rawIdx < len(m.streams) {
			stream := m.streams[rawIdx]
			return m, commands.CheckoutStream(m.client, stream.ID)
		}
	}
	return m, nil
}

func (m Model) handleInputCreateAction() Model {
	if m.view == ViewDefault && m.pane == PaneIssues {
		return m.openCreateIssueDefaultDialog()
	}
	if m.view == ViewDaily {
		switch m.pane {
		case PaneIssues:
			return m.openCreateIssueDefaultDialog()
		case PaneHabits:
			repoName := "-"
			if m.context != nil && m.context.RepoName != nil && *m.context.RepoName != "" {
				repoName = *m.context.RepoName
			}
			streamName := "-"
			if m.context != nil && m.context.StreamName != nil && *m.context.StreamName != "" {
				streamName = *m.context.StreamName
			}
			return m.openCreateHabitDialog(0, streamName, repoName)
		}
	}
	if m.view == ViewMeta {
		switch m.pane {
		case PaneRepos:
			return m.openCreateRepoDialog()
		case PaneStreams:
			repoID, repoName, ok := m.selectedMetaRepo()
			if !ok {
				m.setStatus("Select or checkout a repo first", true)
				return m
			}
			return m.openCreateStreamDialog(repoID, repoName)
		case PaneIssues:
			streamID, streamName, repoName, ok := m.selectedMetaStream()
			if !ok {
				m.setStatus("Select or checkout a stream first", true)
				return m
			}
			return m.openCreateIssueMetaDialog(streamID, streamName, repoName)
		case PaneHabits:
			streamID, streamName, repoName, ok := m.selectedMetaStream()
			if ok {
				return m.openCreateHabitDialog(streamID, streamName, repoName)
			}
			repoName = "-"
			if m.context != nil && m.context.RepoName != nil && *m.context.RepoName != "" {
				repoName = *m.context.RepoName
			}
			streamName = "-"
			if m.context != nil && m.context.StreamName != nil && *m.context.StreamName != "" {
				streamName = *m.context.StreamName
			}
			return m.openCreateHabitDialog(0, streamName, repoName)
		}
	}
	if m.view == ViewWellbeing {
		return m.openCreateCheckInDialog()
	}
	if m.pane == PaneScratchpads {
		return m.openCreateScratchpad()
	}
	return m
}

func (m Model) handleInputOpenEditor() (Model, tea.Cmd, bool) {
	if m.view == ViewConfig && m.pane == PaneConfig {
		if item, ok := selectionpkg.SelectedConfigItem(m.selectionSnapshot()); ok && item.Editable && strings.TrimSpace(item.Path) != "" {
			return m, dialogruntime.OpenEditor(item.Path, func(err error) tea.Msg { return commands.ErrMsg{Err: err} }), true
		}
		return m, nil, true
	}
	if m.view == ViewReports && m.pane == PaneExportReports {
		if report, ok := selectionpkg.SelectedExportReport(m.selectionSnapshot()); ok && strings.TrimSpace(report.Path) != "" {
			return m, dialogruntime.OpenEditor(report.Path, func(err error) tea.Msg { return commands.ErrMsg{Err: err} }), true
		}
		return m, nil, true
	}
	if m.view == ViewDaily && m.pane == PaneHabits {
		if next, ok := m.openSelectedEditDialog(); ok {
			return next, nil, true
		}
		return m, nil, true
	}
	if m.view == ViewWellbeing {
		if m.dailyCheckIn == nil {
			m = m.openCreateCheckInDialog()
		} else {
			m = m.openEditCheckInDialog()
		}
		return m, nil, true
	}
	if m.view == ViewSessionHistory && m.pane == PaneSessions {
		if entry, ok := m.selectedSessionHistoryEntry(); ok {
			m.sessionDetailOpen = true
			m.sessionDetailY = 0
			return m, commands.LoadSessionDetail(m.client, entry.ID), true
		}
		return m, nil, true
	}
	if m.dialog == "" {
		if next, ok := m.openSelectedEditDialog(); ok {
			return next, nil, true
		}
	}
	return m, nil, false
}

func (m Model) handleInputSetHabitFailed() (tea.Cmd, bool) {
	if m.view != ViewDaily || m.pane != PaneHabits {
		return nil, false
	}
	if habit, ok := m.selectedDailyHabitRecord(); ok {
		if habit.Status == sharedtypes.HabitCompletionStatusFailed {
			return commands.UncompleteHabit(m.client, habit.ID, m.currentDashboardDate()), true
		}
		return commands.SetHabitStatus(m.client, habit.ID, m.currentDashboardDate(), sharedtypes.HabitCompletionStatusFailed, habit.DurationMinutes, habit.Notes), true
	}
	return nil, true
}

func (m Model) handleInputDeleteSelection() (Model, tea.Cmd, bool) {
	if m.view == ViewWellbeing {
		if m.dailyCheckIn == nil {
			return m, m.setStatus("No check-in to delete for this date", true), true
		}
		m = m.openConfirmDeleteEntity("checkin", m.currentWellbeingDate(), "this check-in")
		return m, nil, true
	}
	if m.view == ViewReports && m.pane == PaneExportReports {
		if report, ok := selectionpkg.SelectedExportReport(m.selectionSnapshot()); ok {
			m = m.openConfirmDeleteEntity("report", report.Path, report.Name)
			return m, nil, true
		}
		return m, nil, true
	}
	if m.pane == PaneScratchpads {
		rawIdx := selectionpkg.FilteredIndexAtCursor(m.selectionSnapshot(), PaneScratchpads)
		if rawIdx >= 0 && rawIdx < len(m.scratchpads) {
			m = m.openConfirmDelete(m.scratchpads[rawIdx].ID)
			return m, nil, true
		}
	}
	if m.dialog == "" {
		if next, ok := m.openSelectedDeleteDialog(); ok {
			return next, nil, true
		}
	}
	return m, nil, false
}

func (m Model) handleInputOpenSelection() (tea.Cmd, bool) {
	if m.view == ViewUpdates && m.updateStatus != nil {
		return commands.OpenExternalURL(m.updateStatus.ReleaseURL), true
	}
	if m.view == ViewReports && m.pane == PaneExportReports {
		if report, ok := selectionpkg.SelectedExportReport(m.selectionSnapshot()); ok && strings.TrimSpace(report.Path) != "" {
			return dialogruntime.OpenDefaultViewer(report.Path, func(err error) tea.Msg { return commands.ErrMsg{Err: err} }), true
		}
		return nil, true
	}
	return nil, false
}

func (m Model) handleInputEnter() (Model, tea.Cmd, bool) {
	if m.view == ViewConfig && m.pane == PaneConfig {
		if item, ok := selectionpkg.SelectedConfigItem(m.selectionSnapshot()); ok {
			m = m.openViewEntityDialog(item.DetailTitle, item.Label, item.DetailMeta, item.DetailBody)
			return m, nil, true
		}
		return m, nil, true
	}
	if m.view == ViewReports && m.pane == PaneExportReports {
		if report, ok := selectionpkg.SelectedExportReport(m.selectionSnapshot()); ok {
			meta := fmt.Sprintf("Kind %s   Format %s   Modified %s", report.Kind, report.Format, report.ModifiedAt)
			scope := strings.TrimSpace(report.ScopeLabel)
			if scope == "" {
				scope = "-"
			}
			dateLabel := strings.TrimSpace(report.DateLabel)
			if dateLabel == "" {
				dateLabel = report.Date
			}
			body := "Scope\n" + scope + "\n\nDate\n" + dateLabel + "\n\nPath\n" + report.Path + "\n\nSize\n" + fmt.Sprintf("%d bytes", report.SizeBytes)
			m = m.openViewEntityDialog("Export Report", report.Name, meta, body)
			return m, nil, true
		}
		return m, nil, true
	}
	if m.view == ViewSessionHistory && m.pane == PaneSessions {
		if entry, ok := m.selectedSessionHistoryEntry(); ok {
			m.sessionDetailOpen = true
			m.sessionDetailY = 0
			return m, commands.LoadSessionDetail(m.client, entry.ID), true
		}
		return m, nil, true
	}
	if m.dialog == "" {
		if next, ok := m.openSelectedViewDialog(); ok {
			return next, nil, true
		}
	}
	if m.pane == PaneScratchpads {
		rawIdx := selectionpkg.FilteredIndexAtCursor(m.selectionSnapshot(), PaneScratchpads)
		if rawIdx < 0 || rawIdx >= len(m.scratchpads) {
			return m, nil, true
		}
		m.scratchpadOpen = true
		m.scratchpadMeta = helperpkg.ScratchpadMetaAt(m.scratchpads, rawIdx)
		return m, commands.OpenScratchpad(m.client, m.scratchpads, rawIdx), true
	}
	return m, nil, false
}

func (m Model) handleInputToggleHabitCompleted() (tea.Cmd, bool) {
	if m.view == ViewDaily && m.pane == PaneHabits {
		if habit, ok := m.selectedDailyHabitRecord(); ok {
			if habit.Status == sharedtypes.HabitCompletionStatusCompleted {
				return commands.UncompleteHabit(m.client, habit.ID, m.currentDashboardDate()), true
			}
			return commands.SetHabitStatus(m.client, habit.ID, m.currentDashboardDate(), sharedtypes.HabitCompletionStatusCompleted, nil, nil), true
		}
	}
	return nil, false
}

func (m Model) handleInputConfigReset() (tea.Cmd, bool) {
	if m.view != ViewConfig || m.exportAssets == nil {
		return nil, false
	}
	if item, ok := selectionpkg.SelectedConfigItem(m.selectionSnapshot()); ok {
		if item.Label == "Reports directory" && m.exportAssets.ReportsDirCustomized {
			return commands.SetExportReportsDir(m.client, ""), true
		}
		if item.Label == "ICS export directory" && m.exportAssets.ICSDirCustomized {
			return commands.SetExportICSDir(m.client, ""), true
		}
		if item.Resettable {
			return commands.ResetExportTemplate(m.client, item.ReportKind, item.AssetKind), true
		}
	}
	return nil, false
}

func (m Model) handleInputStartFocusFromSelection() (tea.Model, tea.Cmd) {
	if m.view == ViewSessionHistory && (m.timer == nil || m.timer.State == "idle") {
		rawIdx := selectionpkg.FilteredIndexAtCursor(m.selectionSnapshot(), PaneSessions)
		if rawIdx < 0 || rawIdx >= len(m.sessionHistory) {
			return m, nil
		}
		issueID := m.sessionHistory[rawIdx].IssueID
		meta := selectionpkg.IssueMetaByID(m.selectionSnapshot(), issueID)
		if meta == nil {
			return m, m.setStatus("Issue metadata unavailable", true)
		}
		return m, commands.StartFocusSession(m.client, meta.RepoID, meta.StreamID, issueID)
	}
	if m.view == ViewSessionActive {
		if m.timer == nil || m.timer.State == "idle" {
			if m.context == nil || m.context.RepoID == nil || m.context.StreamID == nil || m.context.IssueID == nil {
				return m, m.setStatus("No active issue in context", true)
			}
			meta := selectionpkg.ActiveIssue(m.selectionSnapshot())
			if meta == nil {
				return m, m.setStatus("Active issue metadata unavailable", true)
			}
			return m, commands.StartFocusSession(m.client, *m.context.RepoID, *m.context.StreamID, *m.context.IssueID)
		}
		return m, nil
	}
	if m.pane != PaneIssues {
		return m, nil
	}
	if m.view == ViewDefault {
		rawIdx := selectionpkg.FilteredIndexAtCursor(m.selectionSnapshot(), PaneIssues)
		issues := selectionpkg.DefaultScopedIssues(m.selectionSnapshot())
		if rawIdx < 0 || rawIdx >= len(issues) {
			return m, nil
		}
		issue := issues[rawIdx]
		return m, commands.StartFocusSession(m.client, issue.RepoID, issue.StreamID, issue.ID)
	}
	if m.view == ViewDaily {
		rawIdx := selectionpkg.FilteredIndexAtCursor(m.selectionSnapshot(), PaneIssues)
		issues := selectionpkg.DailyScopedIssues(m.selectionSnapshot())
		if rawIdx < 0 || rawIdx >= len(issues) {
			return m, nil
		}
		issue := issues[rawIdx]
		meta := selectionpkg.IssueMetaByID(m.selectionSnapshot(), issue.ID)
		if meta == nil {
			return m, m.setStatus("Issue metadata unavailable", true)
		}
		return m, commands.StartFocusSession(m.client, meta.RepoID, issue.StreamID, issue.ID)
	}
	rawIdx := selectionpkg.FilteredIndexAtCursor(m.selectionSnapshot(), PaneIssues)
	if rawIdx < 0 || rawIdx >= len(m.issues) {
		return m, nil
	}
	if m.context == nil || m.context.RepoID == nil {
		return m, m.setStatus("No repo in context for selected issue", true)
	}
	issue := m.issues[rawIdx]
	return m, commands.StartFocusSession(m.client, *m.context.RepoID, issue.StreamID, issue.ID)
}
