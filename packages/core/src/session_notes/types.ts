export type SessionNoteSection =
  | "commit"
  | "context"
  | "work"
  | "notes";

export type ParsedSessionNotes = Partial<
  Record<SessionNoteSection, string>
>;

export interface SessionWorkSummary {
  workSeconds: number;
  restSeconds: number;
  workSegments: number;
  restSegments: number;
  totalSeconds: number;
}
