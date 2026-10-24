package sessions

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	sessionmeta "crona/tui/internal/tui/views/sessionmeta"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderActiveView(theme types.Theme, state types.ContentState) string {
	var activeIssue *api.IssueWithMeta
	if state.Timer != nil && state.Timer.IssueID != nil {
		activeIssue = sessionmeta.IssueMetaByID(state.AllIssues, *state.Timer.IssueID)
	}
	if state.Timer != nil && state.Timer.Adhoc {
		activeIssue = nil
	}
	hardLimit := state.Timer != nil && state.Timer.HardLimitActive
	now := time.Now()
	workElapsedSeconds := helperpkg.DerivedSegmentElapsedSeconds(state.Timer, state.Elapsed, now)
	displaySeconds := workElapsedSeconds
	if hardLimit && !state.Timer.HardLimitExpired {
		if remainingSeconds, _, ok := helperpkg.DerivedHardLimitSegmentRemainingSeconds(
			state.Timer,
			state.Elapsed,
			now,
		); ok {
			displaySeconds = remainingSeconds
		}
	} else if state.Timer.State == "paused" {
		displaySeconds = workElapsedSeconds
	}
	elapsed := viewhelpers.FormatClockText(displaySeconds)
	seg := "work"
	if state.Timer.SegmentType != nil {
		seg = string(*state.Timer.SegmentType)
	} else if state.Timer.ReadySegmentType != nil && !hardLimit {
		seg = string(*state.Timer.ReadySegmentType)
	}
	timerTitle := "Focus Session"
	timerHint := "[p] pause  [x] end session  [i] context"
	nextLabel := sessionActionSegmentLabel(state.Timer)
	stateColor := activeTimerColor(theme, state.Timer)
	if hardLimit {
		timerTitle = "Pomodoro Session"
		if sharedtypes.NormalizeTimerHardLimitKind(state.Timer.HardLimitKind) ==
			sharedtypes.TimerHardLimitKindCountdown {
			timerTitle = "Timer Session"
		}
		timerHint = "[x] end session  [i] context"
		if state.Timer.State == "ready" {
			timerHint = "[r] start " + nextLabel + "  " + timerHint
			if state.Timer.ReadySegmentType != nil {
				stateColor = segmentColor(theme, *state.Timer.ReadySegmentType)
			} else if state.Timer.NextSegmentType != nil {
				stateColor = segmentColor(theme, *state.Timer.NextSegmentType)
			}
		}
		if label := helperpkg.PomodoroAdvanceActionLabel(state.Timer); label != "" {
			timerHint = "[z] " + label + "  " + timerHint
		}
	} else if state.Timer.State == "ready" && !hardLimit {
		timerTitle = "Ready For"
		timerHint = "[r] start " + nextLabel + "  [x] end session  [i] context"
	} else if state.Timer.State == "paused" {
		stateColor = theme.ColorYellow
		timerTitle = "Paused For"
		timerHint = "[r] resume  [x] end session  [i] context"
		seg = "paused"
	}
	if hardLimit && state.Timer.State != "ready" {
		timerTitle = strings.TrimSuffix(timerTitle, " Session") + " · " + phaseLabel(seg)
		if seg == "work" && state.Timer.HardLimitCycles > 0 {
			timerTitle += fmt.Sprintf(
				" %d of %d",
				min(state.Timer.CompletedWorkCycles+1, state.Timer.HardLimitCycles),
				state.Timer.HardLimitCycles,
			)
		}
	} else if !hardLimit {
		phase := phaseLabel(seg)
		if state.Timer.State == "ready" {
			phase = "Ready for " + nextLabel
		}
		timerTitle = "Stopwatch · " + phase
	}
	totalH := max(state.Height, 1)
	width := max(1, state.Width-2)
	centerWidth := max(1, width-1)

	// Under the clock: the phase bar (timed sessions only), then one status line.
	below := []string{}
	if hardLimit {
		if bar := hardLimitProgressBar(theme, state, min(centerWidth, 72), now, stateColor); bar != "" {
			below = append(below, "", bar)
		}
	}
	status := []string{}
	if label := sessionTimingLabel(state, now); label != "" {
		status = append(status, label)
	}
	if nextLabel != "next" && state.Timer.State != "ready" {
		status = append(status, "next "+nextLabel)
	}
	if label := sessionEndsAtLabel(state, now); label != "" {
		status = append(status, strings.Replace(label, "Ends At", "ends at", 1))
	}
	if dots := cycleDots(state.Timer, seg); dots != "" {
		below = append(below, "", lipgloss.NewStyle().Foreground(stateColor).Render(dots))
	}
	if len(status) > 0 {
		below = append(
			below,
			"",
			lipgloss.NewStyle().
				Foreground(stateColor).
				Bold(true).
				Render(strings.Join(status, "  ·  ")),
		)
	}
	priorWorkedSeconds := state.Timer.IssueWorkedSeconds
	below = append(below, "", "")
	below = append(
		below,
		sessionIssueLines(
			theme,
			activeIssue,
			state.Timer.Adhoc,
			priorWorkedSeconds+workElapsedSeconds,
			centerWidth,
		)...)
	below = append(below, "", theme.StyleDim.Render(timerHint))

	clockHeight := max(3, min(totalH/2, totalH-1-len(below)-2))
	timerText := sessionmeta.RenderResponsiveClock(
		elapsed,
		centerWidth,
		clockHeight,
		stateColor,
		theme.ColorDim,
	)
	body := splitLines(
		[]string{
			lipgloss.NewStyle().Foreground(stateColor).Bold(true).Render(timerText),
			centerLines(below, centerWidth),
		},
	)
	for i, line := range body {
		body[i] = lipgloss.NewStyle().
			Width(centerWidth).
			AlignHorizontal(lipgloss.Center).
			Render(line)
	}
	if pad := (totalH - 1 - len(body)) / 2; pad > 0 {
		body = append(make([]string, pad), body...)
	}
	titleStyle := lipgloss.NewStyle().Foreground(stateColor).Bold(true)
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title:      timerTitle,
		Meta:       sessionContextLabel(activeIssue, state.Timer.Adhoc),
		TitleStyle: &titleStyle,
		Width:      state.Width,
		Height:     totalH,
		Lines:      body,
	})
}

// cycleDots is the pomodoro's cycles: ● finished, ◉ in focus now, ○ to come.
func cycleDots(timer *api.TimerState, segment string) string {
	if timer == nil || timer.HardLimitCycles <= 0 {
		return ""
	}
	dots := make([]string, timer.HardLimitCycles)
	for i := range dots {
		switch {
		case i < timer.CompletedWorkCycles:
			dots[i] = "●"
		case i == timer.CompletedWorkCycles && segment == "work":
			dots[i] = "◉"
		default:
			dots[i] = "○"
		}
	}
	return strings.Join(dots, " ")
}

// phaseLabel names a segment for the header: Focus, Short break, Long break, Rest, Paused.
func phaseLabel(segment string) string {
	switch segment {
	case "work":
		return "Focus"
	case "short_break":
		return "Short break"
	case "long_break":
		return "Long break"
	case "rest":
		return "Rest"
	case "paused":
		return "Paused"
	default:
		return "Focus"
	}
}

func sessionContextLabel(issue *api.IssueWithMeta, adhoc bool) string {
	if adhoc {
		return "Ad hoc"
	}
	if issue == nil || issue.RepoName == "" {
		return ""
	}
	return issue.RepoName + " › " + issue.StreamName
}

// sessionIssueLines is the issue block under the clock: title, one line of description,
// then worked time against the estimate.
func sessionIssueLines(
	theme types.Theme,
	issue *api.IssueWithMeta,
	adhoc bool,
	workedSeconds, width int,
) []string {
	if issue == nil {
		if adhoc {
			return []string{
				theme.StyleDim.Render("Ad hoc session · attribution pending until you end it"),
			}
		}
		return []string{theme.StyleDim.Render("No issue selected")}
	}
	lines := []string{theme.StyleHeader.Render(viewhelpers.Truncate(issue.Title, width))}
	if issue.Description != nil && strings.TrimSpace(*issue.Description) != "" {
		lines = append(
			lines,
			theme.StyleDim.Render(viewhelpers.Truncate(collapseSpace(*issue.Description), width)),
		)
	}
	worked := "Worked " + helperpkg.FormatCompactDurationMinutes(workedSeconds/60)
	if issue.EstimateMinutes != nil && *issue.EstimateMinutes > 0 {
		worked += fmt.Sprintf(
			" / est. %s  ·  %d%%",
			helperpkg.FormatCompactDurationMinutes(*issue.EstimateMinutes),
			workedSeconds*100/(*issue.EstimateMinutes*60),
		)
	}
	return append(lines, "", theme.StyleDim.Render(worked))
}

func splitLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, strings.Split(line, "\n")...)
	}
	return out
}

func sessionTimingLabel(state types.ContentState, now time.Time) string {
	if state.Timer == nil {
		return ""
	}
	if state.Timer.HardLimitActive {
		if state.Timer.HardLimitExpired {
			return "commit or extend"
		}
		remaining, _, ok := helperpkg.DerivedHardLimitSegmentRemainingSeconds(
			state.Timer,
			state.Elapsed,
			now,
		)

		if !ok {
			return ""
		}
		remainingLabel := formatRemainingMinutes(remaining)
		if remainingLabel == "" {
			return ""
		}
		return remainingLabel + " left"
	}
	if state.Settings != nil && state.Settings.TimerMode == sharedtypes.TimerModeStopwatch {
		return ""
	}
	segment := state.Timer.SegmentType
	if state.Timer.State == "ready" {
		segment = state.Timer.ReadySegmentType
		if segment == nil {
			segment = state.Timer.NextSegmentType
		}
		if segment == nil {
			return ""
		}
		return segmentReadyLabel(*segment) + " ready"
	}
	if segment == nil || state.Settings == nil {
		return ""
	}
	durationSeconds, ok := segmentDurationSeconds(state.Settings, *segment)
	if !ok || durationSeconds <= 0 {
		return ""
	}
	elapsed := helperpkg.DerivedSegmentElapsedSeconds(state.Timer, state.Elapsed, now)
	elapsed = min(elapsed, durationSeconds)
	remaining := durationSeconds - elapsed
	remainingLabel := formatRemainingMinutes(remaining)
	if remainingLabel == "" {
		return ""
	}
	return remainingLabel + " left"
}

func sessionEndsAtLabel(state types.ContentState, now time.Time) string {
	if state.Timer == nil ||
		state.Timer.State == "ready" ||
		state.Timer.State == "paused" ||
		state.Timer.State == "expired" {
		return ""
	}
	remaining := 0
	if state.Timer.HardLimitActive {
		remaining = helperpkg.DerivedHardLimitRemainingSeconds(
			state.Timer,
			state.Elapsed,
			now,
		)
	} else {
		if state.Settings == nil || state.Settings.TimerMode == sharedtypes.TimerModeStopwatch {
			return ""
		}
		segment := state.Timer.SegmentType
		if segment == nil {
			return ""
		}
		durationSeconds, ok := segmentDurationSeconds(state.Settings, *segment)
		if !ok || durationSeconds <= 0 {
			return ""
		}
		elapsed := helperpkg.DerivedSegmentElapsedSeconds(state.Timer, state.Elapsed, now)
		remaining = max(0, durationSeconds-min(elapsed, durationSeconds))
	}
	if remaining <= 0 {
		return ""
	}
	return "Ends At " + now.Add(time.Duration(remaining)*time.Second).Format("15:04")
}

func segmentReadyLabel(segment sharedtypes.SessionSegmentType) string {
	switch segment {
	case sharedtypes.SessionSegmentWork:
		return "work"
	case sharedtypes.SessionSegmentShortBreak, sharedtypes.SessionSegmentLongBreak:
		return "break"
	case sharedtypes.SessionSegmentRest:
		return "rest"
	default:
		return "next"
	}
}

func segmentDurationSeconds(
	settings *api.CoreSettings,
	segment sharedtypes.SessionSegmentType,
) (int, bool) {
	if settings == nil {
		return 0, false
	}
	switch segment {
	case sharedtypes.SessionSegmentWork:
		return settings.WorkDurationMinutes * 60, settings.WorkDurationMinutes > 0
	case sharedtypes.SessionSegmentShortBreak:
		return settings.ShortBreakMinutes * 60, settings.ShortBreakMinutes > 0
	case sharedtypes.SessionSegmentLongBreak:
		return settings.LongBreakMinutes * 60, settings.LongBreakMinutes > 0
	default:
		return 0, false
	}
}

func formatRemainingMinutes(seconds int) string {
	seconds = max(0, seconds)
	minutes := seconds / 60
	if seconds%60 != 0 {
		minutes++
	}
	minutes = max(1, minutes)
	if minutes == 1 {
		return "1 min"
	}
	return fmt.Sprintf("%d mins", minutes)
}

func hardLimitProgressBar(
	theme types.Theme,
	state types.ContentState,
	width int,
	now time.Time,
	color lipgloss.Color,
) string {
	if state.Timer == nil || !state.Timer.HardLimitActive || width < 8 {
		return ""
	}
	remaining, segment, ok := helperpkg.DerivedHardLimitSegmentRemainingSeconds(
		state.Timer,
		state.Elapsed,
		now,
	)
	if !ok {
		return ""
	}
	durationSeconds, ok := hardLimitSegmentDurationSeconds(state.Timer, *segment)
	if !ok || durationSeconds <= 0 {
		return ""
	}
	barWidth := max(10, width)
	done := durationSeconds - remaining
	done = max(0, done)
	filled := min(barWidth, (done*barWidth)/durationSeconds)
	palette := sessionProgressPalette(theme, state.Timer, color)
	return lipgloss.NewStyle().
		Width(width).
		AlignHorizontal(lipgloss.Center).
		Render(viewhelpers.RenderGradientBar(barWidth, filled, palette))
}

func sessionActionSegmentLabel(timer *api.TimerState) string {
	if timer == nil {
		return "next"
	}
	if timer.ReadySegmentType != nil {
		return segmentActionLabel(string(*timer.ReadySegmentType))
	}
	if timer.NextSegmentType != nil {
		return segmentActionLabel(string(*timer.NextSegmentType))
	}
	return "next"
}

func segmentActionLabel(segment string) string {
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

func activeTimerColor(theme types.Theme, timer *api.TimerState) lipgloss.Color {
	if timer == nil {
		return theme.ColorGreen
	}
	if timer.State == "ready" {
		return theme.ColorYellow
	}
	segment := ""
	if timer.SegmentType != nil {
		segment = string(*timer.SegmentType)
	}
	switch segment {
	case "short_break":
		return theme.ColorCyan
	case "long_break":
		return theme.ColorMagenta
	case "rest":
		return theme.ColorYellow
	default:
		if timer.State == "paused" {
			return theme.ColorYellow
		}
		return theme.ColorGreen
	}
}

func segmentColor(theme types.Theme, segment sharedtypes.SessionSegmentType) lipgloss.Color {
	switch segment {
	case sharedtypes.SessionSegmentShortBreak:
		return theme.ColorCyan
	case sharedtypes.SessionSegmentLongBreak:
		return theme.ColorMagenta
	case sharedtypes.SessionSegmentRest:
		return theme.ColorYellow
	default:
		return theme.ColorGreen
	}
}

func sessionProgressPalette(
	theme types.Theme,
	timer *api.TimerState,
	color lipgloss.Color,
) viewhelpers.GradientBarPalette {
	palette := viewhelpers.GradientBarPalette{
		Start: theme.ColorDullGreen,
		End:   color,
		Track: theme.ColorDim,
	}
	if timer == nil || timer.SegmentType == nil {
		return palette
	}
	switch *timer.SegmentType {
	case sharedtypes.SessionSegmentShortBreak:
		palette.Start = theme.ColorBlue
		palette.End = theme.ColorCyan
	case sharedtypes.SessionSegmentLongBreak:
		palette.Start = theme.ColorMagenta
		palette.End = theme.ColorCyan
	case sharedtypes.SessionSegmentRest:
		palette.Start = theme.ColorYellow
		palette.End = theme.ColorOrange
	default:
		palette.End = color
	}
	return palette
}

func hardLimitSegmentDurationSeconds(
	timer *api.TimerState,
	segment sharedtypes.SessionSegmentType,
) (int, bool) {
	if timer == nil || !timer.HardLimitActive {
		return 0, false
	}
	switch segment {
	case sharedtypes.SessionSegmentWork:
		if timer.HardLimitWorkSeconds <= 0 {
			return 0, false
		}
		return timer.HardLimitWorkSeconds, true
	case sharedtypes.SessionSegmentShortBreak:
		if timer.HardLimitBreakSeconds <= 0 {
			return 0, false
		}
		return timer.HardLimitBreakSeconds, true
	case sharedtypes.SessionSegmentLongBreak:
		if timer.HardLimitLongBreakSeconds > 0 {
			return timer.HardLimitLongBreakSeconds, true
		}
		if timer.HardLimitBreakSeconds > 0 {
			return timer.HardLimitBreakSeconds, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func centerLines(lines []string, width int) string {
	centered := make([]string, 0, len(lines))
	for _, line := range lines {
		centered = append(
			centered,
			lipgloss.NewStyle().Width(max(1, width)).AlignHorizontal(lipgloss.Center).Render(line),
		)
	}
	return strings.Join(centered, "\n")
}
