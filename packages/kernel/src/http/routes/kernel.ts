import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
import { readKernelInfo } from "../../kernel-info";

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
  }
}
