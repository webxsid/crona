package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"

	issuecore "crona/tui/internal/tui/views/issuecore"
)

type TablePane struct {
	PaneBase
	Columns     []table.Column
	Rows        []table.Row
	HeaderStyle lipgloss.Style
}

func (p TablePane) View() string {
	return issuecore.IssueTableView(p.Columns, p.Rows, p.HeaderStyle)
}
