package model

import (
	"maps"
	"time"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	commands "crona/tui/internal/tui/commands"
	configitems "crona/tui/internal/tui/configitems"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
	inputpkg "crona/tui/internal/tui/input"
	selectionpkg "crona/tui/internal/tui/selection"
	uistate "crona/tui/internal/tui/state"
	viewruntime "crona/tui/internal/tui/views/runtime"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) selectionSnapshot() selectionpkg.Snapshot {
	return selectionpkg.PrepareSnapshot(selectionpkg.Snapshot{
		View:                m.view,
		Pane:                m.pane,
		DefaultIssueSection: m.defaultIssueSection,
		DailyTaskSection:    m.dailyTaskSection,
		PreferActiveIssue: m.timer != nil && m.timer.State != "idle" &&
			m.view == ViewSessionActive,
		DashboardDate:  m.currentDashboardDate(),
		Cursors:        m.cursor,
		Filters:        m.filters,
		Context:        m.context,
		Timer:          m.timer,
		Repos:          m.repos,
		Streams:        m.streams,
		Issues:         m.dailyIssuesForSelection(),
		Habits:         m.habits,
		AllIssues:      m.allIssues,
		DueHabits:      m.dueHabits,
		MomentumCards:  m.momentumCards,
		HabitHistory:   m.habitHistory,
		ExportReports:  m.exportReports,
		SessionHistory: m.sessionHistory,
		Ops:            m.ops,
		Settings:       m.settings,
		AlertStatus:    m.alertStatus,
		AlertReminders: m.alertReminders,
		ConfigItems:    m.configItemsForSnapshot(),
	})
}

func (m Model) configItemsForSnapshot() []configitems.Item {
	if m.view != ViewConfig && m.pane != PaneConfig {
		return nil
	}
	return configitems.Build(m.exportAssets)
}

func (m Model) dailyIssuesForSelection() []api.Issue {
	if m.dailySummary == nil {
		return m.issues
	}
	if m.view == ViewDaily {
		return m.dailySummary.Issues
	}
	return m.issues
}

func (m Model) inputState() inputpkg.State {
	protected := false
	activeView := m.view
	activePane := m.pane
	cursor := maps.Clone(m.cursor)
	if cursor == nil {
		cursor = map[Pane]int{}
	}
	if nextProtected, _, _ := viewruntime.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02")); nextProtected {
		protected = true
		if activeView != ViewReports && activeView != ViewSessionHistory &&
			activeView != ViewHabitHistory && activeView != ViewSettings {
			activeView = ViewAway
			activePane = uistate.DefaultPane(activeView)
		}
	}
	return inputpkg.State{
		ActiveView:                activeView,
		ActivePane:                activePane,
		ProtectedModeActive:       protected,
		Cursor:                    cursor,
		Filters:                   m.filters,
		DefaultIssueSection:       m.defaultIssueSection,
		DailyTaskSection:          m.dailyTaskSection,
		DashboardDate:             m.dashboardDate,
		RollupStartDate:           m.rollupStartDate,
		RollupEndDate:             m.rollupEndDate,
		MomentumDate:              m.momentumDate,
		MomentumWindowDays:        m.currentMomentumWindowDays(),
		MomentumTab:               string(m.currentMomentumTab()),
		MomentumHistoryY:          m.momentumHistoryCursor,
		WellbeingDate:             m.wellbeingDate,
		WellbeingWindowDays:       m.currentWellbeingWindowDays(),
		Dialog:                    m.dialog,
		DialogState:               m.dialogState(),
		HelpOpen:                  m.helpOpen,
		SessionDetailOpen:         m.sessionDetailOpen,
		SessionDetailY:            m.sessionDetailY,
		SessionContextOpen:        m.sessionContextOpen,
		SessionContextY:           m.sessionContextY,
		OpsLimit:                  m.opsLimit,
		OpsLimitPinned:            m.opsLimitPinned,
		Context:                   m.context,
		Timer:                     m.timer,
		UpdateStatus:              m.updateStatus,
		UpdateChecking:            m.updateChecking,
		UpdateDiagnosticsExpanded: m.updateDiagnosticsExpanded,
		CurrentExecutable:         m.currentExecutablePath,
		RunningIsBeta:             m.isBetaBuild(),
		Settings:                  m.settings,
		AlertStatus:               m.alertStatus,
		AlertReminders:            m.alertReminders,
		ExportAssets:              m.exportAssets,
		DailyCheckIn:              m.dailyCheckIn,
	}
}

func (m Model) applyInputState(state inputpkg.State) Model {
	m.view = state.ActiveView
	m.pane = state.ActivePane
	m.cursor = maps.Clone(state.Cursor)
	if MomentumTab(state.MomentumTab) == MomentumTabFocus {
		m.momentumHistoryCursor = state.Cursor[PaneMomentumCards]
	} else if state.Cursor != nil {
		m.cursor[PaneMomentumCards] = state.Cursor[PaneMomentumCards]
	}
	m.filters = state.Filters
	m.defaultIssueSection = state.DefaultIssueSection
	m.dailyTaskSection = state.DailyTaskSection
	m.dashboardDate = state.DashboardDate
	m.rollupStartDate = state.RollupStartDate
	m.rollupEndDate = state.RollupEndDate
	m.momentumDate = state.MomentumDate
	m.momentumWindowDays = state.MomentumWindowDays
	m.wellbeingDate = state.WellbeingDate
	m.wellbeingWindowDays = state.WellbeingWindowDays
	m = m.withDialogState(state.DialogState)
	if state.DialogState.Kind == "" {
		m.dialog = state.Dialog
	}
	m.helpOpen = state.HelpOpen
	m.sessionDetailOpen = state.SessionDetailOpen
	m.sessionDetailY = state.SessionDetailY
	m.sessionContextOpen = state.SessionContextOpen
	m.sessionContextY = state.SessionContextY
	m.opsLimit = state.OpsLimit
	m.opsLimitPinned = state.OpsLimitPinned
	m.context = state.Context
	m.timer = state.Timer
	m.updateStatus = state.UpdateStatus
	m.updateChecking = state.UpdateChecking
	m.updateDiagnosticsExpanded = state.UpdateDiagnosticsExpanded
	m.currentExecutablePath = state.CurrentExecutable
	m.settings = state.Settings
	m.alertStatus = state.AlertStatus
	m.alertReminders = state.AlertReminders
	m.exportAssets = state.ExportAssets
	m.dailyCheckIn = state.DailyCheckIn
	return m
}

func (m Model) inputDeps() inputpkg.Deps {
	return inputpkg.Deps{
		CloseEventStop: func() { m.stopEventStream() },
		ShutdownKernel: func() tea.Cmd { return commands.ShutdownKernel(m.client) },
		SeedDevData:    func() tea.Cmd { return commands.SeedDevData(m.client) },
		ClearDevData:   func() tea.Cmd { return commands.ClearDevData(m.client) },
		IsDevMode:      func(state inputpkg.State) bool { return m.applyInputState(state).isDevMode() },
		NextActiveSessionView: func(state inputpkg.State, dir int) uistate.View {
			return m.applyInputState(state).nextActiveSessionView(dir)
		},
		NextWorkspaceView: func(state inputpkg.State, dir int) uistate.View {
			return m.applyInputState(state).nextWorkspaceView(dir)
		},
		DefaultPane: uistate.DefaultPane,
		NextPane:    nextPane,
		SetDefaultIssueSection: func(state *inputpkg.State, section uistate.DefaultIssueSection) {
			next := m.applyInputState(*state)
			next.setDefaultIssueSection(section)
			*state = next.inputState()
		},
		SetDailyTaskSection: func(state *inputpkg.State, section uistate.DailyTaskSection) {
			next := m.applyInputState(*state)
			next.setDailyTaskSection(section)
			*state = next.inputState()
		},
		ListLen: func(state inputpkg.State, pane uistate.Pane) int {
			next := m.applyInputState(state)
			return (&next).listLen(pane)
		},
		LoadExportAssets: func() tea.Cmd { return commands.LoadExportAssets(m.client) },
		SetStatus: func(state *inputpkg.State, message string, isErr bool) tea.Cmd {
			next := m.applyInputState(*state)
			cmd := next.setStatus(message, isErr)
			*state = next.inputState()
			return cmd
		},
		LoadDailySummary:     func(date string) tea.Cmd { return commands.LoadDailySummary(m.client, date) },
		LoadDueHabits:        func(date string) tea.Cmd { return commands.LoadDueHabits(m.client, date) },
		LoadDailyStreaks:     func(date string) tea.Cmd { return commands.LoadDailyStreaks(m.client, date) },
		CurrentDashboardDate: func(state inputpkg.State) string { return m.applyInputState(state).currentDashboardDate() },
		LoadRollupSummaries:  func(start, end string) tea.Cmd { return commands.LoadRollupSummaries(m.client, start, end) },
		LoadMomentumRange: func(date string, windowDays int) tea.Cmd {
			return commands.LoadMomentumRange(m.client, date, windowDays)
		},
		CurrentMomentumDate: func(state inputpkg.State) string {
			return m.applyInputState(state).currentMomentumDate()
		},
		CurrentRollupStartDate: func(state inputpkg.State) string {
			return m.applyInputState(state).currentRollupStartDate()
		},
		CurrentRollupEndDate: func(state inputpkg.State) string {
			return m.applyInputState(state).currentRollupEndDate()
		},
		LoadWellbeing: func(date string, windowDays int) tea.Cmd {
			return commands.LoadWellbeingWindow(m.client, date, windowDays)
		},
		CurrentWellbeingDate: func(state inputpkg.State) string { return m.applyInputState(state).currentWellbeingDate() },
		OpenCheckInDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			date := next.currentDashboardDate()
			if next.view == ViewWellbeing {
				date = next.currentWellbeingDate()
			}
			next = next.openCheckInDialogForDate(date)
			*state = next.inputState()
			return true
		},
		OpenHelpDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openHelpDialog()
			*state = next.inputState()
			return true
		},
		ConfigChangeSelected: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			if item, ok := selectionpkg.SelectedConfigItem(snapshot); ok &&
				next.exportAssets != nil {
				switch {
				case item.PresetStyle:
					for _, asset := range next.exportAssets.TemplateAssets {
						if asset.ReportKind != item.ReportKind ||
							asset.AssetKind != item.AssetKind ||
							len(asset.Presets) == 0 {
							continue
						}
						currentID := ""
						if asset.SelectedPreset != nil {
							currentID = asset.SelectedPreset.ID
						}
						nextID := currentID
						for idx, preset := range asset.Presets {
							if preset.ID == currentID {
								nextID = asset.Presets[(idx+1)%len(asset.Presets)].ID
								break
							}
						}
						if nextID == "" && len(asset.Presets) > 0 {
							nextID = asset.Presets[0].ID
						}
						*state = next.inputState()
						return commands.ApplyExportTemplatePreset(
							m.client,
							item.ReportKind,
							item.AssetKind,
							nextID,
						)
					}
				case item.Label == "Reports directory":
					next = next.openExportReportsDirDialog(next.exportAssets.ReportsDir)
				case item.Label == "ICS export directory":
					next = next.openExportICSDirDialog(next.exportAssets.ICSDir)
				default:
					return nil
				}
				*state = next.inputState()
				return nil
			}
			return nil
		},
		OpenCheckoutContextDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openCheckoutContextDialog()
			*state = next.inputState()
			return true
		},
		Checkout: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			model, cmd := next.checkout()
			*state = model.inputState()
			return cmd
		},
		CheckUpdateNow: func() tea.Cmd { return commands.CheckUpdateNow(m.client) },
		ResumeSession:  func(state inputpkg.State) tea.Cmd { return commands.ResumeFocusSession(m.client, state.Timer) },
		PauseSession:   func() tea.Cmd { return commands.PauseFocusSession(m.client) },
		OpenEndSessionDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openSessionMessageDialog("end_session")
			*state = next.inputState()
			return true
		},
		OpenStashSessionDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openSessionMessageDialog("stash_session")
			*state = next.inputState()
			return true
		},
		CanOpenStashList: func(state inputpkg.State) bool {
			next := m.applyInputState(state)
			return (next.timer == nil || next.timer.State == "idle") && next.dialog == ""
		},
		OpenStashListDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openStashListDialog()
			*state = next.inputState()
			return true
		},
		LoadStashes: func() tea.Cmd { return commands.LoadStashes(m.client) },
		ClampFiltered: func(state *inputpkg.State, pane uistate.Pane) {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			max := len(selectionpkg.FilteredIndices(snapshot, pane))
			if max == 0 {
				next.cursor[pane] = 0
			} else if next.cursor[pane] >= max {
				next.cursor[pane] = max - 1
			}
			*state = next.inputState()
		},
		CurrentOpsLimit: func(state inputpkg.State) int {
			next := m.applyInputState(state)
			return (&next).currentOpsLimit()
		},
		LoadOps: func(limit int) tea.Cmd { return commands.LoadOps(m.client, limit) },
		OpenIssueStatusFromSelection: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			if issue, ok := selectionpkg.SelectedIssueDetail(snapshot); ok {
				next = next.withDialogState(next.dialogSnapshot().OpenIssueStatus(issue.Status))
				*state = next.inputState()
				return true
			}
			return false
		},
		AbandonSelectedIssue: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			issue, ok := selectionpkg.SelectedIssueDetail(snapshot)
			if !ok {
				return nil
			}
			if issue.Status == "done" {
				return next.setStatus("Done issues cannot be abandoned", true)
			}
			if issue.Status == "abandoned" {
				return nil
			}
			if next.timer != nil && next.timer.State != "idle" {
				next = next.withDialogState(
					next.dialogSnapshot().OpenIssueSessionTransition(issue.ID, "abandoned"),
				)
				*state = next.inputState()
				return nil
			}
			next = next.withDialogState(
				next.dialogSnapshot().OpenIssueStatusNote("abandoned", "Abandon reason", true),
			)
			next.dialogIssueID = issue.ID
			next.dialogStreamID = issue.StreamID
			*state = next.inputState()
			return nil
		},
		ToggleSelectedIssueToday: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			if issue, ok := selectionpkg.SelectedIssueDetail(snapshot); ok {
				date := ""
				if issue.TodoForDate != nil && *issue.TodoForDate == next.currentDashboardDate() {
					date = ""
				} else {
					date = next.currentDashboardDate()
				}
				return commands.SetIssueTodoDate(
					m.client,
					issue.ID,
					date,
					issue.StreamID,
					next.currentDashboardDate(),
				)
			}
			return nil
		},
		ToggleSelectedIssuePinnedDaily: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			if issue, ok := selectionpkg.SelectedIssue(snapshot); ok {
				return commands.ToggleIssuePinnedDaily(
					m.client,
					issue.ID,
					issue.PinnedDaily,
					issue.StreamID,
					next.currentDashboardDate(),
				)
			}
			return nil
		},
		OpenSelectedIssueTodoDateDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			if issue, ok := selectionpkg.SelectedIssueDetail(snapshot); ok {
				next = next.withDialogState(
					next.dialogSnapshot().OpenDatePicker("", issue.ID, 0, issue.TodoForDate),
				)
			}
			*state = next.inputState()
			return true
		},
		HandleCreateAction: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.handleInputCreateAction()
			*state = next.inputState()
			return true
		},
		OpenExportDailyDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openExportDailyDialog()
			*state = next.inputState()
			return true
		},
		OpenEditorAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			next, cmd, handled := next.handleInputOpenEditor()
			*state = next.inputState()
			return cmd, handled
		},
		OpenHabitLogAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			habit, ok := next.selectedHabitForAction()
			if !ok {
				return nil, false
			}
			next = next.openHabitCompletionDialog(
				habit.ID,
				next.currentDashboardDate(),
				habit.TargetMinutes,
				nil,
			)
			*state = next.inputState()
			return nil, true
		},
		EnterHabitHistoryView: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			next, cmd := next.enterHabitHistoryViewFromSelection()
			*state = next.inputState()
			return cmd, true
		},
		DeleteSelectionAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			next, cmd, handled := next.handleInputDeleteSelection()
			*state = next.inputState()
			return cmd, handled
		},
		OpenSelectionAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			cmd, handled := next.handleInputOpenSelection()
			*state = next.inputState()
			return cmd, handled
		},
		EnterAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			next, cmd, handled := next.handleInputEnter()
			*state = next.inputState()
			return cmd, handled
		},
		ToggleMomentumAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			cmd, handled := next.handleInputToggleMomentum()
			*state = next.inputState()
			return cmd, handled
		},
		ToggleHabitCompletedAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			cmd, handled := next.handleInputToggleHabitCompleted()
			*state = next.inputState()
			return cmd, handled
		},
		SetHabitFailedAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			cmd, handled := next.handleInputSetHabitFailed()
			*state = next.inputState()
			return cmd, handled
		},
		StartFocusFromSelection: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			model, cmd := next.handleInputStartFocusFromSelection()
			*state = model.(Model).inputState()
			return cmd
		},
		OpenManualSessionDialog: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			if next.dialog != "" {
				return nil, false
			}
			if next.view == ViewDaily && next.pane == PaneHabits {
				if habit, ok := next.selectedDailyHabitRecord(); ok {
					next = next.openHabitCompletionDialog(
						habit.ID,
						next.currentDashboardDate(),
						habit.DurationMinutes,
						habit.Notes,
					)
					*state = next.inputState()
					return nil, true
				}
				return nil, false
			}
			cmd := next.preflightIssueActionFromSelection(commands.IssueActionModeManual)
			if cmd == nil {
				return nil, false
			}
			*state = next.inputState()
			return cmd, true
		},
		OpenSessionContextOverlay: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			if next.view != ViewSessionActive || next.timer == nil || next.timer.State == "idle" {
				return false
			}
			snapshot := next.selectionSnapshot()
			if selectionpkg.ActiveIssue(snapshot) == nil {
				return false
			}
			next.sessionContextOpen = true
			next.sessionContextY = 0
			*state = next.inputState()
			return true
		},
		ConfigReset: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			cmd, _ := next.handleInputConfigReset()
			*state = next.inputState()
			return cmd
		},
		PatchSetting: func(key sharedtypes.CoreSettingsKey, value any, repoID, streamID int64, dashboardDate string) tea.Cmd {
			return commands.PatchSetting(m.client, key, value, repoID, streamID, dashboardDate)
		},
		CreateMomentumDefinition: func(def api.HabitStreakDefinition, dashboardDate string, momentumDate string, momentumWindowDays int) tea.Cmd {
			return commands.CreateMomentumDefinition(
				m.client,
				def,
				dashboardDate,
				momentumDate,
				momentumWindowDays,
			)
		},
		UpdateMomentumDefinition: func(def api.HabitStreakDefinition, dashboardDate string, momentumDate string, momentumWindowDays int) tea.Cmd {
			return commands.UpdateMomentumDefinition(
				m.client,
				def,
				dashboardDate,
				momentumDate,
				momentumWindowDays,
			)
		},
		DeleteMomentumDefinition: func(id string, dashboardDate string, momentumDate string, momentumWindowDays int) tea.Cmd {
			return commands.DeleteMomentumDefinition(
				m.client,
				id,
				dashboardDate,
				momentumDate,
				momentumWindowDays,
			)
		},
		TestAlertNotification: func() tea.Cmd { return commands.TestAlertNotification(m.client) },
		TestAlertSound:        func() tea.Cmd { return commands.TestAlertSound(m.client) },
		CreateAlertReminder: func(input shareddto.AlertReminderCreateRequest) tea.Cmd {
			return commands.CreateAlertReminder(m.client, input)
		},
		UpdateAlertReminder: func(input shareddto.AlertReminderUpdateRequest) tea.Cmd {
			return commands.UpdateAlertReminder(m.client, input)
		},
		ToggleAlertReminder: func(id string, enabled bool) tea.Cmd {
			return commands.ToggleAlertReminder(m.client, id, enabled)
		},
		DeleteAlertReminder: func(id string) tea.Cmd {
			return commands.DeleteAlertReminder(m.client, id)
		},
		OpenCreateAlertReminderDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openCreateAlertReminderDialog()
			*state = next.inputState()
			return true
		},
		OpenEditAlertReminderDialog: func(state *inputpkg.State, id string) bool {
			next := m.applyInputState(*state)
			next = next.openEditAlertReminderDialog(id)
			*state = next.inputState()
			return true
		},
		OpenEditDateDisplayFormatDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openEditDateDisplayFormatDialog()
			*state = next.inputState()
			return true
		},
		OpenEditRestProtectionDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openEditRestProtectionDialog()
			*state = next.inputState()
			return true
		},
		OpenEditTelemetrySettingsDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openEditTelemetrySettingsDialog()
			*state = next.inputState()
			return true
		},
		OpenConfirmWipeDataDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openConfirmWipeDataDialog()
			*state = next.inputState()
			return true
		},
		WipeRuntimeData:           func() tea.Cmd { return commands.WipeRuntimeData(m.client) },
		OpenSupportIssueURL:       func() tea.Cmd { return m.openSupportIssueURL() },
		OpenSupportDiscussionsURL: func() tea.Cmd { return m.openSupportDiscussionsURL() },
		OpenSupportReleasesURL:    func() tea.Cmd { return m.openSupportReleasesURL() },
		OpenSupportRoadmapURL:     func() tea.Cmd { return m.openSupportRoadmapURL() },
		CopySupportDiagnostics:    func(state inputpkg.State) tea.Cmd { return m.copySupportDiagnosticsCmd(state) },
		GenerateSupportBundle:     func(state inputpkg.State) tea.Cmd { return m.generateSupportBundleCmd(state) },
		OpenViewJumpDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openViewJumpDialog()
			*state = next.inputState()
			return true
		},
		OpenBetaSupportDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openBetaSupportDialog()
			*state = next.inputState()
			return true
		},
		OpenRollupStartDateDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openDatePickerDialog(
				"rollup_start",
				0,
				0,
				dialogstate.ValueToPointer(next.currentRollupStartDate()),
			)
			*state = next.inputState()
			return true
		},
		OpenRollupEndDateDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openDatePickerDialog(
				"rollup_end",
				0,
				0,
				dialogstate.ValueToPointer(next.currentRollupEndDate()),
			)
			*state = next.inputState()
			return true
		},
	}
}
