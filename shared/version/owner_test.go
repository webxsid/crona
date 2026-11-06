package version

import (
	"os"
	"regexp"
	"testing"
)

// The release scripts keep their own copy of the owner; it must match RepoOwner.
func TestScriptsShareRepoOwner(t *testing.T) {
	meta, err := os.ReadFile("../../scripts/project_meta.sh")
	if err != nil {
		t.Fatalf("read project_meta.sh: %v", err)
	}
	match := regexp.MustCompile(`(?m)^PROJECT_OWNER="([^"]+)"`).FindSubmatch(meta)
	if match == nil || string(match[1]) != RepoOwner {
		t.Fatalf("expected scripts/project_meta.sh PROJECT_OWNER=%q, got %q", RepoOwner, match)
	}
	pkg := regexp.MustCompile(`(?m)^PACKAGE_OWNER="([^"]+)"`).FindSubmatch(meta)
	if pkg == nil || string(pkg[1]) != PackageOwner {
		t.Fatalf("expected scripts/project_meta.sh PACKAGE_OWNER=%q, got %q", PackageOwner, pkg)
	}
}

func TestRepoLinksBuildOnOwner(t *testing.T) {
	if got, want := RepoURL(), "https://github.com/"+RepoOwner+"/"+RepoName; got != want {
		t.Fatalf("RepoURL() = %q, want %q", got, want)
	}
	if got, want := BrewTapFormula("crona"), PackageOwner+"/tap/crona"; got != want {
		t.Fatalf("BrewTapFormula() = %q, want %q", got, want)
	}
}
