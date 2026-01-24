export type IssueStatus = "todo" | "active" | "done";
export interface Issue {
    id: string;
    streamId: string;
    title: string;
    status: IssueStatus;
    estimateMinutes?: number | undefined;
    notes?: string | undefined;
}
