package controller

import (
	"strconv"
	"strings"

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
		FocusChoices:           pomodoroFocusChoices(),
		BreakChoices:           pomodoroBreakChoices(),
		LongBreakChoices:       pomodoroLongBreakChoices(),
		FocusCustomChoice:      pomodoroFocusCustomChoice,
		BreakCustomChoice:      pomodoroBreakCustomChoice,
		LongBreakCustomChoice:  pomodoroLongBreakCustomChoice,
		LongBreakSelected:      longBreakSelected,
		FocusRowActive:         state.FocusIdx == pomodoroFocusRowIdx,
		FocusCustomActive:      state.FocusIdx == pomodoroFocusCustomIdx,
		BreakRowActive:         state.FocusIdx == pomodoroBreakRowIdx,
		BreakCustomActive:      state.FocusIdx == pomodoroBreakCustomIdx,
		LongBreakRowActive:     state.FocusIdx == pomodoroLongBreakRowIdx,
		LongBreakCustomActive:  state.FocusIdx == pomodoroLongBreakCustomIdx,
		CyclesRowActive:        state.FocusIdx == pomodoroCyclesRowIdx,
		LongBreakCyclesActive:  state.FocusIdx == pomodoroLongBreakCyclesIdx,
		CyclesDisabled:         pomodoroBreakDisabled(state),
		LongBreakForcedOff:     longBreakForcedOff,
		LongBreakDisabled:      pomodoroBreakDisabled(state) || pomodoroLongBreakDisabled(state),
		ShowSummary:            values.FocusSeconds > 0 || values.BreakSeconds > 0,
		FocusDisplay:           pomodoroDisplayDuration(state.PomodoroFocusChoice, pomodoroFocusCustomChoice, -1, state.Inputs[0].Value(), values.FocusSeconds),
		ShortBreakDisplay:      pomodoroDisplayDuration(state.PomodoroBreakChoice, pomodoroBreakCustomChoice, pomodoroBreakNoBreakChoice, state.Inputs[1].Value(), values.BreakSeconds),
		LongBreakDisplay:       pomodoroDisplayDuration(longBreakSelected, pomodoroLongBreakCustomChoice, pomodoroLongBreakNoBreakChoice, state.Inputs[2].Value(), values.LongBreakSeconds),
		CyclesSummary:          pomodoroCyclesSummary(state, values),
		EstimatedTotalDuration: pomodoroEstimatedTotalDuration(values),
	}
}

func pomodoroDisplayDuration(choice int, customChoice int, noBreakChoice int, input string, seconds int) string {
	if noBreakChoice >= 0 && choice == noBreakChoice {
		return "No Break"
	}
	if choice == customChoice {
		if value := strings.TrimSpace(input); value != "" {
			return value
		}
	}
	return helperpkg.FormatCompactDurationSeconds(seconds)
}

func pomodoroCyclesSummary(state State, values pomodoroValues) string {
	if values.BreakSeconds <= 0 {
		return "continuous"
	}
	cycles := strings.TrimSpace(state.Inputs[3].Value())
	if cycles == "" && values.Cycles > 0 {
		cycles = strconv.Itoa(values.Cycles)
	}
	if cycles == "" {
		cycles = "0"
	}
	if values.LongBreakSeconds <= 0 {
		return cycles + " / long break off"
	}
	every := strings.TrimSpace(state.Inputs[4].Value())
	if every == "" && values.CyclesBeforeLongBreak > 0 {
		every = strconv.Itoa(values.CyclesBeforeLongBreak)
	}
	if every == "" {
		every = "0"
	}
	return cycles + " / every " + every + " cycles"
}

func pomodoroEstimatedTotalDuration(values pomodoroValues) string {
	if values.TotalSeconds <= 0 {
		return ""
	}
	return "Estimated total " + helperpkg.FormatCompactDurationSeconds(values.TotalSeconds)
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
	if state.PomodoroBreakChoice == pomodoroBreakCustomChoice {
		parsed, err := ParseDurationInput(state.Inputs[1].Value(), validate, "Short break duration")
		if err != nil {
			return values, err
		}
		if parsed > 0 {
			values.BreakSeconds = parsed
		}
	} else if state.PomodoroBreakChoice == pomodoroBreakNoBreakChoice {
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
	if state.PomodoroLongBreakChoice == pomodoroLongBreakCustomChoice {
		parsed, err := ParseDurationInput(state.Inputs[2].Value(), validate, "Long break duration")
		if err != nil {
			return values, err
		}
		if parsed > 0 {
			values.LongBreakSeconds = parsed
		}
	} else if state.PomodoroLongBreakChoice == pomodoroLongBreakNoBreakChoice {
		values.LongBreakSeconds = 0
	}
	parsed, err := ParsePositiveIntInput(state.Inputs[3].Value(), validate, "Number of cycles")
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
			"Cycle before long break",
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
