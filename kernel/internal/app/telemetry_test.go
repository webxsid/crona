package app

import (
	"testing"

	sharedposthog "crona/shared/posthog"
	sharedtypes "crona/shared/types"
)

type recordedEvent struct {
	name       string
	properties sharedposthog.Properties
}

type stubTelemetryClient struct {
	events []recordedEvent
}

func (s *stubTelemetryClient) Enabled() bool                           { return true }
func (s *stubTelemetryClient) UsageEnabled() bool                      { return true }
func (s *stubTelemetryClient) ErrorReportingEnabled() bool             { return true }
func (s *stubTelemetryClient) DistinctID() string                      { return "anon_test" }
func (s *stubTelemetryClient) Identify(sharedposthog.Properties) error { return nil }
func (s *stubTelemetryClient) ReportError(string, error, sharedposthog.Properties) error {
	return nil
}
func (s *stubTelemetryClient) Flush() error { return nil }
func (s *stubTelemetryClient) Close() error { return nil }
func (s *stubTelemetryClient) Capture(event string, properties sharedposthog.Properties) error {
	s.events = append(s.events, recordedEvent{name: event, properties: properties})
	return nil
}

func TestCaptureDaemonStarted(t *testing.T) {
	telemetry := &stubTelemetryClient{}

	if err := captureDaemonStarted(telemetry, "unix", sharedtypes.UpdateChannelStable); err != nil {
		t.Fatalf("captureDaemonStarted: %v", err)
	}
	if len(telemetry.events) != 1 {
		t.Fatalf("expected one event, got %d", len(telemetry.events))
	}
	if got := telemetry.events[0].name; got != sharedposthog.EventDaemonStarted {
		t.Fatalf("expected %q, got %q", sharedposthog.EventDaemonStarted, got)
	}
	if got := telemetry.events[0].properties["entrypoint"]; got != "daemon" {
		t.Fatalf("expected daemon entrypoint, got %#v", got)
	}
	if got := telemetry.events[0].properties["transport"]; got != "unix" {
		t.Fatalf("expected unix transport, got %#v", got)
	}
	if got := telemetry.events[0].properties["running_channel"]; got != string(sharedtypes.UpdateChannelStable) {
		t.Fatalf("expected stable channel, got %#v", got)
	}
}

func TestCaptureDaemonStopped(t *testing.T) {
	telemetry := &stubTelemetryClient{}

	if err := captureDaemonStopped(telemetry, "tcp", sharedtypes.UpdateChannelBeta); err != nil {
		t.Fatalf("captureDaemonStopped: %v", err)
	}
	if len(telemetry.events) != 1 {
		t.Fatalf("expected one event, got %d", len(telemetry.events))
	}
	if got := telemetry.events[0].name; got != sharedposthog.EventDaemonStopped {
		t.Fatalf("expected %q, got %q", sharedposthog.EventDaemonStopped, got)
	}
	if got := telemetry.events[0].properties["transport"]; got != "tcp" {
		t.Fatalf("expected tcp transport, got %#v", got)
	}
	if got := telemetry.events[0].properties["running_channel"]; got != string(sharedtypes.UpdateChannelBeta) {
		t.Fatalf("expected beta channel, got %#v", got)
	}
}

func TestTelemetryUsageAndErrorReportingRequireOnboarding(t *testing.T) {
	settings := &sharedtypes.CoreSettings{
		OnboardingCompleted:   false,
		UsageTelemetryEnabled: true,
		ErrorReportingEnabled: true,
	}
	if telemetryUsageEnabled(settings) {
		t.Fatalf("expected usage telemetry disabled before onboarding completes")
	}
	if telemetryErrorReportingEnabled(settings) {
		t.Fatalf("expected error reporting disabled before onboarding completes")
	}

	settings.OnboardingCompleted = true
	if !telemetryUsageEnabled(settings) {
		t.Fatalf("expected usage telemetry enabled after onboarding completes")
	}
	if !telemetryErrorReportingEnabled(settings) {
		t.Fatalf("expected error reporting enabled after onboarding completes")
	}
}
