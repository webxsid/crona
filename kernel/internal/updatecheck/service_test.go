package updatecheck

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFetchLatestReleaseRequiresInstallerAndChecksumsAssets(t *testing.T) {
	service := &Service{
		client: &http.Client{
			Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				body := `{
					"name":"Crona 0.3.0",
					"tag_name":"v0.3.0",
					"body":"Notes",
					"html_url":"https://example.com/release",
					"published_at":"2026-03-25T00:00:00Z",
					"assets":[
						{"name":"install-crona-tui.sh","browser_download_url":"https://example.com/install"},
						{"name":"checksums.txt","browser_download_url":"https://example.com/checksums"}
					]
				}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	release, err := service.fetchLatestRelease(context.Background())
	if err != nil {
		t.Fatalf("fetchLatestRelease returned error: %v", err)
	}
	if release.InstallURL != "https://example.com/install" {
		t.Fatalf("expected install URL, got %q", release.InstallURL)
	}
	if release.ChecksumsURL != "https://example.com/checksums" {
		t.Fatalf("expected checksums URL, got %q", release.ChecksumsURL)
	}
	if got := release.installUnavailableReason(); got != "" {
		t.Fatalf("expected install to be available, got reason %q", got)
	}
}

func TestInstallUnavailableReason(t *testing.T) {
	tests := []struct {
		name  string
		input latestRelease
		want  string
	}{
		{name: "missing both", input: latestRelease{}, want: "Release is missing installer and checksums assets."},
		{name: "missing installer", input: latestRelease{ChecksumsURL: "https://example.com/checksums"}, want: "Release is missing the install-crona-tui.sh asset."},
		{name: "missing checksums", input: latestRelease{InstallURL: "https://example.com/install"}, want: "Release is missing the checksums.txt asset."},
	}

	for _, tc := range tests {
		if got := tc.input.installUnavailableReason(); got != tc.want {
			t.Fatalf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
