package export

import (
	"crona/kernel/internal/runtime"
	sharedtypes "crona/shared/types"
)

func ReportWriteSpecForTesting(kind sharedtypes.ExportReportKind, label, scopeLabel, date, startDate, endDate string, format sharedtypes.ExportFormat, baseName string) reportWriteSpec {
	return reportWriteSpec{
		Kind:       kind,
		Label:      label,
		ScopeLabel: scopeLabel,
		Date:       date,
		StartDate:  startDate,
		EndDate:    endDate,
		Format:     format,
		BaseName:   baseName,
	}
}

func WriteFileForTesting(path string, body []byte) error {
	return writeFile(path, body)
}

func ResolveReportsDirForTesting(paths runtime.Paths, raw string) (string, error) {
	return normalizeReportsDir(paths, raw)
}

func RenderDetailedIssueGroupForTesting(issue sharedtypes.IssueWithMeta, sessions []sharedtypes.SessionHistoryEntry) []string {
	group := reportIssueGroup{Issue: issue}
	for _, session := range sessions {
		group.Sessions = append(group.Sessions, reportIssueSession{SessionHistoryEntry: session})
	}
	return renderDetailedIssueGroups([]reportIssueGroup{group})
}
