//go:build !windows

package localipc

import (
	"net"
	"os"
	"time"
)

func Listen(endpoint string) (net.Listener, error) {
	if err := os.Remove(endpoint); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	ln, err := net.Listen("unix", endpoint)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(endpoint, 0o600); err != nil {
		_ = ln.Close()
		return nil, err
	}
	return ln, nil
}

func Dial(endpoint string, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: TimeoutOrDefault(timeout)}
	return dialer.Dial("unix", endpoint)
}

func CleanupEndpoint(endpoint string) error {
	if err := os.Remove(endpoint); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
