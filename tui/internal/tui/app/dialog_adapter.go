package app

import (
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
	dialogpkg "crona/tui/internal/tui/app/dialogs"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) openCreateScratchpad() Model {
	return m.withDialogState(dialogpkg.OpenCreateScratchpad(m.dialogState()))
}
func (m Model) openCreateRepoDialog() Model {
	return m.withDialogState(dialogpkg.OpenCreateRepo(m.dialogState()))
}
func (m Model) openEditRepoDialog(repoID int64, name string) Model {
	return m.withDialogState(dialogpkg.OpenEditRepo(m.dialogState(), repoID, name, m.repoDescriptionByID(repoID)))
}
func (m Model) openCreateStreamDialog(repoID int64, repoName string) Model {
	return m.withDialogState(dialogpkg.OpenCreateStream(m.dialogState(), repoID, repoName))
}
func (m Model) openEditStreamDialog(streamID, repoID int64, streamName, repoName string) Model {
	return m.withDialogState(dialogpkg.OpenEditStream(m.dialogState(), streamID, repoID, streamName, repoName, m.streamDescriptionByID(streamID)))
}
func (m Model) openCreateIssueMetaDialog(streamID int64, streamName, repoName string) Model {
	return m.withDialogState(dialogpkg.OpenCreateIssueMeta(m.dialogState(), streamID, streamName, repoName))
}
func (m Model) openCreateHabitDialog(streamID int64, streamName, repoName string) Model {
	next := m.withDialogState(dialogpkg.OpenCreateHabit(m.dialogState()))
	if strings.TrimSpace(repoName) != "" && repoName != "-" {
		next.dialogInputs[0].SetValue(repoName)
		next.dialogRepoIndex = 0
		next.dialogStreamIndex = 0
	}
	if strings.TrimSpace(streamName) != "" && streamName != "-" {
		next.dialogInputs[1].SetValue(streamName)
		next.dialogStreamIndex = 0
	}
	next.dialogFocusIdx = 2
	next = next.withDialogState(dialogpkg.SyncDialogFocus(next.dialogState()))
	_ = streamID
	return next
}
func (m Model) openEditIssueDialog(issueID, streamID int64, title string, description *string, estimateMinutes *int, todoForDate *string) Model {
	return m.withDialogState(dialogpkg.OpenEditIssue(m.dialogState(), issueID, streamID, title, description, estimateMinutes, todoForDate))
}
func (m Model) openEditHabitDialog(habitID, streamID int64, name string, description *string, schedule string, weekdays []int, targetMinutes *int, active bool) Model {
	scheduleValue := schedule
	if schedule == "weekly" {
		scheduleValue = strings.Join(dialogpkg.WeekdayTokens(weekdays), ",")
	}
	return m.withDialogState(dialogpkg.OpenEditHabit(m.dialogState(), habitID, streamID, name, description, scheduleValue, targetMinutes, active))
}
func (m Model) openHabitCompletionDialog(habitID int64, date string, durationMinutes *int, notes *string) Model {
	return m.withDialogState(dialogpkg.OpenHabitCompletion(m.dialogState(), habitID, date, durationMinutes, notes))
}
func (m Model) openCreateIssueDefaultDialog() Model {
	next := m.withDialogState(dialogpkg.OpenCreateIssueDefault(m.dialogState()))
	if next.context == nil {
		return next
	}
	if next.context.RepoName != nil {
		next.dialogInputs[0].SetValue(*next.context.RepoName)
		next.dialogRepoIndex = 0
		next.dialogStreamIndex = 0
	}
	if next.context.StreamName != nil {
		next.dialogInputs[1].SetValue(*next.context.StreamName)
		next.dialogStreamIndex = 0
	}
	next.dialogFocusIdx = 2
	next = next.withDialogState(dialogpkg.SyncDialogFocus(next.dialogState()))
	return next
}
func (m Model) openCheckoutContextDialog() Model {
	next := m.withDialogState(dialogpkg.OpenCheckoutContext(m.dialogState()))
	if next.context == nil {
		return next
	}
	if next.context.RepoName != nil {
		next.dialogInputs[0].SetValue(*next.context.RepoName)
		next.dialogRepoIndex = 0
		next.dialogStreamIndex = 0
		next.dialogFocusIdx = 1
		next = next.withDialogState(dialogpkg.SyncDialogFocus(next.dialogState()))
	}
	if next.context.StreamName != nil {
		next.dialogInputs[1].SetValue(*next.context.StreamName)
		next.dialogStreamIndex = 0
	}
	return next
}
func (m Model) openCreateCheckInDialog() Model {
	return m.withDialogState(dialogpkg.OpenCreateCheckIn(m.dialogState(), m.currentWellbeingDate()))
}
func (m Model) openEditCheckInDialog() Model {
	return m.withDialogState(dialogpkg.OpenEditCheckIn(m.dialogState(), m.dailyCheckIn, m.currentWellbeingDate()))
}
func (m Model) openConfirmDelete(id string) Model {
	return m.withDialogState(dialogpkg.OpenConfirmDelete(m.dialogState(), "scratchpad", id, "this scratchpad", 0, 0))
}
func (m Model) openConfirmDeleteEntity(kind, id, label string) Model {
	return m.withDialogState(dialogpkg.OpenConfirmDelete(m.dialogState(), kind, id, label, m.dialogRepoID, m.dialogStreamID))
}
func (m Model) openStashListDialog() Model {
	return m.withDialogState(dialogpkg.OpenStashList(m.dialogState()))
}
func (m Model) openIssueStatusDialog(status string) Model {
	return m.withDialogState(dialogpkg.OpenIssueStatus(m.dialogState(), status))
}
func (m Model) openIssueStatusNoteDialog(status, label string, required bool) Model {
	return m.withDialogState(dialogpkg.OpenIssueStatusNote(m.dialogState(), m.dialogIssueID, m.dialogStreamID, status, label, required))
}
func (m Model) openSessionMessageDialog(kind string) Model {
	return m.withDialogState(dialogpkg.OpenSessionMessage(m.dialogState(), kind))
}
func (m Model) openIssueSessionTransitionDialog(issueID int64, status string) Model {
	return m.withDialogState(dialogpkg.OpenIssueSessionTransition(m.dialogState(), issueID, status))
}
func (m Model) openAmendSessionDialog(sessionID string, commit string) Model {
	return m.withDialogState(dialogpkg.OpenAmendSession(m.dialogState(), sessionID, commit))
}
func (m Model) openDatePickerDialog(parentDialog string, issueID int64, inputIndex int, initial *string) Model {
	return m.withDialogState(dialogpkg.OpenDatePicker(m.dialogState(), parentDialog, issueID, inputIndex, initial, m.currentDashboardDate()))
}
func (m Model) openViewEntityDialog(title string, name string, meta string, body string) Model {
	return m.withDialogState(dialogpkg.OpenViewEntity(m.dialogState(), title, name, meta, body))
}
func (m Model) openExportDailyDialog() Model {
	includePDF := m.exportAssets != nil && m.exportAssets.PDFRendererAvailable
	return m.withDialogState(dialogpkg.OpenExportDaily(m.dialogState(), m.currentDashboardDate(), includePDF))
}
func (m Model) openExportReportsDirDialog(current string) Model {
	return m.withDialogState(dialogpkg.OpenExportReportsDir(m.dialogState(), current))
}

func (m Model) updateDialog(msg tea.KeyMsg) (Model, tea.Cmd) {
	state, action, status := dialogpkg.Update(m.dialogState(), m.dialogContext(), m.currentDashboardDate(), msg)
	m = m.withDialogState(state)
	if status != "" {
		return m, m.setStatus(status, true)
	}
	if action == nil {
		return m, nil
	}
	return m, m.dialogActionCmd(*action)
}

func (m Model) dialogContext() dialogpkg.UpdateContext {
	ctx := dialogpkg.UpdateContext{
		Repos:     m.repos,
		Streams:   m.streams,
		AllIssues: m.allIssues,
		Context:   m.context,
		Stashes:   m.stashes,
	}
	if issueID, streamID, _, _, ok := m.selectedIssue(); ok {
		ctx.SelectedIssueID = issueID
		ctx.SelectedStreamID = streamID
		ctx.HasSelectedIssue = true
	}
	if issue := m.activeIssueWithMeta(); issue != nil {
		ctx.ActiveIssueStream = issue.StreamID
		ctx.HasActiveIssue = true
	}
	return ctx
}

func (m Model) dialogState() dialogpkg.State {
	return dialogpkg.State{
		Kind:               m.dialog,
		Width:              m.width,
		Inputs:             m.dialogInputs,
		Description:        m.dialogDescription,
		DescriptionEnabled: m.dialogDescriptionOn,
		DescriptionIndex:   m.dialogDescriptionIdx,
		FocusIdx:           m.dialogFocusIdx,
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
		ChoiceCursor:       m.dialogChoiceCursor,
		Processing:         m.dialogProcessing,
		ProcessingLabel:    m.dialogProcessingLabel,
		StatusLabel:        m.dialogStatusLabel,
		StatusRequired:     m.dialogStatusRequired,
		ViewTitle:          m.dialogViewTitle,
		ViewName:           m.dialogViewName,
		ViewMeta:           m.dialogViewMeta,
		ViewBody:           m.dialogViewBody,
	}
}

func (m Model) withDialogState(state dialogpkg.State) Model {
	m.dialog = state.Kind
	m.dialogInputs = state.Inputs
	m.dialogDescription = state.Description
	m.dialogDescriptionOn = state.DescriptionEnabled
	m.dialogDescriptionIdx = state.DescriptionIndex
	m.dialogFocusIdx = state.FocusIdx
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
	m.dialogChoiceCursor = state.ChoiceCursor
	m.dialogProcessing = state.Processing
	m.dialogProcessingLabel = state.ProcessingLabel
	m.dialogStatusLabel = state.StatusLabel
	m.dialogStatusRequired = state.StatusRequired
	m.dialogViewTitle = state.ViewTitle
	m.dialogViewName = state.ViewName
	m.dialogViewMeta = state.ViewMeta
	m.dialogViewBody = state.ViewBody
	return m
}

func (m Model) dialogActionCmd(action dialogpkg.Action) tea.Cmd {
	switch action.Kind {
	case "create_scratchpad":
		return cmdCreateScratchpad(m.client, action.Name, action.Path)
	case "create_repo":
		return cmdCreateRepoOnly(m.client, action.Name, action.Description)
	case "edit_repo":
		return cmdUpdateRepo(m.client, action.RepoID, action.Name, action.Description)
	case "create_stream":
		return cmdCreateStreamOnly(m.client, action.RepoID, action.Name, action.Description)
	case "edit_stream":
		return cmdUpdateStream(m.client, action.RepoID, action.StreamID, action.Name, action.Description)
	case "create_issue_meta":
		return cmdCreateIssueOnly(m.client, action.StreamID, action.Title, action.Description, action.Estimate, action.DueDate)
	case "create_habit":
		return cmdCreateHabitWithPath(m.client, action.RepoName, "", action.StreamName, "", action.Name, action.Description, action.Status, action.Weekdays, action.Estimate)
	case "edit_habit":
		return cmdUpdateHabit(m.client, action.HabitID, action.StreamID, action.Name, action.Description, action.Status, action.Weekdays, action.Estimate, action.Active, m.currentDashboardDate())
	case "create_issue_default":
		return cmdCreateIssueWithPath(m.client, action.RepoName, "", action.StreamName, "", action.Title, action.Description, action.Estimate, action.DueDate)
	case "checkout_context":
		return cmdCheckoutContext(m.client, action.RepoID, action.RepoName, action.StreamID, action.StreamName)
	case "create_checkin", "edit_checkin":
		return cmdUpsertDailyCheckIn(m.client, shareddto.DailyCheckInUpsertRequest{
			Date:              action.CheckInDate,
			Mood:              action.Mood,
			Energy:            action.Energy,
			SleepHours:        action.SleepHours,
			SleepScore:        action.SleepScore,
			ScreenTimeMinutes: action.ScreenTimeMinutes,
			Notes:             action.Note,
		}, action.CheckInDate)
	case "edit_issue":
		return cmdUpdateIssue(m.client, action.IssueID, action.StreamID, action.Title, action.Description, action.Estimate, action.DueDate, m.currentDashboardDate())
	case "complete_habit":
		return cmdSetHabitStatus(m.client, action.HabitID, action.CheckInDate, sharedtypes.HabitCompletionStatusCompleted, action.Estimate, action.Note)
	case "export_report":
		if action.OutputMode == sharedtypes.ExportOutputModeClipboard && action.ReportKind == sharedtypes.ExportReportKindDaily {
			return cmdCopyDailyReport(m.client, action.CheckInDate)
		}
		req := shareddto.ExportReportRequest{
			Kind:       action.ReportKind,
			Date:       action.CheckInDate,
			Format:     action.ReportFormat,
			OutputMode: action.OutputMode,
		}
		if action.ReportKind == sharedtypes.ExportReportKindRepo {
			if m.context == nil || m.context.RepoID == nil {
				return func() tea.Msg { return errMsg{errors.New("repo report requires an active repo context")} }
			}
			req.RepoID = m.context.RepoID
		}
		if action.ReportKind == sharedtypes.ExportReportKindStream {
			if m.context == nil || m.context.StreamID == nil {
				return func() tea.Msg { return errMsg{errors.New("stream report requires an active stream context")} }
			}
			req.StreamID = m.context.StreamID
			if m.context.RepoID != nil {
				req.RepoID = m.context.RepoID
			}
		}
		if action.ReportKind == sharedtypes.ExportReportKindIssueRollup || action.ReportKind == sharedtypes.ExportReportKindCSV {
			if m.context != nil {
				req.RepoID = m.context.RepoID
				req.StreamID = m.context.StreamID
			}
		}
		return cmdGenerateReport(m.client, req)
	case "set_export_reports_dir":
		return cmdSetExportReportsDir(m.client, action.Path)
	case "delete":
		switch action.Name {
		case "repo":
			return cmdDeleteRepo(m.client, dialogpkg.ParseNumericID(action.ID))
		case "stream":
			return cmdDeleteStream(m.client, action.RepoID, dialogpkg.ParseNumericID(action.ID))
		case "issue":
			return cmdDeleteIssue(m.client, dialogpkg.ParseNumericID(action.ID), action.StreamID, m.currentDashboardDate())
		case "habit":
			return cmdDeleteHabit(m.client, dialogpkg.ParseNumericID(action.ID), action.StreamID, m.currentDashboardDate())
		case "checkin":
			return cmdDeleteDailyCheckIn(m.client, action.ID)
		default:
			return cmdDeleteScratchpad(m.client, action.ID)
		}
	case "apply_stash":
		return cmdApplyStash(m.client, action.ID)
	case "drop_stash":
		return cmdDropStash(m.client, action.ID)
	case "change_issue_status":
		return cmdChangeIssueStatus(m.client, action.IssueID, action.Status, action.Note, action.StreamID, m.currentDashboardDate())
	case "amend_session":
		return cmdAmendSessionNote(m.client, action.ID, dialogpkg.ValueOrEmpty(action.Note))
	case "end_session":
		return cmdEndFocusSession(m.client, action.StreamID, m.currentDashboardDate(), action.Payload)
	case "stash_session":
		return cmdStashFocusSession(m.client, dialogpkg.ValueOrEmpty(action.Note))
	case "change_issue_status_and_end_session":
		return cmdChangeIssueStatusAndEndSession(m.client, action.IssueID, action.Status, action.Note, action.StreamID, m.currentDashboardDate(), action.Payload)
	case "set_issue_todo_date":
		due := ""
		if action.DueDate != nil {
			due = *action.DueDate
		}
		return cmdSetIssueTodoDate(m.client, action.IssueID, due, action.StreamID, m.currentDashboardDate())
	default:
		return nil
	}
}

func openEditor(filePath string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}
	c := exec.Command(editor, filePath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errMsg{err}
		}
		return editorDoneMsg{}
	})
}

func openDefaultViewer(filePath string) tea.Cmd {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", filePath)
	case "linux":
		c = exec.Command("xdg-open", filePath)
	case "windows":
		c = exec.Command("cmd", "/c", "start", "", filePath)
	default:
		return func() tea.Msg {
			return errMsg{err: os.ErrInvalid}
		}
	}
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errMsg{err}
		}
		return nil
	})
}

type editorDoneMsg struct{}
