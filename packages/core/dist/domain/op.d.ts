export type OpEntity = "repo" | "stream" | "issue" | "session" | "session_segment" | "active_context" | "stash";
export type OpAction = "create" | "update" | "delete";
export interface Op {
    id: string;
    entity: OpEntity;
    entityId: string;
    action: OpAction;
    payload: unknown;
    timestamp: string;
    userId: string;
    deviceId: string;
}
