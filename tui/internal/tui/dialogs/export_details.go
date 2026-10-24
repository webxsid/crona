package dialogs

import (
	"strings"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	helperpkg "crona/tui/internal/tui/helpers"
)

// renderExportChoice draws the export pickers: a list, the selected item's help (or a
// preview), with the anchor date in the border. While an export runs the list dims and
// the progress label shows instead of the keys.
func renderExportChoice(theme Theme, state controllerpkg.State) string {
	d := listDialog{
		context: helperpkg.FormatDisplayDate(state.CheckInDate, nil),
		border:  theme.ColorGreen,
		hints:   []string{"↑/↓ move"},
	}
	body := []string{}
	switch state.Kind {
	case "export_report", "export_summary_period":
		d.title, d.primary = "Export Report", "enter choose"
		body = append(
			body,
			Field(
				theme,
				"Category",
				len("Category"),
				theme.StyleHeader.Render(exportCategoryTitle(state.ExportCategory)),
				false,
			),
			"",
		)
	case "export_report_category":
		d.title, d.primary = "Export Report", "enter choose category"
	case "export_preset", "export_summary_output":
		d.title, d.primary = "Choose Report Style", "enter use style"
	default:
		d.title, d.primary = "Calendar Export Repo", "enter export"
		body = append(body, theme.StyleDim.Render("Select repo"))
	}
	d.focus = len(body) + state.ChoiceCursor
	if state.Processing {
		for _, item := range state.ChoiceItems {
			body = append(body, theme.StyleDim.Render("  "+item))
		}
		body = append(
			body,
			"",
			theme.StyleHeader.Render(state.ProcessingLabel),
			theme.StyleDim.Render("Please wait..."),
		)
		d.primary, d.hints = "", nil
	} else {
		body = append(body, selectList(theme, state.ChoiceItems, state.ChoiceCursor)...)
	}
	if !state.Processing && state.ChoiceCursor >= 0 &&
		state.ChoiceCursor < len(state.ChoiceDetails) {
		detail := state.ChoiceDetails[state.ChoiceCursor]
		switch {
		case state.Kind == "export_preset" || state.Kind == "export_summary_output":
			body = append(body, "", theme.StyleDim.Render("Preview"))
			body = append(body, strings.Split(renderViewEntityBody(theme, detail), "\n")...)
		case strings.TrimSpace(detail) != "" && state.Kind != "export_calendar_repo" && state.Kind != "export_scope":
			body = append(body, "", theme.StyleDim.Render(detail))
		}
	}
	d.body = body
	return renderListDialog(theme, state, d)
}

// renderExportPathForm draws the reports and ICS directory forms.
func renderExportPathForm(theme Theme, state controllerpkg.State, title string) string {
	inner := formInner(state.Width)
	full, _ := formColumns(inner, len("Path"))
	return renderForm(theme, state, FormSpec{
		Title: title, FocusCount: 1, Primary: plainHint(dialogSubmitHint(state, "save")),
		Rows: []FormRow{
			fieldRow(
				0,
				Field(theme, "Path", len("Path"), bareInput(state.Inputs[0], full, "~/..."), true),
			),
			staticRow("", theme.StyleDim.Render("Use an absolute path or ~/...")),
		},
	})
}

// renderSummaryRangeForm draws Custom Summary Range: Start and End dates.
func renderSummaryRangeForm(theme Theme, state controllerpkg.State) string {
	inner := formInner(state.Width)
	column := len("Start")
	_, half := formColumns(inner, column)
	return renderForm(theme, state, FormSpec{
		Title: "Custom Summary Range", Border: theme.ColorGreen, FocusCount: 2, Primary: "enter continue",
		Hints: func(int) []string { return []string{"tab switch"} },
		Rows: []FormRow{pairRow(
			Field(
				theme,
				"Start",
				column,
				bareInput(state.Inputs[0], half, "YYYY-MM-DD"),
				state.FocusIdx == 0,
			),
			Field(
				theme,
				"End",
				column,
				bareInput(state.Inputs[1], half, "YYYY-MM-DD"),
				state.FocusIdx == 1,
			),
			0,
			1,
			inner,
		)},
	})
}

// renderDetailView draws view_entity, view_momentum_detail and support_bundle_result: the
// name, its meta fields, the body, and (for momentum) a contributor list.
func renderDetailView(theme Theme, state controllerpkg.State) string {
	inner := FrameInner(state.Width, []string{strings.Repeat(" ", frameMaxInner)})
	body := []string{theme.StyleHeader.Render(fallback(state.ViewName, "-"))}
	if state.ViewMeta != "" {
		body = append(body, renderViewMeta(theme, state.ViewMeta, inner)...)
	}
	if strings.TrimSpace(state.ViewBody) != "" {
		body = append(body, "")
		body = append(body, strings.Split(renderViewEntityBody(theme, state.ViewBody), "\n")...)
	}
	d := listDialog{title: state.ViewTitle, esc: "esc close", focus: -1}
	switch state.Kind {
	case "view_momentum_detail":
		body = append(body, "", theme.StyleDim.Render("Contributors"))
		if len(state.ChoiceItems) == 0 {
			body = append(
				body,
				theme.StyleDim.Render("No contributing entries in the current bucket"),
			)
		}
		for i, item := range state.ChoiceItems {
			cursor := -1
			if i == state.ChoiceCursor {
				d.focus, cursor = len(body), 0
			}
			body = append(body, selectList(theme, []string{item}, cursor)...)
			if i < len(state.ChoiceDetails) && strings.TrimSpace(state.ChoiceDetails[i]) != "" {
				body = append(body, theme.StyleDim.Render("    "+state.ChoiceDetails[i]))
			}
		}
		d.hints = []string{"↑/↓ scroll", "e edit"}
	case "support_bundle_result":
		d.border = theme.ColorGreen
		d.hints = []string{"o open folder", "c copy path", "g report issue"}
	default:
		if strings.TrimSpace(state.ViewPath) != "" {
			d.hints = append(d.hints, "e edit")
		}
		if state.ViewEntityAdhoc {
			d.hints = append(d.hints, "e convert to regular issue")
		}
	}
	d.body = body
	return renderListDialog(theme, state, d)
}
