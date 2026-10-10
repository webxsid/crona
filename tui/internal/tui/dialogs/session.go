package dialogs

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	viewchrome "crona/tui/internal/tui/views/chrome"
	issuecore "crona/tui/internal/tui/views/issuecore"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func renderSessionDialog(theme Theme, state controllerpkg.State) string {
	switch state.Kind {
	case "end_session":
		title := "End Session"
		hint := dialogSubmitHint(
			state,
			"confirm",
		) + "   [ctrl+e] open details dialog   [esc] cancel"
		labels := []string{
			"Commit message",
			"Worked on",
			"Outcome",
			"Next step",
			"Blockers",
			"Links",
		}
		rows := []string{theme.StylePaneTitle.Render(title)}
		if state.AdhocTimer {
			labels[0] = "Session summary (required)"
			rows = append(rows, "", theme.StyleDim.Render(adhocElapsedLabel(state)))
			if state.AdhocResolution == sharedtypes.AdhocSessionResolutionKeep {
				rows = append(rows, theme.StyleHeader.Render("Ad hoc / Sessions"), theme.StyleDim.Render("Your summary becomes the issue title, shortened if needed. The issue is marked Done and assigned today."))
				hint = "[ctrl+s] save ad hoc & finish   [ctrl+e] details   [esc] cancel"
			} else {
				rows = append(rows, theme.StyleHeader.Render(state.ViewName), theme.StyleDim.Render(state.RepoName+" / "+state.StreamName))
				hint = "[ctrl+s] attach & finish   [ctrl+e] details   [esc] cancel"
			}
		}
		for i := range state.Inputs {
			rows = append(rows, "", theme.StyleDim.Render(labels[i]), state.Inputs[i].View())
		}
		rows = appendDialogFooter(theme, state, rows, hint)
		return formModal(theme, state, 72, rows)
	case "adhoc_resolution":
		rows := []string{theme.StylePaneTitle.Render(state.ViewTitle), theme.StyleDim.Render(adhocElapsedLabel(state))}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		for i, item := range state.ChoiceItems {
			rows = append(rows, "")
			if i == state.ChoiceCursor {
				rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+item))
			} else {
				rows = append(rows, theme.StyleNormal.Render("  "+item))
			}
			if i < len(state.ChoiceDetails) && state.ChoiceDetails[i] != "" {
				rows = append(rows, theme.StyleDim.Render("    "+state.ChoiceDetails[i]))
			}
		}
		if len(state.ChoiceItems) == 0 {
			rows = append(rows, theme.StyleDim.Render("No eligible issues are available."))
		}
		rows = appendDialogFooter(theme, state, rows, "[↑/↓] move   [enter] choose   [esc] back")
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "adhoc_issue_picker":
		rows := []string{theme.StylePaneTitle.Render(state.ViewTitle)}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		tabLabels := []string{"Due today", "Overdue", "Pinned", "All"}
		tabs := make([]string, 0, len(tabLabels))
		for i, label := range tabLabels {
			label += fmt.Sprintf(" %d", state.AdhocIssueTabCounts[i])
			if i == state.AdhocIssueTab {
				tabs = append(tabs, theme.StyleSelected.Render("["+label+"]"))
			} else {
				tabs = append(tabs, theme.StyleDim.Render(label))
			}
		}
		rows = append(rows, "", strings.Join(tabs, "   "))
		contentWidth := max(12, min(state.Width-8, 92)-6)
		header := strings.Join(rows, "\n")
		footer := "[←/→] tab   [↑/↓] move   [enter] continue   [esc] back"
		footerHeight := lipgloss.Height(lipgloss.NewStyle().Width(contentWidth).Render(footer))
		available := 30
		if state.Height > 0 {
			available = max(3, state.Height-6-lipgloss.Height(lipgloss.NewStyle().Width(contentWidth).Render(header))-footerHeight-4)
		}
		count := max(1, available/3)
		first := max(0, state.ChoiceCursor-count+1)
		last := min(len(state.ChoiceItems), first+count)
		for i := first; i < last; i++ {
			item := ansi.Truncate(strings.ReplaceAll(state.ChoiceItems[i], "\n", " "), contentWidth-2, "…")
			if i == state.ChoiceCursor {
				rows = append(rows, "", theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+item))
			} else {
				rows = append(rows, "", theme.StyleNormal.Render("  "+item))
			}
			if i < len(state.ChoiceDetails) {
				rows = append(rows, theme.StyleDim.Render("  "+ansi.Truncate(state.ChoiceDetails[i], contentWidth-2, "…")))
			}
		}
		if len(state.ChoiceItems) > 0 {
			position := fmt.Sprintf("Issue %d of %d", state.ChoiceCursor+1, len(state.ChoiceItems))
			if first > 0 {
				position += " · ↑ more"
			}
			if last < len(state.ChoiceItems) {
				position += " · ↓ more"
			}
			rows = append(rows, "", theme.StyleDim.Render(position))
		}
		if len(state.AdhocIssueIDs) == 0 {
			if state.AdhocIssueTabCounts[3] == 0 {
				rows = append(rows, "", theme.StyleDim.Render("No eligible issues are available."))
			} else {
				rows = append(rows, "", theme.StyleDim.Render("No issues in this tab. Switch tabs to see other eligible issues."))
			}
		}
		rows = appendDialogFooter(theme, state, rows, footer)
		return modal(theme, state.Width, 92, theme.ColorCyan, rows)
	case "adhoc_new_issue":
		rows := []string{theme.StylePaneTitle.Render(state.ViewTitle)}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		if state.RepoID > 0 && state.StreamID > 0 {
			rows = append(rows, "", theme.StyleDim.Render("Issue will be created in the active repo and stream."))
		} else {
			rows = append(rows, "", theme.StyleError.Render("Select an active repo and stream first."))
		}
		rows = append(rows, "", theme.StyleDim.Render("Issue title"), state.Inputs[0].View())
		rows = appendDialogFooter(theme, state, rows, "[enter] continue   [esc] back")
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "timer_start_type":
		rows := []string{theme.StylePaneTitle.Render("Start Timer")}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		if meta := renderTimerMeta(theme, state, "", 0, false); meta != "" {
			rows = append(rows, "", meta)
		}
		for i, item := range state.ChoiceItems {
			line := timerStartChoiceLabel(i, item)
			if i == state.ChoiceCursor {
				rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+line))
			} else {
				rows = append(rows, theme.StyleNormal.Render("  "+line))
			}
		}
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) {
			rows = append(rows, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
		}
		rows = appendDialogFooter(theme, state, rows, "[↑/↓] move   [enter] choose   [esc] cancel")
		return modal(theme, state.Width, 68, theme.ColorGreen, rows)
	case "pomodoro_start":
		vm := controllerpkg.BuildPomodoroDialogViewModel(state)
		rows := []string{theme.StylePaneTitle.Render("Pomodoro Session")}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		if meta := renderTimerMeta(theme, state, vm.EstimatedTotalDuration, vm.EstimatedTotalSeconds, true); meta != "" {
			rows = append(rows, "", meta)
		}
		if vm.ShowSummary {
			rows = append(
				rows,
				"",
				theme.StyleDim.Render(
					fmt.Sprintf(
						"%s  ·  %s  ·  %s",
						vm.FocusDisplay,
						vm.ShortBreakDisplay,
						vm.LongBreakDisplay,
					),
				),
			)
			rows = append(rows, theme.StyleDim.Render(
				vm.CyclesSummary,
			))
		}
		rows = append(
			rows,
			"",
			pomodoroRowLabel(theme, "Focus", vm.FocusRowActive || vm.FocusCustomActive),
		)
		rows = append(rows, pomodoroChoiceRow(
			theme,
			vm.FocusChoices,
			state.PomodoroFocusChoice,
			vm.FocusCustomChoice,
			vm.FocusCustomActive,
			state.Inputs[0].View(),
		))
		rows = append(
			rows,
			"",
			pomodoroRowLabel(theme, "Short Break", vm.BreakRowActive || vm.BreakCustomActive),
		)
		rows = append(rows, pomodoroChoiceRow(
			theme,
			vm.BreakChoices,
			state.PomodoroBreakChoice,
			vm.BreakCustomChoice,
			vm.BreakCustomActive,
			state.Inputs[1].View(),
		))
		if vm.LongBreakForcedOff {
			rows = append(rows, "", pomodoroDisabledValueRow(theme, "Long Break"))
		} else {
			rows = append(rows, "", pomodoroRowLabel(theme, "Long Break", vm.LongBreakRowActive || vm.LongBreakCustomActive))
			rows = append(rows, pomodoroChoiceRow(
				theme,
				vm.LongBreakChoices,
				vm.LongBreakSelected,
				vm.LongBreakCustomChoice,
				vm.LongBreakCustomActive,
				state.Inputs[2].View(),
			))
		}
		if vm.CyclesDisabled {
			rows = append(rows, "", pomodoroDisabledValueRow(theme, "Cycles"))
		} else {
			rows = append(rows, "", pomodoroRowLabel(theme, "Cycles", vm.CyclesRowActive))
			rows = append(rows, state.Inputs[3].View())
		}
		if vm.LongBreakDisabled {
			rows = append(rows, "", pomodoroDisabledValueRow(theme, "Long Break"))
		} else {
			rows = append(rows, "", pomodoroRowLabel(theme, "Long Break After", vm.LongBreakCyclesActive))
			rows = append(rows, state.Inputs[4].View())
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[←/→] choose   [↑/↓] move   "+dialogSubmitHint(state, "start")+"   [esc] back",
		)
		return modal(theme, state.Width, 72, theme.ColorGreen, rows)
	case "timer_countdown_start":
		rows := []string{theme.StylePaneTitle.Render("Timer Session")}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		if meta := renderTimerMeta(theme, state, "", state.TimerCountdownSeconds, true); meta != "" {
			rows = append(rows, "", meta)
		}
		rows = append(rows, "", theme.StyleDim.Render("Focus Time"), state.Inputs[0].View())
		rows = append(
			rows,
			"",
			theme.StyleDim.Render(
				"Single countdown. No breaks. Reuses the hard-limit session completion flow.",
			),
		)
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			dialogSubmitHint(state, "start")+"   [esc] back",
		)
		return modal(theme, state.Width, 68, theme.ColorGreen, rows)
	case "hard_limit_expired":
		title := "Pomodoro Session Complete"
		description := "Choose how to finish this pomodoro session."
		if state.HardLimitKind == sharedtypes.TimerHardLimitKindCountdown {
			title = "Timer Session Complete"
			description = "Choose how to finish this timer session."
		}
		rows := []string{theme.StylePaneTitle.Render(title)}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		rows = append(rows, "", theme.StyleDim.Render(description))
		for i, item := range state.ChoiceItems {
			line := "  " + item
			if i == state.ChoiceCursor {
				rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+item))
			} else {
				rows = append(rows, theme.StyleNormal.Render(line))
			}
		}
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) {
			rows = append(rows, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
		}
		rows = appendDialogFooter(theme, state, rows, "[↑/↓] move   [enter] choose")
		return modal(theme, state.Width, 68, theme.ColorYellow, rows)
	case "hard_limit_extend":
		vm := controllerpkg.BuildPomodoroDialogViewModel(state)
		rows := []string{theme.StylePaneTitle.Render("Extend Pomodoro Session")}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		if totalPreview := controllerpkg.PreviewPomodoroExtendDuration(state); totalPreview != "" {
			rows = append(rows, "", theme.StyleSelected.Render(totalPreview))
		}
		if vm.ShowSummary {
			rows = append(
				rows,
				"",
				theme.StyleDim.Render(
					fmt.Sprintf(
						"%s  ·  %s  ·  %s",
						vm.FocusDisplay,
						vm.ShortBreakDisplay,
						vm.LongBreakDisplay,
					),
				),
			)
			rows = append(rows, theme.StyleDim.Render(vm.CyclesSummary))
		}
		rows = append(
			rows,
			"",
			pomodoroRowLabel(theme, "Focus", vm.FocusRowActive || vm.FocusCustomActive),
		)
		rows = append(
			rows,
			pomodoroChoiceRow(
				theme,
				vm.FocusChoices,
				state.PomodoroFocusChoice,
				vm.FocusCustomChoice,
				vm.FocusCustomActive,
				state.Inputs[0].View(),
			),
		)
		rows = append(
			rows,
			"",
			pomodoroRowLabel(theme, "Short Break", vm.BreakRowActive || vm.BreakCustomActive),
		)
		rows = append(
			rows,
			pomodoroChoiceRow(
				theme,
				vm.BreakChoices,
				state.PomodoroBreakChoice,
				vm.BreakCustomChoice,
				vm.BreakCustomActive,
				state.Inputs[1].View(),
			),
		)
		if vm.LongBreakForcedOff {
			rows = append(rows, "", pomodoroDisabledValueRow(theme, "Long Break"))
		} else {
			rows = append(rows, "", pomodoroRowLabel(theme, "Long Break", vm.LongBreakRowActive || vm.LongBreakCustomActive))
			rows = append(rows, pomodoroChoiceRow(theme, vm.LongBreakChoices, vm.LongBreakSelected, vm.LongBreakCustomChoice, vm.LongBreakCustomActive, state.Inputs[2].View()))
		}
		if vm.CyclesDisabled {
			rows = append(rows, "", pomodoroDisabledValueRow(theme, "Cycles"))
		} else {
			rows = append(rows, "", pomodoroRowLabel(theme, "Cycles", vm.CyclesRowActive))
			rows = append(rows, state.Inputs[3].View())
		}
		if vm.LongBreakDisabled {
			rows = append(rows, "", pomodoroDisabledValueRow(theme, "Long Break"))
		} else {
			rows = append(rows, "", pomodoroRowLabel(theme, "Long Break After", vm.LongBreakCyclesActive))
			rows = append(rows, state.Inputs[4].View())
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[←/→] choose   [↑/↓] move   [enter] extend   [esc] back",
		)
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "timer_countdown_extend":
		rows := []string{theme.StylePaneTitle.Render("Extend Timer Session")}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		rows = append(
			rows,
			"",
			theme.StyleDim.Render("Additional Time"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Adds time to the current countdown. No breaks or cycles."),
		)
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			dialogSubmitHint(state, "extend")+"   [esc] back",
		)
		return modal(theme, state.Width, 68, theme.ColorCyan, rows)
	case "amend_session":
		rows := []string{
			theme.StylePaneTitle.Render("Amend Session"),
			"",
			theme.StyleDim.Render("Commit message"),
			state.Inputs[0].View(),
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			dialogSubmitHint(state, "save")+"   [esc] cancel",
		)
		return modal(theme, state.Width, 68, theme.ColorCyan, rows)
	case "edit_session_history":
		rows := []string{theme.StylePaneTitle.Render("Edit Session History")}
		if state.SessionHistoryContext != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.SessionHistoryContext))
		}
		labels := []string{"Session start (YYYY-MM-DD HH:mm:ss)", "Session end (YYYY-MM-DD HH:mm:ss)"}
		for i := range state.Inputs {
			rows = append(rows, "", theme.StyleDim.Render(labels[i]), state.Inputs[i].View())
		}
		for i, label := range []string{"Commit / summary", "Work summary", "Notes"} {
			if i < len(state.SessionHistoryNoteInputs) {
				rows = append(rows, "", theme.StyleDim.Render(label), state.SessionHistoryNoteInputs[i].View())
			}
		}
		rows = append(rows, "", theme.StyleDim.Render("Work and break segments"))
		const rowHeight = 3
		visibleRows := max(1, (state.Height-len(rows)-5)/rowHeight)
		start := 0
		if len(state.SessionHistorySegments) > visibleRows {
			focused := max(0, (state.FocusIdx-5)/3)
			start = min(max(0, focused-visibleRows/2), len(state.SessionHistorySegments)-visibleRows)
		}
		end := min(len(state.SessionHistorySegments), start+visibleRows)
		if start > 0 {
			rows = append(rows, theme.StyleDim.Render(fmt.Sprintf("… %d earlier segment(s)", start)))
		}
		for i := start; i < end; i++ {
			segment := state.SessionHistorySegments[i]
			rows = append(rows, historySegmentTypeRow(theme, state, segment, i))
			rows = append(rows, segment.StartTime.View()+"  →  "+segment.EndTime.View())
		}
		if end < len(state.SessionHistorySegments) {
			rows = append(rows, theme.StyleDim.Render(fmt.Sprintf("… %d later segment(s)", len(state.SessionHistorySegments)-end)))
		}
		if len(state.SessionHistorySegments) == 0 {
			rows = append(rows, theme.StyleDim.Render("No segments. Press [a] to add one."))
		}
		rows = appendDialogFooter(theme, state, rows, dialogSubmitHint(state, "save")+"   [tab] next   [ctrl+a] add segment   [ctrl+d] remove   [←/→] change type   [esc] cancel")
		return modal(theme, state.Width, 100, theme.ColorCyan, rows)
	case "manual_session":
		rows := []string{theme.StylePaneTitle.Render("Log Session")}
		if state.IssueID != 0 {
			label := "Issue #" + itoa(state.IssueID)
			if state.ViewName != "" {
				label += "  " + state.ViewName
			}
			if state.IssueEstimateMins != nil && *state.IssueEstimateMins > 0 {
				label += fmt.Sprintf(
					"  · estimate %s",
					controllerpkg.FormatDurationMinutesInput(state.IssueEstimateMins),
				)
			}
			rows = append(rows, "", theme.StyleDim.Render(label))
		}
		labels := []string{
			"Summary",
			"Date",
			"Work duration",
			"Break duration",
			"Start time",
			"End time",
			"Notes",
		}
		for i := range state.Inputs {
			rows = append(rows, "", theme.StyleDim.Render(labels[i]), state.Inputs[i].View())
		}
		rows = appendDialogFooter(theme, state, rows, manualSessionHint(state))
		return modal(theme, state.Width, 72, theme.ColorGreen, rows)
	case "issue_session_transition":
		title := "End Session?"
		body := "Mark this issue and end the active session?"
		hint := "[y/enter] confirm   [n/esc] cancel"
		border := theme.ColorYellow
		switch state.IssueStatus {
		case "done":
			title, body, border = "Complete Issue", "Mark the issue done and end the active session.", theme.ColorGreen
		case "abandoned":
			title, body, border = "Abandon Issue", "Abandon the issue and end the active session.", theme.ColorRed
		}
		rows := []string{theme.StylePaneTitle.Render(title), "", body}
		if (state.IssueStatus == "done" || state.IssueStatus == "abandoned") &&
			len(state.Inputs) > 0 {
			hint = dialogSubmitHint(state, "confirm") + "   [esc] cancel"
			label := "Abandon reason"
			if state.IssueStatus == "done" {
				label = "Completion note"
			}
			rows = append(rows, "", theme.StyleDim.Render(label), state.Inputs[0].View())
		}
		rows = appendDialogFooter(theme, state, rows, hint)
		return modal(theme, state.Width, 68, border, rows)
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
	labelStyle := theme.StyleDim.PaddingRight(2)
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

func pomodoroRowLabel(theme Theme, label string, active bool) string {
	if active {
		return theme.StyleCursor.Render("> " + label)
	}
	return theme.StyleDim.Render("  " + label)
}

func pomodoroDisabledValueRow(theme Theme, label string) string {
	return theme.StyleDim.Render(label + ": disabled")
}

func itoa(v int64) string {
	return fmt.Sprintf("%d", v)
}

func manualSessionHint(state controllerpkg.State) string {
	switch state.FocusIdx {
	case 1:
		return "[ctrl+e] pick date   [g] today   [tab] next   " + dialogSubmitHint(
			state,
			"save",
		) + "   [esc] cancel"
	default:
		return "[tab] next   durations: 90, 90m, 1h30m   " + dialogSubmitHint(
			state,
			"save",
		) + "   [esc] cancel"
	}
}

func adhocElapsedLabel(state controllerpkg.State) string {
	seconds := max(0, state.AdhocElapsedSeconds)
	return fmt.Sprintf("Elapsed %dm %02ds", seconds/60, seconds%60)
}
