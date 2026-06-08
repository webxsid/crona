package issuecore

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	issueTableKeyCursor   = "cursor"
	issueTableKeyIssue    = "issue"
	issueTableKeyStatus   = "status"
	issueTableKeyGap1     = "gap1"
	issueTableKeyGap2     = "gap2"
	issueTableKeyGap3     = "gap3"
	issueTableKeyGap4     = "gap4"
	issueTableKeyGap5     = "gap5"
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
		titleW := width - statusW - contextW - effortW - 19
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
	titleW := width - repoW - streamW - statusW - estimateW - workedW - 25
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
			table.NewColumn(issueTableKeyIssue, "Issue", layout.TitleW).WithStyle(issueColumnStyle(true)),
			table.NewColumn(issueTableKeyGap1, "", 1),
			table.NewColumn(issueTableKeyStatus, "Status", layout.StatusW).WithStyle(issueColumnStyle(true)),
			table.NewColumn(issueTableKeyGap2, "", 1),
			table.NewColumn(issueTableKeyContext, "Context", layout.ContextW).WithStyle(issueColumnStyle(true)),
			table.NewColumn(issueTableKeyGap3, "", 1),
			table.NewColumn(issueTableKeyEffort, "Effort", layout.EffortW).WithStyle(issueColumnStyle(true)),
		}
	}
	return []table.Column{
		table.NewColumn(issueTableKeyCursor, "", 2),
		table.NewColumn(issueTableKeyIssue, "Issue", layout.TitleW).WithStyle(issueColumnStyle(false)),
		table.NewColumn(issueTableKeyGap1, "", 1),
		table.NewColumn(issueTableKeyStatus, "Status", layout.StatusW).WithStyle(issueColumnStyle(false)),
		table.NewColumn(issueTableKeyGap2, "", 1),
		table.NewColumn(issueTableKeyEstimate, "Est.", layout.EstimateW).WithStyle(issueColumnStyle(false)),
		table.NewColumn(issueTableKeyGap3, "", 1),
		table.NewColumn(issueTableKeyWorked, "Worked", layout.WorkedW).WithStyle(issueColumnStyle(false)),
		table.NewColumn(issueTableKeyGap4, "", 1),
		table.NewColumn(issueTableKeyRepo, "Repo", layout.RepoW).WithStyle(issueColumnStyle(false)),
		table.NewColumn(issueTableKeyGap5, "", 1),
		table.NewColumn(issueTableKeyStream, "Stream", layout.StreamW).WithStyle(issueColumnStyle(false)),
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
		issueTableKeyGap1:     "",
		issueTableKeyStatus:   data.Status,
		issueTableKeyGap2:     "",
		issueTableKeyEstimate: data.Estimate,
		issueTableKeyGap3:     "",
		issueTableKeyWorked:   data.Worked,
		issueTableKeyGap4:     "",
		issueTableKeyRepo:     data.Repo,
		issueTableKeyGap5:     "",
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

func issueColumnStyle(compact bool) lipgloss.Style {
	if compact {
		return lipgloss.NewStyle().Padding(0, 2)
	}
	return lipgloss.NewStyle().Padding(0, 1)
}
