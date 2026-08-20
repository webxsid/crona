package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"
)

type Logger struct {
	infoPath  string
	errorPath string
	mu        sync.Mutex
}

func NewLogger(paths Paths) *Logger {
	return &Logger{
		infoPath:  filepath.Join(paths.CurrentLogDir, "info.log"),
		errorPath: filepath.Join(paths.CurrentLogDir, "error.log"),
	}
}

func (l *Logger) Info(msg string) {
	l.write("INFO", msg, "")
}

func (l *Logger) Error(msg string, err error) {
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	l.write("ERROR", msg, detail)
}

// Go runs fn in a daemon-owned goroutine and records panics without allowing
// a worker failure to terminate the daemon process.
func (l *Logger) Go(name string, fn func()) {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				l.Error("panic in "+name, fmt.Errorf("%v\n%s", recovered, debug.Stack()))
			}
		}()
		fn()
	}()
}

func (l *Logger) write(level, msg, detail string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := fmt.Sprintf("[%s] [%s] %s", time.Now().Format(time.RFC3339), level, msg)
	if detail != "" {
		entry += "\n  Detail: " + detail
	}
	entry += "\n"

	if err := appendFile(l.infoPath, entry); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "crona logger: write info log: %v\n", err)
	}
	if level == "ERROR" {
		if err := appendFile(l.errorPath, entry); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "crona logger: write error log: %v\n", err)
		}
	}
}

func appendFile(path, entry string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, FilePerm())
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	_, err = f.WriteString(entry)
	return err
}
