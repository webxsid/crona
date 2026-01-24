export interface Session {
    id: string;
    issueId: string;
    startTime: string;
    endTime?: string | undefined;
    durationSeconds?: number | undefined;
    notes?: string | undefined;
}
