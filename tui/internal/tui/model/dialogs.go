package model

import (
	"strings"
	"time"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	commands "crona/tui/internal/tui/commands"
	dialogruntime "crona/tui/internal/tui/dialog_runtime"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
	helperpkg "crona/tui/internal/tui/helpers"
	selectionpkg "crona/tui/internal/tui/selection"
	uistate "crona/tui/internal/tui/state"
	viewruntime "crona/tui/internal/tui/views/runtime"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) dialogSnapshot() dialogstate.Snapshot {
	selectionSnapshot := m.selectionSnapshot()
	dialogState := m.dialogState()
	if m.settings != nil {
		dialogState.PromptGlyphMode = sharedtypes.NormalizePromptGlyphMode(m.settings.PromptGlyphMode)
	}
	dialogSnapshot := dialogstate.Snapshot{
		Dialog:               dialogState,
		Repos:                m.repos,
		Streams:              m.streams,
		AllHabits:            m.allHabits,
		AllIssues:            m.allIssues,
		Context:              m.context,
		Stashes:              m.stashes,
		DailyCheckIn:         m.dailyCheckIn,
		UpdateStatus:         m.updateStatus,
		ExportAssets:         m.exportAssets,
		Settings:             m.settings,
		AlertReminders:       m.alertReminders,
		CurrentDashboardDate: m.currentDashboardDate(),
		CurrentWellbeingDate: m.currentWellbeingDate(),
		HasActiveTimer:       m.timer != nil && m.timer.State != "idle",
		AvailableViews:       m.jumpAvailableViews(),
	}
	dialogSnapshot.ProtectedModeActive, _, _ = viewruntime.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02"))
	if issue, ok := selectionpkg.SelectedIssueDetail(selectionSnapshot); ok {
		dialogSnapshot.SelectedIssueID = issue.ID
		dialogSnapshot.SelectedStreamID = issue.StreamID
		dialogSnapshot.HasSelectedIssue = true
	}
	if issue := selectionpkg.ActiveIssue(selectionSnapshot); issue != nil {
		dialogSnapshot.ActiveIssueStream = issue.StreamID
		dialogSnapshot.HasActiveIssue = true
	}
	return dialogSnapshot
}

func (m Model) dialogActionCmd(action dialogstate.Action) tea.Cmd {
	return dialogruntime.Resolve(action, m.dialogRuntimeState(), m.dialogRuntimeDeps())
}

func (m Model) handleDialogAction(next Model, action dialogstate.Action) (Model, tea.Cmd) {
	switch action.Kind {
	case "jump_view":
		target := View(strings.TrimSpace(action.TargetView))
		if target == "" {
			return next, nil
		}
		if !next.canJumpToView(target) {
			return next, next.setStatus("That view is not available right now", true)
		}
		if target == ViewSessionActive && (next.timer == nil || next.timer.State == "idle") {
			return next, next.setStatus("Active session view is only available while a timer is running", true)
		}
		if target == ViewAway {
			protected, _, _ := viewruntime.ProtectedRestMode(next.settings, time.Now().Format("2006-01-02"))
			if !protected {
				return next, next.setStatus("Away view is only available when away mode or rest protection is active", true)
			}
		}
		if target == ViewHabitHistory {
			return next.enterHabitHistoryViewFromSelection()
		}
		next.view = target
		next.pane = uistate.DefaultPane(target)
		return next, nil
	case "continue_focus_fresh":
		return next, commands.ContinueFocusSessionFresh(next.client, action.RepoID, action.StreamID, action.IssueID)
	case "copy_support_diagnostics":
		return next, next.copySupportDiagnosticsCmd(next.inputState())
	case "generate_support_bundle":
		return next, next.generateSupportBundleCmd(next.inputState())
	case "open_view_entity_editor":
		path := strings.TrimSpace(action.Path)
		if path == "" {
			return next, nil
		}
		return next, dialogruntime.OpenEditor(path, func(err error) tea.Msg { return commands.ErrMsg{Err: err} })
	case "open_export_reports_dir_dialog":
		if next.exportAssets == nil {
			return next, nil
		}
		return next.openExportReportsDirDialog(next.exportAssets.ReportsDir), nil
	case "open_export_ics_dir_dialog":
		if next.exportAssets == nil {
			return next, nil
		}
		return next.openExportICSDirDialog(next.exportAssets.ICSDir), nil
	case "reset_export_reports_dir":
		if next.exportAssets == nil {
			return next, nil
		}
		return next, commands.SetExportReportsDir(next.client, "")
	case "reset_export_ics_dir":
		if next.exportAssets == nil {
			return next, nil
		}
		return next, commands.SetExportICSDir(next.client, "")
	default:
		return next, next.dialogActionCmd(action)
	}
}

func (m Model) openCreateScratchpad() Model {
	return m.withDialogState(m.dialogSnapshot().OpenCreateScratchpad())
}

func (m Model) openViewJumpDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenViewJump())
}

func (m Model) openBetaSupportDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenBetaSupport())
}

func (m Model) openStashConflictDialog(conflict api.StashConflict, repoID, streamID, issueID int64) Model {
	state := m.dialogSnapshot().OpenStashConflict(conflict)
	state.RepoID = repoID
	state.StreamID = streamID
	state.IssueID = issueID
	return m.withDialogState(state)
}

func (m Model) openCreateRepoDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenCreateRepo())
}

func (m Model) openEditRepoDialog(repoID int64, name string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenEditRepo(repoID, name, m.repoDescriptionByID(repoID)))
}

func (m Model) openCreateStreamDialog(repoID int64, repoName string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenCreateStream(repoID, repoName))
}

func (m Model) openEditStreamDialog(streamID, repoID int64, streamName, repoName string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenEditStream(streamID, repoID, streamName, repoName, m.streamDescriptionByID(streamID)))
}

func (m Model) openCreateIssueMetaDialog(streamID int64, streamName, repoName string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenCreateIssueMeta(streamID, streamName, repoName))
}

func (m Model) openCreateHabitDialog(streamID int64, streamName, repoName string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenCreateHabit(streamID, streamName, repoName))
}

func (m Model) openEditIssueDialog(issueID, streamID int64, title string, description *string, estimateMinutes *int, todoForDate *string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenEditIssue(issueID, streamID, title, description, estimateMinutes, todoForDate))
}

func (m Model) openEditHabitDialog(habitID, streamID int64, name string, description *string, schedule string, weekdays []int, targetMinutes *int, active bool) Model {
	return m.withDialogState(m.dialogSnapshot().OpenEditHabit(habitID, streamID, name, description, schedule, weekdays, targetMinutes, active))
}

func (m Model) openHabitCompletionDialog(habitID int64, date string, durationMinutes *int, notes *string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenHabitCompletion(habitID, date, durationMinutes, notes))
}

func (m Model) openHabitHistoryView() Model {
	m.view = ViewHabitHistory
	m.pane = PaneHabitHistory
	m.habitHistoryHabitID = 0
	m.habitHistoryTitle = "Habit History"
	m.habitHistoryMeta = habitHistoryScopeLabel(m.context)
	m.cursor[PaneHabitHistory] = 0
	return m
}

func (m Model) enterHabitHistoryViewFromSelection() (Model, tea.Cmd) {
	next := m
	next = next.openHabitHistoryView()
	return next, commands.LoadHabitHistory(m.client, m.context, nil)
}

func (m Model) openCreateIssueDefaultDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenCreateIssueDefault())
}

func (m Model) openCheckoutContextDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenCheckoutContext())
}

func (m Model) openCreateCheckInDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenCreateCheckIn())
}

func (m Model) openEditCheckInDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenEditCheckIn())
}

func (m Model) openConfirmDelete(id string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenConfirmDelete(id))
}

func (m Model) openConfirmDeleteEntity(kind, id, label string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenConfirmDeleteEntity(kind, id, label))
}

func (m Model) openStashListDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenStashList())
}

func (m Model) openSessionMessageDialog(kind string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenSessionMessage(kind))
}

func (m Model) openAmendSessionDialog(sessionID string, commit string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenAmendSession(sessionID, commit))
}

func (m Model) openManualSessionDialog(issueID int64, issueLabel string, estimateMinutes *int, date string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenManualSession(issueID, issueLabel, estimateMinutes, date))
}

func (m Model) openDatePickerDialog(parentDialog string, issueID int64, inputIndex int, initial *string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenDatePicker(parentDialog, issueID, inputIndex, initial))
}

func (m Model) openViewEntityDialog(title string, name string, meta string, body string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenViewEntity(title, name, meta, body))
}

func (m Model) openViewEntityDialogWithPath(title string, name string, meta string, body string, path string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenViewEntityWithPath(title, name, meta, body, path))
}

func (m Model) openSupportBundleDialog(path string, sizeBytes int64, windowLabel string) Model {
	meta := strings.Join([]string{
		"Size " + helperpkg.HumanizeSupportBytes(sizeBytes),
		"Window " + strings.TrimSpace(windowLabel),
		"Redaction safe",
	}, "   ")
	body := strings.Join([]string{
		"Bundle",
		path,
		"",
		"Attach this zip to a bug report if you need deeper diagnostics.",
		"Use o to open the folder, c to copy the path, or g to open the issue tracker.",
	}, "\n")
	return m.withDialogState(m.dialogSnapshot().OpenSupportBundleResult(helperpkg.SupportBundleDisplayName(path), meta, body, path))
}

func (m Model) openExportDailyDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenExportDaily())
}

func (m Model) openExportReportsDirDialog(current string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenExportReportsDir(current))
}

func (m Model) openExportICSDirDialog(current string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenExportICSDir(current))
}

func (m Model) openCreateAlertReminderDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenCreateAlertReminder())
}

func (m Model) openEditAlertReminderDialog(id string) Model {
	return m.withDialogState(m.dialogSnapshot().OpenEditAlertReminder(id))
}

func (m Model) openEditDateDisplayFormatDialog() Model {
	current := ""
	if m.settings != nil {
		current = m.settings.DateDisplayFormat
	}
	return m.withDialogState(m.dialogSnapshot().OpenEditDateDisplayFormat(current))
}

func (m Model) openEditRestProtectionDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenEditRestProtection())
}

func (m Model) openEditHabitStreaksDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenEditHabitStreaks())
}

func (m Model) openConfirmWipeDataDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenConfirmWipeData())
}

func (m Model) openConfirmUninstallDialog() Model {
	return m.withDialogState(m.dialogSnapshot().OpenConfirmUninstall())
}

func (m Model) dialogState() dialogstate.State {
	return dialogstate.State{
		Kind:               m.dialog,
		Width:              m.width,
		Inputs:             m.dialogInputs,
		Description:        m.dialogDescription,
		DescriptionEnabled: m.dialogDescriptionOn,
		DescriptionIndex:   m.dialogDescriptionIdx,
		FocusIdx:           m.dialogFocusIdx,
		ErrorMessage:       m.dialogErrorMessage,
		DeleteID:           m.dialogDeleteID,
		DeleteKind:         m.dialogDeleteKind,
		DeleteLabel:        m.dialogDeleteLabel,
		SessionID:          m.dialogSessionID,
		IssueID:            m.dialogIssueID,
		HabitID:            m.dialogHabitID,
		IssueStatus:        m.dialogIssueStatus,
		CheckInDate:        m.dialogCheckInDate,
		RepoID:             m.dialogRepoID,
		RepoName:           m.dialogRepoName,
		RepoItems:          m.dialogRepoItems,
		RepoItemIDs:        m.dialogRepoItemIDs,
		StreamID:           m.dialogStreamID,
		StreamName:         m.dialogStreamName,
		RepoIndex:          m.dialogRepoIndex,
		StreamIndex:        m.dialogStreamIndex,
		Parent:             m.dialogParent,
		DateMonthValue:     m.dialogDateMonth,
		DateCursorValue:    m.dialogDateCursor,
		StashCursor:        m.dialogStashCursor,
		StatusItems:        m.dialogStatusItems,
		StatusCursor:       m.dialogStatusCursor,
		ChoiceItems:        m.dialogChoiceItems,
		ChoiceValues:       m.dialogChoiceValues,
		ChoiceDetails:      m.dialogChoiceDetails,
		TemplateAssets:     m.dialogTemplateAssets,
		ChoiceCursor:       m.dialogChoiceCursor,
		Processing:         m.dialogProcessing,
		ProcessingLabel:    m.dialogProcessingLabel,
		StatusLabel:        m.dialogStatusLabel,
		StatusRequired:     m.dialogStatusRequired,
		ViewTitle:          m.dialogViewTitle,
		ViewName:           m.dialogViewName,
		IssueEstimateMins:  m.dialogIssueEstimateMins,
		ReminderID:         m.dialogReminderID,
		ReminderKind:       m.dialogReminderKind,
		ViewMeta:           m.dialogViewMeta,
		ViewBody:           m.dialogViewBody,
		ViewPath:           m.dialogViewPath,
		SupportBundlePath:  m.dialogSupportBundlePath,
		ProtectionStep:     m.dialogProtectionStep,
		ProtectionCursor:   m.dialogProtectionCursor,
		ProtectionStreaks:  m.dialogProtectionStreaks,
		ProtectionWeekdays: m.dialogProtectionWeekdays,
		ProtectionDates:    m.dialogProtectionDates,
		HabitItems:         m.allHabits,
		HabitStreakStep:    m.dialogHabitStreakStep,
		HabitStreakCursor:  m.dialogHabitStreakCursor,
		HabitStreakDefs:    m.dialogHabitStreakDefs,
		HabitStreakDraft:   m.dialogHabitStreakDraft,
		HabitStreakEditIdx: m.dialogHabitStreakEditIdx,
		ExportPresetKind:   m.dialogExportPresetKind,
		ExportPresetFormat: m.dialogExportPresetFormat,
		ExportPresetOutput: m.dialogExportPresetOutput,
		ExportIncludePDF:   m.dialogExportIncludePDF,
		PromptGlyphMode:    m.dialogPromptGlyphMode,
	}
}

func (m Model) withDialogState(state dialogstate.State) Model {
	m.dialog = state.Kind
	m.dialogInputs = state.Inputs
	m.dialogDescription = state.Description
	m.dialogDescriptionOn = state.DescriptionEnabled
	m.dialogDescriptionIdx = state.DescriptionIndex
	m.dialogFocusIdx = state.FocusIdx
	m.dialogErrorMessage = state.ErrorMessage
	m.dialogDeleteID = state.DeleteID
	m.dialogDeleteKind = state.DeleteKind
	m.dialogDeleteLabel = state.DeleteLabel
	m.dialogSessionID = state.SessionID
	m.dialogIssueID = state.IssueID
	m.dialogHabitID = state.HabitID
	m.dialogIssueStatus = state.IssueStatus
	m.dialogCheckInDate = state.CheckInDate
	m.dialogRepoID = state.RepoID
	m.dialogRepoName = state.RepoName
	m.dialogRepoItems = state.RepoItems
	m.dialogRepoItemIDs = state.RepoItemIDs
	m.dialogStreamID = state.StreamID
	m.dialogStreamName = state.StreamName
	m.dialogRepoIndex = state.RepoIndex
	m.dialogStreamIndex = state.StreamIndex
	m.dialogParent = state.Parent
	m.dialogDateMonth = state.DateMonthValue
	m.dialogDateCursor = state.DateCursorValue
	m.dialogStashCursor = state.StashCursor
	m.dialogStatusItems = state.StatusItems
	m.dialogStatusCursor = state.StatusCursor
	m.dialogChoiceItems = state.ChoiceItems
	m.dialogChoiceValues = state.ChoiceValues
	m.dialogChoiceDetails = state.ChoiceDetails
	m.dialogTemplateAssets = state.TemplateAssets
	m.dialogChoiceCursor = state.ChoiceCursor
	m.dialogProcessing = state.Processing
	m.dialogProcessingLabel = state.ProcessingLabel
	m.dialogStatusLabel = state.StatusLabel
	m.dialogStatusRequired = state.StatusRequired
	m.dialogViewTitle = state.ViewTitle
	m.dialogViewName = state.ViewName
	m.dialogIssueEstimateMins = state.IssueEstimateMins
	m.dialogReminderID = state.ReminderID
	m.dialogReminderKind = state.ReminderKind
	m.dialogViewMeta = state.ViewMeta
	m.dialogViewBody = state.ViewBody
	m.dialogViewPath = state.ViewPath
	m.dialogSupportBundlePath = state.SupportBundlePath
	m.dialogProtectionStep = state.ProtectionStep
	m.dialogProtectionCursor = state.ProtectionCursor
	m.dialogProtectionStreaks = state.ProtectionStreaks
	m.dialogProtectionWeekdays = state.ProtectionWeekdays
	m.dialogProtectionDates = state.ProtectionDates
	m.dialogHabitStreakStep = state.HabitStreakStep
	m.dialogHabitStreakCursor = state.HabitStreakCursor
	m.dialogHabitStreakDefs = state.HabitStreakDefs
	m.dialogHabitStreakDraft = state.HabitStreakDraft
	m.dialogHabitStreakEditIdx = state.HabitStreakEditIdx
	m.dialogExportPresetKind = state.ExportPresetKind
	m.dialogExportPresetFormat = state.ExportPresetFormat
	m.dialogExportPresetOutput = state.ExportPresetOutput
	m.dialogExportIncludePDF = state.ExportIncludePDF
	m.dialogPromptGlyphMode = state.PromptGlyphMode
	return m
}

func (m Model) dialogRuntimeState() dialogruntime.State {
	return dialogruntime.State{
		Context:               m.context,
		Repos:                 m.repos,
		DashboardDate:         m.currentDashboardDate(),
		RollupStartDate:       m.currentRollupStartDate(),
		RollupEndDate:         m.currentRollupEndDate(),
		CurrentExecutablePath: m.currentExecutablePath,
		KernelExecutablePath:  kernelExecutablePath(m.kernelInfo),
		KernelInfo:            m.kernelInfo,
	}
}

func (m Model) dialogRuntimeDeps() dialogruntime.Deps {
	return dialogruntime.Deps{
		CreateScratchpad: func(name, path string) tea.Cmd { return commands.CreateScratchpad(m.client, name, path) },
		CreateRepo: func(name string, description *string) tea.Cmd {
			return commands.CreateRepoOnly(m.client, name, description)
		},
		UpdateRepo: func(repoID int64, name string, description *string) tea.Cmd {
			return commands.UpdateRepo(m.client, repoID, name, description)
		},
		CreateStream: func(repoID int64, name string, description *string) tea.Cmd {
			return commands.CreateStreamOnly(m.client, repoID, name, description)
		},
		UpdateStream: func(repoID, streamID int64, name string, description *string) tea.Cmd {
			return commands.UpdateStream(m.client, repoID, streamID, name, description)
		},
		CreateIssueOnly: func(streamID int64, title string, description *string, estimateMinutes *int, dueDate *string) tea.Cmd {
			return commands.CreateIssueOnly(m.client, streamID, title, description, estimateMinutes, dueDate)
		},
		CreateHabitWithPath: func(repoName, streamName, name string, description *string, schedule string, weekdays []int, estimateMinutes *int) tea.Cmd {
			return commands.CreateHabitWithPath(m.client, repoName, "", streamName, "", name, description, schedule, weekdays, estimateMinutes)
		},
		UpdateHabit: func(habitID, streamID int64, name string, description *string, schedule string, weekdays []int, estimateMinutes *int, active bool, dashboardDate string) tea.Cmd {
			return commands.UpdateHabit(m.client, habitID, streamID, name, description, schedule, weekdays, estimateMinutes, active, dashboardDate)
		},
		CreateIssueWithPath: func(repoName, streamName, title string, description *string, estimateMinutes *int, dueDate *string) tea.Cmd {
			return commands.CreateIssueWithPath(m.client, repoName, "", streamName, "", title, description, estimateMinutes, dueDate)
		},
		CheckoutContext: func(repoID int64, repoName string, streamID int64, streamName string) tea.Cmd {
			return commands.CheckoutContext(m.client, repoID, repoName, streamID, streamName)
		},
		UpsertDailyCheckIn: func(req shareddto.DailyCheckInUpsertRequest, refreshDate string) tea.Cmd {
			return commands.UpsertDailyCheckIn(m.client, req, refreshDate)
		},
		UpdateIssue: func(issueID, streamID int64, title string, description *string, estimateMinutes *int, dueDate *string, dashboardDate string) tea.Cmd {
			return commands.UpdateIssue(m.client, issueID, streamID, title, description, estimateMinutes, dueDate, dashboardDate)
		},
		SetHabitStatus: func(habitID int64, date string, status sharedtypes.HabitCompletionStatus, estimateMinutes *int, note *string) tea.Cmd {
			return commands.SetHabitStatus(m.client, habitID, date, status, estimateMinutes, note)
		},
		CopyDailyReport: func(date string) tea.Cmd { return commands.CopyDailyReport(m.client, date) },
		GenerateCalendarExport: func(req shareddto.ExportCalendarRequest) tea.Cmd {
			return commands.GenerateCalendarExport(m.client, req)
		},
		GenerateReport:      func(req shareddto.ExportReportRequest) tea.Cmd { return commands.GenerateReport(m.client, req) },
		SetExportReportsDir: func(path string) tea.Cmd { return commands.SetExportReportsDir(m.client, path) },
		SetExportICSDir:     func(path string) tea.Cmd { return commands.SetExportICSDir(m.client, path) },
		PatchSetting: func(key sharedtypes.CoreSettingsKey, value any, repoID, streamID int64, dashboardDate string) tea.Cmd {
			return commands.PatchSetting(m.client, key, value, repoID, streamID, dashboardDate)
		},
		CreateAlertReminder: func(input shareddto.AlertReminderCreateRequest) tea.Cmd {
			return commands.CreateAlertReminder(m.client, input)
		},
		UpdateAlertReminder: func(input shareddto.AlertReminderUpdateRequest) tea.Cmd {
			return commands.UpdateAlertReminder(m.client, input)
		},
		DeleteRepo:   func(id int64) tea.Cmd { return commands.DeleteRepo(m.client, id) },
		DeleteStream: func(repoID, streamID int64) tea.Cmd { return commands.DeleteStream(m.client, repoID, streamID) },
		DeleteIssue: func(issueID, streamID int64, dashboardDate string) tea.Cmd {
			return commands.DeleteIssue(m.client, issueID, streamID, dashboardDate)
		},
		DeleteHabit: func(habitID, streamID int64, dashboardDate string) tea.Cmd {
			return commands.DeleteHabit(m.client, habitID, streamID, dashboardDate)
		},
		DeleteDailyCheckIn: func(id string) tea.Cmd { return commands.DeleteDailyCheckIn(m.client, id) },
		DeleteExportReport: func(report api.ExportReportFile) tea.Cmd { return commands.DeleteExportReport(m.client, report) },
		DeleteScratchpad:   func(id string) tea.Cmd { return commands.DeleteScratchpad(m.client, id) },
		ApplyStash:         func(id string) tea.Cmd { return commands.ApplyStash(m.client, id) },
		DropStash:          func(id string) tea.Cmd { return commands.DropStash(m.client, id) },
		ChangeIssueStatus: func(issueID int64, status string, note *string, streamID int64, dashboardDate string) tea.Cmd {
			return commands.ChangeIssueStatus(m.client, issueID, status, note, streamID, dashboardDate)
		},
		AmendSessionNote: func(id string, note string) tea.Cmd { return commands.AmendSessionNote(m.client, id, note) },
		LogManualSession: func(input shareddto.ManualSessionLogRequest) tea.Cmd {
			return commands.LogManualSession(m.client, input)
		},
		EndFocusSession: func(streamID int64, dashboardDate string, payload shareddto.EndSessionRequest) tea.Cmd {
			return commands.EndFocusSession(m.client, streamID, dashboardDate, payload)
		},
		StashFocusSession: func(note string) tea.Cmd { return commands.StashFocusSession(m.client, note) },
		ChangeIssueStatusAndEndSession: func(issueID int64, status string, note *string, streamID int64, dashboardDate string, payload shareddto.EndSessionRequest) tea.Cmd {
			return commands.ChangeIssueStatusAndEndSession(m.client, issueID, status, note, streamID, dashboardDate, payload)
		},
		SetIssueTodoDate: func(issueID int64, date string, streamID int64, dashboardDate string) tea.Cmd {
			return commands.SetIssueTodoDate(m.client, issueID, date, streamID, dashboardDate)
		},
		SetRollupStartDate: func(date, currentEnd string) tea.Cmd {
			if date > currentEnd {
				currentEnd = date
			}
			return commands.SetRollupRange(date, currentEnd)
		},
		SetRollupEndDate: func(currentStart, date string) tea.Cmd {
			if currentStart > date {
				currentStart = date
			}
			return commands.SetRollupRange(currentStart, date)
		},
		WipeRuntimeData: func() tea.Cmd { return commands.WipeRuntimeData(m.client) },
		UninstallCrona: func(currentExecutablePath, kernelExecutablePath string, kernelInfo *api.KernelInfo) tea.Cmd {
			return commands.UninstallCrona(m.client, currentExecutablePath, kernelExecutablePath, kernelInfo)
		},
		OpenSupportIssueURL:       func() tea.Cmd { return m.openSupportIssueURL() },
		OpenSupportDiscussionsURL: func() tea.Cmd { return m.openSupportDiscussionsURL() },
		OpenSupportReleasesURL:    func() tea.Cmd { return m.openSupportReleasesURL() },
		OpenSupportRoadmapURL:     func() tea.Cmd { return m.openSupportRoadmapURL() },
		OpenExternalPath: func(path string) tea.Cmd {
			return commands.OpenExternalPath(path)
		},
		CopyText: func(text, message string) tea.Cmd { return commands.CopyTextToClipboard(text, message) },
		ErrorCmd: func(err error) tea.Cmd { return func() tea.Msg { return commands.ErrMsg{Err: err} } },
		ResolvePatchSettingValue: func(action dialogstate.Action) any {
			switch action.SettingKey {
			case sharedtypes.CoreSettingsKeyRestWeekdays:
				return action.IntList
			case sharedtypes.CoreSettingsKeyDateDisplayFormat:
				return action.Path
			case sharedtypes.CoreSettingsKeyHabitStreakDefs:
				return action.HabitStreakDefs
			default:
				return action.StringList
			}
		},
	}
}
