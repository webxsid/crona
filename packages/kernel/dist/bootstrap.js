"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.bootstrapKernel = bootstrapKernel;
const server_1 = require("./http/server");
const kernel_info_1 = require("./kernel-info");
const core_1 = require("@crona/core");
const os_1 = __importDefault(require("os"));
let shuttingDown = false;
async function gracefulShutdown(ctx) {
    if (shuttingDown) {
        return;
    }
    shuttingDown = true;
    try {
        await (0, core_1.stopSession)(ctx);
    }
    catch (err) {
        console.error("Failed to stop session during shutdown", err);
    }
    try {
        await (0, kernel_info_1.clearKernelInfo)();
    }
    catch (err) {
        console.error("Failed to clear kernel info", err);
    }
    console.log("Kernel shutting down");
    process.exit(0);
}
function _printAllAppRoutes(app) {
    const routes = app.printRoutes?.();
    console.debug("Registered HTTP routes:\n", routes);
}
async function bootstrapKernel({ dbPath }) {
    const eventBus = new core_1.EventBus();
    console.debug("EventBus initialized:", eventBus);
    core_1.SqliteDb.initDb(dbPath);
    await (0, core_1.initDb)();
    const userId = "local";
    const deviceId = os_1.default.hostname() || "device-1";
    const ctx = await (0, core_1.createCommandContext)({
        userId,
        deviceId,
        now: () => new Date().toISOString(),
        events: eventBus,
    });
    try {
        await ctx.coreSettings.initializeDefaults(userId, deviceId);
    }
    catch (err) {
        console.error("Failed to initialize core settings defaults", err);
        process.exit(1);
    }
    // init Active Context
    try {
        await ctx.activeContext.initializeDefaults(userId, deviceId);
    }
    catch (err) {
        console.error("Failed to initialize active context defaults", err);
        process.exit(1);
    }
    // set up inmemeory timer boundaaries
    try {
        const timerService = new core_1.TimerService(ctx);
        await timerService.scheduleNextBoundary();
    }
    catch (err) {
        console.error("Failed to schedule timer boundaries", err);
        process.exit(1);
    }
    const server = (0, server_1.createHttpServer)(ctx);
    await server.listen({ port: 0, host: "127.0.0.1" });
    const addr = server.server.address();
    let port;
    if (typeof addr === "string") {
        // unix socket
        port = addr;
    }
    else if (addr && typeof addr === "object") {
        port = addr.port;
    }
    if (!port) {
        throw new Error("Failed to start kernel HTTP server");
    }
    // printAllAppRoutes(server);
    await (0, kernel_info_1.writeKernelInfo)({
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
    console.log("Kernel running", addr);
}
