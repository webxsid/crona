package helpers

const (
	supportGitHubRepoBaseURL  = "https://github.com/webxsid/crona"
	supportGitHubDocsURL      = "https://github.com/webxsid/crona/blob/main/docs/README.md"
	supportFeedbackRoadmapURL = "https://crona.userjot.com/"
)

func SupportRepoURL() string {
	return supportGitHubRepoBaseURL
}

func SupportFeedbackRoadmapURL() string {
	return supportFeedbackRoadmapURL
}

func SupportReleasesURL() string {
	return supportGitHubRepoBaseURL + "/releases"
}

func SupportDocumentationURL() string {
	return supportGitHubDocsURL
}
