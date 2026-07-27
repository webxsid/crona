package controller

import (
	"testing"

	sharedtypes "crona/shared/types"
)

func TestIssueDialogPromptsUseEmojiGlyphsByDefault(t *testing.T) {
	state := OpenCreateIssueDefault(State{})

	if got := state.Inputs[0].Prompt; got != dialogSearchPromptEmoji {
		t.Fatalf("expected repo search prompt %q, got %q", dialogSearchPromptEmoji, got)
	}
	if got := state.Inputs[1].Prompt; got != dialogSearchPromptEmoji {
		t.Fatalf("expected stream search prompt %q, got %q", dialogSearchPromptEmoji, got)
	}
	if got := state.Inputs[3].Prompt; got != dialogTimePromptEmoji {
		t.Fatalf("expected estimate prompt %q, got %q", dialogTimePromptEmoji, got)
	}
	if got := state.Inputs[4].Prompt; got != dialogDatePromptEmoji {
		t.Fatalf("expected due date prompt %q, got %q", dialogDatePromptEmoji, got)
	}
}

func TestIssueDialogPromptsRespectUnicodeMode(t *testing.T) {
	state := OpenCreateIssueDefault(State{PromptGlyphMode: sharedtypes.PromptGlyphModeUnicode})

	if got := state.Inputs[0].Prompt; got != dialogSearchPromptUnicode {
		t.Fatalf("expected repo search prompt %q, got %q", dialogSearchPromptUnicode, got)
	}
	if got := state.Inputs[3].Prompt; got != dialogTimePromptUnicode {
		t.Fatalf("expected estimate prompt %q, got %q", dialogTimePromptUnicode, got)
	}
	if got := state.Inputs[4].Prompt; got != dialogDatePromptUnicode {
		t.Fatalf("expected due date prompt %q, got %q", dialogDatePromptUnicode, got)
	}
}

func TestManualSessionPromptsRespectASCIIMode(t *testing.T) {
	state := OpenManualSession(
		State{PromptGlyphMode: sharedtypes.PromptGlyphModeASCII},
		42,
		"Fix prompts",
		nil,
		"2026-04-30",
	)

	if got := state.Inputs[1].Prompt; got != dialogDatePromptASCII {
		t.Fatalf("expected manual session date prompt %q, got %q", dialogDatePromptASCII, got)
	}
	for _, idx := range []int{2, 3, 4, 5} {
		if got := state.Inputs[idx].Prompt; got != dialogTimePromptASCII {
			t.Fatalf(
				"expected manual session input %d prompt %q, got %q",
				idx,
				dialogTimePromptASCII,
				got,
			)
		}
	}
}

func TestAlertReminderTimePromptUsesConfiguredGlyphs(t *testing.T) {
	state := OpenCreateAlertReminder(
		State{PromptGlyphMode: sharedtypes.PromptGlyphModeUnicode},
		sharedtypes.AlertReminderKindCheckIn,
	)

	if got := state.Inputs[1].Prompt; got != dialogTimePromptUnicode {
		t.Fatalf("expected reminder time prompt %q, got %q", dialogTimePromptUnicode, got)
	}
}

func TestAlertReminderDefaultsDependOnKind(t *testing.T) {
	checkIn := OpenCreateAlertReminder(State{}, sharedtypes.AlertReminderKindCheckIn)
	if checkIn.ReminderKind != sharedtypes.AlertReminderKindCheckIn ||
		checkIn.Inputs[1].Value() != "20:00" {
		t.Fatalf(
			"unexpected check-in reminder defaults: kind=%q time=%q",
			checkIn.ReminderKind,
			checkIn.Inputs[1].Value(),
		)
	}

	plan := OpenCreateAlertReminder(State{}, sharedtypes.AlertReminderKindDailyPlan)
	if plan.ReminderKind != sharedtypes.AlertReminderKindDailyPlan ||
		plan.Inputs[1].Value() != "09:00" {
		t.Fatalf(
			"unexpected daily-plan reminder defaults: kind=%q time=%q",
			plan.ReminderKind,
			plan.Inputs[1].Value(),
		)
	}
}

func TestEditAlertReminderPreservesKindAndTime(t *testing.T) {
	state := OpenEditAlertReminder(State{}, sharedtypes.AlertReminder{
		ID:           "plan",
		Kind:         sharedtypes.AlertReminderKindDailyPlan,
		ScheduleType: sharedtypes.AlertReminderScheduleDaily,
		TimeHHMM:     "08:30",
	})

	if state.ReminderKind != sharedtypes.AlertReminderKindDailyPlan ||
		state.Inputs[1].Value() != "08:30" {
		t.Fatalf(
			"unexpected edited reminder state: kind=%q time=%q",
			state.ReminderKind,
			state.Inputs[1].Value(),
		)
	}
}

func TestRenderSelectorShowsValueMarker(t *testing.T) {
	rendered := renderSelector(
		Theme{},
		State{PromptGlyphMode: sharedtypes.PromptGlyphModeASCII},
		"Work",
		false,
	)
	if rendered != "[ > Work ]" {
		t.Fatalf("expected selector output with value marker, got %q", rendered)
	}
}
