package posthog

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sharedtypes "crona/shared/types"
	versionpkg "crona/shared/version"
)

func TestNewReturnsNoopWhenDisabled(t *testing.T) {
	runtimeDir := t.TempDir()
	t.Setenv("CRONA_HOME", runtimeDir)

	client, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if client.Enabled() {
		t.Fatal("expected disabled client")
	}
	if client.ErrorReportingEnabled() {
		t.Fatal("expected error reporting disabled client")
	}
	if client.DistinctID() != "" {
		t.Fatalf("expected empty distinct ID, got %q", client.DistinctID())
	}
	if err := client.Capture(EventTUIStarted, Properties{"entrypoint": "tui"}); err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if err := client.Identify(Properties{"app": "tui"}); err != nil {
		t.Fatalf("Identify: %v", err)
	}
	if err := client.ReportError("handled", errors.New("boom"), Properties{"operation": "test"}); err != nil {
		t.Fatalf("ReportError: %v", err)
	}
	body, err := os.ReadFile(posthogLogPath(t, runtimeDir))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	plain := string(body)
	for _, want := range []string{"operation=capture", "result=skipped", "reason=disabled", "operation=identify", "operation=report_error"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected log to contain %q, got %q", want, plain)
		}
	}
}

func TestNewCreatesIdentityFileWhenEnabled(t *testing.T) {
	runtimeDir := t.TempDir()

	client, err := New(Config{
		Enabled:               true,
		UsageEnabled:          true,
		ErrorReportingEnabled: true,
		APIKey:                "test-key",
		Host:                  "http://127.0.0.1:1",
		App:                   "kernel",
		Version:               "test",
		Mode:                  "Dev",
		RuntimeDir:            runtimeDir,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = client.Close() }()

	if !client.Enabled() {
		t.Fatal("expected enabled client")
	}
	if !strings.HasPrefix(client.DistinctID(), "anon_") {
		t.Fatalf("expected anonymous distinct ID, got %q", client.DistinctID())
	}
	path := filepath.Join(runtimeDir, identityFilename)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected identity file: %v", err)
	}
	if err := client.Capture(EventTUIStarted, Properties{"entrypoint": "tui"}); err != nil {
		t.Fatalf("Capture: %v", err)
	}
	body, err := os.ReadFile(posthogLogPath(t, runtimeDir))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	plain := string(body)
	for _, want := range []string{"operation=capture", "result=attempted", "event=tui_started", "property_keys=entrypoint"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected log to contain %q, got %q", want, plain)
		}
	}
}

func TestNewReusesExistingIdentity(t *testing.T) {
	runtimeDir := t.TempDir()

	first, err := New(Config{
		Enabled:               true,
		UsageEnabled:          true,
		ErrorReportingEnabled: true,
		APIKey:                "test-key",
		Host:                  "http://127.0.0.1:1",
		App:                   "tui",
		Version:               "test",
		Mode:                  "Dev",
		RuntimeDir:            runtimeDir,
	})
	if err != nil {
		t.Fatalf("first New: %v", err)
	}
	defer func() { _ = first.Close() }()

	second, err := New(Config{
		Enabled:               true,
		UsageEnabled:          true,
		ErrorReportingEnabled: true,
		APIKey:                "test-key",
		Host:                  "http://127.0.0.1:1",
		App:                   "cli",
		Version:               "test",
		Mode:                  "Dev",
		RuntimeDir:            runtimeDir,
	})
	if err != nil {
		t.Fatalf("second New: %v", err)
	}
	defer func() { _ = second.Close() }()

	if first.DistinctID() == "" || second.DistinctID() == "" {
		t.Fatal("expected non-empty distinct IDs")
	}
	if first.DistinctID() != second.DistinctID() {
		t.Fatalf("expected stable distinct ID, got %q then %q", first.DistinctID(), second.DistinctID())
	}
}

func TestNewReplacesMalformedIdentityFile(t *testing.T) {
	runtimeDir := t.TempDir()
	path := filepath.Join(runtimeDir, identityFilename)
	if err := os.WriteFile(path, []byte("{bad json"), filePerm); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	client, err := New(Config{
		Enabled:               true,
		UsageEnabled:          true,
		ErrorReportingEnabled: true,
		APIKey:                "test-key",
		Host:                  "http://127.0.0.1:1",
		App:                   "kernel",
		Version:               "test",
		Mode:                  "Dev",
		RuntimeDir:            runtimeDir,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = client.Close() }()

	if client.DistinctID() == "" {
		t.Fatal("expected regenerated distinct ID")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.Contains(string(body), "{bad json") {
		t.Fatalf("expected malformed identity file to be replaced, got %q", string(body))
	}
}

func TestReportErrorWorksWhenUsageTelemetryIsDisabled(t *testing.T) {
	runtimeDir := t.TempDir()

	client, err := New(Config{
		Enabled:               true,
		UsageEnabled:          false,
		ErrorReportingEnabled: true,
		APIKey:                "test-key",
		Host:                  "http://127.0.0.1:1",
		App:                   "tui",
		Version:               "test",
		Mode:                  "Dev",
		RuntimeDir:            runtimeDir,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = client.Close() }()

	if client.UsageEnabled() {
		t.Fatal("expected usage telemetry disabled")
	}
	if !client.ErrorReportingEnabled() {
		t.Fatal("expected error reporting enabled")
	}
	if err := client.Capture(EventTUIStarted, Properties{"entrypoint": "tui"}); err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if err := client.ReportError("handled", errors.New("dial unix /tmp/kernel.sock failed"), Properties{"operation": "load_settings"}); err != nil {
		t.Fatalf("ReportError: %v", err)
	}
	body, err := os.ReadFile(posthogLogPath(t, runtimeDir))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	plain := string(body)
	for _, want := range []string{"operation=capture", "result=skipped", "error=usage_disabled", "operation=report_error", "event=error_reported"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected log to contain %q, got %q", want, plain)
		}
	}
}

func TestReportErrorSanitizesPathBearingMessages(t *testing.T) {
	props := sanitizedErrorProperties("handled", errors.New("open /Users/test/secret.txt: permission denied"), Properties{"operation": "load_settings", "entrypoint": "tui"})
	if got := props["message"]; got != "redacted path-bearing error" {
		t.Fatalf("expected redacted path-bearing error, got %#v", got)
	}
	if got := props["operation"]; got != "load_settings" {
		t.Fatalf("expected operation to survive sanitization, got %#v", got)
	}
}

func TestDefaultEventPropertiesIncludeRuntimeChannel(t *testing.T) {
	oldVersion := versionpkg.Version
	versionpkg.Version = "1.3.0-beta.1"
	defer func() { versionpkg.Version = oldVersion }()

	props := defaultEventProperties(Config{
		App:     "tui",
		Version: "1.3.0-beta.1",
		Mode:    "Dev",
	})
	if got := props["app"]; got != "tui" {
		t.Fatalf("expected app property, got %#v", got)
	}
	if got := props["env_mode"]; got != "Dev" {
		t.Fatalf("expected env_mode property, got %#v", got)
	}
	if got, ok := props["running_channel"].(sharedtypes.UpdateChannel); !ok || got != sharedtypes.UpdateChannelBeta {
		t.Fatalf("expected beta running channel, got %#v", props["running_channel"])
	}
	if got := props["running_is_beta"]; got != true {
		t.Fatalf("expected beta running flag, got %#v", got)
	}
	if got := props["app_version"]; got != "1.3.0-beta.1" {
		t.Fatalf("expected app version property, got %#v", got)
	}
}

func posthogLogPath(t *testing.T, runtimeDir string) string {
	t.Helper()
	return filepath.Join(runtimeDir, "logs", logDirName, time.Now().UTC().Format("2006-01-02")+".log")
}
