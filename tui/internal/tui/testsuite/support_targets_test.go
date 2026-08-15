package testsuite

import (
	"net/url"
	"testing"

	helperpkg "crona/tui/internal/tui/helpers"
)

func TestSupportFeedbackRoadmapURL(t *testing.T) {
	raw := helperpkg.SupportFeedbackRoadmapURL()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("url.Parse: %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "crona.userjot.com" {
		t.Fatalf("expected UserJot https URL, got %q", raw)
	}
}

func TestSupportPublicURLs(t *testing.T) {
	tests := map[string]string{
		"feedback-roadmap": helperpkg.SupportFeedbackRoadmapURL(),
		"releases":         helperpkg.SupportReleasesURL(),
		"documentation":    helperpkg.SupportDocumentationURL(),
	}
	for name, raw := range tests {
		parsed, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("%s url.Parse: %v", name, err)
		}
		if parsed.Scheme != "https" {
			t.Fatalf("%s expected https url, got %q", name, raw)
		}
	}
}
