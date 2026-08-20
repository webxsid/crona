package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoggerAppendsEntries(t *testing.T) {
	dir := t.TempDir()
	logger := NewLogger(Paths{CurrentLogDir: dir})

	logger.Info("first")
	logger.Info("second")
	logger.Error("third", os.ErrClosed)

	body, err := os.ReadFile(filepath.Join(dir, "info.log"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{"first", "second", "third"} {
		if !strings.Contains(text, want) {
			t.Fatalf("log missing %q: %s", want, text)
		}
	}
	if strings.Index(text, "first") > strings.Index(text, "second") ||
		strings.Index(text, "second") > strings.Index(text, "third") {
		t.Fatalf("log entries are out of order: %s", text)
	}
}
