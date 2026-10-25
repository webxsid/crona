package controller

import (
	"strconv"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"

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

func OpenEditSessionHistory(state State, detail *api.SessionDetail) State {
	if detail == nil {
		return state
	}
	inputs := []textinput.Model{
		newSessionDetailInput(state, "YYYY-MM-DD HH:mm:ss"),
		newSessionDetailInput(state, "YYYY-MM-DD HH:mm:ss"),
	}
	inputs[0].SetValue(sessionHistoryTimestampValue(detail.StartTime))
	if detail.EndTime != nil {
		inputs[1].SetValue(sessionHistoryTimestampValue(*detail.EndTime))
	}
	noteSections := []sharedtypes.SessionNoteSection{
		sharedtypes.SessionNoteSectionCommit,
		sharedtypes.SessionNoteSectionWork,
		sharedtypes.SessionNoteSectionNotes,
	}
	noteInputs := make([]textarea.Model, 0, len(noteSections))
	for _, section := range noteSections {
		input := newDescriptionInput(64, 3)
		if detail.ParsedNotes != nil {
			input.SetValue(detail.ParsedNotes[section])
		}
		noteInputs = append(noteInputs, input)
	}
	segmentRows := make([]SessionHistorySegmentRow, 0, len(detail.Segments))
	for _, segment := range detail.Segments {
		row := newSessionHistorySegmentRow(state, segment.SegmentType)
		row.StartTime.SetValue(sessionHistoryTimestampValue(segment.StartTime))
		if segment.EndTime != nil {
			row.EndTime.SetValue(sessionHistoryTimestampValue(*segment.EndTime))
		}
		segmentRows = append(segmentRows, row)
	}
	state = Close(state)
	state.Kind = "edit_session_history"
	state.SessionID = detail.ID
	state.SessionHistoryContext = strings.TrimSpace(
		detail.RepoName + " / " + detail.StreamName + " / Issue #" +
			strconv.FormatInt(detail.IssueID, 10) + " " + detail.IssueTitle,
	)
	state.Inputs = inputs
	state.SessionHistoryNoteInputs = noteInputs
	state.SessionHistorySegments = segmentRows
	state.FocusIdx = 0
	return SyncDialogFocus(state)
}

func newSessionHistorySegmentRow(
	state State,
	segmentType sharedtypes.SessionSegmentType,
) SessionHistorySegmentRow {
	if segmentType == "" {
		segmentType = sharedtypes.SessionSegmentWork
	}
	return SessionHistorySegmentRow{
		Type:      segmentType,
		StartTime: newSessionDetailInput(state, "YYYY-MM-DD HH:mm:ss"),
		EndTime:   newSessionDetailInput(state, "YYYY-MM-DD HH:mm:ss"),
	}
}

func sessionHistoryTimestampValue(value string) string {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return strings.TrimSpace(value)
	}
	return parsed.In(time.Local).Format("2006-01-02 15:04:05.999999999 -07:00")
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

func OpenEditStream(
	state State,
	streamID, repoID int64,
	streamName, repoName string,
	descriptionValue *string,
) State {
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
	state = withTagsField(state, 7, nil, state.TagVocab)
	return SyncDialogFocus(initCadence(state))
}

func OpenEditHabit(
	state State,
	habitID, streamID int64,
	name string,
	descriptionValue *string,
	scheduleRaw string,
	targetMinutes *int,
	active bool,
) State {
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
	state = withTagsField(state, 5, nil, state.TagVocab)
	return initCadence(state)
}

func OpenHabitCompletion(
	state State,
	habitID int64,
	date string,
	durationMinutes *int,
	notes *string,
) State {
	return openHabitActivity(
		state,
		"complete_habit",
		"Habit Log",
		habitID,
		date,
		durationMinutes,
		notes,
	)
}

func openHabitActivity(
	state State,
	kind, title string,
	habitID int64,
	date string,
	durationMinutes *int,
	notes *string,
) State {
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
	return withTagsField(state, issueMetaFocusTags, nil, state.TagVocab)
}

func OpenEditIssue(
	state State,
	issueID, streamID int64,
	title string,
	descriptionValue *string,
	estimateMinutes *int,
	todoForDate *string,
) State {
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
	return withTagsField(state, issueMetaFocusTags, nil, state.TagVocab)
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
	state.DescriptionIndex = defaultIssueFocusDescription
	state.FocusIdx = defaultIssueFocusTitle
	state = withTagsField(state, defaultIssueFocusTags, nil, state.TagVocab)
	return SyncDialogFocus(state)
}

func OpenPromoteAdhocIssue(state State, issue api.IssueWithMeta, context *api.ActiveContext) State {
	state = OpenCreateIssueDefault(state)
	state.PromotingAdhoc = true
	state.IssueID = issue.ID
	state.ViewName = "Choose a repo and stream. This issue stays Done; session history is preserved."
	state.Inputs[2].SetValue(issue.Title)
	state.Description.SetValue(ValueOrEmpty(issue.Description))
	if issue.EstimateMinutes != nil {
		state.Inputs[3].SetValue(FormatDurationMinutesInput(issue.EstimateMinutes))
	}
	if issue.TodoForDate != nil {
		state.Inputs[4].SetValue(*issue.TodoForDate)
		state.AdhocDate = strings.TrimSpace(*issue.TodoForDate)
	}
	if context != nil {
		if context.RepoName != nil {
			state.Inputs[0].SetValue(*context.RepoName)
		}
		if context.StreamName != nil {
			state.Inputs[1].SetValue(*context.StreamName)
		}
	}
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
	state.RepoIndex = -1
	state.StreamIndex = -1
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

func OpenIssueStatus(
	state State,
	status, title, repoName, streamName string,
	estimateMinutes *int,
	workedSeconds int,
) State {
	state = Close(state)
	state.Kind = "issue_status"
	state.IssueStatus = status
	state.ViewName = title
	state.RepoName = repoName
	state.StreamName = streamName
	state.IssueEstimateMins = estimateMinutes
	state.IssueWorkedSeconds = workedSeconds
	state.StatusItems = sharedtypes.AllowedIssueStatusTransitions(sharedtypes.IssueStatus(status))
	return state
}

func OpenIssueStatusNote(
	state State,
	issueID, streamID int64,
	status, label string,
	required bool,
) State {
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
	}
	input.Focus()
	input.CharLimit = 200
	input.Width = 48
	state = Close(state)
	state.Kind = kind
	state.Inputs = []textinput.Model{input}
	return state
}

func OpenSessionMessageWithParent(state State, kind string, parent string) State {
	viewName := state.ViewName
	state = OpenSessionMessage(state, kind)
	state.Parent = parent
	state.ViewName = viewName
	return state
}

func OpenTimerStartType(
	state State,
	repoID, streamID, issueID int64,
	issueLabel string,
	estimateMinutes *int,
	workedSeconds int,
) State {
	adhoc := state.AdhocTimer
	state = Close(state)
	state.AdhocTimer = adhoc
	state.Kind = "timer_start_type"
	state.ViewTitle = "Start Timer"
	state.ViewName = strings.TrimSpace(issueLabel)
	state.RepoID = repoID
	state.StreamID = streamID
	state.IssueID = issueID
	state.IssueEstimateMins = estimateMinutes
	state.IssueWorkedSeconds = workedSeconds
	state.ChoiceItems = []string{"Stopwatch", "Pomodoro", "Timer"}
	state.ChoiceValues = []string{"stopwatch", "pomodoro", "timer"}
	state.ChoiceDetails = []string{
		"Start an open-ended focus session.",
		"Configure focus, breaks, and cycles in a single pomodoro setup dialog.",
		"Run a single countdown with no breaks. Defaults from the remaining issue estimate.",
	}
	state.ChoiceCursor = 0
	return state
}

func OpenAdhocTimerStartType(state State) State {
	state = OpenTimerStartType(state, 0, 0, 0, "Ad hoc", nil, 0)
	state.AdhocTimer = true
	return state
}

func OpenAdhocResolution(state State, context *api.ActiveContext) State {
	adhocElapsedSeconds := state.AdhocElapsedSeconds
	adhocDate := state.AdhocDate
	repoName, streamName := state.RepoName, state.StreamName
	state = Close(state)
	state.Kind = "adhoc_resolution"
	state.ViewTitle = "Finish Ad hoc Session"
	state.ViewName = "Choose where this session should be recorded."
	state.AdhocTimer = true
	state.AdhocElapsedSeconds = adhocElapsedSeconds
	state.AdhocDate = adhocDate
	state.RepoName = repoName
	state.StreamName = streamName
	state.ChoiceItems = []string{
		"Attach to an existing issue",
		"Create a new issue",
		"Keep as ad hoc",
	}
	state.ChoiceValues = []string{"attach", "create", "keep"}
	state.ChoiceDetails = []string{
		"Choose an issue, then add a session summary.",
		"Create an issue and finish this session in one step.",
		"Create a completed issue in Ad hoc / Sessions. Convert it later.",
	}
	if context != nil {
		if context.RepoID != nil {
			state.RepoID = *context.RepoID
		}
		if context.RepoName != nil {
			state.RepoName = *context.RepoName
		}
		if context.StreamID != nil {
			state.StreamID = *context.StreamID
		}
		if context.StreamName != nil {
			state.StreamName = *context.StreamName
		}
	}
	return state
}

func OpenAdhocIssuePicker(state State, issues []api.IssueWithMeta, currentDate string) State {
	adhocElapsedSeconds := state.AdhocElapsedSeconds
	adhocDate := state.AdhocDate
	repoID, streamID := state.RepoID, state.StreamID
	repoName, streamName := state.RepoName, state.StreamName
	initialOpening := state.Kind != "adhoc_issue_picker"
	tab := state.AdhocIssueTab
	state = Close(state)
	state.Kind = "adhoc_issue_picker"
	state.ViewTitle = "Attach to Issue"
	state.ViewName = "Choose an issue. You’ll add a session summary before finishing."
	state.AdhocTimer = true
	state.AdhocResolution = sharedtypes.AdhocSessionResolutionAttach
	state.AdhocElapsedSeconds = adhocElapsedSeconds
	state.AdhocDate = adhocDate
	state.RepoID = repoID
	state.StreamID = streamID
	state.RepoName = repoName
	state.StreamName = streamName
	if tab < 0 || tab >= len(adhocIssueTabNames) {
		tab = 0
	}
	state.AdhocIssueTab = tab
	eligible := make([]api.IssueWithMeta, 0, len(issues))
	for _, issue := range issues {
		if !sharedtypes.CanStartFocus(issue.Status) {
			continue
		}
		eligible = append(eligible, issue)
		dueDate := ""
		if issue.TodoForDate != nil {
			dueDate = strings.TrimSpace(*issue.TodoForDate)
		}
		if currentDate != "" && dueDate == currentDate {
			state.AdhocIssueTabCounts[0]++
		}
		if currentDate != "" && dueDate != "" && dueDate < currentDate {
			state.AdhocIssueTabCounts[1]++
		}
		if issue.PinnedDaily {
			state.AdhocIssueTabCounts[2]++
		}
	}
	state.AdhocIssueTabCounts[3] = len(eligible)
	if initialOpening && tab == 0 && state.AdhocIssueTabCounts[0] == 0 &&
		state.AdhocIssueTabCounts[3] > 0 {
		tab = 3
		state.AdhocIssueTab = tab
	}
	state.ChoiceCursor = -1
	for _, issue := range eligible {
		dueDate := ""
		if issue.TodoForDate != nil {
			dueDate = strings.TrimSpace(*issue.TodoForDate)
		}
		matches := tab == 3 || (tab == 0 && currentDate != "" && dueDate == currentDate) ||
			(tab == 1 && currentDate != "" && dueDate != "" && dueDate < currentDate) ||
			(tab == 2 && issue.PinnedDaily)
		if !matches {
			continue
		}
		state.AdhocIssueIDs = append(state.AdhocIssueIDs, issue.ID)
		state.ChoiceValues = append(state.ChoiceValues, strconv.FormatInt(issue.ID, 10))
		state.ChoiceItems = append(state.ChoiceItems, issue.Title)
		status := strings.ReplaceAll(string(issue.Status), "_", " ")
		detail := issue.RepoName + " / " + issue.StreamName + " · " + status
		if dueDate != "" {
			detail += " · " + dueDate
		}
		if issue.PinnedDaily {
			detail += " · pinned"
		}
		state.ChoiceDetails = append(state.ChoiceDetails, detail)
		if state.ChoiceCursor < 0 {
			state.ChoiceCursor = len(state.ChoiceItems) - 1
		}
	}
	return state
}

var adhocIssueTabNames = [...]string{"Due today", "Overdue", "Pinned", "All"}

func OpenAdhocNewIssue(state State) State {
	input := textinput.New()
	input.Placeholder = "Issue title"
	input.Focus()
	input.CharLimit = 120
	input.Width = 52
	repoID, streamID := state.RepoID, state.StreamID
	state = Close(state)
	state.Kind = "adhoc_new_issue"
	state.ViewTitle = "Create Issue"
	state.ViewName = "The completed session will be attached to this issue."
	state.AdhocTimer = true
	state.AdhocResolution = sharedtypes.AdhocSessionResolutionCreate
	state.RepoID = repoID
	state.StreamID = streamID
	state.Inputs = []textinput.Model{input}
	return state
}

func OpenAdhocIssueCreation(state State, context *api.ActiveContext) State {
	adhocElapsedSeconds := state.AdhocElapsedSeconds
	adhocDate := state.AdhocDate
	repoName, streamName := state.RepoName, state.StreamName
	state = OpenCreateIssueDefault(state)
	state.AdhocTimer = true
	state.AdhocResolution = sharedtypes.AdhocSessionResolutionAttach
	state.AdhocElapsedSeconds = adhocElapsedSeconds
	state.AdhocDate = adhocDate
	state.ViewName = "Saving creates the issue, attaches this session, and stops the timer."
	commit := textinput.New()
	commit.Placeholder = "What did you work on?"
	commit.CharLimit = 200
	commit.Width = 52
	state.Inputs = append(state.Inputs, commit)
	if context != nil {
		if context.RepoName != nil {
			repoName = *context.RepoName
		}
		if context.StreamName != nil {
			streamName = *context.StreamName
		}
	}
	state.Inputs[0].SetValue(repoName)
	state.Inputs[1].SetValue(streamName)
	if adhocElapsedSeconds > 0 {
		minutes := ((adhocElapsedSeconds + 15*60 - 1) / (15 * 60)) * 15
		state.Inputs[3].SetValue(FormatDurationMinutesInput(&minutes))
	}
	if adhocDate != "" {
		state.Inputs[4].SetValue(adhocDate)
	}
	state.FocusIdx = defaultIssueFocusTitle
	return SyncDialogFocus(state)
}

func OpenAdhocEndSession(
	state State,
	resolution sharedtypes.AdhocSessionResolution,
	issueID, repoID, streamID int64,
	newIssueTitle string,
) State {
	elapsed := state.AdhocElapsedSeconds
	date := state.AdhocDate
	state = OpenSessionMessage(state, "end_session")
	state.AdhocElapsedSeconds = elapsed
	state.AdhocDate = date
	state.Inputs[0].Placeholder = "Session summary"
	state.AdhocTimer = true
	state.AdhocResolution = resolution
	state.IssueID = issueID
	state.RepoID = repoID
	state.StreamID = streamID
	state.NewIssueTitle = newIssueTitle
	return state
}

func OpenPomodoroStart(
	state State,
	repoID, streamID, issueID int64,
	issueLabel string,
	estimateMinutes *int,
	workedSeconds int,
) State {
	adhoc := state.AdhocTimer
	focusSeconds := state.PomodoroFocusSeconds
	if focusSeconds <= 0 {
		focusSeconds = 25 * 60
	}
	breakSeconds := state.PomodoroBreakSeconds
	if breakSeconds <= 0 {
		breakSeconds = 5 * 60
	}
	longBreakSeconds := state.PomodoroLongBreakSeconds
	if longBreakSeconds <= 0 {
		longBreakSeconds = 15 * 60
	}
	cyclesBeforeLongBreak := state.PomodoroCyclesBeforeLongBreak
	if cyclesBeforeLongBreak <= 0 {
		cyclesBeforeLongBreak = 4
	}
	cycles := state.PomodoroCycles
	if cycles <= 0 {
		cycles = 4
	}
	state = Close(state)
	state.AdhocTimer = adhoc
	state.Kind = "pomodoro_start"
	state.Parent = "timer_start_type"
	state.ViewTitle = "Pomodoro Session"
	state.RepoID = repoID
	state.StreamID = streamID
	state.IssueID = issueID
	state.ViewName = strings.TrimSpace(issueLabel)
	state.IssueEstimateMins = estimateMinutes
	state.IssueWorkedSeconds = workedSeconds
	state.Inputs = newPomodoroDialogInputs(
		state,
		focusSeconds,
		breakSeconds,
		longBreakSeconds,
		cycles,
		cyclesBeforeLongBreak,
	)
	state.PomodoroFocusSeconds = focusSeconds
	state.PomodoroFocusChoice = pomodoroFocusChoiceForSeconds(focusSeconds)
	state.PomodoroBreakSeconds = breakSeconds
	state.PomodoroBreakChoice = pomodoroBreakChoiceForSeconds(breakSeconds)
	state.PomodoroLongBreakSeconds = longBreakSeconds
	state.PomodoroLongBreakChoice = pomodoroLongBreakChoiceForSeconds(longBreakSeconds)
	state.PomodoroCyclesBeforeLongBreak = cyclesBeforeLongBreak
	state.PomodoroCycles = cycles
	state.FocusIdx = 0
	return SyncDialogFocus(state)
}

func OpenSingleTimerStart(
	state State,
	repoID, streamID, issueID int64,
	issueLabel string,
	estimateMinutes *int,
	workedSeconds int,
) State {
	adhoc := state.AdhocTimer
	countdownSeconds := state.TimerCountdownSeconds
	if countdownSeconds <= 0 {
		countdownSeconds = timerDefaultCountdownSeconds(estimateMinutes, workedSeconds)
	}
	input := newSessionDetailInput(state, "25m")
	input.SetValue(pomodoroSeedDurationInput(countdownSeconds, 25*60))
	input.Focus()

	state = Close(state)
	state.AdhocTimer = adhoc
	state.Kind = "timer_countdown_start"
	state.Parent = "timer_start_type"
	state.ViewTitle = "Timer Session"
	state.RepoID = repoID
	state.StreamID = streamID
	state.IssueID = issueID
	state.ViewName = strings.TrimSpace(issueLabel)
	state.IssueEstimateMins = estimateMinutes
	state.IssueWorkedSeconds = workedSeconds
	state.TimerCountdownSeconds = countdownSeconds
	state.Inputs = []textinput.Model{input}
	state.FocusIdx = 0
	return SyncDialogFocus(state)
}

func OpenHardLimitExpired(state State, issueLabel string) State {
	kind := sharedtypes.NormalizeTimerHardLimitKind(state.HardLimitKind)
	totalSeconds := state.HardLimitTotalSeconds
	focusSeconds := state.HardLimitFocusSeconds
	breakSeconds := state.HardLimitBreakSeconds
	longBreakSeconds := state.HardLimitLongBreakSeconds
	cyclesBeforeLongBreak := state.HardLimitCyclesBeforeLongBreak
	state = Close(state)
	state.Kind = "hard_limit_expired"
	state.HardLimitKind = kind
	state.HardLimitTotalSeconds = totalSeconds
	state.HardLimitFocusSeconds = focusSeconds
	state.HardLimitBreakSeconds = breakSeconds
	state.HardLimitLongBreakSeconds = longBreakSeconds
	state.HardLimitCyclesBeforeLongBreak = cyclesBeforeLongBreak
	state.ViewTitle = "Pomodoro Session Complete"
	state.ViewName = strings.TrimSpace(issueLabel)
	state.ChoiceItems = []string{"[c] Commit", "[e] Extend"}
	state.ChoiceValues = []string{"commit", "extend"}
	state.ChoiceDetails = []string{
		"Finish the session and capture what was completed.",
		"Add more Pomodoro sessions and keep the same cadence running.",
	}
	if kind == sharedtypes.TimerHardLimitKindCountdown {
		state.ViewTitle = "Timer Session Complete"
		state.ChoiceDetails[1] = "Add more time and keep the same timer session running."
	}
	state.ChoiceCursor = 0
	return state
}

func OpenHardLimitExtend(state State) State {
	if sharedtypes.NormalizeTimerHardLimitKind(state.HardLimitKind) ==
		sharedtypes.TimerHardLimitKindCountdown {
		return OpenCountdownExtend(state)
	}
	viewName := state.ViewName
	totalSeconds := state.HardLimitTotalSeconds
	focusSeconds := state.HardLimitFocusSeconds
	breakSeconds := state.HardLimitBreakSeconds
	longBreakSeconds := state.HardLimitLongBreakSeconds
	cyclesBeforeLongBreak := state.HardLimitCyclesBeforeLongBreak
	state = Close(state)
	state.Kind = "hard_limit_extend"
	state.Parent = "hard_limit_expired"
	state.ViewName = viewName
	if focusSeconds <= 0 {
		focusSeconds = 25 * 60
	}
	state.HardLimitFocusSeconds = focusSeconds
	state.HardLimitBreakSeconds = breakSeconds
	state.HardLimitLongBreakSeconds = longBreakSeconds
	state.HardLimitCyclesBeforeLongBreak = cyclesBeforeLongBreak
	state.PomodoroFocusSeconds = focusSeconds
	state.PomodoroFocusChoice = pomodoroFocusChoiceForSeconds(focusSeconds)
	state.PomodoroBreakSeconds = breakSeconds
	state.PomodoroBreakChoice = pomodoroBreakChoiceForSeconds(breakSeconds)
	state.PomodoroLongBreakSeconds = longBreakSeconds
	state.PomodoroLongBreakChoice = pomodoroLongBreakChoiceForSeconds(longBreakSeconds)
	state.PomodoroCyclesBeforeLongBreak = cyclesBeforeLongBreak
	if state.PomodoroCyclesBeforeLongBreak < 0 {
		state.PomodoroCyclesBeforeLongBreak = 4
	}
	if state.PomodoroBreakSeconds <= 0 {
		state.PomodoroCycles = 1
		state.PomodoroLongBreakSeconds = 0
		state.PomodoroLongBreakChoice = pomodoroLongBreakNoBreakChoice
		state.PomodoroCyclesBeforeLongBreak = 0
	} else {
		state.PomodoroCycles = inferPomodoroCycles(
			totalSeconds,
			state.PomodoroFocusSeconds,
			state.PomodoroBreakSeconds,
			state.PomodoroLongBreakSeconds,
			state.PomodoroCyclesBeforeLongBreak,
		)
		if state.PomodoroLongBreakSeconds > 0 && state.PomodoroCyclesBeforeLongBreak <= 0 {
			state.PomodoroCyclesBeforeLongBreak = 4
		}
	}
	state.Inputs = newPomodoroDialogInputs(
		state,
		focusSeconds,
		breakSeconds,
		longBreakSeconds,
		max(1, state.PomodoroCycles),
		state.PomodoroCyclesBeforeLongBreak,
	)
	state.FocusIdx = pomodoroFocusRowIdx
	return SyncDialogFocus(state)
}

func OpenCountdownExtend(state State) State {
	viewName := state.ViewName
	durationSeconds := state.HardLimitFocusSeconds
	if durationSeconds <= 0 {
		durationSeconds = state.HardLimitTotalSeconds
	}
	if durationSeconds <= 0 {
		durationSeconds = 25 * 60
	}
	input := withTimePrompt(state, newSessionDetailInput(state, "25m"))
	input.SetValue(pomodoroSeedDurationInput(durationSeconds, 25*60))
	input.Focus()

	state = Close(state)
	state.Kind = "timer_countdown_extend"
	state.Parent = "hard_limit_expired"
	state.ViewName = viewName
	state.HardLimitKind = sharedtypes.TimerHardLimitKindCountdown
	state.HardLimitFocusSeconds = durationSeconds
	state.Inputs = []textinput.Model{input}
	state.FocusIdx = 0
	return SyncDialogFocus(state)
}

func newPomodoroDialogInputs(
	state State,
	focusSeconds, breakSeconds, longBreakSeconds, cycles, cyclesBeforeLongBreak int,
) []textinput.Model {
	inputs := []textinput.Model{
		newSessionDetailInput(state, "25m"),
		newSessionDetailInput(state, "5m"),
		newSessionDetailInput(state, "15m"),
		newSessionDetailInput(state, "4"),
		newSessionDetailInput(state, "4"),
	}
	inputs[0].SetValue(pomodoroSeedDurationInput(focusSeconds, 25*60))
	inputs[1].SetValue(pomodoroSeedDurationInput(breakSeconds, 5*60))
	inputs[2].SetValue(pomodoroSeedDurationInput(longBreakSeconds, 15*60))
	inputs[3].SetValue(strconv.Itoa(max(1, cycles)))
	if cyclesBeforeLongBreak > 0 {
		inputs[4].SetValue(strconv.Itoa(cyclesBeforeLongBreak))
	}
	return inputs
}

func timerDefaultCountdownSeconds(estimateMinutes *int, workedSeconds int) int {
	if estimateMinutes != nil && *estimateMinutes > 0 {
		remaining := (*estimateMinutes * 60) - max(0, workedSeconds)
		if remaining > 0 {
			return remaining
		}
	}
	return 25 * 60
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

func OpenManualSession(
	state State,
	issueID int64,
	issueLabel string,
	estimateMinutes *int,
	date string,
) State {
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

func OpenDatePicker(
	state State,
	parentDialog string,
	issueID int64,
	inputIndex int,
	initial *string,
	currentDate string,
) State {
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

func OpenSummaryCalendar(state State, date string) State {
	selected := ResolveDialogDate(nil, date)
	monthStart := time.Date(selected.Year(), selected.Month(), 1, 0, 0, 0, 0, selected.Location())
	state.Kind = "summary_calendar"
	state.Parent = ""
	state.DateCursorValue = selected.Format("2006-01-02")
	state.DateMonthValue = monthStart.Format("2006-01-02")
	return state
}
