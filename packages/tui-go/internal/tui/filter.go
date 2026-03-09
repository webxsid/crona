package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) paneItems(pane Pane) []string {
	switch pane {
	case PaneRepos:
		items := make([]string, 0, len(m.repos))
		for _, repo := range m.repos {
			items = append(items, repo.Name)
		}
		return items
	case PaneStreams:
		items := make([]string, 0, len(m.streams))
		for _, stream := range m.streams {
			items = append(items, stream.Name)
		}
		return items
	case PaneIssues:
		if m.view == ViewDefault {
			items := make([]string, 0, len(m.allIssues))
			for _, issue := range m.allIssues {
				estimate := ""
				if issue.EstimateMinutes != nil {
					estimate = fmt.Sprintf(" %dm", *issue.EstimateMinutes)
				}
				due := issueDueLabel(issue.TodoForDate)
				if due != "" {
					due = " " + due
				}
				items = append(items, fmt.Sprintf("[%s/%s] %s %s%s%s", issue.RepoName, issue.StreamName, issue.Status, issue.Title, estimate, due))
			}
			return items
		}
		if m.view == ViewDaily {
			items := make([]string, 0)
			if m.dailySummary == nil {
				return items
			}
			for _, issue := range m.dailySummary.Issues {
				meta := m.issueMetaByID(issue.ID)
				repoName := "-"
				streamName := "-"
				if meta != nil {
					repoName = meta.RepoName
					streamName = meta.StreamName
				}
				estimate := ""
				if issue.EstimateMinutes != nil {
					estimate = fmt.Sprintf(" %dm", *issue.EstimateMinutes)
				}
				due := issueDueLabel(issue.TodoForDate)
				if due != "" {
					due = " " + due
				}
				items = append(items, fmt.Sprintf("[%s/%s] %s %s%s%s", repoName, streamName, issue.Status, issue.Title, estimate, due))
			}
			return items
		}
		items := make([]string, 0, len(m.issues))
		for _, issue := range m.issues {
			due := issueDueLabel(issue.TodoForDate)
			if due != "" {
				due = " " + due
			}
			items = append(items, fmt.Sprintf("%s %s%s", issue.Status, issue.Title, due))
		}
		return items
	case PaneScratchpads:
		items := make([]string, 0, len(m.scratchpads))
		for _, scratchpad := range m.scratchpads {
			items = append(items, scratchpad.Name)
		}
		return items
	case PaneOps:
		items := make([]string, 0, len(m.ops))
		for _, op := range m.ops {
			ts := op.Timestamp
			if len(ts) >= 19 {
				ts = strings.Replace(ts[:19], "T", " ", 1)
			}
			items = append(items, fmt.Sprintf("%s %s.%s %s", ts, op.Entity, op.Action, op.EntityID))
		}
		return items
	}
	return nil
}

func (m *Model) filteredIndices(pane Pane) []int {
	items := m.paneItems(pane)
	query := strings.TrimSpace(strings.ToLower(m.filters[pane]))
	if query == "" {
		indices := make([]int, len(items))
		for i := range items {
			indices[i] = i
		}
		return indices
	}

	indices := make([]int, 0, len(items))
	for i, item := range items {
		if strings.Contains(strings.ToLower(item), query) {
			indices = append(indices, i)
		}
	}
	return indices
}

func (m *Model) filteredIndexAtCursor(pane Pane) int {
	indices := m.filteredIndices(pane)
	cur := m.cursor[pane]
	if cur < 0 || cur >= len(indices) {
		return -1
	}
	return indices[cur]
}

func (m *Model) filteredCursorForRawIndex(pane Pane, rawIdx int) int {
	indices := m.filteredIndices(pane)
	for i, idx := range indices {
		if idx == rawIdx {
			return i
		}
	}
	return -1
}

func (m *Model) clampFiltered(pane Pane) {
	m.clamp(pane, len(m.filteredIndices(pane)))
}

func (m *Model) startFilterEdit(pane Pane) {
	input := textinput.New()
	input.Placeholder = "filter..."
	input.SetValue(m.filters[pane])
	input.CursorEnd()
	input.Focus()
	input.CharLimit = 120
	input.Width = 24

	m.filterEditing = true
	m.filterPane = pane
	m.filterInput = input
}

func (m *Model) stopFilterEdit() {
	m.filterEditing = false
	m.filterPane = ""
	m.filterInput.Blur()
}

func (m Model) updateFilter(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.stopFilterEdit()
		return m, nil
	case "enter":
		m.stopFilterEdit()
		return m, nil
	}

	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	m.filters[m.filterPane] = m.filterInput.Value()
	m.cursor[m.filterPane] = 0
	m.clampFiltered(m.filterPane)
	return m, cmd
}

func (m Model) renderFilterLine(pane Pane, width int) string {
	if m.filterEditing && m.filterPane == pane {
		value := m.filterInput.View()
		return styleDim.Render("filter: ") + truncate(value, width-8)
	}

	query := m.filters[pane]
	if strings.TrimSpace(query) == "" {
		return styleDim.Render("filter: /")
	}
	return styleDim.Render("filter: ") + truncate(query, width-8)
}
