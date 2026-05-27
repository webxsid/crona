package viewchrome

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

import uistate "crona/tui/internal/tui/state"

type ActionsState struct {
	View                   string
	Pane                   string
	TimerState             string
	TimerSegment           string
	TimerNextSegment       string
	StructuredTimer        bool
	HardLimitActive        bool
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
		actions = append(
			actions,
			theme.StyleHeader.Render("[f9]")+theme.StyleDim.Render(" beta support"),
		)
	}
	if tabLabel := tabActionLabel(state); tabLabel != "" {
		actions = append(
			actions,
			theme.StyleHeader.Render("[tab]")+theme.StyleDim.Render(" "+tabLabel),
		)
	}
	if supportsGlobalCreate(state.View) {
		actions = append(actions, theme.StyleHeader.Render("[a]")+theme.StyleDim.Render(" new"))
	}
	if supportsGlobalContext(state.View) {
		actions = append(actions, theme.StyleHeader.Render("[c]")+theme.StyleDim.Render(" context"))
	}
	if supportsGlobalExport(state.View) {
		actions = append(actions, theme.StyleHeader.Render("[E]")+theme.StyleDim.Render(" export"))
	}
	if state.View == "daily" || state.View == "wellbeing" {
		actions = append(actions, checkInAction(theme))
		actions = append(actions, awayAction(theme, false))
	}
	if state.View == "away" && state.AwayModeActive {
		actions = append(actions, awayAction(theme, true))
	}
	return actions
}

func ContextualActions(theme Theme, state ActionsState) []string {
	if state.RestModeActive && state.View == "away" {
		if state.AwayModeActive {
			return []string{
				awayAction(theme, true),
			}
		}
		return nil
	}
	if state.View == "session_active" {
		if state.TimerState == "" || state.TimerState == "idle" {
			return []string{
				theme.StyleHeader.Render("[f]") + theme.StyleDim.Render(" start timer"),
				theme.StyleHeader.Render("[Z]") + theme.StyleDim.Render(" stashes"),
			}
		}
		if state.HardLimitActive {
			return []string{
				theme.StyleHeader.Render("[x]") + theme.StyleDim.Render(" commit"),
				theme.StyleHeader.Render("[z]") + theme.StyleDim.Render(" stash"),
				theme.StyleHeader.Render("[i]") + theme.StyleDim.Render(" context"),
				theme.StyleHeader.Render("[s/A]") + theme.StyleDim.Render(" issue"),
			}
		}
		if state.TimerState == "ready" {
			return []string{
				theme.StyleHeader.Render(
					"[r]",
				) + theme.StyleDim.Render(
					" start "+timerActionSegmentLabel(state.TimerNextSegment),
				),
				theme.StyleHeader.Render("[x]") + theme.StyleDim.Render(" end"),
				theme.StyleHeader.Render("[z]") + theme.StyleDim.Render(" stash"),
				theme.StyleHeader.Render("[i]") + theme.StyleDim.Render(" context"),
				theme.StyleHeader.Render("[s/A]") + theme.StyleDim.Render(" issue"),
			}
		}
		return []string{
			theme.StyleHeader.Render("[p]") + theme.StyleDim.Render(" pause"),
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
			theme.StyleHeader.Render("[Z]") + theme.StyleDim.Render(" stashes"),
		}
	}
	if state.View == "habit_history" {
		return []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" details"),
		}
	}
	if state.View == "wellbeing" {
		return []string{
			theme.StyleHeader.Render("[,/.]") + theme.StyleDim.Render(" date"),
			theme.StyleHeader.Render("[g]") + theme.StyleDim.Render(" today"),
			theme.StyleHeader.Render("[+/-]") + theme.StyleDim.Render(" window"),
			theme.StyleHeader.Render("[0]") + theme.StyleDim.Render(" reset window"),
			checkInAction(theme),
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
			theme.StyleHeader.Render("[c/space]") + theme.StyleDim.Render(" change"),
			theme.StyleHeader.Render("[R]") + theme.StyleDim.Render(" rescan tools"),
		}
		if state.TimerState == "" || state.TimerState == "idle" {
			actions = append(
				actions,
				theme.StyleHeader.Render("[r]")+theme.StyleDim.Render(" reset selected"),
			)
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
			actions = append(
				actions,
				theme.StyleHeader.Render("[i]")+theme.StyleDim.Render(" install"),
			)
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
		contextLabel := " checkout"
		if state.View != "meta" {
			contextLabel = " context"
		}
		return []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" view"),
			theme.StyleHeader.Render("[c]") + theme.StyleDim.Render(contextLabel),
			theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" new"),
			theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit"),
			theme.StyleHeader.Render("[d]") + theme.StyleDim.Render(" delete"),
		}
	case "issues":
		timerIdle := state.TimerState == "" || state.TimerState == "idle"
		if state.View == "daily" {
			actions := []string{
				theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" view"),
				theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" new"),
				theme.StyleHeader.Render("[s]") + theme.StyleDim.Render(" status"),
				theme.StyleHeader.Render("[D]") + theme.StyleDim.Render(" due date"),
				theme.StyleHeader.Render("[P]") + theme.StyleDim.Render(" pin"),
			}
			if timerIdle {
				actions = append(
					actions,
					theme.StyleHeader.Render("[f]")+theme.StyleDim.Render(" start timer"),
					theme.StyleHeader.Render("[m]")+theme.StyleDim.Render(" log"),
					theme.StyleHeader.Render("[e]")+theme.StyleDim.Render(" edit"),
				)
			}
			return actions
		}
		actions := []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" view"),
			theme.StyleHeader.Render("[s]") + theme.StyleDim.Render(" status"),
			theme.StyleHeader.Render("[D]") + theme.StyleDim.Render(" due date"),
			theme.StyleHeader.Render("[P]") + theme.StyleDim.Render(" pin"),
		}
		if timerIdle {
			actions = append(
				actions,
				theme.StyleHeader.Render("[f]")+theme.StyleDim.Render(" start timer"),
				theme.StyleHeader.Render("[m]")+theme.StyleDim.Render(" log"),
				theme.StyleHeader.Render("[e/d]")+theme.StyleDim.Render(" edit/delete"),
			)
		}
		return actions
	case "habits":
		if state.View == "daily" {
			actions := []string{
				theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" view"),
				theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" new"),
				theme.StyleHeader.Render("[x]") + theme.StyleDim.Render(" toggle"),
				theme.StyleHeader.Render("[F]") + theme.StyleDim.Render(" fail"),
				theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit"),
				theme.StyleHeader.Render("[d]") + theme.StyleDim.Render(" delete"),
			}
			if state.TimerState == "" || state.TimerState == "idle" {
				actions = append(actions, theme.StyleHeader.Render("[m]")+theme.StyleDim.Render(" log"))
			}
			return actions
		}
		return []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" view"),
			theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" new"),
			theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit"),
			theme.StyleHeader.Render("[m]") + theme.StyleDim.Render(" log"),
			theme.StyleHeader.Render("[d]") + theme.StyleDim.Render(" delete"),
		}
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

func checkInAction(theme Theme) string {
	return theme.StyleHeader.Render("[w]") + theme.StyleDim.Render(" check-in")
}

func awayAction(theme Theme, disabled bool) string {
	if disabled {
		return theme.StyleHeader.Render("[W]") + theme.StyleDim.Render(" disable away")
	}
	return theme.StyleHeader.Render("[W]") + theme.StyleDim.Render(" away")
}

func timerActionSegmentLabel(segment string) string {
	switch strings.TrimSpace(segment) {
	case "short_break":
		return "short break"
	case "long_break":
		return "long break"
	case "work":
		return "work"
	default:
		return "next"
	}
}

func supportsGlobalCreate(view string) bool {
	return view == "default" || view == "daily"
}

func supportsGlobalContext(view string) bool {
	switch view {
	case "default",
		"daily",
		"meta",
		"rollup",
		"wellbeing",
		"reports",
		"ops",
		"session_history",
		"habit_history":
		return true
	default:
		return false
	}
}

func supportsGlobalExport(view string) bool {
	switch view {
	case "default", "daily", "meta", "wellbeing", "reports":
		return true
	default:
		return false
	}
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
