export type IssueStatus = "todo" | "active" | "done" | "abandoned";

export interface Issue {
  id: string;
  streamId: string;
  title: string;
  status: IssueStatus;
  estimateMinutes?: number | undefined;
  notes?: string | undefined;
  todoForDate?: string | undefined; // ISO date string (YYYY-MM-DD)
  completedAt?: string | undefined; // ISO datetime
  abandonedAt?: string | undefined; // ISO datetime
}

export interface IssueWithMeta extends Issue {
  repoId: string;
  repoName: string;
  streamName: string;
}

export interface DailyIssueSummary {
  date: string; // ISO date string (YYYY-MM-DD)
  totalIssues: number;
  issues: Issue[];
  totalEstimatedMinutes: number;
  completedIssues: number;
  abandonedIssues: number;
  workedSeconds: number;
}
