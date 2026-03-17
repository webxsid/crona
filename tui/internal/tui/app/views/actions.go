package views

import (
	sharedtypes "crona/shared/types"
)

type ActionsState struct {
	View           string
	Pane           string
	ScratchpadOpen bool
	TimerState     string
	IsDevMode      bool
}

func GlobalActions(theme Theme, state ActionsState) []string {
	actions := []string{
		theme.StyleHeader.Render("[tab]") + theme.StyleDim.Render(" next pane"),
	}
	if state.View == "default" || state.View == "daily" {
		actions = append(actions,
			theme.StyleHeader.Render("[a]")+theme.StyleDim.Render(" new"),
			theme.StyleHeader.Render("[c]")+theme.StyleDim.Render(" context"),
		)
	}
	return actions
}

func ContextualActions(theme Theme, state ActionsState) []string {
	if state.View == "session_active" {
		if state.TimerState == "" || state.TimerState == "idle" {
			return []string{
				theme.StyleHeader.Render("[f]") + theme.StyleDim.Render(" focus"),
				theme.StyleHeader.Render("[Z]") + theme.StyleDim.Render(" stashes"),
			}
		}
		return []string{
			theme.StyleHeader.Render("[p/r]") + theme.StyleDim.Render(" pause/resume"),
			theme.StyleHeader.Render("[x]") + theme.StyleDim.Render(" end"),
			theme.StyleHeader.Render("[z]") + theme.StyleDim.Render(" stash"),
			theme.StyleHeader.Render("[s/A]") + theme.StyleDim.Render(" issue"),
		}
	}
	if state.View == "session_history" {
		return []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" details"),
			theme.StyleHeader.Render("[f]") + theme.StyleDim.Render(" focus"),
			theme.StyleHeader.Render("[Z]") + theme.StyleDim.Render(" stashes"),
		}
	}
	if state.View == "wellbeing" {
		return []string{
			theme.StyleHeader.Render("[,/.]") + theme.StyleDim.Render(" date"),
			theme.StyleHeader.Render("[g]") + theme.StyleDim.Render(" today"),
			theme.StyleHeader.Render("[a/e]") + theme.StyleDim.Render(" check-in"),
			theme.StyleHeader.Render("[d]") + theme.StyleDim.Render(" delete"),
		}
	}

	switch state.Pane {
	case "repos", "streams":
		return []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" view"),
			theme.StyleHeader.Render("[c]") + theme.StyleDim.Render(" checkout"),
			theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" new"),
			theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit"),
			theme.StyleHeader.Render("[d]") + theme.StyleDim.Render(" delete"),
		}
	case "issues":
		if state.View == "daily" {
			return []string{
				theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" view"),
				theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" new"),
				theme.StyleHeader.Render("[c]") + theme.StyleDim.Render(" context"),
				theme.StyleHeader.Render("[f]") + theme.StyleDim.Render(" focus"),
				theme.StyleHeader.Render("[s]") + theme.StyleDim.Render(" status"),
				theme.StyleHeader.Render("[D]") + theme.StyleDim.Render(" due date"),
			}
		}
		return []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" view"),
			theme.StyleHeader.Render("[f]") + theme.StyleDim.Render(" focus"),
			theme.StyleHeader.Render("[s]") + theme.StyleDim.Render(" status"),
			theme.StyleHeader.Render("[D]") + theme.StyleDim.Render(" due date"),
			theme.StyleHeader.Render("[e/d]") + theme.StyleDim.Render(" edit/delete"),
		}
	case "habits":
		if state.View == "daily" {
			return []string{
				theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" view"),
				theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" new"),
				theme.StyleHeader.Render("[x]") + theme.StyleDim.Render(" toggle"),
				theme.StyleHeader.Render("[F]") + theme.StyleDim.Render(" fail"),
				theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" log"),
			}
		}
		return []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" view"),
			theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" new"),
			theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit"),
			theme.StyleHeader.Render("[d]") + theme.StyleDim.Render(" delete"),
		}
	case "scratchpads":
		if state.ScratchpadOpen {
			actions := []string{
				theme.StyleHeader.Render("[h/l]") + theme.StyleDim.Render(" switch"),
				theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit"),
				theme.StyleHeader.Render("[esc]") + theme.StyleDim.Render(" close"),
			}
			if state.TimerState != "" && state.TimerState != "idle" {
				actions = append(actions, theme.StyleHeader.Render("[s/A]")+theme.StyleDim.Render(" issue"))
			}
			return actions
		}
		actions := []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" open"),
			theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" new"),
			theme.StyleHeader.Render("[d]") + theme.StyleDim.Render(" delete"),
		}
		if state.TimerState != "" && state.TimerState != "idle" {
			actions = append(actions, theme.StyleHeader.Render("[s/A]")+theme.StyleDim.Render(" issue"))
		}
		return actions
	case "ops":
		return []string{
			theme.StyleHeader.Render("[+]") + theme.StyleDim.Render(" more"),
			theme.StyleHeader.Render("[-]") + theme.StyleDim.Render(" less"),
		}
	case "settings":
		return []string{
			theme.StyleHeader.Render("[h/l]") + theme.StyleDim.Render(" change"),
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" toggle/advance"),
		}
	}
	return nil
}

func PaneActions(theme Theme, state ActionsState) []string {
	actions := append([]string{}, GlobalActions(theme, state)...)
	actions = append(actions, ContextualActions(theme, state)...)
	return actions
}

func SettingsItemLabels(settings *sharedtypes.CoreSettings) []string {
	if settings == nil {
		return nil
	}
	return []string{
		"Timer Mode",
		"Breaks Enabled",
		"Work Duration",
		"Short Break",
		"Long Break",
		"Long Break Enabled",
		"Cycles Before Long Break",
		"Auto Start Breaks",
		"Auto Start Work",
		"Repo Sort",
		"Stream Sort",
		"Issue Sort",
	}
}
