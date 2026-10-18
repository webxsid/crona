package layout

import (
	"fmt"
	"slices"
	"strings"

	"crona/tui/internal/logger"
	"crona/tui/internal/tui/chrome"
	"crona/tui/internal/tui/dialogs"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
	helperpkg "crona/tui/internal/tui/helpers"
	uistate "crona/tui/internal/tui/state"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewrenderer "crona/tui/internal/tui/views/renderer"
	viewtypes "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	MinWidth  = 60
	MinHeight = 20
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
	HeaderState         viewchrome.HeaderState
	ContentState        viewtypes.ContentState
	DialogOpen          bool
	DialogState         dialogstate.State
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
	Tabs                []uistate.Tab
	SubTabs             []uistate.View
}

func ViewTheme() viewtypes.Theme {
	return viewtypes.Theme{
		ColorBlue: chrome.ColorBlue, ColorCyan: chrome.ColorCyan, ColorGreen: chrome.ColorGreen, ColorDullGreen: chrome.ColorDullGreen, ColorMagenta: chrome.ColorMagenta,
		ColorSubtle: chrome.ColorSubtle, ColorYellow: chrome.ColorYellow, ColorRed: chrome.ColorRed, ColorDullRed: chrome.ColorDullRed, ColorOrange: chrome.ColorOrange, ColorDim: chrome.ColorDim, ColorWhite: chrome.ColorWhite,
		StyleActive: chrome.StyleActive, StyleInactive: chrome.StyleInactive, StylePaneTitle: chrome.StylePaneTitle, StyleDim: chrome.StyleDim,
		StyleCursor: chrome.StyleCursor, StyleHeader: chrome.StyleHeader, StyleError: chrome.StyleError, StyleSelected: chrome.StyleSelected, StyleSelectedInverse: chrome.StyleSelectedInverse, StyleNormal: chrome.StyleNormal,
	}
}

func DialogTheme() dialogs.Theme {
	return dialogs.Theme{
		ColorCyan: chrome.ColorCyan, ColorYellow: chrome.ColorYellow, ColorOrange: chrome.ColorOrange, ColorRed: chrome.ColorRed, ColorGreen: chrome.ColorGreen, ColorDim: chrome.ColorDim,
		StylePaneTitle: chrome.StylePaneTitle, StyleDim: chrome.StyleDim, StyleCursor: chrome.StyleCursor, StyleHeader: chrome.StyleHeader, StyleError: chrome.StyleError, StyleSelected: chrome.StyleSelected, StyleNormal: chrome.StyleNormal,
	}
}

func Render(state State) string {
	if state.Width == 0 {
		logger.Infof(
			"layout.Render loading: width=0 height=%d view=%q pane=%q dialog_open=%t dialog_kind=%q",
			state.Height,
			state.View,
			state.Pane,
			state.DialogOpen,
			state.DialogState.Kind,
		)
		return "Loading..."
	}
	if state.Width < MinWidth || state.Height < MinHeight {
		logger.Infof(
			"layout.Render minimum-size: width=%d height=%d view=%q pane=%q dialog_open=%t dialog_kind=%q",
			state.Width,
			state.Height,
			state.View,
			state.Pane,
			state.DialogOpen,
			state.DialogState.Kind,
		)
		return renderMinimumSizeWarning(state.Width, state.Height)
	}

	if state.DialogOpen {
		dialogStr := dialogs.Render(DialogTheme(), state.DialogState)
		logger.Infof(
			"layout.Render dialog overlay: view=%q pane=%q dialog_kind=%q dialog_len=%d width=%d height=%d",
			state.View,
			state.Pane,
			state.DialogState.Kind,
			len(dialogStr),
			state.Width,
			state.Height,
		)
		return placeCentered(state, dialogStr)
	}
	if state.SessionDetailOpen {
		return placeCentered(state, renderSessionDetailOverlay(state))
	}
	if state.SessionContextOpen {
		return placeCentered(state, renderSessionContextOverlay(state))
	}

	parts := []string{renderTabLine(state)}
	if subTabs := renderSubTabLine(state); subTabs != "" {
		parts = append(parts, subTabs)
	}
	if dateRow := renderDateRow(state); dateRow != "" {
		parts = append(parts, dateRow)
	}
	content := clipViewportString(
		viewrenderer.RenderContent(ViewTheme(), state.ContentState),
		state.Width,
		ContentHeight(state),
	)
	parts = append(parts, content, renderStatusLine(state))
	result := clipViewportString(strings.Join(parts, "\n"), state.Width, state.Height)
	logger.Infof(
		"layout.Render base: view=%q pane=%q result_len=%d width=%d height=%d status=%q",
		state.View,
		state.Pane,
		len(result),
		state.Width,
		state.Height,
		state.StatusMsg,
	)
	return result
}

func placeCentered(state State, overlay string) string {
	return clipViewportString(
		lipgloss.Place(state.Width, state.Height, lipgloss.Center, lipgloss.Center, overlay),
		state.Width,
		state.Height,
	)
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
	contentWidth := max(
		lipgloss.Width(title),
		max(lipgloss.Width(current), max(lipgloss.Width(required), lipgloss.Width(instruction))),
	)
	boxWidth := min(max(12, contentWidth+8), max(12, width-2))
	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(chrome.ColorYellow).
		Padding(1, 2).
		Width(boxWidth).
		Render(strings.Join(body, "\n"))
	return clipViewportString(
		lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box),
		width,
		height,
	)
}

func renderTabLine(state State) string {
	parts := []string{}
	for i, tab := range uistate.TabOrder() {
		if !slices.Contains(state.Tabs, tab) {
			continue
		}
		label := fmt.Sprintf("[%d] %s", i+1, uistate.TabLabel(tab))
		if tab == uistate.TabForView(state.View) {
			parts = append(parts, chrome.StyleCursor.Render(label))
		} else {
			parts = append(parts, chrome.StyleDim.Render(label))
		}
	}
	left := " " + strings.Join(parts, "  ")
	right := contextLabel(state.RepoName, state.StreamName)
	if state.IsDevMode {
		right += " · dev"
	}
	// The tabs are the only navigation indicator, so the context gives way first.
	room := state.Width - ansi.StringWidth(left) - 2
	if room < 4 {
		return joinEnds(left, "", state.Width)
	}
	right = ansi.Truncate(right, room, "…")
	return joinEnds(left, chrome.StyleHeader.Render(right), state.Width)
}

// contextLabel names the active scope, falling back to All when no repo is selected.
func contextLabel(repo, stream string) string {
	switch {
	case repo == "":
		return "All"
	case stream == "":
		return repo
	default:
		return repo + " › " + stream
	}
}

// renderSubTabLine lists the tab's views between the [ and ] keys that cycle them.
func renderSubTabLine(state State) string {
	if len(state.SubTabs) < 2 {
		return ""
	}
	parts := make([]string, 0, len(state.SubTabs))
	for _, view := range state.SubTabs {
		label := uistate.ViewLabel(view)
		if view == state.View {
			parts = append(parts, chrome.StyleHeader.Render(label))
		} else {
			parts = append(parts, chrome.StyleDim.Render(label))
		}
	}
	key := func(k string) string { return chrome.StyleHeader.Render("[" + k + "]") }
	return " " + key("[") + " " + strings.Join(parts, chrome.StyleDim.Render(" · ")) + " " + key("]")
}

// renderDateRow is a dated view's header row: the shared date's navigation on the left and
// the view's own date keys on the right, which give way first when narrow.
func renderDateRow(state State) string {
	if !viewrenderer.HasDateRow(state.ContentState) {
		return ""
	}
	theme := ViewTheme()
	full := viewrenderer.DateNav(theme, state.ContentState, state.Width)
	keys := viewchrome.FitHints(viewrenderer.DateRowKeys(theme, state.ContentState), state.Width-ansi.StringWidth(full)-6)
	nav := viewrenderer.DateNav(theme, state.ContentState, state.Width-ansi.StringWidth(keys)-4)
	if keys == "" {
		return " " + nav
	}
	return joinEnds(" "+nav, keys+" ", state.Width)
}

func renderStatusLine(state State) string {
	// The focused section's keys lead; global keys fill whatever room is left.
	actions := append(append([]string{}, state.PaneActions...), state.GlobalActions...)
	left := strings.Join(actions, "   ")
	if state.TimerActive {
		left = viewchrome.HeaderSessionLine(ViewTheme(), state.HeaderState)
	}
	right := chrome.StyleHeader.Render("[?]") + chrome.StyleDim.Render(" all keys  ctrl+c quit")
	if state.StatusMsg != "" {
		style := chrome.StyleSelected
		if state.StatusErr {
			style = chrome.StyleError
		}
		msg := strings.Join(strings.Fields(state.StatusMsg), " ")
		right = style.Render(ansi.Truncate(msg, max(1, state.Width-2), "…"))
	}
	return joinEnds(" "+left, right, state.Width)
}

// joinEnds places left and right on one line of exactly width cells, truncating left first.
func joinEnds(left, right string, width int) string {
	rightWidth := ansi.StringWidth(right)
	if rightWidth >= width {
		return ansi.Truncate(right, width, "")
	}
	room := width - rightWidth - 1
	if ansi.StringWidth(left) > room {
		left = ansi.Truncate(left, room, "…")
	}
	gap := width - ansi.StringWidth(left) - rightWidth
	return left + strings.Repeat(" ", gap) + right
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

	offset = max(0, min(offset, maxOffset))

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
	footer := []string{"[↑/↓] scroll   [esc] close"}
	if state.ContentState.Settings != nil && state.ContentState.Settings.ManageHistory {
		footer = []string{"[↑/↓] scroll   [e] edit   [D] delete   [esc] close"}
	}
	return overlayBox(
		"Session Detail",
		visible,
		footer,
		boxWidth,
		chrome.ColorCyan,
		chrome.StyleNormal,
	)
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

	offset = max(0, min(offset, maxOffset))

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
	return overlayBox(
		"Issue Context",
		visible,
		[]string{"[↑/↓] scroll   [esc] close"},
		boxWidth,
		chrome.ColorCyan,
		chrome.StyleNormal,
	)
}

func overlayBox(
	title string,
	body, footer []string,
	width int,
	border lipgloss.Color,
	bodyStyle lipgloss.Style,
) string {
	innerWidth := width - 6
	if innerWidth < 12 {
		innerWidth = 12
		width = innerWidth + 6
	}
	lines := []string{
		chrome.StylePaneTitle.Foreground(border).Render(helperpkg.Truncate(title, innerWidth)),
	}
	if bodyLines := renderOverlaySection(body, innerWidth, bodyStyle); len(bodyLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, bodyLines...)
	}
	if footerLines := renderOverlaySection(footer, innerWidth, chrome.StyleDim); len(
		footerLines,
	) > 0 {
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
	rows := 2
	if len(state.SubTabs) >= 2 {
		rows++
	}
	// The shell's view, not ContentState's: the model sizes content before filling it in.
	if viewrenderer.HasDateRow(viewtypes.ContentState{View: string(state.View)}) {
		rows++
	}
	return max(4, state.Height-rows)
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
