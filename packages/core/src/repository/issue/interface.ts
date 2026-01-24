import type { Issue, IssueStatus } from "../../domain";

export interface IIssueRepository {
  create(
    issue: Issue,
    meta: {
      userId: string;
      now: string;
    }
  ): Promise<Issue>;

  getById(
    issueId: string,
    userId: string
  ): Promise<Issue | null>;

  listByStream(
    streamId: string,
    userId: string
  ): Promise<Issue[]>;

  update(
    issueId: string,
    updates: {
      title?: string | undefined;
      status?: IssueStatus | undefined;
      estimateMinutes?: number | null | undefined;
      notes?: string | null | undefined;
    },
    meta: {
      userId: string;
      now: string;
    }
  ): Promise<Issue>;

  softDelete(
    issueId: string,
    meta: {
      userId: string;
      now: string;
    }
  ): Promise<void>;
}
