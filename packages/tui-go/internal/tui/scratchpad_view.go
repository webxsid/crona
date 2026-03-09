package tui

import (
	"strings"

	"crona/tui/internal/api"
	"crona/tui/internal/logger"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

func (m Model) scratchpadPaneSize() (int, int) {
	availH := m.contentHeight()
	if availH < 4 {
		availH = 4
	}

	return m.mainContentWidth(), availH
}

func (m *Model) syncScratchpadViewport() {
	paneW, paneH := m.scratchpadPaneSize()
	contentW := max(20, paneW-6)
	contentH := paneH - 7
	if contentH < 1 {
		contentH = 1
	}

	m.scratchpadViewport.Width = contentW
	m.scratchpadViewport.Height = contentH
	if m.scratchpadRendered != "" {
		m.scratchpadViewport.SetContent(m.scratchpadRendered)
	}
}

func (m *Model) setActiveScratchpadByIndex(idx int) {
	if idx < 0 || idx >= len(m.scratchpads) {
		return
	}
	pad := m.scratchpads[idx]
	m.scratchpadMeta = &api.ScratchPad{
		ID:           pad.ID,
		Name:         pad.Name,
		Path:         pad.Path,
		Pinned:       pad.Pinned,
		LastOpenedAt: pad.LastOpenedAt,
	}
}

func (m Model) renderScratchpadPane(width, height int, title string) string {
	if !m.scratchpadOpen || m.scratchpadMeta == nil {
		return m.renderPane(title, PaneScratchpads, width, height, "No scratchpads — [a] create new")
	}

	innerW := max(10, width-6)
	contentH := max(1, height-7)

	lines := []string{
		stylePaneTitle.Render(title),
		styleHeader.Render(truncate(m.scratchpadMeta.Name, innerW)),
		styleDim.Render(truncate(m.scratchpadMeta.Path, innerW)),
	}

	view := m.scratchpadViewport.View()
	viewLines := strings.Split(view, "\n")
	if len(viewLines) > contentH {
		viewLines = viewLines[:contentH]
	}
	lines = append(lines, viewLines...)
	lines = append(lines, styleDim.Render("[h/l] switch  [j/k] scroll  [e] edit  [esc] close"))

	body := strings.Join(lines, "\n")

	style := styleInactive
	if m.pane == PaneScratchpads {
		style = styleActive
	}

	return style.
		Width(width-2).
		Height(height-2).
		Padding(0, 1).
		Render(body)
}

func (m Model) renderScratchpadView(width, h int) string {
	return m.renderScratchpadPane(width, h, "Scratchpads")
}

func (m Model) enterScratchpadPane(msg openScratchpadMsg) Model {
	m.scratchpadOpen = true
	m.scratchpadMeta = &api.ScratchPad{
		ID:           msg.meta.ID,
		Name:         msg.meta.Name,
		Path:         msg.meta.Path,
		Pinned:       msg.meta.Pinned,
		LastOpenedAt: msg.meta.LastOpenedAt,
	}
	m.scratchpadFilePath = msg.filePath
	rendered, err := renderMarkdown(msg.content, m.scratchpadRenderWidth())
	if err != nil {
		logger.Errorf("glamour render: %v", err)
		rendered = msg.content
	}
	m.scratchpadRendered = rendered

	m.syncScratchpadViewport()
	m.scratchpadViewport.GotoTop()

	return m
}

func (m Model) scratchpadRenderWidth() int {
	paneW, _ := m.scratchpadPaneSize()
	return max(20, paneW-6)
}

func renderMarkdown(content string, width int) (string, error) {
	w := width - 2
	if w < 24 {
		w = 24
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(w),
	)
	if err != nil {
		return "", err
	}
	return r.Render(content)
}

func (m Model) updateScratchpadPane(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.scratchpadOpen = false
		m.scratchpadMeta = nil
		m.scratchpadFilePath = ""
		m.scratchpadRendered = ""
		return m, nil

	case "e":
		if m.scratchpadFilePath == "" {
			return m, nil
		}
		return m, openEditor(m.scratchpadFilePath)

	case "t":
		if m.timer != nil && m.timer.State != "idle" {
			next, cmd := m.cycleSelectedIssueStatus()
			return next.(Model), cmd
		}
		return m, nil

	case "A":
		if m.timer != nil && m.timer.State != "idle" {
			next, cmd := m.abandonSelectedIssue()
			return next.(Model), cmd
		}
		return m, nil

	case "left", "h":
		if m.cursor[PaneScratchpads] > 0 {
			m.cursor[PaneScratchpads]--
			rawIdx := m.filteredIndexAtCursor(PaneScratchpads)
			if rawIdx >= 0 {
				m.setActiveScratchpadByIndex(rawIdx)
				return m, cmdOpenScratchpad(m.client, m.scratchpads, rawIdx)
			}
		}
		return m, nil

	case "right", "l":
		if m.cursor[PaneScratchpads] < m.listLen(PaneScratchpads)-1 {
			m.cursor[PaneScratchpads]++
			rawIdx := m.filteredIndexAtCursor(PaneScratchpads)
			if rawIdx >= 0 {
				m.setActiveScratchpadByIndex(rawIdx)
				return m, cmdOpenScratchpad(m.client, m.scratchpads, rawIdx)
			}
		}
		return m, nil

	case "j", "down":
		m.scratchpadViewport.LineDown(1)
		return m, nil

	case "k", "up":
		m.scratchpadViewport.LineUp(1)
		return m, nil

	case "d", "ctrl+d":
		m.scratchpadViewport.HalfViewDown()
		return m, nil

	case "u", "ctrl+u":
		m.scratchpadViewport.HalfViewUp()
		return m, nil

	case "g":
		m.scratchpadViewport.GotoTop()
		return m, nil

	case "G":
		m.scratchpadViewport.GotoBottom()
		return m, nil
	}

	var cmd tea.Cmd
	m.scratchpadViewport, cmd = m.scratchpadViewport.Update(msg)
	return m, cmd
}

type scratchpadReloadedMsg struct {
	rendered string
	filePath string
}

func cmdReloadScratchpad(c *api.Client, meta *api.ScratchPad, width int) tea.Cmd {
	return func() tea.Msg {
		if meta == nil {
			return nil
		}
		filePath, content, err := c.ReadScratchpad(meta.ID)
		if err != nil {
			return errMsg{err}
		}
		rendered, err := renderMarkdown(content, width)
		if err != nil {
			logger.Errorf("glamour render: %v", err)
			rendered = content
		}
		return scratchpadReloadedMsg{rendered: rendered, filePath: filePath}
	}
}

func (m Model) scratchpadTabIndexByID(id string) int {
	for i, pad := range m.scratchpads {
		if pad.ID == id {
			return i
		}
	}
	return -1
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
