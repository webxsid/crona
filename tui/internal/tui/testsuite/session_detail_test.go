package testsuite

import (
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	app "crona/tui/internal/tui/app"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSessionDetailOverlayOpensAmendDialog(t *testing.T) {
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

	if got := updated.DialogKind(); got != "amend_session" {
		t.Fatalf("expected amend_session dialog, got %q", got)
	}
	if got := updated.DialogSessionID(); got != detail.ID {
		t.Fatalf("expected dialog session id %q, got %q", detail.ID, got)
	}
	if got := updated.DialogInputValue(0); got != "Refine alerts backend" {
		t.Fatalf("expected commit message to be prefilled, got %q", got)
	}
}
