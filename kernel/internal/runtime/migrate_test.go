package runtime

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"crona/shared/config"
	sharedtypes "crona/shared/types"
)

func TestMigrateLegacyBaseDirMovesLegacyDir(t *testing.T) {
	skipIfUnsupportedUnix(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv(config.EnvVarRuntimeDir, "")
	if runtime.GOOS == "linux" {
		t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".xdg"))
	}

	legacy, err := config.LegacyRuntimeBaseDirForMode(config.ModeProd)
	if err != nil {
		t.Fatalf("LegacyRuntimeBaseDirForMode: %v", err)
	}
	target, err := config.RuntimeBaseDirForMode(config.ModeProd)
	if err != nil {
		t.Fatalf("RuntimeBaseDirForMode: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(legacy, "assets"), 0o700); err != nil {
		t.Fatalf("MkdirAll legacy: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "crona.db"), []byte("db"), 0o600); err != nil {
		t.Fatalf("WriteFile legacy db: %v", err)
	}

	if err := MigrateLegacyBaseDir(config.ModeProd); err != nil {
		t.Fatalf("MigrateLegacyBaseDir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "crona.db")); err != nil {
		t.Fatalf("expected migrated db at target: %v", err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("expected legacy dir to be removed, got err=%v", err)
	}
}

func TestMigrateLegacyBaseDirSkipsOverride(t *testing.T) {
	skipIfUnsupportedUnix(t)

	home := t.TempDir()
	override := filepath.Join(t.TempDir(), "override-crona")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv(config.EnvVarRuntimeDir, override)

	legacy, err := config.LegacyRuntimeBaseDirForMode(config.ModeProd)
	if err != nil {
		t.Fatalf("LegacyRuntimeBaseDirForMode: %v", err)
	}
	if err := os.MkdirAll(legacy, 0o700); err != nil {
		t.Fatalf("MkdirAll legacy: %v", err)
	}

	if err := MigrateLegacyBaseDir(config.ModeProd); err != nil {
		t.Fatalf("MigrateLegacyBaseDir: %v", err)
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Fatalf("expected legacy dir to remain when override is set: %v", err)
	}
	if _, err := os.Stat(override); !os.IsNotExist(err) {
		t.Fatalf("expected override dir to remain untouched, got err=%v", err)
	}
}

func TestMigrateLegacyBaseDirMergesIntoBootstrapTarget(t *testing.T) {
	skipIfUnsupportedUnix(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv(config.EnvVarRuntimeDir, "")
	if runtime.GOOS == "linux" {
		t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".xdg"))
	}

	legacy, err := config.LegacyRuntimeBaseDirForMode(config.ModeProd)
	if err != nil {
		t.Fatalf("LegacyRuntimeBaseDirForMode: %v", err)
	}
	target, err := config.RuntimeBaseDirForMode(config.ModeProd)
	if err != nil {
		t.Fatalf("RuntimeBaseDirForMode: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(legacy, "assets", "user"), 0o700); err != nil {
		t.Fatalf("MkdirAll legacy: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "crona.db"), []byte("db"), 0o600); err != nil {
		t.Fatalf("WriteFile legacy db: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "assets", "user", "custom.txt"), []byte("user-asset"), 0o600); err != nil {
		t.Fatalf("WriteFile legacy asset: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(target, "logs", "tui"), 0o700); err != nil {
		t.Fatalf("MkdirAll target: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(target, "assets", "bundled"), 0o700); err != nil {
		t.Fatalf("MkdirAll target bundled assets: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "assets", "bundled", "fresh.txt"), []byte("bundled"), 0o600); err != nil {
		t.Fatalf("WriteFile target asset: %v", err)
	}

	if err := MigrateLegacyBaseDir(config.ModeProd); err != nil {
		t.Fatalf("MigrateLegacyBaseDir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "crona.db")); err != nil {
		t.Fatalf("expected migrated db at target: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "assets", "user", "custom.txt")); err != nil {
		t.Fatalf("expected legacy user asset at target: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "assets", "bundled", "fresh.txt")); err != nil {
		t.Fatalf("expected existing target asset to remain: %v", err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("expected legacy dir to be removed after merge, got err=%v", err)
	}
}

func TestMigrateLegacyBaseDirReturnsErrorForRunningKernel(t *testing.T) {
	skipIfUnsupportedUnix(t)
	if _, err := exec.LookPath("ps"); err != nil {
		t.Skip("ps unavailable")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv(config.EnvVarRuntimeDir, "")
	if runtime.GOOS == "linux" {
		t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".xdg"))
	}

	legacy, err := config.LegacyRuntimeBaseDirForMode(config.ModeProd)
	if err != nil {
		t.Fatalf("LegacyRuntimeBaseDirForMode: %v", err)
	}
	if err := os.MkdirAll(legacy, 0o700); err != nil {
		t.Fatalf("MkdirAll legacy: %v", err)
	}

	kernelPath := filepath.Join(t.TempDir(), "crona-kernel")
	if err := os.WriteFile(kernelPath, []byte("#!/bin/sh\nsleep 30\n"), 0o755); err != nil {
		t.Fatalf("WriteFile crona-kernel stub: %v", err)
	}
	cmd := exec.Command(kernelPath)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start crona-kernel stub: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}()

	time.Sleep(100 * time.Millisecond)

	info := sharedtypes.KernelInfo{PID: cmd.Process.Pid}
	body, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Marshal kernel info: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "kernel.json"), body, 0o600); err != nil {
		t.Fatalf("WriteFile kernel.json: %v", err)
	}

	err = MigrateLegacyBaseDir(config.ModeProd)
	if err != nil && strings.Contains(err.Error(), "operation not permitted") {
		t.Skipf("process inspection unavailable in sandbox: %v", err)
	}
	if err == nil || !strings.Contains(err.Error(), "stop Crona and retry") {
		t.Fatalf("expected running-kernel migration error, got %v", err)
	}
}

func skipIfUnsupportedUnix(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("unix-only runtime migration")
	}
}
