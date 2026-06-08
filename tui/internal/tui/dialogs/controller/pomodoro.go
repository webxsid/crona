package controller

import (
	"fmt"
	"strconv"
	"strings"

	shareddto "crona/shared/dto"
	helperpkg "crona/tui/internal/tui/helpers"
)

const (
	pomodoroFocusCustomChoice      = 3
	pomodoroBreakNoBreakChoice     = 3
	pomodoroBreakCustomChoice      = 4
	pomodoroLongBreakNoBreakChoice = 3
	pomodoroLongBreakCustomChoice  = 4
	pomodoroFocusRowIdx            = 0
	pomodoroFocusCustomIdx         = 1
	pomodoroBreakRowIdx            = 2
	pomodoroBreakCustomIdx         = 3
	pomodoroLongBreakRowIdx        = 4
	pomodoroLongBreakCustomIdx     = 5
	pomodoroCyclesRowIdx           = 6
	pomodoroLongBreakCyclesIdx     = 7
)

type PomodoroDialogViewModel struct {
	FocusChoices           []string
	BreakChoices           []string
	LongBreakChoices       []string
	FocusCustomChoice      int
	BreakCustomChoice      int
	LongBreakCustomChoice  int
	LongBreakSelected      int
	FocusRowActive         bool
	FocusCustomActive      bool
	BreakRowActive         bool
	BreakCustomActive      bool
	LongBreakRowActive     bool
	LongBreakCustomActive  bool
	CyclesRowActive        bool
	LongBreakCyclesActive  bool
	CyclesDisabled         bool
	LongBreakForcedOff     bool
	LongBreakDisabled      bool
	ShowSummary            bool
	FocusDisplay           string
	ShortBreakDisplay      string
	LongBreakDisplay       string
	CyclesSummary          string
	EstimatedTotalDuration string
}

type pomodoroValues struct {
	FocusSeconds          int
	BreakSeconds          int
	LongBreakSeconds      int
	Cycles                int
	CyclesBeforeLongBreak int
	TotalSeconds          int
}

func pomodoroFocusChoices() []string {
	return []string{"25m", "50m", "90m", "Custom"}
}

func pomodoroBreakChoices() []string {
	return []string{"5m", "10m", "15m", "No Break", "Custom"}
}

func pomodoroLongBreakChoices() []string {
	return []string{"15m", "20m", "30m", "No Break", "Custom"}
}

func BuildPomodoroDialogViewModel(state State) PomodoroDialogViewModel {
	values, _ := pomodoroValuesFromState(state, false)
	longBreakForcedOff := pomodoroBreakDisabled(state)
	longBreakSelected := state.PomodoroLongBreakChoice
	if longBreakForcedOff {
		longBreakSelected = pomodoroLongBreakNoBreakChoice
	}
	return PomodoroDialogViewModel{
		FocusChoices:          pomodoroFocusChoices(),
		BreakChoices:          pomodoroBreakChoices(),
		LongBreakChoices:      pomodoroLongBreakChoices(),
		FocusCustomChoice:     pomodoroFocusCustomChoice,
		BreakCustomChoice:     pomodoroBreakCustomChoice,
		LongBreakCustomChoice: pomodoroLongBreakCustomChoice,
		LongBreakSelected:     longBreakSelected,
		FocusRowActive:        state.FocusIdx == pomodoroFocusRowIdx,
		FocusCustomActive:     state.FocusIdx == pomodoroFocusCustomIdx,
		BreakRowActive:        state.FocusIdx == pomodoroBreakRowIdx,
		BreakCustomActive:     state.FocusIdx == pomodoroBreakCustomIdx,
		LongBreakRowActive:    state.FocusIdx == pomodoroLongBreakRowIdx,
		LongBreakCustomActive: state.FocusIdx == pomodoroLongBreakCustomIdx,
		CyclesRowActive:       state.FocusIdx == pomodoroCyclesRowIdx,
		LongBreakCyclesActive: state.FocusIdx == pomodoroLongBreakCyclesIdx,
		CyclesDisabled:        pomodoroBreakDisabled(state),
		LongBreakForcedOff:    longBreakForcedOff,
		LongBreakDisabled:     pomodoroBreakDisabled(state) || pomodoroLongBreakDisabled(state),
		ShowSummary:           values.FocusSeconds > 0 || values.BreakSeconds > 0,
		FocusDisplay: pomodoroDisplayDuration(
			state.PomodoroFocusChoice,
			pomodoroFocusCustomChoice,
			-1,
			state.Inputs[0].Value(),
			values.FocusSeconds,
			"",
			"Focus",
		),
		ShortBreakDisplay: pomodoroDisplayDuration(
			state.PomodoroBreakChoice,
			pomodoroBreakCustomChoice,
			pomodoroBreakNoBreakChoice,
			state.Inputs[1].Value(),
			values.BreakSeconds,
			"No Short Break",
			"Short Break",
		),
		LongBreakDisplay: pomodoroDisplayDuration(
			longBreakSelected,
			pomodoroLongBreakCustomChoice,
			pomodoroLongBreakNoBreakChoice,
			state.Inputs[2].Value(),
			values.LongBreakSeconds,
			"No Long Break",
			"Long Break",
		),
		CyclesSummary:          pomodoroCyclesSummary(state, values),
		EstimatedTotalDuration: pomodoroEstimatedTotalDuration(values),
	}
}

func pomodoroDisplayDuration(
	choice, customChoice, noBreakChoice int,
	input string,
	seconds int,
	emptyValue, label string,
) string {
	if noBreakChoice >= 0 && choice == noBreakChoice {
		return emptyValue
	}
	if choice == customChoice {
		if value := strings.TrimSpace(input); value != "" {
			return fmt.Sprintf("%s %s", value, label)
		}
	}
	return fmt.Sprintf("%s %s", helperpkg.FormatCompactDurationSeconds(seconds), label)
}

func pomodoroCyclesSummary(state State, values pomodoroValues) string {
	if values.BreakSeconds <= 0 {
		return "Uninterrupted focus"
	}

	cycles := strings.TrimSpace(state.Inputs[3].Value())
	if cycles == "" {
		cycles = strconv.Itoa(values.Cycles)
	}
	if cycles == "" || cycles == "0" {
		return ""
	}

	if values.LongBreakSeconds <= 0 {
		return fmt.Sprintf("%s cycles · no long break", cycles)
	}

	every := strings.TrimSpace(state.Inputs[4].Value())
	if every == "" {
		every = strconv.Itoa(values.CyclesBeforeLongBreak)
	}

	if every == "" || every == "0" {
		return fmt.Sprintf("%s cycles", cycles)
	}

	return fmt.Sprintf(
		"%s cycles · long break every %s cycles",
		cycles,
		every,
	)
}

func pomodoroEstimatedTotalDuration(values pomodoroValues) string {
	if values.TotalSeconds <= 0 {
		return ""
	}
	return "Total Duration " + helperpkg.FormatCompactDurationSeconds(values.TotalSeconds)
}

func inferPomodoroCycles(totalSeconds, focusSeconds, breakSeconds, longBreakSeconds, cyclesBeforeLongBreak int) int {
	if totalSeconds <= 0 || focusSeconds <= 0 {
		return 1
	}
	if breakSeconds <= 0 {
		return 1
	}
	perCycle := focusSeconds + breakSeconds
	if perCycle <= 0 {
		return 1
	}
	for cycles := 1; cycles <= 24; cycles++ {
		candidate := cycles * perCycle
		if longBreakSeconds > 0 && cyclesBeforeLongBreak > 0 {
			candidate += (cycles / cyclesBeforeLongBreak) * (longBreakSeconds - breakSeconds)
		}
		if candidate == totalSeconds {
			return cycles
		}
	}
	return max(1, totalSeconds/perCycle)
}

func pomodoroExtendEstimatedTotalDuration(state State) string {
	config, sessions, err := pomodoroExtendConfigFromState(state, false)
	if err != nil || sessions <= 0 {
		return ""
	}
	addedSeconds := pomodoroExtendAddedSeconds(config, sessions)
	if addedSeconds <= 0 {
		return ""
	}
	return "Added Duration " + helperpkg.FormatCompactDurationSeconds(addedSeconds)
}

func PreviewPomodoroExtendDuration(state State) string {
	return pomodoroExtendEstimatedTotalDuration(state)
}

func pomodoroExtendRequestFromState(state State) (*shareddto.TimerExtendRequest, error) {
	config, sessions, err := pomodoroExtendConfigFromState(state, true)
	if err != nil {
		return nil, err
	}
	additionalSeconds := pomodoroExtendAddedSeconds(config, sessions)
	if additionalSeconds <= 0 {
		return nil, fmt.Errorf("added duration must be positive")
	}
	totalPerSession := config.FocusSeconds + config.BreakSeconds
	if totalPerSession <= 0 {
		totalPerSession = config.FocusSeconds
	}
	return &shareddto.TimerExtendRequest{
		AdditionalSeconds:              additionalSeconds,
		AdditionalSessions:             sessions,
		HardLimitTotalSeconds:          intPtr(totalPerSession),
		HardLimitWorkSeconds:           intPtr(config.FocusSeconds),
		HardLimitBreakSeconds:          intPtr(config.BreakSeconds),
		HardLimitLongBreakSeconds:      intPtr(config.LongBreakSeconds),
		HardLimitCyclesBeforeLongBreak: intPtr(config.CyclesBeforeLongBreak),
	}, nil
}

func pomodoroExtendConfigFromState(state State, validate bool) (pomodoroValues, int, error) {
	values := pomodoroValues{
		FocusSeconds:          state.PomodoroFocusSeconds,
		BreakSeconds:          state.PomodoroBreakSeconds,
		LongBreakSeconds:      state.PomodoroLongBreakSeconds,
		CyclesBeforeLongBreak: state.PomodoroCyclesBeforeLongBreak,
	}

	if state.PomodoroFocusChoice == pomodoroFocusCustomChoice {
		parsed, err := ParseDurationInput(state.Inputs[0].Value(), validate, "Focus duration")
		if err != nil {
			return values, 0, err
		}
		if parsed > 0 {
			values.FocusSeconds = parsed
		}
	}

	switch state.PomodoroBreakChoice {
	case pomodoroBreakCustomChoice:
		parsed, err := ParseDurationInput(state.Inputs[1].Value(), validate, "Short break duration")
		if err != nil {
			return values, 0, err
		}
		if parsed > 0 {
			values.BreakSeconds = parsed
		}
	case pomodoroBreakNoBreakChoice:
		values.BreakSeconds = 0
	}

	switch state.PomodoroLongBreakChoice {
	case pomodoroLongBreakCustomChoice:
		parsed, err := ParseDurationInput(state.Inputs[2].Value(), validate, "Long break duration")
		if err != nil {
			return values, 0, err
		}
		if parsed > 0 {
			values.LongBreakSeconds = parsed
		}
	case pomodoroLongBreakNoBreakChoice:
		values.LongBreakSeconds = 0
	}

	if values.BreakSeconds <= 0 {
		values.LongBreakSeconds = 0
		values.CyclesBeforeLongBreak = 0
	}
	if values.LongBreakSeconds <= 0 {
		values.CyclesBeforeLongBreak = 0
	} else {
		parsed, err := ParsePositiveIntInput(state.Inputs[4].Value(), validate, "Long Break")
		if err != nil {
			return values, 0, err
		}
		if parsed > 0 {
			values.CyclesBeforeLongBreak = parsed
		}
	}
	sessions, err := ParsePositiveIntInput(state.Inputs[3].Value(), validate, "Cycles")
	if err != nil {
		return values, 0, err
	}
	if sessions <= 0 {
		return values, 0, fmt.Errorf("cycles must be positive")
	}
	return values, sessions, nil
}

func pomodoroExtendAddedSeconds(config pomodoroValues, sessions int) int {
	if sessions <= 0 || config.FocusSeconds <= 0 {
		return 0
	}
	if config.BreakSeconds <= 0 {
		return config.FocusSeconds * sessions
	}
	addedSeconds := sessions * (config.FocusSeconds + config.BreakSeconds)
	if config.LongBreakSeconds > 0 && config.CyclesBeforeLongBreak > 0 {
		addedSeconds += (sessions / config.CyclesBeforeLongBreak) * (config.LongBreakSeconds - config.BreakSeconds)
	}
	return addedSeconds
}

func pomodoroValuesFromState(state State, validate bool) (pomodoroValues, error) {
	values := pomodoroValues{
		FocusSeconds:          state.PomodoroFocusSeconds,
		BreakSeconds:          state.PomodoroBreakSeconds,
		LongBreakSeconds:      state.PomodoroLongBreakSeconds,
		Cycles:                state.PomodoroCycles,
		CyclesBeforeLongBreak: state.PomodoroCyclesBeforeLongBreak,
	}

	if state.PomodoroFocusChoice == pomodoroFocusCustomChoice {
		parsed, err := ParseDurationInput(state.Inputs[0].Value(), validate, "Focus duration")
		if err != nil {
			return values, err
		}
		if parsed > 0 {
			values.FocusSeconds = parsed
		}
	}

	switch state.PomodoroBreakChoice {
	case pomodoroBreakCustomChoice:
		parsed, err := ParseDurationInput(state.Inputs[1].Value(), validate, "Short break duration")
		if err != nil {
			return values, err
		}
		if parsed > 0 {
			values.BreakSeconds = parsed
		}
	case pomodoroBreakNoBreakChoice:
		values.BreakSeconds = 0
	}
	if values.BreakSeconds <= 0 {
		values.Cycles = 1
		values.LongBreakSeconds = 0
		values.CyclesBeforeLongBreak = 0
		if values.FocusSeconds > 0 {
			values.TotalSeconds = values.FocusSeconds
		}
		return values, nil
	}

	switch state.PomodoroLongBreakChoice {
	case pomodoroLongBreakCustomChoice:
		parsed, err := ParseDurationInput(state.Inputs[2].Value(), validate, "Long break duration")
		if err != nil {
			return values, err
		}
		if parsed > 0 {
			values.LongBreakSeconds = parsed
		}
	case pomodoroLongBreakNoBreakChoice:
		values.LongBreakSeconds = 0
	}

	parsed, err := ParsePositiveIntInput(state.Inputs[3].Value(), validate, "Cycles")
	if err != nil {
		return values, err
	}
	if parsed > 0 {
		values.Cycles = parsed
	}
	if values.LongBreakSeconds > 0 {
		parsed, err = ParsePositiveIntInput(
			state.Inputs[4].Value(),
			validate,
			"Long Break",
		)
		if err != nil {
			return values, err
		}
		if parsed > 0 {
			values.CyclesBeforeLongBreak = parsed
		}
	} else {
		values.CyclesBeforeLongBreak = 0
	}

	if values.Cycles > 0 && values.FocusSeconds > 0 {
		values.TotalSeconds = values.Cycles * (values.FocusSeconds + values.BreakSeconds)
		if values.LongBreakSeconds > 0 && values.CyclesBeforeLongBreak > 0 {
			values.TotalSeconds += (values.Cycles / values.CyclesBeforeLongBreak) * (values.LongBreakSeconds - values.BreakSeconds)
		}
	}

	return values, nil
}
