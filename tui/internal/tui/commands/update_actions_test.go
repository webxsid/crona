package commands

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"crona/shared/config"
	"crona/tui/internal/api"
)

func TestExpectedChecksum(t *testing.T) {
	checksums := []byte("abc123  install-crona-tui.sh\nzzz999  checksums.txt\n")
	if got := expectedChecksum("install-crona-tui.sh", checksums); got != "abc123" {
		t.Fatalf("expected checksum abc123, got %q", got)
	}
	if got := expectedChecksum("missing", checksums); got != "" {
		t.Fatalf("expected missing checksum to be empty, got %q", got)
	}
}

func TestPrepareInstallCommandAddsForceAndReleaseBaseEnv(t *testing.T) {
	origDownloadFileFn := downloadFileFn
	origDownloadAndVerifyFn := downloadAndVerifyFn
	origUpdateInstallCommandFn := updateInstallCommandFn
	defer func() {
		downloadFileFn = origDownloadFileFn
		downloadAndVerifyFn = origDownloadAndVerifyFn
		updateInstallCommandFn = origUpdateInstallCommandFn
	}()

	downloadFileFn = func(rawURL, path string) (string, error) {
		if err := os.WriteFile(path, []byte("checksums"), 0o600); err != nil {
			return "", err
		}
		return "", nil
	}
	downloadAndVerifyFn = func(rawURL, path, assetName string, checksums []byte) (string, error) {
		if err := os.WriteFile(path, []byte("#!/usr/bin/env sh\n"), 0o700); err != nil {
			return "", err
		}
		return "", nil
	}
	updateInstallCommandFn = func(installerPath string) (*exec.Cmd, error) {
		return exec.Command("sh", "-c", "exit 0"), nil
	}

	status := &api.UpdateStatus{
		InstallAvailable: true,
		ReleaseTag:       "v0.4.0-beta.3",
		InstallScriptURL: "http://127.0.0.1:3210/install-crona-tui.sh",
		ChecksumsURL:     "http://127.0.0.1:3210/checksums.txt",
	}
	cmd, err := prepareInstallCommand(status, true, "")
	if err != nil {
		t.Fatalf("prepareInstallCommand returned error: %v", err)
	}
	env := strings.Join(cmd.Env, "\n")
	for _, want := range []string{
		"CRONA_INSTALL_FORCE=1",
		config.EnvVarReleaseBaseURL + "=http://127.0.0.1:3210",
		config.EnvVarInstallDir + "=",
		config.EnvVarRuntimeDir + "=",
	} {
		if !strings.Contains(env, want) {
			t.Fatalf("expected command env to contain %q, got:\n%s", want, env)
		}
	}
}

func TestPrepareInstallCommandWithLocalReleaseBundle(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("local shell installer smoke test is unix-only")
	}

	version := "v0.4.0-beta.3-local"
	releaseDir := filepath.Join(t.TempDir(), version)
	if err := os.MkdirAll(releaseDir, 0o755); err != nil {
		t.Fatalf("MkdirAll releaseDir: %v", err)
	}

	installerName := config.InstallerAssetNameForGOOS(runtime.GOOS)
	installerBody, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "scripts", "install_tui.sh.tmpl"))
	if err != nil {
		t.Fatalf("ReadFile installer template: %v", err)
	}
	renderedInstaller := strings.ReplaceAll(string(installerBody), "__VERSION__", version)
	renderedInstaller = strings.ReplaceAll(renderedInstaller, "__REPO__", "webxsid/crona")
	if err := os.WriteFile(filepath.Join(releaseDir, installerName), []byte(renderedInstaller), 0o755); err != nil {
		t.Fatalf("WriteFile installer: %v", err)
	}

	bundleName := "crona-bundle-" + version + "-" + runtime.GOOS + "-" + runtime.GOARCH + ".zip"
	assetsName := "crona-assets-" + version + ".tar.gz"
	if err := writeLocalBundle(filepath.Join(releaseDir, bundleName), version); err != nil {
		t.Fatalf("writeLocalBundle: %v", err)
	}
	if err := writeLocalAssets(filepath.Join(releaseDir, assetsName)); err != nil {
		t.Fatalf("writeLocalAssets: %v", err)
	}
	if err := writeChecksums(filepath.Join(releaseDir, "checksums.txt"),
		filepath.Join(releaseDir, installerName),
		filepath.Join(releaseDir, bundleName),
		filepath.Join(releaseDir, assetsName),
	); err != nil {
		t.Fatalf("writeChecksums: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("loopback bind unavailable in this environment: %v", err)
		}
		t.Fatalf("Listen: %v", err)
	}
	server := httptest.NewUnstartedServer(http.FileServer(http.Dir(releaseDir)))
	server.Listener = listener
	server.Start()
	defer server.Close()

	status := &api.UpdateStatus{
		LatestVersion:    strings.TrimPrefix(version, "v"),
		ReleaseTag:       version,
		InstallAvailable: true,
		InstallScriptURL: server.URL + "/" + installerName,
		ChecksumsURL:     server.URL + "/checksums.txt",
	}
	installDir, runtimeDir := localUpdateInstallRoots(status)
	_ = os.RemoveAll(filepath.Dir(installDir))
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(installDir)) })

	cmd, err := prepareInstallCommand(status, true, "")
	if err != nil {
		t.Fatalf("prepareInstallCommand returned error: %v", err)
	}
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err = cmd.Run()
	if err != nil {
		t.Fatalf("expected successful install flow, got error: %v\noutput:\n%s", err, output.String())
	}
	for _, want := range []string{"Installing Crona", "Installing binaries...", "Extracting bundled assets..."} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("expected install output to contain %q, got:\n%s", want, output.String())
		}
	}
	for _, path := range []string{
		filepath.Join(installDir, "crona"),
		filepath.Join(installDir, "crona-kernel"),
		filepath.Join(installDir, "crona-tui"),
		filepath.Join(runtimeDir, "assets", "bundled", "export", "sample.txt"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected installed artifact %s: %v", path, err)
		}
	}
}

func writeLocalBundle(path, version string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	archive := zip.NewWriter(file)
	files := map[string]string{
		"crona-" + version + "-" + runtime.GOOS + "-" + runtime.GOARCH:        "#!/bin/sh\nexit 0\n",
		"crona-kernel-" + version + "-" + runtime.GOOS + "-" + runtime.GOARCH: "#!/bin/sh\nexit 0\n",
		"crona-tui-" + version + "-" + runtime.GOOS + "-" + runtime.GOARCH:    "#!/bin/sh\nexit 0\n",
	}
	for name, body := range files {
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		header.SetMode(0o755)
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := io.Copy(writer, strings.NewReader(body)); err != nil {
			return err
		}
	}
	return archive.Close()
}

func writeLocalAssets(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	gzw := gzip.NewWriter(file)
	defer func() { _ = gzw.Close() }()
	tw := tar.NewWriter(gzw)
	defer func() { _ = tw.Close() }()

	body := []byte("sample asset\n")
	header := &tar.Header{
		Name: "export/sample.txt",
		Mode: 0o644,
		Size: int64(len(body)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err = tw.Write(body)
	return err
}

func writeChecksums(path string, files ...string) error {
	var body bytes.Buffer
	for _, name := range files {
		raw, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(raw)
		body.WriteString(hex.EncodeToString(sum[:]))
		body.WriteString("  ")
		body.WriteString(filepath.Base(name))
		body.WriteString("\n")
	}
	return os.WriteFile(path, body.Bytes(), 0o644)
}
