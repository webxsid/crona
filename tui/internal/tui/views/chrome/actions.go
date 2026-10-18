package viewchrome

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	"github.com/charmbracelet/x/ansi"
)

import uistate "crona/tui/internal/tui/state"

type ActionsState struct {
	View                      string
	Pane                      string
	TimerState                string
	TimerSegment              string
	TimerNextSegment          string
	HardLimitActive           bool
	HardLimitKind             sharedtypes.TimerHardLimitKind
	RestModeActive            bool
	AwayModeActive            bool
	IsDevMode                 bool
	IsBetaBuild               bool
	UpdateVisible             bool
	UpdateInstallAvailable    bool
	UpdateDiagnosticsExpanded bool
	UpdateScriptDeprecated    bool
	UpdateCommand             string
	ManageHistory             bool
	SelectedIssueAdhoc        bool
}

func segmentPtr(value string) *sharedtypes.SessionSegmentType {
	if value == "" {
		return nil
	}
	segment := sharedtypes.SessionSegmentType(value)
	return &segment
}

// NavigationActions move between tabs; they live on the help sheet, not the bottom row.
func NavigationActions(theme Theme, state ActionsState) []string {
	actions := []string{
		theme.StyleHeader.Render("[1-5]") + theme.StyleDim.Render(" tabs"),
		theme.StyleHeader.Render("[[/]]") + theme.StyleDim.Render(" sub-tabs"),
	}
	if state.IsBetaBuild {
		actions = append(
			actions,
			theme.StyleHeader.Render("[f9]")+theme.StyleDim.Render(" beta support"),
		)
	}
	return actions
}

// DateActions are a dated view's date keys, which the header's date row shows; the help
// sheet lists them too.
func DateActions(theme Theme, view string) []string {
	actions := []string{
		theme.StyleHeader.Render("[,/.]") + theme.StyleDim.Render(" date"),
		theme.StyleHeader.Render("[g]") + theme.StyleDim.Render(" today"),
	}
	switch view {
	case "daily":
		return append(actions,
			theme.StyleHeader.Render("[S]")+theme.StyleDim.Render(" day summary"),
			theme.StyleHeader.Render("[C]")+theme.StyleDim.Render(" calendar"),
		)
	case "review":
		return append(actions,
			theme.StyleHeader.Render("[E]")+theme.StyleDim.Render(" calendar"),
			theme.StyleHeader.Render("[+/-/0]")+theme.StyleDim.Render(" window"),
		)
	case "momentum":
		return append(actions, theme.StyleHeader.Render("[+/-/0]")+theme.StyleDim.Render(" window"))
	}
	return nil
}

// WithoutDateKeys drops the date keys a dated view's date row already shows.
func WithoutDateKeys(actions []string) []string {
	out := make([]string, 0, len(actions))
	for _, action := range actions {
		plain := ansi.Strip(action)
		key, _, _ := strings.Cut(plain, " ")
		switch {
		case key == "[,/.]", key == "[g]", key == "[+/-/0]", key == "[+/-]", key == "[0]", plain == "[E] calendar":
			continue
		}
		out = append(out, action)
	}
	return out
}

// GlobalActions are the bottom row's keys. Daily's task and habit keys live in its panes and
// its date keys on the date row.
func GlobalActions(theme Theme, state ActionsState) []string {
	actions := []string{}
	if tabLabel := tabActionLabel(state); tabLabel != "" {
		actions = append(
			actions,
			theme.StyleHeader.Render("[tab]")+theme.StyleDim.Render(" "+tabLabel),
		)
	}
	if supportsGlobalCreate(state.View) && !state.SelectedIssueAdhoc {
		actions = append(actions, theme.StyleHeader.Render("[a]")+theme.StyleDim.Render(" create"))
	}
	if supportsGlobalContext(state.View) && state.Pane != string(uistate.PaneScope) {
		actions = append(
			actions,
			theme.StyleHeader.Render("[c]")+theme.StyleDim.Render(" context"),
		)
	}
	if supportsGlobalExport(state.View) {
		actions = append(
			actions,
			theme.StyleHeader.Render("[E]")+theme.StyleDim.Render(" export"),
		)
	}
	if state.View == "daily" || state.View == "review" {
		actions = append(actions, checkInAction(theme))
		actions = append(actions, awayAction(theme, false))
	}
	if state.View == "away" && state.AwayModeActive {
		actions = append(actions, awayAction(theme, true))
	}
	return actions
}

func ContextualActions(theme Theme, state ActionsState) []string {
	if state.Pane == string(uistate.PaneScope) ||
		(state.View == string(uistate.ViewDefault) && state.Pane == string(uistate.PaneIssues)) ||
		(state.View == string(uistate.ViewHabitList) && state.Pane == string(uistate.PaneHabits)) {
		return nil
	}

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
				theme.StyleHeader.Render("[A]") + theme.StyleDim.Render(" start ad hoc timer"),
			}
		}
		if state.HardLimitActive {
			actions := []string{
				theme.StyleHeader.Render("[x]") + theme.StyleDim.Render(" end session"),
				theme.StyleHeader.Render("[i]") + theme.StyleDim.Render(" context"),
				theme.StyleHeader.Render("[s]") + theme.StyleDim.Render(" change status"),
			}
			if sharedtypes.NormalizeTimerHardLimitKind(state.HardLimitKind) == sharedtypes.TimerHardLimitKindPomodoro {
				if label := helperpkg.PomodoroAdvanceActionLabel(&api.TimerState{
					State:           state.TimerState,
					SegmentType:     segmentPtr(state.TimerSegment),
					NextSegmentType: segmentPtr(state.TimerNextSegment),
					HardLimitActive: true,
					HardLimitKind:   state.HardLimitKind,
				}); label != "" {
					actions = append([]string{theme.StyleHeader.Render("[z]") + theme.StyleDim.Render(" "+label)}, actions...)
				}
			}
			return actions
		}
		if state.TimerState == "ready" {
			return []string{
				theme.StyleHeader.Render(
					"[r]",
				) + theme.StyleDim.Render(
					" start "+timerActionSegmentLabel(state.TimerNextSegment),
				),
				theme.StyleHeader.Render("[x]") + theme.StyleDim.Render(" end session"),
				theme.StyleHeader.Render("[i]") + theme.StyleDim.Render(" context"),
				theme.StyleHeader.Render("[s]") + theme.StyleDim.Render(" change status"),
			}
		}
		return []string{
			theme.StyleHeader.Render("[p]") + theme.StyleDim.Render(" pause"),
			theme.StyleHeader.Render("[x]") + theme.StyleDim.Render(" end session"),
			theme.StyleHeader.Render("[i]") + theme.StyleDim.Render(" context"),
			theme.StyleHeader.Render("[s]") + theme.StyleDim.Render(" change status"),
		}
	}
	if state.View == "session_history" {
		return []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" open details dialog"),
		}
	}
	if state.View == "habit_history" {
		return []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" open details dialog"),
		}
	}
	if state.View == "momentum" {
		return []string{
			theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" create momentum"),
			theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit"),
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" open details dialog"),
			theme.StyleHeader.Render("[D]") + theme.StyleDim.Render(" delete"),
			theme.StyleHeader.Render("[x]") + theme.StyleDim.Render(" toggle enabled"),
		}
	}
	if state.View == "review" {
		return []string{
			theme.StyleHeader.Render("[↑/↓]") + theme.StyleDim.Render(" day"),
			theme.StyleHeader.Render("[S]") + theme.StyleDim.Render(" trends/time"),
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" day (narrow)"),
			checkInAction(theme),
			theme.StyleHeader.Render("[,/.]") + theme.StyleDim.Render(" date"),
			theme.StyleHeader.Render("[g]") + theme.StyleDim.Render(" today"),
			theme.StyleHeader.Render("[E]") + theme.StyleDim.Render(" calendar"),
			theme.StyleHeader.Render("[+/-/0]") + theme.StyleDim.Render(" window"),
		}
	}
	if state.View == "config" {
		actions := []string{
			theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit or open"),
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" open details dialog"),
			theme.StyleHeader.Render("[c/space]") + theme.StyleDim.Render(" change option"),
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
			theme.StyleHeader.Render("[o]") + theme.StyleDim.Render(" open report"),
			theme.StyleHeader.Render("[D]") + theme.StyleDim.Render(" delete"),
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" open details dialog"),
		}
	}
	if state.View == "updates" {
		actions := []string{
			theme.StyleHeader.Render("[r]") + theme.StyleDim.Render(" refresh"),
		}
		openLabel := ""
		copyLabel := ""
		if state.UpdateScriptDeprecated {
			openLabel = "open migration guide"
		} else if state.UpdateVisible {
			openLabel = "open release notes"
		}
		if strings.TrimSpace(state.UpdateCommand) != "" {
			copyLabel = "copy command"
			if IsMigrationCommand(state.UpdateCommand) {
				copyLabel = "copy migration command"
			}
		}
		if openLabel != "" {
			actions = append(
				actions,
				theme.StyleHeader.Render("[o]")+theme.StyleDim.Render(" "+openLabel),
			)
		}
		if copyLabel != "" {
			actions = append(
				actions,
				theme.StyleHeader.Render("[c]")+theme.StyleDim.Render(" "+copyLabel),
			)
		}
		actions = append(
			actions,
			theme.StyleHeader.Render("[d]")+
				theme.StyleDim.Render(" "+diagnosticsActionLabel(state.UpdateDiagnosticsExpanded)),
		)
		return actions
	}
	if state.View == "support" {
		return []string{
			theme.StyleHeader.Render("[o]") + theme.StyleDim.Render(" feedback & roadmap"),
			theme.StyleHeader.Render("[r]") + theme.StyleDim.Render(" open releases"),
			theme.StyleHeader.Render("[g]") + theme.StyleDim.Render(" open documentation"),
			theme.StyleHeader.Render("[c]") + theme.StyleDim.Render(" copy diagnostics"),
			theme.StyleHeader.Render("[b]") + theme.StyleDim.Render(" generate bundle"),
		}
	}
	if state.View == "alerts" {
		return []string{
			theme.StyleHeader.Render("[←/→]") + theme.StyleDim.Render(" change"),
			theme.StyleHeader.Render("[space]") + theme.StyleDim.Render(" toggle"),
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" edit or run"),
			theme.StyleHeader.Render("[D]") + theme.StyleDim.Render(" delete"),
		}
	}

	switch state.Pane {
	case "repos", "streams":
		contextLabel := " change context"
		return []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" open details dialog"),
			theme.StyleHeader.Render("[c]") + theme.StyleDim.Render(contextLabel),
			theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" create"),
			theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit"),
			theme.StyleHeader.Render("[D]") + theme.StyleDim.Render(" delete"),
		}
	case "issues":
		timerIdle := state.TimerState == "" || state.TimerState == "idle"
		if state.SelectedIssueAdhoc {
			return []string{
				theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" review ad hoc session"),
				theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" convert to regular issue"),
			}
		}
		if state.View == "daily" {
			actions := []string{
				theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" open issue details"),
				theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" create issue"),
				theme.StyleHeader.Render("[s]") + theme.StyleDim.Render(" set status"),
				theme.StyleHeader.Render("[d]") + theme.StyleDim.Render(" set due date"),
				theme.StyleHeader.Render("[P]") + theme.StyleDim.Render(" pin to daily"),
			}
			if timerIdle {
				actions = append(
					actions,
					theme.StyleHeader.Render("[f]")+theme.StyleDim.Render(" start focus timer"),
					theme.StyleHeader.Render("[A]")+theme.StyleDim.Render(" start ad hoc timer"),
					theme.StyleHeader.Render("[m]")+theme.StyleDim.Render(" log work"),
					theme.StyleHeader.Render("[e]")+theme.StyleDim.Render(" edit issue"),
				)
			}
			return actions
		}
		actions := []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" open issue details"),
			theme.StyleHeader.Render("[s]") + theme.StyleDim.Render(" set status"),
			theme.StyleHeader.Render("[d]") + theme.StyleDim.Render(" set due date"),
			theme.StyleHeader.Render("[P]") + theme.StyleDim.Render(" pin"),
		}
		if timerIdle {
			actions = append(
				actions,
				theme.StyleHeader.Render("[f]")+theme.StyleDim.Render(" start focus timer"),
				theme.StyleHeader.Render("[A]")+theme.StyleDim.Render(" start ad hoc timer"),
				theme.StyleHeader.Render("[m]")+theme.StyleDim.Render(" log work"),
				theme.StyleHeader.Render("[e/D]")+theme.StyleDim.Render(" edit/delete"),
			)
		}
		return actions
	case "habits":
		if state.View == "daily" {
			actions := []string{
				theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" open habit details"),
				theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" create habit"),
				theme.StyleHeader.Render("[x]") + theme.StyleDim.Render(" toggle completion"),
				theme.StyleHeader.Render("[F]") + theme.StyleDim.Render(" mark failed"),
				theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit habit"),
				theme.StyleHeader.Render("[D]") + theme.StyleDim.Render(" delete"),
			}
			if state.TimerState == "" || state.TimerState == "idle" {
				actions = append(
					actions,
					theme.StyleHeader.Render("[m]")+theme.StyleDim.Render(" log completion"),
				)
			}
			return actions
		}
		actions := []string{
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" open habit details"),
			theme.StyleHeader.Render("[a]") + theme.StyleDim.Render(" create habit"),
			theme.StyleHeader.Render("[e]") + theme.StyleDim.Render(" edit habit"),
		}
		if state.TimerState == "" || state.TimerState == "idle" {
			actions = append(actions, theme.StyleHeader.Render("[D]")+theme.StyleDim.Render(" delete"), theme.StyleHeader.Render("[m]")+theme.StyleDim.Render(" log completion"))
		}
		return actions
	case "ops":
		return []string{
			theme.StyleHeader.Render("[+]") + theme.StyleDim.Render(" more"),
			theme.StyleHeader.Render("[-]") + theme.StyleDim.Render(" less"),
		}
	case "settings":
		return []string{
			theme.StyleHeader.Render("[←/→]") + theme.StyleDim.Render(" change"),
			theme.StyleHeader.Render(
				"[enter]",
			) + theme.StyleDim.Render(
				" edit, toggle, or confirm",
			),
		}
	case "alerts":
		return []string{
			theme.StyleHeader.Render("[←/→]") + theme.StyleDim.Render(" change"),
			theme.StyleHeader.Render("[enter]") + theme.StyleDim.Render(" toggle or test"),
		}
	}
	return nil
}

func IsMigrationCommand(command string) bool {
	return strings.HasPrefix(strings.TrimSpace(command), "brew uninstall ")
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
	return view == "momentum"
}

func supportsGlobalContext(view string) bool {
	switch view {
	case "default",
		"daily",
		"habit_list",
		"review",
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
	case "default", "daily", "habit_list", "reports":
		return true
	default:
		return false
	}
}

func PaneActions(theme Theme, state ActionsState) []string {
	return ContextualActions(theme, state)
}

func diagnosticsActionLabel(expanded bool) string {
	if expanded {
		return "hide diagnostics"
	}
	return "show diagnostics"
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

func compactActionSegments(theme Theme, actions []string) []string {
	out := make([]string, 0, len(actions))
	for _, action := range actions {
		key, label, ok := actionSegmentParts(action)
		if !ok {
			out = append(out, action)
			continue
		}
		if compact := compactActionLabel(key, label); compact != "" && compact != label {
			out = append(
				out,
				theme.StyleHeader.Render(key)+theme.StyleDim.Render(" "+compact),
			)
			continue
		}
		out = append(out, action)
	}
	return out
}

func actionSegmentParts(action string) (key, label string, ok bool) {
	plain := strings.TrimSpace(ansi.Strip(action))
	if plain == "" || !strings.HasPrefix(plain, "[") {
		return "", "", false
	}
	end := strings.IndexByte(plain, ']')
	if end < 0 {
		return "", "", false
	}
	key = plain[:end+1]
	label = strings.TrimSpace(plain[end+1:])
	return key, label, true
}

func compactActionLabel(key, label string) string {
	switch label {
	case "open issue details", "open habit details", "open details dialog", "open day details":
		return "details"
	case "open report":
		return "report"
	case "open release notes":
		return "release"
	case "open migration guide":
		return "guide"
	case "open documentation":
		return "documentation"
	case "feedback & roadmap":
		return "feedback"
	case "change context":
		return "context"
	case "create issue", "create habit", "create momentum", "create":
		return "create"
	case "set status":
		return "status"
	case "set due date":
		return "due"
	case "pin to daily", "pin":
		return "pin"
	case "start focus timer":
		return "focus"
	case "log work", "log completion":
		return "log"
	case "edit issue", "edit habit", "edit":
		return "edit"
	case "toggle completion", "toggle enabled":
		return "toggle"
	case "mark failed":
		return "fail"
	case "delete":
		return "delete"
	case "more":
		return "more"
	case "less":
		return "less"
	case "change option":
		return "option"
	case "rescan tools":
		return "rescan"
	case "edit or open":
		return "edit/open"
	case "edit, toggle, or confirm":
		return "edit/toggle"
	case "edit or run":
		return "edit/run"
	case "copy diagnostics":
		return "copy"
	case "generate bundle":
		return "bundle"
	case "switch views":
		return "views"
	case "keys":
		return "help"
	case "export data":
		return "export"
	case "check-in":
		return "check-in"
	case "away":
		return "away"
	case "disable away":
		return "off"
	case "change":
		return "change"
	case "today":
		return "today"
	case "date":
		return "date"
	case "window":
		return "window"
	case "beta support":
		return "beta"
	}

	switch key {
	case "[d/x]", "[D]":
		if label == "delete" {
			return "delete"
		}
	case "[e/d]", "[e/D]":
		if label == "edit/delete" {
			return "edit/del"
		}
	case "[c/space]":
		if label == "change option" {
			return "option"
		}
	case "[s]":
		if label == "change status" {
			return "status"
		}
	case "[+]":
		if label == "more" {
			return "more"
		}
	case "[-]":
		if label == "less" {
			return "less"
		}
	}

	return label
}

// minStripBeforeHint is the strip room a window hint leaves before it gives way.
const minStripBeforeHint = 40

// WithWindowHint pins the window-size keys to the right of a strip's first row when the
// strip keeps at least minStripBeforeHint cells; render draws the strip at the width given.
func WithWindowHint(theme Theme, days, width int, render func(width int) []string) []string {
	hint := KeyHint(theme, "+/-/0", fmt.Sprintf("%dd window", max(1, days)))
	room := width - ansi.StringWidth(hint) - 3
	if room < minStripBeforeHint {
		return render(width)
	}
	rows := render(room)
	rows[0] += "   " + hint
	for i := 1; i < len(rows); i++ {
		rows[i] += strings.Repeat(" ", width-room)
	}
	return rows
}
