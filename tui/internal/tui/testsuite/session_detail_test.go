package testsuite

import (
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	app "crona/tui/internal/tui/model"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSessionDetailEditHotkeyDoesNothingWhenManageHistoryIsDisabled(t *testing.T) {
	detail := &api.SessionDetail{
		SessionHistoryEntry: sharedtypes.SessionHistoryEntry{
			Session: sharedtypes.Session{
				ID: "sess-1",
			},
			ParsedNotes: sharedtypes.ParsedSessionNotes{
				sharedtypes.SessionNoteSectionCommit: "Refine alerts backend",
			},
		},
	}

	model := app.NewSessionDetailModel(detail)
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	updated := next.(app.Model)

	if got := updated.DialogKind(); got != "" {
		t.Fatalf(
			"expected edit dialog to remain closed when history management is disabled, got %q",
			got,
		)
	}
}
