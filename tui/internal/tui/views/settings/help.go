package settings

import viewchrome "crona/tui/internal/tui/views/chrome"

var (
	toggle     = []string{"Enabled", "Disabled"}
	orderSorts = []string{"Oldest first", "Newest first", "A -> Z", "Z -> A"}
	change     = []string{"[←/→] change"}
	edit       = []string{"[enter] edit"}
)

// settingsHelp explains each Settings row, keyed by its label.
var settingsHelp = map[string]viewchrome.RowHelp{
	"Manage History": {
		Description: "Lets you edit past days' sessions, check-ins and habits. Off keeps history read-only.",
		Options:     toggle, Keys: change,
	},
	"History Edit Window": {
		Description: "How far back history stays editable while Manage History is on.",
		Keys:        change,
	},
	"Update Checks": {
		Description: "Checks for new releases in the background.",
		Options:     toggle, Keys: change,
	},
	"Update Prompt": {
		Description: "Shows a prompt when a newer release is found.",
		Options:     toggle, Keys: change,
	},
	"Update Channel": {
		Description: "Which releases count as updates. Beta gets pre-releases early.",
		Options:     []string{"Stable", "Beta"}, Keys: change,
	},
	"Repo Sort": {
		Description: "Order of repos in lists and pickers.",
		Options:     orderSorts,
		Keys:        change,
	},
	"Stream Sort": {
		Description: "Order of streams within a repo.",
		Options:     orderSorts,
		Keys:        change,
	},
	"Issue Sort": {
		Description: "Order of issues in Issues and Today.",
		Options: []string{
			"Priority",
			"Due date earliest",
			"Due date latest",
			"Oldest first",
			"Newest first",
			"A -> Z",
			"Z -> A",
		},
		Keys: change,
	},
	"Habit Sort": {
		Description: "Order of habits in Today and the habit list.",
		Options: []string{
			"Schedule",
			"Target shortest",
			"Target longest",
			"Oldest first",
			"Newest first",
			"A -> Z",
			"Z -> A",
		},
		Keys: change,
	},
	"Date Format": {
		Description: "How dates are written across the app. Custom uses the format below.",
		Options:     []string{"YYYY-MM-DD", "MM/DD/YYYY", "DD/MM/YYYY", "D MMM YYYY", "Custom"},
		Keys:        change,
	},
	"Custom Date Format": {
		Description: "The pattern used when Date Format is Custom.",
		Keys:        edit,
	},
	"Date Preview": {Description: "Today's date in the current format."},
	"Week Start": {
		Description: "First day of the week for Review, calendars and week totals.",
		Options:     []string{"Monday", "Sunday"}, Keys: change,
	},
	"Prompt Glyphs": {
		Description: "Symbols used in prompts and markers. ASCII suits limited terminals.",
		Options:     []string{"Emoji", "Unicode", "ASCII"}, Keys: change,
	},
	"Away Mode": {
		Description: "Allows days marked away, which pause streaks and hide work views.",
		Options:     toggle, Keys: change,
	},
	"Rollback Window": {
		Description: "How long after midnight yesterday's plan can still be rolled back.",
		Keys:        change,
	},
	"High-Risk Threshold": {
		Description: "Delay penalty at which a planned issue is flagged high risk.",
		Keys:        change,
	},
	"Carry Over": {
		Description: "Moves issues you worked on but didn't finish to the next working day, keeping a copy on each day.",
		Options:     toggle, Keys: change,
	},
	"Carry Threshold": {
		Description: "Work a day needs for an unfinished issue to carry over instead of going overdue.",
		Keys:        change,
	},
	"Stall After": {
		Description: "Carries in a row before an issue is flagged stalled and starts adding to accountability.",
		Keys:        change,
	},
	"Rest & Streak Protection": {
		Description: "Which streaks rest days protect, and which weekdays and dates are rest days.",
		Keys:        edit,
	},
	"Privacy & Diagnostics": {
		Description: "Whether anonymous usage data and crash diagnostics are sent.",
		Keys:        edit,
	},
	"Start of Day": {
		Description: "When a new day begins. Work before this counts toward the previous day.",
		Keys:        edit,
	},
	"End of Day": {
		Description: "When the day winds down; used for end-of-day prompts and summaries.",
		Keys:        edit,
	},
	"Wipe Runtime Data": {
		Description: "Deletes all local data: issues, sessions, habits and settings. This cannot be undone.",
		Keys:        []string{"[enter] confirm"},
	},
}
