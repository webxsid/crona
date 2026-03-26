package localipc

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	TransportUnixSocket       = "unix_socket"
	TransportWindowsNamedPipe = "windows_named_pipe"
)

func DefaultTransport() string {
	if runtime.GOOS == "windows" {
		return TransportWindowsNamedPipe
	}
	return TransportUnixSocket
}

func DefaultEndpoint(baseDir, mode string) string {
	if DefaultTransport() == TransportWindowsNamedPipe {
		return windowsPipeEndpoint(mode)
	}
	return filepath.Join(baseDir, "kernel.sock")
}

func TimeoutOrDefault(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}
	return 5 * time.Second
}

func Label(transport, endpoint string) string {
	if strings.TrimSpace(transport) == "" {
		transport = DefaultTransport()
	}
	switch transport {
	case TransportWindowsNamedPipe:
		return "named pipe " + endpoint
	case TransportUnixSocket:
		return "unix socket " + endpoint
	default:
		return fmt.Sprintf("%s %s", transport, endpoint)
	}
}

func windowsPipeEndpoint(mode string) string {
	name := "crona-kernel"
	if strings.EqualFold(strings.TrimSpace(mode), "dev") {
		name += "-dev"
	}
	return `\\.\pipe\` + name
}
