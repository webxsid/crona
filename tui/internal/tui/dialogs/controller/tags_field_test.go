package controller

import (
	"slices"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func typeInForm(state State, text string) State {
	for _, r := range text {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		if r == ' ' {
			msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
		}
		state, _, _ = Update(state, UpdateContext{}, "2026-10-25", msg)
	}
	return state
}

func TestEditIssueTagsFieldSavesNormalizedTags(t *testing.T) {
	state := OpenEditIssue(State{}, 7, 3, "Fix login", nil, nil, nil)
	state = withTagsField(state, issueMetaFocusTags, []string{"urgent"}, []string{"urgent", "review", "release"})
	if got := state.TagsInput.Value(); got != "urgent " {
		t.Fatalf("prefill: got %q", got)
	}
	for range issueMetaFocusTags {
		state, _, _ = Update(state, UpdateContext{}, "2026-10-25", tea.KeyMsg{Type: tea.KeyTab})
	}
	if !TagsFieldFocused(state) {
		t.Fatalf("expected Tags focused, focus %d", state.FocusIdx)
	}
	state = typeInForm(state, "re")
	state, _, _ = Update(state, UpdateContext{}, "2026-10-25", tea.KeyMsg{Type: tea.KeyRight})
	if got := state.TagsInput.Value(); got != "urgent review" && got != "urgent release" {
		t.Fatalf("→ should complete from the vocabulary, got %q", got)
	}
	state = typeInForm(state, ", Client Acme")
	_, action, status := Update(state, UpdateContext{}, "2026-10-25", tea.KeyMsg{Type: tea.KeyCtrlS})
	if action == nil {
		t.Fatalf("expected a save action, status %q", status)
	}
	if action.Tags == nil || !slices.Contains(*action.Tags, "urgent") || !slices.Contains(*action.Tags, "client") || !slices.Contains(*action.Tags, "acme") {
		t.Fatalf("unexpected tags %v", action.Tags)
	}
}

func TestCreateHabitTagsFieldIsLastStop(t *testing.T) {
	state := OpenCreateHabit(State{})
	if state.TagsFocus != 7 {
		t.Fatalf("create_habit Tags focus %d, want 7", state.TagsFocus)
	}
	state.FocusIdx = 5
	state, _, _ = Update(state, UpdateContext{}, "2026-10-25", tea.KeyMsg{Type: tea.KeyTab})
	if !TagsFieldFocused(state) {
		t.Fatalf("tab from Target should reach Tags, got %d", state.FocusIdx)
	}
}
