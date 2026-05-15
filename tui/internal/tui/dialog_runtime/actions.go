package dialogruntime

import (
	"errors"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	dialogstate "crona/tui/internal/tui/dialogs/controller"

	tea "github.com/charmbracelet/bubbletea"
)

type State struct {
	Context               *api.ActiveContext
	Repos                 []api.Repo
	DashboardDate         string
	RollupStartDate       string
	RollupEndDate         string
	CurrentExecutablePath string
	KernelExecutablePath  string
	KernelInfo            *api.KernelInfo
}

type Deps struct {
	CreateScratchpad               func(name, path string) tea.Cmd
	CreateRepo                     func(name string, description *string) tea.Cmd
	UpdateRepo                     func(repoID int64, name string, description *string) tea.Cmd
	CreateStream                   func(repoID int64, name string, description *string) tea.Cmd
	UpdateStream                   func(repoID, streamID int64, name string, description *string) tea.Cmd
	CreateIssueOnly                func(streamID int64, title string, description *string, estimateMinutes *int, dueDate *string) tea.Cmd
	CreateHabitWithPath            func(repoName, streamName, name string, description *string, schedule string, weekdays []int, estimateMinutes *int) tea.Cmd
	UpdateHabit                    func(habitID, streamID int64, name string, description *string, schedule string, weekdays []int, estimateMinutes *int, active bool, dashboardDate string) tea.Cmd
	CreateIssueWithPath            func(repoName, streamName, title string, description *string, estimateMinutes *int, dueDate *string) tea.Cmd
	CheckoutContext                func(repoID int64, repoName string, streamID int64, streamName string) tea.Cmd
	UpsertDailyCheckIn             func(req shareddto.DailyCheckInUpsertRequest, refreshDate string) tea.Cmd
	UpdateIssue                    func(issueID, streamID int64, title string, description *string, estimateMinutes *int, dueDate *string, dashboardDate string) tea.Cmd
	SetHabitStatus                 func(habitID int64, date string, status sharedtypes.HabitCompletionStatus, estimateMinutes *int, note *string) tea.Cmd
	CopyDailyReport                func(date string) tea.Cmd
	GenerateCalendarExport         func(req shareddto.ExportCalendarRequest) tea.Cmd
	GenerateReport                 func(req shareddto.ExportReportRequest) tea.Cmd
	SetExportReportsDir            func(path string) tea.Cmd
	SetExportICSDir                func(path string) tea.Cmd
	PatchSetting                   func(key sharedtypes.CoreSettingsKey, value any, repoID, streamID int64, dashboardDate string) tea.Cmd
	CreateAlertReminder            func(input shareddto.AlertReminderCreateRequest) tea.Cmd
	UpdateAlertReminder            func(input shareddto.AlertReminderUpdateRequest) tea.Cmd
	DeleteRepo                     func(id int64) tea.Cmd
	DeleteStream                   func(repoID, streamID int64) tea.Cmd
	DeleteIssue                    func(issueID, streamID int64, dashboardDate string) tea.Cmd
	DeleteHabit                    func(habitID, streamID int64, dashboardDate string) tea.Cmd
	DeleteDailyCheckIn             func(id string) tea.Cmd
	DeleteExportReport             func(report api.ExportReportFile) tea.Cmd
	DeleteScratchpad               func(id string) tea.Cmd
	ApplyStash                     func(id string) tea.Cmd
	DropStash                      func(id string) tea.Cmd
	ChangeIssueStatus              func(issueID int64, status string, note *string, streamID int64, dashboardDate string) tea.Cmd
	AmendSessionNote               func(id string, note string) tea.Cmd
	LogManualSession               func(input shareddto.ManualSessionLogRequest) tea.Cmd
	EndFocusSession                func(streamID int64, dashboardDate string, payload shareddto.EndSessionRequest) tea.Cmd
	StashFocusSession              func(note string) tea.Cmd
	ChangeIssueStatusAndEndSession func(issueID int64, status string, note *string, streamID int64, dashboardDate string, payload shareddto.EndSessionRequest) tea.Cmd
	SetIssueTodoDate               func(issueID int64, date string, streamID int64, dashboardDate string) tea.Cmd
	SetRollupStartDate             func(date, currentEnd string) tea.Cmd
	SetRollupEndDate               func(currentStart, date string) tea.Cmd
	WipeRuntimeData                func() tea.Cmd
	UninstallCrona                 func(currentExecutablePath, kernelExecutablePath string, kernelInfo *api.KernelInfo) tea.Cmd
	OpenSupportIssueURL            func() tea.Cmd
	OpenSupportDiscussionsURL      func() tea.Cmd
	OpenSupportReleasesURL         func() tea.Cmd
	OpenSupportRoadmapURL          func() tea.Cmd
	OpenExternalPath               func(path string) tea.Cmd
	CopyText                       func(text, message string) tea.Cmd
	ErrorCmd                       func(error) tea.Cmd
	ResolvePatchSettingValue       func(action dialogstate.Action) any
}

func Resolve(action dialogstate.Action, state State, deps Deps) tea.Cmd {
	r := New[dialogstate.Action, tea.Cmd]()
	r.Register("create_scratchpad", func(action dialogstate.Action) tea.Cmd { return deps.CreateScratchpad(action.Name, action.Path) })
	r.Register("create_repo", func(action dialogstate.Action) tea.Cmd { return deps.CreateRepo(action.Name, action.Description) })
	r.Register("edit_repo", func(action dialogstate.Action) tea.Cmd {
		return deps.UpdateRepo(action.RepoID, action.Name, action.Description)
	})
	r.Register("create_stream", func(action dialogstate.Action) tea.Cmd {
		return deps.CreateStream(action.RepoID, action.Name, action.Description)
	})
	r.Register("edit_stream", func(action dialogstate.Action) tea.Cmd {
		return deps.UpdateStream(action.RepoID, action.StreamID, action.Name, action.Description)
	})
	r.Register("create_issue_meta", func(action dialogstate.Action) tea.Cmd {
		return deps.CreateIssueOnly(action.StreamID, action.Title, action.Description, action.Estimate, action.DueDate)
	})
	r.Register("create_habit", func(action dialogstate.Action) tea.Cmd {
		return deps.CreateHabitWithPath(action.RepoName, action.StreamName, action.Name, action.Description, action.Status, action.Weekdays, action.Estimate)
	})
	r.Register("edit_habit", func(action dialogstate.Action) tea.Cmd {
		return deps.UpdateHabit(action.HabitID, action.StreamID, action.Name, action.Description, action.Status, action.Weekdays, action.Estimate, action.Active, state.DashboardDate)
	})
	r.Register("create_issue_default", func(action dialogstate.Action) tea.Cmd {
		return deps.CreateIssueWithPath(action.RepoName, action.StreamName, action.Title, action.Description, action.Estimate, action.DueDate)
	})
	r.Register("checkout_context", func(action dialogstate.Action) tea.Cmd {
		return deps.CheckoutContext(action.RepoID, action.RepoName, action.StreamID, action.StreamName)
	})
	r.Register("create_checkin", func(action dialogstate.Action) tea.Cmd { return checkinCmd(action, deps) })
	r.Register("edit_checkin", func(action dialogstate.Action) tea.Cmd { return checkinCmd(action, deps) })
	r.Register("edit_issue", func(action dialogstate.Action) tea.Cmd {
		return deps.UpdateIssue(action.IssueID, action.StreamID, action.Title, action.Description, action.Estimate, action.DueDate, state.DashboardDate)
	})
	r.Register("complete_habit", func(action dialogstate.Action) tea.Cmd {
		return deps.SetHabitStatus(action.HabitID, action.CheckInDate, sharedtypes.HabitCompletionStatusCompleted, action.Estimate, action.Note)
	})
	r.Register("export_report", func(action dialogstate.Action) tea.Cmd { return exportCmd(action, state, deps) })
	r.Register("set_export_reports_dir", func(action dialogstate.Action) tea.Cmd { return deps.SetExportReportsDir(action.Path) })
	r.Register("set_export_ics_dir", func(action dialogstate.Action) tea.Cmd { return deps.SetExportICSDir(action.Path) })
	r.Register("patch_setting", func(action dialogstate.Action) tea.Cmd { return patchSettingCmd(action, state, deps) })
	r.Register("create_alert_reminder", func(action dialogstate.Action) tea.Cmd {
		return deps.CreateAlertReminder(shareddto.AlertReminderCreateRequest{
			Kind:         action.ReminderKind,
			ScheduleType: action.ReminderSchedule,
			Weekdays:     action.Weekdays,
			TimeHHMM:     action.ReminderTimeHHMM,
		})
	})
	r.Register("edit_alert_reminder", func(action dialogstate.Action) tea.Cmd {
		return deps.UpdateAlertReminder(shareddto.AlertReminderUpdateRequest{
			ID:           action.ID,
			ScheduleType: &action.ReminderSchedule,
			Weekdays:     action.Weekdays,
			TimeHHMM:     &action.ReminderTimeHHMM,
		})
	})
	r.Register("patch_rest_protection", func(action dialogstate.Action) tea.Cmd { return patchRestProtectionCmd(action, state, deps) })
	r.Register("delete", func(action dialogstate.Action) tea.Cmd { return deleteCmd(action, state, deps) })
	r.Register("apply_stash", func(action dialogstate.Action) tea.Cmd { return deps.ApplyStash(action.ID) })
	r.Register("drop_stash", func(action dialogstate.Action) tea.Cmd { return deps.DropStash(action.ID) })
	r.Register("change_issue_status", func(action dialogstate.Action) tea.Cmd {
		return deps.ChangeIssueStatus(action.IssueID, action.Status, action.Note, action.StreamID, state.DashboardDate)
	})
	r.Register("amend_session", func(action dialogstate.Action) tea.Cmd {
		return deps.AmendSessionNote(action.ID, dialogstate.ValueOrEmpty(action.Note))
	})
	r.Register("manual_session", func(action dialogstate.Action) tea.Cmd {
		if action.ManualSession == nil {
			return deps.ErrorCmd(errors.New("manual session payload is missing"))
		}
		return deps.LogManualSession(*action.ManualSession)
	})
	r.Register("end_session", func(action dialogstate.Action) tea.Cmd {
		return deps.EndFocusSession(action.StreamID, state.DashboardDate, action.Payload)
	})
	r.Register("stash_session", func(action dialogstate.Action) tea.Cmd {
		return deps.StashFocusSession(dialogstate.ValueOrEmpty(action.Note))
	})
	r.Register("change_issue_status_and_end_session", func(action dialogstate.Action) tea.Cmd {
		return deps.ChangeIssueStatusAndEndSession(action.IssueID, action.Status, action.Note, action.StreamID, state.DashboardDate, action.Payload)
	})
	r.Register("set_issue_todo_date", func(action dialogstate.Action) tea.Cmd {
		due := ""
		if action.DueDate != nil {
			due = *action.DueDate
		}
		return deps.SetIssueTodoDate(action.IssueID, due, action.StreamID, state.DashboardDate)
	})
	r.Register("set_rollup_start_date", func(action dialogstate.Action) tea.Cmd {
		date := state.RollupStartDate
		if action.DueDate != nil {
			date = *action.DueDate
		}
		return deps.SetRollupStartDate(date, state.RollupEndDate)
	})
	r.Register("set_rollup_end_date", func(action dialogstate.Action) tea.Cmd {
		date := state.RollupEndDate
		if action.DueDate != nil {
			date = *action.DueDate
		}
		return deps.SetRollupEndDate(state.RollupStartDate, date)
	})
	r.Register("wipe_runtime_data", func(action dialogstate.Action) tea.Cmd { return deps.WipeRuntimeData() })
	r.Register("uninstall_crona", func(action dialogstate.Action) tea.Cmd {
		return deps.UninstallCrona(state.CurrentExecutablePath, state.KernelExecutablePath, state.KernelInfo)
	})
	r.Register("open_support_issue", func(action dialogstate.Action) tea.Cmd { return deps.OpenSupportIssueURL() })
	r.Register("open_support_discussions", func(action dialogstate.Action) tea.Cmd { return deps.OpenSupportDiscussionsURL() })
	r.Register("open_support_releases", func(action dialogstate.Action) tea.Cmd { return deps.OpenSupportReleasesURL() })
	r.Register("open_support_roadmap", func(action dialogstate.Action) tea.Cmd { return deps.OpenSupportRoadmapURL() })
	r.Register("open_support_bundle_folder", func(action dialogstate.Action) tea.Cmd { return deps.OpenExternalPath(action.Path) })
	r.Register("copy_support_bundle_path", func(action dialogstate.Action) tea.Cmd {
		return deps.CopyText(action.Path, "Bundle path copied")
	})
	if cmd, ok := r.Resolve(action.Kind, action); ok {
		return cmd
	}
	return nil
}

func checkinCmd(action dialogstate.Action, deps Deps) tea.Cmd {
	return deps.UpsertDailyCheckIn(shareddto.DailyCheckInUpsertRequest{
		Date:              action.CheckInDate,
		Mood:              action.Mood,
		Energy:            action.Energy,
		SleepHours:        action.SleepHours,
		SleepScore:        action.SleepScore,
		ScreenTimeMinutes: action.ScreenTimeMinutes,
		Notes:             action.Note,
	}, action.CheckInDate)
}

func exportCmd(action dialogstate.Action, state State, deps Deps) tea.Cmd {
	if action.OutputMode == sharedtypes.ExportOutputModeClipboard && action.ReportKind == sharedtypes.ExportReportKindDaily {
		return deps.CopyDailyReport(action.CheckInDate)
	}
	if action.ReportKind == sharedtypes.ExportReportKindCalendar {
		repoID := action.RepoID
		if repoID == 0 {
			if state.Context != nil && state.Context.RepoID != nil {
				repoID = *state.Context.RepoID
			} else if len(state.Repos) > 0 {
				repoID = state.Repos[0].ID
			}
		}
		if repoID == 0 {
			return deps.ErrorCmd(errors.New("calendar export requires a repo"))
		}
		return deps.GenerateCalendarExport(shareddto.ExportCalendarRequest{RepoID: repoID})
	}
	req := shareddto.ExportReportRequest{
		Kind:       action.ReportKind,
		Date:       action.CheckInDate,
		Format:     action.ReportFormat,
		OutputMode: action.OutputMode,
		PresetID:   action.PresetID,
	}
	if action.ReportKind == sharedtypes.ExportReportKindRepo {
		if state.Context == nil || state.Context.RepoID == nil {
			return deps.ErrorCmd(errors.New("repo report requires an active repo context"))
		}
		req.RepoID = state.Context.RepoID
	}
	if action.ReportKind == sharedtypes.ExportReportKindStream {
		if state.Context == nil || state.Context.StreamID == nil {
			return deps.ErrorCmd(errors.New("stream report requires an active stream context"))
		}
		req.StreamID = state.Context.StreamID
		if state.Context.RepoID != nil {
			req.RepoID = state.Context.RepoID
		}
	}
	if action.ReportKind == sharedtypes.ExportReportKindIssueRollup || action.ReportKind == sharedtypes.ExportReportKindCSV {
		if state.Context != nil {
			req.RepoID = state.Context.RepoID
			req.StreamID = state.Context.StreamID
		}
	}
	return deps.GenerateReport(req)
}

func patchSettingCmd(action dialogstate.Action, state State, deps Deps) tea.Cmd {
	value := any(action.StringList)
	if deps.ResolvePatchSettingValue != nil {
		value = deps.ResolvePatchSettingValue(action)
	}
	repoID := int64(0)
	if state.Context != nil && state.Context.RepoID != nil {
		repoID = *state.Context.RepoID
	}
	streamID := int64(0)
	if state.Context != nil && state.Context.StreamID != nil {
		streamID = *state.Context.StreamID
	}
	return deps.PatchSetting(action.SettingKey, value, repoID, streamID, state.DashboardDate)
}

func deleteCmd(action dialogstate.Action, state State, deps Deps) tea.Cmd {
	r := New[dialogstate.Action, tea.Cmd]()
	r.Register("repo", func(action dialogstate.Action) tea.Cmd { return deps.DeleteRepo(dialogstate.ParseNumericID(action.ID)) })
	r.Register("stream", func(action dialogstate.Action) tea.Cmd {
		return deps.DeleteStream(action.RepoID, dialogstate.ParseNumericID(action.ID))
	})
	r.Register("issue", func(action dialogstate.Action) tea.Cmd {
		return deps.DeleteIssue(dialogstate.ParseNumericID(action.ID), action.StreamID, state.DashboardDate)
	})
	r.Register("habit", func(action dialogstate.Action) tea.Cmd {
		return deps.DeleteHabit(dialogstate.ParseNumericID(action.ID), action.StreamID, state.DashboardDate)
	})
	r.Register("checkin", func(action dialogstate.Action) tea.Cmd { return deps.DeleteDailyCheckIn(action.ID) })
	r.Register("report", func(action dialogstate.Action) tea.Cmd {
		return deps.DeleteExportReport(api.ExportReportFile{Name: action.Title, Path: action.ID})
	})
	if cmd, ok := r.Resolve(action.Name, action); ok {
		return cmd
	}
	return deps.DeleteScratchpad(action.ID)
}

func patchRestProtectionCmd(action dialogstate.Action, state State, deps Deps) tea.Cmd {
	repoID := int64(0)
	if state.Context != nil && state.Context.RepoID != nil {
		repoID = *state.Context.RepoID
	}
	streamID := int64(0)
	if state.Context != nil && state.Context.StreamID != nil {
		streamID = *state.Context.StreamID
	}
	return tea.Batch(
		deps.PatchSetting(sharedtypes.CoreSettingsKeyFrozenStreakKinds, action.StreakKinds, repoID, streamID, state.DashboardDate),
		deps.PatchSetting(sharedtypes.CoreSettingsKeyRestWeekdays, action.IntList, repoID, streamID, state.DashboardDate),
		deps.PatchSetting(sharedtypes.CoreSettingsKeyRestSpecificDates, action.RestDates, repoID, streamID, state.DashboardDate),
	)
}
