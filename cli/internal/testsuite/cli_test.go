package testsuite

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	completioncmd "crona/cli/internal/command/completion"
	contextcmd "crona/cli/internal/command/context"
	exportcmd "crona/cli/internal/command/export"
	kernelcmd "crona/cli/internal/command/kernel"
	migrationcmd "crona/cli/internal/command/migration"
	sessioncmd "crona/cli/internal/command/session"
	updatecmd "crona/cli/internal/command/update"
	shareddto "crona/shared/dto"
	"crona/shared/localipc"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestKernelAttachJSONUsesEnsureKernel(t *testing.T) {
	var out bytes.Buffer
	want := &sharedtypes.KernelInfo{
		PID:        42,
		Transport:  localipc.TransportWindowsNamedPipe,
		Endpoint:   `\\.\pipe\crona-kernel`,
		Env:        "Prod",
		ScratchDir: `C:\Users\alice\AppData\Local\Crona\scratch`,
	}

	err := kernelcmd.Run([]string{"attach", "--json"}, kernelcmd.Deps{
		Stdout:       &out,
		CallKernel:   func(string, any, any) error { return nil },
		EnsureKernel: func() (*sharedtypes.KernelInfo, error) { return want, nil },
	})
	if err != nil {
		t.Fatalf("kernel attach: %v", err)
	}

	var got sharedtypes.KernelInfo
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal attach output: %v\noutput=%s", err, out.String())
	}
	if got.PID != want.PID || got.Transport != want.Transport || got.Endpoint != want.Endpoint {
		t.Fatalf("unexpected attach output: %+v", got)
	}
}

func TestKernelStatusTextIncludesWindowsTransportAndEndpoint(t *testing.T) {
	var out bytes.Buffer
	err := kernelcmd.Run([]string{"status"}, kernelcmd.Deps{
		Stdout:       &out,
		EnsureKernel: func() (*sharedtypes.KernelInfo, error) { return nil, nil },
		CallKernel: func(method string, params, target any) error {
			info := target.(*sharedtypes.KernelInfo)
			*info = sharedtypes.KernelInfo{
				PID:        7,
				Transport:  localipc.TransportWindowsNamedPipe,
				Endpoint:   `\\.\pipe\crona-kernel-dev`,
				Env:        "Dev",
				StartedAt:  "2026-03-27T00:00:00Z",
				ScratchDir: `C:\Users\alice\AppData\Local\Crona Dev\scratch`,
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("kernel status: %v", err)
	}

	for _, want := range []string{
		"transport: windows_named_pipe",
		`endpoint: \\.\pipe\crona-kernel-dev`,
		"env: Dev",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out.String())
		}
	}
}

func TestContextSwitchRepoJSONUsesSwitchMethod(t *testing.T) {
	var out bytes.Buffer
	var gotMethod string
	err := contextcmd.Run([]string{"switch-repo", "--id", "12", "--json"}, contextcmd.Deps{
		Stdout: &out,
		CallKernel: func(method string, params, target any) error {
			gotMethod = method
			if method != protocol.MethodContextSwitchRepo {
				t.Fatalf("unexpected method: %s", method)
			}
			body, err := json.Marshal(params)
			if err != nil {
				t.Fatalf("marshal params: %v", err)
			}
			var decoded map[string]int64
			if err := json.Unmarshal(body, &decoded); err != nil {
				t.Fatalf("unmarshal params: %v", err)
			}
			if decoded["repoId"] != 12 {
				t.Fatalf("unexpected repoId payload: %#v", decoded)
			}
			ctx := target.(*sharedtypes.ActiveContext)
			repoName := "Work"
			ctx.RepoName = &repoName
			return nil
		},
	})
	if err != nil {
		t.Fatalf("context switch-repo: %v", err)
	}
	if gotMethod != protocol.MethodContextSwitchRepo {
		t.Fatalf("expected %s, got %s", protocol.MethodContextSwitchRepo, gotMethod)
	}
	if !strings.Contains(out.String(), "\"repoName\": \"Work\"") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestKernelWipeDataRequiresForce(t *testing.T) {
	err := kernelcmd.Run([]string{"wipe-data"}, kernelcmd.Deps{
		Stdout:       &bytes.Buffer{},
		EnsureKernel: func() (*sharedtypes.KernelInfo, error) { return nil, nil },
		CallKernel:   func(string, any, any) error { return nil },
	})
	if err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("expected --force error, got %v", err)
	}
}

func TestTimerStartFromContextUsesCheckedOutIssue(t *testing.T) {
	var gotMethods []string
	err := sessioncmd.RunTimer([]string{"start", "--from-context", "--json"}, sessioncmd.Deps{
		Stdout: &bytes.Buffer{},
		CallKernel: func(method string, params, out any) error {
			gotMethods = append(gotMethods, method)
			switch method {
			case protocol.MethodContextGet:
				ctx := out.(*sharedtypes.ActiveContext)
				issueID := int64(77)
				ctx.IssueID = &issueID
				return nil
			case protocol.MethodTimerStart:
				req, ok := params.(shareddto.TimerStartRequest)
				if !ok || req.IssueID == nil || *req.IssueID != 77 || req.IgnoreExistingStashes {
					t.Fatalf("unexpected timer start params: %#v", params)
				}
				timer := out.(*sharedtypes.TimerState)
				timer.State = "running"
				timer.ElapsedSeconds = 12
				return nil
			default:
				t.Fatalf("unexpected method: %s", method)
				return nil
			}
		},
	})
	if err != nil {
		t.Fatalf("timer start --from-context: %v", err)
	}
	if len(gotMethods) != 2 || gotMethods[0] != protocol.MethodContextGet ||
		gotMethods[1] != protocol.MethodTimerStart {
		t.Fatalf("unexpected method order: %+v", gotMethods)
	}
}

func TestIssueStartFromContextUsesCheckedOutIssue(t *testing.T) {
	err := sessioncmd.RunIssue([]string{"start", "--from-context", "--json"}, sessioncmd.Deps{
		Stdout: &bytes.Buffer{},
		CallKernel: func(method string, params, out any) error {
			switch method {
			case protocol.MethodContextGet:
				ctx := out.(*sharedtypes.ActiveContext)
				issueID := int64(91)
				ctx.IssueID = &issueID
				return nil
			case protocol.MethodTimerStart:
				req, ok := params.(shareddto.TimerStartRequest)
				if !ok || req.IssueID == nil || *req.IssueID != 91 || req.IgnoreExistingStashes {
					t.Fatalf("unexpected timer start params: %#v", params)
				}
				timer := out.(*sharedtypes.TimerState)
				timer.State = "running"
				return nil
			default:
				t.Fatalf("unexpected method: %s", method)
				return nil
			}
		},
	})
	if err != nil {
		t.Fatalf("issue start --from-context: %v", err)
	}
}

func TestIssueStartFormatsStashConflictError(t *testing.T) {
	err := sessioncmd.RunIssue([]string{"start", "--id", "91"}, sessioncmd.Deps{
		Stdout: &bytes.Buffer{},
		CallKernel: func(method string, params, out any) error {
			if method != protocol.MethodTimerStart {
				t.Fatalf("unexpected method: %s", method)
			}
			conflict := sharedtypes.StashConflict{
				IssueID: 91,
				Stashes: []sharedtypes.Stash{{ID: "stash-1"}, {ID: "stash-2"}},
			}
			body, marshalErr := json.Marshal(conflict)
			if marshalErr != nil {
				t.Fatalf("marshal conflict: %v", marshalErr)
			}
			return &protocol.RPCError{
				Code:    protocol.ErrorCodeStashConflict,
				Message: "cannot start focus: 2 stashes exist for issue #91",
				Data:    body,
			}
		},
	})
	if err == nil || !strings.Contains(err.Error(), "2 stashes exist for issue #91") {
		t.Fatalf("expected formatted stash conflict error, got %v", err)
	}
}

func TestExportRepoUsesContextRepoWhenMissing(t *testing.T) {
	var out bytes.Buffer
	var gotMethods []string
	err := exportcmd.Run([]string{"repo", "--json"}, exportcmd.Deps{
		Stdout: &out,
		CallKernel: func(method string, params, target any) error {
			gotMethods = append(gotMethods, method)
			switch method {
			case protocol.MethodContextGet:
				ctx := target.(*sharedtypes.ActiveContext)
				repoID := int64(44)
				ctx.RepoID = &repoID
				return nil
			case protocol.MethodExportRepo:
				body, err := json.Marshal(params)
				if err != nil {
					t.Fatalf("marshal export params: %v", err)
				}
				var decoded map[string]any
				if err := json.Unmarshal(body, &decoded); err != nil {
					t.Fatalf("decode export params: %v", err)
				}
				if int(decoded["repoId"].(float64)) != 44 {
					t.Fatalf("unexpected export payload: %#v", decoded)
				}
				result := target.(*sharedtypes.ExportReportResult)
				result.Kind = sharedtypes.ExportReportKindRepo
				result.Format = sharedtypes.ExportFormatMarkdown
				result.OutputMode = sharedtypes.ExportOutputModeFile
				path := "/tmp/repo-report.md"
				result.FilePath = &path
				return nil
			default:
				t.Fatalf("unexpected method: %s", method)
				return nil
			}
		},
	})
	if err != nil {
		t.Fatalf("export repo: %v", err)
	}
	if len(gotMethods) != 2 || gotMethods[0] != protocol.MethodContextGet ||
		gotMethods[1] != protocol.MethodExportRepo {
		t.Fatalf("unexpected method order: %+v", gotMethods)
	}
	if !strings.Contains(out.String(), "\"filePath\": \"/tmp/repo-report.md\"") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestUpdateRejectsDismissCommand(t *testing.T) {
	var out bytes.Buffer
	err := updatecmd.Run([]string{"dismiss"}, updatecmd.Deps{
		Stdout: &out,
	})
	if err == nil {
		t.Fatalf("expected dismiss command to be rejected")
	}
	if !strings.Contains(err.Error(), "unknown update command") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateStatusTextIncludesInstallSourceAndCommand(t *testing.T) {
	var out bytes.Buffer
	err := updatecmd.Run([]string{"status"}, updatecmd.Deps{
		Stdout: &out,
		CallKernel: func(method string, params, target any) error {
			status := target.(*sharedtypes.UpdateStatus)
			status.CurrentVersion = "1.5.1"
			status.LatestVersion = "1.6.0"
			status.InstallSource = sharedtypes.InstallSourceWinget
			status.UpdateCommand = "winget upgrade --id Webxsid.Crona -e"
			status.InstallUnavailableReason = "managed by winget"
			status.UpdateAvailable = true
			status.Enabled = true
			status.PromptEnabled = true
			return nil
		},
	})
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	for _, want := range []string{"current: 1.5.1", "latest: 1.6.0", "install-source: winget", "update-command: winget upgrade --id Webxsid.Crona -e", "install-unavailable: managed by winget"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out.String())
		}
	}
}

func TestCompletionIncludesMigrationCommands(t *testing.T) {
	var out bytes.Buffer
	if err := completioncmd.Run([]string{"zsh"}, "crona", &out); err != nil {
		t.Fatalf("completion command: %v", err)
	}
	rendered := out.String()
	for _, want := range []string{"backup:Back up data", "restore:Restore data", "backup) ;;", "restore) ;;"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected completion output to contain %q, got:\n%s", want, rendered)
		}
	}
}

func TestBackupCommandWritesBackupFile(t *testing.T) {
	runtimeDir := t.TempDir()
	backupDir := t.TempDir()
	source := filepath.Join(runtimeDir, "crona.db")
	if err := os.WriteFile(source, []byte("database-bytes"), 0o600); err != nil {
		t.Fatalf("write source db: %v", err)
	}

	var out bytes.Buffer
	var shutdownCalls int
	err := migrationcmd.RunBackup([]string{}, migrationcmd.Deps{
		Stdout: &out,
		ShutdownKernel: func() error {
			shutdownCalls++
			return nil
		},
		RuntimeBaseDir: func() (string, error) { return runtimeDir, nil },
		BackupBaseDir:  func() (string, error) { return backupDir, nil },
		Now: func() time.Time {
			return time.Date(2026, 6, 16, 6, 20, 10, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("backup command: %v", err)
	}
	if shutdownCalls != 1 {
		t.Fatalf("expected shutdown to be attempted once, got %d", shutdownCalls)
	}
	rendered := out.String()
	if !strings.Contains(rendered, "backup file:") {
		t.Fatalf("expected backup output to include file path, got %q", rendered)
	}
	lines := strings.Split(strings.TrimSpace(rendered), "\n")
	path := strings.TrimSpace(strings.TrimPrefix(lines[len(lines)-1], "backup file:"))
	if path == "" {
		t.Fatalf("expected backup path in output, got %q", rendered)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read backup file: %v", err)
	}
	if string(body) != "database-bytes" {
		t.Fatalf("unexpected backup contents: %q", string(body))
	}
}

func TestRestoreCommandPromptsAndWritesDatabase(t *testing.T) {
	runtimeDir := t.TempDir()
	backupDir := t.TempDir()
	target := filepath.Join(runtimeDir, "crona.db")
	if err := os.WriteFile(target, []byte("old-data"), 0o600); err != nil {
		t.Fatalf("write current db: %v", err)
	}
	source := filepath.Join(backupDir, "crona-db-20260616-062010.db")
	if err := os.WriteFile(source, []byte("restored-data"), 0o600); err != nil {
		t.Fatalf("write backup db: %v", err)
	}

	var out bytes.Buffer
	var shutdownCalls int
	err := migrationcmd.RunRestore([]string{source}, migrationcmd.Deps{
		Stdout: &out,
		Stdin:  strings.NewReader("y\n"),
		ShutdownKernel: func() error {
			shutdownCalls++
			return nil
		},
		RuntimeBaseDir: func() (string, error) { return runtimeDir, nil },
		BackupBaseDir:  func() (string, error) { return backupDir, nil },
		IsInteractive:  func() bool { return true },
	})
	if err != nil {
		t.Fatalf("restore command: %v", err)
	}
	if shutdownCalls != 1 {
		t.Fatalf("expected shutdown to be attempted once, got %d", shutdownCalls)
	}
	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read restored db: %v", err)
	}
	if string(body) != "restored-data" {
		t.Fatalf("unexpected restored contents: %q", string(body))
	}
	if !strings.Contains(out.String(), "restored database:") {
		t.Fatalf("expected restore output to include destination path, got %q", out.String())
	}
}
