import { createHttpServer } from "./http/server";
import { clearKernelInfo, writeKernelInfo } from "./kernel-info";
import { logError, logInfo } from "./logger";

import type { ICommandContext } from "@crona/core";
import { createCommandContext, EventBus, initDb, SqliteDb, stopSession, TimerService } from "@crona/core";
// import type { FastifyInstance } from "fastify";
import os from "os";

let shuttingDown = false;

async function gracefulShutdown(ctx: ICommandContext) {
  if (shuttingDown) {
    return;
  }
  shuttingDown = true;
  try {
    await stopSession(ctx);
  } catch (err) {
    console.error("Failed to stop session during shutdown", err);
    logError("Failed to stop session during shutdown", err);
  }

  try {
    await clearKernelInfo();
  } catch (err) {
    console.error("Failed to clear kernel info", err);
    logError("Failed to clear kernel info", err);
  }

  logInfo("Kernel shutting down");
  console.log("Kernel shutting down");
  process.exit(0);
}

export interface IBootstrapKernelOptions {
  dbPath?: string | undefined;
}

// function _printAllAppRoutes(app: FastifyInstance) {
//   const routes = app.printRoutes?.();
//   console.debug("Registered HTTP routes:\n", routes);
// }

export async function bootstrapKernel({ dbPath }: IBootstrapKernelOptions) {
  const eventBus = new EventBus();

  console.debug("EventBus initialized:", eventBus);

  SqliteDb.initDb(dbPath);

  await initDb();

  const userId = "local";
  const deviceId = os.hostname() || "device-1";

  const ctx = await createCommandContext({
    userId,
    deviceId,
    now: () => new Date().toISOString(),
    events: eventBus,
  });

  try {
    await ctx.coreSettings.initializeDefaults(
      userId,
      deviceId
    );
  } catch (err) {
    console.error("Failed to initialize core settings defaults", err);
    logError("Failed to initialize core settings defaults", err);
    process.exit(1);
  }

  // init Active Context
  try {
    await ctx.activeContext.initializeDefaults(userId, deviceId);
  } catch (err) {
    console.error("Failed to initialize active context defaults", err);
    logError("Failed to initialize active context defaults", err);
    process.exit(1);
  }

  // set up inmemeory timer boundaaries
  try {
    const timerService = new TimerService(ctx);
    await timerService.scheduleNextBoundary()
  } catch (err) {
    console.error("Failed to schedule timer boundaries", err);
    logError("Failed to schedule timer boundaries", err);
    process.exit(1);
  }

  const server = createHttpServer(ctx);

  await server.listen({ port: 0, host: "127.0.0.1" });

  const addr = server.server.address();

  let port;

  if (typeof addr === "string") {
    // unix socket
    port = addr;
  } else if (addr && typeof addr === "object") {
    port = addr.port;
  }

  if (!port) {
    throw new Error("Failed to start kernel HTTP server");
  }
  // printAllAppRoutes(server);

  await writeKernelInfo({
    port: Number(port),
    token: ctx.authToken,
    pid: process.pid,
    startedAt: new Date().toISOString(),
  });


  process.on("SIGTERM", async () => {
    await gracefulShutdown(ctx);
  });

  process.on("SIGINT", async () => {
    await gracefulShutdown(ctx);
  });

  logInfo(`Kernel listening on port ${String(port)}`);
  console.log("Kernel running", addr);
}

