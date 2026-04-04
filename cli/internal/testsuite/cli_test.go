package testsuite

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	contextcmd "crona/cli/internal/command/context"
	exportcmd "crona/cli/internal/command/export"
	kernelcmd "crona/cli/internal/command/kernel"
	sessioncmd "crona/cli/internal/command/session"
	updatecmd "crona/cli/internal/command/update"
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
				req, ok := params.(struct {
					IssueID *int64 `json:"issueId,omitempty"`
				})
				if !ok || req.IssueID == nil || *req.IssueID != 77 {
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
	if len(gotMethods) != 2 || gotMethods[0] != protocol.MethodContextGet || gotMethods[1] != protocol.MethodTimerStart {
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
				req, ok := params.(struct {
					IssueID *int64 `json:"issueId,omitempty"`
				})
				if !ok || req.IssueID == nil || *req.IssueID != 91 {
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
	if len(gotMethods) != 2 || gotMethods[0] != protocol.MethodContextGet || gotMethods[1] != protocol.MethodExportRepo {
		t.Fatalf("unexpected method order: %+v", gotMethods)
	}
	if !strings.Contains(out.String(), "\"filePath\": \"/tmp/repo-report.md\"") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestUpdateDismissTextHandlesEmptyDismissedVersion(t *testing.T) {
	var out bytes.Buffer
	err := updatecmd.Run([]string{"dismiss"}, updatecmd.Deps{
		Stdout: &out,
		CallKernel: func(method string, params, target any) error {
			status := target.(*sharedtypes.UpdateStatus)
			status.CurrentVersion = "0.4.0"
			return nil
		},
	})
	if err != nil {
		t.Fatalf("update dismiss: %v", err)
	}
	if !strings.Contains(out.String(), "no update dismissed") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}
