package controller

import (
	"strings"
	"time"

	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

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
	repoFilter = withSearchPrompt(state, repoFilter)
	streamFilter := textinput.New()
	streamFilter.Placeholder = "Search stream"
	streamFilter.CharLimit = 80
	streamFilter.Width = 52
	streamFilter = withSearchPrompt(state, streamFilter)
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
	target.Placeholder = "Target duration (e.g. 30m, 1h)"
	target.CharLimit = 8
	target.Width = 52
	target = withTimePrompt(state, target)
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
	target.Placeholder = "Target duration (e.g. 30m, 1h)"
	target.CharLimit = 8
	target.Width = 52
	target = withTimePrompt(state, target)
	if targetMinutes != nil {
		target.SetValue(FormatDurationMinutesInput(targetMinutes))
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
	return openHabitActivity(state, "complete_habit", "Habit Log", habitID, date, durationMinutes, notes)
}

func openHabitActivity(state State, kind, title string, habitID int64, date string, durationMinutes *int, notes *string) State {
	duration := textinput.New()
	duration.Placeholder = "Duration (e.g. 30m, 1h15m)"
	duration.Focus()
	duration.CharLimit = 8
	duration.Width = 52
	duration = withTimePrompt(state, duration)
	if durationMinutes != nil {
		duration.SetValue(FormatDurationMinutesInput(durationMinutes))
	}
	description := newDescriptionInput(52, 5)
	if notes != nil {
		description.SetValue(strings.TrimSpace(*notes))
	}
	state = Close(state)
	state.Kind = kind
	state.Inputs = []textinput.Model{duration}
	state.Description = description
	state.DescriptionEnabled = true
	state.DescriptionIndex = 1
	state.HabitID = habitID
	state.CheckInDate = date
	state.ViewTitle = title
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
	estimate.Placeholder = "Estimate (e.g. 45m, 1h30m)"
	estimate.CharLimit = 8
	estimate.Width = 52
	estimate = withTimePrompt(state, estimate)
	due := textinput.New()
	due.Placeholder = "Due date YYYY-MM-DD (optional)"
	due.CharLimit = 10
	due.Width = 52
	due = withDatePrompt(state, due)
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
	estimateInput.Placeholder = "Estimate (e.g. 45m, 1h30m)"
	estimateInput.CharLimit = 8
	estimateInput.Width = 52
	estimateInput = withTimePrompt(state, estimateInput)
	if estimateMinutes != nil {
		estimateInput.SetValue(FormatDurationMinutesInput(estimateMinutes))
	}
	dueInput := textinput.New()
	dueInput.Placeholder = "Due date YYYY-MM-DD (optional)"
	dueInput.CharLimit = 10
	dueInput.Width = 52
	dueInput = withDatePrompt(state, dueInput)
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
	repoFilter = withSearchPrompt(state, repoFilter)
	streamFilter := textinput.New()
	streamFilter.Placeholder = "Search stream"
	streamFilter.CharLimit = 80
	streamFilter.Width = 52
	streamFilter = withSearchPrompt(state, streamFilter)
	title := textinput.New()
	title.Placeholder = "Issue title"
	title.Focus()
	title.CharLimit = 120
	title.Width = 52
	description := newDescriptionInput(52, 5)
	estimate := textinput.New()
	estimate.Placeholder = "Estimate (e.g. 45m, 1h30m)"
	estimate.CharLimit = 8
	estimate.Width = 52
	estimate = withTimePrompt(state, estimate)
	due := textinput.New()
	due.Placeholder = "Due date YYYY-MM-DD (optional)"
	due.CharLimit = 10
	due.Width = 52
	due = withDatePrompt(state, due)
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
	repoFilter = withSearchPrompt(state, repoFilter)
	streamFilter := textinput.New()
	streamFilter.Placeholder = "Search stream"
	streamFilter.CharLimit = 80
	streamFilter.Width = 52
	streamFilter = withSearchPrompt(state, streamFilter)
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
		newSessionDetailInput(state, "Summary (optional)"),
		newSessionDetailInput(state, "YYYY-MM-DD"),
		newSessionDetailInput(state, "90 | 90m | 1h30m"),
		newSessionDetailInput(state, "15m | 0m"),
		newSessionDetailInput(state, "09:00"),
		newSessionDetailInput(state, "10:45"),
		newSessionDetailInput(state, "Notes (optional)"),
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
