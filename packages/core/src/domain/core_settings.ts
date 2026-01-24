export interface CoreSettings {
  userId: string;
  deviceId: string;
  timerMode: "stopwatch" | "structured";
  breaksEnabled: boolean;
  workDurationMinutes: number;
  shortBreakMinutes: number;
  longBreakMinutes: number;
  longBreakEnabled: boolean;
  cyclesBeforeLongBreak: number;
  autoStartBreaks: boolean;
  autoStartWork: boolean;
  createdAt: string;
  updatedAt: string;
}
