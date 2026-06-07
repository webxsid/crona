package input

import (
	"testing"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleOpenCheckInUsesDashboardDate(t *testing.T) {
	called := false
	state := State{
		ActiveView:    uistate.ViewDaily,
		DashboardDate: "2026-05-27",
	}

	_, _, handled := handleOpenCheckIn(state, Deps{
		OpenCheckInDialog: func(s *State) bool {
			called = true
			if s.DashboardDate != "2026-05-27" {
				t.Fatalf("expected daily check-in opener to preserve dashboard date, got %q", s.DashboardDate)
			}
			return true
		},
	})
	if !handled {
		t.Fatal("expected daily w to be handled")
	}
	if !called {
		t.Fatal("expected daily w to open the check-in dialog")
	}
}

func TestHandleOpenCheckInUsesWellbeingDate(t *testing.T) {
	called := false
	state := State{
		ActiveView:    uistate.ViewWellbeing,
		WellbeingDate: "2026-05-28",
	}

	_, _, handled := handleOpenCheckIn(state, Deps{
		OpenCheckInDialog: func(s *State) bool {
			called = true
			if s.WellbeingDate != "2026-05-28" {
				t.Fatalf("expected wellbeing check-in opener to preserve wellbeing date, got %q", s.WellbeingDate)
			}
			return true
		},
	})
	if !handled {
		t.Fatal("expected wellbeing w to be handled")
	}
	if !called {
		t.Fatal("expected wellbeing w to open the check-in dialog")
	}
}

func TestHandleQuestionOpensHelpDialog(t *testing.T) {
	called := false
	state := State{
		ActiveView: uistate.ViewDaily,
	}
	next, _ := Handle(state, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, Deps{
		OpenHelpDialog: func(s *State) bool {
			called = true
			if s.Dialog != "" {
				t.Fatalf("expected help opener to receive a clean state, got dialog %q", s.Dialog)
			}
			s.Dialog = "view_entity"
			return true
		},
	})
	if !called {
		t.Fatal("expected help dialog opener to be called")
	}
	if next.Dialog != "view_entity" {
		t.Fatalf("expected help dialog to set view_entity, got %q", next.Dialog)
	}
}

func TestHandleStartFilterAllowsMomentumCards(t *testing.T) {
	called := false
	state := State{
		ActivePane: uistate.PaneMomentumCards,
	}
	_, _, handled := handleStartFilter(state, Deps{
		StartFilterEdit: func(s *State, pane uistate.Pane) {
			called = true
			if pane != uistate.PaneMomentumCards {
				t.Fatalf("expected momentum cards filter pane, got %q", pane)
			}
		},
	})
	if !handled {
		t.Fatal("expected / to start filtering momentum cards")
	}
	if !called {
		t.Fatal("expected filter editor to be started for momentum cards")
	}
}

func TestHandleOpenCheckInRejectsOtherViews(t *testing.T) {
	called := false
	_, _, handled := handleOpenCheckIn(State{ActiveView: uistate.ViewDefault}, Deps{
		OpenCheckInDialog: func(*State) bool {
			called = true
			return true
		},
	})
	if handled {
		t.Fatal("expected non-daily check-in binding to fall through")
	}
	if called {
		t.Fatal("did not expect check-in dialog to open outside daily or wellbeing views")
	}
}

func TestRouterRoutesWellbeingWToCheckInAndWAway(t *testing.T) {
	checkInCalled := false
	checkInState := State{
		ActiveView:    uistate.ViewWellbeing,
		WellbeingDate: "2026-05-28",
	}

	next, _ := Handle(checkInState, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}}, Deps{
		OpenCheckInDialog: func(s *State) bool {
			checkInCalled = true
			if s.WellbeingDate != "2026-05-28" {
				t.Fatalf("expected wellbeing w to preserve wellbeing date, got %q", s.WellbeingDate)
			}
			s.Dialog = "checkin"
			return true
		},
	})
	if !checkInCalled {
		t.Fatal("expected wellbeing w to open the check-in dialog")
	}
	if next.Dialog != "checkin" {
		t.Fatalf("expected wellbeing w to route to check-in, got dialog %q", next.Dialog)
	}

	awayState := State{
		ActiveView: uistate.ViewWellbeing,
		Settings:   &api.CoreSettings{AwayModeEnabled: false},
	}
	next, _ = Handle(awayState, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'W'}}, Deps{
		PatchSetting: func(sharedtypes.CoreSettingsKey, any, int64, int64, string) tea.Cmd { return nil },
		SetStatus:    func(*State, string, bool) tea.Cmd { return nil },
	})
	if !next.Settings.AwayModeEnabled {
		t.Fatal("expected wellbeing W to enable away mode")
	}
}

func TestRouterIgnoresWDuringConfiguredRestButStillAllowsManualAwayToggle(t *testing.T) {
	restDate := time.Now().Format("2006-01-02")
	restState := State{
		ActiveView:    uistate.ViewDaily,
		DashboardDate: restDate,
		Settings: &api.CoreSettings{
			AwayModeEnabled:   false,
			RestSpecificDates: []string{restDate},
		},
	}
	next, _ := Handle(restState, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'W'}}, Deps{
		PatchSetting: func(sharedtypes.CoreSettingsKey, any, int64, int64, string) tea.Cmd { return nil },
		SetStatus:    func(*State, string, bool) tea.Cmd { return nil },
	})
	if next.Settings.AwayModeEnabled {
		t.Fatal("expected W to no-op during configured rest")
	}

	manualAwayState := State{
		ActiveView:    uistate.ViewDaily,
		DashboardDate: restDate,
		Settings: &api.CoreSettings{
			AwayModeEnabled:   true,
			RestSpecificDates: []string{restDate},
		},
	}
	next, _ = Handle(manualAwayState, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'W'}}, Deps{
		PatchSetting: func(sharedtypes.CoreSettingsKey, any, int64, int64, string) tea.Cmd { return nil },
		SetStatus:    func(*State, string, bool) tea.Cmd { return nil },
	})
	if next.Settings.AwayModeEnabled {
		t.Fatal("expected W to disable manual away mode even during protected rest")
	}
}
