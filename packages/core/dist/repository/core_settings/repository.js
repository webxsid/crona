"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.CoreSettingsRepository = void 0;
const storage_1 = require("../../storage");
const constant_1 = require("../../storage/constant");
class CoreSettingsRepository {
    constructor() { }
    async getSetting(userId, key) {
        const row = await storage_1.SqliteDb.getDB()
            .selectFrom("core_settings")
            .select(this.keyToColumn(key))
            .where("user_id", "=", userId)
            .executeTakeFirst();
        if (!row)
            return undefined;
        return this.columnToValue(row);
    }
    async setSetting(userId, key, value) {
        await storage_1.SqliteDb.getDB()
            .updateTable("core_settings")
            .set({ [this.keyToColumn(key)]: value })
            .where("user_id", "=", userId)
            .execute();
    }
    async getAllSettings() {
        const rows = await storage_1.SqliteDb.getDB().selectFrom("core_settings").selectAll().execute();
        const result = {};
        for (const row of rows) {
            const userId = row.user_id;
            result[userId] = row;
        }
        return result;
    }
    async initializeDefaults(userId, deviceId) {
        // check if defaults already exist
        const existing = await storage_1.SqliteDb.getDB()
            .selectFrom("core_settings")
            .select("user_id")
            .where("user_id", "=", userId)
            .executeTakeFirst();
        if (existing) {
            return;
        }
        const defaults = constant_1.DEFAULT_CORE_SETTINGS;
        const computedDefaults = {
            user_id: userId,
            device_id: deviceId,
            timer_mode: defaults.timerMode,
            breaks_enabled: defaults.breaksEnabled ? 1 : 0,
            work_duration_minutes: defaults.workDurationMinutes,
            short_break_minutes: defaults.shortBreakMinutes,
            long_break_minutes: defaults.longBreakMinutes,
            long_break_enabled: defaults.longBreakEnabled ? 1 : 0,
            cycles_before_long_break: defaults.cyclesBeforeLongBreak,
            auto_start_breaks: defaults.autoStartBreaks ? 1 : 0,
            auto_start_work: defaults.autoStartWork ? 1 : 0,
            created_at: Date.now().toString(),
            updated_at: Date.now().toString(),
        };
        await storage_1.SqliteDb.getDB().insertInto("core_settings").values(computedDefaults).execute();
    }
    keyToColumn(key) {
        const map = {
            timerMode: "timer_mode",
            breaksEnabled: "breaks_enabled",
            workDurationMinutes: "work_duration_minutes",
            shortBreakMinutes: "short_break_minutes",
            longBreakMinutes: "long_break_minutes",
            longBreakEnabled: "long_break_enabled",
            cyclesBeforeLongBreak: "cycles_before_long_break",
            autoStartBreaks: "auto_start_breaks",
            autoStartWork: "auto_start_work",
        };
        return map[key];
    }
    columnToValue(row) {
        return Object.values(row)[0];
    }
}
exports.CoreSettingsRepository = CoreSettingsRepository;
