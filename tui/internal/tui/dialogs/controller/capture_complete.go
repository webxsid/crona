package controller

import (
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// CaptureSuggestion is one popup entry while typing an annotation in Add issues: Insert
// replaces the word at the cursor, Detail explains it.
type CaptureSuggestion struct {
	Insert string
	Detail string
}

const captureSuggestLimit = 6

var captureEstimates = []string{"15m", "30m", "45m", "1h", "1h30m", "2h", "3h", "4h"}

// captureToken is the annotation word ending at the cursor on a `- ` line, if any.
func captureToken(state State) (string, bool) {
	lines := strings.Split(state.Description.Value(), "\n")
	row := state.Description.Line()
	if row >= len(lines) || !strings.HasPrefix(strings.TrimSpace(lines[row]), "- ") {
		return "", false
	}
	line := []rune(lines[row])
	info := state.Description.LineInfo()
	col := min(len(line), info.StartColumn+info.ColumnOffset)
	if col < len(line) && line[col] != ' ' {
		return "", false
	}
	start := col
	for start > 0 && line[start-1] != ' ' {
		start--
	}
	word := string(line[start:col])
	if word == "" || !strings.ContainsRune("#+~@", rune(word[0])) || start < 2 {
		return "", false
	}
	return word, true
}

// captureSuggestions lists what the word at the cursor could become, best match first.
func captureSuggestions(state State, ctx UpdateContext) []CaptureSuggestion {
	word, ok := captureToken(state)
	if !ok {
		return nil
	}
	trigger, typed := word[:1], strings.ToLower(word[1:])
	var out []CaptureSuggestion
	add := func(insert, detail string, keys ...string) {
		for _, key := range keys {
			if strings.HasPrefix(strings.ToLower(key), typed) {
				out = append(out, CaptureSuggestion{Insert: trigger + insert, Detail: detail})
				return
			}
		}
	}
	switch trigger {
	case "#":
		streams := ctx.AllStreams
		if streams == nil {
			streams = ctx.Streams
		}
		for _, repo := range ctx.Repos {
			for _, stream := range streams {
				if stream.RepoID != repo.ID || strings.ContainsAny(repo.Name+stream.Name, " \t") {
					continue
				}
				path := repo.Name + "/" + stream.Name
				add(path, "", path, stream.Name)
			}
		}
	case "+":
		for _, tag := range state.TagVocab {
			add(tag, "tag", tag)
		}
	case "~":
		for _, estimate := range captureEstimates {
			add(estimate, "estimate", estimate)
		}
	case "@":
		today, err := time.Parse("2006-01-02", ctx.CurrentDate)
		if err != nil {
			return nil
		}
		add("today", today.Format("Mon Jan 2"), "today")
		add("tomorrow", today.AddDate(0, 0, 1).Format("Mon Jan 2"), "tomorrow")
		for offset := 2; offset < 8; offset++ {
			day := today.AddDate(0, 0, offset)
			name := strings.ToLower(day.Weekday().String())
			add(name[:3], day.Format("Mon Jan 2"), name)
		}
	}
	// A word already typed in full needs no popup.
	if len(out) == 1 && strings.EqualFold(out[0].Insert, word) {
		return nil
	}
	if len(out) > captureSuggestLimit {
		out = out[:captureSuggestLimit]
	}
	return out
}

// updateCaptureSuggest handles the popup keys; ok is false when the key is not for it.
func updateCaptureSuggest(state State, msg tea.KeyMsg) (State, bool) {
	if len(state.CaptureSuggestions) == 0 {
		return state, false
	}
	switch msg.String() {
	case "up", "ctrl+p":
		state.CaptureSuggestCursor = (state.CaptureSuggestCursor + len(state.CaptureSuggestions) - 1) % len(state.CaptureSuggestions)
	case "down", "ctrl+n":
		state.CaptureSuggestCursor = (state.CaptureSuggestCursor + 1) % len(state.CaptureSuggestions)
	case "esc":
		state.CaptureSuggestions = nil
	case "tab", "enter":
		word, ok := captureToken(state)
		if !ok {
			state.CaptureSuggestions = nil
			return state, true
		}
		pick := state.CaptureSuggestions[min(state.CaptureSuggestCursor, len(state.CaptureSuggestions)-1)]
		for range []rune(word) {
			state.Description, _ = state.Description.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		}
		state.Description.InsertString(pick.Insert + " ")
		state.CaptureSuggestions = nil
	default:
		return state, false
	}
	return state, true
}

func refreshCaptureSuggestions(state State, ctx UpdateContext) State {
	next := captureSuggestions(state, ctx)
	if !slices.Equal(next, state.CaptureSuggestions) {
		state.CaptureSuggestCursor = 0
	}
	state.CaptureSuggestions = next
	return state
}
