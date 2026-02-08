import type { FastifyInstance } from "fastify";
import type { ScratchPadMeta } from "@crona/core";
import { listScratchpads, pinScratchpad, registerScratchpad, removeScratchpad, type ICommandContext } from "@crona/core";

export class ScratchRoutes {
  constructor(
    private readonly app: FastifyInstance,
    private readonly ctx: ICommandContext
  ) { }

  register() {
    this.registerQueries();
    this.registerCommands();
  }

  private registerQueries() {
    // GET /scratchpads
    this.app.get("/scratchpads", async (req) => {
      const pinnedOnly = (req.query as { pinnedOnly: string }).pinnedOnly === "1";
      return listScratchpads(this.ctx, {
        pinnedOnly
      });
    });
  }

  private registerCommands() {
    // POST /scratchpads/register
    this.app.post("/scratchpads/register", async (req) => {
      const body = req.body as ScratchPadMeta;
      await registerScratchpad(this.ctx, body);
      return { ok: true };
    });

    // PUT /scratchpads/pin?path=...  body: { pinned: boolean }
    this.app.put("/scratchpads/pin", async (req) => {
      const { path } = req.query as { path: string };
      const { pinned } = req.body as { pinned: boolean };
      await pinScratchpad(this.ctx, path, pinned);
      return { ok: true };
    });

    // DELETE /scratchpads?path=...
    this.app.delete("/scratchpads", async (req) => {
      const { path } = req.query as { path: string };
      await removeScratchpad(this.ctx, path);
      return { ok: true };
    });
  }
}
