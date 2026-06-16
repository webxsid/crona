package migration

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	flagspkg "crona/cli/internal/flags"
	"crona/shared/config"
)

type Deps struct {
	Stdout         io.Writer
	Stdin          io.Reader
	ShutdownKernel func() error
	RuntimeBaseDir func() (string, error)
	BackupBaseDir  func() (string, error)
	Now            func() time.Time
	IsInteractive  func() bool
}

func UsageBackup() string {
	return "Usage: crona backup [--json]\n"
}

func UsageRestore() string {
	return "Usage: crona restore <path> [--json]\n"
}

func RunBackup(args []string, deps Deps) error {
	deps = deps.withDefaults()
	if len(args) > 0 && flagspkg.IsHelpArg(args[0]) {
		_, err := fmt.Fprint(deps.Stdout, UsageBackup())
		return err
	}
	fs := flagspkg.New("backup")
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("usage: %s", strings.TrimSpace(UsageBackup()))
	}

	source, err := runtimeDBPath(deps.RuntimeBaseDir)
	if err != nil {
		return err
	}
	if err := deps.ShutdownKernel(); err != nil {
		return err
	}
	backupPath, err := createBackup(source, deps.BackupBaseDir, deps.Now)
	if err != nil {
		return err
	}
	if *jsonOut {
		return printJSON(deps.Stdout, map[string]any{"backupPath": backupPath})
	}
	_, err = fmt.Fprintf(deps.Stdout, "backup file: %s\n", backupPath)
	return err
}

func RunRestore(args []string, deps Deps) error {
	deps = deps.withDefaults()
	if len(args) == 0 || flagspkg.IsHelpArg(args[0]) {
		_, err := fmt.Fprint(deps.Stdout, UsageRestore())
		return err
	}
	sourceArg := strings.TrimSpace(args[0])
	if sourceArg == "" {
		return fmt.Errorf("usage: %s", strings.TrimSpace(UsageRestore()))
	}
	fs := flagspkg.New("restore")
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("usage: %s", strings.TrimSpace(UsageRestore()))
	}

	source, err := filepath.Abs(sourceArg)
	if err != nil {
		return err
	}
	target, err := runtimeDBPath(deps.RuntimeBaseDir)
	if err != nil {
		return err
	}
	if err := deps.ShutdownKernel(); err != nil {
		return err
	}
	if _, err := os.Stat(target); err == nil {
		confirm, err := confirmRestoreOverwrite(deps)
		if err != nil {
			return err
		}
		if !confirm {
			return errors.New("restore cancelled")
		}
	}
	if err := restoreDatabase(source, target); err != nil {
		return err
	}
	if *jsonOut {
		return printJSON(deps.Stdout, map[string]any{"restoredPath": target, "sourcePath": source})
	}
	_, err = fmt.Fprintf(deps.Stdout, "restored database: %s\n", target)
	return err
}

func (d Deps) withDefaults() Deps {
	if d.Stdout == nil {
		d.Stdout = os.Stdout
	}
	if d.Stdin == nil {
		d.Stdin = os.Stdin
	}
	if d.RuntimeBaseDir == nil {
		d.RuntimeBaseDir = config.RuntimeBaseDir
	}
	if d.BackupBaseDir == nil {
		d.BackupBaseDir = config.BackupBaseDir
	}
	if d.Now == nil {
		d.Now = time.Now
	}
	if d.IsInteractive == nil {
		d.IsInteractive = defaultInteractive
	}
	if d.ShutdownKernel == nil {
		d.ShutdownKernel = func() error { return nil }
	}
	return d
}

func runtimeDBPath(resolver func() (string, error)) (string, error) {
	base, err := resolver()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "crona.db"), nil
}

func createBackup(source string, resolver func() (string, error), now func() time.Time) (string, error) {
	info, err := os.Stat(source)
	if err != nil {
		return "", fmt.Errorf("read runtime database: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory", source)
	}
	base, err := resolver()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(base, 0o700); err != nil {
		return "", err
	}
	timestamp := now().UTC().Format("20060102-150405")
	for i := 0; i < 1000; i++ {
		name := fmt.Sprintf("crona-db-%s.db", timestamp)
		if i > 0 {
			name = fmt.Sprintf("crona-db-%s-%d.db", timestamp, i+1)
		}
		target := filepath.Join(base, name)
		if _, err := os.Stat(target); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return "", err
		}
		if err := copyFileAtomic(source, target, info.Mode().Perm()); err != nil {
			return "", err
		}
		return target, nil
	}
	return "", fmt.Errorf("could not allocate a unique backup path in %s", base)
}

func restoreDatabase(source, target string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("read backup: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", source)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	return copyFileAtomic(source, target, info.Mode().Perm())
}

func copyFileAtomic(source, target string, perm os.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() { _ = input.Close() }()

	targetDir := filepath.Dir(target)
	if err := os.MkdirAll(targetDir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(targetDir, filepath.Base(target)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	if _, err := io.Copy(tmp, input); err != nil {
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if _, err := os.Stat(target); err == nil {
		if err := os.Remove(target); err != nil {
			return err
		}
	}
	return os.Rename(tmpName, target)
}

func confirmRestoreOverwrite(deps Deps) (bool, error) {
	if !deps.IsInteractive() {
		return false, errors.New("restore requires an interactive terminal when overwriting the current database")
	}
	if _, err := fmt.Fprintf(
		deps.Stdout,
		"Existing database found. Overwrite it with the backup? [y/N]: ",
	); err != nil {
		return false, err
	}
	reader := bufio.NewReader(deps.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	switch answer {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func defaultInteractive() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func printJSON(w io.Writer, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(body))
	return err
}
