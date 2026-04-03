package helpers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"crona/shared/config"
	"crona/tui/internal/api"
)

const SupportRecentDiagnosticsWindow = time.Hour

func SupportRecentWindowLabel(window time.Duration) string {
	if window%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(window.Hours()))
	}
	if window%time.Minute == 0 {
		return fmt.Sprintf("%dm", int(window.Minutes()))
	}
	return window.String()
}

func SupportRuntimeBaseDir(info *api.KernelInfo) string {
	if info != nil {
		scratchDir := strings.TrimSpace(info.ScratchDir)
		if scratchDir != "" {
			return filepath.Clean(filepath.Dir(scratchDir))
		}
	}
	base, err := config.RuntimeBaseDir()
	if err != nil {
		return ""
	}
	return filepath.Clean(base)
}

func SupportBundleDisplayName(path string) string {
	name := strings.TrimSpace(filepath.Base(path))
	if name == "" {
		return "support-bundle.zip"
	}
	return name
}

func HumanizeSupportBytes(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	units := []string{"KB", "MB", "GB"}
	value := float64(size)
	for _, unit := range units {
		value /= 1024
		if value < 1024 || unit == units[len(units)-1] {
			if value >= 10 {
				return fmt.Sprintf("%.0f %s", value, unit)
			}
			return fmt.Sprintf("%.1f %s", value, unit)
		}
	}
	return fmt.Sprintf("%d B", size)
}

type SupportCollectionMeta struct {
	GeneratedAt        string   `json:"generatedAt"`
	Window             string   `json:"window"`
	WindowStart        string   `json:"windowStart"`
	WindowEnd          string   `json:"windowEnd"`
	RedactionMode      string   `json:"redactionMode"`
	OpsCount           int      `json:"opsCount"`
	TUIErrorCount      int      `json:"tuiErrorCount"`
	KernelErrorCount   int      `json:"kernelErrorCount"`
	CollectionWarnings []string `json:"collectionWarnings,omitempty"`
}

func GenerateSupportBundle(baseDir string, now time.Time, window time.Duration, input SupportDiagnosticsInput, ops []api.Op, tuiErrors, kernelErrors, collectionErrors []string) (string, int64, error) {
	if strings.TrimSpace(baseDir) == "" {
		return "", 0, os.ErrNotExist
	}

	redactor := newSupportRedactor(baseDir, input)
	redactedInput := redactor.RedactInput(input)
	report := SupportDiagnosticsReport(redactedInput)
	redactedOps := redactor.RedactOps(ops)
	redactedTUIErrors := redactor.RedactLines(tuiErrors)
	redactedKernelErrors := redactor.RedactLines(kernelErrors)
	redactedWarnings := redactor.RedactLines(collectionErrors)

	summaryBody := buildSupportSummaryBody(report, now, window)
	opsJSON, err := json.MarshalIndent(redactedOps, "", "  ")
	if err != nil {
		return "", 0, err
	}
	metaJSON, err := json.MarshalIndent(SupportCollectionMeta{
		GeneratedAt:        now.UTC().Format(time.RFC3339),
		Window:             SupportRecentWindowLabel(window),
		WindowStart:        now.UTC().Add(-window).Format(time.RFC3339),
		WindowEnd:          now.UTC().Format(time.RFC3339),
		RedactionMode:      "safe",
		OpsCount:           len(redactedOps),
		TUIErrorCount:      len(redactedTUIErrors),
		KernelErrorCount:   len(redactedKernelErrors),
		CollectionWarnings: redactedWarnings,
	}, "", "  ")
	if err != nil {
		return "", 0, err
	}

	dir := filepath.Join(baseDir, "logs", "support", now.UTC().Format("2006-01-02"))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", 0, err
	}
	path := filepath.Join(dir, fmt.Sprintf("support-bundle-%s.zip", now.UTC().Format("20060102T150405Z")))

	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	if err := writeZipEntry(zw, "summary.txt", []byte(summaryBody)); err != nil {
		return "", 0, err
	}
	if err := writeZipEntry(zw, "recent_ops.json", opsJSON); err != nil {
		return "", 0, err
	}
	if err := writeZipEntry(zw, "recent_errors_tui.log", []byte(joinSupportLines(redactedTUIErrors))); err != nil {
		return "", 0, err
	}
	if err := writeZipEntry(zw, "recent_errors_kernel.log", []byte(joinSupportLines(redactedKernelErrors))); err != nil {
		return "", 0, err
	}
	if err := writeZipEntry(zw, "collection_meta.json", metaJSON); err != nil {
		return "", 0, err
	}
	if err := zw.Close(); err != nil {
		return "", 0, err
	}
	if err := os.WriteFile(path, archive.Bytes(), 0o600); err != nil {
		return "", 0, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, err
	}
	return path, info.Size(), nil
}

func ReadRecentSupportErrorEntries(baseDir string, since, until time.Time) (tuiErrors []string, kernelErrors []string, errs []string) {
	if strings.TrimSpace(baseDir) == "" {
		return nil, nil, []string{"runtime base directory unavailable"}
	}

	tuiErrors, err := readSupportErrorEntries(filepath.Join(baseDir, "logs", "tui"), since, until)
	if err != nil {
		errs = append(errs, "tui errors: "+err.Error())
	}
	kernelErrors, err = readSupportErrorEntries(filepath.Join(baseDir, "logs"), since, until)
	if err != nil {
		errs = append(errs, "kernel errors: "+err.Error())
	}
	return tuiErrors, kernelErrors, errs
}

type supportRedactor struct {
	baseDir         string
	runtimeBase     string
	endpoint        string
	scratchDir      string
	tuiPath         string
	kernelPath      string
	reportsDir      string
	icsDir          string
	pdfRendererPath string
}

func newSupportRedactor(baseDir string, input SupportDiagnosticsInput) supportRedactor {
	r := supportRedactor{
		baseDir:     strings.TrimSpace(baseDir),
		runtimeBase: strings.TrimSpace(baseDir),
		tuiPath:     strings.TrimSpace(input.TUIPath),
		kernelPath:  strings.TrimSpace(input.KernelPath),
	}
	if input.KernelInfo != nil {
		r.endpoint = strings.TrimSpace(input.KernelInfo.Endpoint)
		r.scratchDir = strings.TrimSpace(input.KernelInfo.ScratchDir)
	}
	if input.ExportAssets != nil {
		r.reportsDir = strings.TrimSpace(input.ExportAssets.ReportsDir)
		r.icsDir = strings.TrimSpace(input.ExportAssets.ICSDir)
		r.pdfRendererPath = strings.TrimSpace(input.ExportAssets.PDFRendererPath)
	}
	return r
}

func (r supportRedactor) RedactInput(input SupportDiagnosticsInput) SupportDiagnosticsInput {
	out := input
	out.TUIPath = r.redactPath(out.TUIPath)
	out.KernelPath = r.redactPath(out.KernelPath)
	if input.KernelInfo != nil {
		info := *input.KernelInfo
		info.Endpoint = r.redactEndpoint(info.Endpoint)
		info.SocketPath = r.redactPath(info.SocketPath)
		info.ScratchDir = r.redactPath(info.ScratchDir)
		info.ExecutablePath = r.redactPath(info.ExecutablePath)
		out.KernelInfo = &info
	}
	if input.ExportAssets != nil {
		assets := *input.ExportAssets
		assets.ReportsDir = r.redactPath(assets.ReportsDir)
		assets.DefaultReportsDir = r.redactPath(assets.DefaultReportsDir)
		assets.ICSDir = r.redactPath(assets.ICSDir)
		assets.DefaultICSDir = r.redactPath(assets.DefaultICSDir)
		assets.PDFRendererPath = r.redactPath(assets.PDFRendererPath)
		out.ExportAssets = &assets
	}
	if input.Context != nil {
		ctx := *input.Context
		if ctx.RepoName != nil {
			value := "<redacted>"
			ctx.RepoName = &value
		}
		if ctx.StreamName != nil {
			value := "<redacted>"
			ctx.StreamName = &value
		}
		if ctx.IssueTitle != nil {
			value := "<redacted>"
			ctx.IssueTitle = &value
		}
		out.Context = &ctx
	}
	if input.UpdateStatus != nil {
		status := *input.UpdateStatus
		status.ReleaseURL = ""
		status.InstallScriptURL = ""
		status.ChecksumsURL = ""
		out.UpdateStatus = &status
	}
	return out
}

func (r supportRedactor) RedactOps(ops []api.Op) []api.Op {
	out := make([]api.Op, 0, len(ops))
	for _, op := range ops {
		next := op
		next.EntityID = r.sanitizeString(next.EntityID)
		next.Payload = r.sanitizeValue(next.Payload, "")
		out = append(out, next)
	}
	return out
}

func (r supportRedactor) RedactLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, r.sanitizeString(line))
	}
	return out
}

func (r supportRedactor) sanitizeValue(value any, key string) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for k, v := range typed {
			out[k] = r.sanitizeValue(v, k)
		}
		return out
	case []any:
		out := make([]any, 0, len(typed))
		for _, v := range typed {
			out = append(out, r.sanitizeValue(v, key))
		}
		return out
	case string:
		if isSensitiveSupportKey(key) {
			return "<redacted>"
		}
		return r.sanitizeString(typed)
	default:
		return value
	}
}

func isSensitiveSupportKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, fragment := range []string{"title", "name", "description", "note", "notes", "path", "endpoint", "token"} {
		if strings.Contains(key, fragment) {
			return true
		}
	}
	return false
}

func (r supportRedactor) sanitizeString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	for raw, replacement := range map[string]string{
		r.endpoint:        "<redacted-endpoint>",
		r.scratchDir:      r.redactPath(r.scratchDir),
		r.tuiPath:         r.redactPath(r.tuiPath),
		r.kernelPath:      r.redactPath(r.kernelPath),
		r.reportsDir:      r.redactPath(r.reportsDir),
		r.icsDir:          r.redactPath(r.icsDir),
		r.pdfRendererPath: r.redactPath(r.pdfRendererPath),
		r.runtimeBase:     "<runtime>",
		r.baseDir:         "<runtime>",
	} {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		value = strings.ReplaceAll(value, raw, replacement)
	}
	return value
}

func (r supportRedactor) redactPath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	base := filepath.Base(raw)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "<redacted>"
	}
	return "<redacted>/" + base
}

func (r supportRedactor) redactEndpoint(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return "<redacted-endpoint>"
}

func buildSupportSummaryBody(report string, now time.Time, window time.Duration) string {
	return strings.Join([]string{
		report,
		"",
		"[recent_window]",
		fmt.Sprintf("window=%s", SupportRecentWindowLabel(window)),
		fmt.Sprintf("start=%s", now.UTC().Add(-window).Format(time.RFC3339)),
		fmt.Sprintf("end=%s", now.UTC().Format(time.RFC3339)),
		"redaction=safe",
	}, "\n")
}

func joinSupportLines(lines []string) string {
	if len(lines) == 0 {
		return "none\n"
	}
	return strings.Join(lines, "\n") + "\n"
}

func writeZipEntry(zw *zip.Writer, name string, body []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

func readSupportErrorEntries(logRoot string, since, until time.Time) ([]string, error) {
	entries := []string{}
	for day := startOfUTCDay(since); !day.After(startOfUTCDay(until)); day = day.Add(24 * time.Hour) {
		path := filepath.Join(logRoot, day.Format("2006-01-02"), "error.log")
		body, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, entry := range parseSupportLogEntries(string(body)) {
			ts, ok := parseSupportLogTimestamp(entry)
			if !ok {
				continue
			}
			if ts.Before(since) || ts.After(until) {
				continue
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func parseSupportLogEntries(body string) []string {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	entries := []string{}
	current := []string{}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			if len(current) > 0 {
				entries = append(entries, strings.Join(current, "\n"))
			}
			current = []string{line}
			continue
		}
		if len(current) == 0 {
			current = []string{line}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		entries = append(entries, strings.Join(current, "\n"))
	}
	return entries
}

func parseSupportLogTimestamp(entry string) (time.Time, bool) {
	if !strings.HasPrefix(entry, "[") {
		return time.Time{}, false
	}
	end := strings.Index(entry, "]")
	if end <= 1 {
		return time.Time{}, false
	}
	ts, err := time.Parse(time.RFC3339Nano, entry[1:end])
	if err != nil {
		return time.Time{}, false
	}
	return ts.UTC(), true
}

func startOfUTCDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
