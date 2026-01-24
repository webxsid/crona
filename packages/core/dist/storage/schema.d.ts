import type { OpEntity, SessionSegmentType } from "../domain";
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
    status: "todo" | "active" | "done";
    estimate_minutes: number | null;
    notes: string | null;
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
    action: "create" | "update" | "delete";
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
}
export declare function initSchema(): Promise<void>;
