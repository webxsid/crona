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
	state := OpenManualSession(State{PromptGlyphMode: sharedtypes.PromptGlyphModeASCII}, 42, "Fix prompts", nil, "2026-04-30")

	if got := state.Inputs[1].Prompt; got != dialogDatePromptASCII {
		t.Fatalf("expected manual session date prompt %q, got %q", dialogDatePromptASCII, got)
	}
	for _, idx := range []int{2, 3, 4, 5} {
		if got := state.Inputs[idx].Prompt; got != dialogTimePromptASCII {
			t.Fatalf("expected manual session input %d prompt %q, got %q", idx, dialogTimePromptASCII, got)
		}
	}
}

func TestAlertReminderTimePromptUsesConfiguredGlyphs(t *testing.T) {
	state := OpenCreateAlertReminder(State{PromptGlyphMode: sharedtypes.PromptGlyphModeUnicode})

	if got := state.Inputs[1].Prompt; got != dialogTimePromptUnicode {
		t.Fatalf("expected reminder time prompt %q, got %q", dialogTimePromptUnicode, got)
	}
}

func TestRenderSelectorShowsValueMarker(t *testing.T) {
	rendered := renderSelector(Theme{}, State{PromptGlyphMode: sharedtypes.PromptGlyphModeASCII}, "Work", false)
	if rendered != "[ > Work ]" {
		t.Fatalf("expected selector output with value marker, got %q", rendered)
	}
}
