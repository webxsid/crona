package tui

import (
	"fmt"
	"strings"
	"time"

	"crona/tui/internal/api"

	"github.com/charmbracelet/lipgloss"
)

const pageSize = 12

// ---------- Root View ----------

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Base layout
	base := strings.Join([]string{
		m.renderHeader(),
		m.renderBody(),
		m.renderHelpBar(),
	}, "\n")

	if m.statusMsg != "" {
		base += "\n" + styleError.Render(m.statusMsg)
	}

	// When a dialog is open, centre it over the screen using lipgloss.Place.
	if m.dialog != "" {
		dialogStr := m.renderDialog()
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			dialogStr,
		)
	}

	return base
}

// ---------- Header ----------

func (m Model) renderHeader() string {
	repo := deref(nil)
	stream := deref(nil)
	if m.context != nil {
		repo = firstNonEmpty(m.context.RepoName, m.context.RepoID)
		stream = firstNonEmpty(m.context.StreamName, m.context.StreamID)
	}
	contextLine := fmt.Sprintf(
		"%s %s   %s %s",
		styleDim.Render("repo:"), styleHeader.Render(truncate(repo, max(16, m.width/4))),
		styleDim.Render("stream:"), styleHeader.Render(truncate(stream, max(16, m.width/4))),
	)

	lines := []string{contextLine}
	if secondary := m.renderHeaderSessionLine(); secondary != "" {
		lines = append(lines, secondary)
	}

	return lipgloss.NewStyle().
		Width(m.width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(colorDim).
		Render(strings.Join(lines, "\n"))
}

func (m Model) renderBody() string {
	height := m.contentHeight()
	sidebar := m.renderSidebar(height)
	content := m.renderContent()
	return lipgloss.NewStyle().
		Width(m.width).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content))
}

// ---------- Sidebar ----------

func (m Model) renderSidebar(height int) string {
	width := m.sidebarWidth()
	lines := []string{
		stylePaneTitle.Render("Views"),
		"",
		styleDim.Render("WORKSPACE"),
		m.renderSidebarItem(ViewDefault, "Default"),
		m.renderSidebarItem(ViewMeta, "Meta"),
		m.renderSidebarItem(ViewScratch, "Scratchpads"),
		m.renderSidebarItem(ViewOps, "Ops"),
		"",
		styleDim.Render("DASHBOARD"),
		m.renderSidebarItem(ViewDaily, "Daily Dashboard"),
		"",
		styleDim.Render("SETTINGS"),
		styleDim.Render("  Coming Soon"),
	}

	if (m.timer != nil && m.timer.State != "idle") || m.view == ViewSession {
		lines = append([]string{
			styleDim.Render("SESSION"),
			m.renderSidebarItem(ViewSession, "Session"),
			"",
		}, lines...)
	}

	return styleInactive.
		Width(width-4).
		Height(max(3, height-2)).
		Padding(1, 1).
		Render(strings.Join(lines, "\n"))
}

func (m Model) renderSidebarItem(view View, label string) string {
	if m.view == view {
		return styleCursor.Render("▶ " + label)
	}
	return styleNormal.Render("  " + label)
}

// ---------- Content area ----------

func (m Model) renderContent() string {
	availH := m.contentHeight()
	width := m.mainContentWidth()

	switch m.view {
	case ViewDefault:
		return m.renderDefaultView(width, availH)
	case ViewDaily:
		return m.renderDailyView(width, availH)
	case ViewMeta:
		return m.renderMetaView(width, availH)
	case ViewSession:
		return m.renderSessionView(width, availH)
	case ViewScratch:
		return m.renderScratchpadView(width, availH)
	case ViewOps:
		return m.renderOpsView(width, availH)
	}
	return ""
}

// ---------- Default view ----------

func (m Model) renderDefaultView(width, h int) string {
	return m.renderDefaultIssuesPane(width, h)
}

func (m Model) renderDailyView(width, h int) string {
	summaryH := max(9, h/3)
	listH := h - summaryH
	if listH < 8 {
		listH = 8
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderDailySummary(width, summaryH),
		m.renderDailyIssuesPane(width, listH),
	)
}

// ---------- Meta view ----------

func (m Model) renderMetaView(width, h int) string {
	leftW := width

	topH := h * 30 / 100
	botH := h - topH

	streamsEmpty := "No streams — [a] create new"
	if m.context == nil || m.context.RepoID == nil {
		streamsEmpty = "No repo checked out — [1] then [c]"
	}
	issuesEmpty := "No issues — [a] create new"
	if m.context == nil || m.context.StreamID == nil {
		issuesEmpty = "No stream checked out — [2] then [c]"
	}

	repoPane := m.renderPane("Repos [1]", PaneRepos, leftW/2, topH, "No repos — [a] create new")
	streamPane := m.renderPane("Streams [2]", PaneStreams, leftW/2, topH, streamsEmpty)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, repoPane, streamPane)

	issuePane := m.renderMetaIssuesPane(leftW, botH, issuesEmpty)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, issuePane)
}

// ---------- Session view ----------

func (m Model) renderSessionView(width, h int) string {
	if m.timer == nil || m.timer.State == "idle" {
		msg := lipgloss.NewStyle().
			Width(width).Height(h).
			AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center).
			Render("No active session\n\n[s] start timer on active issue")
		return msg
	}

	var activeIssue *api.IssueWithMeta
	if m.timer.IssueID != nil {
		for i := range m.allIssues {
			if m.allIssues[i].ID == *m.timer.IssueID {
				activeIssue = &m.allIssues[i]
				break
			}
		}
	}

	total := m.timer.ElapsedSeconds + m.elapsed
	elapsed := formatClock(total)

	seg := "work"
	if m.timer.SegmentType != nil {
		seg = *m.timer.SegmentType
	}
	stateColor := colorGreen
	timerTitle := "Focus Session"
	timerHint := "p=pause  x=end  z=stash  ]=scratchpads"
	if m.timer.State == "paused" {
		stateColor = colorYellow
		timerTitle = "Paused For"
		timerHint = "r=resume  x=end  z=stash  ]=scratchpads"
		seg = "paused"
	}

	issueBox := "No issue selected"
	if activeIssue != nil {
		issueBox = fmt.Sprintf("[%s/%s]\n%s", activeIssue.RepoName, activeIssue.StreamName, activeIssue.Title)
	}

	leftW := width - 4
	timerText := renderBigClock(elapsed)
	priorWorkedSeconds, completedSessions := summarizeCompletedSessions(m.issueSessions)
	progress := styleDim.Render(fmt.Sprintf("Completed sessions: %d", completedSessions))
	if activeIssue != nil && activeIssue.EstimateMinutes != nil {
		progress += "\n" + styleDim.Render(formatEstimateProgress(priorWorkedSeconds+total, *activeIssue.EstimateMinutes))
	}

	timerSection := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(stateColor).
		Padding(1, 2).Width(leftW).
		Render(fmt.Sprintf("%s\n\n%s\n\n%s%s",
			timerTitle,
			lipgloss.NewStyle().Foreground(stateColor).Bold(true).Render(timerText),
			styleDim.Render(strings.ToUpper(seg)),
			"\n\n"+progress,
		))

	issueSection := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).BorderForeground(colorCyan).
		Padding(1, 2).Width(leftW).
		Render("Active Issue\n\n" + issueBox + "\n\n" + styleDim.Render(timerHint))

	left := lipgloss.JoinVertical(lipgloss.Left, issueSection, timerSection)
	return left
}

// ---------- Ops view ----------

func (m Model) renderOpsView(width, h int) string {
	active := m.pane == PaneOps
	cur := m.cursor[PaneOps]
	indices := reverseIndices(m.filteredIndices(PaneOps))
	total := len(indices)

	inner := h - 7 // border (2) + title (1) + limit (1) + filter (1) + header (1) + padding (1)
	if inner < 1 {
		inner = 1
	}

	var lines []string
	lines = append(lines, stylePaneTitle.Render("Ops Log"))
	lines = append(lines, styleDim.Render(fmt.Sprintf("limit: %d  [+/-] adjust", m.currentOpsLimit())))
	lines = append(lines, m.renderFilterLine(PaneOps, width-6))

	if total == 0 {
		lines = append(lines, styleDim.Render("No operations recorded"))
	} else {
		timeW := 19
		entityW := max(10, width/8)
		actionW := 10
		markerW := 2
		targetW := width - timeW - entityW - actionW - markerW - 10
		if targetW < 12 {
			targetW = 12
		}

		header := fmt.Sprintf(
			"%-2s %-19s %-*s %-*s %s",
			"",
			"Time",
			entityW, "Entity",
			actionW, "Action",
			"Target",
		)
		lines = append(lines, styleDim.Render(truncate(header, width-6)))

		half := inner / 2
		start := cur - half
		if start < 0 {
			start = 0
		}
		if start+inner > total {
			start = total - inner
		}
		if start < 0 {
			start = 0
		}

		if start > 0 {
			lines = append(lines, styleDim.Render(fmt.Sprintf("   ↑ %d more", start)))
		}

		end := start + inner
		if end > total {
			end = total
		}

		for i := start; i < end; i++ {
			op := m.ops[indices[i]]
			ts := op.Timestamp
			if len(ts) >= 19 {
				ts = strings.Replace(ts[:19], "T", " ", 1)
			}
			target := op.EntityID
			if len(target) > 8 {
				target = target[:8]
			}
			row := fmt.Sprintf(
				"%-2s %-19s %-*s %-*s %s",
				"",
				ts,
				entityW, truncate(op.Entity, entityW),
				actionW, truncate(op.Action, actionW),
				truncate(target, targetW),
			)

			if i == cur && active {
				lines = append(lines, styleCursor.Render("▶ "+truncate(row[2:], width-6)))
			} else if i == cur {
				lines = append(lines, styleSelected.Render("  "+truncate(row[2:], width-6)))
			} else {
				lines = append(lines, styleNormal.Render("  "+truncate(row[2:], width-6)))
			}
		}

		remaining := total - end
		if remaining > 0 {
			lines = append(lines, styleDim.Render(fmt.Sprintf("   ↓ %d more", remaining)))
		}
	}

	content := strings.Join(lines, "\n")
	box := styleInactive
	if active {
		box = styleActive
	}
	return box.
		Width(width-2).
		Height(h-2).
		Padding(0, 1).
		Render(content)
}

// ---------- Generic pane renderer ----------

func (m Model) renderPane(
	title string,
	pane Pane,
	width, height int,
	emptyText string,
) string {
	active := m.pane == pane
	cur := m.cursor[pane]
	items := m.paneItems(pane)
	indices := m.filteredIndices(pane)
	total := len(indices)

	inner := height - 5 // border (2) + title (1) + filter (1) + padding (1)
	if inner < 1 {
		inner = 1
	}

	var lines []string
	lines = append(lines, stylePaneTitle.Render(title))
	lines = append(lines, m.renderFilterLine(pane, width-6))

	if total == 0 {
		lines = append(lines, styleDim.Render(emptyText))
	} else {
		start, end := listWindow(cur, total, inner)
		if start > 0 {
			lines = append(lines, styleDim.Render(fmt.Sprintf("↑ %d more", start)))
		}
		for i := start; i < end; i++ {
			rawIdx := indices[i]
			text := ""
			if rawIdx >= 0 && rawIdx < len(items) {
				text = items[rawIdx]
			}
			lines = append(lines, m.renderPaneRow(i, cur, active, text, width))
		}

		remaining := total - end
		if remaining > 0 {
			lines = append(lines, styleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
		}
	}

	content := strings.Join(lines, "\n")

	return m.renderPaneBox(active, width, height, content)
}

// ---------- Help bar ----------

func (m Model) renderHelpBar() string {
	actions := m.paneActions()
	left := strings.Join(actions, "   ")
	right := styleDim.Render("[[] []] switch view   [K] stop kernel   [q] quit")
	if m.timer != nil && m.timer.State != "idle" {
		right = styleDim.Render("session/scratchpads only   [K] stop kernel   [q] quit")
	}

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return " " + left + strings.Repeat(" ", gap-1) + right
}

func (m Model) paneActions() []string {
	base := []string{
		styleHeader.Render("[tab]") + styleDim.Render(" next pane"),
	}

	if m.view == ViewSession {
		return []string{
			styleHeader.Render("[p/r]") + styleDim.Render(" pause/resume"),
			styleHeader.Render("[x]") + styleDim.Render(" end"),
			styleHeader.Render("[z]") + styleDim.Render(" stash"),
			styleHeader.Render("[t/A]") + styleDim.Render(" issue"),
			styleHeader.Render("[[] []]") + styleDim.Render(" session/scratchpads"),
		}
	}

	switch m.pane {
	case PaneRepos, PaneStreams:
		return append(base,
			styleHeader.Render("[c]")+styleDim.Render(" checkout"),
			styleHeader.Render("[/]")+styleDim.Render(" filter"),
			styleHeader.Render("[a]")+styleDim.Render(" new"),
		)
	case PaneIssues:
		if m.view == ViewDaily {
			return append(base,
				styleHeader.Render("[s]")+styleDim.Render(" focus"),
				styleHeader.Render("[t]")+styleDim.Render(" cycle status"),
				styleHeader.Render("[A]")+styleDim.Render(" abandon"),
				styleHeader.Render("[y]")+styleDim.Render(" today"),
				styleHeader.Render("[D]")+styleDim.Render(" due date"),
				styleHeader.Render("[,/.]")+styleDim.Render(" date"),
				styleHeader.Render("[/]")+styleDim.Render(" filter"),
			)
		}
		return append(base,
			styleHeader.Render("[s]")+styleDim.Render(" focus"),
			styleHeader.Render("[t]")+styleDim.Render(" cycle status"),
			styleHeader.Render("[A]")+styleDim.Render(" abandon"),
			styleHeader.Render("[y]")+styleDim.Render(" today"),
			styleHeader.Render("[D]")+styleDim.Render(" due date"),
			styleHeader.Render("[/]")+styleDim.Render(" filter"),
			styleHeader.Render("[a]")+styleDim.Render(" new"),
		)
	case PaneScratchpads:
		if m.scratchpadOpen {
			actions := append(base,
				styleHeader.Render("[h/l]")+styleDim.Render(" switch"),
				styleHeader.Render("[e]")+styleDim.Render(" edit"),
				styleHeader.Render("[esc]")+styleDim.Render(" close"),
			)
			if m.timer != nil && m.timer.State != "idle" {
				actions = append(actions, styleHeader.Render("[t/A]")+styleDim.Render(" issue"))
			}
			return actions
		}
		actions := append(base,
			styleHeader.Render("[enter]")+styleDim.Render(" open"),
			styleHeader.Render("[/]")+styleDim.Render(" filter"),
			styleHeader.Render("[a]")+styleDim.Render(" new"),
			styleHeader.Render("[d]")+styleDim.Render(" delete"),
		)
		if m.timer != nil && m.timer.State != "idle" {
			actions = append(actions, styleHeader.Render("[t/A]")+styleDim.Render(" issue"))
		}
		return actions
	case PaneOps:
		return append(base,
			styleHeader.Render("[+]")+styleDim.Render(" more"),
			styleHeader.Render("[-]")+styleDim.Render(" less"),
			styleHeader.Render("[/]")+styleDim.Render(" filter"),
		)
	}
	return base
}

func (m Model) renderDefaultIssuesPane(width, height int) string {
	active := m.pane == PaneIssues
	cur := m.cursor[PaneIssues]
	indices := m.filteredIndices(PaneIssues)
	total := len(indices)

	inner := height - 6
	if inner < 1 {
		inner = 1
	}

	repoW := max(12, width/8)
	streamW := max(10, width/8)
	statusW := 11
	estimateW := 8
	titleW := width - repoW - streamW - statusW - estimateW - 14
	if titleW < 16 {
		titleW = 16
	}

	lines := []string{
		stylePaneTitle.Render("Issues [1]"),
		m.renderFilterLine(PaneIssues, width-6),
		styleDim.Render(truncate(fmt.Sprintf("%-2s %-*s %-*s %-*s %-*s %-*s", "", titleW, "Issue", statusW, "Status", estimateW, "Estimate", repoW, "Repo", streamW, "Stream"), width-6)),
	}

	if total == 0 {
		lines = append(lines, styleDim.Render("No issues — [a] create new"))
	} else {
		start, end := listWindow(cur, total, inner)
		if start > 0 {
			lines = append(lines, styleDim.Render(fmt.Sprintf("   ↑ %d more", start)))
		}
		for i := start; i < end; i++ {
			issue := m.allIssues[indices[i]]
			estimate := "-"
			if issue.EstimateMinutes != nil {
				estimate = fmt.Sprintf("%dm", *issue.EstimateMinutes)
			}
			title := issue.Title + issueDueSuffix(issue.TodoForDate)
			row := fmt.Sprintf(
				"%-2s %-*s %-*s %-*s %-*s %-*s",
				"",
				titleW, truncate(title, titleW),
				statusW, truncate(plainIssueStatus(issue.Status), statusW),
				estimateW, estimate,
				repoW, truncate(issue.RepoName, repoW),
				streamW, truncate(issue.StreamName, streamW),
			)
			lines = append(lines, m.renderPaneRowStyled(i, cur, active, row, renderIssueStatus(issue.Status), width))
		}
		if remaining := total - end; remaining > 0 {
			lines = append(lines, styleDim.Render(fmt.Sprintf("   ↓ %d more", remaining)))
		}
	}

	return m.renderPaneBox(active, width, height, strings.Join(lines, "\n"))
}

func (m Model) renderDailySummary(width, height int) string {
	dateText := m.currentDashboardDate()
	totalIssues := 0
	totalEstimate := 0
	completedCount := 0
	abandonedCount := 0
	workedSeconds := 0
	if m.dailySummary != nil {
		dateText = m.dailySummary.Date
		totalIssues = m.dailySummary.TotalIssues
		totalEstimate = m.dailySummary.TotalEstimatedMinutes
		completedCount = m.dailySummary.CompletedIssues
		abandonedCount = m.dailySummary.AbandonedIssues
		workedSeconds = m.dailySummary.WorkedSeconds
	}

	resolvedCount := completedCount + abandonedCount
	progressText := fmt.Sprintf("%d/%d planned resolved", resolvedCount, totalIssues)
	workedEstimateText := fmt.Sprintf("%s / %dm", formatClock(workedSeconds), totalEstimate)
	lines := []string{
		stylePaneTitle.Render("Daily Dashboard"),
		styleDim.Render(fmt.Sprintf("date: %s   [,] prev   [.] next   [g] today", dateText)),
		"",
		progressText,
		m.renderProgressBar(completedCount, abandonedCount, max(0, totalIssues-resolvedCount), max(20, width-10)),
		"",
		fmt.Sprintf(
			"%s   %s   %s",
			styleHeader.Render(fmt.Sprintf("planned %d", totalIssues)),
			styleHeader.Render(fmt.Sprintf("worked %s", workedEstimateText)),
			renderIssueStatus("done").Render(fmt.Sprintf("done %d", completedCount))+"  "+
				renderIssueStatus("abandoned").Render(fmt.Sprintf("abandoned %d", abandonedCount)),
		),
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 2).
		Width(width - 2).
		Height(max(7, height-2)).
		Render(strings.Join(lines, "\n"))
}

func (m Model) renderDailyIssuesPane(width, height int) string {
	active := m.pane == PaneIssues
	cur := m.cursor[PaneIssues]
	indices := m.filteredIndices(PaneIssues)
	total := len(indices)

	inner := height - 5
	if inner < 1 {
		inner = 1
	}

	lines := []string{
		stylePaneTitle.Render("Planned Tasks"),
		m.renderFilterLine(PaneIssues, width-6),
	}

	if m.dailySummary == nil || len(m.dailySummary.Issues) == 0 || total == 0 {
		lines = append(lines, styleDim.Render("No planned tasks for this date"))
		return m.renderPaneBox(active, width, height, strings.Join(lines, "\n"))
	}

	start, end := listWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, styleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		issue := m.dailySummary.Issues[indices[i]]
		meta := m.issueMetaByID(issue.ID)
		repoName := "-"
		streamName := "-"
		if meta != nil {
			repoName = meta.RepoName
			streamName = meta.StreamName
		}
		estimate := ""
		if issue.EstimateMinutes != nil {
			estimate = fmt.Sprintf("  %dm", *issue.EstimateMinutes)
		}
		row := fmt.Sprintf(
			"[%s] %s%s\n%s/%s",
			plainIssueStatus(issue.Status),
			issue.Title,
			estimate+issueDueSuffix(issue.TodoForDate),
			repoName,
			streamName,
		)
		lines = append(lines, m.renderPaneRowStyled(i, cur, active, row, renderIssueStatus(issue.Status), width))
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, styleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return m.renderPaneBox(active, width, height, strings.Join(lines, "\n"))
}

func (m Model) renderMetaIssuesPane(width, height int, emptyText string) string {
	active := m.pane == PaneIssues
	cur := m.cursor[PaneIssues]
	indices := m.filteredIndices(PaneIssues)
	total := len(indices)

	inner := height - 5
	if inner < 1 {
		inner = 1
	}

	lines := []string{
		stylePaneTitle.Render("Issues [3]"),
		m.renderFilterLine(PaneIssues, width-6),
	}

	if total == 0 {
		lines = append(lines, styleDim.Render(emptyText))
	} else {
		start, end := listWindow(cur, total, inner)
		if start > 0 {
			lines = append(lines, styleDim.Render(fmt.Sprintf("↑ %d more", start)))
		}
		for i := start; i < end; i++ {
			issue := m.issues[indices[i]]
			text := fmt.Sprintf("[%s] %s%s", plainIssueStatus(issue.Status), issue.Title, issueDueSuffix(issue.TodoForDate))
			lines = append(lines, m.renderPaneRowStyled(i, cur, active, text, renderIssueStatus(issue.Status), width))
		}
		if remaining := total - end; remaining > 0 {
			lines = append(lines, styleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
		}
	}

	return m.renderPaneBox(active, width, height, strings.Join(lines, "\n"))
}

func (m Model) renderPaneRow(i, cur int, active bool, text string, width int) string {
	return m.renderPaneRowStyled(i, cur, active, text, nil, width)
}

func (m Model) renderPaneRowStyled(i, cur int, active bool, text string, contentStyle *lipgloss.Style, width int) string {
	line := truncate(text, width-6)
	if contentStyle != nil {
		line = contentStyle.Render(line)
	}
	if i == cur && active {
		return styleCursor.Render("▶ " + line)
	}
	if i == cur {
		return styleSelected.Render("  " + line)
	}
	return styleNormal.Render("  " + line)
}

func (m Model) renderPaneBox(active bool, width, height int, content string) string {
	box := styleInactive
	if active {
		box = styleActive
	}
	return box.
		Width(width-2).
		Height(height-2).
		Padding(0, 1).
		Render(content)
}

func listWindow(cur, total, inner int) (int, int) {
	half := inner / 2
	start := cur - half
	if start < 0 {
		start = 0
	}
	if start+inner > total {
		start = total - inner
	}
	if start < 0 {
		start = 0
	}
	end := start + inner
	if end > total {
		end = total
	}
	return start, end
}

func plainIssueStatus(status string) string {
	switch status {
	case "todo":
		return "todo"
	case "active":
		return "active"
	case "done":
		return "done"
	case "abandoned":
		return "abandoned"
	default:
		return status
	}
}

func issueDueSuffix(todoForDate *string) string {
	label := issueDueLabel(todoForDate)
	if label == "" {
		return ""
	}
	return "  [" + label + "]"
}

func issueDueLabel(todoForDate *string) string {
	if todoForDate == nil {
		return ""
	}
	date := strings.TrimSpace(*todoForDate)
	if date == "" {
		return ""
	}
	if date == time.Now().Format("2006-01-02") {
		return "today"
	}
	return "due " + date
}

func renderIssueStatus(status string) *lipgloss.Style {
	switch status {
	case "done":
		s := lipgloss.NewStyle().Foreground(colorGreen)
		return &s
	case "abandoned":
		s := lipgloss.NewStyle().Foreground(colorRed)
		return &s
	case "active":
		s := lipgloss.NewStyle().Foreground(colorYellow)
		return &s
	default:
		s := lipgloss.NewStyle().Foreground(colorWhite)
		return &s
	}
}

// ---------- Helpers ----------

func deref(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}

func firstNonEmpty(a, b *string) string {
	if a != nil && *a != "" {
		return *a
	}
	return deref(b)
}

func truncate(s string, max int) string {
	if max < 4 {
		max = 4
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

func reverseIndices(in []int) []int {
	out := make([]int, len(in))
	for i := range in {
		out[i] = in[len(in)-1-i]
	}
	return out
}

func formatClock(totalSeconds int) string {
	mm := totalSeconds / 60
	ss := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", mm, ss)
}

func formatEstimateProgress(elapsedSeconds, estimateMinutes int) string {
	return fmt.Sprintf("%s / %dm", formatClock(elapsedSeconds), estimateMinutes)
}

func renderBigClock(clock string) string {
	glyphs := map[rune][]string{
		'0': {"███", "█ █", "█ █", "█ █", "███"},
		'1': {" ██", "██ ", " ██", " ██", "███"},
		'2': {"███", "  █", "███", "█  ", "███"},
		'3': {"███", "  █", "███", "  █", "███"},
		'4': {"█ █", "█ █", "███", "  █", "  █"},
		'5': {"███", "█  ", "███", "  █", "███"},
		'6': {"███", "█  ", "███", "█ █", "███"},
		'7': {"███", "  █", "  █", "  █", "  █"},
		'8': {"███", "█ █", "███", "█ █", "███"},
		'9': {"███", "█ █", "███", "  █", "███"},
		':': {"   ", " █ ", "   ", " █ ", "   "},
	}

	lines := make([]string, 5)
	for _, char := range clock {
		glyph, ok := glyphs[char]
		if !ok {
			continue
		}
		for i := range lines {
			if lines[i] != "" {
				lines[i] += "  "
			}
			lines[i] += glyph[i]
		}
	}
	return strings.Join(lines, "\n")
}

func summarizeCompletedSessions(sessions []api.Session) (workedSeconds int, completedCount int) {
	for _, session := range sessions {
		if session.DurationSeconds == nil || session.EndTime == nil {
			continue
		}
		workedSeconds += *session.DurationSeconds
		completedCount++
	}
	return workedSeconds, completedCount
}

func (m Model) activeIssueWithMeta() *api.IssueWithMeta {
	issueID := ""
	if m.timer != nil && m.timer.IssueID != nil {
		issueID = *m.timer.IssueID
	} else if m.context != nil && m.context.IssueID != nil {
		issueID = *m.context.IssueID
	}
	if issueID == "" {
		return nil
	}
	for i := range m.allIssues {
		if m.allIssues[i].ID == issueID {
			return &m.allIssues[i]
		}
	}
	return nil
}

func (m Model) issueMetaByID(issueID string) *api.IssueWithMeta {
	for i := range m.allIssues {
		if m.allIssues[i].ID == issueID {
			return &m.allIssues[i]
		}
	}
	return nil
}

func (m Model) renderSessionHeaderSummary() string {
	if m.timer == nil || m.timer.State == "idle" {
		return ""
	}

	total := m.timer.ElapsedSeconds + m.elapsed
	stateText := "WORK"
	stateColor := colorGreen
	if m.timer.State == "paused" {
		stateText = "PAUSED"
		stateColor = colorYellow
	}
	if m.timer.SegmentType != nil && *m.timer.SegmentType != "" && *m.timer.SegmentType != "work" {
		stateText = strings.ToUpper(*m.timer.SegmentType)
		stateColor = colorYellow
	}

	parts := []string{
		lipgloss.NewStyle().Foreground(stateColor).Bold(true).Render(stateText),
		styleHeader.Render(formatClock(total)),
	}

	priorWorkedSeconds, completedSessions := summarizeCompletedSessions(m.issueSessions)
	parts = append(parts, styleDim.Render(fmt.Sprintf("sessions:%d", completedSessions)))

	if issue := m.activeIssueWithMeta(); issue != nil && issue.EstimateMinutes != nil {
		parts = append(parts, styleDim.Render(formatEstimateProgress(priorWorkedSeconds+total, *issue.EstimateMinutes)))
	}

	return strings.Join(parts, styleDim.Render("  ·  "))
}

func (m Model) renderHeaderPrimary() string {
	if summary := m.renderSessionHeaderSummary(); summary != "" {
		return summary
	}
	if m.view == ViewDaily {
		return fmt.Sprintf("daily dashboard  %s", m.currentDashboardDate())
	}
	return m.renderHealthChip()
}

func (m Model) renderHeaderSecondary() string {
	parts := []string{}
	if m.timer != nil && m.timer.State != "idle" {
		parts = append(parts, m.renderHealthChip())
		if issue := m.activeIssueWithMeta(); issue != nil {
			parts = append(parts, "status:"+renderIssueStatus(issue.Status).Render(strings.ToUpper(plainIssueStatus(issue.Status))))
		}
	} else if m.view == ViewDaily {
		parts = append(parts, m.renderHealthChip())
	}
	return strings.Join(compactNonEmpty(parts), "  ·  ")
}

func (m Model) renderHeaderSessionLine() string {
	if m.timer == nil || m.timer.State == "idle" {
		return ""
	}
	return truncate(m.renderSessionHeaderSummary()+"  ·  "+m.renderHeaderSecondary(), max(20, m.width-4))
}

func (m Model) renderHealthChip() string {
	if m.health == nil {
		return "kernel: checking"
	}
	if m.health.OK == 1 && m.health.DB {
		return "kernel: ok"
	}
	return "kernel: degraded"
}

func compactNonEmpty(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func (m Model) renderProgressBar(done, abandoned, remaining, width int) string {
	if width < 10 {
		width = 10
	}
	total := done + abandoned + remaining
	if total <= 0 {
		total = 1
	}
	doneW := (done * width) / total
	abandonedW := (abandoned * width) / total
	if doneW < 0 {
		doneW = 0
	}
	if abandonedW < 0 {
		abandonedW = 0
	}
	if doneW+abandonedW > width {
		abandonedW = max(0, width-doneW)
	}
	remainingW := width - doneW - abandonedW
	if remainingW < 0 {
		remainingW = 0
	}
	return lipgloss.NewStyle().Foreground(colorGreen).Render(strings.Repeat("█", doneW)) +
		lipgloss.NewStyle().Foreground(colorRed).Render(strings.Repeat("█", abandonedW)) +
		lipgloss.NewStyle().Foreground(colorDim).Render(strings.Repeat("░", remainingW))
}
