package dialogs

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"

	"crona/tui/internal/api"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UpdateContext struct {
	Repos             []api.Repo
	Streams           []api.Stream
	AllIssues         []api.IssueWithMeta
	Context           *api.ActiveContext
	Stashes           []api.Stash
	SelectedIssueID   int64
	SelectedStreamID  int64
	HasSelectedIssue  bool
	ActiveIssueStream int64
	HasActiveIssue    bool
}

type Action struct {
	Kind              string
	ReportKind        sharedtypes.ExportReportKind
	ReportFormat      sharedtypes.ExportFormat
	OutputMode        sharedtypes.ExportOutputMode
	PresetID          string
	ID                string
	RepoID            int64
	StreamID          int64
	IssueID           int64
	HabitID           int64
	Name              string
	Path              string
	CheckInDate       string
	RepoName          string
	StreamName        string
	Title             string
	Description       *string
	Status            string
	Weekdays          []int
	Active            bool
	Estimate          *int
	DueDate           *string
	Note              *string
	SettingKey        sharedtypes.CoreSettingsKey
	StringList        []string
	IntList           []int
	StreakKinds       []string
	RestDates         []string
	Mood              int
	Energy            int
	SleepHours        *float64
	SleepScore        *int
	ScreenTimeMinutes *int
	Payload           shareddto.EndSessionRequest
	ManualSession     *shareddto.ManualSessionLogRequest
}

func OpenCreateScratchpad(state State) State {
	name := textinput.New()
	name.Placeholder = "My notes"
	name.Focus()
	name.CharLimit = 80
	name.Width = 40
	path := textinput.New()
	path.Placeholder = "notes/[[date]].md"
	path.CharLimit = 120
	path.Width = 40
	state = Close(state)
	state.Kind = "create_scratchpad"
	state.Inputs = []textinput.Model{name, path}
	return state
}

func OpenExportDaily(state State, date string, includePDF bool, repos []api.Repo, checkedRepoID *int64, assets *api.ExportAssetStatus) State {
	state = Close(state)
	state.Kind = "export_report_category"
	state.CheckInDate = date
	state.ExportIncludePDF = includePDF
	state.ExportCategory = ""
	items, values, details := exportReportCategories()
	state.ChoiceItems = items
	state.ChoiceValues = values
	state.ChoiceDetails = details
	state.RepoItems = make([]string, 0, len(repos))
	state.RepoItemIDs = make([]int64, 0, len(repos))
	selectedRepoIdx := 0
	for i, repo := range repos {
		state.RepoItems = append(state.RepoItems, repo.Name)
		state.RepoItemIDs = append(state.RepoItemIDs, repo.ID)
		if checkedRepoID != nil && repo.ID == *checkedRepoID {
			selectedRepoIdx = i
		}
	}
	if len(state.RepoItems) > 0 {
		state.RepoIndex = selectedRepoIdx
		state.RepoID = state.RepoItemIDs[selectedRepoIdx]
		state.RepoName = state.RepoItems[selectedRepoIdx]
	}
	state.ChoiceCursor = 0
	if assets != nil {
		state.TemplateAssets = append([]sharedtypes.ExportTemplateAsset(nil), assets.TemplateAssets...)
	}
	return state
}

func exportReportCategories() ([]string, []string, []string) {
	return []string{
			"Narrative Reports",
			"Project Reports",
			"Data Exports",
		}, []string{
			"narrative",
			"project",
			"data",
		}, []string{
			"Daily and weekly summaries, including Markdown, PDF, and clipboard output.",
			"Repo, stream, and issue rollup reports for focused project review.",
			"CSV session dumps and calendar exports for external tools and automations.",
		}
}

func exportReportChoices(category string, includePDF bool) ([]string, []string) {
	switch category {
	case "project":
		return []string{
				"Repo report: write Markdown file",
				"Stream report: write Markdown file",
				"Issue rollup: write Markdown file",
			}, []string{
				"Project-level summaries and issue rollups.",
				"Focused stream review with grouped issue output.",
				"Per-issue work summary and notes rollup.",
			}
	case "data":
		return []string{
				"CSV session export: write file",
				"Calendar export: write ICS file",
			}, []string{
				"Raw session data for spreadsheets or external analysis.",
				"Write repo-scoped ICS files for local calendar automation.",
			}
	default:
		items := []string{
			"Daily report: write Markdown file",
			"Daily report: copy to clipboard",
			"Weekly summary: write Markdown file",
		}
		details := []string{
			"Generate a readable daily report as Markdown.",
			"Copy the daily report directly to the clipboard.",
			"Generate a weekly narrative summary as Markdown.",
		}
		if includePDF {
			items = append([]string{
				"Daily report: write PDF file",
				"Weekly summary: write PDF file",
			}, items...)
			details = append([]string{
				"Generate a styled daily PDF report.",
				"Generate a styled weekly PDF report.",
			}, details...)
		}
		return items, details
	}
}

func OpenExportPreset(state State, reportKind sharedtypes.ExportReportKind, format sharedtypes.ExportFormat, outputMode sharedtypes.ExportOutputMode) State {
	state.Parent = "export_report"
	state.Kind = "export_preset"
	state.ExportPresetKind = reportKind
	state.ExportPresetFormat = format
	state.ExportPresetOutput = outputMode
	state.ChoiceItems = nil
	state.ChoiceValues = nil
	state.ChoiceDetails = nil
	state.ChoiceCursor = 0
	assetKinds := []sharedtypes.ExportAssetKind{sharedtypes.ExportAssetKindTemplateMarkdown}
	if format == sharedtypes.ExportFormatPDF {
		assetKinds = []sharedtypes.ExportAssetKind{
			sharedtypes.ExportAssetKindTemplatePDFHTML,
			sharedtypes.ExportAssetKindTemplatePDF,
		}
	}
	for _, asset := range state.TemplateAssets {
		if asset.ReportKind != reportKind || !slices.Contains(assetKinds, asset.AssetKind) {
			continue
		}
		if asset.Customized {
			currentLabel := "Current custom template"
			currentDetail := "Uses your edited active template without replacing it."
			if format == sharedtypes.ExportFormatPDF {
				currentLabel = "Current custom HTML + CSS"
				currentDetail = "Uses your edited active PDF HTML template and paired stylesheet without replacing them."
			}
			state.ChoiceItems = append(state.ChoiceItems, currentLabel)
			state.ChoiceValues = append(state.ChoiceValues, "__current__")
			state.ChoiceDetails = append(state.ChoiceDetails, currentDetail)
		}
		selectedIdx := 0
		for _, preset := range asset.Presets {
			label := preset.Label
			if asset.SelectedPreset != nil && preset.ID == asset.SelectedPreset.ID {
				label += "  [saved default]"
				selectedIdx = len(state.ChoiceItems)
			}
			state.ChoiceItems = append(state.ChoiceItems, label)
			state.ChoiceValues = append(state.ChoiceValues, preset.ID)
			state.ChoiceDetails = append(state.ChoiceDetails, preset.PreviewBody)
		}
		state.ChoiceCursor = selectedIdx
		break
	}
	return state
}

func OpenExportReportChoices(state State, category string) State {
	state.Kind = "export_report"
	state.Parent = "export_report_category"
	state.ExportCategory = category
	state.ChoiceCursor = 0
	state.ChoiceItems, state.ChoiceDetails = exportReportChoices(category, state.ExportIncludePDF)
	state.ChoiceValues = nil
	return state
}

func OpenExportCalendarRepo(state State) State {
	if len(state.RepoItems) == 0 {
		return state
	}
	state.Kind = "export_calendar_repo"
	state.Parent = "export_report"
	state.ChoiceItems = append([]string(nil), state.RepoItems...)
	if state.RepoIndex < 0 || state.RepoIndex >= len(state.ChoiceItems) {
		state.RepoIndex = 0
	}
	state.ChoiceCursor = state.RepoIndex
	return state
}

func OpenExportReportsDir(state State, current string) State {
	input := textinput.New()
	input.Placeholder = "Reports directory"
	input.SetValue(strings.TrimSpace(current))
	input.Focus()
	input.CharLimit = 240
	input.Width = 56
	state = Close(state)
	state.Kind = "edit_export_reports_dir"
	state.Inputs = []textinput.Model{input}
	return state
}

func OpenExportICSDir(state State, current string) State {
	input := textinput.New()
	input.Placeholder = "ICS export directory"
	input.SetValue(strings.TrimSpace(current))
	input.Focus()
	input.CharLimit = 240
	input.Width = 56
	state = Close(state)
	state.Kind = "edit_export_ics_dir"
	state.Inputs = []textinput.Model{input}
	return state
}

func OpenEditRestProtection(state State, streaks []sharedtypes.StreakKind, weekdays []int, dates []string) State {
	state = Close(state)
	state.Kind = "edit_rest_protection"
	state.ProtectionStep = 0
	state.ProtectionCursor = 0
	if len(streaks) == 0 {
		state.ProtectionStreaks = append([]sharedtypes.StreakKind(nil), sharedtypes.AvailableStreakKinds()...)
	} else {
		state.ProtectionStreaks = append([]sharedtypes.StreakKind(nil), streaks...)
	}
	state.ProtectionWeekdays = append([]int(nil), weekdays...)
	state.ProtectionDates = normalizedDateList(dates)
	return state
}

func OpenCreateCheckIn(state State, date string) State {
	return openCheckInDialog(state, "create_checkin", date, nil)
}

func OpenEditCheckIn(state State, checkIn *api.DailyCheckIn, date string) State {
	return openCheckInDialog(state, "edit_checkin", date, checkIn)
}

func openCheckInDialog(state State, kind string, date string, checkIn *api.DailyCheckIn) State {
	mood := textinput.New()
	mood.Placeholder = "Mood 1-5"
	mood.CharLimit = 1
	mood.Width = 12
	mood.Focus()
	energy := textinput.New()
	energy.Placeholder = "Energy 1-5"
	energy.CharLimit = 1
	energy.Width = 12
	sleepHours := textinput.New()
	sleepHours.Placeholder = "Sleep hours"
	sleepHours.CharLimit = 5
	sleepHours.Width = 16
	sleepScore := textinput.New()
	sleepScore.Placeholder = "Sleep score"
	sleepScore.CharLimit = 3
	sleepScore.Width = 16
	screenTime := textinput.New()
	screenTime.Placeholder = "Screen minutes"
	screenTime.CharLimit = 4
	screenTime.Width = 16
	notes := textinput.New()
	notes.Placeholder = "Notes (optional)"
	notes.CharLimit = 200
	notes.Width = 52
	if checkIn != nil {
		mood.SetValue(strconv.Itoa(checkIn.Mood))
		energy.SetValue(strconv.Itoa(checkIn.Energy))
		if checkIn.SleepHours != nil {
			sleepHours.SetValue(strconv.FormatFloat(*checkIn.SleepHours, 'f', -1, 64))
		}
		if checkIn.SleepScore != nil {
			sleepScore.SetValue(strconv.Itoa(*checkIn.SleepScore))
		}
		if checkIn.ScreenTimeMinutes != nil {
			screenTime.SetValue(strconv.Itoa(*checkIn.ScreenTimeMinutes))
		}
		if checkIn.Notes != nil {
			notes.SetValue(strings.TrimSpace(*checkIn.Notes))
		}
	}
	state = Close(state)
	state.Kind = kind
	state.CheckInDate = date
	state.Inputs = []textinput.Model{mood, energy, sleepHours, sleepScore, screenTime, notes}
	return state
}

func OpenCreateRepo(state State) State {
	name := textinput.New()
	name.Placeholder = "Repo name"
	name.Focus()
	name.CharLimit = 80
	name.Width = 40
	description := newDescriptionInput(40, 4)
	state = Close(state)
	state.Kind = "create_repo"
	state.Inputs = []textinput.Model{name}
	state.Description = description
	state.DescriptionEnabled = true
	state.DescriptionIndex = 1
	return state
}

func OpenEditRepo(state State, repoID int64, name string, descriptionValue *string) State {
	input := textinput.New()
	input.Placeholder = "Repo name"
	input.SetValue(name)
	input.Focus()
	input.CharLimit = 80
	input.Width = 40
	description := newDescriptionInput(40, 4)
	if descriptionValue != nil {
		description.SetValue(strings.TrimSpace(*descriptionValue))
	}
	state = Close(state)
	state.Kind = "edit_repo"
	state.Inputs = []textinput.Model{input}
	state.Description = description
	state.DescriptionEnabled = true
	state.DescriptionIndex = 1
	state.RepoID = repoID
	return state
}

func OpenCreateStream(state State, repoID int64, repoName string) State {
	input := textinput.New()
	input.Placeholder = "Stream name"
	input.Focus()
	input.CharLimit = 80
	input.Width = 40
	description := newDescriptionInput(40, 4)
	state = Close(state)
	state.Kind = "create_stream"
	state.Inputs = []textinput.Model{input}
	state.Description = description
	state.DescriptionEnabled = true
	state.DescriptionIndex = 1
	state.RepoID = repoID
	state.RepoName = repoName
	return state
}

func OpenEditStream(state State, streamID, repoID int64, streamName, repoName string, descriptionValue *string) State {
	input := textinput.New()
	input.Placeholder = "Stream name"
	input.SetValue(streamName)
	input.Focus()
	input.CharLimit = 80
	input.Width = 40
	description := newDescriptionInput(40, 4)
	if descriptionValue != nil {
		description.SetValue(strings.TrimSpace(*descriptionValue))
	}
	state = Close(state)
	state.Kind = "edit_stream"
	state.Inputs = []textinput.Model{input}
	state.Description = description
	state.DescriptionEnabled = true
	state.DescriptionIndex = 1
	state.StreamID = streamID
	state.RepoID = repoID
	state.RepoName = repoName
	return state
}

func OpenCreateHabit(state State) State {
	repoFilter := textinput.New()
	repoFilter.Placeholder = "Search repo"
	repoFilter.CharLimit = 80
	repoFilter.Width = 52
	streamFilter := textinput.New()
	streamFilter.Placeholder = "Search stream"
	streamFilter.CharLimit = 80
	streamFilter.Width = 52
	name := textinput.New()
	name.Placeholder = "Habit name"
	name.Focus()
	name.CharLimit = 120
	name.Width = 52
	description := newDescriptionInput(52, 5)
	schedule := textinput.New()
	schedule.Placeholder = "daily | weekdays | mon,wed,fri"
	schedule.CharLimit = 32
	schedule.Width = 52
	target := textinput.New()
	target.Placeholder = "Target minutes (optional)"
	target.CharLimit = 6
	target.Width = 52
	state = Close(state)
	state.Kind = "create_habit"
	state.Inputs = []textinput.Model{repoFilter, streamFilter, name, schedule, target}
	state.Description = description
	state.DescriptionEnabled = true
	state.DescriptionIndex = 3
	state.FocusIdx = 2
	return SyncDialogFocus(state)
}

func OpenEditHabit(state State, habitID, streamID int64, name string, descriptionValue *string, scheduleRaw string, targetMinutes *int, active bool) State {
	nameInput := textinput.New()
	nameInput.Placeholder = "Habit name"
	nameInput.SetValue(name)
	nameInput.Focus()
	nameInput.CharLimit = 120
	nameInput.Width = 52
	description := newDescriptionInput(52, 5)
	if descriptionValue != nil {
		description.SetValue(strings.TrimSpace(*descriptionValue))
	}
	schedule := textinput.New()
	schedule.Placeholder = "daily | weekdays | mon,wed,fri"
	schedule.SetValue(scheduleRaw)
	schedule.CharLimit = 32
	schedule.Width = 52
	target := textinput.New()
	target.Placeholder = "Target minutes (optional)"
	target.CharLimit = 6
	target.Width = 52
	if targetMinutes != nil {
		target.SetValue(strconv.Itoa(*targetMinutes))
	}
	state = Close(state)
	state.Kind = "edit_habit"
	state.Inputs = []textinput.Model{nameInput, schedule, target}
	state.Description = description
	state.DescriptionEnabled = true
	state.DescriptionIndex = 1
	state.HabitID = habitID
	state.StreamID = streamID
	state.StatusLabel = map[bool]string{true: "active", false: "inactive"}[active]
	return state
}

func OpenHabitCompletion(state State, habitID int64, date string, durationMinutes *int, notes *string) State {
	duration := textinput.New()
	duration.Placeholder = "Duration minutes (optional)"
	duration.Focus()
	duration.CharLimit = 6
	duration.Width = 52
	if durationMinutes != nil {
		duration.SetValue(strconv.Itoa(*durationMinutes))
	}
	description := newDescriptionInput(52, 5)
	if notes != nil {
		description.SetValue(strings.TrimSpace(*notes))
	}
	state = Close(state)
	state.Kind = "complete_habit"
	state.Inputs = []textinput.Model{duration}
	state.Description = description
	state.DescriptionEnabled = true
	state.DescriptionIndex = 1
	state.HabitID = habitID
	state.CheckInDate = date
	return state
}

func OpenCreateIssueMeta(state State, streamID int64, streamName, repoName string) State {
	title := textinput.New()
	title.Placeholder = "Issue title"
	title.Focus()
	title.CharLimit = 120
	title.Width = 52
	description := newDescriptionInput(52, 5)
	estimate := textinput.New()
	estimate.Placeholder = "Estimate minutes (optional)"
	estimate.CharLimit = 6
	estimate.Width = 52
	due := textinput.New()
	due.Placeholder = "Due date YYYY-MM-DD (optional)"
	due.CharLimit = 10
	due.Width = 52
	state = Close(state)
	state.Kind = "create_issue_meta"
	state.Inputs = []textinput.Model{title, estimate, due}
	state.Description = description
	state.DescriptionEnabled = true
	state.DescriptionIndex = 1
	state.StreamID = streamID
	state.StreamName = streamName
	state.RepoName = repoName
	return state
}

func OpenEditIssue(state State, issueID, streamID int64, title string, descriptionValue *string, estimateMinutes *int, todoForDate *string) State {
	titleInput := textinput.New()
	titleInput.Placeholder = "Issue title"
	titleInput.SetValue(title)
	titleInput.Focus()
	titleInput.CharLimit = 120
	titleInput.Width = 52
	descriptionInput := newDescriptionInput(52, 5)
	if descriptionValue != nil {
		descriptionInput.SetValue(strings.TrimSpace(*descriptionValue))
	}
	estimateInput := textinput.New()
	estimateInput.Placeholder = "Estimate minutes (optional)"
	estimateInput.CharLimit = 6
	estimateInput.Width = 52
	if estimateMinutes != nil {
		estimateInput.SetValue(strconv.Itoa(*estimateMinutes))
	}
	dueInput := textinput.New()
	dueInput.Placeholder = "Due date YYYY-MM-DD (optional)"
	dueInput.CharLimit = 10
	dueInput.Width = 52
	if todoForDate != nil {
		dueInput.SetValue(strings.TrimSpace(*todoForDate))
	}
	state = Close(state)
	state.Kind = "edit_issue"
	state.Inputs = []textinput.Model{titleInput, estimateInput, dueInput}
	state.Description = descriptionInput
	state.DescriptionEnabled = true
	state.DescriptionIndex = 1
	state.IssueID = issueID
	state.StreamID = streamID
	return state
}

func OpenCreateIssueDefault(state State) State {
	repoFilter := textinput.New()
	repoFilter.Placeholder = "Search repo"
	repoFilter.CharLimit = 80
	repoFilter.Width = 52
	streamFilter := textinput.New()
	streamFilter.Placeholder = "Search stream"
	streamFilter.CharLimit = 80
	streamFilter.Width = 52
	title := textinput.New()
	title.Placeholder = "Issue title"
	title.Focus()
	title.CharLimit = 120
	title.Width = 52
	description := newDescriptionInput(52, 5)
	estimate := textinput.New()
	estimate.Placeholder = "Estimate minutes (optional)"
	estimate.CharLimit = 6
	estimate.Width = 52
	due := textinput.New()
	due.Placeholder = "Due date YYYY-MM-DD (optional)"
	due.CharLimit = 10
	due.Width = 52
	state = Close(state)
	state.Kind = "create_issue_default"
	state.Inputs = []textinput.Model{repoFilter, streamFilter, title, estimate, due}
	state.Description = description
	state.DescriptionEnabled = true
	state.DescriptionIndex = 3
	state.FocusIdx = 2
	return SyncDialogFocus(state)
}

func OpenCheckoutContext(state State) State {
	repoFilter := textinput.New()
	repoFilter.Placeholder = "Search repo"
	repoFilter.CharLimit = 80
	repoFilter.Width = 52
	repoFilter.Focus()
	streamFilter := textinput.New()
	streamFilter.Placeholder = "Search stream"
	streamFilter.CharLimit = 80
	streamFilter.Width = 52
	state = Close(state)
	state.Kind = "checkout_context"
	state.Inputs = []textinput.Model{repoFilter, streamFilter}
	return state
}

func OpenConfirmDelete(state State, kind, id, label string, repoID, streamID int64) State {
	state = Close(state)
	state.Kind = "confirm_delete"
	state.DeleteKind = kind
	state.DeleteID = id
	state.DeleteLabel = label
	state.RepoID = repoID
	state.StreamID = streamID
	return state
}

func OpenConfirmWipeData(state State) State {
	state = Close(state)
	state.Kind = "confirm_wipe"
	state.DeleteLabel = "all Crona runtime data"
	return state
}

func OpenConfirmUninstall(state State) State {
	state = Close(state)
	state.Kind = "confirm_uninstall"
	state.DeleteLabel = "the installed Crona binaries and runtime data"
	return state
}

func newDescriptionInput(width, height int) textarea.Model {
	input := textarea.New()
	input.Placeholder = "Description (optional)"
	input.SetWidth(width)
	input.SetHeight(height)
	input.CharLimit = 2000
	input.ShowLineNumbers = false
	input.FocusedStyle.CursorLine = lipgloss.NewStyle()
	input.Blur()
	return input
}

func OpenStashList(state State) State {
	state = Close(state)
	state.Kind = "stash_list"
	return state
}

func OpenIssueStatus(state State, status string) State {
	state = Close(state)
	state.Kind = "issue_status"
	state.StatusItems = sharedtypes.AllowedIssueStatusTransitions(sharedtypes.IssueStatus(status))
	return state
}

func OpenIssueStatusNote(state State, issueID, streamID int64, status, label string, required bool) State {
	input := textinput.New()
	input.Placeholder = label
	input.Focus()
	input.CharLimit = 200
	input.Width = 48
	state = Close(state)
	state.Kind = "issue_status_note"
	state.Inputs = []textinput.Model{input}
	state.IssueID = issueID
	state.StreamID = streamID
	state.IssueStatus = status
	state.StatusLabel = label
	state.StatusRequired = required
	return state
}

func OpenSessionMessage(state State, kind string) State {
	input := textinput.New()
	if kind == "end_session" {
		input.Placeholder = "Commit message"
	} else {
		input.Placeholder = "Stash note"
	}
	input.Focus()
	input.CharLimit = 200
	input.Width = 48
	state = Close(state)
	state.Kind = kind
	state.Inputs = []textinput.Model{input}
	return state
}

func OpenIssueSessionTransition(state State, issueID int64, status string) State {
	state = Close(state)
	state.Kind = "issue_session_transition"
	state.IssueID = issueID
	state.IssueStatus = status
	switch status {
	case "done":
		input := textinput.New()
		input.Placeholder = "Completion note (optional)"
		input.Focus()
		input.CharLimit = 200
		input.Width = 48
		state.Inputs = []textinput.Model{input}
	case "abandoned":
		input := textinput.New()
		input.Placeholder = "Abandon reason"
		input.Focus()
		input.CharLimit = 200
		input.Width = 48
		state.Inputs = []textinput.Model{input}
	}
	return state
}

func OpenAmendSession(state State, sessionID string, commit string) State {
	input := textinput.New()
	input.Placeholder = "Commit message"
	input.SetValue(strings.TrimSpace(commit))
	input.Focus()
	input.CharLimit = 200
	input.Width = 48
	state = Close(state)
	state.Kind = "amend_session"
	state.SessionID = sessionID
	state.Inputs = []textinput.Model{input}
	return state
}

func OpenManualSession(state State, issueID int64, issueLabel string, estimateMinutes *int, date string) State {
	inputs := []textinput.Model{
		newSessionDetailInput("Summary (optional)"),
		newSessionDetailInput("YYYY-MM-DD"),
		newSessionDetailInput("90m"),
		newSessionDetailInput("15m"),
		newSessionDetailInput("09:00"),
		newSessionDetailInput("10:45"),
		newSessionDetailInput("Notes (optional)"),
	}
	inputs[1].SetValue(strings.TrimSpace(date))
	inputs[2].Focus()
	state = Close(state)
	state.Kind = "manual_session"
	state.IssueID = issueID
	state.ViewName = strings.TrimSpace(issueLabel)
	state.IssueEstimateMins = estimateMinutes
	state.Inputs = inputs
	state.FocusIdx = 2
	return SyncDialogFocus(state)
}

func OpenDatePicker(state State, parentDialog string, issueID int64, inputIndex int, initial *string, currentDate string) State {
	selected := ResolveDialogDate(initial, currentDate)
	monthStart := time.Date(selected.Year(), selected.Month(), 1, 0, 0, 0, 0, selected.Location())
	state.Parent = parentDialog
	state.IssueID = issueID
	state.Kind = "pick_date"
	state.DateCursorValue = selected.Format("2006-01-02")
	state.DateMonthValue = monthStart.Format("2006-01-02")
	state.FocusIdx = inputIndex
	return state
}

func Close(state State) State {
	state.Kind = ""
	state.ErrorMessage = ""
	state.Inputs = nil
	state.Description = textarea.Model{}
	state.DescriptionEnabled = false
	state.DescriptionIndex = 0
	state.FocusIdx = 0
	state.DeleteID = ""
	state.DeleteKind = ""
	state.DeleteLabel = ""
	state.SessionID = ""
	state.IssueID = 0
	state.HabitID = 0
	state.IssueStatus = ""
	state.RepoID = 0
	state.RepoName = ""
	state.RepoItems = nil
	state.RepoItemIDs = nil
	state.StreamID = 0
	state.StreamName = ""
	state.RepoIndex = 0
	state.StreamIndex = 0
	state.Parent = ""
	state.DateMonthValue = ""
	state.DateCursorValue = ""
	state.StashCursor = 0
	state.StatusItems = nil
	state.StatusCursor = 0
	state.ChoiceItems = nil
	state.ChoiceValues = nil
	state.ChoiceDetails = nil
	state.ChoiceCursor = 0
	state.Processing = false
	state.ProcessingLabel = ""
	state.StatusLabel = ""
	state.StatusRequired = false
	state.CheckInDate = ""
	state.ViewTitle = ""
	state.ViewName = ""
	state.IssueEstimateMins = nil
	state.ViewMeta = ""
	state.ViewBody = ""
	state.ExportPresetKind = ""
	state.ExportPresetFormat = ""
	state.ExportPresetOutput = ""
	state.ExportIncludePDF = false
	return state
}

func SyncDialogFocus(state State) State {
	for i := range state.Inputs {
		state.Inputs[i].Blur()
	}
	if state.DescriptionEnabled {
		state.Description.Blur()
	}
	if state.DescriptionEnabled && state.FocusIdx == state.DescriptionIndex {
		state.Description.Focus()
		return state
	}
	if inputIdx, ok := dialogInputIndex(state, state.FocusIdx); ok {
		state.Inputs[inputIdx].Focus()
	}
	return state
}

func clearDialogError(state State) State {
	state.ErrorMessage = ""
	return state
}

func dialogInputIndex(state State, focusIdx int) (int, bool) {
	if focusIdx < 0 {
		return 0, false
	}
	if state.DescriptionEnabled && focusIdx == state.DescriptionIndex {
		return 0, false
	}
	inputIdx := focusIdx
	if state.DescriptionEnabled && focusIdx > state.DescriptionIndex {
		inputIdx--
	}
	if inputIdx < 0 || inputIdx >= len(state.Inputs) {
		return 0, false
	}
	return inputIdx, true
}

func ToggleEndSessionAdvanced(state State) State {
	if state.Kind != "end_session" {
		return state
	}
	if len(state.Inputs) > 1 {
		commit := state.Inputs[0].Value()
		input := newSessionDetailInput("Commit message")
		input.SetValue(commit)
		input.Focus()
		state.Inputs = []textinput.Model{input}
		state.FocusIdx = 0
		return state
	}
	commit := state.Inputs[0].Value()
	inputs := []textinput.Model{
		newSessionDetailInput("Commit message"),
		newSessionDetailInput("Worked on"),
		newSessionDetailInput("Outcome"),
		newSessionDetailInput("Next step"),
		newSessionDetailInput("Blockers"),
		newSessionDetailInput("Links"),
	}
	inputs[0].SetValue(commit)
	inputs[0].Focus()
	state.Inputs = inputs
	state.FocusIdx = 0
	return state
}

func Update(state State, ctx UpdateContext, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	switch state.Kind {
	case "create_repo":
		return updateNameDescription(state, msg, "Repo name is required", func(name string, description *string) *Action {
			return &Action{Kind: "create_repo", Name: name, Description: description}
		})
	case "edit_repo":
		return updateNameDescription(state, msg, "Repo name is required", func(name string, description *string) *Action {
			return &Action{Kind: "edit_repo", RepoID: state.RepoID, Name: name, Description: description}
		})
	case "create_stream":
		return updateNameDescription(state, msg, "Stream name is required", func(name string, description *string) *Action {
			return &Action{Kind: "create_stream", RepoID: state.RepoID, Name: name, Description: description}
		})
	case "edit_stream":
		return updateNameDescription(state, msg, "Stream name is required", func(name string, description *string) *Action {
			return &Action{Kind: "edit_stream", RepoID: state.RepoID, StreamID: state.StreamID, Name: name, Description: description}
		})
	case "create_scratchpad":
		return updateCreateScratchpad(state, msg)
	case "confirm_delete":
		return updateConfirmDelete(state, msg)
	case "confirm_wipe":
		return updateConfirmWipe(state, msg)
	case "confirm_uninstall":
		return updateConfirmUninstall(state, msg)
	case "stash_list":
		return updateStashList(state, ctx, msg)
	case "issue_status":
		return updateIssueStatus(state, ctx, currentDate, msg)
	case "issue_status_note":
		return updateIssueStatusNote(state, currentDate, msg)
	case "end_session", "stash_session":
		return updateSessionMessage(state, ctx, currentDate, msg)
	case "amend_session":
		return updateAmendSession(state, msg)
	case "manual_session":
		return updateManualSession(state, msg)
	case "issue_session_transition":
		return updateIssueSessionTransition(state, ctx, currentDate, msg)
	case "pick_date":
		return updateDatePicker(state, ctx, currentDate, msg)
	case "create_issue_meta":
		return updateCreateIssueMeta(state, currentDate, msg)
	case "create_issue_default":
		return updateCreateIssueDefault(state, ctx, currentDate, msg)
	case "create_habit":
		return updateCreateHabit(state, ctx, msg)
	case "edit_habit":
		return updateHabitEditor(state, msg, "edit_habit")
	case "complete_habit":
		return updateHabitCompletion(state, msg)
	case "checkout_context":
		return updateCheckoutContext(state, ctx, msg)
	case "edit_issue":
		return updateEditIssue(state, currentDate, msg)
	case "create_checkin", "edit_checkin":
		return updateCheckIn(state, msg)
	case "export_report":
		return updateExportDaily(state, msg)
	case "export_report_category":
		return updateExportCategory(state, msg)
	case "export_preset":
		return updateExportPreset(state, msg)
	case "export_calendar_repo":
		return updateExportCalendarRepo(state, msg)
	case "edit_export_reports_dir":
		return updateSingleInput(state, msg, "Reports directory is required", func(value string) *Action {
			return &Action{Kind: "set_export_reports_dir", Path: value}
		})
	case "edit_export_ics_dir":
		return updateSingleInput(state, msg, "ICS export directory is required", func(value string) *Action {
			return &Action{Kind: "set_export_ics_dir", Path: value}
		})
	case "edit_rest_protection":
		return updateRestProtection(state, currentDate, msg)
	case "view_entity":
		return updateViewEntity(state, msg)
	case "support_bundle_result":
		return updateSupportBundleResult(state, msg)
	default:
		return state, nil, ""
	}
}

func newSessionDetailInput(placeholder string) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
	input.CharLimit = 200
	input.Width = 48
	return input
}

func updateSingleInput(state State, msg tea.KeyMsg, requiredMsg string, submit func(string) *Action) (State, *Action, string) {
	if len(state.Inputs) == 0 {
		return Close(state), nil, "dialog input unavailable"
	}
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			name := strings.TrimSpace(state.Inputs[0].Value())
			if name == "" {
				return state, nil, requiredMsg
			}
			return Close(state), submit(name), ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[0], cmd = state.Inputs[0].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateValidatedSingleInput(state State, msg tea.KeyMsg, requiredMsg string, submit func(string) (*Action, string)) (State, *Action, string) {
	if len(state.Inputs) == 0 {
		return Close(state), nil, "dialog input unavailable"
	}
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			value := strings.TrimSpace(state.Inputs[0].Value())
			if requiredMsg != "" && value == "" {
				return state, nil, requiredMsg
			}
			action, status := submit(value)
			if status != "" {
				return state, nil, status
			}
			return Close(state), action, ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[0], cmd = state.Inputs[0].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateNameDescription(state State, msg tea.KeyMsg, requiredMsg string, submit func(string, *string) *Action) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		dir := 1
		if msg.String() == "shift+tab" || msg.String() == "up" {
			dir = -1
		}
		state.FocusIdx = (state.FocusIdx + dir + 2) % 2
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			name := strings.TrimSpace(state.Inputs[0].Value())
			if name == "" {
				return state, nil, requiredMsg
			}
			return Close(state), submit(name, ValueToPointer(strings.TrimSpace(state.Description.Value()))), ""
		}
	}
	if state.DescriptionEnabled && state.FocusIdx == state.DescriptionIndex {
		var cmd tea.Cmd
		state.Description, cmd = state.Description.Update(msg)
		_ = cmd
		return clearDialogError(state), nil, ""
	}
	var cmd tea.Cmd
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateCreateScratchpad(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		dir := 1
		if msg.String() == "shift+tab" || msg.String() == "up" {
			dir = -1
		}
		state.FocusIdx = (state.FocusIdx + dir + len(state.Inputs)) % len(state.Inputs)
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			name := strings.TrimSpace(state.Inputs[0].Value())
			path := strings.TrimSpace(state.Inputs[1].Value())
			if name == "" || path == "" {
				return state, nil, "Name and path are required"
			}
			return Close(state), &Action{Kind: "create_scratchpad", Name: name, Path: path}, ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateCheckIn(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		dir := 1
		if msg.String() == "shift+tab" || msg.String() == "up" {
			dir = -1
		}
		state.FocusIdx = (state.FocusIdx + dir + len(state.Inputs)) % len(state.Inputs)
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			mood, err := strconv.Atoi(strings.TrimSpace(state.Inputs[0].Value()))
			if err != nil || mood < 1 || mood > 5 {
				return state, nil, "Mood must be between 1 and 5"
			}
			energy, err := strconv.Atoi(strings.TrimSpace(state.Inputs[1].Value()))
			if err != nil || energy < 1 || energy > 5 {
				return state, nil, "Energy must be between 1 and 5"
			}
			sleepHours, status := parseOptionalFloat(strings.TrimSpace(state.Inputs[2].Value()))
			if status != "" {
				return state, nil, status
			}
			sleepScore, status := parseOptionalIntRange(strings.TrimSpace(state.Inputs[3].Value()), 0, 100, "Sleep score must be between 0 and 100")
			if status != "" {
				return state, nil, status
			}
			screenTime, status := parseOptionalIntRange(strings.TrimSpace(state.Inputs[4].Value()), 0, 100000, "Screen time must be 0 or more")
			if status != "" {
				return state, nil, status
			}
			note := ValueToPointer(strings.TrimSpace(state.Inputs[5].Value()))
			return Close(state), &Action{
				Kind:              state.Kind,
				CheckInDate:       state.CheckInDate,
				Mood:              mood,
				Energy:            energy,
				SleepHours:        sleepHours,
				SleepScore:        sleepScore,
				ScreenTimeMinutes: screenTime,
				Note:              note,
			}, ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func parseOptionalFloat(raw string) (*float64, string) {
	if raw == "" {
		return nil, ""
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value < 0 {
		return nil, "Sleep hours must be 0 or more"
	}
	return &value, ""
}

func parseOptionalIntRange(raw string, min int, max int, message string) (*int, string) {
	if raw == "" {
		return nil, ""
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		return nil, message
	}
	return &value, ""
}

func updateConfirmDelete(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "enter":
		action := &Action{Kind: "delete", ID: state.DeleteID, RepoID: state.RepoID, StreamID: state.StreamID}
		action.Name = state.DeleteKind
		action.Title = state.DeleteLabel
		return Close(state), action, ""
	}
	return state, nil, ""
}

func updateConfirmWipe(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "enter":
		return Close(state), &Action{Kind: "wipe_runtime_data"}, ""
	}
	return state, nil, ""
}

func updateConfirmUninstall(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "enter":
		return Close(state), &Action{Kind: "uninstall_crona"}, ""
	}
	return state, nil, ""
}

func updateStashList(state State, ctx UpdateContext, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc", "q":
		return Close(state), nil, ""
	case "j", "down":
		if state.StashCursor < len(ctx.Stashes)-1 {
			state.StashCursor++
		}
	case "k", "up":
		if state.StashCursor > 0 {
			state.StashCursor--
		}
	case "enter":
		if len(ctx.Stashes) == 0 || state.StashCursor < 0 || state.StashCursor >= len(ctx.Stashes) {
			return state, nil, ""
		}
		return Close(state), &Action{Kind: "apply_stash", ID: ctx.Stashes[state.StashCursor].ID}, ""
	case "x":
		if len(ctx.Stashes) == 0 || state.StashCursor < 0 || state.StashCursor >= len(ctx.Stashes) {
			return state, nil, ""
		}
		return Close(state), &Action{Kind: "drop_stash", ID: ctx.Stashes[state.StashCursor].ID}, ""
	}
	return state, nil, ""
}

func updateIssueStatus(state State, ctx UpdateContext, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc", "q":
		return Close(state), nil, ""
	case "j", "down":
		if state.StatusCursor < len(state.StatusItems)-1 {
			state.StatusCursor++
		}
	case "k", "up":
		if state.StatusCursor > 0 {
			state.StatusCursor--
		}
	case "enter":
		if !ctx.HasSelectedIssue || len(state.StatusItems) == 0 || state.StatusCursor < 0 || state.StatusCursor >= len(state.StatusItems) {
			return state, nil, ""
		}
		status := string(state.StatusItems[state.StatusCursor])
		switch status {
		case "blocked":
			return OpenIssueStatusNote(state, ctx.SelectedIssueID, ctx.SelectedStreamID, status, "Blocker reason", true), nil, ""
		case "in_review":
			return OpenIssueStatusNote(state, ctx.SelectedIssueID, ctx.SelectedStreamID, status, "Review note (optional)", false), nil, ""
		case "done":
			return OpenIssueStatusNote(state, ctx.SelectedIssueID, ctx.SelectedStreamID, status, "Completion note (optional)", false), nil, ""
		case "abandoned":
			return OpenIssueStatusNote(state, ctx.SelectedIssueID, ctx.SelectedStreamID, status, "Abandon reason", true), nil, ""
		default:
			return Close(state), &Action{Kind: "change_issue_status", IssueID: ctx.SelectedIssueID, StreamID: ctx.SelectedStreamID, Status: status}, ""
		}
	}
	return state, nil, ""
}

func updateIssueStatusNote(state State, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			note := ValueToPointer(state.Inputs[0].Value())
			if state.StatusRequired && note == nil {
				return state, nil, state.StatusLabel + " is required"
			}
			return Close(state), &Action{Kind: "change_issue_status", IssueID: state.IssueID, StreamID: state.StreamID, Status: state.IssueStatus, Note: note}, ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[0], cmd = state.Inputs[0].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateSessionMessage(state State, ctx UpdateContext, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "ctrl+e", "f2":
		return ToggleEndSessionAdvanced(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		if len(state.Inputs) == 1 {
			return state, nil, ""
		}
		dir := 1
		if msg.String() == "shift+tab" || msg.String() == "up" {
			dir = -1
		}
		state.FocusIdx = (state.FocusIdx + dir + len(state.Inputs)) % len(state.Inputs)
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			payload := EndSessionRequest(state.Inputs)
			kind := state.Kind
			state = Close(state)
			if kind == "end_session" {
				if !ctx.HasActiveIssue {
					return state, nil, "Active issue metadata unavailable"
				}
				return state, &Action{Kind: "end_session", StreamID: ctx.ActiveIssueStream, Payload: payload}, ""
			}
			return state, &Action{Kind: "stash_session", Note: payload.CommitMessage}, ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateIssueSessionTransition(state State, ctx UpdateContext, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "n", "N":
		if state.IssueStatus != "done" && state.IssueStatus != "abandoned" {
			return Close(state), nil, ""
		}
	case "y", "Y":
		if state.IssueStatus != "done" && state.IssueStatus != "abandoned" {
			return Close(state), &Action{Kind: "change_issue_status_and_end_session", IssueID: state.IssueID, StreamID: ctx.ActiveIssueStream, Status: state.IssueStatus, Payload: shareddto.EndSessionRequest{}}, ""
		}
	default:
		if isDialogSubmitKey(state, msg.String()) {
			note := ValueToPointer("")
			if len(state.Inputs) > 0 {
				note = ValueToPointer(state.Inputs[0].Value())
			}
			if state.IssueStatus == "abandoned" && note == nil {
				return state, nil, "Abandon reason is required"
			}
			return Close(state), &Action{Kind: "change_issue_status_and_end_session", IssueID: state.IssueID, StreamID: ctx.ActiveIssueStream, Status: state.IssueStatus, Note: note, Payload: shareddto.EndSessionRequest{CommitMessage: note}}, ""
		}
	}
	if (state.IssueStatus == "done" || state.IssueStatus == "abandoned") && len(state.Inputs) > 0 {
		var cmd tea.Cmd
		state.Inputs[0], cmd = state.Inputs[0].Update(msg)
		_ = cmd
	}
	return clearDialogError(state), nil, ""
}

func updateAmendSession(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			note := strings.TrimSpace(state.Inputs[0].Value())
			if note == "" {
				return state, nil, "Commit message is required"
			}
			return Close(state), &Action{Kind: "amend_session", ID: state.SessionID, Note: ValueToPointer(note)}, ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[0], cmd = state.Inputs[0].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateManualSession(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "f2", "ctrl+y":
		if state.FocusIdx == 1 {
			return OpenDatePicker(state, "manual_session", state.IssueID, 1, ValueToPointer(state.Inputs[1].Value()), strings.TrimSpace(state.Inputs[1].Value())), nil, ""
		}
	case "g":
		if state.FocusIdx == 1 {
			state.Inputs[1].SetValue(time.Now().Format("2006-01-02"))
			return state, nil, ""
		}
	case "tab", "shift+tab", "down", "up":
		state.FocusIdx = (state.FocusIdx + ternaryDir(msg.String()) + len(state.Inputs)) % len(state.Inputs)
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			date := strings.TrimSpace(state.Inputs[1].Value())
			if _, err := time.Parse("2006-01-02", date); err != nil {
				return state, nil, "Date must be YYYY-MM-DD"
			}
			workSeconds, err := ParseDurationInput(state.Inputs[2].Value(), true, "Work duration")
			if err != nil {
				return state, nil, err.Error()
			}
			breakSeconds, err := ParseDurationInput(state.Inputs[3].Value(), false, "Break duration")
			if err != nil {
				return state, nil, err.Error()
			}
			startTime, err := ParseClockInput(state.Inputs[4].Value())
			if err != nil {
				return state, nil, err.Error()
			}
			endTime, err := ParseClockInput(state.Inputs[5].Value())
			if err != nil {
				return state, nil, err.Error()
			}
			req := shareddto.ManualSessionLogRequest{
				IssueID:              state.IssueID,
				Date:                 date,
				WorkDurationSeconds:  workSeconds,
				BreakDurationSeconds: breakSeconds,
				StartTime:            startTime,
				EndTime:              endTime,
				CommitMessage:        ValueToPointer(state.Inputs[0].Value()),
				Notes:                ValueToPointer(state.Inputs[6].Value()),
			}
			return Close(state), &Action{Kind: "manual_session", IssueID: state.IssueID, ManualSession: &req}, ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateCreateIssueMeta(state State, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	if state.FocusIdx == 3 && (msg.String() == "f2" || msg.String() == "ctrl+y") {
		return OpenDatePicker(state, "create_issue_meta", 0, 3, ValueToPointer(state.Inputs[2].Value()), currentDate), nil, ""
	}
	return updateMultiInputIssue(state, msg, 4, func(state State) (*Action, string) {
		title := strings.TrimSpace(state.Inputs[0].Value())
		if title == "" {
			return nil, "Issue title is required"
		}
		description := ValueToPointer(strings.TrimSpace(state.Description.Value()))
		estimate, err := ParseEstimateInput(state.Inputs[1].Value())
		if err != nil {
			return nil, err.Error()
		}
		dueDate, err := ParseDueDateInput(state.Inputs[2].Value())
		if err != nil {
			return nil, err.Error()
		}
		return &Action{Kind: "create_issue_meta", StreamID: state.StreamID, Title: title, Description: description, Estimate: estimate, DueDate: dueDate}, ""
	})
}

func updateCreateIssueDefault(state State, ctx UpdateContext, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "f2", "ctrl+y":
		if state.FocusIdx == 5 {
			return OpenDatePicker(state, "create_issue_default", 0, 5, ValueToPointer(state.Inputs[4].Value()), currentDate), nil, ""
		}
	case "g":
		if state.FocusIdx == 5 {
			state.Inputs[4].SetValue(currentDate)
			return state, nil, ""
		}
	case "tab", "shift+tab", "down", "up":
		if (msg.String() == "down" || msg.String() == "up") && (state.FocusIdx == 0 || state.FocusIdx == 1) {
			if state.FocusIdx == 0 {
				state.RepoIndex = ShiftSelection(state.RepoIndex, len(DefaultRepoOptions(state.Inputs, ctx.Repos)), ternaryDir(msg.String()))
				state.StreamIndex = 0
			} else {
				state.StreamIndex = ShiftSelection(state.StreamIndex, len(DefaultStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)), ternaryDir(msg.String()))
			}
			return state, nil, ""
		}
		state.FocusIdx = (state.FocusIdx + ternaryDir(msg.String()) + 6) % 6
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	case "left":
		if state.FocusIdx == 0 {
			state.RepoIndex = ShiftSelection(state.RepoIndex, len(DefaultRepoOptions(state.Inputs, ctx.Repos)), -1)
			state.StreamIndex = 0
			return clearDialogError(state), nil, ""
		}
		if state.FocusIdx == 1 {
			state.StreamIndex = ShiftSelection(state.StreamIndex, len(DefaultStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)), -1)
			return clearDialogError(state), nil, ""
		}
	case "right":
		if state.FocusIdx == 0 {
			state.RepoIndex = ShiftSelection(state.RepoIndex, len(DefaultRepoOptions(state.Inputs, ctx.Repos)), 1)
			state.StreamIndex = 0
			return clearDialogError(state), nil, ""
		}
		if state.FocusIdx == 1 {
			state.StreamIndex = ShiftSelection(state.StreamIndex, len(DefaultStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)), 1)
			return clearDialogError(state), nil, ""
		}
	}
	if isDialogSubmitKey(state, msg.String()) {
		repoName, streamName := DefaultIssueDialogNames(state.Inputs, state.RepoIndex, state.StreamIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)
		title := strings.TrimSpace(state.Inputs[2].Value())
		if repoName == "" || streamName == "" || title == "" {
			return state, nil, "Repo, stream, and issue title are required"
		}
		description := ValueToPointer(strings.TrimSpace(state.Description.Value()))
		estimate, err := ParseEstimateInput(state.Inputs[3].Value())
		if err != nil {
			return state, nil, err.Error()
		}
		dueDate, err := ParseDueDateInput(state.Inputs[4].Value())
		if err != nil {
			return state, nil, err.Error()
		}
		return Close(state), &Action{Kind: "create_issue_default", RepoName: repoName, StreamName: streamName, Title: title, Description: description, Estimate: estimate, DueDate: dueDate}, ""
	}
	if state.DescriptionEnabled && state.FocusIdx == state.DescriptionIndex {
		var cmd tea.Cmd
		state.Description, cmd = state.Description.Update(msg)
		_ = cmd
		return clearDialogError(state), nil, ""
	}
	inputIdx, ok := dialogInputIndex(state, state.FocusIdx)
	if !ok {
		return state, nil, ""
	}
	var cmd tea.Cmd
	state.Inputs[inputIdx], cmd = state.Inputs[inputIdx].Update(msg)
	_ = cmd
	if inputIdx == 0 {
		state.RepoIndex = 0
		state.StreamIndex = 0
	}
	if inputIdx == 1 {
		state.StreamIndex = 0
	}
	return clearDialogError(state), nil, ""
}

func updateCheckoutContext(state State, ctx UpdateContext, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "c":
		return Close(state), &Action{Kind: "checkout_context"}, ""
	case "tab", "shift+tab", "down", "up":
		state.FocusIdx = (state.FocusIdx + ternaryDir(msg.String()) + 2) % 2
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	case "left":
		if state.FocusIdx == 0 {
			state.RepoIndex = ShiftSelection(state.RepoIndex, len(CheckoutRepoOptions(state.Inputs, ctx.Repos)), -1)
			state.StreamIndex = 0
			return state, nil, ""
		}
		if state.FocusIdx == 1 {
			state.StreamIndex = ShiftSelection(state.StreamIndex, len(CheckoutStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)), -1)
			return state, nil, ""
		}
	case "right":
		if state.FocusIdx == 0 {
			state.RepoIndex = ShiftSelection(state.RepoIndex, len(CheckoutRepoOptions(state.Inputs, ctx.Repos)), 1)
			state.StreamIndex = 0
			return state, nil, ""
		}
		if state.FocusIdx == 1 {
			state.StreamIndex = ShiftSelection(state.StreamIndex, len(CheckoutStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)), 1)
			return state, nil, ""
		}
	case "enter":
		repoID, repoName, streamID, streamName := CheckoutDialogSelection(state.Inputs, state.RepoIndex, state.StreamIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)
		if strings.TrimSpace(repoName) == "" && streamID == nil && strings.TrimSpace(streamName) == "" {
			return Close(state), &Action{Kind: "checkout_context"}, ""
		}
		if strings.TrimSpace(repoName) == "" {
			return state, nil, "Repo is required"
		}
		if streamID != nil {
			return Close(state), &Action{Kind: "checkout_context", RepoID: repoID, RepoName: repoName, StreamID: *streamID, StreamName: streamName}, ""
		}
		if strings.TrimSpace(streamName) == "" {
			return Close(state), &Action{Kind: "checkout_context", RepoID: repoID, RepoName: repoName}, ""
		}
		return Close(state), &Action{Kind: "checkout_context", RepoID: repoID, RepoName: repoName, StreamName: streamName}, ""
	}
	var cmd tea.Cmd
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	if state.FocusIdx == 0 {
		state.RepoIndex = 0
		state.StreamIndex = 0
	}
	if state.FocusIdx == 1 {
		state.StreamIndex = 0
	}
	return clearDialogError(state), nil, ""
}

func updateEditIssue(state State, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	if state.FocusIdx == 3 && (msg.String() == "f2" || msg.String() == "ctrl+y") {
		return OpenDatePicker(state, "edit_issue", state.IssueID, 3, ValueToPointer(state.Inputs[2].Value()), currentDate), nil, ""
	}
	return updateMultiInputIssue(state, msg, 4, func(state State) (*Action, string) {
		title := strings.TrimSpace(state.Inputs[0].Value())
		if title == "" {
			return nil, "Issue title is required"
		}
		description := ValueToPointer(strings.TrimSpace(state.Description.Value()))
		estimate, err := ParseEstimateInput(state.Inputs[1].Value())
		if err != nil {
			return nil, err.Error()
		}
		dueDate, err := ParseDueDateInput(state.Inputs[2].Value())
		if err != nil {
			return nil, err.Error()
		}
		return &Action{Kind: "edit_issue", IssueID: state.IssueID, StreamID: state.StreamID, Title: title, Description: description, Estimate: estimate, DueDate: dueDate}, ""
	})
}

func updateMultiInputIssue(state State, msg tea.KeyMsg, inputCount int, submit func(State) (*Action, string)) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		state.FocusIdx = (state.FocusIdx + ternaryDir(msg.String()) + inputCount) % inputCount
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			action, status := submit(state)
			if action == nil {
				return state, nil, status
			}
			return Close(state), action, status
		}
	}
	if state.DescriptionEnabled && state.FocusIdx == state.DescriptionIndex {
		var cmd tea.Cmd
		state.Description, cmd = state.Description.Update(msg)
		_ = cmd
		return clearDialogError(state), nil, ""
	}
	inputIdx, ok := dialogInputIndex(state, state.FocusIdx)
	if !ok {
		return clearDialogError(state), nil, ""
	}
	var cmd tea.Cmd
	state.Inputs[inputIdx], cmd = state.Inputs[inputIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateExportCategory(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc", "q":
		if state.Processing {
			return state, nil, ""
		}
		return Close(state), nil, ""
	case "j", "down":
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
	case "k", "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
	case "enter":
		if state.ChoiceCursor < 0 || state.ChoiceCursor >= len(state.ChoiceValues) {
			return state, nil, "Select an export category"
		}
		return OpenExportReportChoices(state, state.ChoiceValues[state.ChoiceCursor]), nil, ""
	}
	return clearDialogError(state), nil, ""
}

func updateExportDaily(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc", "q":
		if state.Processing {
			return state, nil, ""
		}
		items, values, details := exportReportCategories()
		state.Kind = "export_report_category"
		state.Parent = ""
		state.ExportCategory = ""
		state.ChoiceItems = items
		state.ChoiceValues = values
		state.ChoiceDetails = details
		state.ChoiceCursor = 0
		return clearDialogError(state), nil, ""
	case "j", "down":
		if state.Processing {
			return state, nil, ""
		}
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
	case "k", "up":
		if state.Processing {
			return state, nil, ""
		}
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
	case "enter":
		if state.Processing {
			return state, nil, ""
		}
		selected := ""
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceItems) {
			selected = state.ChoiceItems[state.ChoiceCursor]
		}
		action := Action{Kind: "export_report", CheckInDate: state.CheckInDate}
		switch selected {
		case "Daily report: write PDF file":
			return OpenExportPreset(state, sharedtypes.ExportReportKindDaily, sharedtypes.ExportFormatPDF, sharedtypes.ExportOutputModeFile), nil, ""
		case "Weekly summary: write PDF file":
			return OpenExportPreset(state, sharedtypes.ExportReportKindWeekly, sharedtypes.ExportFormatPDF, sharedtypes.ExportOutputModeFile), nil, ""
		case "Daily report: write Markdown file":
			return OpenExportPreset(state, sharedtypes.ExportReportKindDaily, sharedtypes.ExportFormatMarkdown, sharedtypes.ExportOutputModeFile), nil, ""
		case "Daily report: copy to clipboard":
			return OpenExportPreset(state, sharedtypes.ExportReportKindDaily, sharedtypes.ExportFormatMarkdown, sharedtypes.ExportOutputModeClipboard), nil, ""
		case "Weekly summary: write Markdown file":
			return OpenExportPreset(state, sharedtypes.ExportReportKindWeekly, sharedtypes.ExportFormatMarkdown, sharedtypes.ExportOutputModeFile), nil, ""
		case "Repo report: write Markdown file":
			action.ReportKind = sharedtypes.ExportReportKindRepo
			action.ReportFormat = sharedtypes.ExportFormatMarkdown
			action.OutputMode = sharedtypes.ExportOutputModeFile
			state.ProcessingLabel = "Generating repo report..."
		case "Stream report: write Markdown file":
			action.ReportKind = sharedtypes.ExportReportKindStream
			action.ReportFormat = sharedtypes.ExportFormatMarkdown
			action.OutputMode = sharedtypes.ExportOutputModeFile
			state.ProcessingLabel = "Generating stream report..."
		case "Issue rollup: write Markdown file":
			action.ReportKind = sharedtypes.ExportReportKindIssueRollup
			action.ReportFormat = sharedtypes.ExportFormatMarkdown
			action.OutputMode = sharedtypes.ExportOutputModeFile
			state.ProcessingLabel = "Generating issue rollup..."
		case "Calendar export: write ICS file":
			if len(state.RepoItems) == 0 {
				return state, nil, "Calendar export requires at least one repo"
			}
			return OpenExportCalendarRepo(state), nil, ""
		default:
			action.ReportKind = sharedtypes.ExportReportKindCSV
			action.ReportFormat = sharedtypes.ExportFormatCSV
			action.OutputMode = sharedtypes.ExportOutputModeFile
			state.ProcessingLabel = "Generating CSV export..."
		}
		state.Processing = true
		return state, &action, ""
	}
	return state, nil, ""
}

func updateExportPreset(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc", "q":
		state.Kind = "export_report"
		state.Parent = ""
		state.ChoiceItems, state.ChoiceDetails = exportReportChoices(state.ExportCategory, state.ExportIncludePDF)
		state.ChoiceValues = nil
		state.ExportPresetKind = ""
		state.ExportPresetFormat = ""
		state.ExportPresetOutput = ""
		state.ChoiceCursor = 0
		return clearDialogError(state), nil, ""
	case "j", "down":
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
	case "k", "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
	case "enter":
		if state.ChoiceCursor < 0 || state.ChoiceCursor >= len(state.ChoiceValues) {
			return state, nil, "Select a report style"
		}
		presetID := state.ChoiceValues[state.ChoiceCursor]
		if presetID == "__current__" {
			presetID = ""
		}
		state.Processing = true
		switch state.ExportPresetKind {
		case sharedtypes.ExportReportKindDaily:
			if state.ExportPresetOutput == sharedtypes.ExportOutputModeClipboard {
				state.ProcessingLabel = "Copying report..."
			} else {
				state.ProcessingLabel = "Generating daily report..."
			}
		case sharedtypes.ExportReportKindWeekly:
			state.ProcessingLabel = "Generating weekly report..."
		default:
			state.ProcessingLabel = "Generating report..."
		}
		return state, &Action{
			Kind:         "export_report",
			ReportKind:   state.ExportPresetKind,
			ReportFormat: state.ExportPresetFormat,
			OutputMode:   state.ExportPresetOutput,
			CheckInDate:  state.CheckInDate,
			PresetID:     presetID,
		}, ""
	}
	return clearDialogError(state), nil, ""
}

func updateExportCalendarRepo(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc", "q":
		state.Kind = "export_report"
		state.Parent = ""
		state.ChoiceItems, state.ChoiceDetails = exportReportChoices(state.ExportCategory, state.ExportIncludePDF)
		return clearDialogError(state), nil, ""
	case "j", "down":
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
	case "k", "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
	case "enter":
		if len(state.RepoItemIDs) == 0 || state.ChoiceCursor < 0 || state.ChoiceCursor >= len(state.RepoItemIDs) {
			return state, nil, "Calendar export requires a repo"
		}
		state.RepoIndex = state.ChoiceCursor
		state.RepoID = state.RepoItemIDs[state.ChoiceCursor]
		state.RepoName = state.RepoItems[state.ChoiceCursor]
		state.Kind = "export_report"
		state.ChoiceItems = nil
		state.ChoiceCursor = 0
		state.Parent = ""
		state.Processing = true
		state.ProcessingLabel = "Generating calendar export..."
		return state, &Action{
			Kind:         "export_report",
			ReportKind:   sharedtypes.ExportReportKindCalendar,
			ReportFormat: sharedtypes.ExportFormatICS,
			OutputMode:   sharedtypes.ExportOutputModeFile,
			CheckInDate:  state.CheckInDate,
			RepoID:       state.RepoID,
			RepoName:     state.RepoName,
		}, ""
	}
	return clearDialogError(state), nil, ""
}

func updateHabitEditor(state State, msg tea.KeyMsg, kind string) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		state.FocusIdx = (state.FocusIdx + ternaryDir(msg.String()) + 4) % 4
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	}
	if isDialogSubmitKey(state, msg.String()) {
		name := strings.TrimSpace(state.Inputs[0].Value())
		if name == "" {
			return state, nil, "Habit name is required"
		}
		scheduleType, weekdays, err := ParseHabitSchedule(state.Inputs[1].Value())
		if err != nil {
			return state, nil, err.Error()
		}
		target, status := parseOptionalIntRange(strings.TrimSpace(state.Inputs[2].Value()), 0, 100000, "Target minutes must be 0 or more")
		if status != "" {
			return state, nil, status
		}
		action := &Action{
			Kind:        kind,
			HabitID:     state.HabitID,
			StreamID:    state.StreamID,
			Name:        name,
			Description: ValueToPointer(strings.TrimSpace(state.Description.Value())),
			Status:      scheduleType,
			Weekdays:    weekdays,
			Active:      state.StatusLabel != "inactive",
			Estimate:    target,
		}
		return Close(state), action, ""
	}
	if state.DescriptionEnabled && state.FocusIdx == state.DescriptionIndex {
		var cmd tea.Cmd
		state.Description, cmd = state.Description.Update(msg)
		_ = cmd
		return clearDialogError(state), nil, ""
	}
	inputIdx, ok := dialogInputIndex(state, state.FocusIdx)
	if !ok {
		return state, nil, ""
	}
	var cmd tea.Cmd
	state.Inputs[inputIdx], cmd = state.Inputs[inputIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateCreateHabit(state State, ctx UpdateContext, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		if (msg.String() == "down" || msg.String() == "up") && (state.FocusIdx == 0 || state.FocusIdx == 1) {
			if state.FocusIdx == 0 {
				state.RepoIndex = ShiftSelection(state.RepoIndex, len(DefaultRepoOptions(state.Inputs, ctx.Repos)), ternaryDir(msg.String()))
				state.StreamIndex = 0
			} else {
				state.StreamIndex = ShiftSelection(state.StreamIndex, len(DefaultStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)), ternaryDir(msg.String()))
			}
			return clearDialogError(state), nil, ""
		}
		state.FocusIdx = (state.FocusIdx + ternaryDir(msg.String()) + 6) % 6
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	case "left":
		if state.FocusIdx == 0 {
			state.RepoIndex = ShiftSelection(state.RepoIndex, len(DefaultRepoOptions(state.Inputs, ctx.Repos)), -1)
			state.StreamIndex = 0
			return clearDialogError(state), nil, ""
		}
		if state.FocusIdx == 1 {
			state.StreamIndex = ShiftSelection(state.StreamIndex, len(DefaultStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)), -1)
			return clearDialogError(state), nil, ""
		}
	case "right":
		if state.FocusIdx == 0 {
			state.RepoIndex = ShiftSelection(state.RepoIndex, len(DefaultRepoOptions(state.Inputs, ctx.Repos)), 1)
			state.StreamIndex = 0
			return clearDialogError(state), nil, ""
		}
		if state.FocusIdx == 1 {
			state.StreamIndex = ShiftSelection(state.StreamIndex, len(DefaultStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)), 1)
			return clearDialogError(state), nil, ""
		}
	}
	if isDialogSubmitKey(state, msg.String()) {
		repoName, streamName := DefaultIssueDialogNames(state.Inputs, state.RepoIndex, state.StreamIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)
		name := strings.TrimSpace(state.Inputs[2].Value())
		if repoName == "" || streamName == "" || name == "" {
			return state, nil, "Repo, stream, and habit name are required"
		}
		scheduleType, weekdays, err := ParseHabitSchedule(state.Inputs[3].Value())
		if err != nil {
			return state, nil, err.Error()
		}
		target, status := parseOptionalIntRange(strings.TrimSpace(state.Inputs[4].Value()), 0, 100000, "Target minutes must be 0 or more")
		if status != "" {
			return state, nil, status
		}
		return Close(state), &Action{
			Kind:        "create_habit",
			RepoName:    repoName,
			StreamName:  streamName,
			Name:        name,
			Description: ValueToPointer(strings.TrimSpace(state.Description.Value())),
			Status:      scheduleType,
			Weekdays:    weekdays,
			Estimate:    target,
		}, ""
	}
	if state.DescriptionEnabled && state.FocusIdx == state.DescriptionIndex {
		var cmd tea.Cmd
		state.Description, cmd = state.Description.Update(msg)
		_ = cmd
		return clearDialogError(state), nil, ""
	}
	inputIdx, ok := dialogInputIndex(state, state.FocusIdx)
	if !ok {
		return state, nil, ""
	}
	var cmd tea.Cmd
	state.Inputs[inputIdx], cmd = state.Inputs[inputIdx].Update(msg)
	_ = cmd
	if inputIdx == 0 {
		state.RepoIndex = 0
		state.StreamIndex = 0
	}
	if inputIdx == 1 {
		state.StreamIndex = 0
	}
	return clearDialogError(state), nil, ""
}

func updateHabitCompletion(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		state.FocusIdx = (state.FocusIdx + ternaryDir(msg.String()) + 2) % 2
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	}
	if isDialogSubmitKey(state, msg.String()) {
		duration, status := parseOptionalIntRange(strings.TrimSpace(state.Inputs[0].Value()), 0, 100000, "Duration minutes must be 0 or more")
		if status != "" {
			return state, nil, status
		}
		return Close(state), &Action{
			Kind:        "complete_habit",
			HabitID:     state.HabitID,
			CheckInDate: state.CheckInDate,
			Estimate:    duration,
			Note:        ValueToPointer(strings.TrimSpace(state.Description.Value())),
		}, ""
	}
	if state.DescriptionEnabled && state.FocusIdx == state.DescriptionIndex {
		var cmd tea.Cmd
		state.Description, cmd = state.Description.Update(msg)
		_ = cmd
		return clearDialogError(state), nil, ""
	}
	var cmd tea.Cmd
	state.Inputs[0], cmd = state.Inputs[0].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func ParseHabitSchedule(raw string) (string, []int, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	switch value {
	case "", "daily":
		return "daily", nil, nil
	case "weekdays":
		return "weekdays", nil, nil
	}
	parts := strings.Split(value, ",")
	weekdays := make([]int, 0, len(parts))
	for _, part := range parts {
		switch strings.TrimSpace(part) {
		case "sun":
			weekdays = append(weekdays, 0)
		case "mon":
			weekdays = append(weekdays, 1)
		case "tue":
			weekdays = append(weekdays, 2)
		case "wed":
			weekdays = append(weekdays, 3)
		case "thu":
			weekdays = append(weekdays, 4)
		case "fri":
			weekdays = append(weekdays, 5)
		case "sat":
			weekdays = append(weekdays, 6)
		default:
			return "", nil, errors.New("schedule must be daily, weekdays, or comma-separated weekdays like mon,wed,fri")
		}
	}
	if len(weekdays) == 0 {
		return "", nil, errors.New("schedule must be daily, weekdays, or comma-separated weekdays like mon,wed,fri")
	}
	return "weekly", weekdays, nil
}

func ParseStreakKinds(raw string) ([]sharedtypes.StreakKind, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	out := make([]sharedtypes.StreakKind, 0, len(parts))
	seen := map[sharedtypes.StreakKind]struct{}{}
	for _, part := range parts {
		var kind sharedtypes.StreakKind
		switch strings.TrimSpace(part) {
		case "focus", "focus_days":
			kind = sharedtypes.StreakKindFocusDays
		case "checkin", "checkins", "check-in", "check_in_days", "checkin_days":
			kind = sharedtypes.StreakKindCheckInDays
		default:
			return nil, errors.New("streaks must be comma-separated values from focus_days,checkin_days")
		}
		if _, ok := seen[kind]; ok {
			continue
		}
		seen[kind] = struct{}{}
		out = append(out, kind)
	}
	return out, nil
}

func ParseWeekdayList(raw string) ([]int, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	out := make([]int, 0, len(parts))
	seen := map[int]struct{}{}
	for _, part := range parts {
		var weekday int
		switch strings.TrimSpace(part) {
		case "sun":
			weekday = 0
		case "mon":
			weekday = 1
		case "tue":
			weekday = 2
		case "wed":
			weekday = 3
		case "thu":
			weekday = 4
		case "fri":
			weekday = 5
		case "sat":
			weekday = 6
		default:
			return nil, errors.New("weekdays must be comma-separated tokens like mon,wed,fri")
		}
		if _, ok := seen[weekday]; ok {
			continue
		}
		seen[weekday] = struct{}{}
		out = append(out, weekday)
	}
	return out, nil
}

func ParseSpecificDates(raw string) ([]string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		date := strings.TrimSpace(part)
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return nil, errors.New("rest dates must use YYYY-MM-DD")
		}
		if _, ok := seen[date]; ok {
			continue
		}
		seen[date] = struct{}{}
		out = append(out, date)
	}
	return out, nil
}

func ParseRecurringDates(raw string) ([]string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		date := strings.TrimSpace(part)
		if _, err := time.Parse("01-02", date); err != nil {
			return nil, errors.New("recurring rest dates must use MM-DD")
		}
		if _, ok := seen[date]; ok {
			continue
		}
		seen[date] = struct{}{}
		out = append(out, date)
	}
	return out, nil
}

func streakKindTokens(values []sharedtypes.StreakKind) []string {
	if len(values) == 0 {
		values = sharedtypes.AvailableStreakKinds()
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

func streakKindStrings(values []sharedtypes.StreakKind) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

func WeekdayTokens(days []int) []string {
	names := map[int]string{0: "sun", 1: "mon", 2: "tue", 3: "wed", 4: "thu", 5: "fri", 6: "sat"}
	out := make([]string, 0, len(days))
	for _, day := range days {
		if name, ok := names[day]; ok {
			out = append(out, name)
		}
	}
	return out
}

func updateDatePicker(state State, ctx UpdateContext, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		if state.Parent == "rollup_start" || state.Parent == "rollup_end" {
			return Close(state), nil, ""
		}
		return closeDatePicker(state), nil, ""
	case "enter", " ":
		selected := state.DateCursorValue
		if state.Parent == "create_issue_meta" || state.Parent == "create_issue_default" || state.Parent == "edit_issue" {
			if idx, ok := dialogInputIndex(state, state.FocusIdx); ok {
				state.Inputs[idx].SetValue(selected)
			}
			return closeDatePicker(state), nil, ""
		}
		if state.Parent == "edit_rest_protection" {
			state.ProtectionDates = normalizedDateList(append(state.ProtectionDates, selected))
			return closeDatePicker(state), nil, ""
		}
		if state.Parent == "rollup_start" {
			return Close(state), &Action{Kind: "set_rollup_start_date", DueDate: ValueToPointer(selected)}, ""
		}
		if state.Parent == "rollup_end" {
			return Close(state), &Action{Kind: "set_rollup_end_date", DueDate: ValueToPointer(selected)}, ""
		}
		return Close(state), &Action{Kind: "set_issue_todo_date", IssueID: state.IssueID, StreamID: ctx.ActiveIssueStream, DueDate: ValueToPointer(selected)}, ""
	case "backspace", "delete", "c":
		if state.Parent == "create_issue_meta" || state.Parent == "create_issue_default" || state.Parent == "edit_issue" {
			if idx, ok := dialogInputIndex(state, state.FocusIdx); ok {
				state.Inputs[idx].SetValue("")
			}
			return closeDatePicker(state), nil, ""
		}
		if state.Parent == "edit_rest_protection" {
			return closeDatePicker(state), nil, ""
		}
		if state.Parent == "rollup_start" || state.Parent == "rollup_end" {
			return Close(state), nil, ""
		}
		return Close(state), &Action{Kind: "set_issue_todo_date", IssueID: state.IssueID, StreamID: ctx.ActiveIssueStream, DueDate: ValueToPointer("")}, ""
	case "left", "h":
		return shiftDatePicker(state, 0, 0, -1), nil, ""
	case "right", "l":
		return shiftDatePicker(state, 0, 0, 1), nil, ""
	case "up", "k":
		return shiftDatePicker(state, 0, 0, -7), nil, ""
	case "down", "j":
		return shiftDatePicker(state, 0, 0, 7), nil, ""
	case ",":
		return shiftDatePicker(state, 0, -1, 0), nil, ""
	case ".":
		return shiftDatePicker(state, 0, 1, 0), nil, ""
	case "g":
		return OpenDatePicker(state, state.Parent, state.IssueID, state.FocusIdx, ValueToPointer(time.Now().Format("2006-01-02")), currentDate), nil, ""
	}
	return state, nil, ""
}

func ternaryDir(key string) int {
	if key == "shift+tab" || key == "up" {
		return -1
	}
	return 1
}

func closeDatePicker(state State) State {
	parent := state.Parent
	dateField := state.FocusIdx
	state.Kind = parent
	state.Parent = ""
	state.DateMonthValue = ""
	state.DateCursorValue = ""
	if parent == "create_issue_meta" || parent == "create_issue_default" || parent == "edit_issue" {
		state.FocusIdx = dateField
		state = SyncDialogFocus(state)
		return state
	}
	for i := range state.Inputs {
		if i == state.FocusIdx {
			state.Inputs[i].Focus()
		} else {
			state.Inputs[i].Blur()
		}
	}
	return state
}

func shiftDatePicker(state State, years, months, days int) State {
	selected := DialogDate(state, time.Now().Format("2006-01-02")).AddDate(years, months, days)
	monthStart := time.Date(selected.Year(), selected.Month(), 1, 0, 0, 0, 0, selected.Location())
	state.DateCursorValue = selected.Format("2006-01-02")
	state.DateMonthValue = monthStart.Format("2006-01-02")
	return state
}

func updateRestProtection(state State, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "right", "l":
		state.ProtectionStep = (state.ProtectionStep + 1) % 4
		state.ProtectionCursor = 0
		return state, nil, ""
	case "shift+tab", "left", "h":
		state.ProtectionStep = (state.ProtectionStep + 3) % 4
		state.ProtectionCursor = 0
		return state, nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			return Close(state), &Action{
				Kind:        "patch_rest_protection",
				StreakKinds: streakKindStrings(state.ProtectionStreaks),
				IntList:     normalizedWeekdays(state.ProtectionWeekdays),
				RestDates:   normalizedDateList(state.ProtectionDates),
			}, ""
		}
	}
	switch state.ProtectionStep {
	case 0:
		return updateRestProtectionStreaks(state, msg)
	case 1:
		return updateRestProtectionWeekdays(state, msg)
	case 2:
		return updateRestProtectionDates(state, currentDate, msg)
	case 3:
		return clearDialogError(state), nil, ""
	}
	return clearDialogError(state), nil, ""
}

func updateRestProtectionStreaks(state State, msg tea.KeyMsg) (State, *Action, string) {
	total := len(sharedtypes.AvailableStreakKinds())
	switch msg.String() {
	case "j", "down":
		state.ProtectionCursor = ShiftSelection(state.ProtectionCursor, total, 1)
	case "k", "up":
		state.ProtectionCursor = ShiftSelection(state.ProtectionCursor, total, -1)
	case " ":
		current := sharedtypes.AvailableStreakKinds()[state.ProtectionCursor]
		state.ProtectionStreaks = toggleStreakKind(state.ProtectionStreaks, current)
	case "x":
		current := sharedtypes.AvailableStreakKinds()[state.ProtectionCursor]
		state.ProtectionStreaks = toggleStreakKind(state.ProtectionStreaks, current)
	case "a":
		state.ProtectionStreaks = append([]sharedtypes.StreakKind(nil), sharedtypes.AvailableStreakKinds()...)
	case "c":
		state.ProtectionStreaks = nil
	case "enter":
		state.ProtectionStep = 1
		state.ProtectionCursor = 0
	}
	return state, nil, ""
}

func updateRestProtectionWeekdays(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "j", "down":
		state.ProtectionCursor = ShiftSelection(state.ProtectionCursor, 7, 1)
	case "k", "up":
		state.ProtectionCursor = ShiftSelection(state.ProtectionCursor, 7, -1)
	case " ", "x":
		state.ProtectionWeekdays = toggleWeekday(state.ProtectionWeekdays, state.ProtectionCursor)
	case "c":
		state.ProtectionWeekdays = nil
	case "enter":
		state.ProtectionStep = 2
		state.ProtectionCursor = 0
	}
	return state, nil, ""
}

func updateRestProtectionDates(state State, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "j", "down":
		state.ProtectionCursor = ShiftSelection(state.ProtectionCursor, max(1, len(state.ProtectionDates)), 1)
	case "k", "up":
		state.ProtectionCursor = ShiftSelection(state.ProtectionCursor, max(1, len(state.ProtectionDates)), -1)
	case "a":
		initial := currentDate
		if strings.TrimSpace(initial) == "" {
			initial = time.Now().Format("2006-01-02")
		}
		return OpenDatePicker(state, "edit_rest_protection", 0, 0, ValueToPointer(initial), currentDate), nil, ""
	case "d", "backspace", "delete":
		if len(state.ProtectionDates) == 0 {
			return state, nil, ""
		}
		idx := state.ProtectionCursor
		if idx >= len(state.ProtectionDates) {
			idx = len(state.ProtectionDates) - 1
		}
		state.ProtectionDates = append(state.ProtectionDates[:idx], state.ProtectionDates[idx+1:]...)
		if state.ProtectionCursor >= len(state.ProtectionDates) && state.ProtectionCursor > 0 {
			state.ProtectionCursor--
		}
	case "enter":
		state.ProtectionStep = 3
		state.ProtectionCursor = 0
	}
	return state, nil, ""
}

func normalizedWeekdays(values []int) []int {
	items := make([]int, 0, len(values))
	seen := map[int]struct{}{}
	for _, value := range values {
		if value < 0 || value > 6 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		items = append(items, value)
	}
	slices.Sort(items)
	return items
}

func normalizedDateList(values []string) []string {
	items := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", value); err != nil {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		items = append(items, value)
	}
	slices.Sort(items)
	return items
}

func toggleWeekday(values []int, target int) []int {
	out := make([]int, 0, len(values))
	found := false
	for _, value := range values {
		if value == target {
			found = true
			continue
		}
		out = append(out, value)
	}
	if !found {
		out = append(out, target)
	}
	return normalizedWeekdays(out)
}

func toggleStreakKind(values []sharedtypes.StreakKind, target sharedtypes.StreakKind) []sharedtypes.StreakKind {
	out := make([]sharedtypes.StreakKind, 0, len(values))
	found := false
	for _, value := range values {
		if value == target {
			found = true
			continue
		}
		out = append(out, value)
	}
	if !found {
		out = append(out, target)
	}
	slices.Sort(out)
	return out
}
