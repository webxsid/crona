package sessions

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	sessionmeta "crona/tui/internal/tui/views/sessionmeta"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var noteSectionOrder = []sharedtypes.SessionNoteSection{
	sharedtypes.SessionNoteSectionCommit,
	sharedtypes.SessionNoteSectionContext,
	sharedtypes.SessionNoteSectionWork,
	sharedtypes.SessionNoteSectionNotes,
}

// renderLog lays the session history out as a strip over a day-grouped list and a detail pane.
func renderLog(theme types.Theme, state types.ContentState) string {
	width, height := state.Width, state.Height
	indices := sessionmeta.FilteredSessionIndices(state.SessionHistory, state.Filters["sessions"])
	entries := make([]api.SessionHistoryEntry, 0, len(indices))
	for _, index := range indices {
		entries = append(entries, state.SessionHistory[index])
	}
	cursor := min(max(0, state.Cursors["sessions"]), max(0, len(entries)-1))
	strip := viewchrome.StripRows(
		sessionStrip(state.SessionHistory, state.WeekStart, time.Now()),
		width,
	)
	paneHeight := max(1, height-len(strip)-1)
	listWidth, detailWidth := viewchrome.LogPaneWidths(width)
	panes := viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Sessions", Meta: fmt.Sprint(len(entries)), Focused: true,
		Width: listWidth, Height: paneHeight,
		Lines: sessionList(theme, entries, cursor, listWidth-1, paneHeight-1),
	})
	if detailWidth > 0 {
		title, lines := "Detail", []string{theme.StyleDim.Render("Nothing selected")}
		if len(entries) > 0 {
			entry := entries[cursor]
			title = localTime(
				entry.StartTime,
			).Format("Mon 2 Jan") +
				"  " + timeRange(
				entry.StartTime,
				entry.EndTime,
			)
			lines = sessionDetailLines(theme, state, entry, detailWidth-1)
		}
		detail := viewchrome.RenderSection(theme, viewchrome.Section{
			Title: title, Width: detailWidth, Height: paneHeight, Lines: lines,
		})
		panes = viewchrome.JoinColumns(theme, paneHeight, panes, detail)
	}
	return strings.Join(append(strip, "", panes), "\n")
}

func sessionStrip(
	history []api.SessionHistoryEntry,
	weekStart sharedtypes.WeekStart,
	now time.Time,
) [][2]string {
	total, longest, week := 0, 0, 0
	from := startOfWeek(now.Local(), weekStart)
	for _, entry := range history {
		seconds := durationSeconds(entry)
		total += seconds
		longest = max(longest, seconds)
		end := entry.StartTime
		if entry.EndTime != nil && *entry.EndTime != "" {
			end = *entry.EndTime
		}
		if at := localTime(end); !at.Before(from) && at.Before(from.AddDate(0, 0, 7)) {
			week += seconds
		}
	}
	avg := 0
	if len(history) > 0 {
		avg = total / len(history)
	}
	return [][2]string{
		{fmt.Sprintf("Sessions %d", len(history)), "This week " + compact(week)},
		{"total " + compact(total), "longest " + compact(longest)},
		{"avg " + compact(avg), ""},
	}
}

func sessionList(
	theme types.Theme,
	entries []api.SessionHistoryEntry,
	cursor, width, height int,
) []string {
	if len(entries) == 0 {
		return []string{theme.StyleDim.Render("No sessions recorded")}
	}
	rows := make([]viewchrome.DayRow, 0, len(entries))
	counts, sums := map[string]int{}, map[string]int{}
	for _, entry := range entries {
		date := localTime(entry.StartTime).Format("2006-01-02")
		counts[date]++
		sums[date] += durationSeconds(entry)
		note := theme.StyleDim.Render("no note")
		if first := firstLine(sessionNote(entry)); first != "" {
			note = first
		}
		line := fmt.Sprintf(
			"%-11s %6s  ",
			timeRange(entry.StartTime, entry.EndTime),
			compact(durationSeconds(entry)),
		) + note
		rows = append(rows, viewchrome.DayRow{Date: date, Line: line})
	}
	summary := func(date string) string { return fmt.Sprintf("%d · %s", counts[date], compact(sums[date])) }
	return viewchrome.DayList(theme, rows, cursor, width, height, summary)
}

func sessionDetailLines(
	theme types.Theme,
	state types.ContentState,
	entry api.SessionHistoryEntry,
	width int,
) []string {
	label := func(name, value string) string {
		return theme.StyleDim.Render(
			fmt.Sprintf("%-7s", name),
		) + " " + ansi.Truncate(
			value,
			max(1, width-8),
			"…",
		)
	}
	lines := []string{}
	preview := state.SessionPreview
	if preview == nil || preview.ID != entry.ID {
		lines = append(
			lines,
			label("Time", timeRange(entry.StartTime, entry.EndTime)),
			label("Length", compact(durationSeconds(entry))),
		)
		status := "Loading…"
		if state.SessionPreviewFailed {
			status = "Details unavailable"
		}
		lines = append(lines, theme.StyleDim.Render(status))
		return append(lines, noteLines(theme, entry, width)...)
	}
	lines = append(
		lines,
		label("Issue", orDash(preview.IssueTitle)),
		label("Repo", orDash(preview.RepoName)),
		label("Stream", orDash(preview.StreamName)),
		"",
		label(
			"Work",
			compact(
				preview.WorkSummary.WorkSeconds,
			)+"  "+theme.StyleDim.Render(
				"rest "+compact(preview.WorkSummary.RestSeconds),
			),
		),
		"        "+workRestBar(theme, preview.WorkSummary, max(1, width-12)),
	)
	if len(preview.Segments) > 0 {
		lines = append(lines, "", theme.StyleDim.Render("Segments"))
		for _, segment := range preview.Segments {
			seconds := 0
			if segment.EndTime != nil {
				seconds = int(
					localTime(*segment.EndTime).Sub(localTime(segment.StartTime)).Seconds(),
				)
			}
			kind := strings.ReplaceAll(string(segment.SegmentType), "_", " ")
			lines = append(
				lines,
				ansi.Truncate(
					fmt.Sprintf(
						"%s  %s  %s",
						kind,
						timeRange(segment.StartTime, segment.EndTime),
						compact(seconds),
					),
					width,
					"…",
				),
			)
		}
	}
	return append(lines, noteLines(theme, preview.SessionHistoryEntry, width)...)
}

func noteLines(theme types.Theme, entry api.SessionHistoryEntry, width int) []string {
	lines := []string{}
	wrap := func(text string) []string {
		return strings.Split(ansi.Wordwrap(strings.TrimSpace(text), max(1, width-1), ""), "\n")
	}
	found := false
	for _, section := range noteSectionOrder {
		text := strings.TrimSpace(entry.ParsedNotes[section])
		if text == "" {
			continue
		}
		found = true
		lines = append(
			lines,
			"",
			theme.StyleDim.Render(strings.ToUpper(string(section[:1]))+string(section[1:])),
		)
		lines = append(lines, wrap(text)...)
	}
	if !found && entry.Notes != nil && strings.TrimSpace(*entry.Notes) != "" {
		lines = append(lines, "", theme.StyleDim.Render("Note"))
		lines = append(lines, wrap(*entry.Notes)...)
	}
	return lines
}

func workRestBar(theme types.Theme, summary sharedtypes.SessionWorkSummary, width int) string {
	total := summary.WorkSeconds + summary.RestSeconds
	if total <= 0 {
		return theme.StyleDim.Render(strings.Repeat("─", width))
	}
	work := summary.WorkSeconds * width / total
	return lipgloss.NewStyle().Foreground(theme.ColorDullGreen).Render(strings.Repeat("█", work)) +
		theme.StyleDim.Render(strings.Repeat("█", width-work))
}

func sessionNote(entry api.SessionHistoryEntry) string {
	for _, section := range []sharedtypes.SessionNoteSection{sharedtypes.SessionNoteSectionCommit, sharedtypes.SessionNoteSectionNotes} {
		if text := strings.TrimSpace(entry.ParsedNotes[section]); text != "" {
			return text
		}
	}
	if entry.Notes != nil {
		return strings.TrimSpace(*entry.Notes)
	}
	return ""
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return strings.TrimSpace(line)
}

func timeRange(start string, end *string) string {
	to := "now"
	if end != nil && *end != "" {
		to = localTime(*end).Format("15:04")
	}
	return localTime(start).Format("15:04") + "–" + to
}

func durationSeconds(entry api.SessionHistoryEntry) int {
	if entry.DurationSeconds == nil {
		return 0
	}
	return *entry.DurationSeconds
}

func compact(seconds int) string {
	return helperpkg.FormatCompactDurationMinutes(seconds / 60)
}

func localTime(timestamp string) time.Time {
	at, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return time.Time{}
	}
	return at.Local()
}

func startOfWeek(now time.Time, weekStart sharedtypes.WeekStart) time.Time {
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	offset := (int(day.Weekday()) + 6) % 7
	if weekStart == sharedtypes.WeekStartSunday {
		offset = int(day.Weekday())
	}
	return day.AddDate(0, 0, -offset)
}

func orDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "–"
	}
	return value
}
