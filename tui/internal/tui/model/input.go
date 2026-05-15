package model

import (
	"time"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	commands "crona/tui/internal/tui/commands"
	configitems "crona/tui/internal/tui/configitems"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
	filteringpkg "crona/tui/internal/tui/filtering"
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
		PreferActiveIssue:   m.timer != nil && m.timer.State != "idle" && (m.view == ViewSessionActive || m.view == ViewScratch),
		DashboardDate:       m.currentDashboardDate(),
		Cursors:             m.cursor,
		Filters:             m.filters,
		Context:             m.context,
		Timer:               m.timer,
		Repos:               m.repos,
		Streams:             m.streams,
		Issues:              m.dailyIssuesForSelection(),
		Habits:              m.habits,
		AllIssues:           m.allIssues,
		DueHabits:           m.dueHabits,
		HabitHistory:        m.habitHistory,
		ExportReports:       m.exportReports,
		Scratchpads:         m.scratchpads,
		SessionHistory:      m.sessionHistory,
		Ops:                 m.ops,
		Settings:            m.settings,
		AlertStatus:         m.alertStatus,
		AlertReminders:      m.alertReminders,
		ConfigItems:         m.configItemsForSnapshot(),
	})
}

func (m Model) configItemsForSnapshot() []configitems.Item {
	if m.view != ViewConfig && m.pane != PaneConfig && (!m.filterEditing || m.filterPane != PaneConfig) {
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
	if nextProtected, _, _ := viewruntime.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02")); nextProtected {
		protected = true
		if activeView != ViewReports && activeView != ViewSessionHistory && activeView != ViewHabitHistory {
			activeView = ViewAway
			activePane = uistate.DefaultPane(activeView)
		}
	}
	return inputpkg.State{
		ActiveView:          activeView,
		ActivePane:          activePane,
		ProtectedModeActive: protected,
		Cursor:              m.cursor,
		Filters:             m.filters,
		DefaultIssueSection: m.defaultIssueSection,
		DailyTaskSection:    m.dailyTaskSection,
		DashboardDate:       m.dashboardDate,
		RollupStartDate:     m.rollupStartDate,
		RollupEndDate:       m.rollupEndDate,
		WellbeingDate:       m.wellbeingDate,
		Dialog:              m.dialog,
		DialogState:         m.dialogState(),
		HelpOpen:            m.helpOpen,
		SessionDetailOpen:   m.sessionDetailOpen,
		SessionDetailY:      m.sessionDetailY,
		SessionContextOpen:  m.sessionContextOpen,
		SessionContextY:     m.sessionContextY,
		ScratchpadOpen:      m.scratchpadOpen,
		OpsLimit:            m.opsLimit,
		OpsLimitPinned:      m.opsLimitPinned,
		Context:             m.context,
		Timer:               m.timer,
		UpdateStatus:        m.updateStatus,
		UpdateChecking:      m.updateChecking,
		UpdateInstalling:    m.updateInstalling,
		UpdateInstallError:  m.updateInstallError,
		CurrentExecutable:   m.currentExecutablePath,
		RunningIsBeta:       m.isBetaBuild(),
		Settings:            m.settings,
		AlertStatus:         m.alertStatus,
		AlertReminders:      m.alertReminders,
		ExportAssets:        m.exportAssets,
		DailyCheckIn:        m.dailyCheckIn,
	}
}

func (m Model) applyInputState(state inputpkg.State) Model {
	m.view = state.ActiveView
	m.pane = state.ActivePane
	m.cursor = state.Cursor
	m.filters = state.Filters
	m.defaultIssueSection = state.DefaultIssueSection
	m.dailyTaskSection = state.DailyTaskSection
	m.dashboardDate = state.DashboardDate
	m.rollupStartDate = state.RollupStartDate
	m.rollupEndDate = state.RollupEndDate
	m.wellbeingDate = state.WellbeingDate
	m = m.withDialogState(state.DialogState)
	if state.DialogState.Kind == "" {
		m.dialog = state.Dialog
	}
	m.helpOpen = state.HelpOpen
	m.sessionDetailOpen = state.SessionDetailOpen
	m.sessionDetailY = state.SessionDetailY
	m.sessionContextOpen = state.SessionContextOpen
	m.sessionContextY = state.SessionContextY
	m.scratchpadOpen = state.ScratchpadOpen
	m.opsLimit = state.OpsLimit
	m.opsLimitPinned = state.OpsLimitPinned
	m.context = state.Context
	m.timer = state.Timer
	m.updateStatus = state.UpdateStatus
	m.updateChecking = state.UpdateChecking
	m.updateInstalling = state.UpdateInstalling
	m.updateInstallError = state.UpdateInstallError
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
		CloseEventStop:     func() { m.stopEventStream() },
		ShutdownKernel:     func() tea.Cmd { return commands.ShutdownKernel(m.client) },
		SeedDevData:        func() tea.Cmd { return commands.SeedDevData(m.client) },
		ClearDevData:       func() tea.Cmd { return commands.ClearDevData(m.client) },
		PrepareLocalUpdate: func() tea.Cmd { return commands.PrepareLocalUpdate(m.client) },
		IsDevMode:          func(state inputpkg.State) bool { return m.applyInputState(state).isDevMode() },
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
		CurrentDashboardDate: func(state inputpkg.State) string { return m.applyInputState(state).currentDashboardDate() },
		LoadRollupSummaries:  func(start, end string) tea.Cmd { return commands.LoadRollupSummaries(m.client, start, end) },
		CurrentRollupStartDate: func(state inputpkg.State) string {
			return m.applyInputState(state).currentRollupStartDate()
		},
		CurrentRollupEndDate: func(state inputpkg.State) string {
			return m.applyInputState(state).currentRollupEndDate()
		},
		LoadWellbeing:        func(date string) tea.Cmd { return commands.LoadWellbeing(m.client, date) },
		CurrentWellbeingDate: func(state inputpkg.State) string { return m.applyInputState(state).currentWellbeingDate() },
		ConfigChangeSelected: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			if item, ok := selectionpkg.SelectedConfigItem(snapshot); ok && next.exportAssets != nil {
				switch {
				case item.PresetStyle:
					for _, asset := range next.exportAssets.TemplateAssets {
						if asset.ReportKind != item.ReportKind || asset.AssetKind != item.AssetKind || len(asset.Presets) == 0 {
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
						return commands.ApplyExportTemplatePreset(m.client, item.ReportKind, item.AssetKind, nextID)
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
		CheckUpdateNow:             func() tea.Cmd { return commands.CheckUpdateNow(m.client) },
		SelfUpdateInstallAvailable: func(state inputpkg.State) bool { return m.applyInputState(state).selfUpdateInstallAvailable() },
		SelfUpdateUnsupportedReason: func(state inputpkg.State) string {
			return m.applyInputState(state).selfUpdateUnsupportedReason()
		},
		InstallUpdate: func(state inputpkg.State) tea.Cmd {
			next := m.applyInputState(state)
			return commands.InstallUpdate(next.updateStatus, next.selfUpdateInstallAvailable(), next.selfUpdateUnsupportedReason())
		},
		DismissUpdate: func() tea.Cmd { return commands.DismissUpdate(m.client) },
		ResumeSession: func() tea.Cmd { return commands.ResumeFocusSession(m.client) },
		PauseSession:  func() tea.Cmd { return commands.PauseFocusSession(m.client) },
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
			filterState := next.filterState()
			if deps := next.filterDeps(); deps.Clamp != nil && deps.ItemCount != nil {
				deps.Clamp(filterState.Cursor, pane, deps.ItemCount(filterState, pane))
			}
			next = next.applyFilterState(filterState)
			*state = next.inputState()
		},
		CurrentOpsLimit: func(state inputpkg.State) int {
			next := m.applyInputState(state)
			return (&next).currentOpsLimit()
		},
		LoadOps: func(limit int) tea.Cmd { return commands.LoadOps(m.client, limit) },
		StartFilterEdit: func(state *inputpkg.State, pane uistate.Pane) {
			next := m.applyInputState(*state)
			next = next.applyFilterState(filteringpkg.Start(next.filterState(), pane))
			*state = next.inputState()
		},
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
				next = next.withDialogState(next.dialogSnapshot().OpenIssueSessionTransition(issue.ID, "abandoned"))
				*state = next.inputState()
				return nil
			}
			next = next.withDialogState(next.dialogSnapshot().OpenIssueStatusNote("abandoned", "Abandon reason", true))
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
				return commands.SetIssueTodoDate(m.client, issue.ID, date, issue.StreamID, next.currentDashboardDate())
			}
			return nil
		},
		ToggleSelectedIssuePinnedDaily: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			if issue, ok := selectionpkg.SelectedIssue(snapshot); ok {
				return commands.ToggleIssuePinnedDaily(m.client, issue.ID, issue.PinnedDaily, issue.StreamID, next.currentDashboardDate())
			}
			return nil
		},
		OpenSelectedIssueTodoDateDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			if issue, ok := selectionpkg.SelectedIssueDetail(snapshot); ok {
				next = next.withDialogState(next.dialogSnapshot().OpenDatePicker("", issue.ID, 0, issue.TodoForDate))
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
			next = next.openHabitCompletionDialog(habit.ID, next.currentDashboardDate(), habit.TargetMinutes, nil)
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
		OpenManualSessionDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			if next.dialog != "" {
				return false
			}
			if next.view == ViewDaily && next.pane == PaneHabits {
				if habit, ok := next.selectedDailyHabitRecord(); ok {
					next = next.openHabitCompletionDialog(habit.ID, next.currentDashboardDate(), habit.DurationMinutes, habit.Notes)
					*state = next.inputState()
					return true
				}
				return false
			}
			snapshot := next.selectionSnapshot()
			if issue, ok := selectionpkg.SelectedIssueDetail(snapshot); ok {
				issueLabel := ""
				var estimateMinutes *int
				if meta := selectionpkg.IssueMetaByID(snapshot, issue.ID); meta != nil {
					issueLabel = meta.Title
					estimateMinutes = meta.EstimateMinutes
				}
				next = next.openManualSessionDialog(issue.ID, issueLabel, estimateMinutes, next.currentDashboardDate())
				*state = next.inputState()
				return true
			}
			return false
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
		OpenEditHabitStreaksDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openEditHabitStreaksDialog()
			*state = next.inputState()
			return true
		},
		OpenEditRestProtectionDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openEditRestProtectionDialog()
			*state = next.inputState()
			return true
		},
		OpenConfirmWipeDataDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openConfirmWipeDataDialog()
			*state = next.inputState()
			return true
		},
		OpenConfirmUninstallDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openConfirmUninstallDialog()
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
			next = next.openDatePickerDialog("rollup_start", 0, 0, dialogstate.ValueToPointer(next.currentRollupStartDate()))
			*state = next.inputState()
			return true
		},
		OpenRollupEndDateDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openDatePickerDialog("rollup_end", 0, 0, dialogstate.ValueToPointer(next.currentRollupEndDate()))
			*state = next.inputState()
			return true
		},
	}
}
