package ops

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// renderLog lays the ops out as a strip over a day-grouped list and a detail pane.
func renderLog(theme types.Theme, state types.ContentState) string {
	width, height := state.Width, state.Height
	indices := reverseIndices(filteredOpIndices(state.Ops, state.Filters["ops"]))
	ops := make([]api.Op, 0, len(indices))
	for _, index := range indices {
		ops = append(ops, state.Ops[index])
	}
	cursor := min(max(0, state.Cursors["ops"]), max(0, len(ops)-1))
	strip := viewchrome.StripRows(opsStrip(state.Ops), width)
	paneHeight := max(1, height-len(strip)-1)
	listWidth, detailWidth := viewchrome.LogPaneWidths(width)
	panes := viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Ops", Meta: fmt.Sprintf("%d · [+/-] limit", len(ops)), Focused: true,
		Width: listWidth, Height: paneHeight,
		Lines: opsList(theme, ops, cursor, listWidth-1, paneHeight-1),
	})
	if detailWidth > 0 {
		title, lines := "Detail", []string{theme.StyleDim.Render("Nothing selected")}
		if len(ops) > 0 {
			op := ops[cursor]
			title = humanEntity(op.Entity) + " " + string(op.Action)
			lines = opDetailLines(theme, op, detailWidth-1)
		}
		detail := viewchrome.RenderSection(theme, viewchrome.Section{
			Title: title, Width: detailWidth, Height: paneHeight, Lines: lines,
		})
		panes = viewchrome.JoinColumns(theme, paneHeight, panes, detail)
	}
	return strings.Join(append(strip, "", panes), "\n")
}

func opsStrip(ops []api.Op) [][2]string {
	counts := map[string]int{}
	oldest := ""
	for _, op := range ops {
		counts[string(op.Action)]++
		if oldest == "" || op.Timestamp < oldest {
			oldest = op.Timestamp
		}
	}
	since := "Since –"
	if oldest != "" {
		since = "Since " + localTime(oldest).Format("2006-01-02 15:04")
	}
	return [][2]string{
		{fmt.Sprintf("Ops %d", len(ops)), since},
		{fmt.Sprintf("create %d  update %d  delete %d", counts["create"], counts["update"], counts["delete"]), ""},
	}
}

func opsList(theme types.Theme, ops []api.Op, cursor, width, height int) []string {
	if len(ops) == 0 {
		return []string{theme.StyleDim.Render("No operations recorded")}
	}
	entityWidth := 0
	for _, op := range ops {
		entityWidth = max(entityWidth, ansi.StringWidth(humanEntity(op.Entity)))
	}
	rows := make([]viewchrome.DayRow, 0, len(ops))
	counts := map[string]int{}
	for _, op := range ops {
		at := localTime(op.Timestamp)
		date := at.Format("2006-01-02")
		counts[date]++
		action := fmt.Sprintf("%-6s", op.Action)
		line := at.Format("15:04:05") + "  " + fmt.Sprintf("%-*s", entityWidth, humanEntity(op.Entity)) + "  " +
			actionStyle(theme, string(op.Action)).Render(action) + "  " + op.EntityID
		rows = append(rows, viewchrome.DayRow{Date: date, Line: line})
	}
	summary := func(date string) string { return fmt.Sprint(counts[date]) }
	return viewchrome.DayList(theme, rows, cursor, width, height, summary)
}

func opDetailLines(theme types.Theme, op api.Op, width int) []string {
	fields := [][2]string{
		{"Target", op.EntityID},
		{"Time", localTime(op.Timestamp).Format("2006-01-02 15:04:05")},
		{"Device", op.DeviceID},
		{"User", op.UserID},
	}
	lines := make([]string, 0, len(fields)+8)
	for _, field := range fields {
		value := field[1]
		if strings.TrimSpace(value) == "" {
			value = "–"
		}
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("%-7s", field[0]))+" "+ansi.Truncate(value, max(1, width-8), "…"))
	}
	lines = append(lines, "", theme.StyleDim.Render("Payload"))
	return append(lines, PayloadLines(op.Payload, width)...)
}

// PayloadLines renders a JSON object as sorted, indented `key  value` lines and anything
// else as compact JSON, wrapping every line to width.
func PayloadLines(payload any, width int) []string {
	if payload == nil {
		return []string{"–"}
	}
	if object, ok := payload.(map[string]any); ok {
		return objectLines(object, 0, width)
	}
	return wrapLines(compactJSON(payload), "", width)
}

func objectLines(object map[string]any, indent, width int) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	pad := strings.Repeat(" ", indent)
	lines := []string{}
	for _, key := range keys {
		if nested, ok := object[key].(map[string]any); ok {
			lines = append(lines, ansi.Truncate(pad+key, width, "…"))
			lines = append(lines, objectLines(nested, indent+2, width)...)
			continue
		}
		value := compactJSON(object[key])
		if text, ok := object[key].(string); ok {
			value = text
		}
		lines = append(lines, wrapLines(pad+key+"  "+value, pad+"  ", width)...)
	}
	return lines
}

func wrapLines(text, indent string, width int) []string {
	lines := strings.Split(ansi.Wrap(text, max(1, width), ""), "\n")
	for i := 1; i < len(lines); i++ {
		lines[i] = ansi.Truncate(indent+lines[i], width, "")
	}
	return lines
}

func compactJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(data)
}

func localTime(timestamp string) time.Time {
	at, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return time.Time{}
	}
	return at.Local()
}

func humanEntity(entity sharedtypes.OpEntity) string {
	return strings.ReplaceAll(string(entity), "_", " ")
}

func actionStyle(theme types.Theme, action string) lipgloss.Style {
	switch action {
	case "create":
		return lipgloss.NewStyle().Foreground(theme.ColorDullGreen)
	case "update":
		return lipgloss.NewStyle().Foreground(theme.ColorYellow)
	case "delete":
		return lipgloss.NewStyle().Foreground(theme.ColorDullRed)
	default:
		return theme.StyleNormal
	}
}
