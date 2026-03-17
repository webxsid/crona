package store

var DefaultCoreSettings = map[string]any{
	"timerMode":             "stopwatch",
	"breaksEnabled":         false,
	"workDurationMinutes":   25,
	"shortBreakMinutes":     5,
	"longBreakMinutes":      15,
	"longBreakEnabled":      true,
	"cyclesBeforeLongBreak": 4,
	"autoStartBreaks":       true,
	"autoStartWork":         false,
	"repoSort":              string("chronological_asc"),
	"streamSort":            string("chronological_asc"),
	"issueSort":             string("priority"),
}
