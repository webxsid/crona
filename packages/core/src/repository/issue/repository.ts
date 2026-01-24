import type { Issue, IssueStatus } from "../../domain";
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
    };
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
    }));
  }

  async update(
    issueId: string,
    updates: {
      title?: string;
      status?: IssueStatus;
      estimateMinutes?: number | null;
      notes?: string | null;
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
}
