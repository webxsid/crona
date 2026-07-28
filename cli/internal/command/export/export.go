package export

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	contextcmd "crona/cli/internal/command/context"
	flagspkg "crona/cli/internal/flags"
	outputpkg "crona/cli/internal/output"
	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

type Deps struct {
	Stdout     io.Writer
	CallKernel func(method string, params, out any) error
}

func Usage() string {
	return "Usage: crona export <summary|daily|weekly|repo|stream|issue-rollup|csv|calendar|reports> ...\n"
}

func Run(args []string, deps Deps) error {
	if len(args) == 0 || flagspkg.IsHelpArg(args[0]) {
		_, err := fmt.Fprint(deps.Stdout, Usage())
		return err
	}
	switch args[0] {
	case "summary", "glance":
		return runReport(args[1:], sharedtypes.ExportReportKindGlance, deps)
	case "daily":
		return runReport(args[1:], sharedtypes.ExportReportKindDaily, deps)
	case "weekly":
		return runReport(args[1:], sharedtypes.ExportReportKindWeekly, deps)
	case "repo":
		return runReport(args[1:], sharedtypes.ExportReportKindRepo, deps)
	case "stream":
		return runReport(args[1:], sharedtypes.ExportReportKindStream, deps)
	case "issue-rollup":
		return runReport(args[1:], sharedtypes.ExportReportKindIssueRollup, deps)
	case "csv":
		return runReport(args[1:], sharedtypes.ExportReportKindCSV, deps)
	case "calendar":
		return runCalendar(args[1:], deps)
	case "reports":
		return runReports(args[1:], deps)
	default:
		return fmt.Errorf("unknown export command: %s", args[0])
	}
}

func runCalendar(args []string, deps Deps) error {
	fs := flagspkg.New("export calendar")
	repoID := fs.Int64("repo-id", 0, "")
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	selectedRepoID, err := resolveCalendarRepoID(*repoID, deps)
	if err != nil {
		return err
	}
	req := shareddto.ExportCalendarRequest{RepoID: selectedRepoID}
	var out sharedtypes.CalendarExportResult
	if err := deps.CallKernel(protocol.MethodExportCalendar, req, &out); err != nil {
		return err
	}
	if strings.TrimSpace(out.IssuesFilePath) == "" ||
		strings.TrimSpace(out.SessionsFilePath) == "" {
		return errors.New(
			"calendar export response is incomplete; restart the kernel so the updated export handler is loaded",
		)
	}
	if *jsonOut {
		return outputpkg.PrintJSON(deps.Stdout, out)
	}
	if _, err := fmt.Fprintf(deps.Stdout, "calendar issues export written: %s\n", out.IssuesFilePath); err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		deps.Stdout,
		"calendar sessions export written: %s\n",
		out.SessionsFilePath,
	)
	return err
}

func runReports(args []string, deps Deps) error {
	if len(args) == 0 || flagspkg.IsHelpArg(args[0]) {
		return fmt.Errorf("usage: crona export reports <list|delete>")
	}
	switch args[0] {
	case "list":
		jsonOut := flagspkg.HasJSON(args[1:])
		var out []sharedtypes.ExportReportFile
		if err := deps.CallKernel(protocol.MethodExportReportsList, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return outputpkg.PrintJSON(deps.Stdout, out)
		}
		if len(out) == 0 {
			_, err := fmt.Fprintln(deps.Stdout, "no exported reports")
			return err
		}
		for _, report := range out {
			if _, err := fmt.Fprintf(deps.Stdout, "%s  [%s] %s\n", outputpkg.ExportReportLabel(report), report.Format, report.Path); err != nil {
				return err
			}
		}
		return nil
	case "delete":
		fs := flagspkg.New("export reports delete")
		path := fs.String("path", "", "")
		jsonOut := fs.Bool("json", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*path) == "" {
			return errors.New("report path is required")
		}
		trimmedPath := strings.TrimSpace(*path)
		if err := deps.CallKernel(protocol.MethodExportReportsDelete, shareddto.ExportReportDeleteRequest{Path: trimmedPath}, nil); err != nil {
			return err
		}
		if *jsonOut {
			return outputpkg.PrintJSON(deps.Stdout, map[string]any{"ok": true, "path": trimmedPath})
		}
		_, err := fmt.Fprintf(deps.Stdout, "report deleted: %s\n", trimmedPath)
		return err
	default:
		return fmt.Errorf("unknown export reports command: %s", args[0])
	}
}

func runReport(args []string, kind sharedtypes.ExportReportKind, deps Deps) error {
	fs := flagspkg.New("export " + string(kind))
	date := fs.String("date", "", "")
	start := fs.String("start", "", "")
	end := fs.String("end", "", "")
	repoID := fs.Int64("repo-id", 0, "")
	streamID := fs.Int64("stream-id", 0, "")
	issueID := fs.Int64("issue-id", 0, "")
	yesterday := fs.Bool("yesterday", false, "")
	week := fs.Bool("week", false, "")
	month := fs.Bool("month", false, "")
	lastXDays := fs.Int("last-x-days", 0, "")
	formatValue := fs.String("format", "", "")
	outputValue := fs.String("output", "", "")
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	input := shareddto.ExportReportRequest{
		Kind:       kind,
		Date:       strings.TrimSpace(*date),
		Start:      strings.TrimSpace(*start),
		End:        strings.TrimSpace(*end),
		Format:     parseExportFormat(*formatValue, kind),
		OutputMode: parseExportOutput(*outputValue, kind),
	}
	if *repoID > 0 {
		input.RepoID = repoID
	}
	if *streamID > 0 {
		input.StreamID = streamID
	}
	if *issueID > 0 {
		input.IssueID = issueID
	}
	if kind == sharedtypes.ExportReportKindGlance {
		if err := applyGlancePeriod(&input, *yesterday, *week, *month, *lastXDays, time.Now()); err != nil {
			return err
		}
		if input.Start != "" && input.End != "" && input.Start != input.End {
			input.Kind = sharedtypes.ExportReportKindSummaryRange
		}
	}
	if err := requireExplicitScope(input); err != nil {
		return err
	}
	if err := applyScopeDefaults(&input, deps); err != nil {
		return err
	}

	var out sharedtypes.ExportReportResult
	if err := deps.CallKernel(exportMethodForKind(kind), input, &out); err != nil {
		return err
	}
	if *jsonOut {
		return outputpkg.PrintJSON(deps.Stdout, out)
	}
	if input.OutputMode == sharedtypes.ExportOutputModeClipboard {
		if err := copyToClipboard(out.Content); err != nil {
			return err
		}
		_, err := fmt.Fprintln(deps.Stdout, "Markdown copied to clipboard")
		return err
	}
	return outputpkg.PrintExportResult(deps.Stdout, out)
}

func requireExplicitScope(input shareddto.ExportReportRequest) error {
	switch input.Kind {
	case sharedtypes.ExportReportKindRepo:
		if input.RepoID == nil {
			return errors.New("repo export requires --repo-id")
		}
	case sharedtypes.ExportReportKindStream:
		if input.StreamID == nil {
			return errors.New("stream export requires --stream-id")
		}
	case sharedtypes.ExportReportKindIssueRollup:
		if input.IssueID == nil {
			return errors.New("issue export requires --issue-id")
		}
	}
	return nil
}

func applyGlancePeriod(input *shareddto.ExportReportRequest, yesterday, week, month bool, lastXDays int, now time.Time) error {
	selected := 0
	for _, active := range []bool{yesterday, week, month, lastXDays > 0} {
		if active {
			selected++
		}
	}
	if selected > 1 || (selected > 0 && (input.Date != "" || input.Start != "" || input.End != "")) {
		return errors.New("choose one summary period")
	}
	if (input.Start == "") != (input.End == "") {
		return errors.New("--start and --end must be used together")
	}
	anchor := now
	switch {
	case yesterday:
		anchor = anchor.AddDate(0, 0, -1)
		input.Date = anchor.Format(time.DateOnly)
	case week:
		offset := (int(anchor.Weekday()) + 6) % 7
		input.Start = anchor.AddDate(0, 0, -offset).Format(time.DateOnly)
		input.End = anchor.Format(time.DateOnly)
	case month:
		input.Start = time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, anchor.Location()).Format(time.DateOnly)
		input.End = anchor.Format(time.DateOnly)
	case lastXDays > 0:
		input.Start = anchor.AddDate(0, 0, -(lastXDays - 1)).Format(time.DateOnly)
		input.End = anchor.Format(time.DateOnly)
	case input.Date == "" && input.Start == "":
		input.Date = anchor.Format(time.DateOnly)
	}
	return nil
}

func copyToClipboard(content string) error {
	commands := [][]string{{"pbcopy"}, {"wl-copy"}, {"xclip", "-selection", "clipboard"}, {"xsel", "--clipboard", "--input"}, {"clip"}}
	for _, candidate := range commands {
		path, err := exec.LookPath(candidate[0])
		if err != nil {
			continue
		}
		cmd := exec.Command(path, candidate[1:]...)
		cmd.Stdin = strings.NewReader(content)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	return errors.New("no supported clipboard command found")
}

func resolveCalendarRepoID(explicit int64, deps Deps) (int64, error) {
	if explicit > 0 {
		return explicit, nil
	}
	var ctxOut sharedtypes.ActiveContext
	if err := deps.CallKernel(protocol.MethodContextGet, nil, &ctxOut); err == nil &&
		ctxOut.RepoID != nil &&
		*ctxOut.RepoID > 0 {
		return *ctxOut.RepoID, nil
	}
	var repos []sharedtypes.Repo
	if err := deps.CallKernel(protocol.MethodRepoList, nil, &repos); err != nil {
		return 0, err
	}
	if len(repos) == 0 {
		return 0, errors.New("calendar export requires at least one repo")
	}
	return repos[0].ID, nil
}

func applyScopeDefaults(input *shareddto.ExportReportRequest, deps Deps) error {
	if input == nil {
		return nil
	}
	contextDeps := contextcmd.Deps{Stdout: deps.Stdout, CallKernel: deps.CallKernel}
	switch input.Kind {
	case sharedtypes.ExportReportKindCSV:
		ctxOut, err := contextcmd.ResolveActive(contextDeps)
		if err != nil {
			return err
		}
		if input.StreamID == nil && ctxOut.StreamID != nil && *ctxOut.StreamID > 0 {
			input.StreamID = ctxOut.StreamID
		} else if input.RepoID == nil && ctxOut.RepoID != nil && *ctxOut.RepoID > 0 {
			input.RepoID = ctxOut.RepoID
		}
	}
	return nil
}

func exportMethodForKind(kind sharedtypes.ExportReportKind) string {
	switch kind {
	case sharedtypes.ExportReportKindGlance, sharedtypes.ExportReportKindSummaryRange:
		return protocol.MethodExportGlance
	case sharedtypes.ExportReportKindWeekly:
		return protocol.MethodExportWeekly
	case sharedtypes.ExportReportKindRepo:
		return protocol.MethodExportRepo
	case sharedtypes.ExportReportKindStream:
		return protocol.MethodExportStream
	case sharedtypes.ExportReportKindIssueRollup:
		return protocol.MethodExportIssueRollup
	case sharedtypes.ExportReportKindCSV:
		return protocol.MethodExportCSV
	default:
		return protocol.MethodExportDaily
	}
}

func parseExportFormat(value string, kind sharedtypes.ExportReportKind) sharedtypes.ExportFormat {
	switch kind {
	case sharedtypes.ExportReportKindCSV:
		return sharedtypes.ExportFormatCSV
	default:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "pdf":
			return sharedtypes.ExportFormatPDF
		default:
			return sharedtypes.ExportFormatMarkdown
		}
	}
}

func parseExportOutput(
	value string,
	kind sharedtypes.ExportReportKind,
) sharedtypes.ExportOutputMode {
	switch kind {
	case sharedtypes.ExportReportKindCSV:
		return sharedtypes.ExportOutputModeFile
	default:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "clipboard":
			return sharedtypes.ExportOutputModeClipboard
		default:
			return sharedtypes.ExportOutputModeFile
		}
	}
}
