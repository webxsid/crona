package testsuite

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"crona/kernel/internal/export"
	"crona/kernel/internal/runtime"
	"crona/kernel/internal/sessionnotes"
	"crona/kernel/internal/store"
	sharedtypes "crona/shared/types"
)

func TestHabitCascadeDeleteAndRestoreByStreamAndRepo(t *testing.T) {
	ctx := context.Background()
	db := openTestStore(t)
	now := time.Now().UTC().Format(time.RFC3339)
	userID := "local"

	repos := store.NewRepoRepository(db.DB())
	streams := store.NewStreamRepository(db.DB())
	habits := store.NewHabitRepository(db.DB())

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

func mustCreateRepo(t *testing.T, ctx context.Context, repos *store.RepoRepository, userID, now string, id int64, name string) sharedtypes.Repo {
	t.Helper()
	repo, err := repos.Create(ctx, sharedtypes.Repo{ID: id, Name: name}, userID, now)
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	return repo
}

func mustCreateStream(t *testing.T, ctx context.Context, streams *store.StreamRepository, userID, now string, id, repoID int64, name string) sharedtypes.Stream {
	t.Helper()
	stream, err := streams.Create(ctx, sharedtypes.Stream{ID: id, RepoID: repoID, Name: name, Visibility: sharedtypes.StreamVisibilityPersonal}, userID, now)
	if err != nil {
		t.Fatalf("create stream: %v", err)
	}
	return stream
}

func mustCreateHabit(t *testing.T, ctx context.Context, habits *store.HabitRepository, userID, now string, id, streamID int64, name string) sharedtypes.Habit {
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
	if err := os.MkdirAll(filepath.Join(paths.BundledAssetsDir, "export"), 0o700); err != nil {
		t.Fatalf("mkdir export assets: %v", err)
	}
	if err := export.WriteFileForTesting(filepath.Join(paths.BundledAssetsDir, "export", "daily-report.default.hbs"), []byte("{{date}}")); err != nil {
		t.Fatalf("write markdown template: %v", err)
	}
	if err := export.WriteFileForTesting(filepath.Join(paths.BundledAssetsDir, "export", "daily-report.pdf.default.hbs"), []byte("{{date}}")); err != nil {
		t.Fatalf("write pdf template: %v", err)
	}
	if err := export.WriteFileForTesting(filepath.Join(paths.BundledAssetsDir, "export", "daily-report.variables.md"), []byte("vars")); err != nil {
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

func TestExportReportsDirNormalizesLegacyDailyPathToReportsRoot(t *testing.T) {
	base := t.TempDir()
	paths := runtime.Paths{
		BaseDir:    base,
		ReportsDir: filepath.Join(base, "reports"),
	}
	got, err := export.ResolveReportsDirForTesting(paths, filepath.Join(base, "reports", "daily"))
	if err != nil {
		t.Fatalf("resolve reports dir: %v", err)
	}
	if got != filepath.Join(base, "reports") {
		t.Fatalf("expected legacy reports/daily to normalize to reports root, got %q", got)
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
