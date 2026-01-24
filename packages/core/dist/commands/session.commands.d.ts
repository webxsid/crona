import type { Session } from "../domain/session";
import type { ICommandContext } from "./context";
import type { SessionSegmentType } from "../domain";
import type { ParsedSessionNotes } from "../session_notes";
/**
 * Start a session for an issue
 * Enforces: only one active session per user
 */
export declare function startSession(ctx: ICommandContext, issueId: string): Promise<Session>;
/**
 * Stop the currently active session
 * Idempotent: calling stop with no active session is a no-op
 */
export declare function stopSession(ctx: ICommandContext, commitMessage?: string | undefined): Promise<Session | null>;
export declare function ammendSessionNotes(ctx: ICommandContext, message: string, sessionId?: string | undefined): Promise<Session>;
/**
 * Pause the current session
 * Stores elapsed time in stash
 */
export declare function pauseSession(ctx: ICommandContext, nextSegmentType?: SessionSegmentType): Promise<void>;
export declare function resumeSession(ctx: ICommandContext): Promise<void>;
/**
 * List Session History
 * Read-only
 */
export declare function listSessionHistory(ctx: ICommandContext, query: {
    repoId: string | undefined;
    streamId: string | undefined;
    issueId: string | undefined;
    since: string | undefined;
    until: string | undefined;
    limit: number | undefined;
    offset: number | undefined;
}, useContext?: boolean): Promise<Array<Session & {
    parsedNotes: ParsedSessionNotes;
}>>;
/**
 * List all sessions for an issue
 * Read-only
 */
export declare function listSessionsByIssue(ctx: ICommandContext, issueId: string): Promise<Session[]>;
