export const DEFAULT_CORE_SETTINGS = {
  timerMode: "stopwatch" as const,
  breaksEnabled: false,

  workDurationMinutes: 25,
  shortBreakMinutes: 5,
  longBreakMinutes: 15,

  longBreakEnabled: true,
  cyclesBeforeLongBreak: 4,

  autoStartBreaks: true,
  autoStartWork: false,
};
