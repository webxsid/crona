package daily

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	shareddatefmt "crona/shared/datefmt"
	"crona/tui/internal/api"
	issuecore "crona/tui/internal/tui/views/issuecore"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// issueHistoryLimit is how many status changes the detail pane lists.
const issueHistoryLimit = 5

// issueStatusEntry matches the line the daemon appends to an issue's notes on a status
// change: "[<RFC3339>] <from> -> <to>" with an optional ": <note>".
var issueStatusEntry = regexp.MustCompile(`^\[([^\]]+)\] (\S+) -> (\S+)(?:: (.*))?$`)

type issueNoteEntry struct {
	at             time.Time
	from, to, note string
}

// splitIssueNotes separates free-text note lines from status-change entries, newest entry first.
func splitIssueNotes(text string) ([]string, []issueNoteEntry) {
	var free []string
	var entries []issueNoteEntry
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		match := issueStatusEntry.FindStringSubmatch(strings.TrimSpace(line))
		if match == nil {
			free = append(free, line)
			continue
		}
		at, err := time.Parse(time.RFC3339, match[1])
		if err != nil {
			free = append(free, line)
			continue
		}
		entries = append([]issueNoteEntry{{at: at, from: match[2], to: match[3], note: strings.TrimSpace(match[4])}}, entries...)
	}
	return free, entries
}

// issueNotesLines renders an issue's notes as the free text, then its status history.
func issueNotesLines(theme types.Theme, settings *api.CoreSettings, text *string, width int) []string {
	if text == nil || strings.TrimSpace(*text) == "" {
		return nil
	}
	free, entries := splitIssueNotes(*text)
	freeText := strings.TrimSpace(strings.Join(free, "\n"))
	var lines []string
	if freeText != "" {
		lines = append(lines, detailText(theme, "Notes", &freeText, width)...)
	}
	if len(entries) == 0 {
		return lines
	}
	written := map[string]bool{}
	for _, line := range free {
		written[strings.TrimSpace(line)] = true
	}
	lines = append(lines, "", theme.StyleDim.Bold(true).Render("History"))
	for _, entry := range entries[:min(len(entries), issueHistoryLimit)] {
		local := entry.at.Local()
		row := theme.StyleDim.Render(shareddatefmt.FormatDate(local, settings)+"  "+local.Format("15:04")) + "  " +
			theme.StyleNormal.Render(issuecore.PlainIssueStatus(entry.from)+" → ") + issueStatusText(theme, entry.to)
		note := entry.note
		if written[note] {
			note = ""
		}
		switch {
		case note == "":
			lines = append(lines, row)
		case ansi.StringWidth(row)+2+ansi.StringWidth(note) <= width:
			lines = append(lines, row+"  "+theme.StyleDim.Render(note))
		default:
			lines = append(lines, row)
			for _, wrapped := range wrapDetail(theme.StyleDim, note, width-2) {
				lines = append(lines, "  "+wrapped)
			}
		}
	}
	if hidden := len(entries) - issueHistoryLimit; hidden > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("+%d earlier", hidden)))
	}
	return lines
}

func issueStatusText(theme types.Theme, status string) string {
	style := theme.StyleNormal
	if s := issuecore.IssueStatusStyle(theme, status); s != nil {
		style = *s
	}
	return style.Render(issuecore.PlainIssueStatus(status))
}
