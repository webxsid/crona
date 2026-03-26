//go:build !windows

package runtime

import (
	"errors"
	"syscall"
)

func processExists(pid int) (bool, error) {
	if err := syscall.Kill(pid, 0); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
