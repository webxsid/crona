import { randomUUID } from "crypto";
import type { Repo } from "../domain/repo";
import type { ICommandContext } from "./context";

/**
 * Create a new repo
 */
export async function createRepo(
  ctx: ICommandContext,
  input: {
    name: string;
    color?: string;
  }
): Promise<Repo> {
  if (!input.name.trim()) {
    throw new Error("Repo name cannot be empty");
  }

  const repo: Repo = {
    id: randomUUID(),
    name: input.name.trim(),
    color: input.color,
  };

  const now = ctx.now();

  await ctx.repos.create(repo, {
    userId: ctx.userId,
    now,
  });

  await ctx.ops.append({
    id: randomUUID(),
    entity: "repo",
    entityId: repo.id,
    action: "create",
    payload: repo,
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  return repo;
}

/**
 * Rename / recolor a repo
 */
export async function updateRepo(
  ctx: ICommandContext,
  repoId: string,
  updates: {
    name?: string;
    color?: string;
  }
): Promise<Repo> {
  if (updates.name !== undefined && !updates.name.trim()) {
    throw new Error("Repo name cannot be empty");
  }

  const now = ctx.now();

  const updateObj: Partial<Pick<Repo, "name" | "color">> = {};

  if (updates.name !== undefined) {
    updateObj["name"] = updates.name.trim();
  }
  if (Object.prototype.hasOwnProperty.call(updates, "color")) {
    updateObj["color"] = updates.color;
  }

  const updated = await ctx.repos.update(
    repoId,
    updateObj,
    {
      userId: ctx.userId,
      now,
    }
  );

  await ctx.ops.append({
    id: randomUUID(),
    entity: "repo",
    entityId: repoId,
    action: "update",
    payload: updates,
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  return updated;
}

/**
 * Delete a repo (soft delete)
 * NOTE: cascading deletes are handled at storage or command level later
 */
export async function deleteRepo(
  ctx: ICommandContext,
  repoId: string
): Promise<void> {
  const now = ctx.now();

  await ctx.repos.softDelete(repoId, {
    userId: ctx.userId,
    now,
  });

  await ctx.ops.append({
    id: randomUUID(),
    entity: "repo",
    entityId: repoId,
    action: "delete",
    payload: null,
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });
}

/**
 * List all repos for current user
 * Read-only, no ops emitted
 */
export async function listRepos(
  ctx: ICommandContext
): Promise<Repo[]> {
  return ctx.repos.list(ctx.userId);
}
