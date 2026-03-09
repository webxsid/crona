import type { Issue, IssueStatus, IssueWithMeta } from "../../domain";

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

  listAll(userId: string): Promise<IssueWithMeta[]>;

  listByStream(
    streamId: string,
    userId: string
  ): Promise<Issue[]>;

  listDeletedByStream(
    streamId: string,
    userId: string
  ): Promise<Issue[]>;

  listByTodoForDate(
    todoForDate: string,
    userId: string
  ): Promise<Issue[]>;

  update(
    issueId: string,
    updates: {
      title?: string | undefined;
      status?: IssueStatus | undefined;
      estimateMinutes?: number | null | undefined;
      notes?: string | null | undefined;
      todoForDate?: string | null | undefined;
      completedAt?: string | null | undefined;
      abandonedAt?: string | null | undefined;
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

  cascadeSoftDeleteByStreamId(
    streamId: string,
    meta: {
      userId: string;
      now: string;
    }
  ): Promise<void>

  cascadeSoftDeleteByRepoId(
    repoId: string,
    meta: {
      userId: string;
      now: string;
    }
  ): Promise<void>

  restoreDeletedById(
    issueId: string,
    meta: {
      userId: string;
      now: string;
    }
  ): Promise<void>;

  restoreDeletedByStreamId(
    streamId: string,
    meta: {
      userId: string;
      now: string;
    }
  ): Promise<void>;

  restoreDeletedByRepoId(
    repoId: string,
    meta: {
      userId: string;
      now: string;
    }
  ): Promise<void>;
}
