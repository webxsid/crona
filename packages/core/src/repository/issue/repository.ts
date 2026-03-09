import type { Issue, IssueStatus, IssueWithMeta } from "../../domain";
import { SqliteDb } from "../../storage";
import type { IIssueRepository } from "./interface";

export class SqliteIssueRepository implements IIssueRepository {
  async create(
    issue: Issue,
    meta: { userId: string; now: string }
  ): Promise<Issue> {
    await SqliteDb.getDB()
      .insertInto("issues")
      .values({
        id: issue.id,
        stream_id: issue.streamId,
        title: issue.title,
        status: issue.status,
        estimate_minutes: issue.estimateMinutes ?? null,
        todo_for_date: issue.todoForDate ?? null,
        completed_at: issue.completedAt ?? null,
        abandoned_at: issue.abandonedAt ?? null,
        notes: issue.notes ?? null,
        user_id: meta.userId,
        created_at: meta.now,
        updated_at: meta.now,
        deleted_at: null,
      })
      .execute();

    return issue;
  }

  async getById(
    issueId: string,
    userId: string
  ): Promise<Issue | null> {
    const row = await SqliteDb.getDB()
      .selectFrom("issues")
      .select([
        "id",
        "stream_id",
        "title",
        "status",
        "estimate_minutes",
        "notes",
        "todo_for_date",
        "completed_at",
        "abandoned_at",
      ])
      .where("id", "=", issueId)
      .where("user_id", "=", userId)
      .where("deleted_at", "is", null)
      .executeTakeFirst();

    if (!row) return null;

    return {
      id: row.id,
      streamId: row.stream_id,
      title: row.title,
      status: row.status as IssueStatus,
      estimateMinutes: row.estimate_minutes ?? undefined,
      notes: row.notes ?? undefined,
      todoForDate: row.todo_for_date ?? undefined,
      completedAt: row.completed_at ?? undefined,
      abandonedAt: row.abandoned_at ?? undefined,
    };
  }

  async listAll(userId: string): Promise<IssueWithMeta[]> {
    const rows = await SqliteDb.getDB()
      .selectFrom("issues")
      .innerJoin("streams", "streams.id", "issues.stream_id")
      .innerJoin("repos", "repos.id", "streams.repo_id")
      .select([
        "issues.id",
        "issues.stream_id",
        "issues.title",
        "issues.status",
        "issues.estimate_minutes",
        "issues.notes",
        "issues.todo_for_date",
        "issues.completed_at",
        "issues.abandoned_at",
        "streams.name as stream_name",
        "repos.id as repo_id",
        "repos.name as repo_name",
      ])
      .where("issues.user_id", "=", userId)
      .where("issues.deleted_at", "is", null)
      .where("streams.deleted_at", "is", null)
      .where("repos.deleted_at", "is", null)
      .orderBy("issues.created_at", "asc")
      .execute();

    return rows.map((r) => ({
      id: r.id,
      streamId: r.stream_id,
      title: r.title,
      status: r.status as IssueStatus,
      estimateMinutes: r.estimate_minutes ?? undefined,
      notes: r.notes ?? undefined,
      todoForDate: r.todo_for_date ?? undefined,
      completedAt: r.completed_at ?? undefined,
      abandonedAt: r.abandoned_at ?? undefined,
      repoId: r.repo_id,
      repoName: r.repo_name,
      streamName: r.stream_name,
    }));
  }

  async listByStream(
    streamId: string,
    userId: string
  ): Promise<Issue[]> {
    const rows = await SqliteDb.getDB()
      .selectFrom("issues")
      .select([
        "id",
        "stream_id",
        "title",
        "status",
        "estimate_minutes",
        "notes",
        "todo_for_date",
        "completed_at",
        "abandoned_at",
      ])
      .where("stream_id", "=", streamId)
      .where("user_id", "=", userId)
      .where("deleted_at", "is", null)
      .orderBy("created_at", "asc")
      .execute();

    return rows.map((r) => ({
      id: r.id,
      streamId: r.stream_id,
      title: r.title,
      status: r.status as IssueStatus,
      estimateMinutes: r.estimate_minutes ?? undefined,
      notes: r.notes ?? undefined,
      todoForDate: r.todo_for_date ?? undefined,
      completedAt: r.completed_at ?? undefined,
      abandonedAt: r.abandoned_at ?? undefined,
    }));
  }

  async listDeletedByStream(
    streamId: string,
    userId: string
  ): Promise<Issue[]> {
    // join with streams to ensure the stream exists and is not deleted
    const rows = await SqliteDb.getDB()
      .selectFrom("issues")
      .innerJoin("streams", "streams.id", "issues.stream_id")
      .select([
        "issues.id",
        "issues.stream_id",
        "issues.title",
        "issues.status",
        "issues.estimate_minutes",
        "issues.notes",
        "issues.todo_for_date",
        "issues.completed_at",
        "issues.abandoned_at",
      ])
      .where("issues.stream_id", "=", streamId)
      .where("issues.user_id", "=", userId)
      .where("issues.deleted_at", "is not", null)
      .where("streams.deleted_at", "is", null)
      .execute();

    return rows.map((r) => ({
      id: r.id,
      streamId: r.stream_id,
      title: r.title,
      status: r.status as IssueStatus,
      estimateMinutes: r.estimate_minutes ?? undefined,
      notes: r.notes ?? undefined,
      todoForDate: r.todo_for_date ?? undefined,
      completedAt: r.completed_at ?? undefined,
      abandonedAt: r.abandoned_at ?? undefined,
    }));
  }

  async update(
    issueId: string,
    updates: {
      title?: string;
      status?: IssueStatus;
      estimateMinutes?: number | null;
      notes?: string | null;
      todoForDate?: string | null | undefined;
      completedAt?: string | null | undefined;
      abandonedAt?: string | null | undefined;
    },
    meta: { userId: string; now: string }
  ): Promise<Issue> {
    const result = await SqliteDb.getDB()
      .updateTable("issues")
      .set({
        title: updates.title,
        status: updates.status,
        estimate_minutes:
          updates.estimateMinutes === undefined
            ? undefined
            : updates.estimateMinutes,
        notes:
          updates.notes === undefined ? undefined : updates.notes,
        todo_for_date:
          updates.todoForDate === undefined ? undefined : updates.todoForDate,
        completed_at:
          updates.completedAt === undefined ? undefined : updates.completedAt,
        abandoned_at:
          updates.abandonedAt === undefined ? undefined : updates.abandonedAt,
        updated_at: meta.now,
      })
      .where("id", "=", issueId)
      .where("user_id", "=", meta.userId)
      .where("deleted_at", "is", null)
      .executeTakeFirst();

    if (result.numUpdatedRows === BigInt(0)) {
      throw new Error("Issue not found or already deleted");
    }

    const updated = await this.getById(issueId, meta.userId);
    if (!updated) {
      throw new Error("Issue disappeared after update");
    }

    return updated;
  }

  async softDelete(
    issueId: string,
    meta: { userId: string; now: string }
  ): Promise<void> {
    const result = await SqliteDb.getDB()
      .updateTable("issues")
      .set({
        deleted_at: meta.now,
        updated_at: meta.now,
      })
      .where("id", "=", issueId)
      .where("user_id", "=", meta.userId)
      .where("deleted_at", "is", null)
      .executeTakeFirst();

    if (result.numUpdatedRows === BigInt(0)) {
      throw new Error("Issue not found or already deleted");
    }
  }

  async cascadeSoftDeleteByStreamId(
    streamId: string,
    meta: { userId: string; now: string }
  ): Promise<void> {
    await SqliteDb.getDB()
      .updateTable("issues")
      .set({
        deleted_at: meta.now,
        updated_at: meta.now,
      })
      .where("stream_id", "=", streamId)
      .where("user_id", "=", meta.userId)
      .where("deleted_at", "is", null)
      .execute();
  }

  async cascadeSoftDeleteByRepoId(
    repoId: string,
    meta: { userId: string; now: string }
  ): Promise<void> {
    await SqliteDb.getDB()
      .updateTable("issues")
      .set({
        deleted_at: meta.now,
        updated_at: meta.now,
      })
      .where("stream_id", "in",
        SqliteDb.getDB()
          .selectFrom("streams")
          .select("id")
          .where("repo_id", "=", repoId)
          .where("user_id", "=", meta.userId)
          .where("deleted_at", "is", null)
      )
      .where("user_id", "=", meta.userId)
      .where("deleted_at", "is", null)
      .execute();
  }

  async restoreDeletedById(
    issueId: string,
    meta: { userId: string; now: string }
  ): Promise<void> {
    const result = await SqliteDb.getDB()
      .updateTable("issues")
      .set({
        deleted_at: null,
        updated_at: meta.now,
      })
      .where("id", "=", issueId)
      .where("user_id", "=", meta.userId)
      .where("deleted_at", "is not", null)
      .executeTakeFirst();

    if (result.numUpdatedRows === BigInt(0)) {
      throw new Error("Issue not found or not deleted");
    }
  }

  async restoreDeletedByStreamId(
    streamId: string,
    meta: { userId: string; now: string }
  ): Promise<void> {
    await SqliteDb.getDB()
      .updateTable("issues")
      .set({
        deleted_at: null,
        updated_at: meta.now,
      })
      .where("stream_id", "=", streamId)
      .where("user_id", "=", meta.userId)
      .where("deleted_at", "is not", null)
      .execute();
  }

  async restoreDeletedByRepoId(
    repoId: string,
    meta: { userId: string; now: string }
  ): Promise<void> {
    await SqliteDb.getDB()
      .updateTable("issues")
      .set({
        deleted_at: null,
        updated_at: meta.now,
      })
      .where("stream_id", "in",
        SqliteDb.getDB()
          .selectFrom("streams")
          .select("id")
          .where("repo_id", "=", repoId)
          .where("user_id", "=", meta.userId)
          .where("deleted_at", "is", null)
      )
      .where("user_id", "=", meta.userId)
      .where("deleted_at", "is not", null)
      .execute();
  }

  async listByTodoForDate(
    todoForDate: string,
    userId: string
  ): Promise<Issue[]> {
    const rows = await SqliteDb.getDB()
      .selectFrom("issues")
      .select([
        "id",
        "stream_id",
        "title",
        "status",
        "estimate_minutes",
        "notes",
        "todo_for_date",
        "completed_at",
        "abandoned_at",
      ])
      .where("todo_for_date", "=", todoForDate)
      .where("user_id", "=", userId)
      .where("deleted_at", "is", null)
      .orderBy("created_at", "asc")
      .execute();

    return rows.map((r) => ({
      id: r.id,
      streamId: r.stream_id,
      title: r.title,
      status: r.status as IssueStatus,
      estimateMinutes: r.estimate_minutes ?? undefined,
      notes: r.notes ?? undefined,
      todoForDate: r.todo_for_date ?? undefined,
      completedAt: r.completed_at ?? undefined,
      abandonedAt: r.abandoned_at ?? undefined,
    }));
  }
}
