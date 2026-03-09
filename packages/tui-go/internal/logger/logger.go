package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var logBase string

func init() {
	home, _ := os.UserHomeDir()
	logBase = filepath.Join(home, ".crona", "logs", "tui-go")
}

func logDir() string {
	date := time.Now().UTC().Format("2006-01-02")
	dir := filepath.Join(logBase, date)
	_ = os.MkdirAll(dir, 0o700)
	return dir
}

func write(level, msg string) {
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	entry := fmt.Sprintf("[%s] [%s] %s\n", ts, level, msg)
	dir := logDir()

	if f, err := os.OpenFile(filepath.Join(dir, "info.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); err == nil {
		_, _ = f.WriteString(entry)
		f.Close()
	}
	if level == "ERROR" {
		if f, err := os.OpenFile(filepath.Join(dir, "error.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); err == nil {
			_, _ = f.WriteString(entry)
			f.Close()
		}
	}
}

func Info(msg string)                        { write("INFO", msg) }
func Infof(format string, args ...any)       { write("INFO", fmt.Sprintf(format, args...)) }
func Error(msg string)                       { write("ERROR", msg) }
func Errorf(format string, args ...any)      { write("ERROR", fmt.Sprintf(format, args...)) }
