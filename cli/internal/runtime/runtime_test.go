package runtime

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"crona/shared/localipc"
	sharedtypes "crona/shared/types"
)

func TestNormalizeKernelInfoPreservesWindowsNamedPipeFields(t *testing.T) {
	info := &sharedtypes.KernelInfo{
		Transport:  localipc.TransportWindowsNamedPipe,
		Endpoint:   `\\.\pipe\crona-daemon`,
		SocketPath: "",
	}

	NormalizeKernelInfo(info)

	if info.Transport != localipc.TransportWindowsNamedPipe {
		t.Fatalf("expected windows transport, got %q", info.Transport)
	}
	if info.Endpoint != `\\.\pipe\crona-daemon` {
		t.Fatalf("expected endpoint to be preserved, got %q", info.Endpoint)
	}
	if info.SocketPath != "" {
		t.Fatalf(
			"expected socket path to stay empty for windows transport, got %q",
			info.SocketPath,
		)
	}
}

func TestKernelLaunchCandidatesIncludeWindowsExeSources(t *testing.T) {
	root := t.TempDir()
	repoRoot := filepath.Join(root, "repo")
	exeDir := filepath.Join(root, "app")
	pathDir := filepath.Join(root, "path")
	if err := os.MkdirAll(filepath.Join(repoRoot, "kernel", "cmd"), 0o755); err != nil {
		t.Fatalf("mkdir repo kernel cmd: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, "bin"), 0o755); err != nil {
		t.Fatalf("mkdir repo bin: %v", err)
	}
	if err := os.MkdirAll(exeDir, 0o755); err != nil {
		t.Fatalf("mkdir exe dir: %v", err)
	}
	if err := os.MkdirAll(pathDir, 0o755); err != nil {
		t.Fatalf("mkdir path dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "go.work"), []byte("go 1.26.1\n"), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "kernel", "cmd", "crona-kernel"), []byte("stub"), 0o644); err != nil {
		t.Fatalf("write repo daemon marker: %v", err)
	}

	sibling := filepath.Join(exeDir, "crona-daemon.exe")
	pathCandidate := filepath.Join(pathDir, "crona-daemon.exe")
	repoBin := filepath.Join(repoRoot, "bin", "crona-daemon.exe")
	for _, file := range []string{sibling, pathCandidate, repoBin} {
		if err := os.WriteFile(file, []byte("stub"), 0o755); err != nil {
			t.Fatalf("write candidate %s: %v", file, err)
		}
	}

	got := KernelLaunchCandidates(Deps{
		KernelBinary: func() string { return "crona-daemon.exe" },
		OSExecutable: func() (string, error) { return filepath.Join(exeDir, "crona.exe"), nil },
		OSGetwd:      func() (string, error) { return repoRoot, nil },
		ExecLookPath: func(file string) (string, error) {
			switch file {
			case "crona-daemon.exe":
				return pathCandidate, nil
			case "go":
				return filepath.Join(pathDir, "go"), nil
			default:
				return "", os.ErrNotExist
			}
		},
		OSStat: os.Stat,
	})
	if len(got) < 3 {
		t.Fatalf("expected multiple windows candidates, got %+v", got)
	}

	names := []string{}
	cmds := []string{}
	for _, candidate := range got {
		names = append(names, candidate.Name)
		cmds = append(cmds, candidate.Cmd)
	}

	for _, want := range []string{
		"sibling crona-daemon.exe",
		"PATH crona-daemon.exe",
		"repo bin crona-daemon.exe",
	} {
		if !slices.Contains(names, want) {
			t.Fatalf("expected candidate %q in %+v", want, names)
		}
	}

	for _, want := range []string{sibling, pathCandidate, repoBin} {
		if !slices.Contains(cmds, want) {
			t.Fatalf("expected command %q in %+v", want, cmds)
		}
	}
}
