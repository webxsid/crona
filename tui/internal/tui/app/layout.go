package app

import (
	"fmt"
	"strings"

	"crona/tui/internal/tui/app/dialogs"
	helperpkg "crona/tui/internal/tui/app/helpers"
	"crona/tui/internal/tui/app/views"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	minTUIWidth  = 135
	minTUIHeight = 30
)

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}
	if m.isUndersized() {
		return m.renderMinimumSizeWarning()
	}

	base := strings.Join([]string{
		m.renderHeader(),
		m.renderBody(),
		m.renderHelpBar(),
	}, "\n")

	if m.dialog != "" {
		dialogStr := dialogs.Render(dialogTheme(), m.dialogRenderState())
		return clipViewportString(lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			dialogStr,
		), m.width, m.height)
	}

	if m.sessionDetailOpen {
		overlay := m.renderSessionDetailOverlay()
		return clipViewportString(m.renderOverlay(base, overlay, max(0, (m.width-overlayWidth(overlay))/2), max(0, (m.height-overlayHeight(overlay))/2)), m.width, m.height)
	}
	if m.helpOpen {
		overlay := m.renderHelpOverlay()
		return clipViewportString(m.renderOverlay(base, overlay, max(0, (m.width-overlayWidth(overlay))/2), max(0, (m.height-overlayHeight(overlay))/2)), m.width, m.height)
	}
	if m.statusMsg != "" {
		overlay := m.renderStatusToast()
		return clipViewportString(m.renderOverlay(base, overlay, 1, max(0, m.height-overlayHeight(overlay)-1)), m.width, m.height)
	}

	return clipViewportString(base, m.width, m.height)
}

func (m Model) isUndersized() bool {
	return m.width < minTUIWidth || m.height < minTUIHeight
}

func (m Model) renderMinimumSizeWarning() string {
	title := "Terminal Too Small"
	current := fmt.Sprintf("Current: %dx%d", m.width, m.height)
	required := fmt.Sprintf("Required: %dx%d", minTUIWidth, minTUIHeight)
	instruction := "Resize the terminal to continue."

	body := []string{
		stylePaneTitle.Render(title),
		"",
		styleNormal.Render(current),
		styleNormal.Render(required),
		"",
		styleDim.Render(instruction),
	}

	contentWidth := max(
		lipgloss.Width(title),
		max(lipgloss.Width(current), max(lipgloss.Width(required), lipgloss.Width(instruction))),
	)
	boxWidth := min(max(12, contentWidth+8), max(12, m.width-2))
	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorYellow).
		Padding(1, 2).
		Width(boxWidth).
		Render(strings.Join(body, "\n"))

	return clipViewportString(lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	), m.width, m.height)
}

func (m Model) renderHeader() string {
	repo := helperpkg.Deref(nil)
	stream := helperpkg.Deref(nil)
	mode := ""
	if m.context != nil {
		repo = helperpkg.FirstNonEmpty(m.context.RepoName, nil)
		stream = helperpkg.FirstNonEmpty(m.context.StreamName, nil)
	}
	if m.isDevMode() {
		mode = "   " + styleDim.Render("env:") + " " + styleHeader.Render("Dev")
	}
	contextLine := fmt.Sprintf(
		"%s %s   %s %s%s",
		styleDim.Render("repo:"), styleHeader.Render(helperpkg.Truncate(repo, max(16, m.width/4))),
		styleDim.Render("stream:"), styleHeader.Render(helperpkg.Truncate(stream, max(16, m.width/4))),
		mode,
	)

	lines := []string{contextLine}
	if secondary := views.HeaderSessionLine(viewTheme(), views.HeaderState{
		Width:         m.width,
		View:          string(m.view),
		Elapsed:       m.elapsed,
		Timer:         m.timer,
		IssueSessions: m.issueSessions,
		AllIssues:     m.allIssues,
		Health:        m.health,
		UpdateStatus:  m.updateStatus,
	}); secondary != "" {
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
	sidebarWidth, contentWidth := m.bodyWidths()
	sidebar := m.renderSidebar(sidebarWidth, height)
	content := m.renderContent(contentWidth, height)
	return lipgloss.NewStyle().
		Width(m.width).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content))
}

func (m Model) renderSidebar(width, height int) string {
	if m.timer != nil && m.timer.State != "idle" {
		lines := []string{
			stylePaneTitle.Render("Active Session"),
			styleDim.Render("[ / ] switch"),
			"",
			styleDim.Render("SESSION"),
			m.renderSidebarItem(ViewSessionActive, "Session"),
			m.renderSidebarItem(ViewSessionHistory, "History"),
			m.renderSidebarItem(ViewScratch, "Scratchpads"),
		}
		return styleInactive.
			Width(width-4).
			Height(max(3, height-2)).
			Padding(1, 1).
			Render(strings.Join(lines, "\n"))
	}

	lines := []string{
		stylePaneTitle.Render("Views"),
		styleDim.Render("[ / ] switch"),
		"",
		styleDim.Render("DASHBOARD"),
		m.renderSidebarItem(ViewDaily, "Daily"),
		m.renderSidebarItem(ViewWellbeing, "Wellbeing"),
		"",
		styleDim.Render("EXPORT"),
		m.renderSidebarItem(ViewReports, "Reports"),
		"",
		styleDim.Render("WORKSPACE"),
		m.renderSidebarItem(ViewDefault, "Issues"),
		m.renderSidebarItem(ViewMeta, "Meta"),
		m.renderSidebarItem(ViewScratch, "Scratchpads"),
		m.renderSidebarItem(ViewOps, "Ops"),
		m.renderSidebarItem(ViewConfig, "Config"),
		m.renderSidebarItem(ViewSettings, "Settings"),
	}
	lines = append(lines, m.renderSidebarItem(ViewUpdates, "Updates"))
	lines = append(lines,
		"",
		styleDim.Render("SESSION"),
		m.renderSidebarItem(ViewSessionHistory, "History"),
	)

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

func (m Model) renderContent(width, availH int) string {
	return views.RenderContent(viewTheme(), m.viewContentState(width, availH))
}

func (m Model) renderPane(title string, pane Pane, width, height int, emptyText string) string {
	active := m.pane == pane
	cur := m.cursor[pane]
	items := m.paneItems(pane)
	indices := m.filteredIndices(pane)
	total := len(indices)

	inner := height - 5
	if inner < 1 {
		inner = 1
	}

	var lines []string
	lines = append(lines, stylePaneTitle.Render(title))
	actions := []string(nil)
	if active {
		timerState := ""
		if m.timer != nil {
			timerState = m.timer.State
		}
		actions = views.ContextualActions(viewTheme(), views.ActionsState{
			View:                   string(m.view),
			Pane:                   string(pane),
			ScratchpadOpen:         m.scratchpadOpen,
			TimerState:             timerState,
			IsDevMode:              m.isDevMode(),
			UpdateVisible:          viewsShouldShowUpdate(m.updateStatus),
			UpdateInstallAvailable: m.selfUpdateInstallAvailable(),
		})
	}
	lines = append(lines, views.RenderPaneActionLine(viewTheme(), m.filters[pane], width-6, actions))

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
		if remaining := total - end; remaining > 0 {
			lines = append(lines, styleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
		}
	}

	return m.renderPaneBox(active, width, height, strings.Join(lines, "\n"))
}

func (m Model) renderHelpBar() string {
	timerState := ""
	if m.timer != nil {
		timerState = m.timer.State
	}
	actions := views.GlobalActions(viewTheme(), views.ActionsState{
		View:                   string(m.view),
		Pane:                   string(m.pane),
		ScratchpadOpen:         m.scratchpadOpen,
		TimerState:             timerState,
		IsDevMode:              m.isDevMode(),
		UpdateVisible:          viewsShouldShowUpdate(m.updateStatus),
		UpdateInstallAvailable: m.selfUpdateInstallAvailable(),
	})
	devRightAction := ""
	if m.isDevMode() {
		devRightAction = "[f6] seed dev data   [f7] clear dev data   "
	}
	leftActions := actions
	rightText := "[K] stop kernel   [q] quit"
	if m.timer != nil && m.timer.State != "idle" {
		rightText = "session/history/scratchpads only   [K] stop kernel   [q] quit"
	}
	rightText = devRightAction + "[K] stop kernel   [q] quit"
	if m.width < 200 && len(leftActions) > 5 {
		leftActions = leftActions[:5]
		rightText = "[?] more   [q] quit"
	}
	if m.width < 120 && len(leftActions) > 4 {
		leftActions = leftActions[:4]
		rightText = "[?] more   [q] quit"
	}
	if m.width < 96 && len(leftActions) > 3 {
		leftActions = leftActions[:3]
		rightText = "[?] more   [q] quit"
	}
	if m.width < 76 && len(leftActions) > 2 {
		leftActions = leftActions[:2]
		rightText = "[q] quit"
	}

	left := strings.Join(leftActions, "   ")
	right := styleDim.Render(rightText)
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return " " + left + strings.Repeat(" ", gap-1) + right
}

func (m Model) contentHeight() int {
	headerH := 4
	if m.width > 0 {
		headerH = lipgloss.Height(m.renderHeader())
	}
	helpH := 1
	if m.width > 0 {
		helpH = lipgloss.Height(m.renderHelpBar())
	}
	availableHeight := m.height - headerH - helpH
	if availableHeight < 4 {
		availableHeight = 4
	}
	return availableHeight
}

func (m Model) renderOverlay(base, overlay string, x, y int) string {
	baseLines := strings.Split(base, "\n")
	if len(baseLines) < m.height {
		for len(baseLines) < m.height {
			baseLines = append(baseLines, "")
		}
	}
	for i := range baseLines {
		baseLines[i] = padRight(baseLines[i], m.width)
	}

	overlayLines := strings.Split(overlay, "\n")
	for row, line := range overlayLines {
		targetRow := y + row
		if targetRow < 0 || targetRow >= len(baseLines) {
			continue
		}
		baseRunes := []rune(baseLines[targetRow])
		overlayRunes := []rune(line)
		for col, r := range overlayRunes {
			targetCol := x + col
			if targetCol < 0 || targetCol >= len(baseRunes) {
				continue
			}
			baseRunes[targetCol] = r
		}
		baseLines[targetRow] = string(baseRunes)
	}
	return strings.Join(baseLines, "\n")
}

func (m Model) renderStatusToast() string {
	if m.statusMsg == "" {
		return ""
	}
	maxWidth := min(max(28, m.width/2), max(28, m.width-4))
	title := "Notice"
	border := colorCyan
	bodyStyle := styleNormal
	if m.statusErr {
		title = "ERROR"
		border = colorRed
		bodyStyle = styleError.Bold(true)
	}
	return overlayBox(title, wrapText(m.statusMsg, maxWidth-6), nil, maxWidth, border, bodyStyle)
}

func (m Model) renderHelpOverlay() string {
	bodyLines := []string{"Press ? or esc to close", ""}
	for _, action := range m.paneActions() {
		bodyLines = append(bodyLines, action)
	}
	boxWidth := min(max(42, m.width-8), 88)
	return overlayBox("Keys", bodyLines, []string{"[?] close"}, boxWidth, colorCyan, styleNormal)
}

func (m Model) renderSessionDetailOverlay() string {
	body := m.sessionDetailContentLines()
	boxWidth := min(max(52, m.width-10), 96)
	innerWidth := boxWidth - 4
	visibleHeight := m.sessionDetailViewportHeight()
	wrapped := make([]string, 0, len(body))
	for _, line := range body {
		if line == "" {
			wrapped = append(wrapped, "")
			continue
		}
		wrapped = append(wrapped, wrapText(line, innerWidth)...)
	}
	if len(wrapped) == 0 {
		wrapped = []string{"No session details available"}
	}
	maxOffset := max(0, len(wrapped)-visibleHeight)
	offset := m.sessionDetailY
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}
	visible := wrapped[offset:]
	if len(visible) > visibleHeight {
		visible = visible[:visibleHeight]
	}
	if offset > 0 {
		visible = append([]string{"[more above]"}, visible...)
	}
	if offset+visibleHeight < len(wrapped) {
		visible = append(visible, "[more below]")
	}
	footer := []string{"[j/k] scroll   [e] amend   [esc] close"}
	return overlayBox("Session Detail", visible, footer, boxWidth, colorCyan, styleNormal)
}

func overlayBox(title string, body, footer []string, width int, border lipgloss.Color, bodyStyle lipgloss.Style) string {
	innerWidth := width - 6
	if innerWidth < 12 {
		innerWidth = 12
		width = innerWidth + 6
	}

	titleLine := stylePaneTitle.Foreground(border).Render(helperpkg.Truncate(title, innerWidth))
	bodyLines := renderOverlaySection(body, innerWidth, bodyStyle)
	footerLines := renderOverlaySection(footer, innerWidth, styleDim)

	lines := []string{titleLine}
	if len(bodyLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, bodyLines...)
	}
	if len(footerLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, footerLines...)
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(1, 2).
		Width(width).
		Render(strings.Join(lines, "\n"))
}

func renderOverlaySection(lines []string, width int, lineStyle lipgloss.Style) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			out = append(out, "")
			continue
		}
		for _, wrapped := range wrapText(line, width) {
			out = append(out, lineStyle.Render(helperpkg.Truncate(wrapped, width)))
		}
	}
	return out
}

func wrapText(text string, width int) []string {
	if width < 4 {
		return []string{text}
	}
	if strings.TrimSpace(text) == "" {
		return []string{""}
	}
	words := strings.Fields(text)
	lines := make([]string, 0, len(words))
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
			continue
		}
		if len([]rune(current))+1+len([]rune(word)) <= width {
			current += " " + word
			continue
		}
		lines = append(lines, current)
		current = word
	}
	if current != "" {
		lines = append(lines, current)
	}
	if len(lines) == 0 {
		return []string{text}
	}
	return lines
}

func overlayWidth(overlay string) int {
	width := 0
	for _, line := range strings.Split(overlay, "\n") {
		if w := len([]rune(line)); w > width {
			width = w
		}
	}
	return width
}

func overlayHeight(overlay string) int {
	if overlay == "" {
		return 0
	}
	return len(strings.Split(overlay, "\n"))
}

func padRight(s string, width int) string {
	if width < 1 {
		return ""
	}
	if ansi.StringWidth(s) >= width {
		return ansi.Truncate(s, width, "")
	}
	return s + strings.Repeat(" ", width-ansi.StringWidth(s))
}

func clipViewportString(s string, width, height int) string {
	if height < 1 || width < 1 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	for i := range lines {
		lines[i] = padRight(lines[i], width)
	}
	return strings.Join(lines, "\n")
}

func (m Model) sidebarWidth() int {
	if m.width < 64 {
		return max(14, m.width/4)
	}
	if m.width < 90 {
		return 18
	}
	return 22
}

func (m Model) mainContentWidth() int {
	_, contentWidth := m.bodyWidths()
	return contentWidth
}

func (m Model) bodyWidths() (int, int) {
	sidebarWidth := m.sidebarWidth()
	contentWidth := m.width - sidebarWidth
	if contentWidth < 24 {
		contentWidth = 24
		sidebarWidth = max(14, m.width-contentWidth)
		contentWidth = m.width - sidebarWidth
	}
	if contentWidth < 0 {
		contentWidth = 0
	}
	return sidebarWidth, contentWidth
}

func (m Model) renderPaneRow(i, cur int, active bool, text string, width int) string {
	return m.renderPaneRowStyled(i, cur, active, text, nil, width)
}

func (m Model) renderPaneRowStyled(i, cur int, active bool, text string, contentStyle *lipgloss.Style, width int) string {
	line := helperpkg.Truncate(text, width-6)
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
		Render(viewsClipBoxContent(content, height-2))
}

func viewsClipBoxContent(content string, maxLines int) string {
	if maxLines < 1 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}
	if maxLines == 1 {
		return "..."
	}
	clipped := append([]string{}, lines[:maxLines-1]...)
	clipped = append(clipped, "...")
	return strings.Join(clipped, "\n")
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

func (m Model) isDevMode() bool {
	return m.kernelInfo != nil && m.kernelInfo.Env == "Dev"
}

func viewTheme() views.Theme {
	return views.Theme{
		ColorBlue: colorBlue, ColorCyan: colorCyan, ColorGreen: colorGreen, ColorMagenta: colorMagenta,
		ColorSubtle: colorSubtle, ColorYellow: colorYellow, ColorRed: colorRed, ColorDim: colorDim, ColorWhite: colorWhite,
		StyleActive: styleActive, StyleInactive: styleInactive, StylePaneTitle: stylePaneTitle, StyleDim: styleDim,
		StyleCursor: styleCursor, StyleHeader: styleHeader, StyleError: styleError, StyleSelected: styleSelected, StyleNormal: styleNormal,
	}
}

func dialogTheme() dialogs.Theme {
	return dialogs.Theme{
		ColorCyan: colorCyan, ColorYellow: colorYellow, ColorRed: colorRed, ColorGreen: colorGreen,
		StylePaneTitle: stylePaneTitle, StyleDim: styleDim, StyleCursor: styleCursor, StyleHeader: styleHeader, StyleError: styleError, StyleSelected: styleSelected, StyleNormal: styleNormal,
	}
}

func (m Model) viewContentState(width, height int) views.ContentState {
	state := views.ContentState{
		View: string(m.view), Pane: string(m.pane), Width: width, Height: height, Elapsed: m.elapsed, DashboardDate: m.dashboardDate, WellbeingDate: m.currentWellbeingDate(), DefaultIssueSection: string(m.defaultIssueSection), SessionHistoryTitle: m.sessionHistoryTitle(), SessionHistoryMeta: m.sessionHistorySubtitle(),
		Cursors:            map[string]int{"repos": m.cursor[PaneRepos], "streams": m.cursor[PaneStreams], "issues": m.cursor[PaneIssues], "habits": m.cursor[PaneHabits], "sessions": m.cursor[PaneSessions], "scratchpads": m.cursor[PaneScratchpads], "ops": m.cursor[PaneOps], "export_reports": m.cursor[PaneExportReports], "config": m.cursor[PaneConfig], "settings": m.cursor[PaneSettings]},
		Filters:            map[string]string{"repos": m.filters[PaneRepos], "streams": m.filters[PaneStreams], "issues": m.filters[PaneIssues], "habits": m.filters[PaneHabits], "sessions": m.filters[PaneSessions], "scratchpads": m.filters[PaneScratchpads], "ops": m.filters[PaneOps], "export_reports": m.filters[PaneExportReports], "config": m.filters[PaneConfig], "settings": m.filters[PaneSettings]},
		ScratchpadOpen:     m.scratchpadOpen,
		ScratchpadRendered: m.scratchpadViewport.View(),
		Repos:              m.repos, Streams: m.streams, Issues: m.issues, DailyIssues: m.dailyScopedIssues(), Habits: m.habits, AllIssues: m.allIssues, DefaultIssues: m.defaultScopedIssues(), DueHabits: m.filteredDueHabits(), DailySummary: m.dailySummary, DailyCheckIn: m.dailyCheckIn, MetricsRange: m.metricsRange, MetricsRollup: m.metricsRollup, Streaks: m.streaks, ExportAssets: m.exportAssets, ExportReports: m.exportReports, IssueSessions: m.issueSessions, SessionHistory: m.sessionHistory, Scratchpads: m.scratchpads, Ops: m.ops, Context: m.context, Timer: m.timer, Health: m.health, UpdateStatus: m.updateStatus, UpdateChecking: m.updateChecking, UpdateInstalling: m.updateInstalling, UpdateInstallOutput: m.updateInstallOutput, UpdateInstallError: m.updateInstallError, UpdateInstallAvailable: m.selfUpdateInstallAvailable(), UpdateManualReason: m.selfUpdateUnsupportedReason(), TUIExecutablePath: m.currentExecutablePath, KernelExecutablePath: kernelExecutablePath(m.kernelInfo), Settings: m.settings,
	}
	if m.scratchpadMeta != nil {
		state.ScratchpadName = m.scratchpadMeta.Name
		state.ScratchpadPath = m.scratchpadMeta.Path
	}
	return state
}

func (m Model) dialogRenderState() dialogs.State {
	state := m.dialogState()
	state.Width = m.width
	if m.dialog == "create_issue_default" || m.dialog == "create_habit" {
		state.RepoSelectorLabel, state.StreamSelectorLabel = dialogs.DefaultIssueDialogLabels(m.dialogInputs, m.dialogRepoIndex, m.dialogStreamIndex, m.repos, m.allIssues, m.streams, m.context)
	}
	if m.dialog == "checkout_context" {
		state.RepoSelectorLabel, state.StreamSelectorLabel = dialogs.CheckoutDialogLabels(m.dialogInputs, m.dialogRepoIndex, m.dialogStreamIndex, m.repos, m.allIssues, m.streams, m.context)
	}
	if m.dialog == "pick_date" {
		state = dialogs.PopulateDatePresentation(dialogTheme(), state, m.currentDashboardDate())
	}
	for _, stash := range m.stashes {
		label := stash.CreatedAt
		if stash.Note != nil && strings.TrimSpace(*stash.Note) != "" {
			label = *stash.Note
		}
		contextBits := []string{}
		if stash.RepoID != nil {
			contextBits = append(contextBits, fmt.Sprintf("repo:%d", *stash.RepoID))
		}
		if stash.StreamID != nil {
			contextBits = append(contextBits, fmt.Sprintf("stream:%d", *stash.StreamID))
		}
		if stash.IssueID != nil {
			contextBits = append(contextBits, fmt.Sprintf("issue:%d", *stash.IssueID))
		}
		meta := stash.CreatedAt
		if len(contextBits) > 0 {
			meta += "  " + strings.Join(contextBits, "  ")
		}
		state.Stashes = append(state.Stashes, dialogs.StashItem{Label: helperpkg.Truncate(label, 42), Meta: helperpkg.Truncate(meta, 48)})
	}
	return state
}
