import type { Session } from "../../domain";
import type { ParsedSessionNotes } from "../../session_notes";
import type { ISessionRepository } from "./interface";
export declare class SessionRepository implements ISessionRepository {
    start(session: Session, meta: {
        userId: string;
        deviceId: string;
        now: string;
    }): Promise<Session>;
    stop(sessionId: string, updates: {
        endTime: string;
        durationSeconds: number;
        notes: string | undefined;
    }, meta: {
        userId: string;
        deviceId: string;
        now: string;
    }): Promise<Session>;
    getActiveSession(userId: string): Promise<Session | null>;
    listByIssue(issueId: string, userId: string): Promise<Session[]>;
    getSessiobById(sessionId: string, userId: string): Promise<Session | null>;
    getLastSessionForIssue(issueId: string, userId: string): Promise<Session | null>;
    ammendSessionNotes(sessionId: string, newNotes: string, meta: {
        userId: string;
        deviceId: string;
        now: string;
    }): Promise<Session>;
    getLastSessionForUser(userId: string): Promise<Session | null>;
    listEnded(input: {
        userId: string;
        repoId?: string | undefined;
        streamId?: string | undefined;
        issueId?: string | undefined;
        since?: string | undefined;
        until?: string | undefined;
        limit?: number | undefined;
        offset?: number | undefined;
    }): Promise<Array<Session & {
        parsedNotes: ParsedSessionNotes;
    }>>;
}
