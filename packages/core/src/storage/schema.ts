import type { OpEntity, SessionSegmentType } from "../domain";
import { sql } from "kysely";
import { SqliteDb } from "./db";

export interface RepoTable {
  id: string;
  name: string;
  color: string | null;
  user_id: string;
  created_at: string;
  updated_at: string;
  deleted_at: string | null;
}

export interface StreamTable {
  id: string;
  repo_id: string;
  name: string;
  visibility: "personal" | "shared";
  user_id: string;
  created_at: string;
  updated_at: string;
  deleted_at: string | null;
}

export interface IssueTable {
  id: string;
  stream_id: string;
  title: string;
  status: "todo" | "active" | "done" | "abandoned";
  estimate_minutes: number | null;
  notes: string | null;
  todo_for_date: string | null;
  completed_at: string | null;
  abandoned_at: string | null;
  user_id: string;
  created_at: string;
  updated_at: string;
  deleted_at: string | null;
}

export interface SessionTable {
  id: string;
  issue_id: string;
  start_time: string;
  end_time: string | null;
  duration_seconds: number | null;
  notes: string | null;
  user_id: string;
  device_id: string;
  created_at: string;
  updated_at: string;
  deleted_at: string | null;
}

export interface StashTable {
  id: string;
  repo_id: string | null;
  stream_id: string | null;
  issue_id: string | null;
  session_id: string | null;
  segment_type: SessionSegmentType | null;
  segment_started_at: string | null;
  elapsed_seconds: number | null;
  note: string | null;
  user_id: string;
  device_id: string;
  created_at: string;
  updated_at: string;
  deleted_at: string | null;
}

export interface OpTable {
  id: string;
  user_id: string;
  device_id: string;
  entity: OpEntity;
  entity_id: string;
  action: "create" | "update" | "delete" | "restore";
  payload: string;
  timestamp: string;
}

export interface CoreSettingsTable {
  user_id: string;
  device_id: string;
  timer_mode: "stopwatch" | "structured";
  breaks_enabled: number;
  work_duration_minutes: number;
  short_break_minutes: number;
  long_break_minutes: number;
  long_break_enabled: number;
  cycles_before_long_break: number;
  auto_start_breaks: number;
  auto_start_work: number;
  created_at: string;
  updated_at: string;
}

export interface SessionSegmentsTable {
  id: string;
  user_id: string;
  device_id: string;
  session_id: string;
  segment_type: SessionSegmentType;
  elapsed_offset_seconds: number | null;
  start_time: string;
  end_time: string | null;
  created_at: string;
}

export interface ActiveContextTable {
  user_id: string;
  device_id: string;
  repo_id: string | null;
  stream_id: string | null;
  issue_id: string | null;
  updated_at: string;
}

export interface ScratchPadMetaTable {
  id: string;
  user_id: string;
  device_id: string;
  name: string;
  path: string;
  last_opened_at: string;
  pinned: number;
}

export interface DB {
  repos: RepoTable;
  streams: StreamTable;
  issues: IssueTable;
  sessions: SessionTable;
  stash: StashTable;
  ops: OpTable;
  core_settings: CoreSettingsTable;
  session_segments: SessionSegmentsTable;
  active_context: ActiveContextTable;
  scratch_pad_meta: ScratchPadMetaTable;
}

export async function initSchema(): Promise<void> {
  const db = SqliteDb.getDB();
  await db.schema
    .createTable("repos")
    .ifNotExists()
    .addColumn("id", "text", (col) => col.primaryKey())
    .addColumn("name", "text", (col) => col.notNull().unique())
    .addColumn("color", "text")
    .addColumn("user_id", "text", (col) => col.notNull())
    .addColumn("created_at", "text", (col) => col.notNull())
    .addColumn("updated_at", "text", (col) => col.notNull())
    .addColumn("deleted_at", "text")
    .execute();

  await db.schema
    .createTable("streams")
    .ifNotExists()
    .addColumn("id", "text", (col) => col.primaryKey())
    .addColumn("repo_id", "text", (col) => col.notNull())
    .addColumn("name", "text", (col) => col.notNull())
    .addColumn("visibility", "text", (col) => col.notNull())
    .addColumn("user_id", "text", (col) => col.notNull())
    .addColumn("created_at", "text", (col) => col.notNull())
    .addColumn("updated_at", "text", (col) => col.notNull())
    .addColumn("deleted_at", "text")
    .execute();

  await db.schema
    .createTable("issues")
    .ifNotExists()
    .addColumn("id", "text", (col) => col.primaryKey())
    .addColumn("stream_id", "text", (col) => col.notNull())
    .addColumn("title", "text", (col) => col.notNull())
    .addColumn("status", "text", (col) => col.notNull())
    .addColumn("estimate_minutes", "integer")
    .addColumn("notes", "text")
    .addColumn("todo_for_date", "text")
    .addColumn("completed_at", "text")
    .addColumn("abandoned_at", "text")
    .addColumn("user_id", "text", (col) => col.notNull())
    .addColumn("created_at", "text", (col) => col.notNull())
    .addColumn("updated_at", "text", (col) => col.notNull())
    .addColumn("deleted_at", "text")
    .execute();

  await ensureIssueColumn("completed_at");
  await ensureIssueColumn("abandoned_at");

  await db.schema
    .createTable("sessions")
    .ifNotExists()
    .addColumn("id", "text", (col) => col.primaryKey())
    .addColumn("issue_id", "text", (col) => col.notNull())
    .addColumn("start_time", "text", (col) => col.notNull())
    .addColumn("end_time", "text")
    .addColumn("duration_seconds", "integer")
    .addColumn("notes", "text")
    .addColumn("user_id", "text", (col) => col.notNull())
    .addColumn("device_id", "text", (col) => col.notNull())
    .addColumn("created_at", "text", (col) => col.notNull())
    .addColumn("updated_at", "text", (col) => col.notNull())
    .addColumn("deleted_at", "text")
    .execute();

  await db.schema
    .createTable("stash")
    .ifNotExists()
    .addColumn("id", "text", (col) => col.primaryKey())
    .addColumn("repo_id", "text")
    .addColumn("stream_id", "text")
    .addColumn("issue_id", "text")
    .addColumn("session_id", "text")
    .addColumn("segment_type", "text")
    .addColumn("segment_started_at", "text")
    .addColumn("elapsed_seconds", "integer")
    .addColumn("note", "text")
    .addColumn("user_id", "text", (col) => col.notNull())
    .addColumn("device_id", "text", (col) => col.notNull())
    .addColumn("created_at", "text", (col) => col.notNull())
    .addColumn("updated_at", "text", (col) => col.notNull())
    .addColumn("deleted_at", "text")
    .execute();

  await db.schema
    .createTable("ops")
    .ifNotExists()
    .addColumn("id", "text", (col) => col.primaryKey())
    .addColumn("user_id", "text", (col) => col.notNull())
    .addColumn("device_id", "text", (col) => col.notNull())
    .addColumn("entity", "text", (col) => col.notNull())
    .addColumn("entity_id", "text", (col) => col.notNull())
    .addColumn("action", "text", (col) => col.notNull())
    .addColumn("payload", "text", (col) => col.notNull())
    .addColumn("timestamp", "text", (col) => col.notNull())
    .execute();

  await db.schema
    .createTable("core_settings")
    .ifNotExists()
    .addColumn("user_id", "text", (col) => col.primaryKey())
    .addColumn("device_id", "text", (col) => col.notNull())
    .addColumn("timer_mode", "text", (col) => col.notNull())
    .addColumn("breaks_enabled", "integer", (col) => col.notNull())
    .addColumn("work_duration_minutes", "integer", (col) => col.notNull())
    .addColumn("short_break_minutes", "integer", (col) => col.notNull())
    .addColumn("long_break_minutes", "integer", (col) => col.notNull())
    .addColumn("long_break_enabled", "integer", (col) => col.notNull())
    .addColumn("cycles_before_long_break", "integer", (col) => col.notNull())
    .addColumn("auto_start_breaks", "integer", (col) => col.notNull())
    .addColumn("auto_start_work", "integer", (col) => col.notNull())
    .addColumn("created_at", "text", (col) => col.notNull())
    .addColumn("updated_at", "text", (col) => col.notNull())
    .execute();

  await db.schema
    .createTable("session_segments")
    .ifNotExists()
    .addColumn("id", "text", (col) => col.primaryKey())
    .addColumn("user_id", "text", (col) => col.notNull())
    .addColumn("device_id", "text", (col) => col.notNull())
    .addColumn("session_id", "text", (col) => col.notNull())
    .addColumn("segment_type", "text", (col) => col.notNull())
    .addColumn("elapsed_offset_seconds", "integer")
    .addColumn("start_time", "text", (col) => col.notNull())
    .addColumn("end_time", "text")
    .addColumn("created_at", "text", (col) => col.notNull())
    .execute();

  await db.schema
    .createTable("active_context")
    .ifNotExists()
    .addColumn("user_id", "text", (col) => col.primaryKey())
    .addColumn("device_id", "text", (col) => col.notNull())
    .addColumn("repo_id", "text")
    .addColumn("stream_id", "text")
    .addColumn("issue_id", "text")
    .addColumn("updated_at", "text", (col) => col.notNull())
    .execute();

  await db.schema
    .createTable("scratch_pad_meta")
    .ifNotExists()
    .addColumn("id", "text", (col) => col.primaryKey())
    .addColumn("user_id", "text", (col) => col.notNull())
    .addColumn("device_id", "text", (col) => col.notNull())
    .addColumn("name", "text", (col) => col.notNull())
    .addColumn("path", "text", (col) => col.notNull().unique())
    .addColumn("last_opened_at", "text", (col) => col.notNull())
    .addColumn("pinned", "integer", (col) => col.notNull())
    .execute();


  // indexes
  await db.schema
    .createIndex("idx_streams_repo_id")
    .on("streams")
    .column("repo_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_issues_stream_id")
    .on("issues")
    .column("stream_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_sessions_issue_id")
    .on("sessions")
    .column("issue_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_stash_repo_id")
    .on("stash")
    .column("repo_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_stash_stream_id")
    .on("stash")
    .column("stream_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_stash_issue_id")
    .on("stash")
    .column("issue_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_ops_entity_entity_id")
    .on("ops")
    .columns(["entity", "entity_id"])
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_repos_user_id")
    .on("repos")
    .column("user_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_streams_user_id")
    .on("streams")
    .column("user_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_issues_user_id")
    .on("issues")
    .column("user_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_sessions_user_id")
    .on("sessions")
    .column("user_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_stash_user_id")
    .on("stash")
    .column("user_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_ops_user_id")
    .on("ops")
    .column("user_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_session_segments_session_id")
    .on("session_segments")
    .column("session_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_session_segments_user_id")
    .on("session_segments")
    .column("user_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_active_context_user_id")
    .on("active_context")
    .column("user_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_active_context_device_id")
    .on("active_context")
    .column("device_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_scratch_pad_meta_user_id")
    .on("scratch_pad_meta")
    .column("user_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_scratch_pad_meta_device_id")
    .on("scratch_pad_meta")
    .column("device_id")
    .ifNotExists()
    .execute();

  await db.schema
    .createIndex("idx_scratch_pad_meta_last_opened_at")
    .on("scratch_pad_meta")
    .column("last_opened_at")
    .ifNotExists()
    .execute();
}

async function ensureIssueColumn(columnName: "completed_at" | "abandoned_at"): Promise<void> {
  const rows = await sql<{ name: string }>`PRAGMA table_info('issues')`.execute(SqliteDb.getDB());
  const exists = rows.rows.some((row) => row.name === columnName);
  if (exists) {
    return;
  }

  await SqliteDb.getDB().schema
    .alterTable("issues")
    .addColumn(columnName, "text")
    .execute();
}
