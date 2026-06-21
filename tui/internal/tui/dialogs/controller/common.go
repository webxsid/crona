package controller

import (
	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/bubbles/textinput"
)

type dialogInputAction int

const (
	dialogActionNone dialogInputAction = iota
	dialogActionCancel
	dialogActionPrimary
	dialogActionFocusNext
	dialogActionFocusPrev
	dialogActionToggle
	dialogActionActivate
	dialogActionMoveUp
	dialogActionMoveDown
	dialogActionMoveLeft
	dialogActionMoveRight
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

func dialogActionForKey(state State, key string) dialogInputAction {
	switch key {
	case "esc":
		return dialogActionCancel
	case "ctrl+s":
		return dialogActionPrimary
	case "tab":
		return dialogActionFocusNext
	case "shift+tab":
		return dialogActionFocusPrev
	case " ":
		return dialogActionToggle
	case "enter":
		return dialogActionActivate
	case "down":
		return dialogActionMoveDown
	case "up":
		return dialogActionMoveUp
	case "left":
		return dialogActionMoveLeft
	case "right":
		return dialogActionMoveRight
	default:
		return dialogActionNone
	}
}

func dialogVerticalMoveDir(action dialogInputAction) int {
	switch action {
	case dialogActionMoveDown:
		return 1
	case dialogActionMoveUp:
		return -1
	default:
		return 0
	}
}

func dialogFocusMoveDir(action dialogInputAction) int {
	switch action {
	case dialogActionFocusNext:
		return 1
	case dialogActionFocusPrev:
		return -1
	default:
		return 0
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
			Search: "⌕ ",
			Value:  dialogValuePrompt,
			Time:   "◷ ",
			Date:   "◫ ",
		}
	case sharedtypes.PromptGlyphModeASCII:
		return dialogPromptSet{
			Search: "? ",
			Value:  dialogValuePrompt,
			Time:   "t ",
			Date:   "d ",
		}
	default:
		return dialogPromptSet{
			Search: "🔎 ",
			Value:  dialogValuePrompt,
			Time:   "🕒 ",
			Date:   "📅 ",
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
			return "[type] filter   [←/→] choose   [↑/↓/tab] move   " + dialogSubmitHint(
				state,
				submitLabel,
			) + "   [esc] cancel"
		case 3:
			return "[enter] newline   [tab] next   " + dialogSubmitHint(
				state,
				submitLabel,
			) + "   [esc] cancel"
		case 5:
			return "[ctrl+e] calendar   [g] today   [tab] next   " + dialogSubmitHint(
				state,
				submitLabel,
			) + "   [esc] cancel"
		default:
			return "[tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		}
	case "create_issue_meta", "edit_issue":
		switch state.FocusIdx {
		case 1:
			return "[enter] newline   [tab] next   " + dialogSubmitHint(
				state,
				submitLabel,
			) + "   [esc] cancel"
		case 3:
			return "[ctrl+e] calendar   [g] today   [tab] next   " + dialogSubmitHint(
				state,
				submitLabel,
			) + "   [esc] cancel"
		default:
			return "[tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		}
	default:
		return dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
	}
}
