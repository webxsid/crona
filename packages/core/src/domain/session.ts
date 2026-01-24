export interface Session {
  id: string;
  issueId: string;
  startTime: string;   // ISO timestamp
  endTime?: string | undefined;    // ISO timestamp
  durationSeconds?: number | undefined;
  notes?: string | undefined;
}
