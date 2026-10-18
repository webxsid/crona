package input

import tea "github.com/charmbracelet/bubbletea"

// supportActions are Support's actions in list order; each letter key and [enter] on the
// selected row run the same command.
var supportActions = []struct {
	key string
	run func(State, Deps) tea.Cmd
}{
	{"o", func(_ State, deps Deps) tea.Cmd { return deps.OpenSupportFeedbackRoadmapURL() }},
	{"r", func(_ State, deps Deps) tea.Cmd { return deps.OpenSupportReleasesURL() }},
	{"g", func(_ State, deps Deps) tea.Cmd { return deps.OpenSupportDocumentationURL() }},
	{"c", func(s State, deps Deps) tea.Cmd { return deps.CopySupportDiagnostics(s) }},
	{"b", func(s State, deps Deps) tea.Cmd { return deps.GenerateSupportBundle(s) }},
}

// SupportActionKeys are the action keys in list order, for the view to match.
func SupportActionKeys() []string {
	keys := make([]string, len(supportActions))
	for i, action := range supportActions {
		keys[i] = action.key
	}
	return keys
}
