"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.EventsRoutes = void 0;
/**
 * Server-Sent Events (SSE)
 * - Realtime kernel → client events
 */
class EventsRoutes {
    app;
    ctx;
    constructor(app, ctx) {
        this.app = app;
        this.ctx = ctx;
    }
    register() {
        this.app.get("/events", async (req, reply) => {
            console.debug("[EventsRoutes] Client connected to /events SSE endpoint");
            // Required SSE headers
            reply.raw.setHeader("Content-Type", "text/event-stream");
            reply.raw.setHeader("Cache-Control", "no-cache");
            reply.raw.setHeader("Connection", "keep-alive");
            // Flush headers immediately
            reply.raw.flushHeaders?.();
            // Initial ping so client knows stream is alive
            reply.raw.write(":ok\n\n");
            // Subscribe to internal event bus
            const unsubscribe = this.ctx.events.subscribe((event) => {
                reply.raw.write(`event: ${event.type}\n` +
                    `data: ${JSON.stringify(event.payload)}\n\n`);
            });
            // Cleanup on disconnect
            req.raw.on("close", () => {
                unsubscribe();
            });
        });
    }
}
exports.EventsRoutes = EventsRoutes;
