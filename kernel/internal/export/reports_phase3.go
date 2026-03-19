package export

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/runtime"
	"crona/kernel/internal/sessionnotes"
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

type reportFileMetadata struct {
	Kind       sharedtypes.ExportReportKind `json:"kind"`
	Label      string                       `json:"label"`
	ScopeLabel string                       `json:"scopeLabel,omitempty"`
	Date       string                       `json:"date,omitempty"`
	StartDate  string                       `json:"startDate,omitempty"`
	EndDate    string                       `json:"endDate,omitempty"`
	Format     sharedtypes.ExportFormat     `json:"format"`
}

type reportWriteSpec struct {
	Kind       sharedtypes.ExportReportKind
	Label      string
	ScopeLabel string
	Date       string
	StartDate  string
	EndDate    string
	Format     sharedtypes.ExportFormat
	BaseName   string
}

type reportIssueSession struct {
	sharedtypes.SessionHistoryEntry
}

type reportIssueGroup struct {
	Issue    sharedtypes.IssueWithMeta
	Sessions []reportIssueSession
}

type csvExportSpec struct {
	Columns []csvExportColumn `json:"columns"`
}

type csvExportColumn struct {
	Header string `json:"header"`
	Field  string `json:"field"`
}

func GenerateReport(ctx context.Context, c *core.Context, paths runtime.Paths, input shareddto.ExportReportRequest) (*sharedtypes.ExportReportResult, error) {
	kind := normalizeReportKind(input.Kind)
	switch kind {
	case sharedtypes.ExportReportKindDaily:
		return generateDailyExport(ctx, c, paths, input)
	case sharedtypes.ExportReportKindWeekly:
		return generateWeeklyExport(ctx, c, paths, input)
	case sharedtypes.ExportReportKindRepo:
		return generateRepoExport(ctx, c, paths, input)
	case sharedtypes.ExportReportKindStream:
		return generateStreamExport(ctx, c, paths, input)
	case sharedtypes.ExportReportKindIssueRollup:
		return generateIssueRollupExport(ctx, c, paths, input)
	case sharedtypes.ExportReportKindCSV:
		return generateCSVExport(ctx, c, paths, input)
	default:
		return nil, fmt.Errorf("unsupported report kind %q", input.Kind)
	}
}

func generateDailyExport(ctx context.Context, c *core.Context, paths runtime.Paths, input shareddto.ExportReportRequest) (*sharedtypes.ExportReportResult, error) {
	date := normalizeReportDate(input.Date)
	format := normalizeNarrativeFormat(input.Format)
	return generateDailyReportWithKind(ctx, c, paths, date, format, input.OutputMode)
}

func generateWeeklyExport(ctx context.Context, c *core.Context, paths runtime.Paths, input shareddto.ExportReportRequest) (*sharedtypes.ExportReportResult, error) {
	start, end := normalizeRange(input.Start, input.End, input.Date)
	format := normalizeNarrativeFormat(input.Format)

	days, err := corecommands.ComputeMetricsRange(ctx, c, start, end)
	if err != nil {
		return nil, err
	}
	rollup, err := corecommands.ComputeMetricsRollup(ctx, c, start, end)
	if err != nil {
		return nil, err
	}
	streaks, err := corecommands.ComputeMetricsStreaks(ctx, c, start, end)
	if err != nil {
		return nil, err
	}
	checkIns, err := corecommands.ListDailyCheckInsInRange(ctx, c, start, end)
	if err != nil {
		return nil, err
	}

	checkInByDate := make(map[string]sharedtypes.DailyCheckIn, len(checkIns))
	for _, checkIn := range checkIns {
		checkInByDate[checkIn.Date] = checkIn
	}
	dayItems := make([]map[string]any, 0, len(days))
	for _, day := range days {
		item := map[string]any{
			"date":            day.Date,
			"workedTime":      formatDurationHMS(day.WorkedSeconds),
			"sessionCount":    day.SessionCount,
			"totalIssues":     day.TotalIssues,
			"completedIssues": day.CompletedIssues,
			"abandonedIssues": day.AbandonedIssues,
			"checkIn":         nil,
		}
		if checkIn, ok := checkInByDate[day.Date]; ok {
			item["checkIn"] = map[string]any{"mood": checkIn.Mood, "energy": checkIn.Energy}
		}
		dayItems = append(dayItems, item)
	}

	data := map[string]any{
		"startDate":   start,
		"endDate":     end,
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"summary": map[string]any{
			"days":            rollup.Days,
			"checkInDays":     rollup.CheckInDays,
			"focusDays":       rollup.FocusDays,
			"workedTime":      formatDurationHMS(rollup.WorkedSeconds),
			"restTime":        formatDurationHMS(rollup.RestSeconds),
			"completedIssues": rollup.CompletedIssues,
			"abandonedIssues": rollup.AbandonedIssues,
			"estimatedTime":   fmt.Sprintf("%dm", rollup.TotalEstimatedMinutes),
			"averageMood":     formatOptionalFloat(rollup.AverageMood),
			"averageEnergy":   formatOptionalFloat(rollup.AverageEnergy),
		},
		"streaks": map[string]any{
			"currentFocusDays":   streaks.CurrentFocusDays,
			"longestFocusDays":   streaks.LongestFocusDays,
			"currentCheckInDays": streaks.CurrentCheckInDays,
			"longestCheckInDays": streaks.LongestCheckInDays,
		},
		"days": dayItems,
	}

	return renderNarrativeReport(paths, sharedtypes.ExportReportKindWeekly, data, reportWriteSpec{
		Kind:      sharedtypes.ExportReportKindWeekly,
		Label:     "Weekly Summary",
		Date:      end,
		StartDate: start,
		EndDate:   end,
		Format:    format,
		BaseName:  fmt.Sprintf("weekly-%s-to-%s", start, end),
	}, input.OutputMode)
}

func generateRepoExport(ctx context.Context, c *core.Context, paths runtime.Paths, input shareddto.ExportReportRequest) (*sharedtypes.ExportReportResult, error) {
	if input.RepoID == nil || *input.RepoID == 0 {
		return nil, errors.New("repo report requires repoId")
	}
	repo, err := c.Repos.GetByID(ctx, *input.RepoID, c.UserID)
	if err != nil {
		return nil, err
	}
	if repo == nil {
		return nil, errors.New("repo not found")
	}
	start, end := normalizeRange(input.Start, input.End, input.Date)
	format := normalizeNarrativeFormat(input.Format)
	streams, err := corecommands.ListStreamsByRepo(ctx, c, repo.ID)
	if err != nil {
		return nil, err
	}
	allIssues, err := corecommands.ListAllIssues(ctx, c)
	if err != nil {
		return nil, err
	}
	var issues []sharedtypes.IssueWithMeta
	for _, issue := range allIssues {
		if issue.RepoID == repo.ID {
			issues = append(issues, issue)
		}
	}
	var habits []sharedtypes.Habit
	for _, stream := range streams {
		streamHabits, err := corecommands.ListHabitsByStream(ctx, c, stream.ID)
		if err != nil {
			return nil, err
		}
		habits = append(habits, streamHabits...)
	}
	sessions, err := corecommands.ListSessionHistory(ctx, c, struct {
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
		Since    *string
		Until    *string
		Limit    *int
		Offset   *int
	}{RepoID: &repo.ID, Since: datePtr(start + "T00:00:00.000Z"), Until: datePtr(end + "T23:59:59.999Z")}, false)
	if err != nil {
		return nil, err
	}
	issueGroups := buildIssueGroups(issues, sessions)
	data := map[string]any{
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"startDate":   start,
		"endDate":     end,
		"repo": map[string]any{
			"name":        repo.Name,
			"description": optionalString(repo.Description),
		},
		"summary": map[string]any{
			"streamCount":  len(streams),
			"issueCount":   len(issues),
			"habitCount":   len(habits),
			"sessionCount": len(sessions),
		},
		"streams": mapStreams(streams),
		"issues":  mapDetailedIssueGroups(issueGroups),
		"habits":  mapHabits(habits),
	}
	return renderNarrativeReport(paths, sharedtypes.ExportReportKindRepo, data, reportWriteSpec{
		Kind:       sharedtypes.ExportReportKindRepo,
		Label:      "Repo Report",
		ScopeLabel: repo.Name,
		Date:       end,
		StartDate:  start,
		EndDate:    end,
		Format:     format,
		BaseName:   fmt.Sprintf("repo-%d-%s-%s-to-%s", repo.ID, slugify(repo.Name), start, end),
	}, input.OutputMode, &sharedtypes.ExportReportScope{RepoID: &repo.ID, RepoName: &repo.Name})
}

func generateStreamExport(ctx context.Context, c *core.Context, paths runtime.Paths, input shareddto.ExportReportRequest) (*sharedtypes.ExportReportResult, error) {
	if input.StreamID == nil || *input.StreamID == 0 {
		return nil, errors.New("stream report requires streamId")
	}
	stream, err := c.Streams.GetByID(ctx, *input.StreamID, c.UserID)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, errors.New("stream not found")
	}
	repo, err := c.Repos.GetByID(ctx, stream.RepoID, c.UserID)
	if err != nil {
		return nil, err
	}
	start, end := normalizeRange(input.Start, input.End, input.Date)
	format := normalizeNarrativeFormat(input.Format)
	issues, err := corecommands.ListIssuesByStream(ctx, c, stream.ID)
	if err != nil {
		return nil, err
	}
	habits, err := corecommands.ListHabitsByStream(ctx, c, stream.ID)
	if err != nil {
		return nil, err
	}
	sessions, err := corecommands.ListSessionHistory(ctx, c, struct {
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
		Since    *string
		Until    *string
		Limit    *int
		Offset   *int
	}{StreamID: &stream.ID, Since: datePtr(start + "T00:00:00.000Z"), Until: datePtr(end + "T23:59:59.999Z")}, false)
	if err != nil {
		return nil, err
	}
	repoName := "-"
	if repo != nil {
		repoName = repo.Name
	}
	issueGroups := buildIssueGroupsWithStream(issues, repoName, stream.Name, sessions)
	scope := &sharedtypes.ExportReportScope{
		RepoID:     &stream.RepoID,
		StreamID:   &stream.ID,
		StreamName: &stream.Name,
	}
	if repo != nil {
		scope.RepoName = &repo.Name
	}
	data := map[string]any{
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"startDate":   start,
		"endDate":     end,
		"repo": map[string]any{
			"name": repoName,
		},
		"stream": map[string]any{
			"name":        stream.Name,
			"description": optionalString(stream.Description),
		},
		"summary": map[string]any{
			"issueCount":   len(issues),
			"habitCount":   len(habits),
			"sessionCount": len(sessions),
		},
		"issues": mapDetailedIssueGroups(issueGroups),
		"habits": mapHabits(habits),
	}
	return renderNarrativeReport(paths, sharedtypes.ExportReportKindStream, data, reportWriteSpec{
		Kind:       sharedtypes.ExportReportKindStream,
		Label:      "Stream Report",
		ScopeLabel: joinNonEmpty(" / ", repoName, stream.Name),
		Date:       end,
		StartDate:  start,
		EndDate:    end,
		Format:     format,
		BaseName:   fmt.Sprintf("stream-%d-%s-%s-to-%s", stream.ID, slugify(stream.Name), start, end),
	}, input.OutputMode, scope)
}

func generateIssueRollupExport(ctx context.Context, c *core.Context, paths runtime.Paths, input shareddto.ExportReportRequest) (*sharedtypes.ExportReportResult, error) {
	start, end := normalizeRange(input.Start, input.End, input.Date)
	format := normalizeNarrativeFormat(input.Format)
	entries, err := corecommands.ListSessionHistory(ctx, c, struct {
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
		Since    *string
		Until    *string
		Limit    *int
		Offset   *int
	}{RepoID: input.RepoID, StreamID: input.StreamID, Since: datePtr(start + "T00:00:00.000Z"), Until: datePtr(end + "T23:59:59.999Z")}, false)
	if err != nil {
		return nil, err
	}
	allIssues, err := corecommands.ListAllIssues(ctx, c)
	if err != nil {
		return nil, err
	}
	issueMeta := make(map[int64]sharedtypes.IssueWithMeta, len(allIssues))
	for _, issue := range allIssues {
		issueMeta[issue.ID] = issue
	}
	type rollup struct {
		issue        sharedtypes.IssueWithMeta
		title        string
		status       sharedtypes.IssueStatus
		repoName     string
		streamName   string
		estimateMins int
		sessions     int
		workedSecs   int
		entries      []reportIssueSession
	}
	rollups := map[int64]*rollup{}
	for _, entry := range entries {
		meta, ok := issueMeta[entry.IssueID]
		if !ok {
			continue
		}
		parsed := sessionnotes.Parse(entry.Notes)
		entry.ParsedNotes = parsed
		item := rollups[entry.IssueID]
		if item == nil {
			item = &rollup{
				issue:      meta,
				title:      meta.Title,
				status:     meta.Status,
				repoName:   meta.RepoName,
				streamName: meta.StreamName,
			}
			if meta.EstimateMinutes != nil {
				item.estimateMins = *meta.EstimateMinutes
			}
			rollups[entry.IssueID] = item
		}
		item.sessions++
		if entry.DurationSeconds != nil {
			item.workedSecs += *entry.DurationSeconds
		}
		item.entries = append(item.entries, reportIssueSession{SessionHistoryEntry: entry})
	}
	ids := make([]int64, 0, len(rollups))
	for id := range rollups {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return rollups[ids[i]].workedSecs > rollups[ids[j]].workedSecs })
	detailedGroups := make([]reportIssueGroup, 0, len(ids))
	for _, id := range ids {
		item := rollups[id]
		detailedGroups = append(detailedGroups, reportIssueGroup{
			Issue:    item.issue,
			Sessions: item.entries,
		})
	}
	data := map[string]any{
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"startDate":   start,
		"endDate":     end,
		"issues":      mapDetailedIssueGroups(detailedGroups, map[int64]map[string]any{}),
	}
	return renderNarrativeReport(paths, sharedtypes.ExportReportKindIssueRollup, data, reportWriteSpec{
		Kind:      sharedtypes.ExportReportKindIssueRollup,
		Label:     "Session to Issue Rollup",
		Date:      end,
		StartDate: start,
		EndDate:   end,
		Format:    format,
		BaseName:  fmt.Sprintf("issue-rollup-%s-to-%s", start, end),
	}, input.OutputMode)
}

func generateCSVExport(ctx context.Context, c *core.Context, paths runtime.Paths, input shareddto.ExportReportRequest) (*sharedtypes.ExportReportResult, error) {
	start, end := normalizeRange(input.Start, input.End, input.Date)
	entries, err := corecommands.ListSessionHistory(ctx, c, struct {
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
		Since    *string
		Until    *string
		Limit    *int
		Offset   *int
	}{RepoID: input.RepoID, StreamID: input.StreamID, Since: datePtr(start + "T00:00:00.000Z"), Until: datePtr(end + "T23:59:59.999Z")}, false)
	if err != nil {
		return nil, err
	}
	allIssues, err := corecommands.ListAllIssues(ctx, c)
	if err != nil {
		return nil, err
	}
	issueMeta := make(map[int64]sharedtypes.IssueWithMeta, len(allIssues))
	for _, issue := range allIssues {
		issueMeta[issue.ID] = issue
	}
	rowMaps := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		meta := issueMeta[entry.IssueID]
		parsed := sessionnotes.Parse(entry.Notes)
		duration := ""
		if entry.DurationSeconds != nil {
			duration = strconv.Itoa(*entry.DurationSeconds)
		}
		endTime := ""
		if entry.EndTime != nil {
			endTime = *entry.EndTime
		}
		rowMaps = append(rowMaps, map[string]any{
			"sessionId":       entry.ID,
			"issueId":         entry.IssueID,
			"issueTitle":      meta.Title,
			"status":          string(meta.Status),
			"repo":            meta.RepoName,
			"stream":          meta.StreamName,
			"startTime":       entry.StartTime,
			"endTime":         endTime,
			"durationSeconds": duration,
			"commit":          strings.TrimSpace(parsed[sharedtypes.SessionNoteSectionCommit]),
			"context":         strings.TrimSpace(parsed[sharedtypes.SessionNoteSectionContext]),
			"work":            strings.TrimSpace(parsed[sharedtypes.SessionNoteSectionWork]),
			"notes":           strings.TrimSpace(parsed[sharedtypes.SessionNoteSectionNotes]),
		})
	}
	content, err := renderCSVExport(paths, rowMaps)
	if err != nil {
		return nil, err
	}
	return finalizeReport(paths, reportWriteSpec{
		Kind:      sharedtypes.ExportReportKindCSV,
		Label:     "CSV Export",
		Date:      end,
		StartDate: start,
		EndDate:   end,
		Format:    sharedtypes.ExportFormatCSV,
		BaseName:  fmt.Sprintf("sessions-%s-to-%s", start, end),
	}, content, sharedtypes.ExportOutputModeFile)
}

func finalizeReport(paths runtime.Paths, spec reportWriteSpec, content string, mode sharedtypes.ExportOutputMode, scope ...*sharedtypes.ExportReportScope) (*sharedtypes.ExportReportResult, error) {
	format := normalizeFormatForKind(spec.Kind, spec.Format)
	if spec.Kind == sharedtypes.ExportReportKindCSV && mode != sharedtypes.ExportOutputModeFile {
		return nil, errors.New("csv export only supports file output")
	}
	if format == sharedtypes.ExportFormatPDF && mode != sharedtypes.ExportOutputModeFile {
		return nil, errors.New("pdf export only supports file output")
	}
	assets, err := EnsureAssets(paths)
	if err != nil {
		return nil, err
	}
	result := &sharedtypes.ExportReportResult{
		Kind:       spec.Kind,
		Label:      spec.Label,
		Date:       spec.Date,
		StartDate:  spec.StartDate,
		EndDate:    spec.EndDate,
		Format:     format,
		OutputMode: mode,
		Content:    content,
		Assets:     assets,
	}
	if len(scope) > 0 {
		result.Scope = scope[0]
	}
	if mode != sharedtypes.ExportOutputModeFile {
		return result, nil
	}
	if format == sharedtypes.ExportFormatPDF {
		filePath, renderer, err := RenderPDFReport(paths, spec, content)
		if err != nil {
			return nil, err
		}
		result.FilePath = &filePath
		result.Renderer = &renderer
		return result, nil
	}
	filePath, err := WriteReport(paths, spec, []byte(content))
	if err != nil {
		return nil, err
	}
	result.FilePath = &filePath
	return result, nil
}

func renderNarrativeReport(paths runtime.Paths, kind sharedtypes.ExportReportKind, data map[string]any, spec reportWriteSpec, mode sharedtypes.ExportOutputMode, scope ...*sharedtypes.ExportReportScope) (*sharedtypes.ExportReportResult, error) {
	templateBody, _, err := LoadActiveReportTemplate(paths, kind, spec.Format)
	if err != nil {
		return nil, err
	}
	rendered, err := RenderTemplate(string(templateBody), data)
	if err != nil {
		return nil, err
	}
	return finalizeReport(paths, spec, strings.TrimSpace(rendered), mode, scope...)
}

func normalizeReportKind(kind sharedtypes.ExportReportKind) sharedtypes.ExportReportKind {
	switch kind {
	case sharedtypes.ExportReportKindWeekly, sharedtypes.ExportReportKindRepo, sharedtypes.ExportReportKindStream, sharedtypes.ExportReportKindIssueRollup, sharedtypes.ExportReportKindCSV:
		return kind
	default:
		return sharedtypes.ExportReportKindDaily
	}
}

func normalizeReportDate(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Now().Format("2006-01-02")
	}
	return trimmed
}

func normalizeRange(start, end, date string) (string, string) {
	trimmedStart := strings.TrimSpace(start)
	trimmedEnd := strings.TrimSpace(end)
	if trimmedStart != "" && trimmedEnd != "" {
		return trimmedStart, trimmedEnd
	}
	anchor := normalizeReportDate(date)
	t, err := time.Parse("2006-01-02", anchor)
	if err != nil {
		return anchor, anchor
	}
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := t.AddDate(0, 0, -(weekday - 1))
	return weekStart.Format("2006-01-02"), weekStart.AddDate(0, 0, 6).Format("2006-01-02")
}

func normalizeNarrativeFormat(format sharedtypes.ExportFormat) sharedtypes.ExportFormat {
	if format == sharedtypes.ExportFormatPDF {
		return sharedtypes.ExportFormatPDF
	}
	return sharedtypes.ExportFormatMarkdown
}

func normalizeFormatForKind(kind sharedtypes.ExportReportKind, format sharedtypes.ExportFormat) sharedtypes.ExportFormat {
	if kind == sharedtypes.ExportReportKindCSV {
		return sharedtypes.ExportFormatCSV
	}
	return normalizeNarrativeFormat(format)
}

func slugify(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return "report"
	}
	var b strings.Builder
	lastDash := false
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "report"
	}
	return out
}

func joinNonEmpty(sep string, values ...string) string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" && trimmed != "-" {
			out = append(out, trimmed)
		}
	}
	return strings.Join(out, sep)
}

func estimateLabel(value *int) string {
	if value == nil {
		return "no estimate"
	}
	return fmt.Sprintf("%dm", *value)
}

func estimateTableLabel(value int) string {
	if value <= 0 {
		return "-"
	}
	return fmt.Sprintf("%dm", value)
}

func formatOptionalFloat(value *float64) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%.1f", *value)
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func escapeTable(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}

func csvField(value string) string {
	if strings.ContainsAny(value, ",\"\n") {
		return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
	}
	return value
}

func renderCSVExport(paths runtime.Paths, rows []map[string]any) (string, error) {
	spec, err := loadCSVExportSpec(paths)
	if err != nil {
		return "", err
	}
	if len(spec.Columns) == 0 {
		return "", errors.New("csv export spec has no columns")
	}
	headers := make([]string, 0, len(spec.Columns))
	for _, column := range spec.Columns {
		headers = append(headers, csvField(strings.TrimSpace(column.Header)))
	}
	lines := []string{strings.Join(headers, ",")}
	for _, row := range rows {
		values := make([]string, 0, len(spec.Columns))
		for _, column := range spec.Columns {
			values = append(values, csvField(fmt.Sprint(resolvePath(strings.TrimSpace(column.Field), row, row))))
		}
		lines = append(lines, strings.Join(values, ","))
	}
	return strings.Join(lines, "\n"), nil
}

func loadCSVExportSpec(paths runtime.Paths) (*csvExportSpec, error) {
	status, err := EnsureAssets(paths)
	if err != nil {
		return nil, err
	}
	for _, asset := range status.TemplateAssets {
		if asset.ReportKind == sharedtypes.ExportReportKindCSV && asset.AssetKind == sharedtypes.ExportAssetKindCSVSpec {
			body, err := os.ReadFile(asset.UserPath)
			if err != nil {
				return nil, err
			}
			var spec csvExportSpec
			if err := json.Unmarshal(body, &spec); err != nil {
				return nil, err
			}
			return &spec, nil
		}
	}
	return nil, errors.New("csv export spec not found")
}

func buildIssueGroups(issues []sharedtypes.IssueWithMeta, sessions []sharedtypes.SessionHistoryEntry) []reportIssueGroup {
	groups := make([]reportIssueGroup, 0, len(issues))
	indexByIssue := make(map[int64]int, len(issues))
	for _, issue := range issues {
		indexByIssue[issue.ID] = len(groups)
		groups = append(groups, reportIssueGroup{Issue: issue})
	}
	for _, session := range sessions {
		idx, ok := indexByIssue[session.IssueID]
		if !ok {
			continue
		}
		session.ParsedNotes = sessionnotes.Parse(session.Notes)
		groups[idx].Sessions = append(groups[idx].Sessions, reportIssueSession{SessionHistoryEntry: session})
	}
	return groups
}

func buildIssueGroupsWithStream(issues []sharedtypes.Issue, repoName string, streamName string, sessions []sharedtypes.SessionHistoryEntry) []reportIssueGroup {
	withMeta := make([]sharedtypes.IssueWithMeta, 0, len(issues))
	for _, issue := range issues {
		withMeta = append(withMeta, sharedtypes.IssueWithMeta{
			Issue:      issue,
			RepoName:   repoName,
			StreamName: streamName,
		})
	}
	return buildIssueGroups(withMeta, sessions)
}

func renderDetailedIssueGroups(groups []reportIssueGroup) []string {
	lines := make([]string, 0, len(groups)*8)
	for _, group := range groups {
		issue := group.Issue
		lines = append(lines, fmt.Sprintf("### #%d %s", issue.ID, issue.Title))
		lines = append(lines, fmt.Sprintf("- Status: %s", issue.Status))
		lines = append(lines, fmt.Sprintf("- Estimate: %s", estimateLabel(issue.EstimateMinutes)))
		if strings.TrimSpace(issue.RepoName) != "" || strings.TrimSpace(issue.StreamName) != "" {
			lines = append(lines, fmt.Sprintf("- Scope: %s", joinNonEmpty(" / ", issue.RepoName, issue.StreamName)))
		}
		if issue.Description != nil && strings.TrimSpace(*issue.Description) != "" {
			lines = append(lines, "", "Description", strings.TrimSpace(*issue.Description))
		}
		if issue.Notes != nil && strings.TrimSpace(*issue.Notes) != "" {
			lines = append(lines, "", "Issue Notes", strings.TrimSpace(*issue.Notes))
		}
		lines = append(lines, "")
		lines = append(lines, renderIssueSessionsSection(group.Sessions)...)
		lines = append(lines, "")
	}
	return trimTrailingBlanks(lines)
}

func renderIssueSessionsSection(sessions []reportIssueSession) []string {
	if len(sessions) == 0 {
		return []string{"Sessions", "- No sessions in range"}
	}
	lines := []string{"Sessions"}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].StartTime < sessions[j].StartTime })
	for _, session := range sessions {
		lines = append(lines, fmt.Sprintf("- %s", renderSessionSummaryLine(session.SessionHistoryEntry)))
		sessionLines := renderParsedSessionNotes(session.SessionHistoryEntry)
		for _, line := range sessionLines {
			lines = append(lines, "  "+line)
		}
	}
	return lines
}

func renderSessionSummaryLine(session sharedtypes.SessionHistoryEntry) string {
	duration := "unknown duration"
	if session.DurationSeconds != nil {
		duration = formatDurationHMS(*session.DurationSeconds)
	}
	endTime := "open"
	if session.EndTime != nil && strings.TrimSpace(*session.EndTime) != "" {
		endTime = *session.EndTime
	}
	return fmt.Sprintf("%s -> %s (%s)", session.StartTime, endTime, duration)
}

func renderParsedSessionNotes(session sharedtypes.SessionHistoryEntry) []string {
	lines := []string{}
	sections := []struct {
		label string
		key   sharedtypes.SessionNoteSection
	}{
		{label: "Commit", key: sharedtypes.SessionNoteSectionCommit},
		{label: "Context", key: sharedtypes.SessionNoteSectionContext},
		{label: "Work", key: sharedtypes.SessionNoteSectionWork},
		{label: "Notes", key: sharedtypes.SessionNoteSectionNotes},
	}
	for _, section := range sections {
		text := strings.TrimSpace(session.ParsedNotes[section.key])
		if text == "" {
			continue
		}
		lines = append(lines, section.label+": "+firstLine(text))
		rest := remainingLines(text)
		for _, line := range rest {
			lines = append(lines, "  "+line)
		}
	}
	if len(lines) > 0 {
		return lines
	}
	if session.Notes != nil && strings.TrimSpace(*session.Notes) != "" {
		return []string{"Notes: " + strings.TrimSpace(*session.Notes)}
	}
	return []string{"Notes: -"}
}

func firstLine(text string) string {
	parts := strings.Split(strings.TrimSpace(text), "\n")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func remainingLines(text string) []string {
	parts := strings.Split(strings.TrimSpace(text), "\n")
	if len(parts) <= 1 {
		return nil
	}
	out := make([]string, 0, len(parts)-1)
	for _, part := range parts[1:] {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func trimTrailingBlanks(lines []string) []string {
	end := len(lines)
	for end > 0 && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return lines[:end]
}

func mapStreams(streams []sharedtypes.Stream) []map[string]any {
	items := make([]map[string]any, 0, len(streams))
	for _, stream := range streams {
		items = append(items, map[string]any{
			"id":          stream.ID,
			"name":        stream.Name,
			"description": optionalString(stream.Description),
		})
	}
	return items
}

func mapHabits(habits []sharedtypes.Habit) []map[string]any {
	items := make([]map[string]any, 0, len(habits))
	for _, habit := range habits {
		items = append(items, map[string]any{
			"id":           habit.ID,
			"name":         habit.Name,
			"scheduleType": string(habit.ScheduleType),
		})
	}
	return items
}

func mapDetailedIssueGroups(groups []reportIssueGroup, enrich ...map[int64]map[string]any) []map[string]any {
	var extra map[int64]map[string]any
	if len(enrich) > 0 {
		extra = enrich[0]
	}
	items := make([]map[string]any, 0, len(groups))
	for _, group := range groups {
		issue := group.Issue
		item := map[string]any{
			"id":           issue.ID,
			"title":        issue.Title,
			"status":       string(issue.Status),
			"scope":        joinNonEmpty(" / ", issue.RepoName, issue.StreamName),
			"estimateTime": estimateLabel(issue.EstimateMinutes),
			"description":  optionalString(issue.Description),
			"notes":        optionalString(issue.Notes),
			"sessionCount": len(group.Sessions),
			"workedTime":   workedTimeForSessions(group.Sessions),
			"sessions":     mapIssueSessions(group.Sessions),
		}
		if extra != nil {
			for key, value := range extra[issue.ID] {
				item[key] = value
			}
		}
		items = append(items, item)
	}
	return items
}

func workedTimeForSessions(sessions []reportIssueSession) string {
	total := 0
	for _, session := range sessions {
		if session.DurationSeconds != nil {
			total += *session.DurationSeconds
		}
	}
	return formatDurationHMS(total)
}

func mapIssueSessions(sessions []reportIssueSession) []map[string]any {
	items := make([]map[string]any, 0, len(sessions))
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].StartTime < sessions[j].StartTime })
	for _, session := range sessions {
		items = append(items, map[string]any{
			"id":      session.ID,
			"summary": renderSessionSummaryLine(session.SessionHistoryEntry),
			"commit":  strings.TrimSpace(session.ParsedNotes[sharedtypes.SessionNoteSectionCommit]),
			"context": strings.TrimSpace(session.ParsedNotes[sharedtypes.SessionNoteSectionContext]),
			"work":    strings.TrimSpace(session.ParsedNotes[sharedtypes.SessionNoteSectionWork]),
			"notes":   sessionNoteText(session.SessionHistoryEntry),
		})
	}
	return items
}

func sessionNoteText(session sharedtypes.SessionHistoryEntry) string {
	if text := strings.TrimSpace(session.ParsedNotes[sharedtypes.SessionNoteSectionNotes]); text != "" {
		return text
	}
	if session.Notes != nil {
		return strings.TrimSpace(*session.Notes)
	}
	return ""
}

func datePtr(value string) *string {
	return &value
}

func writeReportMetadata(path string, metadata reportFileMetadata) error {
	body, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, body)
}

func readReportMetadata(path string) (*reportFileMetadata, error) {
	body, err := readFile(path)
	if err != nil {
		return nil, err
	}
	var metadata reportFileMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

func metadataPathForReport(path string) string {
	return path + ".meta.json"
}

func reportDisplayDateLabel(date, start, end string) string {
	switch {
	case strings.TrimSpace(start) != "" && strings.TrimSpace(end) != "":
		return start + " to " + end
	case strings.TrimSpace(date) != "":
		return date
	default:
		return ""
	}
}

func writeFile(path string, body []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, body, runtime.FilePerm())
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
