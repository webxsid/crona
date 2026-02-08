import type { ScratchPadMeta } from "../domain";
import type { ICommandContext } from "./context";

function handleVariablesInPath(path: string): string {
  // the vriables will be in the format [[variableName]]
  // supported variables are:
  // - [[date]]: current date in YYYY-MM-DD format
  // - [[time]]: current time in HH-mm-ss format
  // - [[datetime]]: current date and time in YYYY-MM-DD_HH-mm-ss format
  // - [[timestamp]]: current timestamp in milliseconds
  // - [[random]]: random string of 8 characters

  const allowedVariables = ["date", "time", "datetime", "timestamp", "random"];

  const varibaleRegex = /\[\[(\w+)\]\]/g;

  const foundVariables = path.matchAll(varibaleRegex);

  if (!foundVariables) {
    return path;
  }

  if (Array.from(foundVariables).some((match) => {
    const variableName = match[1];
    return !allowedVariables.includes(variableName!);
  })) {
    throw new Error(
      `Invalid variable in path. Allowed variables are: ${allowedVariables.join(
        ", "
      )}`
    );
  }

  let resultPath = path;

  for (const match of path.matchAll(varibaleRegex)) {
    const variableName = match[1];
    let replacement = "";
    const now = new Date();
    switch (variableName) {
      case "date":
        replacement = now.toISOString().split("T")[0]!;
        break;
      case "time":
        replacement = now.toTimeString().split(" ")[0]!.replace(/:/g, "-");
        break;
      case "datetime":
        replacement = now
          .toISOString()
          .replace(/[:.]/g, "-")
          .replace("T", "_")
          .split("Z")[0]!;
        break;
      case "timestamp":
        replacement = now.getTime().toString();
        break;
      case "random":
        replacement = Math.random().toString(36).substring(2, 10);
        break;
    }
    resultPath = resultPath.replace(match[0], replacement);
  }

  return resultPath;

}

export async function registerScratchpad(
  ctx: ICommandContext,
  meta: ScratchPadMeta
): Promise<void> {
  const incomingPath = meta.path;
  if (!incomingPath) {
    throw new Error("Path is required to register a scratchpad");
  }

  const processedPath = handleVariablesInPath(incomingPath);
  await ctx.scratchPads.upsert({
    id: `${ctx.userId}:${ctx.deviceId}:${processedPath}`,
    name: meta.name,
    path: processedPath,
    pinned: meta.pinned ?? false,
    lastOpenedAt: new Date(ctx.now()),
  }, {
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });
}

export async function listScratchpads(
  ctx: ICommandContext,
  options: {
    pinnedOnly?: boolean;
  }
): Promise<ScratchPadMeta[]> {
  return ctx.scratchPads.list(
    ctx.userId,
    ctx.deviceId,
    {
      pinnedOnly: options.pinnedOnly ?? false,
    }
  );
}

export async function pinScratchpad(
  ctx: ICommandContext,
  path: string,
  pinned: boolean
): Promise<void> {
  const existing = await ctx.scratchPads.get(path, {
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });
  if (!existing) throw new Error("Scratchpad not found");

  await ctx.scratchPads.upsert({ ...existing, pinned }, {
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });
}

export async function removeScratchpad(
  ctx: ICommandContext,
  path: string
): Promise<void> {
  await ctx.scratchPads.remove(path, {
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });
}
