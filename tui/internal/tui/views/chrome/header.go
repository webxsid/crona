package viewchrome

import (
	"fmt"
	"strings"
	"time"

	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"

	"github.com/charmbracelet/lipgloss"
)

type HeaderState struct {
	Width         int
	View          string
	Elapsed       int
	Timer         *api.TimerState
	IssueSessions []api.Session
	AllIssues     []api.IssueWithMeta
	Health        *api.Health
	UpdateStatus  *api.UpdateStatus
}

// HeaderSessionLine is the status bar's session chip: phase and cycle, the clock, the issue
// and its worked time against the estimate. The session view already shows all of it, so
// there the chip is empty. The engine shows only when it is not ok.
func HeaderSessionLine(theme Theme, state HeaderState) string {
	if state.Timer == nil || state.Timer.State == "idle" || state.View == "session_active" {
		return ""
	}
	timer := state.Timer
	now := time.Now()
	clock := helperpkg.DerivedSegmentElapsedSeconds(timer, state.Elapsed, now)
	if timer.HardLimitActive && !timer.HardLimitExpired {
		if remaining, _, ok := helperpkg.DerivedHardLimitSegmentRemainingSeconds(timer, state.Elapsed, now); ok {
			clock = remaining
		}
	}
	phase, color := sessionPhase(theme, timer)
	parts := []string{
		LipStyle(theme, color).Render(phase),
		theme.StyleHeader.Render(viewhelpers.FormatClockText(clock)),
	}
	if issue := activeIssueWithMeta(state); issue != nil {
		parts = append(parts, viewhelpers.Truncate(issue.Title, 40))
		worked := helperpkg.DerivedSegmentElapsedSeconds(timer, state.Elapsed, now)
		if timer.SegmentType != nil && *timer.SegmentType != "work" {
			worked = 0
		}
		label := helperpkg.FormatCompactDurationMinutes((timer.IssueWorkedSeconds + worked) / 60)
		if issue.EstimateMinutes != nil && *issue.EstimateMinutes > 0 {
			label += " / est. " + helperpkg.FormatCompactDurationMinutes(*issue.EstimateMinutes)
		}
		parts = append(parts, theme.StyleDim.Render(label))
	} else if timer.Adhoc {
		parts = append(parts, theme.StyleDim.Render("ad hoc"))
	}
	if chip := healthChip(state.Health); chip != "" {
		parts = append(parts, LipStyle(theme, theme.ColorYellow).Render(chip))
	}
	return strings.Join(parts, theme.StyleDim.Render("  ·  "))
}

// sessionPhase names the timer's phase for the chip, with the pomodoro cycle when known.
func sessionPhase(theme Theme, timer *api.TimerState) (string, lipgloss.Color) {
	switch {
	case timer.HardLimitExpired:
		return "DONE", theme.ColorYellow
	case timer.State == "ready":
		return "READY", theme.ColorYellow
	case timer.State == "paused":
		return "PAUSED", theme.ColorYellow
	}
	segment := "work"
	if timer.SegmentType != nil {
		segment = string(*timer.SegmentType)
	}
	switch segment {
	case "short_break":
		return "BREAK", theme.ColorCyan
	case "long_break":
		return "LONG BREAK", theme.ColorMagenta
	case "rest":
		return "REST", theme.ColorYellow
	}
	if timer.HardLimitCycles > 0 {
		return fmt.Sprintf("FOCUS %d/%d", min(timer.CompletedWorkCycles+1, timer.HardLimitCycles), timer.HardLimitCycles), theme.ColorGreen
	}
	return "FOCUS", theme.ColorGreen
}

func healthChip(health *api.Health) string {
	if health == nil || (health.OK == 1 && health.DB) {
		return ""
	}
	return "engine degraded"
}

func activeIssueWithMeta(state HeaderState) *api.IssueWithMeta {
	if state.Timer == nil || state.Timer.IssueID == nil {
		return nil
	}
	for i := range state.AllIssues {
		if state.AllIssues[i].ID == *state.Timer.IssueID {
			return &state.AllIssues[i]
		}
	}
	return nil
}
