import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";

/**
 * Server-Sent Events (SSE)
 * - Realtime kernel → client events
 */
export class EventsRoutes {
  constructor(
    private readonly app: FastifyInstance,
    private readonly ctx: ICommandContext
  ) { }

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
        reply.raw.write(
          `data: ${JSON.stringify({ type: event.type, payload: event.payload })}\n\n`
        );
      });

      // Cleanup on disconnect
      req.raw.on("close", () => {
        unsubscribe();
      });
    });
  }
}
