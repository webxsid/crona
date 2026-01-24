import type { Issue } from "../domain/issue";
import type { IssueStatus } from "../domain/issue";
import type { ICommandContext } from "./context";
/**
 * Create a new issue under a stream
 */
export declare function createIssue(ctx: ICommandContext, input: {
    streamId: string;
    title: string;
    estimateMinutes?: number | undefined;
    notes?: string | undefined;
}): Promise<Issue>;
/**
 * Update issue metadata (title, estimate, notes)
 * Does NOT handle status transitions
 */
export declare function updateIssue(ctx: ICommandContext, issueId: string, updates: {
    title?: string | undefined;
    estimateMinutes?: number | null | undefined;
    notes?: string | null | undefined;
}): Promise<Issue>;
/**
 * Change issue status
 * Explicit command to keep state machine clean
 */
export declare function changeIssueStatus(ctx: ICommandContext, issueId: string, nextStatus: IssueStatus): Promise<Issue>;
/**
 * Delete an issue (soft delete)
 */
export declare function deleteIssue(ctx: ICommandContext, issueId: string): Promise<void>;
/**
 * List issues in a stream
 * Read-only
 */
export declare function listIssuesByStream(ctx: ICommandContext, streamId: string): Promise<Issue[]>;
