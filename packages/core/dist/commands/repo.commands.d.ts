import type { Repo } from "../domain/repo";
import type { ICommandContext } from "./context";
/**
 * Create a new repo
 */
export declare function createRepo(ctx: ICommandContext, input: {
    name: string;
    color?: string;
}): Promise<Repo>;
/**
 * Rename / recolor a repo
 */
export declare function updateRepo(ctx: ICommandContext, repoId: string, updates: {
    name?: string;
    color?: string;
}): Promise<Repo>;
/**
 * Delete a repo (soft delete)
 * NOTE: cascading deletes are handled at storage or command level later
 */
export declare function deleteRepo(ctx: ICommandContext, repoId: string): Promise<void>;
/**
 * List all repos for current user
 * Read-only, no ops emitted
 */
export declare function listRepos(ctx: ICommandContext): Promise<Repo[]>;
