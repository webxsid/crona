"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.SettingsRoutes = void 0;
/**
 * Core settings HTTP APIs
 * - Timer mode
 * - Break configuration
 * - Future kernel-level behavior
 */
class SettingsRoutes {
    app;
    ctx;
    constructor(app, ctx) {
        this.app = app;
        this.ctx = ctx;
    }
    register() {
        this.registerQueries();
        this.registerCommands();
    }
    // ---------- Queries ----------
    registerQueries() {
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
            const { key } = req.params;
            return this.ctx.coreSettings.getSetting(this.ctx.userId, key);
        });
    }
    // ---------- Commands ----------
    registerCommands() {
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
            const { key, value } = req.body;
            return this.ctx.coreSettings.setSetting(this.ctx.userId, key, value);
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
            const settings = req.body;
            const updatedSettings = {};
            for (const key of Object.keys(settings)) {
                const value = settings[key];
                if (value !== undefined) {
                    const updatedValue = await this.ctx.coreSettings.setSetting(this.ctx.userId, key, value);
                    updatedSettings[key] = updatedValue;
                }
            }
            return updatedSettings;
        });
    }
}
exports.SettingsRoutes = SettingsRoutes;
