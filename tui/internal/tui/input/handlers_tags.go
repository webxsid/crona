package input

import (
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	keyregistry "crona/tui/internal/tui/key_registry"
	uistate "crona/tui/internal/tui/state"
	tagsmeta "crona/tui/internal/tui/views/tagsmeta"

	tea "github.com/charmbracelet/bubbletea"
)

func registerTagKeys(r *keyregistry.Router[State, uistate.View, uistate.Pane], deps Deps) {
	keyregistry.RegisterSettings(
		r,
		uistate.ViewTags,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCycleTagColor(s, deps, -1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCycleTagColor(s, deps, 1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShowTaggedItems(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShowTaggedItems(s, deps) },
	)
	r.RegisterView(uistate.ViewTags, "r", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return handleRenameSelectedTag(s, deps)
	})
	r.RegisterView(uistate.ViewTags, "U", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		if deps.RemoveUnusedTags == nil {
			return s, nil, true
		}
		return s, deps.RemoveUnusedTags(), true
	})
}

func selectedTag(s State) (api.Tag, bool) {
	return tagsmeta.Selected(s.Tags, s.FilterState(uistate.PaneTags), s.Cursor[uistate.PaneTags])
}

// handleShowTaggedItems opens Issues filtered to the selected tag, or Habits when only habits carry it.
func handleShowTaggedItems(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	tag, ok := selectedTag(s)
	if !ok || deps.SetPaneFilter == nil {
		return s, nil, true
	}
	view, pane := uistate.ViewDefault, uistate.PaneIssues
	if tag.IssueCount == 0 && tag.HabitCount > 0 {
		view, pane = uistate.ViewHabitList, uistate.PaneHabits
	}
	next, _, _ := enterView(s, deps, view)
	s = next.(State)
	return s, deps.SetPaneFilter(pane, "+"+tag.Name), true
}

func handleRenameSelectedTag(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if tag, ok := selectedTag(s); ok && !tag.System && deps.OpenRenameTagDialog != nil {
		deps.OpenRenameTagDialog(&s, tag.Name)
	}
	return s, nil, true
}

// handleCycleTagColor steps the selected tag through no color and the palette.
func handleCycleTagColor(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	tag, ok := selectedTag(s)
	if !ok || deps.SetTagColor == nil {
		return s, nil, true
	}
	options := append([]string{""}, sharedtypes.TagColors...)
	current := ""
	if tag.Color != nil {
		current = *tag.Color
	}
	next := options[nextIndex(current, options, dir)]
	var color *string
	if next != "" {
		color = &next
	}
	return s, deps.SetTagColor(tag.Name, color), true
}

func handleDeleteSelectedTag(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	tag, ok := selectedTag(s)
	if !ok || tag.System || deps.OpenConfirmDeleteTagDialog == nil {
		return s, nil, true
	}
	deps.OpenConfirmDeleteTagDialog(&s, tag)
	return s, nil, true
}
