package controller

import (
	"strings"

	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/bubbles/textinput"
)

const (
	dialogSearchPromptEmoji   = "🔎 "
	dialogTimePromptEmoji     = "🕒 "
	dialogDatePromptEmoji     = "📅 "
	dialogSearchPromptUnicode = "⌕ "
	dialogTimePromptUnicode   = "◷ "
	dialogDatePromptUnicode   = "◫ "
	dialogSearchPromptASCII   = "? "
	dialogTimePromptASCII     = "t "
	dialogDatePromptASCII     = "d "
	dialogValuePrompt         = "> "
)

func plainIssueStatus(status string) string {
	switch status {
	case "in_progress":
		return "in progress"
	case "in_review":
		return "in review"
	default:
		return status
	}
}

func fallback(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func dialogSubmitChord(state State) string {
	return "ctrl+s"
}

func dialogSubmitHint(state State, label string) string {
	return "[" + dialogSubmitChord(state) + "] " + label
}

func renderSelector(theme Theme, state State, label string, active bool) string {
	style := theme.StyleNormal
	if active {
		style = theme.StyleCursor
	}
	return style.Render("[ " + promptGlyphSet(state).Value + label + " ]")
}

func isDialogSubmitKey(state State, key string) bool {
	switch key {
	case "ctrl+s":
		return true
	default:
		return false
	}
}

func withDialogPrompt(input textinput.Model, prompt string) textinput.Model {
	input.Prompt = prompt
	return input
}

type dialogPromptSet struct {
	Search string
	Value  string
	Time   string
	Date   string
}

func promptGlyphSet(state State) dialogPromptSet {
	switch sharedtypes.NormalizePromptGlyphMode(state.PromptGlyphMode) {
	case sharedtypes.PromptGlyphModeUnicode:
		return dialogPromptSet{
			Search: dialogSearchPromptUnicode,
			Value:  dialogValuePrompt,
			Time:   dialogTimePromptUnicode,
			Date:   dialogDatePromptUnicode,
		}
	case sharedtypes.PromptGlyphModeASCII:
		return dialogPromptSet{
			Search: dialogSearchPromptASCII,
			Value:  dialogValuePrompt,
			Time:   dialogTimePromptASCII,
			Date:   dialogDatePromptASCII,
		}
	default:
		return dialogPromptSet{
			Search: dialogSearchPromptEmoji,
			Value:  dialogValuePrompt,
			Time:   dialogTimePromptEmoji,
			Date:   dialogDatePromptEmoji,
		}
	}
}

func withSearchPrompt(state State, input textinput.Model) textinput.Model {
	return withDialogPrompt(input, promptGlyphSet(state).Search)
}

func withTimePrompt(state State, input textinput.Model) textinput.Model {
	return withDialogPrompt(input, promptGlyphSet(state).Time)
}

func withDatePrompt(state State, input textinput.Model) textinput.Model {
	return withDialogPrompt(input, promptGlyphSet(state).Date)
}

func issueDialogHint(state State, submitLabel string) string {
	switch state.Kind {
	case "create_issue_default":
		switch state.FocusIdx {
		case 0, 1:
			return "[type] filter   [left/right] choose   [up/down/tab] move   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		case 3:
			return "[enter] newline   [tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		case 5:
			return "[f2] calendar   [g] today   [tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		default:
			return "[tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		}
	case "create_issue_meta", "edit_issue":
		switch state.FocusIdx {
		case 1:
			return "[enter] newline   [tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		case 3:
			return "[f2] calendar   [g] today   [tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		default:
			return "[tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		}
	default:
		return dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
	}
}
