package layout

import (
	"fmt"
	"strings"

	"crona/tui/internal/tui/chrome"
	"crona/tui/internal/tui/dialogs"
	helperpkg "crona/tui/internal/tui/helpers"
	uistate "crona/tui/internal/tui/state"
	"crona/tui/internal/tui/views"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	MinWidth  = 135
	MinHeight = 30
)

type State struct {
	Width               int
	Height              int
	View                uistate.View
	Pane                uistate.Pane
	ProtectedMode       bool
	IsDevMode           bool
	RepoName            string
	StreamName          string
	TimerActive         bool
	HeaderState         views.HeaderState
	ContentState        views.ContentState
	DialogOpen          bool
	DialogState         dialogs.State
	HelpOpen            bool
	SessionDetailOpen   bool
	SessionDetailY      int
	SessionDetailLines  []string
	SessionContextOpen  bool
	SessionContextY     int
	SessionContextLines []string
	StatusMsg           string
	StatusErr           bool
	PaneActions         []string
	GlobalActions       []string
}

func ViewTheme() views.Theme {
	return views.Theme{
		ColorBlue: chrome.ColorBlue, ColorCyan: chrome.ColorCyan, ColorGreen: chrome.ColorGreen, ColorMagenta: chrome.ColorMagenta,
		ColorSubtle: chrome.ColorSubtle, ColorYellow: chrome.ColorYellow, ColorRed: chrome.ColorRed, ColorDim: chrome.ColorDim, ColorWhite: chrome.ColorWhite,
		StyleActive: chrome.StyleActive, StyleInactive: chrome.StyleInactive, StylePaneTitle: chrome.StylePaneTitle, StyleDim: chrome.StyleDim,
		StyleCursor: chrome.StyleCursor, StyleHeader: chrome.StyleHeader, StyleError: chrome.StyleError, StyleSelected: chrome.StyleSelected, StyleNormal: chrome.StyleNormal,
	}
}

func DialogTheme() dialogs.Theme {
	return dialogs.Theme{
		ColorCyan: chrome.ColorCyan, ColorYellow: chrome.ColorYellow, ColorRed: chrome.ColorRed, ColorGreen: chrome.ColorGreen,
		StylePaneTitle: chrome.StylePaneTitle, StyleDim: chrome.StyleDim, StyleCursor: chrome.StyleCursor, StyleHeader: chrome.StyleHeader, StyleError: chrome.StyleError, StyleSelected: chrome.StyleSelected, StyleNormal: chrome.StyleNormal,
	}
}

func Render(state State) string {
	if state.Width == 0 {
		return "Loading..."
	}
	if state.Width < MinWidth || state.Height < MinHeight {
		return renderMinimumSizeWarning(state.Width, state.Height)
	}

	header := ""
	help := ""
	if !state.ProtectedMode {
		header = renderHeader(state)
		help = renderHelpBar(state)
	}
	contentHeight := contentHeightFromChrome(state.ProtectedMode, state.Height, header, help)
	baseParts := []string{}
	if header != "" {
		baseParts = append(baseParts, header)
	}
	baseParts = append(baseParts, renderBody(state, contentHeight))
	if help != "" {
		baseParts = append(baseParts, help)
	}
	base := strings.Join(baseParts, "\n")

	if state.DialogOpen {
		dialogStr := dialogs.Render(DialogTheme(), state.DialogState)
		return clipViewportString(lipgloss.Place(state.Width, state.Height, lipgloss.Center, lipgloss.Center, dialogStr), state.Width, state.Height)
	}
	if state.SessionDetailOpen {
		dialogStr := renderSessionDetailOverlay(state)
		return clipViewportString(lipgloss.Place(state.Width, state.Height, lipgloss.Center, lipgloss.Center, dialogStr), state.Width, state.Height)
	}
	if state.SessionContextOpen {
		dialogStr := renderSessionContextOverlay(state)
		return clipViewportString(lipgloss.Place(state.Width, state.Height, lipgloss.Center, lipgloss.Center, dialogStr), state.Width, state.Height)
	}
	if state.HelpOpen {
		overlay := renderHelpOverlay(state)
		return clipViewportString(renderOverlay(base, overlay, max(0, (state.Width-overlayWidth(overlay))/2), max(0, (state.Height-overlayHeight(overlay))/2), state.Width, state.Height), state.Width, state.Height)
	}
	if state.StatusMsg != "" {
		overlay := renderStatusToast(state)
		return clipViewportString(renderOverlay(base, overlay, 1, max(0, state.Height-overlayHeight(overlay)-1), state.Width, state.Height), state.Width, state.Height)
	}
	return clipViewportString(base, state.Width, state.Height)
}

func renderMinimumSizeWarning(width, height int) string {
	title := "Terminal Too Small"
	current := fmt.Sprintf("Current: %dx%d", width, height)
	required := fmt.Sprintf("Required: %dx%d", MinWidth, MinHeight)
	instruction := "Resize the terminal to continue."
	body := []string{
		chrome.StylePaneTitle.Render(title),
		"",
		chrome.StyleNormal.Render(current),
		chrome.StyleNormal.Render(required),
		"",
		chrome.StyleDim.Render(instruction),
	}
	contentWidth := max(lipgloss.Width(title), max(lipgloss.Width(current), max(lipgloss.Width(required), lipgloss.Width(instruction))))
	boxWidth := min(max(12, contentWidth+8), max(12, width-2))
	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(chrome.ColorYellow).
		Padding(1, 2).
		Width(boxWidth).
		Render(strings.Join(body, "\n"))
	return clipViewportString(lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box), width, height)
}

func renderHeader(state State) string {
	mode := ""
	if state.IsDevMode {
		mode = "   " + chrome.StyleDim.Render("env:") + " " + chrome.StyleHeader.Render("Dev")
	}
	contextLine := fmt.Sprintf("%s %s   %s %s%s",
		chrome.StyleDim.Render("repo:"), chrome.StyleHeader.Render(helperpkg.Truncate(state.RepoName, max(16, state.Width/4))),
		chrome.StyleDim.Render("stream:"), chrome.StyleHeader.Render(helperpkg.Truncate(state.StreamName, max(16, state.Width/4))),
		mode,
	)
	lines := []string{contextLine}
	if secondary := views.HeaderSessionLine(ViewTheme(), state.HeaderState); secondary != "" {
		lines = append(lines, secondary)
	}
	return lipgloss.NewStyle().
		Width(state.Width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(chrome.ColorDim).
		Render(strings.Join(lines, "\n"))
}

func renderBody(state State, height int) string {
	sidebarWidth, _ := bodyWidths(state.Width)
	sidebar := renderSidebar(state, sidebarWidth, height)
	content := views.RenderContent(ViewTheme(), state.ContentState)
	return lipgloss.NewStyle().
		Width(state.Width).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content))
}

func renderSidebar(state State, width, height int) string {
	var lines []string
	if state.ProtectedMode {
		lines = []string{
			chrome.StylePaneTitle.Render("Away"),
			"",
			chrome.StyleDim.Render("AVAILABLE"),
			renderSidebarItem(state, uistate.ViewAway, "Away"),
			renderSidebarItem(state, uistate.ViewReports, "Reports"),
			renderSidebarItem(state, uistate.ViewSessionHistory, "History"),
		}
		return chrome.StyleInactive.Width(width-4).Height(max(3, height-2)).Padding(1, 1).Render(strings.Join(lines, "\n"))
	}
	if state.TimerActive {
		lines = []string{
			chrome.StylePaneTitle.Render("Active Session"),
			chrome.StyleDim.Render("[ / ] switch"),
			chrome.StyleDim.Render("[v] jump"),
			"",
			chrome.StyleDim.Render("SESSION"),
			renderSidebarItem(state, uistate.ViewSessionActive, "Session"),
			renderSidebarItem(state, uistate.ViewSessionHistory, "History"),
			renderSidebarItem(state, uistate.ViewScratch, "Scratchpads"),
		}
	} else {
		lines = []string{
			chrome.StylePaneTitle.Render("Views"),
			chrome.StyleDim.Render("[ / ] switch"),
			chrome.StyleDim.Render("[v] jump"),
			"",
			chrome.StyleDim.Render("DASHBOARD"),
			renderSidebarItem(state, uistate.ViewDaily, "Daily"),
			renderSidebarItem(state, uistate.ViewRollup, "Rollup"),
			renderSidebarItem(state, uistate.ViewWellbeing, "Wellbeing"),
			"",
			chrome.StyleDim.Render("EXPORT"),
			renderSidebarItem(state, uistate.ViewReports, "Reports"),
			renderSidebarItem(state, uistate.ViewConfig, "Config"),
			"",
			chrome.StyleDim.Render("WORKSPACE"),
			renderSidebarItem(state, uistate.ViewDefault, "Issues"),
			renderSidebarItem(state, uistate.ViewMeta, "Meta"),
			renderSidebarItem(state, uistate.ViewScratch, "Scratchpads"),
			renderSidebarItem(state, uistate.ViewOps, "Ops"),
			renderSidebarItem(state, uistate.ViewSettings, "Settings"),
			renderSidebarItem(state, uistate.ViewUpdates, "Updates"),
			renderSidebarItem(state, uistate.ViewSupport, "Support"),
			"",
			chrome.StyleDim.Render("SESSION"),
			renderSidebarItem(state, uistate.ViewSessionHistory, "History"),
		}
	}
	return chrome.StyleInactive.Width(width-4).Height(max(3, height-2)).Padding(1, 1).Render(strings.Join(lines, "\n"))
}

func renderSidebarItem(state State, view uistate.View, label string) string {
	if state.View == view {
		return chrome.StyleCursor.Render("▶ " + label)
	}
	return chrome.StyleNormal.Render("  " + label)
}

func renderHelpBar(state State) string {
	devRightAction := ""
	if state.IsDevMode {
		devRightAction = "[f6] seed dev data   [f7] clear dev data   [f8] local update   "
	}
	leftActions := append([]string(nil), state.GlobalActions...)
	rightText := devRightAction + "[K] stop kernel   [q] quit"
	if state.Width < 200 && len(leftActions) > 5 {
		leftActions = leftActions[:5]
		rightText = "[?] more   [q] quit"
	}
	if state.Width < 120 && len(leftActions) > 4 {
		leftActions = leftActions[:4]
		rightText = "[?] more   [q] quit"
	}
	if state.Width < 96 && len(leftActions) > 3 {
		leftActions = leftActions[:3]
		rightText = "[?] more   [q] quit"
	}
	if state.Width < 76 && len(leftActions) > 2 {
		leftActions = leftActions[:2]
		rightText = "[q] quit"
	}
	left := strings.Join(leftActions, "   ")
	right := chrome.StyleDim.Render(rightText)
	gap := state.Width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return " " + left + strings.Repeat(" ", gap-1) + right
}

func renderOverlay(base, overlay string, x, y, width, height int) string {
	baseLines := strings.Split(base, "\n")
	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}
	for i := range baseLines {
		baseLines[i] = padRight(baseLines[i], width)
	}
	for row, line := range strings.Split(overlay, "\n") {
		targetRow := y + row
		if targetRow < 0 || targetRow >= len(baseLines) {
			continue
		}
		baseRunes := []rune(baseLines[targetRow])
		for col, r := range []rune(line) {
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

func renderStatusToast(state State) string {
	maxWidth := min(max(28, state.Width/2), max(28, state.Width-4))
	title := "Notice"
	border := chrome.ColorCyan
	bodyStyle := chrome.StyleNormal
	if state.StatusErr {
		title = "ERROR"
		border = chrome.ColorRed
		bodyStyle = chrome.StyleError.Bold(true)
	}
	return overlayBox(title, WrapText(state.StatusMsg, maxWidth-6), nil, maxWidth, border, bodyStyle)
}

func renderHelpOverlay(state State) string {
	bodyLines := []string{"Press ? or esc to close", ""}
	bodyLines = append(bodyLines, state.PaneActions...)
	boxWidth := min(max(42, state.Width-8), 88)
	return overlayBox("Keys", bodyLines, []string{"[?] close"}, boxWidth, chrome.ColorCyan, chrome.StyleNormal)
}

func renderSessionDetailOverlay(state State) string {
	boxWidth := min(max(52, state.Width-10), 96)
	innerWidth := boxWidth - 4
	visibleHeight := max(6, state.Height-10)
	wrapped := make([]string, 0, len(state.SessionDetailLines))
	for _, line := range state.SessionDetailLines {
		if line == "" {
			wrapped = append(wrapped, "")
			continue
		}
		wrapped = append(wrapped, WrapText(line, innerWidth)...)
	}
	if len(wrapped) == 0 {
		wrapped = []string{"No session details available"}
	}
	maxOffset := max(0, len(wrapped)-visibleHeight)
	offset := state.SessionDetailY
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
	return overlayBox("Session Detail", visible, []string{"[j/k] scroll   [e] amend   [esc] close"}, boxWidth, chrome.ColorCyan, chrome.StyleNormal)
}

func renderSessionContextOverlay(state State) string {
	boxWidth := min(max(50, state.Width-10), 92)
	innerWidth := boxWidth - 4
	visibleHeight := max(6, state.Height-10)
	wrapped := make([]string, 0, len(state.SessionContextLines))
	for _, line := range state.SessionContextLines {
		if line == "" {
			wrapped = append(wrapped, "")
			continue
		}
		wrapped = append(wrapped, WrapText(line, innerWidth)...)
	}
	if len(wrapped) == 0 {
		wrapped = []string{"No issue context available"}
	}
	maxOffset := max(0, len(wrapped)-visibleHeight)
	offset := state.SessionContextY
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
	return overlayBox("Issue Context", visible, []string{"[j/k] scroll   [esc] close"}, boxWidth, chrome.ColorCyan, chrome.StyleNormal)
}

func overlayBox(title string, body, footer []string, width int, border lipgloss.Color, bodyStyle lipgloss.Style) string {
	innerWidth := width - 6
	if innerWidth < 12 {
		innerWidth = 12
		width = innerWidth + 6
	}
	lines := []string{chrome.StylePaneTitle.Foreground(border).Render(helperpkg.Truncate(title, innerWidth))}
	if bodyLines := renderOverlaySection(body, innerWidth, bodyStyle); len(bodyLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, bodyLines...)
	}
	if footerLines := renderOverlaySection(footer, innerWidth, chrome.StyleDim); len(footerLines) > 0 {
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
		for _, wrapped := range WrapText(line, width) {
			out = append(out, lineStyle.Render(helperpkg.Truncate(wrapped, width)))
		}
	}
	return out
}

func ContentHeight(state State) int {
	if state.ProtectedMode {
		return contentHeightFromChrome(true, state.Height, "", "")
	}
	return contentHeightFromChrome(false, state.Height, renderHeader(state), renderHelpBar(state))
}

func contentHeightFromChrome(protected bool, height int, header, help string) int {
	if protected {
		if height < 4 {
			return 4
		}
		return height
	}
	headerH := 4
	if header != "" {
		headerH = lipgloss.Height(header)
	}
	helpH := 1
	if help != "" {
		helpH = lipgloss.Height(help)
	}
	availableHeight := height - headerH - helpH
	if availableHeight < 4 {
		return 4
	}
	return availableHeight
}

func bodyWidths(width int) (int, int) {
	sidebarWidth := 22
	if width < 64 {
		sidebarWidth = max(14, width/4)
	} else if width < 90 {
		sidebarWidth = 18
	}
	contentWidth := width - sidebarWidth
	if contentWidth < 24 {
		contentWidth = 24
		sidebarWidth = max(14, width-contentWidth)
		contentWidth = width - sidebarWidth
	}
	if contentWidth < 0 {
		contentWidth = 0
	}
	return sidebarWidth, contentWidth
}

func WrapText(text string, width int) []string {
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
