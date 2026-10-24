package controller

import (
	"slices"
	"strings"

	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// TagTargetIssue and TagTargetHabit say which kind of item the tag picker edits.
const (
	TagTargetIssue = "issue"
	TagTargetHabit = "habit"
)

// OpenTagPicker opens the tag picker for one issue or habit: a checklist of the vocabulary
// under a text input that filters it or names a new tag.
func OpenTagPicker(state State, target string, id int64, title string, current, vocab []string) State {
	input := textinput.New()
	input.Placeholder = "Filter or new tag"
	input.CharLimit = sharedtypes.MaxTagNameLength + 1
	input.Width = 40
	input.Prompt = "+ "
	input.Focus()
	state = Close(state)
	state.Kind = "edit_tags"
	state.Inputs = []textinput.Model{input}
	state.TagTarget = target
	state.ViewName = title
	if target == TagTargetHabit {
		state.HabitID = id
	} else {
		state.IssueID = id
	}
	state.TagSelected = slices.Clone(current)
	state.TagVocab = slices.Clone(vocab)
	return state
}

// TagPickerRow is one line of the picker: a tag and whether it is checked, or the offer to
// create the tag typed in the input.
type TagPickerRow struct {
	Name    string
	Checked bool
	Create  bool
}

// TagPickerRows lists the vocabulary (plus tags checked but not yet in it) that match the input,
// led by a create row when the input names a valid tag that does not exist yet.
func TagPickerRows(state State) []TagPickerRow {
	query := ""
	if len(state.Inputs) > 0 {
		query = strings.ToLower(strings.TrimSpace(state.Inputs[0].Value()))
	}
	names := slices.Clone(state.TagVocab)
	for _, name := range state.TagSelected {
		if !slices.Contains(names, name) {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	rows := []TagPickerRow{}
	if query != "" {
		if name, err := sharedtypes.NormalizeTagName(query); err == nil && !slices.Contains(names, name) {
			rows = append(rows, TagPickerRow{Name: name, Create: true})
		}
	}
	for _, name := range names {
		if query == "" || strings.Contains(name, strings.TrimPrefix(query, "+")) {
			rows = append(rows, TagPickerRow{Name: name, Checked: slices.Contains(state.TagSelected, name)})
		}
	}
	return rows
}

func updateTagPicker(state State, msg tea.KeyMsg) (State, *Action, string) {
	rows := TagPickerRows(state)
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "up", "ctrl+p":
		if state.TagCursor > 0 {
			state.TagCursor--
		}
		return state, nil, ""
	case "down", "ctrl+n":
		if state.TagCursor < len(rows)-1 {
			state.TagCursor++
		}
		return state, nil, ""
	case " ", "space":
		if state.TagCursor >= 0 && state.TagCursor < len(rows) {
			state = toggleTagRow(state, rows[state.TagCursor])
		}
		return state, nil, ""
	case "enter":
		// A tag typed but not yet checked is meant to be added.
		if len(rows) > 0 && rows[0].Create {
			state = toggleTagRow(state, rows[0])
		}
		selected := slices.Clone(state.TagSelected)
		slices.Sort(selected)
		action := &Action{Kind: "set_tags", StringList: selected, Status: state.TagTarget}
		if state.TagTarget == TagTargetHabit {
			action.HabitID = state.HabitID
		} else {
			action.IssueID = state.IssueID
		}
		return Close(state), action, ""
	}
	var cmd tea.Cmd
	state.Inputs[0], cmd = state.Inputs[0].Update(msg)
	_ = cmd
	state.TagCursor = 0
	return clearDialogError(state), nil, ""
}

func toggleTagRow(state State, row TagPickerRow) State {
	if idx := slices.Index(state.TagSelected, row.Name); idx >= 0 {
		state.TagSelected = slices.Delete(state.TagSelected, idx, idx+1)
	} else {
		state.TagSelected = append(state.TagSelected, row.Name)
	}
	if row.Create {
		state.Inputs[0].SetValue("")
		state.TagCursor = 0
	}
	return state
}
