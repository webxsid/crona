//go:build windows

package localipc

import (
	"context"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

func Listen(endpoint string) (net.Listener, error) {
	return winio.ListenPipe(endpoint, nil)
}

func Dial(endpoint string, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutOrDefault(timeout))
	defer cancel()
	return winio.DialPipeContext(ctx, endpoint)
}

func CleanupEndpoint(endpoint string) error {
	return nil
}
