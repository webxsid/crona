package posthog

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	sharedconfig "crona/shared/config"
	"crona/shared/protocol"
	versionpkg "crona/shared/version"

	vendorposthog "github.com/posthog/posthog-go"
)

const (
	identityFilename = "telemetry.json"
	dirPerm          = 0o700
	filePerm         = 0o600
	logDirName       = "posthog"

	EventTUIStarted    = "tui_started"
	EventDaemonStarted = "daemon_started"
	EventDaemonStopped = "daemon_stopped"
	EventErrorReported = "error_reported"
)

type Config struct {
	APIKey                string
	Host                  string
	Enabled               bool
	UsageEnabled          bool
	ErrorReportingEnabled bool
	App                   string
	Version               string
	Mode                  string
	RuntimeDir            string
}

type Properties map[string]any

type Client interface {
	Enabled() bool
	UsageEnabled() bool
	ErrorReportingEnabled() bool
	DistinctID() string
	Capture(event string, properties Properties) error
	Identify(properties Properties) error
	ReportError(kind string, err error, properties Properties) error
	Flush() error
	Close() error
}

type client struct {
	inner                 vendorposthog.Client
	distinctID            string
	usageEnabled          bool
	errorReportingEnabled bool
	logger                *attemptLogger
	app                   string
	host                  string
}

type noopClient struct {
	distinctID  string
	logger      *attemptLogger
	app         string
	host        string
	usageReason string
	errorReason string
}

type identityState struct {
	DistinctID string `json:"distinctId"`
	CreatedAt  string `json:"createdAt"`
}

type noopLogger struct{}

type attemptLogger struct {
	path string
	mu   sync.Mutex
}

func LoadConfig(app string) Config {
	env := sharedconfig.Current()
	return Config{
		APIKey:                sharedconfig.PostHogAPIKey(),
		Host:                  sharedconfig.PostHogHost(),
		Enabled:               sharedconfig.PostHogEnabled(),
		UsageEnabled:          sharedconfig.PostHogEnabled(),
		ErrorReportingEnabled: sharedconfig.PostHogEnabled(),
		App:                   strings.TrimSpace(app),
		Version:               versionpkg.Current(),
		Mode:                  env.Mode,
	}
}

func New(cfg Config) (Client, error) {
	cfg = normalizeConfig(cfg)
	logger, err := newAttemptLogger(cfg.RuntimeDir)
	if err != nil {
		return noopClient{}, err
	}
	if !cfg.UsageEnabled && !cfg.ErrorReportingEnabled {
		return noopClient{
			logger:      logger,
			app:         cfg.App,
			host:        cfg.Host,
			usageReason: "disabled",
			errorReason: "disabled",
		}, nil
	}
	if cfg.APIKey == "" {
		return noopClient{
			logger:      logger,
			app:         cfg.App,
			host:        cfg.Host,
			usageReason: "missing_api_key",
			errorReason: "missing_api_key",
		}, nil
	}

	state, err := loadOrCreateIdentity(cfg.RuntimeDir)
	if err != nil {
		return noopClient{}, err
	}

	disableGeoIP := true
	inner, err := vendorposthog.NewWithConfig(cfg.APIKey, vendorposthog.Config{
		Endpoint:     cfg.Host,
		Logger:       noopLogger{},
		DisableGeoIP: &disableGeoIP,
		DefaultEventProperties: vendorposthog.Properties{
			"app":         cfg.App,
			"app_version": cfg.Version,
			"env_mode":    cfg.Mode,
			"os":          runtime.GOOS,
			"goos":        runtime.GOOS,
			"arch":        runtime.GOARCH,
		},
	})
	if err != nil {
		return noopClient{}, err
	}

	return &client{
		inner:                 inner,
		distinctID:            state.DistinctID,
		usageEnabled:          cfg.UsageEnabled,
		errorReportingEnabled: cfg.ErrorReportingEnabled,
		logger:                logger,
		app:                   cfg.App,
		host:                  cfg.Host,
	}, nil
}

func (c *client) Enabled() bool {
	return c != nil && (c.usageEnabled || c.errorReportingEnabled)
}

func (c *client) UsageEnabled() bool {
	return c != nil && c.usageEnabled
}

func (c *client) ErrorReportingEnabled() bool {
	return c != nil && c.errorReportingEnabled
}

func (c *client) DistinctID() string {
	if c == nil {
		return ""
	}
	return c.distinctID
}

func (c *client) Capture(event string, properties Properties) error {
	if c == nil || !c.usageEnabled {
		if c != nil {
			c.logAttempt("capture", strings.TrimSpace(event), properties, "skipped", "usage_disabled")
		}
		return nil
	}
	event = strings.TrimSpace(event)
	c.logAttempt("capture", event, properties, "attempted", "")
	err := c.inner.Enqueue(vendorposthog.Capture{
		DistinctId: c.distinctID,
		Event:      event,
		Properties: vendorposthog.Properties(properties),
	})
	if err != nil {
		c.logAttempt("capture", event, properties, "failed", err.Error())
	}
	return err
}

func (c *client) Identify(properties Properties) error {
	if c == nil || !c.usageEnabled {
		if c != nil {
			c.logAttempt("identify", "", properties, "skipped", "usage_disabled")
		}
		return nil
	}
	c.logAttempt("identify", "", properties, "attempted", "")
	err := c.inner.Enqueue(vendorposthog.Identify{
		DistinctId: c.distinctID,
		Properties: vendorposthog.Properties(properties),
	})
	if err != nil {
		c.logAttempt("identify", "", properties, "failed", err.Error())
	}
	return err
}

func (c *client) ReportError(kind string, err error, properties Properties) error {
	if c == nil || !c.errorReportingEnabled {
		if c != nil {
			c.logAttempt("report_error", EventErrorReported, properties, "skipped", "error_reporting_disabled")
		}
		return nil
	}
	props := sanitizedErrorProperties(kind, err, properties)
	c.logAttempt("report_error", EventErrorReported, props, "attempted", "")
	enqueueErr := c.inner.Enqueue(vendorposthog.Capture{
		DistinctId: c.distinctID,
		Event:      EventErrorReported,
		Properties: vendorposthog.Properties(props),
	})
	if enqueueErr != nil {
		c.logAttempt("report_error", EventErrorReported, props, "failed", enqueueErr.Error())
	}
	return enqueueErr
}

func (c *client) Flush() error {
	if c == nil || c.inner == nil {
		return nil
	}
	c.logAttempt("flush", "", nil, "attempted", "")
	err := c.inner.Close()
	if err != nil {
		c.logAttempt("flush", "", nil, "failed", err.Error())
		return err
	}
	c.logAttempt("flush", "", nil, "closed", "")
	return nil
}

func (c *client) Close() error {
	if c == nil || c.inner == nil {
		return nil
	}
	c.logAttempt("close", "", nil, "attempted", "")
	err := c.inner.Close()
	if err != nil {
		c.logAttempt("close", "", nil, "failed", err.Error())
		return err
	}
	c.logAttempt("close", "", nil, "closed", "")
	return nil
}

func (n noopClient) Enabled() bool               { return false }
func (n noopClient) UsageEnabled() bool          { return false }
func (n noopClient) ErrorReportingEnabled() bool { return false }
func (n noopClient) DistinctID() string          { return n.distinctID }
func (n noopClient) Capture(event string, properties Properties) error {
	n.logSkip("capture", strings.TrimSpace(event), properties, n.usageReason)
	return nil
}
func (n noopClient) Identify(properties Properties) error {
	n.logSkip("identify", "", properties, n.usageReason)
	return nil
}
func (n noopClient) ReportError(kind string, err error, properties Properties) error {
	props := sanitizedErrorProperties(kind, err, properties)
	n.logSkip("report_error", EventErrorReported, props, n.errorReason)
	return nil
}
func (n noopClient) Flush() error {
	if n.logger != nil {
		n.logger.log(map[string]string{
			"app":       n.app,
			"enabled":   "false",
			"host":      n.host,
			"operation": "flush",
			"reason":    n.errorReason,
			"result":    "skipped",
		})
	}
	return nil
}
func (n noopClient) Close() error {
	if n.logger != nil {
		n.logger.log(map[string]string{
			"app":       n.app,
			"enabled":   "false",
			"host":      n.host,
			"operation": "close",
			"reason":    skipReason(n.usageReason, n.errorReason),
			"result":    "skipped",
		})
	}
	return nil
}
func (noopLogger) Debugf(string, ...any) {}
func (noopLogger) Logf(string, ...any)   {}
func (noopLogger) Warnf(string, ...any)  {}
func (noopLogger) Errorf(string, ...any) {}

func (c *client) logAttempt(operation, event string, properties Properties, result, detail string) {
	if c == nil || c.logger == nil {
		return
	}
	fields := map[string]string{
		"app":            c.app,
		"distinct_id":    c.distinctID,
		"enabled":        "true",
		"event":          event,
		"host":           c.host,
		"operation":      operation,
		"property_count": fmt.Sprintf("%d", len(properties)),
		"result":         result,
	}
	if keys := propertyKeys(properties); keys != "" {
		fields["property_keys"] = keys
	}
	if detail != "" {
		fields["error"] = detail
	}
	c.logger.log(fields)
}

func (n noopClient) logSkip(operation, event string, properties Properties, reason string) {
	if n.logger == nil {
		return
	}
	fields := map[string]string{
		"app":            n.app,
		"enabled":        "false",
		"event":          event,
		"host":           n.host,
		"operation":      operation,
		"property_count": fmt.Sprintf("%d", len(properties)),
		"reason":         reason,
		"result":         "skipped",
	}
	if keys := propertyKeys(properties); keys != "" {
		fields["property_keys"] = keys
	}
	n.logger.log(fields)
}

func normalizeConfig(cfg Config) Config {
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.Host = strings.TrimSpace(cfg.Host)
	cfg.App = strings.TrimSpace(cfg.App)
	cfg.Version = strings.TrimSpace(cfg.Version)
	cfg.Mode = strings.TrimSpace(cfg.Mode)
	cfg.RuntimeDir = strings.TrimSpace(cfg.RuntimeDir)
	if !cfg.Enabled {
		cfg.UsageEnabled = false
		cfg.ErrorReportingEnabled = false
	}
	if cfg.RuntimeDir == "" {
		if base, err := sharedconfig.RuntimeBaseDir(); err == nil {
			cfg.RuntimeDir = base
		}
	}
	return cfg
}

func sanitizedErrorProperties(kind string, err error, properties Properties) Properties {
	out := Properties{
		"error_kind": normalizeErrorKind(kind),
	}
	classification, message, rpcCode := classifyError(err)
	out["error_class"] = classification
	out["message"] = sanitizeErrorMessage(message)
	if rpcCode != "" {
		out["rpc_code"] = rpcCode
	}
	for key, value := range properties {
		switch strings.TrimSpace(key) {
		case "operation", "entrypoint":
			if text := sanitizePropertyText(value); text != "" {
				out[key] = text
			}
		}
	}
	return out
}

func normalizeErrorKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "panic":
		return "panic"
	default:
		return "handled"
	}
}

func classifyError(err error) (class string, message string, rpcCode string) {
	if err == nil {
		return "unknown", "unknown error", ""
	}
	var rpcErr *protocol.RPCError
	if errors.As(err, &rpcErr) {
		return "rpc_error", rpcErr.Message, strings.TrimSpace(rpcErr.Code)
	}
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return "io_error", pathErr.Err.Error(), ""
	}
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return "io_error", netErr.Err.Error(), ""
	}
	msg := strings.TrimSpace(err.Error())
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "required") || strings.Contains(lower, "invalid") {
		return "validation_error", msg, ""
	}
	if strings.Contains(lower, "timeout") || strings.Contains(lower, "timed out") {
		return "io_error", msg, ""
	}
	return "unknown", msg, ""
}

func sanitizeErrorMessage(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return "unknown error"
	}
	if strings.Contains(message, "/") || strings.Contains(message, `\`) || strings.Contains(message, "~") {
		return "redacted path-bearing error"
	}
	if len(message) > 160 {
		message = strings.TrimSpace(message[:160])
	}
	return message
}

func sanitizePropertyText(value any) string {
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" {
		return ""
	}
	if strings.Contains(text, "/") || strings.Contains(text, `\`) {
		return ""
	}
	if len(text) > 64 {
		text = text[:64]
	}
	return text
}

func skipReason(primary, secondary string) string {
	if primary == secondary || secondary == "" {
		return primary
	}
	if primary == "" {
		return secondary
	}
	return primary + "," + secondary
}

func newAttemptLogger(runtimeDir string) (*attemptLogger, error) {
	runtimeDir = strings.TrimSpace(runtimeDir)
	if runtimeDir == "" {
		base, err := sharedconfig.RuntimeBaseDir()
		if err != nil {
			return nil, fmt.Errorf("resolve telemetry log runtime dir: %w", err)
		}
		runtimeDir = base
	}
	dir := filepath.Join(runtimeDir, "logs", logDirName)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return nil, fmt.Errorf("ensure telemetry logs dir: %w", err)
	}
	return &attemptLogger{
		path: filepath.Join(dir, time.Now().UTC().Format("2006-01-02")+".log"),
	}, nil
}

func loadOrCreateIdentity(runtimeDir string) (identityState, error) {
	runtimeDir = strings.TrimSpace(runtimeDir)
	if runtimeDir == "" {
		return identityState{}, fmt.Errorf("resolve telemetry runtime dir: empty runtime dir")
	}
	if err := os.MkdirAll(runtimeDir, dirPerm); err != nil {
		return identityState{}, fmt.Errorf("ensure telemetry runtime dir: %w", err)
	}

	path := filepath.Join(runtimeDir, identityFilename)
	body, err := os.ReadFile(path)
	if err == nil {
		var state identityState
		if jsonErr := json.Unmarshal(body, &state); jsonErr == nil && strings.TrimSpace(state.DistinctID) != "" {
			return state, nil
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return identityState{}, fmt.Errorf("read telemetry identity: %w", err)
	}

	state, err := newIdentityState()
	if err != nil {
		return identityState{}, err
	}
	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return identityState{}, fmt.Errorf("encode telemetry identity: %w", err)
	}
	if err := os.WriteFile(path, encoded, filePerm); err != nil {
		return identityState{}, fmt.Errorf("write telemetry identity: %w", err)
	}
	return state, nil
}

func newIdentityState() (identityState, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return identityState{}, fmt.Errorf("generate telemetry identity: %w", err)
	}
	return identityState{
		DistinctID: "anon_" + hex.EncodeToString(buf),
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (l *attemptLogger) log(fields map[string]string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	keys := make([]string, 0, len(fields))
	for key, value := range fields {
		if strings.TrimSpace(value) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys)+1)
	parts = append(parts, fmt.Sprintf("ts=%s", time.Now().UTC().Format(time.RFC3339Nano)))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, sanitizeLogValue(fields[key])))
	}
	entry := strings.Join(parts, " ") + "\n"

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePerm)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.WriteString(entry)
}

func propertyKeys(properties Properties) string {
	if len(properties) == 0 {
		return ""
	}
	keys := make([]string, 0, len(properties))
	for key := range properties {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

func sanitizeLogValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "\\r")
	return value
}
