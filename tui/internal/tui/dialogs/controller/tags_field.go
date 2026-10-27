package controller

import (
	"slices"
	"strings"

	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// withTagsField adds the Tags field to an issue or habit form at focus, prefilled with
// current. The field completes the name being typed from vocab; → accepts the completion.
func withTagsField(state State, focus int, current, vocab []string) State {
	input := textinput.New()
	input.Placeholder = "comma or space separated"
	input.CharLimit = 200
	input.Width = 52
	input.ShowSuggestions = true
	input.KeyMap.AcceptSuggestion = key.NewBinding(key.WithKeys("right"))
	if len(current) > 0 {
		input.SetValue(strings.Join(current, " ") + " ")
	}
	state.TagsInput = input
	state.TagsEnabled = true
	state.TagsFocus = focus
	state.TagVocab = slices.Clone(vocab)
	return refreshTagSuggestions(state)
}

// TagsFieldFocused reports whether the Tags field holds focus.
func TagsFieldFocused(state State) bool {
	return state.TagsEnabled && state.FocusIdx == state.TagsFocus
}

// updateTagsField types into the focused Tags field. It leaves tab, esc, enter and
// shortcuts to the form, and reports whether it consumed the key.
func updateTagsField(state State, msg tea.KeyMsg) (State, bool) {
	if !TagsFieldFocused(state) {
		return state, false
	}
	switch msg.Type {
	case tea.KeyRunes, tea.KeySpace, tea.KeyBackspace, tea.KeyDelete,
		tea.KeyLeft, tea.KeyRight, tea.KeyHome, tea.KeyEnd:
	default:
		return state, false
	}
	state.TagsInput, _ = state.TagsInput.Update(msg)
	return clearDialogError(refreshTagSuggestions(state)), true
}

// refreshTagSuggestions offers the vocabulary names that complete the last word typed. The
// input matches suggestions against its whole value, so each one keeps the earlier words.
func refreshTagSuggestions(state State) State {
	value := state.TagsInput.Value()
	cut := strings.LastIndexAny(value, ", ") + 1
	prefix, raw := value[:cut], value[cut:]
	lead := ""
	if strings.HasPrefix(raw, "+") {
		lead = "+"
	}
	word := strings.ToLower(strings.TrimPrefix(raw, "+"))
	if word == "" {
		state.TagsInput.SetSuggestions(nil)
		return state
	}
	typed := splitTagWords(prefix)
	suggestions := []string{}
	for _, name := range state.TagVocab {
		if strings.HasPrefix(name, word) && name != word && !slices.Contains(typed, name) {
			suggestions = append(suggestions, prefix+lead+name)
		}
	}
	state.TagsInput.SetSuggestions(suggestions)
	return state
}

// tagsFromField parses the Tags field, or returns nil when the form has none.
func tagsFromField(state State) (*[]string, error) {
	if !state.TagsEnabled {
		return nil, nil
	}
	names, err := sharedtypes.NormalizeTagNames(splitTagWords(state.TagsInput.Value()))
	if err != nil {
		return nil, err
	}
	return &names, nil
}

func splitTagWords(value string) []string {
	words := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' })
	out := make([]string, 0, len(words))
	for _, word := range words {
		if word = strings.TrimPrefix(word, "+"); word != "" {
			out = append(out, strings.ToLower(word))
		}
	}
	return out
}
