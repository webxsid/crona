package viewchrome

import (
	"fmt"
	"strings"
	"time"

	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	issuecore "crona/tui/internal/tui/views/issuecore"
	sessionmeta "crona/tui/internal/tui/views/sessionmeta"
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

func HeaderSessionLine(theme Theme, state HeaderState) string {
	if state.Timer == nil || state.Timer.State == "idle" {
		return ""
	}
	return viewhelpers.Truncate(
		headerSessionSummary(theme, state)+"  ·  "+headerSecondary(theme, state),
		max(20, state.Width-4),
	)
}

func headerSessionSummary(theme Theme, state HeaderState) string {
	if state.Timer == nil || state.Timer.State == "idle" {
		return ""
	}

	now := time.Now()
	total := helperpkg.DerivedSegmentElapsedSeconds(state.Timer, state.Elapsed, now)
	if state.Timer.HardLimitActive && !state.Timer.HardLimitExpired {
		if remaining, _, ok := helperpkg.DerivedHardLimitSegmentRemainingSeconds(
			state.Timer,
			state.Elapsed,
			now,
		); ok {
			total = remaining
		}
	} else if state.Timer.State == "paused" {
		total = helperpkg.DerivedSegmentElapsedSeconds(state.Timer, state.Elapsed, now)
	}
	stateText := "WORK"
	stateColor := theme.ColorGreen
	if state.Timer.HardLimitExpired {
		stateText = "LIMIT"
		stateColor = theme.ColorYellow
	} else if state.Timer.HardLimitActive {
		stateText = "LIMIT"
		stateColor = theme.ColorYellow
	} else if state.Timer.State == "ready" {
		stateText = "READY"
		stateColor = theme.ColorYellow
	} else if state.Timer.State == "paused" {
		stateText = "PAUSED"
		stateColor = theme.ColorYellow
	}
	if state.Timer.State == "ready" && !state.Timer.HardLimitActive &&
		state.Timer.ReadySegmentType != nil &&
		*state.Timer.ReadySegmentType != "" {
		stateText = "READY:" + strings.ToUpper(string(*state.Timer.ReadySegmentType))
	} else if state.Timer.SegmentType != nil && *state.Timer.SegmentType != "" && *state.Timer.SegmentType != "work" {
		stateText = strings.ToUpper(string(*state.Timer.SegmentType))
		stateColor = theme.ColorYellow
	}

	parts := []string{
		LipStyle(theme, stateColor).Render(stateText),
		theme.StyleHeader.Render(viewhelpers.FormatClockText(total)),
	}

	priorWorkedSeconds, completedSessions := sessionmeta.SummarizeCompletedSessions(
		state.IssueSessions,
	)
	parts = append(parts, theme.StyleDim.Render(fmt.Sprintf("sessions:%d", completedSessions)))

	if issue := activeIssueWithMeta(state); issue != nil && issue.EstimateMinutes != nil {
		workElapsed := helperpkg.DerivedSegmentElapsedSeconds(state.Timer, state.Elapsed, now)
		parts = append(
			parts,
			theme.StyleDim.Render(
				sessionmeta.FormatEstimateProgress(
					priorWorkedSeconds+workElapsed,
					*issue.EstimateMinutes,
				),
			),
		)
	}

	return strings.Join(parts, theme.StyleDim.Render("  ·  "))
}

func headerSecondary(theme Theme, state HeaderState) string {
	parts := []string{}
	if state.Timer != nil && state.Timer.State != "idle" {
		parts = append(parts, healthChip(state.Health))
		if issue := activeIssueWithMeta(state); issue != nil {
			parts = append(
				parts,
				"status:"+issuecore.IssueStatusStyle(theme, string(issue.Status)).
					Render(strings.ToUpper(issuecore.PlainIssueStatus(string(issue.Status)))),
			)
		}
	} else if state.View == "daily" || state.View == "wellbeing" {
		parts = append(parts, healthChip(state.Health))
	}
	return strings.Join(compactNonEmpty(parts), "  ·  ")
}

func healthChip(health *api.Health) string {
	if health == nil {
		return "engine: checking"
	}
	if health.OK == 1 && health.DB {
		return "engine: ok"
	}
	return "engine: degraded"
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

func compactNonEmpty(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			out = append(out, part)
		}
	}
	return out
}
