package model

import (
	"crona/tui/internal/tui/scope"
	"fmt"
	"slices"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/tui/commands"
	helperpkg "crona/tui/internal/tui/helpers"
	layoutpkg "crona/tui/internal/tui/layout"
	navigationutil "crona/tui/internal/tui/navigationutil"
	selectionpkg "crona/tui/internal/tui/selection"
	uistate "crona/tui/internal/tui/state"
	issuecore "crona/tui/internal/tui/views/issuecore"
	viewruntime "crona/tui/internal/tui/views/runtime"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) availableViews() []View {
	if protected, _, _ := viewruntime.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02")); protected {
		return []View{ViewAway, ViewReports, ViewSessionHistory, ViewHabitHistory, ViewSettings}
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
		return []View{ViewAway, ViewReports, ViewSessionHistory, ViewHabitHistory, ViewSettings}
	}
	if m.timer != nil && m.timer.State != "idle" {
		return []View{ViewSessionActive, ViewSessionHistory}
	}
	return m.availableViews()
}

func (m Model) canJumpToView(target View) bool {
	canJump := slices.Contains(m.jumpAvailableViews(), target)

	return canJump
}

func (m Model) visibleTabViews(tab uistate.Tab) []View {
	available := m.jumpAvailableViews()
	views := []View{}
	for _, view := range uistate.TabViews(tab) {
		if slices.Contains(available, view) {
			views = append(views, view)
		}
	}
	return views
}

func (m Model) visibleTabs() []uistate.Tab {
	tabs := []uistate.Tab{}
	for _, tab := range uistate.TabOrder() {
		if len(m.visibleTabViews(tab)) > 0 {
			tabs = append(tabs, tab)
		}
	}
	return tabs
}

func (m Model) tabTargetView(tab uistate.Tab) (View, bool) {
	views := m.visibleTabViews(tab)
	if len(views) == 0 {
		return "", false
	}
	if slices.Contains(views, m.view) {
		return m.view, true
	}
	return views[0], true
}

func (m Model) nextSubTabView(dir int) View {
	views := m.visibleTabViews(uistate.TabForView(m.view))
	if !slices.Contains(views, m.view) {
		if len(views) > 0 {
			return views[0]
		}
		return m.view
	}
	if len(views) < 2 {
		return m.view
	}
	return navigationutil.NextView(views, m.view, dir)
}

func nextPane(view View, current Pane, dir int) Pane {
	return uistate.NextPane(view, current, dir)
}

// enterIssuesTab starts the scope tree on the checked-out context.
func (m *Model) enterIssuesTab() {
	m.browseScope = scope.FromContext(m.context)
	if m.cursor == nil {
		m.cursor = map[Pane]int{}
	}
	m.cursor[PaneScope] = scope.IndexOf(m.scopeNodes(), m.browseScope)
}

func (m Model) scopeNodes() []scope.Node {
	return scope.Nodes(m.repos, m.allStreams)
}

// browseToScopeCursor scopes the lists to the tree's node under its cursor.
func (m *Model) browseToScopeCursor() {
	nodes := m.scopeNodes()
	if cur := m.cursor[PaneScope]; cur >= 0 && cur < len(nodes) {
		m.browseScope = nodes[cur].Scope
	}
	m.clampIssuesTabLists()
}

// clampIssuesTabLists derives each list from its own view, independent of focus.
func (m *Model) clampIssuesTabLists() {
	snapshot := m.selectionSnapshot()
	snapshot.View = ViewDefault
	m.clamp(PaneIssues, len(selectionpkg.FilteredIndices(snapshot, PaneIssues)))
	snapshot.View = ViewHabitList
	m.clamp(PaneHabits, len(selectionpkg.FilteredIndices(snapshot, PaneHabits)))
	m.cursor[PaneIssues] = max(0, m.cursor[PaneIssues])
	m.cursor[PaneHabits] = max(0, m.cursor[PaneHabits])
}

func (m *Model) initializeIssuesTransition(previous View) {
	if uistate.TabForView(m.view) != uistate.TabIssues {
		return
	}
	entering := uistate.TabForView(previous) != uistate.TabIssues
	if entering {
		m.enterIssuesTab()
	}
	m.clampIssuesTabLists()
}

func (m Model) issuesTransitionRefresh(previous View) tea.Cmd {
	entering := uistate.TabForView(m.view) == uistate.TabIssues && uistate.TabForView(previous) != uistate.TabIssues
	if entering || m.view == ViewHabitList && previous != ViewHabitList {
		return commands.LoadTodayHabits(m.client, m.currentLogicalDate())
	}
	return nil
}

// reconcileBrowseScope keeps the tree's cursor on the browsed node once all streams are
// known, falling back to All when that repo or stream is gone.
func (m *Model) reconcileBrowseScope() {
	nodes := m.scopeNodes()
	idx := scope.IndexOf(nodes, m.browseScope)
	if nodes[idx].Scope != m.browseScope {
		m.browseScope, idx = scope.Scope{}, 0
	}
	if m.cursor == nil {
		m.cursor = map[Pane]int{}
	}
	m.cursor[PaneScope] = idx
}

// selectedScopeNode is the tree's node under its cursor while the tree has focus.
func (m Model) selectedScopeNode() (scope.Node, bool) {
	return selectionpkg.SelectedScopeNode(m.selectionSnapshot())
}

func (m Model) scopeStream(id int64) (api.Stream, bool) {
	for _, stream := range m.allStreams {
		if stream.ID == id {
			return stream, true
		}
	}
	return api.Stream{}, false
}

func (m *Model) setDefaultIssueSection(section DefaultIssueSection) {
	if section != DefaultIssueSectionOpen && section != DefaultIssueSectionCompleted {
		section = DefaultIssueSectionOpen
	}
	m.defaultIssueSection = section
	if m.view != ViewDefault || m.pane != PaneIssues {
		return
	}
	m.cursor[PaneIssues] = 0
	snapshot := m.selectionSnapshot()
	indices := selectionpkg.FilteredIndices(snapshot, PaneIssues)
	m.clamp(PaneIssues, len(indices))
}

func (m *Model) setDailyTaskSection(section DailyTaskSection) {
	if section != DailyTaskSectionPlanned && section != DailyTaskSectionOverdue &&
		section != DailyTaskSectionUpcoming {
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

func (m Model) selectedIssueRecord() (*api.Issue, bool) {
	snapshot := m.selectionSnapshot()
	if m.timer != nil && m.timer.State != "idle" &&
		m.view == ViewSessionActive {
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

func (m Model) selectedMomentumCard() (*api.MomentumCard, bool) {
	return selectionpkg.SelectedMomentumCard(m.selectionSnapshot())
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
	case PaneScope:
		node, ok := m.selectedScopeNode()
		if !ok || node.Scope.IsAll() {
			return m, false
		}
		if node.Scope.StreamID == 0 {
			return m.openEditRepoDialog(node.Scope.RepoID, node.Name), true
		}
		return m.openEditStreamDialog(node.Scope.StreamID, node.Scope.RepoID, node.Name, m.repoNameByID(node.Scope.RepoID)), true
	case PaneIssues:
		if issue, ok := m.selectedIssueRecord(); ok {
			if issue.Adhoc {
				if meta, found := m.issueMetaByID(issue.ID); found {
					return m.withDialogState(m.dialogSnapshot().OpenPromoteAdhocIssue(*meta)), true
				}
			}
			return m.openEditIssueDialog(
				issue.ID,
				issue.StreamID,
				issue.Title,
				issue.Description,
				issue.EstimateMinutes,
				issue.TodoForDate,
			), true
		}
	case PaneHabits:
		if m.view == ViewDaily {
			if habit, ok := m.selectedDailyHabitRecord(); ok {
				return m.openEditHabitDialog(
					habit.ID,
					habit.StreamID,
					habit.Name,
					habit.Description,
					string(habit.ScheduleType),
					habit.Weekdays,
					habit.TargetMinutes,
					habit.Active,
				), true
			}
		}
		if habit, ok := m.selectedHabitRecord(); ok {
			return m.openEditHabitDialog(
				habit.ID,
				habit.StreamID,
				habit.Name,
				habit.Description,
				string(habit.ScheduleType),
				habit.Weekdays,
				habit.TargetMinutes,
				habit.Active,
			), true
		}
	}
	return m, false
}

// viewEntityBodyWidth is the text width inside the view_entity modal (76 wide, padded 3 a side).
func viewEntityBodyWidth(width int) int {
	return max(24, min(width-8, 76)-6)
}

func (m Model) openSelectedViewDialog() (Model, bool) {
	snapshot := m.selectionSnapshot()
	switch m.pane {
	case PaneScope:
		node, ok := m.selectedScopeNode()
		if !ok || node.Scope.IsAll() {
			return m, false
		}
		if node.Scope.StreamID == 0 {
			body := strings.Join([]string{"Description", optionalText(m.repoDescriptionByID(node.Scope.RepoID))}, "\n")
			return m.openViewEntityDialog("Repo", node.Name, fmt.Sprintf("ID %d", node.Scope.RepoID), body), true
		}
		stream, _ := m.scopeStream(node.Scope.StreamID)
		meta := strings.Join([]string{
			fmt.Sprintf("Repo %s", m.repoNameByID(node.Scope.RepoID)),
			fmt.Sprintf("Visibility %s", stream.Visibility),
			fmt.Sprintf("ID %d", node.Scope.StreamID),
		}, "   ")
		body := strings.Join([]string{"Description", optionalText(stream.Description)}, "\n")
		return m.openViewEntityDialog("Stream", node.Name, meta, body), true
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
	case PaneIssues:
		issue, ok := m.selectedIssueRecord()
		if !ok {
			return m, false
		}
		repoName, streamName := "-", "-"
		var issueMeta *api.IssueWithMeta
		if meta := selectionpkg.IssueMetaByID(snapshot, issue.ID); meta != nil {
			issueMeta = meta
			repoName = meta.RepoName
			streamName = meta.StreamName
		}
		spentSeconds := issue.WorkedSeconds
		if issueMeta != nil && issueMeta.WorkedSeconds > spentSeconds {
			spentSeconds = issueMeta.WorkedSeconds
		}
		due := "-"
		if issue.TodoForDate != nil && strings.TrimSpace(*issue.TodoForDate) != "" {
			due = strings.TrimSpace(*issue.TodoForDate)
		}
		metaBits := []string{
			fmt.Sprintf("Repo %s", repoName),
			fmt.Sprintf("Stream %s", streamName),
			fmt.Sprintf("Status %s", issue.Status),
			issuecore.IssueWorkedEstimateSummary(spentSeconds, issue.EstimateMinutes),
			fmt.Sprintf("Due %s", due),
			fmt.Sprintf("ID %d", issue.ID),
		}
		body := []string{
			"Description",
			optionalText(issue.Description),
		}
		body = append(body, issuecore.NotesLines(layoutpkg.ViewTheme(), m.settings, issue.Notes, viewEntityBodyWidth(m.width))...)
		dialog := m.dialogSnapshot().OpenViewEntity(
			"Issue",
			issue.Title,
			strings.Join(metaBits, "   "),
			strings.Join(body, "\n"),
		)
		if issue.Adhoc {
			dialog.ViewEntityAdhoc = true
			dialog.IssueID = issue.ID
		}
		return m.withDialogState(dialog), true
	case PaneHabits:
		if m.view == ViewHabitList {
			if habit, ok := m.selectedHabitRecord(); ok {
				meta := []string{
					fmt.Sprintf(
						"Schedule %s",
						formatHabitSchedule(habit.ScheduleType, habit.Weekdays),
					),
					fmt.Sprintf("Target %s", formatHabitTarget(habit.TargetMinutes)),
					fmt.Sprintf("Active %t", habit.Active),
					fmt.Sprintf("ID %d", habit.ID),
				}
				body := strings.Join([]string{"Description", optionalText(habit.Description)}, "\n")
				return m.openViewEntityDialog(
					"Habit",
					habit.Name,
					strings.Join(meta, "   "),
					body,
				), true
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
					fmt.Sprintf(
						"Schedule %s",
						formatHabitSchedule(habit.ScheduleType, habit.Weekdays),
					),
					fmt.Sprintf("Status %s", habitStatus),
					fmt.Sprintf("Duration %s", formatHabitTarget(habit.DurationMinutes)),
				}
				body := []string{"Description", optionalText(habit.Description)}
				if habit.Notes != nil && strings.TrimSpace(*habit.Notes) != "" {
					body = append(body, "", "Notes", strings.TrimSpace(*habit.Notes))
				}
				return m.openViewEntityDialog(
					"Habit",
					habit.Name,
					strings.Join(meta, "   "),
					strings.Join(body, "\n"),
				), true
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
	switch m.pane {
	case PaneScope:
		node, ok := m.selectedScopeNode()
		if !ok || node.Scope.IsAll() {
			return m, false
		}
		if node.Scope.StreamID == 0 {
			return m.openConfirmDeleteEntity("repo", fmt.Sprintf("%d", node.Scope.RepoID), node.Name), true
		}
		next := m.openConfirmDeleteEntity("stream", fmt.Sprintf("%d", node.Scope.StreamID), node.Name)
		next.dialogRepoID = node.Scope.RepoID
		return next, true
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

// currentDate is the shared date every dated view shows; unset, it follows the logical today.
func (m Model) currentDate() string {
	if m.date != "" {
		return m.date
	}
	return m.currentLogicalDate()
}

func (m Model) currentLogicalDate() string {
	if m.health != nil && strings.TrimSpace(m.health.CurrentDate) != "" {
		return strings.TrimSpace(m.health.CurrentDate)
	}
	return time.Now().Format("2006-01-02")
}

func (m Model) currentRollupWindowDays() int {
	if m.rollupWindowDays < 1 {
		return 7
	}
	return m.rollupWindowDays
}

// reloadDateSurfaces reloads every surface bound to the shared date.
func (m Model) reloadDateSurfaces() tea.Cmd {
	date := m.currentDate()
	return tea.Batch(
		commands.LoadDailySummary(m.client, date),
		commands.LoadDailyPlan(m.client, date),
		commands.LoadDueHabits(m.client, date),
		commands.LoadDailyStreaks(m.client, date),
		commands.LoadSummarySnapshot(m.client, date),
		commands.LoadMomentumRange(m.client, date, m.currentMomentumWindowDays()),
		commands.LoadWellbeingWindow(m.client, date, m.currentWellbeingWindowDays()),
		commands.LoadRollupSummaries(m.client, m.currentRollupStartDate(), date),
	)
}

// currentRollupStartDate opens the Rollup window that ends on the shared date.
func (m Model) currentRollupStartDate() string {
	return shiftISODate(m.currentDate(), -(m.currentRollupWindowDays() - 1))
}

func (m Model) currentMomentumWindowDays() int {
	if m.momentumWindowDays < 1 {
		return 30
	}
	return m.momentumWindowDays
}

func (m Model) currentMomentumTab() MomentumTab {
	return MomentumTabCustom
}

func (m Model) momentumCursorForView() int {
	if m.currentMomentumTab() == MomentumTabFocus {
		return m.momentumHistoryCursor
	}
	return m.cursor[PaneMomentumCards]
}

func (m Model) currentWellbeingWindowDays() int {
	if m.wellbeingWindowDays < 1 {
		return 7
	}
	return m.wellbeingWindowDays
}

func (m Model) currentWeekStart() sharedtypes.WeekStart {
	if m.settings == nil {
		return sharedtypes.WeekStartMonday
	}
	return sharedtypes.NormalizeWeekStart(m.settings.WeekStart)
}

func shiftISODate(date string, days int) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}
