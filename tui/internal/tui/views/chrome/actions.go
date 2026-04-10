package viewchrome

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

import uistate "crona/tui/internal/tui/state"

type ActionsState struct {
	View                   string
	Pane                   string
	ScratchpadOpen         bool
	TimerState             string
	RestModeActive         bool
	AwayModeActive         bool
	IsDevMode              bool
	IsBetaBuild            bool
	UpdateVisible          bool
	UpdateInstallAvailable bool
}

func GlobalActions(theme Theme, state ActionsState) []string {
	actions := []string{
		theme.StyleHeader.Render("[v]") + theme.StyleDim.Render(" views"),
	}
	if state.IsBetaBuild {
		actions = append(actions, theme.StyleHeader.Render("[f9]")+theme.StyleDim.Render(" beta support"))
	}
	if tabLabel := tabActionLabel(state); tabLabel != "" {
		actions = append(actions, theme.StyleHeader.Render("[tab]")+theme.StyleDim.Render(" "+tabLabel))
	}
	if state.View == "default" || state.View == "daily" {
		actions = append(actions,
			theme.StyleHeader.Render("[a]")+theme.StyleDim.Render(" new"),
			theme.StyleHeader.Render("[c]")+theme.StyleDim.Render(" context"),
		)
	}
	if state.View == "daily" {
		actions = append(actions,
			theme.StyleHeader.Render("[E]")+theme.StyleDim.Render(" export"),
			theme.StyleHeader.Render("[w]")+theme.StyleDim.Render(" away"),
		)
	}
	if state.UpdateVisible {
		actions = append(actions, theme.StyleHeader.Render("[u]")+theme.StyleDim.Render(" updates"))
	}
	return actions
}

func ContextualActions(theme Theme, state ActionsState) []string {
	if state.RestModeActive && state.View == "away" {
		if state.AwayModeActive {
			return []string{
				theme.StyleHeader.Render("[w]") + theme.StyleDim.Render(" disable away"),
			}
		}
		return nil
	}
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
			theme.StyleHeader.Render("[i]") + theme.StyleDim.Render(" context"),
			theme.StyleHeader.Render("[s/A]") + theme.StyleDim.Render(" issue"),
		}
	}
	if state.View == "session_history" {
		if state.RestModeActive {
			return []string{
				theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" details"),
			}
		}
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
	if state.View == "rollup" {
		return []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" day details"),
			theme.StyleHeader.Render("[S/E]") + theme.StyleDim.Render(" calendar"),
			theme.StyleHeader.Render("[h/l]") + theme.StyleDim.Render(" start"),
			theme.StyleHeader.Render("[,/.]") + theme.StyleDim.Render(" end"),
			theme.StyleHeader.Render("[g]") + theme.StyleDim.Render(" weekly"),
		}
	}
	if state.View == "config" {
		actions := []string{
			theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit/open"),
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" details"),
			theme.StyleHeader.Render("[space]") + theme.StyleDim.Render(" change"),
			theme.StyleHeader.Render("[R]") + theme.StyleDim.Render(" rescan tools"),
		}
		if state.TimerState == "" || state.TimerState == "idle" {
			actions = append(actions, theme.StyleHeader.Render("[r]")+theme.StyleDim.Render(" reset selected"))
		}
		return actions
	}
	if state.View == "reports" {
		return []string{
			theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit"),
			theme.StyleHeader.Render("[o]") + theme.StyleDim.Render(" open"),
			theme.StyleHeader.Render("[d]") + theme.StyleDim.Render(" delete"),
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" details"),
		}
	}
	if state.View == "updates" {
		actions := []string{
			theme.StyleHeader.Render("[r]") + theme.StyleDim.Render(" check now"),
			theme.StyleHeader.Render("[o]") + theme.StyleDim.Render(" open release"),
			theme.StyleHeader.Render("[U]") + theme.StyleDim.Render(" dismiss"),
		}
		if state.UpdateInstallAvailable {
			actions = append(actions, theme.StyleHeader.Render("[i]")+theme.StyleDim.Render(" install"))
		} else {
			actions = append(actions, theme.StyleDim.Render("[i] install unavailable"))
		}
		return actions
	}
	if state.View == "support" {
		return []string{
			theme.StyleHeader.Render("[o]") + theme.StyleDim.Render(" report bug"),
			theme.StyleHeader.Render("[d]") + theme.StyleDim.Render(" discussions"),
			theme.StyleHeader.Render("[r]") + theme.StyleDim.Render(" releases"),
			theme.StyleHeader.Render("[g]") + theme.StyleDim.Render(" roadmap"),
			theme.StyleHeader.Render("[c]") + theme.StyleDim.Render(" copy diagnostics"),
			theme.StyleHeader.Render("[b]") + theme.StyleDim.Render(" bundle"),
		}
	}
	if state.View == "alerts" {
		return []string{
			theme.StyleHeader.Render("[h/l]") + theme.StyleDim.Render(" change"),
			theme.StyleHeader.Render("[space]") + theme.StyleDim.Render(" toggle"),
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" edit/run"),
			theme.StyleHeader.Render("[d/x]") + theme.StyleDim.Render(" delete"),
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
				theme.StyleHeader.Render("[m]") + theme.StyleDim.Render(" log"),
				theme.StyleHeader.Render("[s]") + theme.StyleDim.Render(" status"),
				theme.StyleHeader.Render("[D]") + theme.StyleDim.Render(" due date"),
			}
		}
		return []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" view"),
			theme.StyleHeader.Render("[f]") + theme.StyleDim.Render(" focus"),
			theme.StyleHeader.Render("[m]") + theme.StyleDim.Render(" log"),
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
				theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit"),
				theme.StyleHeader.Render("[m]") + theme.StyleDim.Render(" log"),
				theme.StyleHeader.Render("[d]") + theme.StyleDim.Render(" delete"),
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
				theme.StyleHeader.Render("[o]") + theme.StyleDim.Render(" open"),
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
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" edit/toggle/confirm"),
		}
	case "alerts":
		return []string{
			theme.StyleHeader.Render("[h/l]") + theme.StyleDim.Render(" change"),
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" toggle/test"),
		}
	}
	return nil
}

func PaneActions(theme Theme, state ActionsState) []string {
	return ContextualActions(theme, state)
}

func DedupeActionKeys(actions []string, suppress []string) []string {
	if len(actions) == 0 || len(suppress) == 0 {
		return actions
	}
	blocked := map[string]struct{}{}
	for _, action := range suppress {
		for _, token := range actionTokens(action) {
			blocked[token] = struct{}{}
		}
	}
	out := make([]string, 0, len(actions))
	for _, action := range actions {
		tokens := actionTokens(action)
		if len(tokens) == 0 {
			out = append(out, action)
			continue
		}
		duplicate := false
		for _, token := range tokens {
			if _, ok := blocked[token]; ok {
				duplicate = true
				break
			}
		}
		if !duplicate {
			out = append(out, action)
		}
	}
	return out
}

func tabActionLabel(state ActionsState) string {
	if state.View == string(uistate.ViewDefault) && state.Pane == string(uistate.PaneIssues) {
		return "open/resolved"
	}
	if len(uistate.ViewPanes(uistate.View(state.View))) > 1 {
		return "next pane"
	}
	return ""
}

func actionTokens(action string) []string {
	plain := ansi.Strip(action)
	out := []string{}
	for {
		start := strings.IndexByte(plain, '[')
		if start < 0 {
			break
		}
		plain = plain[start+1:]
		end := strings.IndexByte(plain, ']')
		if end < 0 {
			break
		}
		token := strings.TrimSpace(plain[:end])
		if token != "" {
			out = append(out, token)
		}
		plain = plain[end+1:]
	}
	return out
}
