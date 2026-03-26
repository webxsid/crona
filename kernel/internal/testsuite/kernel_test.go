package testsuite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/events"
	"crona/kernel/internal/export"
	"crona/kernel/internal/runtime"
	"crona/kernel/internal/sessionnotes"
	"crona/kernel/internal/store"
	"crona/kernel/internal/store/repositories"
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

func TestHabitCascadeDeleteAndRestoreByStreamAndRepo(t *testing.T) {
	ctx := context.Background()
	db := openTestStore(t)
	now := time.Now().UTC().Format(time.RFC3339)
	userID := "local"

	registry := store.NewRegistry(db.DB())
	repos := registry.Repos
	streams := registry.Streams
	habits := registry.Habits

	repo := mustCreateRepo(t, ctx, repos, userID, now, 1, "Work")
	stream := mustCreateStream(t, ctx, streams, userID, now, 1, repo.ID, "app")
	habit := mustCreateHabit(t, ctx, habits, userID, now, 1, stream.ID, "Inbox Zero")

	if err := habits.SoftDeleteByStream(ctx, stream.ID, userID, now); err != nil {
		t.Fatalf("soft delete by stream: %v", err)
	}
	got, err := habits.GetByID(ctx, habit.ID, userID)
	if err != nil {
		t.Fatalf("get deleted habit: %v", err)
	}
	if got != nil {
		t.Fatalf("expected habit deleted by stream cascade")
	}
	if err := habits.RestoreDeletedByStream(ctx, stream.ID, userID, now); err != nil {
		t.Fatalf("restore by stream: %v", err)
	}
	got, err = habits.GetByID(ctx, habit.ID, userID)
	if err != nil {
		t.Fatalf("get restored habit: %v", err)
	}
	if got == nil || got.Name != "Inbox Zero" {
		t.Fatalf("expected habit restored by stream, got %+v", got)
	}

	if err := habits.SoftDeleteByRepo(ctx, repo.ID, userID, now); err != nil {
		t.Fatalf("soft delete by repo: %v", err)
	}
	got, err = habits.GetByID(ctx, habit.ID, userID)
	if err != nil {
		t.Fatalf("get repo-deleted habit: %v", err)
	}
	if got != nil {
		t.Fatalf("expected habit deleted by repo cascade")
	}
	if err := habits.RestoreDeletedByRepo(ctx, repo.ID, userID, now); err != nil {
		t.Fatalf("restore by repo: %v", err)
	}
	got, err = habits.GetByID(ctx, habit.ID, userID)
	if err != nil {
		t.Fatalf("get repo-restored habit: %v", err)
	}
	if got == nil || got.Name != "Inbox Zero" {
		t.Fatalf("expected habit restored by repo, got %+v", got)
	}
}

func openTestStore(t *testing.T) *store.Store {
	t.Helper()

	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "crona.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if err := store.InitSchema(context.Background(), db.DB()); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	return db
}

func newTestCoreContext(t *testing.T, now func() string) (*core.Context, runtime.Paths) {
	t.Helper()

	base := t.TempDir()
	paths := runtime.Paths{
		BaseDir:          base,
		AssetsDir:        filepath.Join(base, "assets"),
		BundledAssetsDir: filepath.Join(base, "assets", "bundled"),
		UserAssetsDir:    filepath.Join(base, "assets", "user"),
		ReportsDir:       filepath.Join(base, "reports"),
		ICSDir:           filepath.Join(base, "calendar"),
		LogsDir:          filepath.Join(base, "logs"),
		CurrentLogDir:    filepath.Join(base, "logs", "today"),
		ScratchDir:       filepath.Join(base, "scratch"),
	}
	if err := runtime.EnsurePaths(paths); err != nil {
		t.Fatalf("ensure paths: %v", err)
	}
	db, err := store.Open(filepath.Join(base, "crona.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if err := store.InitSchema(context.Background(), db.DB()); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	coreCtx := core.NewContext(db, store.NewRegistry(db.DB()), "local", "test-device", paths.ScratchDir, now, events.NewBus())
	if err := coreCtx.InitDefaults(context.Background()); err != nil {
		t.Fatalf("init defaults: %v", err)
	}
	return coreCtx, paths
}

func mustCreateRepo(t *testing.T, ctx context.Context, repos *repositories.RepoRepository, userID, now string, id int64, name string) sharedtypes.Repo {
	t.Helper()
	repo, err := repos.Create(ctx, sharedtypes.Repo{ID: id, Name: name}, userID, now)
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	return repo
}

func mustCreateStream(t *testing.T, ctx context.Context, streams *repositories.StreamRepository, userID, now string, id, repoID int64, name string) sharedtypes.Stream {
	t.Helper()
	stream, err := streams.Create(ctx, sharedtypes.Stream{ID: id, RepoID: repoID, Name: name, Visibility: sharedtypes.StreamVisibilityPersonal}, userID, now)
	if err != nil {
		t.Fatalf("create stream: %v", err)
	}
	return stream
}

func mustCreateHabit(t *testing.T, ctx context.Context, habits *repositories.HabitRepository, userID, now string, id, streamID int64, name string) sharedtypes.Habit {
	t.Helper()
	habit, err := habits.Create(ctx, sharedtypes.Habit{ID: id, StreamID: streamID, Name: name, ScheduleType: sharedtypes.HabitScheduleDaily, Active: true}, userID, now)
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}
	return habit
}

func TestExportReportsListIncludesKindScopeAndDateMetadata(t *testing.T) {
	base := t.TempDir()
	paths := runtime.Paths{
		BaseDir:          base,
		AssetsDir:        filepath.Join(base, "assets"),
		BundledAssetsDir: filepath.Join(base, "assets", "bundled"),
		UserAssetsDir:    filepath.Join(base, "assets", "user"),
		ReportsDir:       filepath.Join(base, "reports"),
		LogsDir:          filepath.Join(base, "logs"),
		CurrentLogDir:    filepath.Join(base, "logs", "today"),
		ScratchDir:       filepath.Join(base, "scratch"),
	}
	if err := runtime.EnsurePaths(paths); err != nil {
		t.Fatalf("ensure paths: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(paths.BundledAssetsDir, "export", "daily"), 0o700); err != nil {
		t.Fatalf("mkdir export assets: %v", err)
	}
	if err := export.WriteFileForTesting(filepath.Join(paths.BundledAssetsDir, "export", "daily", "report.default.hbs"), []byte("{{date}}")); err != nil {
		t.Fatalf("write markdown template: %v", err)
	}
	if err := export.WriteFileForTesting(filepath.Join(paths.BundledAssetsDir, "export", "daily", "report.pdf.default.hbs"), []byte("{{date}}")); err != nil {
		t.Fatalf("write pdf template: %v", err)
	}
	if err := export.WriteFileForTesting(filepath.Join(paths.BundledAssetsDir, "export", "daily", "report.variables.md"), []byte("vars")); err != nil {
		t.Fatalf("write variable docs: %v", err)
	}

	repoSpec := export.ReportWriteSpecForTesting(sharedtypes.ExportReportKindRepo, "Repo Report", "Work", "", "2026-03-17", "2026-03-23", sharedtypes.ExportFormatMarkdown, "repo-work")
	if _, err := export.WriteReport(paths, repoSpec, []byte("# Repo Report")); err != nil {
		t.Fatalf("write repo report: %v", err)
	}
	csvSpec := export.ReportWriteSpecForTesting(sharedtypes.ExportReportKindCSV, "CSV Export", "", "", "2026-03-17", "2026-03-23", sharedtypes.ExportFormatCSV, "sessions")
	if _, err := export.WriteReport(paths, csvSpec, []byte("h1,h2\nv1,v2")); err != nil {
		t.Fatalf("write csv report: %v", err)
	}

	reports, err := export.ListReports(paths)
	if err != nil {
		t.Fatalf("list reports: %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}

	var foundRepo, foundCSV bool
	for _, report := range reports {
		switch report.Kind {
		case sharedtypes.ExportReportKindRepo:
			foundRepo = true
			if report.ScopeLabel != "Work" {
				t.Fatalf("expected repo scope label Work, got %q", report.ScopeLabel)
			}
			if report.DateLabel != "2026-03-17 to 2026-03-23" {
				t.Fatalf("unexpected repo date label %q", report.DateLabel)
			}
		case sharedtypes.ExportReportKindCSV:
			foundCSV = true
			if report.Format != string(sharedtypes.ExportFormatCSV) {
				t.Fatalf("expected csv format, got %q", report.Format)
			}
		}
	}
	if !foundRepo || !foundCSV {
		t.Fatalf("expected repo and csv reports, got %+v", reports)
	}
}

func TestCalendarICSFilesWriteOutsideReportsDirectory(t *testing.T) {
	base := t.TempDir()
	paths := runtime.Paths{
		BaseDir:    base,
		ReportsDir: filepath.Join(base, "reports"),
		ICSDir:     filepath.Join(base, "calendar"),
	}
	if err := os.MkdirAll(paths.ReportsDir, 0o700); err != nil {
		t.Fatalf("mkdir reports dir: %v", err)
	}
	if err := os.MkdirAll(paths.ICSDir, 0o700); err != nil {
		t.Fatalf("mkdir ics dir: %v", err)
	}
	spec := export.ReportWriteSpecForTesting(sharedtypes.ExportReportKindCalendar, "Calendar Export", "Work / app", "", "2026-03-17", "2026-03-23", sharedtypes.ExportFormatICS, "calendar-2026-03-17-to-2026-03-23")
	filePath, err := export.WriteReport(paths, spec, []byte("BEGIN:VCALENDAR\r\nEND:VCALENDAR"))
	if err != nil {
		t.Fatalf("write calendar report: %v", err)
	}
	if !strings.HasPrefix(filePath, paths.ICSDir) {
		t.Fatalf("expected calendar export in ics dir, got %q", filePath)
	}

	reports, err := export.ListReports(paths)
	if err != nil {
		t.Fatalf("list reports: %v", err)
	}
	if len(reports) != 0 {
		t.Fatalf("expected calendar exports excluded from reports list, got %d entries", len(reports))
	}
}

func TestGenerateCalendarExportWritesRepoScopedIssuesAndSessionsFiles(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-03-20T09:00:00Z"
	coreCtx, paths := newTestCoreContext(t, func() string { return currentNow })

	workRepo, err := corecommands.CreateRepo(ctx, coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{Name: "Work"})
	if err != nil {
		t.Fatalf("create work repo: %v", err)
	}
	personalRepo, err := corecommands.CreateRepo(ctx, coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{Name: "Personal"})
	if err != nil {
		t.Fatalf("create personal repo: %v", err)
	}
	appStream, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: workRepo.ID, Name: "app"})
	if err != nil {
		t.Fatalf("create app stream: %v", err)
	}
	opsStream, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: workRepo.ID, Name: "ops"})
	if err != nil {
		t.Fatalf("create ops stream: %v", err)
	}
	homeStream, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: personalRepo.ID, Name: "home"})
	if err != nil {
		t.Fatalf("create home stream: %v", err)
	}

	estimate := 60
	workDesc := "Ship the keyboard-first flow"
	workNotes := "Needs design review"
	plannedDate := "2026-03-25"
	plannedIssue, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{
		StreamID:        appStream.ID,
		Title:           "Keyboard-first command palette",
		Description:     &workDesc,
		EstimateMinutes: &estimate,
		Notes:           &workNotes,
		TodoForDate:     &plannedDate,
	})
	if err != nil {
		t.Fatalf("create planned issue: %v", err)
	}
	backlogIssue, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: opsStream.ID, Title: "Backlog issue without date"})
	if err != nil {
		t.Fatalf("create backlog issue: %v", err)
	}
	otherDate := "2026-03-26"
	otherRepoIssue, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: homeStream.ID, Title: "Personal task", TodoForDate: &otherDate})
	if err != nil {
		t.Fatalf("create personal issue: %v", err)
	}

	currentNow = "2026-03-20T10:00:00Z"
	if _, err := corecommands.StartSession(ctx, coreCtx, plannedIssue.ID); err != nil {
		t.Fatalf("start work session: %v", err)
	}
	currentNow = "2026-03-20T11:30:00Z"
	if _, err := corecommands.StopSession(ctx, coreCtx, corecommands.SessionEndInput{WorkedOn: strPtr("Built export flow")}); err != nil {
		t.Fatalf("stop work session: %v", err)
	}

	currentNow = "2026-03-21T08:00:00Z"
	if _, err := corecommands.StartSession(ctx, coreCtx, otherRepoIssue.ID); err != nil {
		t.Fatalf("start personal session: %v", err)
	}
	currentNow = "2026-03-21T09:00:00Z"
	if _, err := corecommands.StopSession(ctx, coreCtx, corecommands.SessionEndInput{WorkedOn: strPtr("Personal work")}); err != nil {
		t.Fatalf("stop personal session: %v", err)
	}

	result, err := export.GenerateCalendarExport(ctx, coreCtx, paths, shareddto.ExportCalendarRequest{RepoID: workRepo.ID})
	if err != nil {
		t.Fatalf("generate calendar export: %v", err)
	}
	if !strings.HasSuffix(result.IssuesFilePath, filepath.Join("calendar", "1-work", "issues.ics")) {
		t.Fatalf("unexpected issues path %q", result.IssuesFilePath)
	}
	if !strings.HasSuffix(result.SessionsFilePath, filepath.Join("calendar", "1-work", "sessions.ics")) {
		t.Fatalf("unexpected sessions path %q", result.SessionsFilePath)
	}

	issuesBody, err := os.ReadFile(result.IssuesFilePath)
	if err != nil {
		t.Fatalf("read issues.ics: %v", err)
	}
	issuesText := string(issuesBody)
	if !strings.Contains(issuesText, "SUMMARY:Keyboard-first command palette") {
		t.Fatalf("expected planned issue in issues calendar")
	}
	if !strings.Contains(issuesText, "DTSTART;VALUE=DATE:20260325") || !strings.Contains(issuesText, "DTEND;VALUE=DATE:20260326") {
		t.Fatalf("expected all-day issue event in issues calendar, got:\n%s", issuesText)
	}
	if strings.Contains(issuesText, backlogIssue.Title) {
		t.Fatalf("did not expect issue without todo date in issues calendar")
	}
	if strings.Contains(issuesText, otherRepoIssue.Title) {
		t.Fatalf("did not expect other repo issue in issues calendar")
	}

	sessionsBody, err := os.ReadFile(result.SessionsFilePath)
	if err != nil {
		t.Fatalf("read sessions.ics: %v", err)
	}
	sessionsText := string(sessionsBody)
	if !strings.Contains(sessionsText, "SUMMARY:Keyboard-first command palette") {
		t.Fatalf("expected work session in sessions calendar")
	}
	if !strings.Contains(sessionsText, "DTSTART:20260320T100000Z") || !strings.Contains(sessionsText, "DTEND:20260320T113000Z") {
		t.Fatalf("expected timed session event in sessions calendar, got:\n%s", sessionsText)
	}
	if strings.Contains(sessionsText, otherRepoIssue.Title) {
		t.Fatalf("did not expect other repo session in sessions calendar")
	}
}

func TestExportReportsDirNormalizesLegacyDailyPathToReportsRoot(t *testing.T) {
	base := t.TempDir()
	paths := runtime.Paths{
		BaseDir:    base,
		ReportsDir: filepath.Join(base, "reports"),
		ICSDir:     filepath.Join(base, "calendar"),
	}
	got, err := export.ResolveReportsDirForTesting(paths, filepath.Join(base, "reports", "daily"))
	if err != nil {
		t.Fatalf("resolve reports dir: %v", err)
	}
	if got != filepath.Join(base, "reports") {
		t.Fatalf("expected legacy reports/daily to normalize to reports root, got %q", got)
	}
}

func TestExportICSDirNormalizesTildePath(t *testing.T) {
	base := t.TempDir()
	paths := runtime.Paths{
		BaseDir: base,
		ICSDir:  filepath.Join(base, "calendar"),
	}
	got, err := export.ResolveICSDirForTesting(paths, "calendar-exports")
	if err != nil {
		t.Fatalf("resolve ics dir: %v", err)
	}
	if got != filepath.Join(base, "calendar-exports") {
		t.Fatalf("expected relative ics dir under base, got %q", got)
	}
}

func TestEnsureAssetsMigratesLegacyDailyUserTemplateToNestedPath(t *testing.T) {
	base := t.TempDir()
	paths := runtime.Paths{
		BaseDir:          base,
		AssetsDir:        filepath.Join(base, "assets"),
		BundledAssetsDir: filepath.Join(base, "assets", "bundled"),
		UserAssetsDir:    filepath.Join(base, "assets", "user"),
		ReportsDir:       filepath.Join(base, "reports"),
		ICSDir:           filepath.Join(base, "calendar"),
		LogsDir:          filepath.Join(base, "logs"),
		CurrentLogDir:    filepath.Join(base, "logs", "today"),
		ScratchDir:       filepath.Join(base, "scratch"),
	}
	if err := runtime.EnsurePaths(paths); err != nil {
		t.Fatalf("ensure paths: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(paths.UserAssetsDir, "export"), 0o700); err != nil {
		t.Fatalf("mkdir user export dir: %v", err)
	}
	legacyPath := filepath.Join(paths.UserAssetsDir, "export", "daily-report.user.hbs")
	if err := export.WriteFileForTesting(legacyPath, []byte("legacy-template")); err != nil {
		t.Fatalf("write legacy user template: %v", err)
	}

	status, err := export.EnsureAssets(paths)
	if err != nil {
		t.Fatalf("ensure assets: %v", err)
	}
	if status.TemplatePath != filepath.Join(paths.UserAssetsDir, "export", "daily", "report.hbs") {
		t.Fatalf("expected nested template path, got %q", status.TemplatePath)
	}
	if status.ICSDir != paths.ICSDir {
		t.Fatalf("expected default ics dir %q, got %q", paths.ICSDir, status.ICSDir)
	}
	body, err := os.ReadFile(status.TemplatePath)
	if err != nil {
		t.Fatalf("read migrated template: %v", err)
	}
	if string(body) != "legacy-template" {
		t.Fatalf("expected legacy template content to migrate, got %q", string(body))
	}
}

func TestDeleteReportRemovesReportAndMetadataWithinReportsDir(t *testing.T) {
	base := t.TempDir()
	paths := runtime.Paths{
		BaseDir:    base,
		ReportsDir: filepath.Join(base, "reports"),
	}
	if err := os.MkdirAll(paths.ReportsDir, 0o700); err != nil {
		t.Fatalf("mkdir reports dir: %v", err)
	}
	reportPath := filepath.Join(paths.ReportsDir, "daily-2026-03-19.md")
	if err := export.WriteFileForTesting(reportPath, []byte("# report")); err != nil {
		t.Fatalf("write report: %v", err)
	}
	if err := export.WriteFileForTesting(reportPath+".meta.json", []byte(`{"kind":"daily"}`)); err != nil {
		t.Fatalf("write report metadata: %v", err)
	}

	if err := export.DeleteReport(paths, reportPath); err != nil {
		t.Fatalf("delete report: %v", err)
	}
	if _, err := os.Stat(reportPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected report removed, got %v", err)
	}
	if _, err := os.Stat(reportPath + ".meta.json"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected report metadata removed, got %v", err)
	}
}

func TestDeleteReportRejectsPathOutsideReportsDir(t *testing.T) {
	base := t.TempDir()
	paths := runtime.Paths{
		BaseDir:    base,
		ReportsDir: filepath.Join(base, "reports"),
	}
	err := export.DeleteReport(paths, filepath.Join(base, "elsewhere", "bad.md"))
	if err == nil {
		t.Fatalf("expected outside-path delete to fail")
	}
}

func TestDetailedIssueGroupRenderingIncludesDescriptionsIssueNotesAndSessionNotes(t *testing.T) {
	description := "Detailed issue description"
	notes := "Issue notes body"
	rawNotes := sessionnotes.Serialize(sharedtypes.ParsedSessionNotes{
		sharedtypes.SessionNoteSectionCommit:  "feat: report depth",
		sharedtypes.SessionNoteSectionWork:    "Worked on report detail sections",
		sharedtypes.SessionNoteSectionNotes:   "Need to verify exported markdown",
		sharedtypes.SessionNoteSectionContext: "Repo ID: 1",
	})
	rendered := strings.Join(export.RenderDetailedIssueGroupForTesting(sharedtypes.IssueWithMeta{
		Issue: sharedtypes.Issue{
			ID:              42,
			Title:           "Expand export reports",
			Status:          sharedtypes.IssueStatusInProgress,
			Description:     &description,
			Notes:           &notes,
			EstimateMinutes: intPtr(90),
		},
		RepoName:   "Work",
		StreamName: "app",
	}, []sharedtypes.SessionHistoryEntry{
		{
			Session: sharedtypes.Session{
				ID:              "session-1",
				IssueID:         42,
				StartTime:       "2026-03-19T10:00:00Z",
				EndTime:         strPtr("2026-03-19T11:00:00Z"),
				DurationSeconds: intPtr(3600),
				Notes:           &rawNotes,
			},
			ParsedNotes: sessionnotes.Parse(&rawNotes),
		},
	}), "\n")

	for _, want := range []string{
		"Description",
		"Detailed issue description",
		"Issue Notes",
		"Issue notes body",
		"Sessions",
		"Commit: feat: report depth",
		"Work: Worked on report detail sections",
		"Notes: Need to verify exported markdown",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected rendered detailed issue group to contain %q, got %q", want, rendered)
		}
	}
}

func intPtr(v int) *int {
	return &v
}

func strPtr(v string) *string {
	return &v
}
