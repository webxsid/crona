package controller

import (
	"slices"
	"strconv"
	"strings"

	shareddto "crona/shared/dto"

	tea "github.com/charmbracelet/bubbletea"
)

// captureState is everything the Add issues dialog and its review keep while a row is
// edited in the New Issue form, which closes and reopens dialogs.
type captureState struct {
	text     string
	target   string
	streamID int64
	rows     []shareddto.IssueCaptureProposal
	selected []bool
	cursor   int
	vocab    []string
}

func saveCapture(state State) captureState {
	return captureState{
		text: state.CaptureText, target: state.CaptureTarget, streamID: state.CaptureStreamID,
		rows: state.CaptureRows, selected: state.CaptureSelected, cursor: state.CaptureCursor,
		vocab: state.TagVocab,
	}
}

func restoreCapture(state State, saved captureState, kind string) State {
	state = Close(state)
	state.Kind = kind
	state.CaptureText, state.CaptureTarget, state.CaptureStreamID = saved.text, saved.target, saved.streamID
	state.CaptureRows, state.CaptureSelected, state.CaptureCursor = saved.rows, saved.selected, saved.cursor
	state.TagVocab = saved.vocab
	if kind == "capture_issues" {
		description := newDescriptionInput(60, 10)
		description.Placeholder = "- fix login redirect #api ~45m @today +urgent"
		description.CharLimit = 20000
		description.SetValue(saved.text)
		description.Focus()
		state.Description = description
		state.DescriptionEnabled = true
		state.DescriptionIndex = 0
	}
	return state
}

// OpenCaptureIssues opens the Add issues dialog. Lines without a `#` target go to streamID
// (target names it), or to the Inbox stream when streamID is 0.
func OpenCaptureIssues(state State, streamID int64, target string, vocab []string) State {
	return restoreCapture(state, captureState{target: target, streamID: streamID, vocab: slices.Clone(vocab)}, "capture_issues")
}

// OpenCaptureReview opens the review of parsed rows. Rows start selected unless flagged.
func OpenCaptureReview(state State, text string, rows []shareddto.IssueCaptureProposal) State {
	saved := saveCapture(state)
	saved.text, saved.rows, saved.cursor = text, rows, 0
	saved.selected = make([]bool, len(rows))
	for i, row := range rows {
		saved.selected[i] = len(row.Flags) == 0
	}
	return restoreCapture(state, saved, "capture_review")
}

// CaptureSelectedCount counts the rows that will be imported.
func CaptureSelectedCount(state State) int {
	count := 0
	for i := range state.CaptureRows {
		if i < len(state.CaptureSelected) && state.CaptureSelected[i] {
			count++
		}
	}
	return count
}

func updateCaptureIssues(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "ctrl+s":
		text := state.Description.Value()
		if strings.TrimSpace(text) == "" {
			return state, nil, "Type one issue per `- ` line"
		}
		// Unchanged text reopens the review with its selection and row edits.
		if state.CaptureRows != nil && text == state.CaptureText {
			return restoreCapture(state, saveCapture(state), "capture_review"), nil, ""
		}
		return state, &Action{Kind: "capture_parse", Title: text, StreamID: state.CaptureStreamID}, ""
	}
	var cmd tea.Cmd
	state.Description, cmd = state.Description.Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateCaptureReview(state State, ctx UpdateContext, msg tea.KeyMsg) (State, *Action, string) {
	rows := state.CaptureRows
	switch msg.String() {
	case "esc":
		return restoreCapture(state, saveCapture(state), "capture_issues"), nil, ""
	case "up", "k", "ctrl+p":
		if state.CaptureCursor > 0 {
			state.CaptureCursor--
		}
	case "down", "j", "ctrl+n":
		if state.CaptureCursor < len(rows)-1 {
			state.CaptureCursor++
		}
	case " ", "space":
		if i := state.CaptureCursor; i < len(rows) {
			if len(rows[i].Flags) > 0 {
				return state, nil, "Fix this row first: " + strings.Join(rows[i].Flags, ", ")
			}
			state.CaptureSelected = slices.Clone(state.CaptureSelected)
			state.CaptureSelected[i] = !state.CaptureSelected[i]
		}
	case "a":
		all := true
		for i, row := range rows {
			if len(row.Flags) == 0 && !state.CaptureSelected[i] {
				all = false
			}
		}
		state.CaptureSelected = make([]bool, len(rows))
		for i, row := range rows {
			state.CaptureSelected[i] = !all && len(row.Flags) == 0
		}
	case "enter":
		if state.CaptureCursor < len(rows) {
			return openCaptureRowEditor(state, ctx), nil, ""
		}
	case "ctrl+s":
		picked := make([]shareddto.IssueCaptureProposal, 0, len(rows))
		for i, row := range rows {
			if state.CaptureSelected[i] {
				picked = append(picked, row)
			}
		}
		if len(picked) == 0 {
			return state, nil, "Select at least one row to import"
		}
		state.Processing = true
		state.ProcessingLabel = "Importing…"
		return state, &Action{Kind: "capture_import", CaptureRows: picked, CaptureLeft: len(rows) - len(picked)}, ""
	default:
		return state, nil, ""
	}
	return clearDialogError(state), nil, ""
}

// openCaptureRowEditor edits the row under the cursor in the New Issue form; saving the form
// writes the row back instead of creating an issue (see updateCaptureRowEditor).
func openCaptureRowEditor(state State, ctx UpdateContext) State {
	saved := saveCapture(state)
	row := saved.rows[saved.cursor]
	next := OpenCreateIssueDefault(state)
	next = selectCaptureRowTarget(next, ctx, row)
	next.Inputs[2].SetValue(row.Title)
	next.Description.SetValue(ValueOrEmpty(row.Description))
	if row.EstimateMinutes != nil {
		next.Inputs[3].SetValue(FormatDurationMinutesInput(row.EstimateMinutes))
	}
	next.Inputs[4].SetValue(ValueOrEmpty(row.TodoForDate))
	next = withTagsField(next, defaultIssueFocusTags, row.Tags, saved.vocab)
	next.CaptureText, next.CaptureTarget, next.CaptureStreamID = saved.text, saved.target, saved.streamID
	next.CaptureRows, next.CaptureSelected, next.CaptureCursor = saved.rows, saved.selected, saved.cursor
	next.CaptureEditing = true
	return SyncDialogFocus(next)
}

// selectCaptureRowTarget points the form's pickers at the row's repo and stream. A part the
// parse left unresolved starts unchosen, so saving cannot quietly take the first option.
// The Inbox before its first import exists only by name.
func selectCaptureRowTarget(state State, ctx UpdateContext, row shareddto.IssueCaptureProposal) State {
	if row.Inbox && row.StreamID == 0 {
		state.Inputs[0].SetValue(row.RepoName)
		state.Inputs[1].SetValue(row.StreamName)
		state.RepoIndex, state.StreamIndex = 0, 0
		return state
	}
	state.RepoIndex, state.StreamIndex = -1, -1
	if row.RepoID == 0 {
		return state
	}
	state.RepoIndex = optionIndex(DefaultRepoOptions(state.Inputs, ctx.Repos), row.RepoID)
	if state.RepoIndex >= 0 && row.StreamID > 0 {
		streams := DefaultStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)
		state.StreamIndex = optionIndex(streams, row.StreamID)
	}
	return state
}

func optionIndex(options []SelectorOption, id int64) int {
	want := strconv.FormatInt(id, 10)
	for i, option := range options {
		if option.ID == want {
			return i
		}
	}
	return -1
}

// updateCaptureRowEditor runs the New Issue form for a review row. The form validates as
// usual; its create action becomes the row's new fields, and esc returns unchanged.
func updateCaptureRowEditor(state State, ctx UpdateContext, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	saved := saveCapture(state)
	if msg.String() == "esc" {
		return restoreCapture(state, saved, "capture_review"), nil, ""
	}
	if !isIssueCreateSubmitKey(msg.String()) {
		state.CaptureEditing = false
		next, action, status := Update(state, ctx, currentDate, msg)
		next.CaptureEditing = true
		return next, action, status
	}
	row := saved.rows[saved.cursor]
	repoID, streamID, inbox, ok := captureRowTarget(state, ctx, row)
	if !ok {
		return state, nil, "Pick an existing repo and stream"
	}
	state.CaptureEditing = false
	next, action, status := Update(state, ctx, currentDate, tea.KeyMsg{Type: tea.KeyCtrlS})
	if action == nil || action.Kind != "create_issue_default" {
		next.CaptureEditing = true
		return next, action, status
	}
	row.Title, row.Description, row.EstimateMinutes, row.TodoForDate = action.Title, action.Description, action.Estimate, action.DueDate
	row.RepoName, row.StreamName = action.RepoName, action.StreamName
	row.RepoID, row.StreamID, row.Inbox = repoID, streamID, inbox
	if action.Tags != nil {
		row.Tags = *action.Tags
	}
	row.Flags = nil
	saved.rows = slices.Clone(saved.rows)
	saved.rows[saved.cursor] = row
	saved.selected = slices.Clone(saved.selected)
	saved.selected[saved.cursor] = true
	return restoreCapture(next, saved, "capture_review"), nil, ""
}

// captureRowTarget resolves the form's repo and stream to existing ones. The Inbox counts
// as existing before its first import creates it.
func captureRowTarget(state State, ctx UpdateContext, row shareddto.IssueCaptureProposal) (int64, int64, bool, bool) {
	repo, ok := selectedOption(DefaultRepoOptions(state.Inputs, ctx.Repos), state.RepoIndex)
	if ok && repo.ID != "__new__" {
		streams := DefaultStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)
		if stream, ok := selectedOption(streams, state.StreamIndex); ok && stream.ID != "__new__" {
			repoID, _ := strconv.ParseInt(repo.ID, 10, 64)
			streamID, _ := strconv.ParseInt(stream.ID, 10, 64)
			if repoID > 0 && streamID > 0 {
				return repoID, streamID, false, true
			}
		}
	}
	if row.Inbox && strings.EqualFold(strings.TrimSpace(state.Inputs[0].Value()), row.RepoName) &&
		strings.EqualFold(strings.TrimSpace(state.Inputs[1].Value()), row.StreamName) {
		return 0, 0, true, true
	}
	return 0, 0, false, false
}
