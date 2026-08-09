package controller

import (
	"strconv"
	"strings"
	"time"

	sharedtypes "crona/shared/types"

	tea "github.com/charmbracelet/bubbletea"
)

func updateCreateIssueMeta(
	state State,
	currentDate string,
	msg tea.KeyMsg,
) (State, *Action, string) {
	if state.FocusIdx == 3 && isDatePickerShortcut(msg.String()) {
		return OpenDatePicker(
			state,
			"create_issue_meta",
			0,
			3,
			ValueToPointer(state.Inputs[2].Value()),
			currentDate,
		), nil, ""
	}
	if state.FocusIdx == 3 && msg.String() == "g" {
		if err := ValidateDueDateAt(state, currentDate, currentDate); err != nil {
			state.ErrorMessage = err.Error()
			return state, nil, ""
		}
		state.Inputs[2].SetValue(currentDate)
		return clearDialogError(state), nil, ""
	}
	return updateMultiInputIssue(state, msg, 4, func(state State) (*Action, string) {
		title := strings.TrimSpace(state.Inputs[0].Value())
		if title == "" {
			return nil, "Issue title is required"
		}
		description := ValueToPointer(strings.TrimSpace(state.Description.Value()))
		estimate, err := ParseEstimateInput(state.Inputs[1].Value())
		if err != nil {
			return nil, err.Error()
		}
		dueDate, err := ParseDueDateInput(state.Inputs[2].Value())
		if err != nil {
			return nil, err.Error()
		}
		if err := ValidateDueDateAt(state, ValueOrEmpty(dueDate), currentDate); err != nil {
			return nil, err.Error()
		}
		return &Action{
			Kind:        "create_issue_meta",
			StreamID:    state.StreamID,
			Title:       title,
			Description: description,
			Estimate:    estimate,
			DueDate:     dueDate,
		}, ""
	})
}

func updateCreateIssueDefault(
	state State,
	ctx UpdateContext,
	currentDate string,
	msg tea.KeyMsg,
) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "ctrl+e", "ctrl+y":
		if state.FocusIdx == 5 {
			return OpenDatePicker(
				state,
				"create_issue_default",
				0,
				5,
				ValueToPointer(state.Inputs[4].Value()),
				currentDate,
			), nil, ""
		}
	case "g":
		if state.FocusIdx == 5 {
			if err := ValidateDueDateAt(state, currentDate, currentDate); err != nil {
				state.ErrorMessage = err.Error()
				return state, nil, ""
			}
			state.Inputs[4].SetValue(currentDate)
			return state, nil, ""
		}
	case "tab", "shift+tab", "down", "up":
		if (msg.String() == "down" || msg.String() == "up") &&
			(state.FocusIdx == 0 || state.FocusIdx == 1) {
			if state.FocusIdx == 0 {
				state.RepoIndex = ShiftSelection(
					state.RepoIndex,
					len(DefaultRepoOptions(state.Inputs, ctx.Repos)),
					ternaryDir(msg.String()),
				)
				state.StreamIndex = 0
			} else {
				state.StreamIndex = ShiftSelection(state.StreamIndex, len(DefaultStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)), ternaryDir(msg.String()))
			}
			return state, nil, ""
		}
		state.FocusIdx = (state.FocusIdx + ternaryDir(msg.String()) + 6) % 6
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	case "left":
		if state.FocusIdx == 0 {
			state.RepoIndex = ShiftSelection(
				state.RepoIndex,
				len(DefaultRepoOptions(state.Inputs, ctx.Repos)),
				-1,
			)
			state.StreamIndex = 0
			return clearDialogError(state), nil, ""
		}
		if state.FocusIdx == 1 {
			state.StreamIndex = ShiftSelection(
				state.StreamIndex,
				len(
					DefaultStreamOptions(
						state.Inputs,
						state.RepoIndex,
						ctx.Repos,
						ctx.AllIssues,
						ctx.Streams,
						ctx.Context,
					),
				),
				-1,
			)
			return clearDialogError(state), nil, ""
		}
	case "right":
		if state.FocusIdx == 0 {
			state.RepoIndex = ShiftSelection(
				state.RepoIndex,
				len(DefaultRepoOptions(state.Inputs, ctx.Repos)),
				1,
			)
			state.StreamIndex = 0
			return clearDialogError(state), nil, ""
		}
		if state.FocusIdx == 1 {
			state.StreamIndex = ShiftSelection(
				state.StreamIndex,
				len(
					DefaultStreamOptions(
						state.Inputs,
						state.RepoIndex,
						ctx.Repos,
						ctx.AllIssues,
						ctx.Streams,
						ctx.Context,
					),
				),
				1,
			)
			return clearDialogError(state), nil, ""
		}
	}
	if isDialogSubmitKey(state, msg.String()) {
		repoName, streamName := DefaultIssueDialogNames(
			state.Inputs,
			state.RepoIndex,
			state.StreamIndex,
			ctx.Repos,
			ctx.AllIssues,
			ctx.Streams,
			ctx.Context,
		)
		title := strings.TrimSpace(state.Inputs[2].Value())
		if repoName == "" || streamName == "" || title == "" {
			return state, nil, "Repo, stream, and issue title are required"
		}
		description := ValueToPointer(strings.TrimSpace(state.Description.Value()))
		estimate, err := ParseEstimateInput(state.Inputs[3].Value())
		if err != nil {
			return state, nil, err.Error()
		}
		dueDate, err := ParseDueDateInput(state.Inputs[4].Value())
		if err != nil {
			return state, nil, err.Error()
		}
		if err := ValidateDueDateAt(state, ValueOrEmpty(dueDate), currentDate); err != nil {
			return state, nil, err.Error()
		}
		return Close(
				state,
			), &Action{
				Kind:        "create_issue_default",
				RepoName:    repoName,
				StreamName:  streamName,
				Title:       title,
				Description: description,
				Estimate:    estimate,
				DueDate:     dueDate,
			}, ""
	}
	if state.DescriptionEnabled && state.FocusIdx == state.DescriptionIndex {
		var cmd tea.Cmd
		state.Description, cmd = state.Description.Update(msg)
		_ = cmd
		return clearDialogError(state), nil, ""
	}
	inputIdx, ok := dialogInputIndex(state, state.FocusIdx)
	if !ok {
		return state, nil, ""
	}
	var cmd tea.Cmd
	state.Inputs[inputIdx], cmd = state.Inputs[inputIdx].Update(msg)
	_ = cmd
	if inputIdx == 0 {
		state.RepoIndex = 0
		state.StreamIndex = 0
	}
	if inputIdx == 1 {
		state.StreamIndex = 0
	}
	return clearDialogError(state), nil, ""
}

func updateCheckoutContext(
	state State,
	ctx UpdateContext,
	msg tea.KeyMsg,
) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "c":
		return Close(state), &Action{Kind: "checkout_context"}, ""
	case "tab", "shift+tab", "down", "up":
		state.FocusIdx = (state.FocusIdx + ternaryDir(msg.String()) + 2) % 2
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	case "left":
		if state.FocusIdx == 0 {
			state.RepoIndex = ShiftSelection(
				state.RepoIndex,
				len(CheckoutRepoOptions(state.Inputs, ctx.Repos)),
				-1,
			)
			if strings.TrimSpace(state.Inputs[1].Value()) == "" {
				state.StreamIndex = -1
			}
			return state, nil, ""
		}
		if state.FocusIdx == 1 {
			state.StreamIndex = ShiftSelection(
				state.StreamIndex,
				len(
					CheckoutStreamOptions(
						state.Inputs,
						state.RepoIndex,
						ctx.Repos,
						ctx.AllIssues,
						ctx.Streams,
						ctx.Context,
					),
				),
				-1,
			)
			return state, nil, ""
		}
	case "right":
		if state.FocusIdx == 0 {
			state.RepoIndex = ShiftSelection(
				state.RepoIndex,
				len(CheckoutRepoOptions(state.Inputs, ctx.Repos)),
				1,
			)
			if strings.TrimSpace(state.Inputs[1].Value()) == "" {
				state.StreamIndex = -1
			}
			return state, nil, ""
		}
		if state.FocusIdx == 1 {
			state.StreamIndex = ShiftSelection(
				state.StreamIndex,
				len(
					CheckoutStreamOptions(
						state.Inputs,
						state.RepoIndex,
						ctx.Repos,
						ctx.AllIssues,
						ctx.Streams,
						ctx.Context,
					),
				),
				1,
			)
			return state, nil, ""
		}
	case "enter":
		repoID, repoName, streamID, streamName := CheckoutDialogSelection(
			state.Inputs,
			state.RepoIndex,
			state.StreamIndex,
			ctx.Repos,
			ctx.AllIssues,
			ctx.Streams,
			ctx.Context,
		)
		if strings.TrimSpace(repoName) == "" && streamID == nil &&
			strings.TrimSpace(streamName) == "" {
			return Close(state), &Action{Kind: "checkout_context"}, ""
		}
		if strings.TrimSpace(repoName) == "" {
			return state, nil, "Repo is required"
		}
		if streamID != nil {
			return Close(
					state,
				), &Action{
					Kind:       "checkout_context",
					RepoID:     repoID,
					RepoName:   repoName,
					StreamID:   *streamID,
					StreamName: streamName,
				}, ""
		}
		if strings.TrimSpace(streamName) == "" {
			return Close(
					state,
				), &Action{
					Kind:     "checkout_context",
					RepoID:   repoID,
					RepoName: repoName,
				}, ""
		}
		return Close(
				state,
			), &Action{
				Kind:       "checkout_context",
				RepoID:     repoID,
				RepoName:   repoName,
				StreamName: streamName,
			}, ""
	}
	var cmd tea.Cmd
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	if state.FocusIdx == 0 {
		if strings.TrimSpace(state.Inputs[0].Value()) == "" {
			state.RepoIndex = -1
			state.StreamIndex = -1
		} else {
			state.RepoIndex = 0
			state.StreamIndex = 0
		}
	}
	if state.FocusIdx == 1 {
		if strings.TrimSpace(state.Inputs[1].Value()) == "" {
			state.StreamIndex = -1
		} else {
			state.StreamIndex = 0
		}
	}
	return clearDialogError(state), nil, ""
}

func updateEditIssue(state State, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	if state.FocusIdx == 3 && isDatePickerShortcut(msg.String()) {
		return OpenDatePicker(
			state,
			"edit_issue",
			state.IssueID,
			3,
			ValueToPointer(state.Inputs[2].Value()),
			currentDate,
		), nil, ""
	}
	if state.FocusIdx == 3 && msg.String() == "g" {
		if err := ValidateDueDateAt(state, currentDate, currentDate); err != nil {
			state.ErrorMessage = err.Error()
			return state, nil, ""
		}
		state.Inputs[2].SetValue(currentDate)
		return clearDialogError(state), nil, ""
	}
	return updateMultiInputIssue(state, msg, 4, func(state State) (*Action, string) {
		title := strings.TrimSpace(state.Inputs[0].Value())
		if title == "" {
			return nil, "Issue title is required"
		}
		description := ValueToPointer(strings.TrimSpace(state.Description.Value()))
		estimate, err := ParseEstimateInput(state.Inputs[1].Value())
		if err != nil {
			return nil, err.Error()
		}
		dueDate, err := ParseDueDateInput(state.Inputs[2].Value())
		if err != nil {
			return nil, err.Error()
		}
		if err := ValidateDueDateAt(state, ValueOrEmpty(dueDate), currentDate); err != nil {
			return nil, err.Error()
		}
		return &Action{
			Kind:        "edit_issue",
			IssueID:     state.IssueID,
			StreamID:    state.StreamID,
			Title:       title,
			Description: description,
			Estimate:    estimate,
			DueDate:     dueDate,
		}, ""
	})
}

func updateMultiInputIssue(
	state State,
	msg tea.KeyMsg,
	inputCount int,
	submit func(State) (*Action, string),
) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		state.FocusIdx = (state.FocusIdx + ternaryDir(msg.String()) + inputCount) % inputCount
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			action, status := submit(state)
			if action == nil {
				return state, nil, status
			}
			return Close(state), action, status
		}
	}
	if state.DescriptionEnabled && state.FocusIdx == state.DescriptionIndex {
		var cmd tea.Cmd
		state.Description, cmd = state.Description.Update(msg)
		_ = cmd
		return clearDialogError(state), nil, ""
	}
	inputIdx, ok := dialogInputIndex(state, state.FocusIdx)
	if !ok {
		return clearDialogError(state), nil, ""
	}
	var cmd tea.Cmd
	state.Inputs[inputIdx], cmd = state.Inputs[inputIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func isDatePickerShortcut(key string) bool {
	return key == "ctrl+e" || key == "ctrl+y"
}

func updateExportCategory(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		if state.Processing {
			return state, nil, ""
		}
		return Close(state), nil, ""
	case "down":
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
	case "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
	case "enter":
		if state.ChoiceCursor < 0 || state.ChoiceCursor >= len(state.ChoiceValues) {
			return state, nil, "Select an export category"
		}
		return OpenExportReportChoices(state, state.ChoiceValues[state.ChoiceCursor]), nil, ""
	}
	return clearDialogError(state), nil, ""
}

func updateExportDaily(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		if state.Processing {
			return state, nil, ""
		}
		items, values, details := exportReportCategories()
		state.Kind = "export_report_category"
		state.Parent = ""
		state.ExportCategory = ""
		state.ChoiceItems = items
		state.ChoiceValues = values
		state.ChoiceDetails = details
		state.ChoiceCursor = 0
		return clearDialogError(state), nil, ""
	case "down":
		if state.Processing {
			return state, nil, ""
		}
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
	case "up":
		if state.Processing {
			return state, nil, ""
		}
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
	case "enter":
		if state.Processing {
			return state, nil, ""
		}
		selected := ""
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceItems) {
			selected = state.ChoiceItems[state.ChoiceCursor]
		}
		action := Action{Kind: "export_report", CheckInDate: state.CheckInDate}
		switch selected {
		case "Daily report: write PDF file":
			return OpenExportPreset(
				state,
				sharedtypes.ExportReportKindDaily,
				sharedtypes.ExportFormatPDF,
				sharedtypes.ExportOutputModeFile,
			), nil, ""
		case "Weekly summary: write PDF file":
			return OpenExportPreset(
				state,
				sharedtypes.ExportReportKindWeekly,
				sharedtypes.ExportFormatPDF,
				sharedtypes.ExportOutputModeFile,
			), nil, ""
		case "Daily report: write Markdown file":
			return OpenExportPreset(
				state,
				sharedtypes.ExportReportKindDaily,
				sharedtypes.ExportFormatMarkdown,
				sharedtypes.ExportOutputModeFile,
			), nil, ""
		case "Daily report: copy to clipboard":
			return OpenExportPreset(
				state,
				sharedtypes.ExportReportKindDaily,
				sharedtypes.ExportFormatMarkdown,
				sharedtypes.ExportOutputModeClipboard,
			), nil, ""
		case "Weekly summary: write Markdown file":
			return OpenExportPreset(
				state,
				sharedtypes.ExportReportKindWeekly,
				sharedtypes.ExportFormatMarkdown,
				sharedtypes.ExportOutputModeFile,
			), nil, ""
		case "Repo report: write Markdown file":
			return OpenExportScope(state, sharedtypes.ExportReportKindRepo), nil, ""
		case "Stream report: write Markdown file":
			return OpenExportScope(state, sharedtypes.ExportReportKindStream), nil, ""
		case "Issue rollup: write Markdown file":
			return OpenExportScope(state, sharedtypes.ExportReportKindIssueRollup), nil, ""
		case "Calendar export: write ICS file":
			if len(state.RepoItems) == 0 {
				return state, nil, "Calendar export requires at least one repo"
			}
			return OpenExportCalendarRepo(state), nil, ""
		default:
			action.ReportKind = sharedtypes.ExportReportKindCSV
			action.ReportFormat = sharedtypes.ExportFormatCSV
			action.OutputMode = sharedtypes.ExportOutputModeFile
			state.ProcessingLabel = "Generating CSV export..."
		}
		state.Processing = true
		return state, &action, ""
	}
	return state, nil, ""
}

func updateExportSummaryPeriod(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		items, values, details := exportReportCategories()
		state.Kind = "export_report_category"
		state.ExportCategory = ""
		state.ChoiceItems, state.ChoiceValues, state.ChoiceDetails = items, values, details
		state.ChoiceCursor = 0
	case "down":
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
	case "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
	case "enter":
		if state.ChoiceCursor < 0 || state.ChoiceCursor >= len(state.ChoiceItems) {
			return state, nil, "Select a summary period"
		}
		return OpenSummaryOutput(state, state.ChoiceItems[state.ChoiceCursor]), nil, ""
	}
	return clearDialogError(state), nil, ""
}

func updateExportScope(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return OpenExportReportChoices(state, "project"), nil, ""
	case "down":
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
	case "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
	case "enter":
		if state.ChoiceCursor < 0 || state.ChoiceCursor >= len(state.ChoiceValues) {
			return state, nil, "Select an export target"
		}
		id, err := strconv.ParseInt(state.ChoiceValues[state.ChoiceCursor], 10, 64)
		if err != nil {
			return state, nil, "Invalid export target"
		}
		action := &Action{Kind: "export_report", ReportKind: state.ExportPresetKind, ReportFormat: sharedtypes.ExportFormatMarkdown, OutputMode: sharedtypes.ExportOutputModeFile, CheckInDate: state.CheckInDate}
		switch state.ExportPresetKind {
		case sharedtypes.ExportReportKindRepo:
			action.RepoID = id
		case sharedtypes.ExportReportKindStream:
			action.StreamID = id
		case sharedtypes.ExportReportKindIssueRollup:
			action.IssueID = id
		}
		state.Processing = true
		state.ProcessingLabel = "Generating report..."
		return state, action, ""
	}
	return clearDialogError(state), nil, ""
}

func updateExportSummaryOutput(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return OpenExportReportChoices(state, "summary"), nil, ""
	case "down":
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
	case "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
	case "enter":
		selected := state.ChoiceItems[state.ChoiceCursor]
		format, output := sharedtypes.ExportFormatMarkdown, sharedtypes.ExportOutputModeFile
		if selected == "Write PDF file" {
			format = sharedtypes.ExportFormatPDF
		}
		if selected == "Copy Markdown" {
			output = sharedtypes.ExportOutputModeClipboard
		}
		reportKind := sharedtypes.ExportReportKindGlance
		if state.ExportStart != "" && state.ExportEnd != "" && state.ExportStart != state.ExportEnd {
			reportKind = sharedtypes.ExportReportKindSummaryRange
		}
		return OpenExportPreset(state, reportKind, format, output), nil, ""
	}
	return clearDialogError(state), nil, ""
}

func updateExportSummaryRange(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return OpenExportReportChoices(state, "summary"), nil, ""
	case "tab", "shift+tab", "up", "down":
		if len(state.Inputs) == 2 {
			state.Inputs[state.FocusIdx].Blur()
			state.FocusIdx = 1 - state.FocusIdx
			state.Inputs[state.FocusIdx].Focus()
		}
		return clearDialogError(state), nil, ""
	case "enter":
		start, end := strings.TrimSpace(state.Inputs[0].Value()), strings.TrimSpace(state.Inputs[1].Value())
		startDate, startErr := time.Parse(time.DateOnly, start)
		endDate, endErr := time.Parse(time.DateOnly, end)
		if startErr != nil || endErr != nil {
			return state, nil, "Use dates in YYYY-MM-DD format"
		}
		if endDate.Before(startDate) {
			return state, nil, "End date must be on or after start date"
		}
		state.ExportStart, state.ExportEnd = start, end
		state.Inputs = nil
		state.Kind = "export_summary_output"
		state.ChoiceItems = []string{"Write Markdown file", "Copy Markdown"}
		state.ChoiceDetails = []string{"Save the dashboard as Markdown.", "Copy the dashboard Markdown to the clipboard."}
		if state.ExportIncludePDF {
			state.ChoiceItems = append([]string{"Write PDF file"}, state.ChoiceItems...)
			state.ChoiceDetails = append([]string{"Save a styled, landscape PDF dashboard."}, state.ChoiceDetails...)
		}
		state.ChoiceCursor = 0
		return state, nil, ""
	}
	if len(state.Inputs) > 0 {
		var cmd tea.Cmd
		state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
		_ = cmd
	}
	return clearDialogError(state), nil, ""
}

func updateExportPreset(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		state.Kind = "export_report"
		state.Parent = ""
		state.ChoiceItems, state.ChoiceDetails = exportReportChoices(
			state.ExportCategory,
			state.ExportIncludePDF,
		)
		state.ChoiceValues = nil
		state.ExportPresetKind = ""
		state.ExportPresetFormat = ""
		state.ExportPresetOutput = ""
		state.ChoiceCursor = 0
		return clearDialogError(state), nil, ""
	case "down":
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
	case "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
	case "enter":
		if state.ChoiceCursor < 0 || state.ChoiceCursor >= len(state.ChoiceValues) {
			return state, nil, "Select a report style"
		}
		presetID := state.ChoiceValues[state.ChoiceCursor]
		if presetID == "__current__" {
			presetID = ""
		}
		state.Processing = true
		switch state.ExportPresetKind {
		case sharedtypes.ExportReportKindDaily:
			if state.ExportPresetOutput == sharedtypes.ExportOutputModeClipboard {
				state.ProcessingLabel = "Copying report..."
			} else {
				state.ProcessingLabel = "Generating daily report..."
			}
		case sharedtypes.ExportReportKindWeekly:
			state.ProcessingLabel = "Generating weekly report..."
		default:
			state.ProcessingLabel = "Generating report..."
		}
		return state, &Action{
			Kind:         "export_report",
			ReportKind:   state.ExportPresetKind,
			ReportFormat: state.ExportPresetFormat,
			OutputMode:   state.ExportPresetOutput,
			CheckInDate:  state.CheckInDate,
			StartDate:    state.ExportStart,
			EndDate:      state.ExportEnd,
			PresetID:     presetID,
		}, ""
	}
	return clearDialogError(state), nil, ""
}

func updateExportCalendarRepo(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		state.Kind = "export_report"
		state.Parent = ""
		state.ChoiceItems, state.ChoiceDetails = exportReportChoices(
			state.ExportCategory,
			state.ExportIncludePDF,
		)
		state.ChoiceValues = nil
		state.ChoiceCursor = 0
		return clearDialogError(state), nil, ""
	case "down":
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
	case "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
	case "enter":
		if len(state.RepoItemIDs) == 0 || state.ChoiceCursor < 0 ||
			state.ChoiceCursor >= len(state.RepoItemIDs) {
			return state, nil, "Calendar export requires a repo"
		}
		state.RepoIndex = state.ChoiceCursor
		state.RepoID = state.RepoItemIDs[state.ChoiceCursor]
		state.RepoName = state.RepoItems[state.ChoiceCursor]
		state.Kind = "export_report"
		state.ChoiceItems = nil
		state.ChoiceCursor = 0
		state.Parent = ""
		state.Processing = true
		state.ProcessingLabel = "Generating calendar export..."
		return state, &Action{
			Kind:         "export_report",
			ReportKind:   sharedtypes.ExportReportKindCalendar,
			ReportFormat: sharedtypes.ExportFormatICS,
			OutputMode:   sharedtypes.ExportOutputModeFile,
			CheckInDate:  state.CheckInDate,
			RepoID:       state.RepoID,
			RepoName:     state.RepoName,
		}, ""
	}
	return clearDialogError(state), nil, ""
}
