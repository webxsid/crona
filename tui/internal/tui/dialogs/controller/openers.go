package controller

import (
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	uistate "crona/tui/internal/tui/state"
)

type Snapshot struct {
	Dialog               State
	Repos                []api.Repo
	Streams              []api.Stream
	AllHabits            []api.HabitWithMeta
	AllIssues            []api.IssueWithMeta
	Context              *api.ActiveContext
	Stashes              []api.Stash
	DailyCheckIn         *api.DailyCheckIn
	UpdateStatus         *api.UpdateStatus
	ExportAssets         *api.ExportAssetStatus
	Settings             *api.CoreSettings
	AlertReminders       []api.AlertReminder
	CurrentDashboardDate string
	CurrentWellbeingDate string
	ProtectedModeActive  bool
	HasActiveTimer       bool
	AvailableViews       []uistate.View
	HasSelectedIssue     bool
	SelectedIssueID      int64
	SelectedStreamID     int64
	HasActiveIssue       bool
	ActiveIssueStream    int64
}

func (s Snapshot) OpenCreateScratchpad() State {
	return OpenCreateScratchpad(s.Dialog)
}
func (s Snapshot) OpenCreateRepo() State { return OpenCreateRepo(s.Dialog) }

func (s Snapshot) OpenEditRepo(repoID int64, name string, description *string) State {
	return OpenEditRepo(s.Dialog, repoID, name, description)
}

func (s Snapshot) OpenCreateStream(repoID int64, repoName string) State {
	return OpenCreateStream(s.Dialog, repoID, repoName)
}

func (s Snapshot) OpenEditStream(streamID, repoID int64, streamName, repoName string, description *string) State {
	return OpenEditStream(s.Dialog, streamID, repoID, streamName, repoName, description)
}

func (s Snapshot) OpenCreateIssueMeta(streamID int64, streamName, repoName string) State {
	return OpenCreateIssueMeta(s.Dialog, streamID, streamName, repoName)
}

func (s Snapshot) OpenCreateHabit(streamID int64, streamName, repoName string) State {
	next := OpenCreateHabit(s.Dialog)
	if strings.TrimSpace(repoName) != "" && repoName != "-" {
		next.Inputs[0].SetValue(repoName)
		next.RepoIndex = 0
		next.StreamIndex = 0
	}
	if strings.TrimSpace(streamName) != "" && streamName != "-" {
		next.Inputs[1].SetValue(streamName)
		next.StreamIndex = 0
	}
	next.FocusIdx = 2
	next = SyncDialogFocus(next)
	_ = streamID
	return next
}

func (s Snapshot) OpenEditIssue(issueID, streamID int64, title string, description *string, estimateMinutes *int, todoForDate *string) State {
	return OpenEditIssue(s.Dialog, issueID, streamID, title, description, estimateMinutes, todoForDate)
}

func (s Snapshot) OpenEditHabit(habitID, streamID int64, name string, description *string, schedule string, weekdays []int, targetMinutes *int, active bool) State {
	scheduleValue := schedule
	if schedule == "weekly" {
		scheduleValue = strings.Join(WeekdayTokens(weekdays), ",")
	}
	return OpenEditHabit(s.Dialog, habitID, streamID, name, description, scheduleValue, targetMinutes, active)
}

func (s Snapshot) OpenHabitCompletion(habitID int64, date string, durationMinutes *int, notes *string) State {
	return OpenHabitCompletion(s.Dialog, habitID, date, durationMinutes, notes)
}

func (s Snapshot) OpenCreateIssueDefault() State {
	next := OpenCreateIssueDefault(s.Dialog)
	if s.Context == nil {
		return next
	}
	if s.Context.RepoName != nil {
		next.Inputs[0].SetValue(*s.Context.RepoName)
		next.RepoIndex = 0
		next.StreamIndex = 0
	}
	if s.Context.StreamName != nil {
		next.Inputs[1].SetValue(*s.Context.StreamName)
		next.StreamIndex = 0
	}
	next.FocusIdx = 2
	return SyncDialogFocus(next)
}

func (s Snapshot) OpenCheckoutContext() State {
	next := OpenCheckoutContext(s.Dialog)
	if s.Context == nil {
		return next
	}
	if s.Context.RepoName != nil {
		next.Inputs[0].SetValue(*s.Context.RepoName)
		next.RepoIndex = 0
		next.StreamIndex = 0
		next.FocusIdx = 1
		next = SyncDialogFocus(next)
	}
	if s.Context.StreamName != nil {
		next.Inputs[1].SetValue(*s.Context.StreamName)
		next.StreamIndex = 0
	}
	return next
}

func (s Snapshot) OpenCreateCheckIn() State {
	return OpenCreateCheckIn(s.Dialog, s.CurrentWellbeingDate)
}

func (s Snapshot) OpenEditCheckIn() State {
	return OpenEditCheckIn(s.Dialog, s.DailyCheckIn, s.CurrentWellbeingDate)
}

func (s Snapshot) OpenEditHabitStreaks() State {
	return OpenEditHabitStreaks(s.Dialog, s.Settings, s.AllHabits)
}

func (s Snapshot) OpenConfirmDelete(id string) State {
	return OpenConfirmDelete(s.Dialog, "scratchpad", id, "this scratchpad", 0, 0)
}

func (s Snapshot) OpenConfirmDeleteEntity(kind, id, label string) State {
	return OpenConfirmDelete(s.Dialog, kind, id, label, s.Dialog.RepoID, s.Dialog.StreamID)
}

func (s Snapshot) OpenConfirmWipeData() State {
	return OpenConfirmWipeData(s.Dialog)
}

func (s Snapshot) OpenConfirmUninstall() State {
	return OpenConfirmUninstall(s.Dialog)
}

func (s Snapshot) OpenStashList() State { return OpenStashList(s.Dialog) }

func (s Snapshot) OpenIssueStatus(status string) State {
	return OpenIssueStatus(s.Dialog, status)
}

func (s Snapshot) OpenIssueStatusNote(status, label string, required bool) State {
	return OpenIssueStatusNote(s.Dialog, s.Dialog.IssueID, s.Dialog.StreamID, status, label, required)
}

func (s Snapshot) OpenSessionMessage(kind string) State {
	return OpenSessionMessage(s.Dialog, kind)
}

func (s Snapshot) OpenIssueSessionTransition(issueID int64, status string) State {
	return OpenIssueSessionTransition(s.Dialog, issueID, status)
}

func (s Snapshot) OpenAmendSession(sessionID string, commit string) State {
	return OpenAmendSession(s.Dialog, sessionID, commit)
}

func (s Snapshot) OpenManualSession(issueID int64, issueLabel string, estimateMinutes *int, date string) State {
	return OpenManualSession(s.Dialog, issueID, issueLabel, estimateMinutes, date)
}

func (s Snapshot) OpenDatePicker(parentDialog string, issueID int64, inputIndex int, initial *string) State {
	return OpenDatePicker(s.Dialog, parentDialog, issueID, inputIndex, initial, s.CurrentDashboardDate)
}

func (s Snapshot) OpenViewEntity(title, name, meta, body string) State {
	return OpenViewEntity(s.Dialog, title, name, meta, body)
}

func (s Snapshot) OpenViewEntityWithPath(title, name, meta, body, path string) State {
	return OpenViewEntityWithPath(s.Dialog, title, name, meta, body, path)
}

func (s Snapshot) OpenViewJump() State {
	return OpenViewJump(s.Dialog, s.AvailableViews)
}

func (s Snapshot) OpenBetaSupport() State {
	return OpenBetaSupport(s.Dialog)
}

func (s Snapshot) OpenStashConflict(conflict sharedtypes.StashConflict) State {
	return OpenStashConflict(s.Dialog, conflict)
}

func (s Snapshot) OpenSupportBundleResult(name, meta, body, path string) State {
	return OpenSupportBundleResult(s.Dialog, name, meta, body, path)
}

func (s Snapshot) OpenUpdateNotes() State {
	if s.UpdateStatus == nil {
		return s.Dialog
	}
	name := "v" + strings.TrimSpace(s.UpdateStatus.LatestVersion)
	if title := strings.TrimSpace(s.UpdateStatus.ReleaseName); title != "" {
		name += "  " + title
	}
	metaParts := []string{}
	if strings.TrimSpace(s.UpdateStatus.PublishedAt) != "" {
		metaParts = append(metaParts, "Published "+strings.TrimSpace(s.UpdateStatus.PublishedAt))
	}
	if strings.TrimSpace(s.UpdateStatus.ReleaseURL) != "" {
		metaParts = append(metaParts, "URL "+strings.TrimSpace(s.UpdateStatus.ReleaseURL))
	}
	body := strings.TrimSpace(s.UpdateStatus.ReleaseNotes)
	if body == "" {
		body = "No release notes were published for this release."
	}
	return OpenViewEntity(s.Dialog, "Update Notes", name, strings.Join(metaParts, "   "), body)
}

func (s Snapshot) OpenExportDaily() State {
	includePDF := s.ExportAssets != nil && s.ExportAssets.PDFRendererAvailable
	var checkedRepoID *int64
	if s.Context != nil {
		checkedRepoID = s.Context.RepoID
	}
	return OpenExportDaily(s.Dialog, s.CurrentDashboardDate, includePDF, s.Repos, checkedRepoID, s.ExportAssets)
}

func (s Snapshot) OpenExportReportsDir(current string) State {
	return OpenExportReportsDir(s.Dialog, current)
}

func (s Snapshot) OpenExportICSDir(current string) State {
	return OpenExportICSDir(s.Dialog, current)
}

func (s Snapshot) OpenEditDateDisplayFormat(current string) State {
	return OpenEditDateDisplayFormat(s.Dialog, current)
}

func (s Snapshot) OpenEditRestProtection() State {
	if s.Settings == nil {
		return OpenEditRestProtection(s.Dialog, nil, nil, nil)
	}
	return OpenEditRestProtection(s.Dialog, s.Settings.FrozenStreakKinds, s.Settings.RestWeekdays, s.Settings.RestSpecificDates)
}

func (s Snapshot) OpenEditTelemetrySettings() State {
	if s.Settings == nil {
		return OpenEditTelemetrySettings(s.Dialog, false, false)
	}
	return OpenEditTelemetrySettings(s.Dialog, s.Settings.UsageTelemetryEnabled, s.Settings.ErrorReportingEnabled)
}

func (s Snapshot) OpenOnboarding() State {
	if s.Settings == nil {
		return OpenOnboarding(s.Dialog, false, false)
	}
	return OpenOnboarding(s.Dialog, s.Settings.UsageTelemetryEnabled, s.Settings.ErrorReportingEnabled)
}

func (s Snapshot) OpenCreateAlertReminder() State {
	return OpenCreateAlertReminder(s.Dialog)
}

func (s Snapshot) OpenEditAlertReminder(id string) State {
	for _, reminder := range s.AlertReminders {
		if reminder.ID == strings.TrimSpace(id) {
			return OpenEditAlertReminder(s.Dialog, reminder)
		}
	}
	return s.Dialog
}
