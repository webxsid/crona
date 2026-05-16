package config

import "testing"

func TestPostHogDefaultsInProdMode(t *testing.T) {
	t.Setenv(EnvVarMode, ModeProd)
	t.Setenv(EnvVarPostHogAPIKey, "")
	t.Setenv(EnvVarPostHogHost, "")
	t.Setenv(EnvVarPostHogEnabled, "")

	if got := PostHogAPIKey(); got == "" {
		t.Fatal("expected prod posthog api key default")
	}
	if got := PostHogHost(); got == "" {
		t.Fatal("expected prod posthog host default")
	}
	if !PostHogEnabled() {
		t.Fatal("expected prod posthog telemetry default to be enabled")
	}
}

func TestPostHogDefaultsStayOffInDevMode(t *testing.T) {
	t.Setenv(EnvVarMode, ModeDev)
	t.Setenv(EnvVarPostHogAPIKey, "")
	t.Setenv(EnvVarPostHogHost, "")
	t.Setenv(EnvVarPostHogEnabled, "")

	if got := PostHogAPIKey(); got != "" {
		t.Fatalf("expected dev posthog api key default to stay empty, got %q", got)
	}
	if got := PostHogHost(); got != "" {
		t.Fatalf("expected dev posthog host default to stay empty, got %q", got)
	}
	if PostHogEnabled() {
		t.Fatal("expected dev posthog telemetry default to remain disabled")
	}
}

func TestPostHogEnvOverridesWin(t *testing.T) {
	t.Setenv(EnvVarMode, ModeProd)
	t.Setenv(EnvVarPostHogAPIKey, "env-key")
	t.Setenv(EnvVarPostHogHost, "https://example.com")
	t.Setenv(EnvVarPostHogEnabled, "false")

	if got := PostHogAPIKey(); got != "env-key" {
		t.Fatalf("expected env override api key, got %q", got)
	}
	if got := PostHogHost(); got != "https://example.com" {
		t.Fatalf("expected env override host, got %q", got)
	}
	if PostHogEnabled() {
		t.Fatal("expected explicit false override to disable telemetry")
	}
}
