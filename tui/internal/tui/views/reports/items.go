package reports

import (
	"fmt"
	"path/filepath"

	"crona/tui/internal/api"
)

func reportItems(reports []api.ExportReportFile) []string {
	items := make([]string, 0, len(reports))
	for _, report := range reports {
		scope := report.ScopeLabel
		if scope == "" {
			scope = "-"
		}
		dateLabel := report.DateLabel
		if dateLabel == "" {
			dateLabel = report.Date
		}
		items = append(items, fmt.Sprintf("[%s] %s    %s    [%s] %s    %d B", report.Kind, scope, dateLabel, report.Format, filepath.Base(report.Path), report.SizeBytes))
	}
	return items
}
