package terminaltitle

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

const maxTitleRunes = 80

func Command(title string) tea.Cmd {
	return func() tea.Msg {
		_ = Write(os.Stdout, title)
		return nil
	}
}

func Write(w io.Writer, title string) error {
	_, err := fmt.Fprint(w, Sequence(title))
	return err
}

func Reset(w io.Writer) error {
	return Write(w, "")
}

func Sequence(title string) string {
	return "\x1b]0;" + Sanitize(title) + "\a"
}

func Sanitize(title string) string {
	title = strings.TrimSpace(title)
	out := make([]rune, 0, len(title))
	for _, r := range title {
		if r == '\x1b' || r == '\a' || r == '\n' || r == '\r' || unicode.IsControl(r) {
			continue
		}
		out = append(out, r)
	}
	if len(out) <= maxTitleRunes {
		return string(out)
	}
	return string(out[:maxTitleRunes-3]) + "..."
}
