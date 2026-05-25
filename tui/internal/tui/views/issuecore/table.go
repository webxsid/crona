package issuecore

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	issueTableKeyCursor   = "cursor"
	issueTableKeyIssue    = "issue"
	issueTableKeyStatus   = "status"
	issueTableKeyEstimate = "estimate"
	issueTableKeySpent    = "spent"
	issueTableKeyRepo     = "repo"
	issueTableKeyStream   = "stream"
)

func IssueTableColumns(titleW, statusW, estimateW, spentW, repoW, streamW int) []table.Column {
	return []table.Column{
		table.NewColumn(issueTableKeyCursor, "", 2),
		table.NewColumn(issueTableKeyIssue, "Issue", titleW),
		table.NewColumn(issueTableKeyStatus, "Status", statusW),
		table.NewColumn(issueTableKeyEstimate, "Estimate", estimateW),
		table.NewColumn(issueTableKeySpent, "Spent", spentW),
		table.NewColumn(issueTableKeyRepo, "Repo", repoW),
		table.NewColumn(issueTableKeyStream, "Stream", streamW),
	}
}

func IssueTableRow(
	cursor, issue, status, estimate, spent, repo, stream string,
	rowStyle lipgloss.Style,
) table.Row {
	return table.NewRow(table.RowData{
		issueTableKeyCursor:   cursor,
		issueTableKeyIssue:    issue,
		issueTableKeyStatus:   status,
		issueTableKeyEstimate: estimate,
		issueTableKeySpent:    spent,
		issueTableKeyRepo:     repo,
		issueTableKeyStream:   stream,
	}).WithStyle(rowStyle)
}

func IssueTableView(columns []table.Column, rows []table.Row, headerStyle lipgloss.Style) string {
	return table.New(columns).
		WithRows(rows).
		WithNoPagination().
		WithFooterVisibility(false).
		WithBaseStyle(lipgloss.NewStyle().Align(lipgloss.Left)).
		HeaderStyle(headerStyle.Copy().Align(lipgloss.Left)).
		Border(table.Border{}).
		View()
}
