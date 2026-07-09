package updatecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	runtimepkg "crona/kernel/internal/runtime"
	"crona/shared/config"
	sharedtypes "crona/shared/types"
	versionpkg "crona/shared/version"
)

type latestRelease struct {
	Version      string
	Tag          string
	Name         string
	Notes        string
	URL          string
	InstallURL   string
	InstallAsset string
	ChecksumsURL string
	PublishedAt  string
	IsPrerelease bool
	IsBeta       bool
}

func (s *Service) fetchLatestRelease(
	ctx context.Context,
	channel sharedtypes.UpdateChannel,
) (latestRelease, error) {
	if release, ok := s.currentLocalRelease(); ok {
		return release, nil
	}
	if sharedtypes.NormalizeUpdateChannel(channel) == sharedtypes.UpdateChannelBeta {
		return s.fetchLatestReleaseFromList(ctx)
	}
	return s.fetchLatestStableRelease(ctx)
}

func (s *Service) fetchLatestStableRelease(ctx context.Context) (latestRelease, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/releases/latest",
			versionpkg.RepoOwner,
			versionpkg.RepoName,
		),
		nil,
	)
	if err != nil {
		return latestRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "crona/"+versionpkg.Current())

	resp, err := s.client.Do(req)
	if err != nil {
		return latestRelease{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return latestRelease{}, fmt.Errorf("github releases returned %s", resp.Status)
	}

	var payload githubReleasePayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return latestRelease{}, err
	}
	return s.releaseFromPayload(payload)
}

func (s *Service) fetchLatestReleaseFromList(ctx context.Context) (latestRelease, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/releases",
			versionpkg.RepoOwner,
			versionpkg.RepoName,
		),
		nil,
	)
	if err != nil {
		return latestRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "crona/"+versionpkg.Current())

	resp, err := s.client.Do(req)
	if err != nil {
		return latestRelease{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return latestRelease{}, fmt.Errorf("github releases returned %s", resp.Status)
	}

	var payloads []githubReleasePayload
	if err := json.NewDecoder(resp.Body).Decode(&payloads); err != nil {
		return latestRelease{}, err
	}

	best := latestRelease{}
	bestVersion := semver{}
	found := false
	for _, payload := range payloads {
		if payload.Draft {
			continue
		}
		release, err := s.releaseFromPayload(payload)
		if err != nil {
			continue
		}
		version, ok := parseSemver(release.Version)
		if !ok {
			continue
		}
		if !found || compareSemver(version, bestVersion) > 0 {
			best = release
			bestVersion = version
			found = true
		}
	}
	if !found {
		return latestRelease{}, fmt.Errorf("no eligible releases found")
	}
	return best, nil
}

type githubReleasePayload struct {
	Name        string `json:"name"`
	TagName     string `json:"tag_name"`
	Body        string `json:"body"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Prerelease  bool   `json:"prerelease"`
	Draft       bool   `json:"draft"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (s *Service) releaseFromPayload(payload githubReleasePayload) (latestRelease, error) {
	tag := strings.TrimSpace(payload.TagName)
	version := normalizeVersion(payload.TagName)
	if version == "" {
		return latestRelease{}, fmt.Errorf("latest release tag is empty")
	}
	installURL := ""
	installAsset := config.InstallerAssetNameForGOOS(s.targetGOOS())
	checksumsURL := ""
	for _, asset := range payload.Assets {
		switch strings.TrimSpace(asset.Name) {
		case installAsset:
			installURL = strings.TrimSpace(asset.BrowserDownloadURL)
		case "checksums.txt":
			checksumsURL = strings.TrimSpace(asset.BrowserDownloadURL)
		}
	}
	return latestRelease{
		Version:      version,
		Tag:          tag,
		Name:         strings.TrimSpace(payload.Name),
		Notes:        strings.TrimSpace(payload.Body),
		URL:          strings.TrimSpace(payload.HTMLURL),
		InstallURL:   installURL,
		InstallAsset: installAsset,
		ChecksumsURL: checksumsURL,
		PublishedAt:  strings.TrimSpace(payload.PublishedAt),
		IsPrerelease: payload.Prerelease,
		IsBeta: payload.Prerelease || versionpkg.IsBetaVersion(version) ||
			versionpkg.IsBetaVersion(tag),
	}, nil
}

func (r latestRelease) installUnavailableReason() string {
	installerAsset := strings.TrimSpace(r.InstallAsset)
	if installerAsset == "" {
		installerAsset = config.InstallerAssetName()
	}
	switch {
	case strings.TrimSpace(r.InstallURL) == "" && strings.TrimSpace(r.ChecksumsURL) == "":
		return "Release is missing installer and checksums assets."
	case strings.TrimSpace(r.InstallURL) == "":
		return fmt.Sprintf("Release is missing the %s asset.", installerAsset)
	case strings.TrimSpace(r.ChecksumsURL) == "":
		return "Release is missing the checksums.txt asset."
	default:
		return ""
	}
}

func effectiveUpdateChannel(settings *sharedtypes.CoreSettings) sharedtypes.UpdateChannel {
	if settings == nil {
		return sharedtypes.UpdateChannelStable
	}
	return sharedtypes.NormalizeUpdateChannel(settings.UpdateChannel)
}

func (s *Service) targetGOOS() string {
	if strings.TrimSpace(s.goos) != "" {
		return s.goos
	}
	return runtime.GOOS
}

func (s *Service) resolveInstallMetadataLocked(
	localReleaseActive bool,
) (sharedtypes.InstallSource, string, sharedtypes.UpdateChannel) {
	if source, formula, releaseChannel := s.persistedInstallMetadataLocked(); source != sharedtypes.InstallSourceUnknown {
		return source, formula, releaseChannel
	}
	if source := sourceFromEnv(); source != sharedtypes.InstallSourceUnknown {
		releaseChannel := defaultReleaseChannel(source, "")
		_ = runtimepkg.WriteInstallSourceDetails(
			s.installPath,
			source,
			defaultBrewFormula(source),
			releaseChannel,
		)
		return source, defaultBrewFormula(source), releaseChannel
	}
	if source := sourceFromExecutablePath(s.executablePathLocked()); source != sharedtypes.InstallSourceUnknown {
		releaseChannel := defaultReleaseChannel(source, s.executablePathLocked())
		_ = runtimepkg.WriteInstallSourceDetails(
			s.installPath,
			source,
			defaultBrewFormula(source),
			releaseChannel,
		)
		return source, defaultBrewFormula(source), releaseChannel
	}
	if localReleaseActive {
		_ = runtimepkg.WriteInstallSourceDetails(
			s.installPath,
			sharedtypes.InstallSourceScript,
			"",
			sharedtypes.UpdateChannelStable,
		)
		return sharedtypes.InstallSourceScript, "", sharedtypes.UpdateChannelStable
	}
	return sharedtypes.InstallSourceUnknown, "", sharedtypes.UpdateChannelStable
}

func (s *Service) persistedInstallMetadataLocked() (sharedtypes.InstallSource, string, sharedtypes.UpdateChannel) {
	if file, err := runtimepkg.LoadInstallSourceFile(s.installPath); err == nil {
		return sharedtypes.NormalizeInstallSource(file.InstallSource), strings.TrimSpace(file.BrewFormula), sharedtypes.NormalizeUpdateChannel(file.ReleaseChannel)
	}
	return sharedtypes.NormalizeInstallSource(s.status.InstallSource), strings.TrimSpace(s.status.BrewFormula), sharedtypes.NormalizeUpdateChannel(s.status.ReleaseChannel)
}

func (s *Service) executablePathLocked() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}
	return path
}

func sourceFromEnv() sharedtypes.InstallSource {
	if source := sharedtypes.ParseInstallSource(versionpkg.InstallSource); source != sharedtypes.InstallSourceUnknown {
		return source
	}
	return sharedtypes.ParseInstallSource(strings.TrimSpace(os.Getenv(config.EnvVarInstallSource)))
}

func sourceFromExecutablePath(path string) sharedtypes.InstallSource {
	normalized := strings.ToLower(strings.TrimSpace(path))
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	if normalized == "" {
		return sharedtypes.InstallSourceUnknown
	}
	if strings.Contains(normalized, "/opt/homebrew/") ||
		strings.Contains(normalized, "/usr/local/cellar/") ||
		strings.Contains(normalized, "/home/linuxbrew/.linuxbrew/") ||
		strings.Contains(normalized, "/homebrew/") {
		return sharedtypes.InstallSourceBrew
	}
	if strings.Contains(normalized, "/scoop/apps/") ||
		strings.Contains(normalized, "/scoop/shims/") {
		return sharedtypes.InstallSourceScoop
	}
	if strings.Contains(normalized, "/go/bin/") ||
		strings.Contains(normalized, "/gobin/") {
		return sharedtypes.InstallSourceGo
	}
	return sharedtypes.InstallSourceUnknown
}

func defaultBrewFormula(source sharedtypes.InstallSource) string {
	switch sharedtypes.NormalizeInstallSource(source) {
	case sharedtypes.InstallSourceBrew:
		return currentBrewFormula()
	default:
		return ""
	}
}

func defaultReleaseChannel(source sharedtypes.InstallSource, path string) sharedtypes.UpdateChannel {
	if sharedtypes.NormalizeInstallSource(source) == sharedtypes.InstallSourceUnknown {
		return sharedtypes.UpdateChannelStable
	}
	if value := strings.TrimSpace(os.Getenv(config.EnvVarReleaseChannel)); value != "" {
		return sharedtypes.NormalizeUpdateChannel(sharedtypes.UpdateChannel(value))
	}
	normalized := strings.ToLower(strings.TrimSpace(path))
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	if strings.Contains(normalized, "/scoop/apps/crona-beta/") {
		return sharedtypes.UpdateChannelBeta
	}
	return versionpkg.RunningChannel()
}

func updateCommandForStatus(status sharedtypes.UpdateStatus) string {
	switch sharedtypes.NormalizeInstallSource(status.InstallSource) {
	case sharedtypes.InstallSourceBrew:
		return brewCommandForStatus(status)
	case sharedtypes.InstallSourceScoop:
		return scoopCommandForStatus(status)
	case sharedtypes.InstallSourceScript:
		if versionpkg.InstallScriptDeprecationEnabled() {
			return versionpkg.InstallScriptMigrationURL
		}
		return "curl -fsSL https://crona.work/install.sh | sh"
	case sharedtypes.InstallSourceGo:
		return "go install github.com/webxsid/crona/...@latest"
	default:
		if strings.TrimSpace(status.ReleaseURL) != "" {
			return strings.TrimSpace(status.ReleaseURL)
		}
		return "Open GitHub release page"
	}
}

func currentBrewFormula() string {
	if versionpkg.IsBetaRelease() {
		return "crona-beta"
	}
	return "crona"
}

func previousBrewFormula() string {
	if versionpkg.IsBetaRelease() {
		return "crona"
	}
	return "crona-beta"
}

func brewFormulaMismatch(status sharedtypes.UpdateStatus) bool {
	if sharedtypes.NormalizeInstallSource(status.InstallSource) != sharedtypes.InstallSourceBrew {
		return false
	}
	recorded := strings.TrimSpace(status.BrewFormula)
	if recorded == "" {
		return false
	}
	return !strings.EqualFold(recorded, currentBrewFormula())
}

func brewMigrationCommand(status sharedtypes.UpdateStatus) string {
	if sharedtypes.NormalizeInstallSource(status.InstallSource) != sharedtypes.InstallSourceBrew {
		return ""
	}
	installed := strings.TrimSpace(status.BrewFormula)
	if installed == "" {
		installed = previousBrewFormula()
	}
	target := currentBrewFormula()
	return fmt.Sprintf("brew uninstall %s && brew install webxsid/tap/%s", installed, target)
}

func brewMigrationReason(status sharedtypes.UpdateStatus) string {
	if !brewFormulaMismatch(status) {
		return ""
	}
	installed := strings.TrimSpace(status.BrewFormula)
	if installed == "" {
		installed = previousBrewFormula()
	}
	return fmt.Sprintf(
		"Homebrew formula mismatch: installed via %s while this build expects %s. Run %s.",
		installed,
		currentBrewFormula(),
		brewMigrationCommand(status),
	)
}

func brewCommandForStatus(status sharedtypes.UpdateStatus) string {
	if brewFormulaMismatch(status) {
		return brewMigrationCommand(status)
	}
	formula := strings.TrimSpace(status.BrewFormula)
	if formula == "" {
		formula = currentBrewFormula()
	}
	return "brew upgrade " + formula
}

func scoopCommandForStatus(status sharedtypes.UpdateStatus) string {
	if sharedtypes.NormalizeUpdateChannel(status.ReleaseChannel) == sharedtypes.UpdateChannelBeta {
		return "scoop update crona-beta"
	}
	return "scoop update crona"
}

func (s *Service) hasLocalRelease() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.localRelease != nil
}

func (s *Service) currentLocalRelease() (latestRelease, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.localRelease == nil {
		return latestRelease{}, false
	}
	return *s.localRelease, true
}

func (s *Service) stopLocalReleaseServer() {
	s.mu.Lock()
	stop := s.stopLocalRelease
	s.stopLocalRelease = nil
	s.localRelease = nil
	s.localReleaseBase = ""
	s.mu.Unlock()
	if stop != nil {
		_ = stop()
	}
}
