import type { ICommandContext } from "./context";
import type { ActiveContext } from "../domain";
/**
 * Get current active context (read-only)
 */
export declare function getActiveContext(ctx: ICommandContext): Promise<ActiveContext | null>;
/**
 * Switch active repo
 * Clears downstream context (stream, issue)
 */
export declare function switchRepo(ctx: ICommandContext, repoId: string): Promise<ActiveContext>;
/**
 * Switch active stream
 * Requires repo to be set
 * Clears issue
 */
export declare function switchStream(ctx: ICommandContext, streamId: string): Promise<ActiveContext>;
/**
 * Switch active issue
 * Requires stream to be set
 */
export declare function switchIssue(ctx: ICommandContext, issueId: string): Promise<ActiveContext>;
/**
 * Clear active issue only
 * (equivalent to git checkout --)
 */
export declare function clearIssue(ctx: ICommandContext): Promise<ActiveContext>;
/**
 * Clear entire context
 * (equivalent to git checkout --detach)
 */
export declare function clearContext(ctx: ICommandContext): Promise<void>;
export declare function emitContextChanged(ctx: ICommandContext): Promise<void>;
