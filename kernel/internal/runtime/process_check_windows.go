//go:build windows

package runtime

func processExists(pid int) (bool, error) {
	return false, nil
}
