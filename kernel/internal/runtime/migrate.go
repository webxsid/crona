package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"crona/shared/config"
	sharedtypes "crona/shared/types"
)

func MigrateLegacyBaseDir(mode string) error {
	if runtimeMigrationSkipped() {
		return nil
	}

	target, err := config.RuntimeBaseDirForMode(mode)
	if err != nil {
		return err
	}
	legacy, err := config.LegacyRuntimeBaseDirForMode(mode)
	if err != nil {
		return err
	}
	if strings.TrimSpace(legacy) == "" || target == legacy {
		return nil
	}
	if !exists(legacy) {
		return nil
	}

	if running, err := legacyKernelRunning(legacy); err != nil {
		return err
	} else if running {
		return fmt.Errorf("found a running Crona kernel using %s; stop Crona and retry so runtime data can be migrated to %s", legacy, target)
	}

	if err := os.MkdirAll(filepath.Dir(target), dirPerm); err != nil {
		return err
	}
	if exists(target) {
		if !shouldMergeLegacyIntoTarget(legacy, target) {
			return nil
		}
		return mergeLegacyIntoTarget(legacy, target)
	}
	if err := os.Rename(legacy, target); err == nil {
		return nil
	} else if !isCrossDevice(err) {
		return err
	}

	return copyThenReplaceRuntimeDir(legacy, target)
}

func runtimeMigrationSkipped() bool {
	if config.RuntimeBaseDirOverride() != "" {
		return true
	}
	switch runtime.GOOS {
	case "darwin", "linux":
		return false
	default:
		return true
	}
}

func legacyKernelRunning(base string) (bool, error) {
	infoPath := filepath.Join(base, "kernel.json")
	body, err := os.ReadFile(infoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	var info sharedtypes.KernelInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return false, nil
	}
	if info.PID <= 0 {
		return false, nil
	}
	if err := syscall.Kill(info.PID, 0); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return false, nil
		}
		return false, err
	}

	cmd, err := exec.Command("ps", "-p", fmt.Sprintf("%d", info.PID), "-o", "command=").Output()
	if err != nil {
		return false, fmt.Errorf("could not inspect process %d referenced by %s: %w", info.PID, infoPath, err)
	}
	return strings.Contains(strings.TrimSpace(string(cmd)), "crona-kernel"), nil
}

func copyThenReplaceRuntimeDir(source, target string) error {
	parent := filepath.Dir(target)
	tmpDir, err := os.MkdirTemp(parent, filepath.Base(target)+".migrate-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	if err := copyDirContents(source, tmpDir); err != nil {
		return err
	}
	if err := os.Rename(tmpDir, target); err != nil {
		return err
	}
	if err := os.RemoveAll(source); err != nil {
		_ = os.RemoveAll(target)
		return err
	}
	return nil
}

func shouldMergeLegacyIntoTarget(source, target string) bool {
	return exists(filepath.Join(source, "crona.db")) && !exists(filepath.Join(target, "crona.db"))
}

func mergeLegacyIntoTarget(source, target string) error {
	if err := mergeDirContents(source, target); err != nil {
		return err
	}
	return os.RemoveAll(source)
}

func mergeDirContents(source, target string) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(target, dirPerm); err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		targetPath := filepath.Join(target, entry.Name())

		info, err := os.Lstat(sourcePath)
		if err != nil {
			return err
		}
		if !exists(targetPath) {
			if err := os.Rename(sourcePath, targetPath); err == nil {
				continue
			} else if !isCrossDevice(err) {
				return err
			}
			if info.IsDir() {
				if err := copyThenReplaceRuntimeDir(sourcePath, targetPath); err != nil {
					return err
				}
				continue
			}
			if info.Mode()&os.ModeSymlink != 0 {
				linkTarget, err := os.Readlink(sourcePath)
				if err != nil {
					return err
				}
				if err := os.Symlink(linkTarget, targetPath); err != nil {
					return err
				}
				if err := os.Remove(sourcePath); err != nil {
					return err
				}
				continue
			}
			if err := copyFile(sourcePath, targetPath, info.Mode().Perm()); err != nil {
				return err
			}
			if err := os.Remove(sourcePath); err != nil {
				return err
			}
			continue
		}

		targetInfo, err := os.Lstat(targetPath)
		if err != nil {
			return err
		}
		if info.IsDir() && targetInfo.IsDir() {
			if err := mergeDirContents(sourcePath, targetPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyDirContents(source, target string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", source)
	}
	if err := os.Chmod(target, info.Mode().Perm()); err != nil {
		return err
	}

	return filepath.Walk(source, func(path string, walkInfo os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == source {
			return nil
		}

		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(target, rel)
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}

		switch {
		case info.Mode()&os.ModeSymlink != 0:
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(linkTarget, dest)
		case info.IsDir():
			return os.MkdirAll(dest, info.Mode().Perm())
		default:
			return copyFile(path, dest, info.Mode().Perm())
		}
	})
}

func copyFile(source, target string, perm os.FileMode) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isCrossDevice(err error) bool {
	return errors.Is(err, syscall.EXDEV)
}
