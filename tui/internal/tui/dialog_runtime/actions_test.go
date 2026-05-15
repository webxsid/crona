package dialogruntime

import (
	"strings"
	"testing"

	commands "crona/tui/internal/tui/commands"
	dialogstate "crona/tui/internal/tui/dialogs/controller"

	tea "github.com/charmbracelet/bubbletea"
)

func TestResolveTelemetryActionsCallHooks(t *testing.T) {
	cases := []struct {
		name       string
		kind       string
		register   func(*bool) Deps
	}{
		{
			name: "patch telemetry settings",
			kind: "patch_telemetry_settings",
			register: func(called *bool) Deps {
				return Deps{
					PatchTelemetrySettings: func(usageEnabled, errorReportingEnabled bool, restartNow bool) tea.Cmd {
						*called = true
						return func() tea.Msg {
							return nil
						}
					},
				}
			},
		},
		{
			name: "complete onboarding",
			kind: "complete_onboarding",
			register: func(called *bool) Deps {
				return Deps{
					CompleteOnboarding: func(usageEnabled, errorReportingEnabled bool, restartNow bool) tea.Cmd {
						*called = true
						return func() tea.Msg {
							return nil
						}
					},
				}
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			called := false
			deps := tc.register(&called)
			action := dialogstate.Action{
				Kind:             tc.kind,
				UsageTelemetry:   true,
				ErrorReporting:   false,
				RestartAfterSave:  strings.Contains(tc.kind, "complete"),
			}

			cmd := Resolve(action, State{}, deps)
			if cmd == nil {
				t.Fatal("expected command to be returned")
			}
			if msg := cmd(); msg != nil {
				t.Fatalf("expected nil message from stub command, got %#v", msg)
			}
			if !called {
				t.Fatal("expected telemetry hook to be invoked")
			}
		})
	}
}

func TestResolveTelemetryActionsMissingHookReturnsErrorMsg(t *testing.T) {
	cases := []struct {
		name string
		kind string
	}{
		{name: "patch telemetry settings", kind: "patch_telemetry_settings"},
		{name: "complete onboarding", kind: "complete_onboarding"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cmd := Resolve(dialogstate.Action{Kind: tc.kind}, State{}, Deps{})
			if cmd == nil {
				t.Fatal("expected fallback error command")
			}
			msg := cmd()
			errMsg, ok := msg.(commands.ErrMsg)
			if !ok {
				t.Fatalf("expected commands.ErrMsg, got %#v", msg)
			}
			if errMsg.Err == nil {
				t.Fatal("expected an error")
			}
			if want := tc.kind + " runtime hook is not configured"; !strings.Contains(errMsg.Err.Error(), want) {
				t.Fatalf("expected error to contain %q, got %q", want, errMsg.Err.Error())
			}
		})
	}
}
