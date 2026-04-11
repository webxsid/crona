package testsuite

import (
	"archive/zip"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
)

func TestGenerateSupportBundleCreatesExpectedFilesAndRedactsSensitiveDetails(t *testing.T) {
	baseDir := t.TempDir()
	now := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)
	window := time.Hour

	repoName := "Work"
	streamName := "platform"
	issueTitle := "Secret Issue"
	sessionID := "session-123"
	contextUpdated := "2026-04-02T11:45:00Z"

	input := helperpkg.SupportDiagnosticsInput{
		View:            "support",
		Pane:            "support",
		Width:           100,
		Height:          30,
		DashboardDate:   "2026-04-02",
		RollupStartDate: "2026-03-27",
		RollupEndDate:   "2026-04-02",
		WellbeingDate:   "2026-04-02",
		Context: &api.ActiveContext{
			RepoName:   &repoName,
			StreamName: &streamName,
			IssueTitle: &issueTitle,
			UpdatedAt:  &contextUpdated,
		},
		Timer: &api.TimerState{
			State:       "running",
			SessionID:   &sessionID,
			SegmentType: segmentPtr(sharedtypes.SessionSegmentWork),
		},
		KernelInfo: &api.KernelInfo{
			Endpoint:       filepath.Join(baseDir, "socket", "crona.sock"),
			ScratchDir:     filepath.Join(baseDir, "scratch"),
			ExecutablePath: filepath.Join(baseDir, "bin", "crona-kernel"),
		},
		ExportAssets: &api.ExportAssetStatus{
			ReportsDir:           filepath.Join(baseDir, "reports"),
			ICSDir:               filepath.Join(baseDir, "calendar"),
			PDFRendererPath:      filepath.Join(baseDir, "bin", "weasyprint"),
			PDFRendererName:      "weasyprint",
			PDFRendererAvailable: true,
		},
		TUIPath:    filepath.Join(baseDir, "bin", "crona-tui"),
		KernelPath: filepath.Join(baseDir, "bin", "crona-kernel"),
	}

	ops := []api.Op{{
		ID:        "1",
		Entity:    sharedtypes.OpEntityIssue,
		EntityID:  "issue-1",
		Action:    sharedtypes.OpActionUpdate,
		Timestamp: "2026-04-02T11:50:00Z",
		Payload: map[string]any{
			"title": "Secret Issue",
			"path":  filepath.Join(baseDir, "scratch", "notes.md"),
			"note":  "private note",
		},
	}}
	tuiErrors := []string{"[2026-04-02T11:55:00Z] [ERROR] failed to open " + filepath.Join(baseDir, "scratch", "notes.md")}
	kernelErrors := []string{"[2026-04-02T11:56:00Z] [ERROR] ipc write failed\n  Detail: " + filepath.Join(baseDir, "socket", "crona.sock")}

	path, sizeBytes, err := helperpkg.GenerateSupportBundle(baseDir, now, window, input, ops, tuiErrors, kernelErrors, nil)
	if err != nil {
		t.Fatalf("GenerateSupportBundle returned error: %v", err)
	}
	if sizeBytes <= 0 {
		t.Fatalf("expected non-zero bundle size, got %d", sizeBytes)
	}

	reader, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("zip.OpenReader: %v", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	files := map[string]string{}
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open zip entry %s: %v", file.Name, err)
		}
		body, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read zip entry %s: %v", file.Name, err)
		}
		files[file.Name] = string(body)
	}

	for _, name := range []string{"summary.txt", "recent_ops.json", "recent_errors_tui.log", "recent_errors_kernel.log", "collection_meta.json"} {
		if _, ok := files[name]; !ok {
			t.Fatalf("expected zip entry %s", name)
		}
	}

	if strings.Contains(files["summary.txt"], baseDir) {
		t.Fatalf("expected summary to redact runtime paths, got %q", files["summary.txt"])
	}
	if strings.Contains(files["summary.txt"], "Secret Issue") {
		t.Fatalf("expected summary to redact issue title, got %q", files["summary.txt"])
	}
	if !strings.Contains(files["summary.txt"], "<redacted-endpoint>") {
		t.Fatalf("expected summary to redact endpoint, got %q", files["summary.txt"])
	}

	if strings.Contains(files["recent_ops.json"], "Secret Issue") || strings.Contains(files["recent_ops.json"], "private note") || strings.Contains(files["recent_ops.json"], baseDir) {
		t.Fatalf("expected ops json to redact payload details, got %q", files["recent_ops.json"])
	}
	if strings.Contains(files["recent_errors_tui.log"], baseDir) || strings.Contains(files["recent_errors_kernel.log"], baseDir) {
		t.Fatalf("expected error logs to redact runtime paths")
	}

	var meta helperpkg.SupportCollectionMeta
	if err := json.Unmarshal([]byte(files["collection_meta.json"]), &meta); err != nil {
		t.Fatalf("json.Unmarshal collection_meta: %v", err)
	}
	if meta.RedactionMode != "safe" {
		t.Fatalf("expected redaction mode safe, got %q", meta.RedactionMode)
	}
}

func segmentPtr(value sharedtypes.SessionSegmentType) *sharedtypes.SessionSegmentType {
	return &value
}
