package sessionmeta

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"

	"github.com/charmbracelet/lipgloss"
)

const (
	clockLitRune = '█'
)

type clockTileSpec struct {
	scale int
}

var clockTileSpecs = []clockTileSpec{
	{scale: 1},
	{scale: 2},
	{scale: 3},
}

var clockGlyphRows = map[rune][]string{
	'0': {
		"###",
		"# #",
		"# #",
		"# #",
		"###",
	},
	'1': {
		" # ",
		"## ",
		" # ",
		" # ",
		"###",
	},
	'2': {
		"###",
		"  #",
		"###",
		"#  ",
		"###",
	},
	'3': {
		"###",
		"  #",
		"###",
		"  #",
		"###",
	},
	'4': {
		"# #",
		"# #",
		"###",
		"  #",
		"  #",
	},
	'5': {
		"###",
		"#  ",
		"###",
		"  #",
		"###",
	},
	'6': {
		"###",
		"#  ",
		"###",
		"# #",
		"###",
	},
	'7': {
		"###",
		"  #",
		"  #",
		"  #",
		"  #",
	},
	'8': {
		"###",
		"# #",
		"###",
		"# #",
		"###",
	},
	'9': {
		"###",
		"# #",
		"###",
		"  #",
		"###",
	},
	':': {
		"   ",
		" # ",
		"   ",
		" # ",
		"   ",
	},
}

func FormatEstimateProgress(elapsedSeconds, estimateMinutes int) string {
	return fmt.Sprintf("%s / %s", viewhelpers.FormatClockText(elapsedSeconds), helperpkg.FormatCompactDurationMinutes(estimateMinutes))
}

func SessionHistorySummary(entry api.SessionHistoryEntry) string {
	if entry.ParsedNotes != nil {
		if m := strings.TrimSpace(entry.ParsedNotes[sharedtypes.SessionNoteSectionCommit]); m != "" {
			return m
		}
		if n := strings.TrimSpace(entry.ParsedNotes[sharedtypes.SessionNoteSectionNotes]); n != "" {
			return n
		}
	}
	if entry.Notes != nil && strings.TrimSpace(*entry.Notes) != "" {
		return strings.TrimSpace(*entry.Notes)
	}
	return fmt.Sprintf("Issue #%d", entry.IssueID)
}

func FormatSessionTimestamp(value string) string {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Local().Format("2006-01-02 15:04")
	}
	if len(value) >= 16 {
		return strings.Replace(value[:16], "T", " ", 1)
	}
	return value
}

func FormatSessionDuration(durationSeconds *int, start string, end *string) string {
	if durationSeconds != nil {
		return viewhelpers.FormatClockText(*durationSeconds)
	}
	if end != nil && *end != "" {
		st, se := time.Parse(time.RFC3339, start)
		et, ee := time.Parse(time.RFC3339, *end)
		if se == nil && ee == nil {
			return viewhelpers.FormatClockText(int(et.Sub(st).Seconds()))
		}
	}
	return "-"
}

func SummarizeCompletedSessions(s []api.Session) (workedSeconds int, completedCount int) {
	for _, session := range s {
		if session.DurationSeconds == nil || session.EndTime == nil {
			continue
		}
		workedSeconds += *session.DurationSeconds
		completedCount++
	}
	return
}

func RenderClockTiny(clock string) string {
	return strings.TrimSpace(clock)
}

func RenderClockSmall(clock string) string {
	return renderClockCanvas(clock, clockTileSpecs[0], lipgloss.Color(""))
}

func RenderClockMedium(clock string) string {
	return renderClockCanvas(clock, clockTileSpecs[1], lipgloss.Color(""))
}

func RenderClockLarge(clock string) string {
	return renderClockCanvas(clock, clockTileSpecs[2], lipgloss.Color(""))
}

func RenderResponsiveClock(clock string, width, height int, litColor, dimColor lipgloss.Color) string {
	for i := len(clockTileSpecs) - 1; i >= 0; i-- {
		spec := clockTileSpecs[i]
		if spec.fits(clock, width, height) {
			return renderClockCanvas(clock, spec, litColor)
		}
	}
	return RenderClockTiny(clock)
}

func RenderBigClock(clock string) string {
	return RenderClockLarge(clock)
}

func renderClockCanvas(clock string, spec clockTileSpec, litColor lipgloss.Color) string {
	runes := []rune(strings.TrimSpace(clock))
	if len(runes) == 0 {
		return ""
	}
	width := spec.canvasWidth(len(runes))
	height := spec.canvasHeight()
	rows := make([][]bool, height)
	for y := range rows {
		rows[y] = make([]bool, width)
	}
	for i, char := range runes {
		spec.paint(rows, i*spec.advance(), char)
	}
	out := make([]string, height)
	litStyle := lipgloss.NewStyle()
	if litColor != "" {
		litStyle = litStyle.Foreground(litColor).Bold(true)
	}
	for y := range rows {
		var line strings.Builder
		line.Grow(len(rows[y]))
		for _, cell := range rows[y] {
			if cell {
				line.WriteString(litStyle.Render(string(clockLitRune)))
				continue
			}
			line.WriteByte(' ')
		}
		out[y] = line.String()
	}
	return strings.Join(out, "\n")
}

func (s clockTileSpec) fits(clock string, width, height int) bool {
	if width <= 0 || height <= 0 {
		return false
	}
	widthNeeded := s.canvasWidth(len([]rune(strings.TrimSpace(clock))))
	if widthNeeded > width {
		return false
	}
	return s.canvasHeight() <= height
}

func (s clockTileSpec) canvasWidth(chars int) int {
	if chars <= 0 {
		return 0
	}
	return chars*s.glyphWidth() + (chars-1)*s.glyphGap()
}

func (s clockTileSpec) canvasHeight() int {
	return s.glyphHeight()
}

func (s clockTileSpec) glyphWidth() int {
	return 3 * s.scale
}

func (s clockTileSpec) glyphHeight() int {
	return 5 * s.scale
}

func (s clockTileSpec) glyphGap() int {
	if s.scale < 1 {
		return 1
	}
	return s.scale
}

func (s clockTileSpec) advance() int {
	return s.glyphWidth() + s.glyphGap()
}

func (s clockTileSpec) paint(canvas [][]bool, xOff int, char rune) {
	if char == ':' {
		s.paintColon(canvas, xOff)
		return
	}
	rows, ok := clockGlyphRows[char]
	if !ok {
		return
	}
	s.paintBitmap(canvas, xOff, rows)
}

func (s clockTileSpec) paintColon(canvas [][]bool, xOff int) {
	s.paintBitmap(canvas, xOff, clockGlyphRows[':'])
}

func (s clockTileSpec) paintBitmap(canvas [][]bool, xOff int, rows []string) {
	scale := s.scale
	if scale < 1 {
		scale = 1
	}
	for y, row := range rows {
		for x, cell := range row {
			if cell != '#' {
				continue
			}
			for yy := 0; yy < scale; yy++ {
				canvasY := y*scale + yy
				if canvasY < 0 || canvasY >= len(canvas) {
					continue
				}
				for xx := 0; xx < scale; xx++ {
					canvasX := xOff + x*scale + xx
					if canvasX < 0 || canvasX >= len(canvas[canvasY]) {
						continue
					}
					canvas[canvasY][canvasX] = true
				}
			}
		}
	}
}

func IssueMetaByID(all []api.IssueWithMeta, issueID int64) *api.IssueWithMeta {
	for i := range all {
		if all[i].ID == issueID {
			return &all[i]
		}
	}
	return nil
}

func FilteredSessionIndices(entries []api.SessionHistoryEntry, filter string) []int {
	filter = normalizeFilter(filter)
	out := []int{}
	for i, entry := range entries {
		text := strings.ToLower(SessionHistorySummary(entry))
		if filter == "" || strings.Contains(text, filter) {
			out = append(out, i)
		}
	}
	return out
}

func normalizeFilter(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}
