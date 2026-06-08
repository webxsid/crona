package controller

import (
	"strings"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"

	"crona/tui/internal/api"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type UpdateContext struct {
	Repos             []api.Repo
	Streams           []api.Stream
	AllIssues         []api.IssueWithMeta
	Context           *api.ActiveContext
	Stashes           []api.Stash
	SelectedIssueID   int64
	SelectedStreamID  int64
	HasSelectedIssue  bool
	ActiveIssueStream int64
	HasActiveIssue    bool
}

type Action struct {
	Kind               string
	TargetView         string
	ReportKind         sharedtypes.ExportReportKind
	ReportFormat       sharedtypes.ExportFormat
	OutputMode         sharedtypes.ExportOutputMode
	PresetID           string
	ID                 string
	RepoID             int64
	StreamID           int64
	IssueID            int64
	HabitID            int64
	Name               string
	Path               string
	CheckInDate        string
	RepoName           string
	StreamName         string
	Title              string
	Description        *string
	Status             string
	Weekdays           []int
	Active             bool
	Estimate           *int
	DueDate            *string
	Note               *string
	ReminderKind       sharedtypes.AlertReminderKind
	ReminderSchedule   sharedtypes.AlertReminderScheduleType
	ReminderTimeHHMM   string
	SettingKey         sharedtypes.CoreSettingsKey
	StringList         []string
	IntList            []int
	StreakKinds        []string
	HabitStreakDefs    []sharedtypes.HabitStreakDefinition
	RestDates          []string
	UsageTelemetry     bool
	ErrorReporting     bool
	RestartAfterSave   bool
	OnboardingDone     bool
	Mood               int
	Energy             int
	SleepHours         *float64
	SleepScore         *int
	ScreenTimeMinutes  *int
	Payload            shareddto.EndSessionRequest
	ManualSession      *shareddto.ManualSessionLogRequest
	TimerStart         *shareddto.TimerStartRequest
	TimerExtend        *shareddto.TimerExtendRequest
	AdditionalSessions int
}

func Close(state State) State {
	state.Kind = ""
	state.ErrorMessage = ""
	state.Inputs = nil
	state.Description = textarea.Model{}
	state.DescriptionEnabled = false
	state.DescriptionIndex = 0
	state.FocusIdx = 0
	state.DeleteID = ""
	state.DeleteKind = ""
	state.DeleteLabel = ""
	state.SessionID = ""
	state.IssueID = 0
	state.HabitID = 0
	state.TargetView = ""
	state.IssueStatus = ""
	state.RepoID = 0
	state.RepoName = ""
	state.RepoItems = nil
	state.RepoItemIDs = nil
	state.StreamID = 0
	state.StreamName = ""
	state.RepoIndex = 0
	state.StreamIndex = 0
	state.Parent = ""
	state.DateMonthValue = ""
	state.DateCursorValue = ""
	state.StashCursor = 0
	state.StatusItems = nil
	state.StatusCursor = 0
	state.ChoiceItems = nil
	state.ChoiceValues = nil
	state.ChoiceDetails = nil
	state.ChoiceCursor = 0
	state.Processing = false
	state.ProcessingLabel = ""
	state.StatusLabel = ""
	state.StatusRequired = false
	state.CheckInDate = ""
	state.ViewTitle = ""
	state.ViewName = ""
	state.IssueEstimateMins = nil
	state.ReminderID = ""
	state.ReminderKind = ""
	state.ViewMeta = ""
	state.ViewBody = ""
	state.ViewPath = ""
	state.ExportPresetKind = ""
	state.ExportPresetFormat = ""
	state.ExportPresetOutput = ""
	state.ExportIncludePDF = false
	state.PomodoroFocusSeconds = 0
	state.PomodoroFocusChoice = 0
	state.PomodoroBreakSeconds = 0
	state.PomodoroBreakChoice = 0
	state.PomodoroLongBreakSeconds = 0
	state.PomodoroLongBreakChoice = 0
	state.PomodoroCyclesBeforeLongBreak = 0
	state.PomodoroCycles = 0
	state.HardLimitTotalSeconds = 0
	state.HardLimitFocusSeconds = 0
	state.HardLimitBreakSeconds = 0
	state.HardLimitLongBreakSeconds = 0
	state.HardLimitCyclesBeforeLongBreak = 0
	return state
}

func SyncDialogFocus(state State) State {
	for i := range state.Inputs {
		state.Inputs[i].Blur()
	}
	if state.DescriptionEnabled {
		state.Description.Blur()
	}
	if state.Kind == "pomodoro_start" || state.Kind == "hard_limit_extend" {
		if inputIdx, ok := pomodoroDialogInputIndex(state, state.FocusIdx); ok {
			state.Inputs[inputIdx].Focus()
		}
		return state
	}
	if state.DescriptionEnabled && state.FocusIdx == state.DescriptionIndex {
		state.Description.Focus()
		return state
	}
	if inputIdx, ok := dialogInputIndex(state, state.FocusIdx); ok {
		state.Inputs[inputIdx].Focus()
	}
	return state
}

func clearDialogError(state State) State {
	state.ErrorMessage = ""
	return state
}

func dialogInputIndex(state State, focusIdx int) (int, bool) {
	if focusIdx < 0 {
		return 0, false
	}
	if state.DescriptionEnabled && focusIdx == state.DescriptionIndex {
		return 0, false
	}
	inputIdx := focusIdx
	if state.DescriptionEnabled && focusIdx > state.DescriptionIndex {
		inputIdx--
	}
	if inputIdx < 0 || inputIdx >= len(state.Inputs) {
		return 0, false
	}
	return inputIdx, true
}

func pomodoroDialogInputIndex(state State, focusIdx int) (int, bool) {
	switch focusIdx {
	case 1:
		if state.PomodoroFocusChoice == pomodoroFocusCustomChoice {
			return 0, true
		}
	case 3:
		if state.PomodoroBreakChoice == pomodoroBreakCustomChoice {
			return 1, true
		}
	case 5:
		if state.PomodoroLongBreakChoice == pomodoroLongBreakCustomChoice {
			return 2, true
		}
	case 6:
		return 3, true
	case 7:
		if !pomodoroLongBreakDisabled(state) {
			return 4, true
		}
	}
	return 0, false
}

func ToggleEndSessionAdvanced(state State) State {
	if state.Kind != "end_session" {
		return state
	}
	if len(state.Inputs) > 1 {
		commit := state.Inputs[0].Value()
		input := newSessionDetailInput(state, "Commit message")
		input.SetValue(commit)
		input.Focus()
		state.Inputs = []textinput.Model{input}
		state.FocusIdx = 0
		return state
	}
	commit := state.Inputs[0].Value()
	inputs := []textinput.Model{
		newSessionDetailInput(state, "Commit message"),
		newSessionDetailInput(state, "Worked on"),
		newSessionDetailInput(state, "Outcome"),
		newSessionDetailInput(state, "Next step"),
		newSessionDetailInput(state, "Blockers"),
		newSessionDetailInput(state, "Links"),
	}
	inputs[0].SetValue(commit)
	inputs[0].Focus()
	state.Inputs = inputs
	state.FocusIdx = 0
	return state
}

func Update(
	state State,
	ctx UpdateContext,
	currentDate string,
	msg tea.KeyMsg,
) (State, *Action, string) {
	switch state.Kind {
	case "create_repo":
		return updateNameDescription(
			state,
			msg,
			"Repo name is required",
			func(name string, description *string) *Action {
				return &Action{Kind: "create_repo", Name: name, Description: description}
			},
		)
	case "edit_repo":
		return updateNameDescription(
			state,
			msg,
			"Repo name is required",
			func(name string, description *string) *Action {
				return &Action{
					Kind:        "edit_repo",
					RepoID:      state.RepoID,
					Name:        name,
					Description: description,
				}
			},
		)
	case "create_stream":
		return updateNameDescription(
			state,
			msg,
			"Stream name is required",
			func(name string, description *string) *Action {
				return &Action{
					Kind:        "create_stream",
					RepoID:      state.RepoID,
					Name:        name,
					Description: description,
				}
			},
		)
	case "edit_stream":
		return updateNameDescription(
			state,
			msg,
			"Stream name is required",
			func(name string, description *string) *Action {
				return &Action{
					Kind:        "edit_stream",
					RepoID:      state.RepoID,
					StreamID:    state.StreamID,
					Name:        name,
					Description: description,
				}
			},
		)
	case "confirm_delete":
		return updateConfirmDelete(state, msg)
	case "confirm_wipe":
		return updateConfirmWipe(state, msg)
	case "confirm_uninstall":
		return updateConfirmUninstall(state, msg)
	case "stash_list":
		return updateStashList(state, ctx, msg)
	case "issue_status":
		return updateIssueStatus(state, ctx, currentDate, msg)
	case "issue_status_note":
		return updateIssueStatusNote(state, currentDate, msg)
	case "end_session", "stash_session":
		return updateSessionMessage(state, ctx, currentDate, msg)
	case "timer_start_type":
		return updateTimerStartType(state, msg)
	case "pomodoro_start":
		return updatePomodoroStart(state, msg)
	case "hard_limit_expired":
		return updateHardLimitExpired(state, msg)
	case "hard_limit_extend":
		return updateHardLimitExtend(state, msg)
	case "amend_session":
		return updateAmendSession(state, msg)
	case "manual_session":
		return updateManualSession(state, msg)
	case "issue_session_transition":
		return updateIssueSessionTransition(state, ctx, currentDate, msg)
	case "pick_date":
		return updateDatePicker(state, ctx, currentDate, msg)
	case "create_issue_meta":
		return updateCreateIssueMeta(state, currentDate, msg)
	case "create_issue_default":
		return updateCreateIssueDefault(state, ctx, currentDate, msg)
	case "create_habit":
		return updateCreateHabit(state, ctx, msg)
	case "edit_habit":
		return updateHabitEditor(state, msg, "edit_habit")
	case "complete_habit":
		return updateHabitCompletion(state, msg)
	case "checkout_context":
		return updateCheckoutContext(state, ctx, msg)
	case "edit_issue":
		return updateEditIssue(state, currentDate, msg)
	case "create_checkin", "edit_checkin":
		return updateCheckIn(state, msg)
	case "export_report":
		return updateExportDaily(state, msg)
	case "export_report_category":
		return updateExportCategory(state, msg)
	case "export_preset":
		return updateExportPreset(state, msg)
	case "export_calendar_repo":
		return updateExportCalendarRepo(state, msg)
	case "edit_export_reports_dir":
		return updateSingleInput(
			state,
			msg,
			"Reports directory is required",
			func(value string) *Action {
				return &Action{Kind: "set_export_reports_dir", Path: value}
			},
		)
	case "edit_export_ics_dir":
		return updateSingleInput(
			state,
			msg,
			"ICS export directory is required",
			func(value string) *Action {
				return &Action{Kind: "set_export_ics_dir", Path: value}
			},
		)
	case "edit_date_display_format":
		return updateSingleInput(state, msg, "Date format is required", func(value string) *Action {
			return &Action{
				Kind:       "patch_setting",
				SettingKey: sharedtypes.CoreSettingsKeyDateDisplayFormat,
				Path:       value,
			}
		})
	case "edit_rest_protection":
		return updateRestProtection(state, currentDate, msg)
	case "create_momentum", "edit_momentum":
		return updateHabitStreaks(state, msg)
	case "edit_telemetry_settings":
		return updateTelemetrySettings(state, msg)
	case "onboarding":
		return updateOnboarding(state, msg)
	case "create_alert_reminder", "edit_alert_reminder":
		return updateAlertReminder(state, msg)
	case "view_entity":
		return updateViewEntity(state, msg)
	case "support_bundle_result":
		return updateSupportBundleResult(state, msg)
	case "view_jump":
		return updateViewJump(state, msg)
	case "beta_support":
		return updateBetaSupport(state, msg)
	case "stash_conflict_pick":
		return updateStashConflictPick(state, msg)
	case "stash_conflict":
		return updateStashConflict(state, msg)
	default:
		return state, nil, ""
	}
}

func newSessionDetailInput(state State, placeholder string) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
	input.CharLimit = 200
	input.Width = 48
	switch placeholder {
	case "YYYY-MM-DD":
		input = withDatePrompt(state, input)
	case "90 | 90m | 1h30m", "15m | 0m", "09:00", "10:45":
		input = withTimePrompt(state, input)
	}
	return input
}

func updateStashList(state State, ctx UpdateContext, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc", "q":
		return Close(state), nil, ""
	case "j", "down":
		if state.StashCursor < len(ctx.Stashes)-1 {
			state.StashCursor++
		}
	case "k", "up":
		if state.StashCursor > 0 {
			state.StashCursor--
		}
	case "enter":
		if len(ctx.Stashes) == 0 || state.StashCursor < 0 || state.StashCursor >= len(ctx.Stashes) {
			return state, nil, ""
		}
		return Close(state), &Action{Kind: "apply_stash", ID: ctx.Stashes[state.StashCursor].ID}, ""
	case "x":
		if len(ctx.Stashes) == 0 || state.StashCursor < 0 || state.StashCursor >= len(ctx.Stashes) {
			return state, nil, ""
		}
		return Close(state), &Action{Kind: "drop_stash", ID: ctx.Stashes[state.StashCursor].ID}, ""
	}
	return state, nil, ""
}

func updateIssueStatus(
	state State,
	ctx UpdateContext,
	currentDate string,
	msg tea.KeyMsg,
) (State, *Action, string) {
	switch msg.String() {
	case "esc", "q":
		return Close(state), nil, ""
	case "j", "down":
		if state.StatusCursor < len(state.StatusItems)-1 {
			state.StatusCursor++
		}
	case "k", "up":
		if state.StatusCursor > 0 {
			state.StatusCursor--
		}
	case "enter":
		if !ctx.HasSelectedIssue || len(state.StatusItems) == 0 || state.StatusCursor < 0 ||
			state.StatusCursor >= len(state.StatusItems) {
			return state, nil, ""
		}
		status := string(state.StatusItems[state.StatusCursor])
		switch status {
		case "blocked":
			return OpenIssueStatusNote(
				state,
				ctx.SelectedIssueID,
				ctx.SelectedStreamID,
				status,
				"Blocker reason",
				true,
			), nil, ""
		case "in_review":
			return OpenIssueStatusNote(
				state,
				ctx.SelectedIssueID,
				ctx.SelectedStreamID,
				status,
				"Review note (optional)",
				false,
			), nil, ""
		case "done":
			return OpenIssueStatusNote(
				state,
				ctx.SelectedIssueID,
				ctx.SelectedStreamID,
				status,
				"Completion note (optional)",
				false,
			), nil, ""
		case "abandoned":
			return OpenIssueStatusNote(
				state,
				ctx.SelectedIssueID,
				ctx.SelectedStreamID,
				status,
				"Abandon reason",
				true,
			), nil, ""
		default:
			return Close(
					state,
				), &Action{
					Kind:     "change_issue_status",
					IssueID:  ctx.SelectedIssueID,
					StreamID: ctx.SelectedStreamID,
					Status:   status,
				}, ""
		}
	}
	return state, nil, ""
}

func updateIssueStatusNote(
	state State,
	currentDate string,
	msg tea.KeyMsg,
) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			note := ValueToPointer(state.Inputs[0].Value())
			if state.StatusRequired && note == nil {
				return state, nil, state.StatusLabel + " is required"
			}
			return Close(
					state,
				), &Action{
					Kind:     "change_issue_status",
					IssueID:  state.IssueID,
					StreamID: state.StreamID,
					Status:   state.IssueStatus,
					Note:     note,
				}, ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[0], cmd = state.Inputs[0].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateSessionMessage(
	state State,
	ctx UpdateContext,
	currentDate string,
	msg tea.KeyMsg,
) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		if state.Parent == "hard_limit_expired" {
			return OpenHardLimitExpired(state, state.ViewName), nil, ""
		}
		return Close(state), nil, ""
	case "ctrl+e", "f2":
		return ToggleEndSessionAdvanced(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		if len(state.Inputs) == 1 {
			return state, nil, ""
		}
		dir := 1
		if msg.String() == "shift+tab" || msg.String() == "up" {
			dir = -1
		}
		state.FocusIdx = (state.FocusIdx + dir + len(state.Inputs)) % len(state.Inputs)
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			payload := EndSessionRequest(state.Inputs)
			kind := state.Kind
			state = Close(state)
			if kind == "end_session" {
				if !ctx.HasActiveIssue {
					return state, nil, "Active issue metadata unavailable"
				}
				return state, &Action{
					Kind:     "end_session",
					StreamID: ctx.ActiveIssueStream,
					Payload:  payload,
				}, ""
			}
			return state, &Action{Kind: "stash_session", Note: payload.CommitMessage}, ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateTimerStartType(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "s":
		state.ChoiceCursor = 0
		return updateTimerStartType(state, tea.KeyMsg{Type: tea.KeyEnter})
	case "p":
		state.ChoiceCursor = 1
		return updateTimerStartType(state, tea.KeyMsg{Type: tea.KeyEnter})
	case "j", "down":
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
		return clearDialogError(state), nil, ""
	case "k", "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
		return clearDialogError(state), nil, ""
	case "enter":
		switch currentChoiceValue(state) {
		case "stopwatch":
			return Close(state), &Action{
				Kind: "start_focus_session",
				TimerStart: &shareddto.TimerStartRequest{
					RepoID:   int64Ptr(state.RepoID),
					StreamID: int64Ptr(state.StreamID),
					IssueID:  int64Ptr(state.IssueID),
				},
			}, ""
		case "pomodoro":
			return OpenPomodoroStart(state, state.RepoID, state.StreamID, state.IssueID, state.ViewName), nil, ""
		default:
			return state, nil, "Choose a timer type"
		}
	default:
		if isDialogSubmitKey(state, msg.String()) {
			switch currentChoiceValue(state) {
			case "stopwatch":
				return Close(state), &Action{
					Kind: "start_focus_session",
					TimerStart: &shareddto.TimerStartRequest{
						RepoID:   int64Ptr(state.RepoID),
						StreamID: int64Ptr(state.StreamID),
						IssueID:  int64Ptr(state.IssueID),
					},
				}, ""
			case "pomodoro":
				return OpenPomodoroStart(state, state.RepoID, state.StreamID, state.IssueID, state.ViewName), nil, ""
			default:
				return state, nil, "Choose a timer type"
			}
		}
	}
	return state, nil, ""
}

func pomodoroLongBreakDisabled(state State) bool {
	return state.PomodoroLongBreakChoice == pomodoroLongBreakNoBreakChoice
}

func pomodoroBreakDisabled(state State) bool {
	return state.PomodoroBreakChoice == pomodoroBreakNoBreakChoice
}

func pomodoroFocusIdxEnabled(state State, idx int) bool {
	switch idx {
	case pomodoroFocusCustomIdx:
		return state.PomodoroFocusChoice == pomodoroFocusCustomChoice
	case pomodoroBreakCustomIdx:
		return state.PomodoroBreakChoice == pomodoroBreakCustomChoice
	case pomodoroLongBreakCustomIdx:
		return state.PomodoroLongBreakChoice == pomodoroLongBreakCustomChoice
	case pomodoroLongBreakRowIdx:
		return !pomodoroLongBreakDisabled(state) && !pomodoroBreakDisabled(state)
	case pomodoroCyclesRowIdx:
		return !pomodoroBreakDisabled(state)
	case pomodoroLongBreakCyclesIdx:
		return !pomodoroBreakDisabled(state) && !pomodoroLongBreakDisabled(state)
	default:
		return true
	}
}

func pomodoroNextFocusIdx(state State, idx int, dir int) int {
	if dir == 0 {
		return idx
	}
	limit := 8
	for i := 0; i < limit; i++ {
		idx = (idx + dir + limit) % limit
		if !pomodoroFocusIdxEnabled(state, idx) {
			continue
		}
		return idx
	}
	return idx
}

func pomodoroFocusChoiceForSeconds(seconds int) int {
	switch seconds {
	case 25 * 60:
		return 0
	case 50 * 60:
		return 1
	case 90 * 60:
		return 2
	default:
		return pomodoroFocusCustomChoice
	}
}

func pomodoroBreakChoiceForSeconds(seconds int) int {
	switch seconds {
	case 5 * 60:
		return 0
	case 10 * 60:
		return 1
	case 15 * 60:
		return 2
	case 0:
		return pomodoroBreakNoBreakChoice
	default:
		return pomodoroBreakCustomChoice
	}
}

func pomodoroLongBreakChoiceForSeconds(seconds int) int {
	switch seconds {
	case 15 * 60:
		return 0
	case 20 * 60:
		return 1
	case 30 * 60:
		return 2
	case 0:
		return pomodoroLongBreakNoBreakChoice
	default:
		return pomodoroLongBreakCustomChoice
	}
}

func pomodoroSeedDurationInput(seconds int, fallbackSeconds int) string {
	if seconds <= 0 {
		seconds = fallbackSeconds
	}
	minutes := seconds / 60
	return FormatDurationMinutesInput(&minutes)
}

func submitPomodoroStart(state State) (State, *Action, string) {
	values, err := pomodoroValuesFromState(state, true)
	if err != nil {
		return state, nil, err.Error()
	}
	if values.TotalSeconds <= 0 {
		return state, nil, "Total duration must be positive"
	}
	req := &shareddto.TimerStartRequest{
		RepoID:                         int64Ptr(state.RepoID),
		StreamID:                       int64Ptr(state.StreamID),
		IssueID:                        int64Ptr(state.IssueID),
		HardLimitTotalSeconds:          intPtr(values.TotalSeconds),
		HardLimitWorkSeconds:           intPtr(values.FocusSeconds),
		HardLimitBreakSeconds:          intPtr(values.BreakSeconds),
		HardLimitLongBreakSeconds:      intPtr(values.LongBreakSeconds),
		HardLimitCyclesBeforeLongBreak: intPtr(values.CyclesBeforeLongBreak),
	}
	return Close(state), &Action{Kind: "start_focus_session", TimerStart: req}, ""
}

func refreshPomodoroPreviewFields(state State, inputIdx int) State {
	switch inputIdx {
	case 0:
		if state.PomodoroFocusChoice == pomodoroFocusCustomChoice {
			values, err := pomodoroValuesFromState(state, false)
			if err == nil {
				state.PomodoroFocusSeconds = values.FocusSeconds
			}
		}
	case 1:
		if state.PomodoroBreakChoice == pomodoroBreakCustomChoice {
			values, err := pomodoroValuesFromState(state, false)
			if err == nil {
				state.PomodoroBreakSeconds = values.BreakSeconds
			}
		}
	case 2:
		if state.PomodoroLongBreakChoice == pomodoroLongBreakCustomChoice {
			values, err := pomodoroValuesFromState(state, false)
			if err == nil {
				state.PomodoroLongBreakSeconds = values.LongBreakSeconds
			}
		}
	case 3:
		values, err := pomodoroValuesFromState(state, false)
		if err == nil {
			state.PomodoroCycles = values.Cycles
		}
	case 4:
		values, err := pomodoroValuesFromState(state, false)
		if err == nil {
			state.PomodoroCyclesBeforeLongBreak = values.CyclesBeforeLongBreak
		}
	}
	return state
}

func updatePomodoroStart(state State, msg tea.KeyMsg) (State, *Action, string) {
	if msg.String() == "left" {
		switch state.FocusIdx {
		case pomodoroFocusCustomIdx:
			if state.Inputs[0].Position() == 0 {
				state.FocusIdx = pomodoroFocusRowIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
		case pomodoroBreakCustomIdx:
			if state.Inputs[1].Position() == 0 {
				state.FocusIdx = pomodoroBreakRowIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
		case pomodoroLongBreakCustomIdx:
			if state.Inputs[2].Position() == 0 {
				state.FocusIdx = pomodoroLongBreakRowIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
		}
	}
	switch msg.String() {
	case "esc":
		switch state.FocusIdx {
		case pomodoroFocusCustomIdx:
			state.FocusIdx = pomodoroFocusRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroBreakCustomIdx:
			state.FocusIdx = pomodoroBreakRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroLongBreakCustomIdx:
			state.FocusIdx = pomodoroLongBreakRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		default:
			return OpenTimerStartType(state, state.RepoID, state.StreamID, state.IssueID, state.ViewName), nil, ""
		}
	case "tab", "down":
		state.FocusIdx = pomodoroNextFocusIdx(state, state.FocusIdx, 1)
		return SyncDialogFocus(clearDialogError(state)), nil, ""
	case "shift+tab", "up":
		state.FocusIdx = pomodoroNextFocusIdx(state, state.FocusIdx, -1)
		return SyncDialogFocus(clearDialogError(state)), nil, ""
	case "left":
		switch state.FocusIdx {
		case pomodoroFocusRowIdx:
			if state.PomodoroFocusChoice > 0 {
				state.PomodoroFocusChoice--
			}
			state.PomodoroFocusSeconds = []int{25 * 60, 50 * 60, 90 * 60, state.PomodoroFocusSeconds}[state.PomodoroFocusChoice]
			return clearDialogError(state), nil, ""
		case pomodoroBreakRowIdx:
			if state.PomodoroBreakChoice > 0 {
				state.PomodoroBreakChoice--
			}
			switch state.PomodoroBreakChoice {
			case 0:
				state.PomodoroBreakSeconds = 5 * 60
			case 1:
				state.PomodoroBreakSeconds = 10 * 60
			case 2:
				state.PomodoroBreakSeconds = 15 * 60
			case pomodoroBreakNoBreakChoice:
				state.PomodoroBreakSeconds = 0
			}
			return clearDialogError(state), nil, ""
		case pomodoroLongBreakRowIdx:
			if state.PomodoroLongBreakChoice > 0 {
				state.PomodoroLongBreakChoice--
			}
			switch state.PomodoroLongBreakChoice {
			case 0:
				state.PomodoroLongBreakSeconds = 15 * 60
			case 1:
				state.PomodoroLongBreakSeconds = 20 * 60
			case 2:
				state.PomodoroLongBreakSeconds = 30 * 60
			case pomodoroLongBreakNoBreakChoice:
				state.PomodoroLongBreakSeconds = 0
			}
			if pomodoroLongBreakDisabled(state) && state.FocusIdx == pomodoroLongBreakCyclesIdx {
				state.FocusIdx = pomodoroLongBreakRowIdx
			}
			return clearDialogError(state), nil, ""
		}
	case "right":
		switch state.FocusIdx {
		case pomodoroFocusRowIdx:
			if state.PomodoroFocusChoice < pomodoroFocusCustomChoice {
				state.PomodoroFocusChoice++
			}
			switch state.PomodoroFocusChoice {
			case 0:
				state.PomodoroFocusSeconds = 25 * 60
			case 1:
				state.PomodoroFocusSeconds = 50 * 60
			case 2:
				state.PomodoroFocusSeconds = 90 * 60
			case pomodoroFocusCustomChoice:
				state.FocusIdx = pomodoroFocusCustomIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
			return clearDialogError(state), nil, ""
		case pomodoroBreakRowIdx:
			if state.PomodoroBreakChoice < pomodoroBreakCustomChoice {
				state.PomodoroBreakChoice++
			}
			switch state.PomodoroBreakChoice {
			case 0:
				state.PomodoroBreakSeconds = 5 * 60
			case 1:
				state.PomodoroBreakSeconds = 10 * 60
			case 2:
				state.PomodoroBreakSeconds = 15 * 60
			case pomodoroBreakNoBreakChoice:
				state.PomodoroBreakSeconds = 0
			case pomodoroBreakCustomChoice:
				state.FocusIdx = pomodoroBreakCustomIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
			return clearDialogError(state), nil, ""
		case pomodoroLongBreakRowIdx:
			if state.PomodoroLongBreakChoice < pomodoroLongBreakCustomChoice {
				state.PomodoroLongBreakChoice++
			}
			switch state.PomodoroLongBreakChoice {
			case 0:
				state.PomodoroLongBreakSeconds = 15 * 60
			case 1:
				state.PomodoroLongBreakSeconds = 20 * 60
			case 2:
				state.PomodoroLongBreakSeconds = 30 * 60
			case pomodoroLongBreakNoBreakChoice:
				state.PomodoroLongBreakSeconds = 0
			case pomodoroLongBreakCustomChoice:
				state.FocusIdx = pomodoroLongBreakCustomIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
			return clearDialogError(state), nil, ""
		}
	case "enter":
		switch state.FocusIdx {
		case pomodoroFocusRowIdx:
			if state.PomodoroFocusChoice == pomodoroFocusCustomChoice {
				state.FocusIdx = pomodoroFocusCustomIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
			state.FocusIdx = pomodoroBreakRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroFocusCustomIdx:
			state.FocusIdx = pomodoroBreakRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroBreakRowIdx:
			if state.PomodoroBreakChoice == pomodoroBreakCustomChoice {
				state.FocusIdx = pomodoroBreakCustomIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
			if pomodoroBreakDisabled(state) {
				return submitPomodoroStart(state)
			}
			if pomodoroLongBreakDisabled(state) {
				state.FocusIdx = pomodoroCyclesRowIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
			state.FocusIdx = pomodoroLongBreakRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroBreakCustomIdx:
			if pomodoroLongBreakDisabled(state) {
				if pomodoroBreakDisabled(state) {
					return submitPomodoroStart(state)
				}
				state.FocusIdx = pomodoroCyclesRowIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
			state.FocusIdx = pomodoroLongBreakRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroLongBreakRowIdx:
			if state.PomodoroLongBreakChoice == pomodoroLongBreakCustomChoice {
				state.FocusIdx = pomodoroLongBreakCustomIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
			if pomodoroBreakDisabled(state) {
				return submitPomodoroStart(state)
			}
			state.FocusIdx = pomodoroCyclesRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroLongBreakCustomIdx:
			if pomodoroBreakDisabled(state) {
				return submitPomodoroStart(state)
			}
			state.FocusIdx = pomodoroCyclesRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroCyclesRowIdx:
			if !pomodoroLongBreakDisabled(state) {
				state.FocusIdx = pomodoroLongBreakCyclesIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
			return submitPomodoroStart(state)
		case pomodoroLongBreakCyclesIdx:
			return submitPomodoroStart(state)
		}
	default:
		if isDialogSubmitKey(state, msg.String()) {
			return submitPomodoroStart(state)
		}
	}
	var cmd tea.Cmd
	if inputIdx, ok := pomodoroDialogInputIndex(state, state.FocusIdx); ok {
		state.Inputs[inputIdx], cmd = state.Inputs[inputIdx].Update(msg)
		_ = cmd
		state = refreshPomodoroPreviewFields(state, inputIdx)
		return clearDialogError(state), nil, ""
	}
	return clearDialogError(state), nil, ""
}

func updateHardLimitExpired(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "j", "down":
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
		return clearDialogError(state), nil, ""
	case "k", "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
		return clearDialogError(state), nil, ""
	case "c":
		return OpenSessionMessageWithParent(state, "end_session", "hard_limit_expired"), nil, ""
	case "z":
		return OpenSessionMessageWithParent(state, "stash_session", "hard_limit_expired"), nil, ""
	case "e":
		return OpenHardLimitExtend(state), nil, ""
	case "enter":
		switch currentChoiceValue(state) {
		case "commit":
			return OpenSessionMessageWithParent(state, "end_session", "hard_limit_expired"), nil, ""
		case "stash":
			return OpenSessionMessageWithParent(
				state,
				"stash_session",
				"hard_limit_expired",
			), nil, ""
		default:
			return OpenHardLimitExtend(state), nil, ""
		}
	default:
		if isDialogSubmitKey(state, msg.String()) {
			switch currentChoiceValue(state) {
			case "commit":
				return OpenSessionMessageWithParent(
					state,
					"end_session",
					"hard_limit_expired",
				), nil, ""
			case "stash":
				return OpenSessionMessageWithParent(
					state,
					"stash_session",
					"hard_limit_expired",
				), nil, ""
			default:
				return OpenHardLimitExtend(state), nil, ""
			}
		}
	}
	return state, nil, ""
}

func updateHardLimitExtend(state State, msg tea.KeyMsg) (State, *Action, string) {
	if msg.String() == "left" {
		switch state.FocusIdx {
		case pomodoroFocusCustomIdx:
			if state.Inputs[0].Position() == 0 {
				state.FocusIdx = pomodoroFocusRowIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
		case pomodoroBreakCustomIdx:
			if state.Inputs[1].Position() == 0 {
				state.FocusIdx = pomodoroBreakRowIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
		case pomodoroLongBreakCustomIdx:
			if state.Inputs[2].Position() == 0 {
				state.FocusIdx = pomodoroLongBreakRowIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
		}
	}
	switch msg.String() {
	case "esc":
		switch state.FocusIdx {
		case pomodoroFocusCustomIdx:
			state.FocusIdx = pomodoroFocusRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroBreakCustomIdx:
			state.FocusIdx = pomodoroBreakRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroLongBreakCustomIdx:
			state.FocusIdx = pomodoroLongBreakRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		default:
			return OpenHardLimitExpired(state, state.ViewName), nil, ""
		}
	case "tab", "down":
		state.FocusIdx = pomodoroNextFocusIdx(state, state.FocusIdx, 1)
		return SyncDialogFocus(clearDialogError(state)), nil, ""
	case "shift+tab", "up":
		state.FocusIdx = pomodoroNextFocusIdx(state, state.FocusIdx, -1)
		return SyncDialogFocus(clearDialogError(state)), nil, ""
	case "left":
		switch state.FocusIdx {
		case pomodoroFocusRowIdx:
			if state.PomodoroFocusChoice > 0 {
				state.PomodoroFocusChoice--
			}
			state.PomodoroFocusSeconds = []int{25 * 60, 50 * 60, 90 * 60, state.PomodoroFocusSeconds}[state.PomodoroFocusChoice]
			return clearDialogError(state), nil, ""
		case pomodoroBreakRowIdx:
			if state.PomodoroBreakChoice > 0 {
				state.PomodoroBreakChoice--
			}
			switch state.PomodoroBreakChoice {
			case 0:
				state.PomodoroBreakSeconds = 5 * 60
			case 1:
				state.PomodoroBreakSeconds = 10 * 60
			case 2:
				state.PomodoroBreakSeconds = 15 * 60
			case pomodoroBreakNoBreakChoice:
				state.PomodoroBreakSeconds = 0
			}
			return clearDialogError(state), nil, ""
		case pomodoroLongBreakRowIdx:
			if state.PomodoroLongBreakChoice > 0 {
				state.PomodoroLongBreakChoice--
			}
			switch state.PomodoroLongBreakChoice {
			case 0:
				state.PomodoroLongBreakSeconds = 15 * 60
			case 1:
				state.PomodoroLongBreakSeconds = 20 * 60
			case 2:
				state.PomodoroLongBreakSeconds = 30 * 60
			case pomodoroLongBreakNoBreakChoice:
				state.PomodoroLongBreakSeconds = 0
			}
			return clearDialogError(state), nil, ""
		}
	case "right":
		switch state.FocusIdx {
		case pomodoroFocusRowIdx:
			if state.PomodoroFocusChoice < pomodoroFocusCustomChoice {
				state.PomodoroFocusChoice++
			}
			switch state.PomodoroFocusChoice {
			case 0:
				state.PomodoroFocusSeconds = 25 * 60
			case 1:
				state.PomodoroFocusSeconds = 50 * 60
			case 2:
				state.PomodoroFocusSeconds = 90 * 60
			case pomodoroFocusCustomChoice:
				state.FocusIdx = pomodoroFocusCustomIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
			return clearDialogError(state), nil, ""
		case pomodoroBreakRowIdx:
			if state.PomodoroBreakChoice < pomodoroBreakCustomChoice {
				state.PomodoroBreakChoice++
			}
			switch state.PomodoroBreakChoice {
			case 0:
				state.PomodoroBreakSeconds = 5 * 60
			case 1:
				state.PomodoroBreakSeconds = 10 * 60
			case 2:
				state.PomodoroBreakSeconds = 15 * 60
			case pomodoroBreakNoBreakChoice:
				state.PomodoroBreakSeconds = 0
			case pomodoroBreakCustomChoice:
				state.FocusIdx = pomodoroBreakCustomIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
			return clearDialogError(state), nil, ""
		case pomodoroLongBreakRowIdx:
			if state.PomodoroLongBreakChoice < pomodoroLongBreakCustomChoice {
				state.PomodoroLongBreakChoice++
			}
			switch state.PomodoroLongBreakChoice {
			case 0:
				state.PomodoroLongBreakSeconds = 15 * 60
			case 1:
				state.PomodoroLongBreakSeconds = 20 * 60
			case 2:
				state.PomodoroLongBreakSeconds = 30 * 60
			case pomodoroLongBreakNoBreakChoice:
				state.PomodoroLongBreakSeconds = 0
			case pomodoroLongBreakCustomChoice:
				state.FocusIdx = pomodoroLongBreakCustomIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
			return clearDialogError(state), nil, ""
		}
	case "enter":
		switch state.FocusIdx {
		case pomodoroFocusRowIdx:
			if state.PomodoroFocusChoice == pomodoroFocusCustomChoice {
				state.FocusIdx = pomodoroFocusCustomIdx
			} else {
				state.FocusIdx = pomodoroBreakRowIdx
			}
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroFocusCustomIdx:
			state.FocusIdx = pomodoroBreakRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroBreakRowIdx:
			if state.PomodoroBreakChoice == pomodoroBreakCustomChoice {
				state.FocusIdx = pomodoroBreakCustomIdx
			} else if pomodoroBreakDisabled(state) {
				state.FocusIdx = pomodoroCyclesRowIdx
			} else if pomodoroLongBreakDisabled(state) {
				state.FocusIdx = pomodoroCyclesRowIdx
			} else {
				state.FocusIdx = pomodoroLongBreakRowIdx
			}
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroBreakCustomIdx:
			if pomodoroBreakDisabled(state) {
				state.FocusIdx = pomodoroCyclesRowIdx
			} else if pomodoroLongBreakDisabled(state) {
				state.FocusIdx = pomodoroCyclesRowIdx
			} else {
				state.FocusIdx = pomodoroLongBreakRowIdx
			}
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroLongBreakRowIdx:
			if state.PomodoroLongBreakChoice == pomodoroLongBreakCustomChoice {
				state.FocusIdx = pomodoroLongBreakCustomIdx
			} else {
				state.FocusIdx = pomodoroCyclesRowIdx
			}
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroLongBreakCustomIdx:
			state.FocusIdx = pomodoroCyclesRowIdx
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		case pomodoroCyclesRowIdx:
			if !pomodoroLongBreakDisabled(state) {
				state.FocusIdx = pomodoroLongBreakCyclesIdx
				return SyncDialogFocus(clearDialogError(state)), nil, ""
			}
			req, err := pomodoroExtendRequestFromState(state)
			if err != nil {
				return state, nil, err.Error()
			}
			return Close(state), &Action{Kind: "extend_hard_limit", AdditionalSessions: req.AdditionalSessions, TimerExtend: req}, ""
		case pomodoroLongBreakCyclesIdx:
			req, err := pomodoroExtendRequestFromState(state)
			if err != nil {
				return state, nil, err.Error()
			}
			return Close(state), &Action{Kind: "extend_hard_limit", AdditionalSessions: req.AdditionalSessions, TimerExtend: req}, ""
		}
	default:
		if isDialogSubmitKey(state, msg.String()) {
			req, err := pomodoroExtendRequestFromState(state)
			if err != nil {
				return state, nil, err.Error()
			}
			return Close(state), &Action{Kind: "extend_hard_limit", AdditionalSessions: req.AdditionalSessions, TimerExtend: req}, ""
		}
	}
	var cmd tea.Cmd
	if inputIdx, ok := pomodoroDialogInputIndex(state, state.FocusIdx); ok {
		state.Inputs[inputIdx], cmd = state.Inputs[inputIdx].Update(msg)
		_ = cmd
		state = refreshPomodoroPreviewFields(state, inputIdx)
		return clearDialogError(state), nil, ""
	}
	return clearDialogError(state), nil, ""
}

func currentChoiceValue(state State) string {
	if state.ChoiceCursor < 0 || state.ChoiceCursor >= len(state.ChoiceValues) {
		return ""
	}
	return strings.TrimSpace(state.ChoiceValues[state.ChoiceCursor])
}

func intPtr(value int) *int {
	if value <= 0 {
		return nil
	}
	return &value
}

func int64Ptr(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}

func updateIssueSessionTransition(
	state State,
	ctx UpdateContext,
	currentDate string,
	msg tea.KeyMsg,
) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "n", "N":
		if state.IssueStatus != "done" && state.IssueStatus != "abandoned" {
			return Close(state), nil, ""
		}
	case "y", "Y":
		if state.IssueStatus != "done" && state.IssueStatus != "abandoned" {
			return Close(
					state,
				), &Action{
					Kind:     "change_issue_status_and_end_session",
					IssueID:  state.IssueID,
					StreamID: ctx.ActiveIssueStream,
					Status:   state.IssueStatus,
					Payload:  shareddto.EndSessionRequest{},
				}, ""
		}
	default:
		if isDialogSubmitKey(state, msg.String()) {
			note := ValueToPointer("")
			if len(state.Inputs) > 0 {
				note = ValueToPointer(state.Inputs[0].Value())
			}
			if state.IssueStatus == "abandoned" && note == nil {
				return state, nil, "Abandon reason is required"
			}
			return Close(
					state,
				), &Action{
					Kind:     "change_issue_status_and_end_session",
					IssueID:  state.IssueID,
					StreamID: ctx.ActiveIssueStream,
					Status:   state.IssueStatus,
					Note:     note,
					Payload:  shareddto.EndSessionRequest{CommitMessage: note},
				}, ""
		}
	}
	if (state.IssueStatus == "done" || state.IssueStatus == "abandoned") && len(state.Inputs) > 0 {
		var cmd tea.Cmd
		state.Inputs[0], cmd = state.Inputs[0].Update(msg)
		_ = cmd
	}
	return clearDialogError(state), nil, ""
}

func updateHabitEditor(state State, msg tea.KeyMsg, kind string) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		state.FocusIdx = (state.FocusIdx + ternaryDir(msg.String()) + 4) % 4
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	}
	if isDialogSubmitKey(state, msg.String()) {
		name := strings.TrimSpace(state.Inputs[0].Value())
		if name == "" {
			return state, nil, "Habit name is required"
		}
		scheduleType, weekdays, err := ParseHabitSchedule(state.Inputs[1].Value())
		if err != nil {
			return state, nil, err.Error()
		}
		target, err := ParseOptionalDurationMinutes(state.Inputs[2].Value(), "Target")
		if err != nil {
			return state, nil, err.Error()
		}
		action := &Action{
			Kind:        kind,
			HabitID:     state.HabitID,
			StreamID:    state.StreamID,
			Name:        name,
			Description: ValueToPointer(strings.TrimSpace(state.Description.Value())),
			Status:      scheduleType,
			Weekdays:    weekdays,
			Active:      state.StatusLabel != "inactive",
			Estimate:    target,
		}
		return Close(state), action, ""
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
	return clearDialogError(state), nil, ""
}

func updateCreateHabit(state State, ctx UpdateContext, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
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
			return clearDialogError(state), nil, ""
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
		name := strings.TrimSpace(state.Inputs[2].Value())
		if repoName == "" || streamName == "" || name == "" {
			return state, nil, "Repo, stream, and habit name are required"
		}
		scheduleType, weekdays, err := ParseHabitSchedule(state.Inputs[3].Value())
		if err != nil {
			return state, nil, err.Error()
		}
		target, err := ParseOptionalDurationMinutes(state.Inputs[4].Value(), "Target")
		if err != nil {
			return state, nil, err.Error()
		}
		return Close(state), &Action{
			Kind:        "create_habit",
			RepoName:    repoName,
			StreamName:  streamName,
			Name:        name,
			Description: ValueToPointer(strings.TrimSpace(state.Description.Value())),
			Status:      scheduleType,
			Weekdays:    weekdays,
			Estimate:    target,
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
