package dialogstate

import (
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	dialogpkg "crona/tui/internal/tui/dialogs"
	uistate "crona/tui/internal/tui/state"
)

type Snapshot struct {
	Dialog               dialogpkg.State
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

func OpenCreateScratchpad(s Snapshot) dialogpkg.State {
	return dialogpkg.OpenCreateScratchpad(s.Dialog)
}
func OpenCreateRepo(s Snapshot) dialogpkg.State { return dialogpkg.OpenCreateRepo(s.Dialog) }

func OpenEditRepo(s Snapshot, repoID int64, name string, description *string) dialogpkg.State {
	return dialogpkg.OpenEditRepo(s.Dialog, repoID, name, description)
}

func OpenCreateStream(s Snapshot, repoID int64, repoName string) dialogpkg.State {
	return dialogpkg.OpenCreateStream(s.Dialog, repoID, repoName)
}

func OpenEditStream(s Snapshot, streamID, repoID int64, streamName, repoName string, description *string) dialogpkg.State {
	return dialogpkg.OpenEditStream(s.Dialog, streamID, repoID, streamName, repoName, description)
}

func OpenCreateIssueMeta(s Snapshot, streamID int64, streamName, repoName string) dialogpkg.State {
	return dialogpkg.OpenCreateIssueMeta(s.Dialog, streamID, streamName, repoName)
}

func OpenCreateHabit(s Snapshot, streamID int64, streamName, repoName string) dialogpkg.State {
	next := dialogpkg.OpenCreateHabit(s.Dialog)
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
	next = dialogpkg.SyncDialogFocus(next)
	_ = streamID
	return next
}

func OpenEditIssue(s Snapshot, issueID, streamID int64, title string, description *string, estimateMinutes *int, todoForDate *string) dialogpkg.State {
	return dialogpkg.OpenEditIssue(s.Dialog, issueID, streamID, title, description, estimateMinutes, todoForDate)
}

func OpenEditHabit(s Snapshot, habitID, streamID int64, name string, description *string, schedule string, weekdays []int, targetMinutes *int, active bool) dialogpkg.State {
	scheduleValue := schedule
	if schedule == "weekly" {
		scheduleValue = strings.Join(dialogpkg.WeekdayTokens(weekdays), ",")
	}
	return dialogpkg.OpenEditHabit(s.Dialog, habitID, streamID, name, description, scheduleValue, targetMinutes, active)
}

func OpenHabitCompletion(s Snapshot, habitID int64, date string, durationMinutes *int, notes *string) dialogpkg.State {
	return dialogpkg.OpenHabitCompletion(s.Dialog, habitID, date, durationMinutes, notes)
}

func OpenCreateIssueDefault(s Snapshot) dialogpkg.State {
	next := dialogpkg.OpenCreateIssueDefault(s.Dialog)
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
	return dialogpkg.SyncDialogFocus(next)
}

func OpenCheckoutContext(s Snapshot) dialogpkg.State {
	next := dialogpkg.OpenCheckoutContext(s.Dialog)
	if s.Context == nil {
		return next
	}
	if s.Context.RepoName != nil {
		next.Inputs[0].SetValue(*s.Context.RepoName)
		next.RepoIndex = 0
		next.StreamIndex = 0
		next.FocusIdx = 1
		next = dialogpkg.SyncDialogFocus(next)
	}
	if s.Context.StreamName != nil {
		next.Inputs[1].SetValue(*s.Context.StreamName)
		next.StreamIndex = 0
	}
	return next
}

func OpenCreateCheckIn(s Snapshot) dialogpkg.State {
	return dialogpkg.OpenCreateCheckIn(s.Dialog, s.CurrentWellbeingDate)
}

func OpenEditCheckIn(s Snapshot) dialogpkg.State {
	return dialogpkg.OpenEditCheckIn(s.Dialog, s.DailyCheckIn, s.CurrentWellbeingDate)
}

func OpenEditHabitStreaks(s Snapshot) dialogpkg.State {
	return dialogpkg.OpenEditHabitStreaks(s.Dialog, s.Settings, s.AllHabits)
}

func OpenConfirmDelete(s Snapshot, id string) dialogpkg.State {
	return dialogpkg.OpenConfirmDelete(s.Dialog, "scratchpad", id, "this scratchpad", 0, 0)
}

func OpenConfirmDeleteEntity(s Snapshot, kind, id, label string) dialogpkg.State {
	return dialogpkg.OpenConfirmDelete(s.Dialog, kind, id, label, s.Dialog.RepoID, s.Dialog.StreamID)
}

func OpenConfirmWipeData(s Snapshot) dialogpkg.State {
	return dialogpkg.OpenConfirmWipeData(s.Dialog)
}

func OpenConfirmUninstall(s Snapshot) dialogpkg.State {
	return dialogpkg.OpenConfirmUninstall(s.Dialog)
}

func OpenStashList(s Snapshot) dialogpkg.State { return dialogpkg.OpenStashList(s.Dialog) }

func OpenIssueStatus(s Snapshot, status string) dialogpkg.State {
	return dialogpkg.OpenIssueStatus(s.Dialog, status)
}

func OpenIssueStatusNote(s Snapshot, status, label string, required bool) dialogpkg.State {
	return dialogpkg.OpenIssueStatusNote(s.Dialog, s.Dialog.IssueID, s.Dialog.StreamID, status, label, required)
}

func OpenSessionMessage(s Snapshot, kind string) dialogpkg.State {
	return dialogpkg.OpenSessionMessage(s.Dialog, kind)
}

func OpenIssueSessionTransition(s Snapshot, issueID int64, status string) dialogpkg.State {
	return dialogpkg.OpenIssueSessionTransition(s.Dialog, issueID, status)
}

func OpenAmendSession(s Snapshot, sessionID string, commit string) dialogpkg.State {
	return dialogpkg.OpenAmendSession(s.Dialog, sessionID, commit)
}

func OpenManualSession(s Snapshot, issueID int64, issueLabel string, estimateMinutes *int, date string) dialogpkg.State {
	return dialogpkg.OpenManualSession(s.Dialog, issueID, issueLabel, estimateMinutes, date)
}

func OpenDatePicker(s Snapshot, parentDialog string, issueID int64, inputIndex int, initial *string) dialogpkg.State {
	return dialogpkg.OpenDatePicker(s.Dialog, parentDialog, issueID, inputIndex, initial, s.CurrentDashboardDate)
}

func OpenViewEntity(s Snapshot, title, name, meta, body string) dialogpkg.State {
	return dialogpkg.OpenViewEntity(s.Dialog, title, name, meta, body)
}

func OpenViewEntityWithPath(s Snapshot, title, name, meta, body, path string) dialogpkg.State {
	return dialogpkg.OpenViewEntityWithPath(s.Dialog, title, name, meta, body, path)
}

func OpenViewJump(s Snapshot) dialogpkg.State {
	return dialogpkg.OpenViewJump(s.Dialog, s.AvailableViews)
}

func OpenBetaSupport(s Snapshot) dialogpkg.State {
	return dialogpkg.OpenBetaSupport(s.Dialog)
}

func OpenStashConflict(s Snapshot, conflict sharedtypes.StashConflict) dialogpkg.State {
	return dialogpkg.OpenStashConflict(s.Dialog, conflict)
}

func OpenSupportBundleResult(s Snapshot, name, meta, body, path string) dialogpkg.State {
	return dialogpkg.OpenSupportBundleResult(s.Dialog, name, meta, body, path)
}

func OpenUpdateNotes(s Snapshot) dialogpkg.State {
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
	return dialogpkg.OpenViewEntity(s.Dialog, "Update Notes", name, strings.Join(metaParts, "   "), body)
}

func OpenExportDaily(s Snapshot) dialogpkg.State {
	includePDF := s.ExportAssets != nil && s.ExportAssets.PDFRendererAvailable
	var checkedRepoID *int64
	if s.Context != nil {
		checkedRepoID = s.Context.RepoID
	}
	return dialogpkg.OpenExportDaily(s.Dialog, s.CurrentDashboardDate, includePDF, s.Repos, checkedRepoID, s.ExportAssets)
}

func OpenExportReportsDir(s Snapshot, current string) dialogpkg.State {
	return dialogpkg.OpenExportReportsDir(s.Dialog, current)
}

func OpenExportICSDir(s Snapshot, current string) dialogpkg.State {
	return dialogpkg.OpenExportICSDir(s.Dialog, current)
}

func OpenEditDateDisplayFormat(s Snapshot, current string) dialogpkg.State {
	return dialogpkg.OpenEditDateDisplayFormat(s.Dialog, current)
}

func OpenEditRestProtection(s Snapshot) dialogpkg.State {
	if s.Settings == nil {
		return dialogpkg.OpenEditRestProtection(s.Dialog, nil, nil, nil)
	}
	return dialogpkg.OpenEditRestProtection(s.Dialog, s.Settings.FrozenStreakKinds, s.Settings.RestWeekdays, s.Settings.RestSpecificDates)
}

func OpenEditTelemetrySettings(s Snapshot) dialogpkg.State {
	if s.Settings == nil {
		return dialogpkg.OpenEditTelemetrySettings(s.Dialog, false, false)
	}
	return dialogpkg.OpenEditTelemetrySettings(s.Dialog, s.Settings.UsageTelemetryEnabled, s.Settings.ErrorReportingEnabled)
}

func OpenOnboarding(s Snapshot) dialogpkg.State {
	if s.Settings == nil {
		return dialogpkg.OpenOnboarding(s.Dialog, false, false)
	}
	return dialogpkg.OpenOnboarding(s.Dialog, s.Settings.UsageTelemetryEnabled, s.Settings.ErrorReportingEnabled)
}

func OpenCreateAlertReminder(s Snapshot) dialogpkg.State {
	return dialogpkg.OpenCreateAlertReminder(s.Dialog)
}

func OpenEditAlertReminder(s Snapshot, id string) dialogpkg.State {
	for _, reminder := range s.AlertReminders {
		if reminder.ID == strings.TrimSpace(id) {
			return dialogpkg.OpenEditAlertReminder(s.Dialog, reminder)
		}
	}
	return s.Dialog
}
