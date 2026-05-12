package main

import (
	"testing"

	sharedposthog "crona/shared/posthog"
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

func TestCaptureTUIStarted(t *testing.T) {
	telemetry := &stubTelemetryClient{}

	if err := captureTUIStarted(telemetry); err != nil {
		t.Fatalf("captureTUIStarted: %v", err)
	}
	if len(telemetry.events) != 1 {
		t.Fatalf("expected one event, got %d", len(telemetry.events))
	}
	if got := telemetry.events[0].name; got != sharedposthog.EventTUIStarted {
		t.Fatalf("expected %q, got %q", sharedposthog.EventTUIStarted, got)
	}
	if got := telemetry.events[0].properties["entrypoint"]; got != "tui" {
		t.Fatalf("expected tui entrypoint, got %#v", got)
	}
}
