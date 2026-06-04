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
	issueTableKeyWorked   = "worked"
	issueTableKeyRepo     = "repo"
	issueTableKeyStream   = "stream"
	issueTableKeyContext  = "context"
	issueTableKeyEffort   = "effort"

	issueTableCompactBreakpoint = 96
)

type IssueTableLayout struct {
	Compact bool

	TitleW    int
	StatusW   int
	EstimateW int
	WorkedW   int
	RepoW     int
	StreamW   int
	ContextW  int
	EffortW   int
}

func IssueTableLayoutForWidth(width int) IssueTableLayout {
	if width < issueTableCompactBreakpoint {
		statusW := 14
		contextW := max(18, width/4)
		effortW := max(18, width/4)
		titleW := width - statusW - contextW - effortW - 16
		titleW = max(14, titleW)
		return IssueTableLayout{
			Compact:  true,
			TitleW:   titleW,
			StatusW:  statusW,
			ContextW: contextW,
			EffortW:  effortW,
		}
	}
	statusW := 14
	estimateW := 7
	workedW := 11
	repoW := max(10, width/8)
	streamW := max(10, width/8)
	titleW := width - repoW - streamW - statusW - estimateW - workedW - 20
	titleW = max(14, titleW)
	return IssueTableLayout{
		TitleW:    titleW,
		StatusW:   statusW,
		EstimateW: estimateW,
		WorkedW:   workedW,
		RepoW:     repoW,
		StreamW:   streamW,
	}
}

func IssueTableColumns(layout IssueTableLayout) []table.Column {
	if layout.Compact {
		return []table.Column{
			table.NewColumn(issueTableKeyCursor, "", 2),
			table.NewColumn(issueTableKeyIssue, "Issue", layout.TitleW).WithStyle(issueColumnStyle()),
			table.NewColumn(issueTableKeyStatus, "Status", layout.StatusW).WithStyle(issueColumnStyle()),
			table.NewColumn(issueTableKeyContext, "Context", layout.ContextW).WithStyle(issueColumnStyle()),
			table.NewColumn(issueTableKeyEffort, "Effort", layout.EffortW).WithStyle(issueColumnStyle()),
		}
	}
	return []table.Column{
		table.NewColumn(issueTableKeyCursor, "", 2),
		table.NewColumn(issueTableKeyIssue, "Issue", layout.TitleW).WithStyle(issueColumnStyle()),
		table.NewColumn(issueTableKeyStatus, "Status", layout.StatusW).WithStyle(issueColumnStyle()),
		table.NewColumn(issueTableKeyEstimate, "Est.", layout.EstimateW).WithStyle(issueColumnStyle()),
		table.NewColumn(issueTableKeyWorked, "Worked", layout.WorkedW).WithStyle(issueColumnStyle()),
		table.NewColumn(issueTableKeyRepo, "Repo", layout.RepoW).WithStyle(issueColumnStyle()),
		table.NewColumn(issueTableKeyStream, "Stream", layout.StreamW).WithStyle(issueColumnStyle()),
	}
}

func IssueTableRow(
	cursor string,
	data IssueTableData,
	rowStyle lipgloss.Style,
) table.Row {
	return table.NewRow(table.RowData{
		issueTableKeyCursor:   cursor,
		issueTableKeyIssue:    data.Issue,
		issueTableKeyStatus:   data.Status,
		issueTableKeyEstimate: data.Estimate,
		issueTableKeyWorked:   data.Worked,
		issueTableKeyRepo:     data.Repo,
		issueTableKeyStream:   data.Stream,
		issueTableKeyContext:  data.Context,
		issueTableKeyEffort:   data.Effort,
	}).WithStyle(rowStyle)
}

func IssueTableView(columns []table.Column, rows []table.Row, headerStyle lipgloss.Style) string {
	return table.New(columns).
		WithRows(rows).
		WithNoPagination().
		WithFooterVisibility(false).
		WithBaseStyle(lipgloss.NewStyle().Align(lipgloss.Left)).
		HeaderStyle(headerStyle.Align(lipgloss.Left)).
		Border(table.Border{}).
		View()
}

type IssueTableData struct {
	Issue    string
	Status   string
	Estimate string
	Worked   string
	Repo     string
	Stream   string
	Context  string
	Effort   string
}

func issueColumnStyle() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1)
}
