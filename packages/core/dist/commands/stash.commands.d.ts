import type { Stash } from "../domain/stash";
import type { ICommandContext } from "./context";
/**
 * Stash current context (and session if running)
 */
export declare function stashPush(ctx: ICommandContext, stashNote?: string): Promise<Stash>;
export declare function stashPop(ctx: ICommandContext, stashId: string): Promise<void>;
export declare function stashDrop(ctx: ICommandContext, stashId: string): Promise<void>;
