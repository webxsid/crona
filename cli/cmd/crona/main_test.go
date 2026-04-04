package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRunNoArgsLaunchesTUI(t *testing.T) {
	orig := launchTUIFn
	defer func() { launchTUIFn = orig }()

	called := false
	launchTUIFn = func() error {
		called = true
		return nil
	}

	if err := run(nil); err != nil {
		t.Fatalf("run no args: %v", err)
	}
	if !called {
		t.Fatalf("expected no-arg run to launch TUI")
	}
}

func TestRunHelpFlagStillPrintsUsage(t *testing.T) {
	orig := launchTUIFn
	defer func() { launchTUIFn = orig }()

	called := false
	launchTUIFn = func() error {
		called = true
		return nil
	}

	output := captureStdout(t, func() {
		if err := run([]string{"--help"}); err != nil {
			t.Fatalf("run help: %v", err)
		}
	})

	if called {
		t.Fatalf("expected help flag to avoid launching TUI")
	}
	if !strings.Contains(output, "Run without a command to open the TUI.") {
		t.Fatalf("expected updated usage output, got:\n%s", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	fn()

	_ = w.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return strings.TrimSpace(buf.String())
}
