import type { FastifyInstance } from "fastify";
import type { ICommandContext, CoreSettingsKey } from "@crona/core";

/**
 * Core settings HTTP APIs
 * - Timer mode
 * - Break configuration
 * - Future kernel-level behavior
 */
export class SettingsRoutes {
  constructor(
    private readonly app: FastifyInstance,
    private readonly ctx: ICommandContext
  ) { }

  register() {
    this.registerQueries();
    this.registerCommands();
  }

  // ---------- Queries ----------

  private registerQueries() {
    /**
     * Get all core settings for current user
     * GET /settings/core
     */
    this.app.get("/settings/core", async () => {
      return this.ctx.coreSettings.getAllSettings();
    });

    /**
     * Get a single core setting by key
     * GET /settings/core/:key
     */
    this.app.get("/settings/core/:key", async (req) => {
      const { key } = req.params as { key: CoreSettingsKey };
      return this.ctx.coreSettings.getSetting<CoreSettingsKey>(this.ctx.userId, key);
    });
  }

  // ---------- Commands ----------

  private registerCommands() {
    /**
     * Patch core settings
     * PATCH /settings/core
     *
     * Body:
     * {
     *   key: CoreSettingsKey;
     *   value: unknown;
     * }
     */
    this.app.patch("/settings/core", async (req) => {
      const { key, value } = req.body as { key: CoreSettingsKey; value: unknown };

      return this.ctx.coreSettings.setSetting<CoreSettingsKey, unknown>(
        this.ctx.userId,
        key,
        value
      );
    });

    /**
     * Update multiple core settings
     * PUT /settings/core
     *
     * Body:
     * {
     *   [key: CoreSettingsKey]: unknown;
     * }
     */
    this.app.put("/settings/core", async (req) => {
      const settings = req.body as Partial<Record<CoreSettingsKey, unknown>>;

      const updatedSettings: Partial<Record<CoreSettingsKey, unknown>> = {};

      for (const key of Object.keys(settings) as CoreSettingsKey[]) {
        const value = settings[key];
        if (value !== undefined) {
          const updatedValue = await this.ctx.coreSettings.setSetting<
            CoreSettingsKey,
            unknown
          >(this.ctx.userId, key, value);
          updatedSettings[key] = updatedValue;
        }
      }

      return updatedSettings;
    });
  }
}
