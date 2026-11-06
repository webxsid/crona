package helpers

import versionpkg "crona.work/shared/version"

func SupportRepoURL() string {
	return versionpkg.RepoURL()
}

func SupportFeedbackRoadmapURL() string {
	return versionpkg.RepoURL() + "/discussions"
}

func SupportReleasesURL() string {
	return versionpkg.RepoURL() + "/releases"
}

func SupportDocumentationURL() string {
	return versionpkg.RepoURL() + "/blob/main/docs/README.md"
}
