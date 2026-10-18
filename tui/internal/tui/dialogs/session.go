package dialogs

import (
	"fmt"
	"strings"
	"time"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	viewchrome "crona/tui/internal/tui/views/chrome"
	issuecore "crona/tui/internal/tui/views/issuecore"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func renderSessionDialog(theme Theme, state controllerpkg.State) string {
	switch state.Kind {
	case "end_session":
		return renderEndSession(theme, state)
	case "adhoc_resolution":
		return renderAdhocResolution(theme, state)
	case "adhoc_issue_picker":
		return renderAdhocIssuePicker(theme, state)
	case "adhoc_new_issue":
		return renderAdhocNewIssue(theme, state)
	case "timer_start_type":
		return renderTimerStartType(theme, state)
	case "pomodoro_start", "hard_limit_extend":
		return renderPomodoro(theme, state)
	case "timer_countdown_start", "timer_countdown_extend":
		return renderCountdownForm(theme, state)
	case "hard_limit_expired":
		return renderHardLimitExpired(theme, state)
	case "amend_session":
		return renderAmendSession(theme, state)
	case "edit_session_history":
		return renderEditSessionHistory(theme, state)
	case "manual_session":
		return renderManualSession(theme, state)
	case "issue_session_transition":
		return renderIssueSessionTransition(theme, state)
	default:
		return ""
	}
}

func timerStartChoiceLabel(idx int, item string) string {
	switch idx {
	case 0:
		return "[s] " + item
	case 1:
		return "[p] " + item
	case 2:
		return "[t] " + item
	default:
		return item
	}
}

func historySegmentTypeRow(
	theme Theme,
	state controllerpkg.State,
	segment controllerpkg.SessionHistorySegmentRow,
	index int,
) string {
	typeName := cases.Title(language.Und).String(strings.ReplaceAll(string(segment.Type), "_", " "))
	if typeName == "" {
		typeName = "Work"
	}
	focus := (state.FocusIdx-5)/3 == index && state.FocusIdx >= 5
	line := fmt.Sprintf("%02d  ‹ %s ›", index+1, typeName)
	if focus {
		return theme.StyleCursor.Render(viewchrome.SelectionCursor + " " + line)
	}
	return theme.StyleNormal.Render("  " + line)
}

func timerIssueSummary(state controllerpkg.State) string {
	if state.IssueEstimateMins == nil && state.IssueWorkedSeconds <= 0 {
		return ""
	}
	return issuecore.IssueWorkedEstimateSummary(state.IssueWorkedSeconds, state.IssueEstimateMins)
}

func renderTimerMeta(
	theme Theme,
	state controllerpkg.State,
	totalLabel string,
	totalSeconds int,
	includeEndsAt bool,
) string {
	rows := make([]string, 0, 3)
	if summary := timerIssueSummary(state); summary != "" {
		rows = append(rows, timerMetaRow(theme, "Worked / Est", summary))
	}
	if strings.TrimSpace(totalLabel) != "" {
		rows = append(
			rows,
			timerMetaRow(theme, "Total", strings.TrimPrefix(totalLabel, "Total Duration ")),
		)
	}
	if includeEndsAt {
		if endsAt := timerEndsAtLabel(time.Now(), totalSeconds); endsAt != "" {
			rows = append(rows, timerMetaRow(theme, "Ends At", endsAt))
		}
	}
	if len(rows) == 0 {
		return ""
	}
	return strings.Join(rows, "\n")
}

func timerMetaRow(theme Theme, label string, value string) string {
	labelStyle := theme.StyleDim.Width(len("Worked / Est") + 2)
	valueStyle := theme.StyleNormal
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		labelStyle.Render(label),
		valueStyle.Render(strings.TrimSpace(value)),
	)
}

func timerEndsAtLabel(base time.Time, totalSeconds int) string {
	if totalSeconds <= 0 {
		return ""
	}
	return base.Add(time.Duration(totalSeconds) * time.Second).Format("15:04")
}

func pomodoroChoiceRow(
	theme Theme,
	choices []string,
	selected int,
	customIdx int,
	customActive bool,
	customView string,
) string {
	parts := make([]string, 0, len(choices)+1)
	for i, choice := range choices {
		style := theme.StyleNormal
		if i == selected {
			style = theme.StyleCursor
		}
		parts = append(parts, style.Render(choice))
	}
	if selected == customIdx {
		style := theme.StyleNormal
		if customActive {
			style = theme.StyleCursor
		}
		parts = append(parts, style.Render(customView))
	}
	return strings.Join(parts, "   ")
}

func adhocElapsedLabel(state controllerpkg.State) string {
	seconds := max(0, state.AdhocElapsedSeconds)
	return fmt.Sprintf("Elapsed %dm %02ds", seconds/60, seconds%60)
}
