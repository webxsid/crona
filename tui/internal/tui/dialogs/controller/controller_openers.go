package controller

import (
	"slices"
	"strconv"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"

	"github.com/charmbracelet/bubbles/textinput"
)

func OpenCreateScratchpad(state State) State {
	name := textinput.New()
	name.Placeholder = "My notes"
	name.Focus()
	name.CharLimit = 80
	name.Width = 40
	path := textinput.New()
	path.Placeholder = "notes/[[date]].md"
	path.CharLimit = 120
	path.Width = 40
	state = Close(state)
	state.Kind = "create_scratchpad"
	state.Inputs = []textinput.Model{name, path}
	return state
}

func OpenExportDaily(state State, date string, includePDF bool, repos []api.Repo, checkedRepoID *int64, assets *api.ExportAssetStatus) State {
	state = Close(state)
	state.Kind = "export_report_category"
	state.CheckInDate = date
	state.ExportIncludePDF = includePDF
	state.ExportCategory = ""
	items, values, details := exportReportCategories()
	state.ChoiceItems = items
	state.ChoiceValues = values
	state.ChoiceDetails = details
	state.RepoItems = make([]string, 0, len(repos))
	state.RepoItemIDs = make([]int64, 0, len(repos))
	selectedRepoIdx := 0
	for i, repo := range repos {
		state.RepoItems = append(state.RepoItems, repo.Name)
		state.RepoItemIDs = append(state.RepoItemIDs, repo.ID)
		if checkedRepoID != nil && repo.ID == *checkedRepoID {
			selectedRepoIdx = i
		}
	}
	if len(state.RepoItems) > 0 {
		state.RepoIndex = selectedRepoIdx
		state.RepoID = state.RepoItemIDs[selectedRepoIdx]
		state.RepoName = state.RepoItems[selectedRepoIdx]
	}
	state.ChoiceCursor = 0
	if assets != nil {
		state.TemplateAssets = append([]sharedtypes.ExportTemplateAsset(nil), assets.TemplateAssets...)
	}
	return state
}

func exportReportCategories() ([]string, []string, []string) {
	return []string{
			"Narrative Reports",
			"Project Reports",
			"Data Exports",
		}, []string{
			"narrative",
			"project",
			"data",
		}, []string{
			"Daily and weekly summaries, including Markdown, PDF, and clipboard output.",
			"Repo, stream, and issue rollup reports for focused project review.",
			"CSV session dumps and calendar exports for external tools and automations.",
		}
}

func exportReportChoices(category string, includePDF bool) ([]string, []string) {
	switch category {
	case "project":
		return []string{
				"Repo report: write Markdown file",
				"Stream report: write Markdown file",
				"Issue rollup: write Markdown file",
			}, []string{
				"Project-level summaries and issue rollups.",
				"Focused stream review with grouped issue output.",
				"Per-issue work summary and notes rollup.",
			}
	case "data":
		return []string{
				"CSV session export: write file",
				"Calendar export: write ICS file",
			}, []string{
				"Raw session data for spreadsheets or external analysis.",
				"Write repo-scoped ICS files for local calendar automation.",
			}
	default:
		items := []string{
			"Daily report: write Markdown file",
			"Daily report: copy to clipboard",
			"Weekly summary: write Markdown file",
		}
		details := []string{
			"Generate a readable daily report as Markdown.",
			"Copy the daily report directly to the clipboard.",
			"Generate a weekly narrative summary as Markdown.",
		}
		if includePDF {
			items = append([]string{
				"Daily report: write PDF file",
				"Weekly summary: write PDF file",
			}, items...)
			details = append([]string{
				"Generate a styled daily PDF report.",
				"Generate a styled weekly PDF report.",
			}, details...)
		}
		return items, details
	}
}

func OpenExportPreset(state State, reportKind sharedtypes.ExportReportKind, format sharedtypes.ExportFormat, outputMode sharedtypes.ExportOutputMode) State {
	state.Parent = "export_report"
	state.Kind = "export_preset"
	state.ExportPresetKind = reportKind
	state.ExportPresetFormat = format
	state.ExportPresetOutput = outputMode
	state.ChoiceItems = nil
	state.ChoiceValues = nil
	state.ChoiceDetails = nil
	state.ChoiceCursor = 0
	assetKinds := []sharedtypes.ExportAssetKind{sharedtypes.ExportAssetKindTemplateMarkdown}
	if format == sharedtypes.ExportFormatPDF {
		assetKinds = []sharedtypes.ExportAssetKind{
			sharedtypes.ExportAssetKindTemplatePDFHTML,
			sharedtypes.ExportAssetKindTemplatePDF,
		}
	}
	for _, asset := range state.TemplateAssets {
		if asset.ReportKind != reportKind || !slices.Contains(assetKinds, asset.AssetKind) {
			continue
		}
		if asset.Customized {
			currentLabel := "Current custom template"
			currentDetail := "Uses your edited active template without replacing it."
			if format == sharedtypes.ExportFormatPDF {
				currentLabel = "Current custom HTML + CSS"
				currentDetail = "Uses your edited active PDF HTML template and paired stylesheet without replacing them."
			}
			state.ChoiceItems = append(state.ChoiceItems, currentLabel)
			state.ChoiceValues = append(state.ChoiceValues, "__current__")
			state.ChoiceDetails = append(state.ChoiceDetails, currentDetail)
		}
		selectedIdx := 0
		for _, preset := range asset.Presets {
			label := preset.Label
			if asset.SelectedPreset != nil && preset.ID == asset.SelectedPreset.ID {
				label += "  [saved default]"
				selectedIdx = len(state.ChoiceItems)
			}
			state.ChoiceItems = append(state.ChoiceItems, label)
			state.ChoiceValues = append(state.ChoiceValues, preset.ID)
			state.ChoiceDetails = append(state.ChoiceDetails, preset.PreviewBody)
		}
		state.ChoiceCursor = selectedIdx
		break
	}
	return state
}

func OpenExportReportChoices(state State, category string) State {
	state.Kind = "export_report"
	state.Parent = "export_report_category"
	state.ExportCategory = category
	state.ChoiceCursor = 0
	state.ChoiceItems, state.ChoiceDetails = exportReportChoices(category, state.ExportIncludePDF)
	state.ChoiceValues = nil
	return state
}

func OpenExportCalendarRepo(state State) State {
	if len(state.RepoItems) == 0 {
		return state
	}
	state.Kind = "export_calendar_repo"
	state.Parent = "export_report"
	state.ChoiceItems = append([]string(nil), state.RepoItems...)
	if state.RepoIndex < 0 || state.RepoIndex >= len(state.ChoiceItems) {
		state.RepoIndex = 0
	}
	state.ChoiceCursor = state.RepoIndex
	return state
}

func OpenExportReportsDir(state State, current string) State {
	input := textinput.New()
	input.Placeholder = "Reports directory"
	input.SetValue(strings.TrimSpace(current))
	input.Focus()
	input.CharLimit = 240
	input.Width = 56
	state = Close(state)
	state.Kind = "edit_export_reports_dir"
	state.Inputs = []textinput.Model{input}
	return state
}

func OpenExportICSDir(state State, current string) State {
	input := textinput.New()
	input.Placeholder = "ICS export directory"
	input.SetValue(strings.TrimSpace(current))
	input.Focus()
	input.CharLimit = 240
	input.Width = 56
	state = Close(state)
	state.Kind = "edit_export_ics_dir"
	state.Inputs = []textinput.Model{input}
	return state
}

func OpenEditDateDisplayFormat(state State, current string) State {
	input := textinput.New()
	input.Placeholder = "YYYY-MM-DD | Do MMM YYYY | MM/DD/YYYY"
	input.SetValue(strings.TrimSpace(current))
	input.Focus()
	input.CharLimit = 120
	input.Width = 56
	state = Close(state)
	state.Kind = "edit_date_display_format"
	state.Inputs = []textinput.Model{input}
	return state
}

func OpenEditRestProtection(state State, streaks []sharedtypes.StreakKind, weekdays []int, dates []string) State {
	state = Close(state)
	state.Kind = "edit_rest_protection"
	state.ProtectionStep = 0
	state.ProtectionCursor = 0
	if len(streaks) == 0 {
		state.ProtectionStreaks = append([]sharedtypes.StreakKind(nil), sharedtypes.AvailableStreakKinds()...)
	} else {
		state.ProtectionStreaks = append([]sharedtypes.StreakKind(nil), streaks...)
	}
	state.ProtectionWeekdays = append([]int(nil), weekdays...)
	state.ProtectionDates = normalizedDateList(dates)
	return state
}

func OpenEditHabitStreaks(state State, settings *api.CoreSettings, habits []api.HabitWithMeta) State {
	state = Close(state)
	state.Kind = "edit_habit_streaks"
	state.HabitItems = append([]sharedtypes.HabitWithMeta(nil), habits...)
	if settings != nil {
		state.HabitStreakDefs = append([]sharedtypes.HabitStreakDefinition(nil), settings.HabitStreakDefs...)
	}
	state.HabitStreakDefs = sharedtypes.NormalizeHabitStreakDefinitions(state.HabitStreakDefs)
	state.HabitStreakStep = 0
	state.HabitStreakCursor = 0
	state.HabitStreakEditIdx = -1
	state.HabitStreakDraft = sharedtypes.HabitStreakDefinition{
		Enabled:       true,
		Period:        sharedtypes.HabitStreakPeriodDay,
		RequiredCount: 1,
	}
	return state
}

func OpenCreateCheckIn(state State, date string) State {
	return openCheckInDialog(state, "create_checkin", date, nil)
}

func OpenEditCheckIn(state State, checkIn *api.DailyCheckIn, date string) State {
	return openCheckInDialog(state, "edit_checkin", date, checkIn)
}

func openCheckInDialog(state State, kind string, date string, checkIn *api.DailyCheckIn) State {
	mood := textinput.New()
	mood.Placeholder = "Mood 1-5"
	mood.CharLimit = 1
	mood.Width = 12
	mood.Focus()
	energy := textinput.New()
	energy.Placeholder = "Energy 1-5"
	energy.CharLimit = 1
	energy.Width = 12
	sleepHours := textinput.New()
	sleepHours.Placeholder = "7.5h | 7h30m | 450m"
	sleepHours.CharLimit = 8
	sleepHours.Width = 20
	sleepHours = withTimePrompt(state, sleepHours)
	sleepScore := textinput.New()
	sleepScore.Placeholder = "Sleep score"
	sleepScore.CharLimit = 3
	sleepScore.Width = 16
	screenTime := textinput.New()
	screenTime.Placeholder = "45m | 1h20m"
	screenTime.CharLimit = 8
	screenTime.Width = 20
	screenTime = withTimePrompt(state, screenTime)
	notes := textinput.New()
	notes.Placeholder = "Notes (optional)"
	notes.CharLimit = 200
	notes.Width = 52
	if checkIn != nil {
		mood.SetValue(strconv.Itoa(checkIn.Mood))
		energy.SetValue(strconv.Itoa(checkIn.Energy))
		if checkIn.SleepHours != nil {
			sleepHours.SetValue(FormatDurationHoursInput(checkIn.SleepHours))
		}
		if checkIn.SleepScore != nil {
			sleepScore.SetValue(strconv.Itoa(*checkIn.SleepScore))
		}
		if checkIn.ScreenTimeMinutes != nil {
			screenTime.SetValue(FormatDurationMinutesInput(checkIn.ScreenTimeMinutes))
		}
		if checkIn.Notes != nil {
			notes.SetValue(strings.TrimSpace(*checkIn.Notes))
		}
	}
	state = Close(state)
	state.Kind = kind
	state.CheckInDate = date
	state.Inputs = []textinput.Model{mood, energy, sleepHours, sleepScore, screenTime, notes}
	return state
}
