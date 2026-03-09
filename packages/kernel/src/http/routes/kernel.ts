import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
import { readKernelInfo } from "../../kernel-info";
import { logInfo } from "../../logger";

export class KernelRoutes {
  constructor(
    private readonly app: FastifyInstance,
    private readonly _ctx: ICommandContext
  ) { }

  register() {
    this.app.get("/kernel/info", async () => {
      const info = await readKernelInfo();
      return info;
    });

    this.app.post("/kernel/restart", async (_req, reply) => {
      logInfo("Kernel restart requested via HTTP");
      void reply.send({ ok: true });
      process.nextTick(() => {
        process.kill(process.pid, "SIGTERM");
      });
    });

    this.app.post("/kernel/shutdown", async (_req, reply) => {
      logInfo("Kernel shutdown requested via HTTP");
      void reply.send({ ok: true });
      process.nextTick(() => {
        process.kill(process.pid, "SIGTERM");
      });
    });
  }
}
