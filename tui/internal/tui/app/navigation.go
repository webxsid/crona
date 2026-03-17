package app

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/logger"
	"crona/tui/internal/tui/app/views"

	tea "github.com/charmbracelet/bubbletea"
)

func nextView(current View, dir int) View {
	for i, v := range viewOrder {
		if v == current {
			return viewOrder[(i+dir+len(viewOrder))%len(viewOrder)]
		}
	}
	return current
}

func nextPane(view View, current Pane, dir int) Pane {
	panes := viewPanes[view]
	if len(panes) == 0 {
		return current
	}
	for i, p := range panes {
		if p == current {
			return panes[(i+dir+len(panes))%len(panes)]
		}
	}
	return panes[0]
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
	m.clampFiltered(PaneIssues)
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

func (m Model) selectedIssue() (int64, int64, string, *string, bool) {
	if m.timer != nil && m.timer.State != "idle" && (m.view == ViewSessionActive || m.view == ViewScratch) {
		if issue := m.activeIssueWithMeta(); issue != nil {
			return issue.ID, issue.StreamID, string(issue.Status), issue.TodoForDate, true
		}
	}
	if m.pane != PaneIssues {
		return 0, 0, "", nil, false
	}
	switch m.view {
	case ViewDefault:
		scoped := m.defaultScopedIssues()
		rawIdx := m.filteredIndexAtCursor(PaneIssues)
		if rawIdx < 0 || rawIdx >= len(scoped) {
			return 0, 0, "", nil, false
		}
		issue := scoped[rawIdx]
		return issue.ID, issue.StreamID, string(issue.Status), issue.TodoForDate, true
	case ViewDaily:
		rawIdx := m.filteredIndexAtCursor(PaneIssues)
		issues := m.dailyScopedIssues()
		if rawIdx < 0 || rawIdx >= len(issues) {
			return 0, 0, "", nil, false
		}
		issue := issues[rawIdx]
		return issue.ID, issue.StreamID, string(issue.Status), issue.TodoForDate, true
	case ViewMeta:
		rawIdx := m.filteredIndexAtCursor(PaneIssues)
		if rawIdx < 0 || rawIdx >= len(m.issues) {
			return 0, 0, "", nil, false
		}
		issue := m.issues[rawIdx]
		return issue.ID, issue.StreamID, string(issue.Status), issue.TodoForDate, true
	default:
		return 0, 0, "", nil, false
	}
}

func (m Model) selectedMetaRepo() (int64, string, bool) {
	if m.view != ViewMeta {
		return 0, "", false
	}
	rawIdx := m.filteredIndexAtCursor(PaneRepos)
	if rawIdx >= 0 && rawIdx < len(m.repos) {
		return m.repos[rawIdx].ID, m.repos[rawIdx].Name, true
	}
	if m.context != nil && m.context.RepoID != nil {
		return *m.context.RepoID, firstNonEmpty(m.context.RepoName, nil), true
	}
	return 0, "", false
}

func (m Model) selectedMetaStream() (int64, string, string, bool) {
	if m.view != ViewMeta {
		return 0, "", "", false
	}
	rawIdx := m.filteredIndexAtCursor(PaneStreams)
	if rawIdx >= 0 && rawIdx < len(m.streams) {
		stream := m.streams[rawIdx]
		return stream.ID, stream.Name, m.repoNameByID(stream.RepoID), true
	}
	if m.context != nil && m.context.StreamID != nil {
		repoName := "-"
		if m.context.RepoName != nil {
			repoName = *m.context.RepoName
		}
		return *m.context.StreamID, firstNonEmpty(m.context.StreamName, nil), repoName, true
	}
	return 0, "", "", false
}

func (m Model) selectedIssueRecord() (*api.Issue, bool) {
	if m.timer != nil && m.timer.State != "idle" && (m.view == ViewSessionActive || m.view == ViewScratch) {
		if issue := m.activeIssueWithMeta(); issue != nil {
			copy := issue.Issue
			return &copy, true
		}
	}
	if m.pane != PaneIssues {
		return nil, false
	}
	switch m.view {
	case ViewDefault:
		scoped := m.defaultScopedIssues()
		rawIdx := m.filteredIndexAtCursor(PaneIssues)
		if rawIdx < 0 || rawIdx >= len(scoped) {
			return nil, false
		}
		copy := scoped[rawIdx].Issue
		return &copy, true
	case ViewDaily:
		rawIdx := m.filteredIndexAtCursor(PaneIssues)
		issues := m.dailyScopedIssues()
		if rawIdx < 0 || rawIdx >= len(issues) {
			return nil, false
		}
		copy := issues[rawIdx]
		return &copy, true
	case ViewMeta:
		rawIdx := m.filteredIndexAtCursor(PaneIssues)
		if rawIdx < 0 || rawIdx >= len(m.issues) {
			return nil, false
		}
		copy := m.issues[rawIdx]
		return &copy, true
	default:
		return nil, false
	}
}

func (m Model) selectedHabitRecord() (*api.Habit, bool) {
	if m.pane != PaneHabits {
		return nil, false
	}
	switch m.view {
	case ViewMeta:
		rawIdx := m.filteredIndexAtCursor(PaneHabits)
		if rawIdx < 0 || rawIdx >= len(m.habits) {
			return nil, false
		}
		copy := m.habits[rawIdx]
		return &copy, true
	default:
		return nil, false
	}
}

func (m Model) selectedDailyHabitRecord() (*api.HabitDailyItem, bool) {
	if m.view != ViewDaily || m.pane != PaneHabits {
		return nil, false
	}
	rawIdx := m.filteredIndexAtCursor(PaneHabits)
	habits := m.filteredDueHabits()
	if rawIdx < 0 || rawIdx >= len(habits) {
		return nil, false
	}
	copy := habits[rawIdx]
	return &copy, true
}

func (m Model) selectedSessionHistoryEntry() (*api.SessionHistoryEntry, bool) {
	if m.view != ViewSessionHistory || m.pane != PaneSessions {
		return nil, false
	}
	rawIdx := m.filteredIndexAtCursor(PaneSessions)
	if rawIdx < 0 || rawIdx >= len(m.sessionHistory) {
		return nil, false
	}
	return &m.sessionHistory[rawIdx], true
}

func (m Model) openSelectedEditDialog() (Model, bool) {
	switch m.pane {
	case PaneRepos:
		if repoID, repoName, ok := m.selectedMetaRepo(); ok {
			return m.openEditRepoDialog(repoID, repoName), true
		}
	case PaneStreams:
		rawIdx := m.filteredIndexAtCursor(PaneStreams)
		if rawIdx >= 0 && rawIdx < len(m.streams) {
			stream := m.streams[rawIdx]
			return m.openEditStreamDialog(stream.ID, stream.RepoID, stream.Name, m.repoNameByID(stream.RepoID)), true
		}
	case PaneIssues:
		if issue, ok := m.selectedIssueRecord(); ok {
			return m.openEditIssueDialog(issue.ID, issue.StreamID, issue.Title, issue.Description, issue.EstimateMinutes, issue.TodoForDate), true
		}
	case PaneHabits:
		if habit, ok := m.selectedHabitRecord(); ok {
			return m.openEditHabitDialog(habit.ID, habit.StreamID, habit.Name, habit.Description, string(habit.ScheduleType), habit.Weekdays, habit.TargetMinutes, habit.Active), true
		}
	}
	return m, false
}

func (m Model) openSelectedViewDialog() (Model, bool) {
	switch m.pane {
	case PaneRepos:
		if m.view != ViewMeta {
			return m, false
		}
		rawIdx := m.filteredIndexAtCursor(PaneRepos)
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
		rawIdx := m.filteredIndexAtCursor(PaneStreams)
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
		meta := m.issueMetaByID(issue.ID)
		repoName := "-"
		streamName := "-"
		if meta != nil {
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
		rawIdx := m.filteredIndexAtCursor(PaneStreams)
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

func optionalText(value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return "-"
	}
	return strings.TrimSpace(*value)
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

func (m Model) abandonSelectedIssue() (tea.Model, tea.Cmd) {
	issueID, streamID, status, _, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}
	if status == "done" {
		return m, m.setStatus("Done issues cannot be abandoned", true)
	}
	if status == "abandoned" {
		return m, nil
	}
	if m.timer != nil && m.timer.State != "idle" {
		return m.openIssueSessionTransitionDialog(issueID, "abandoned"), nil
	}
	next := m.openIssueStatusNoteDialog("abandoned", "Abandon reason", true)
	next.dialogIssueID = issueID
	next.dialogStreamID = streamID
	return next, nil
}

func (m Model) toggleSelectedIssueToday() (tea.Model, tea.Cmd) {
	issueID, streamID, _, todoForDate, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}
	return m, cmdToggleIssueToday(m.client, issueID, todoForDate != nil && *todoForDate != "", streamID, m.currentDashboardDate())
}

func (m Model) openSelectedIssueTodoDateDialog() Model {
	issueID, _, _, todoForDate, ok := m.selectedIssue()
	if !ok {
		return m
	}
	return m.openDatePickerDialog("", issueID, 0, todoForDate)
}

func (m Model) currentDashboardDate() string {
	if m.dashboardDate != "" {
		return m.dashboardDate
	}
	return time.Now().Format("2006-01-02")
}

func (m Model) currentWellbeingDate() string {
	if m.wellbeingDate != "" {
		return m.wellbeingDate
	}
	return time.Now().Format("2006-01-02")
}

func shiftISODate(date string, days int) string {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Now().AddDate(0, 0, days).Format("2006-01-02")
	}
	return parsed.AddDate(0, 0, days).Format("2006-01-02")
}

func (m Model) checkout() (tea.Model, tea.Cmd) {
	switch m.pane {
	case PaneRepos:
		rawIdx := m.filteredIndexAtCursor(PaneRepos)
		if rawIdx < 0 || rawIdx >= len(m.repos) {
			return m, nil
		}
		repo := m.repos[rawIdx]
		logger.Infof("checkout repo: %s (%d)", repo.Name, repo.ID)
		return m, cmdCheckoutRepo(m.client, repo.ID)
	case PaneStreams:
		rawIdx := m.filteredIndexAtCursor(PaneStreams)
		if rawIdx < 0 || rawIdx >= len(m.streams) {
			return m, nil
		}
		stream := m.streams[rawIdx]
		logger.Infof("checkout stream: %s (%d)", stream.Name, stream.ID)
		return m, cmdCheckoutStream(m.client, stream.ID)
	default:
		return m, nil
	}
}
