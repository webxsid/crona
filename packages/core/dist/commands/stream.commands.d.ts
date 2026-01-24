import type { Stream, StreamVisibility } from "../domain/stream";
import type { ICommandContext } from "./context";
/**
 * Create a new stream under a repo
 */
export declare function createStream(ctx: ICommandContext, input: {
    repoId: string;
    name: string;
    visibility?: StreamVisibility;
}): Promise<Stream>;
/**
 * Rename / change visibility of a stream
 */
export declare function updateStream(ctx: ICommandContext, streamId: string, updates: {
    name?: string;
    visibility?: StreamVisibility;
}): Promise<Stream>;
/**
 * Delete a stream (soft delete)
 */
export declare function deleteStream(ctx: ICommandContext, streamId: string): Promise<void>;
/**
 * List streams for a repo
 * Read-only → no ops
 */
export declare function listStreamsByRepo(ctx: ICommandContext, repoId: string): Promise<Stream[]>;
