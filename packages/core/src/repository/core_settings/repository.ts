import type { CoreSettings } from "../../domain";
import type { CoreSettingsTable } from "../../storage";
import { SqliteDb } from "../../storage";
import { DEFAULT_CORE_SETTINGS } from "../../storage/constant";
import type { CoreSettingsKey, ICoreSettingsRepository } from "./interface";

export class CoreSettingsRepository implements ICoreSettingsRepository {
  constructor() { }

  async getSetting<K extends CoreSettingsKey>(
    userId: string,
    key: K
  ): Promise<CoreSettings[K] | undefined> {
    const row = await SqliteDb.getDB()
      .selectFrom("core_settings")
      .select(this.keyToColumn(key) as keyof CoreSettingsTable)
      .where("user_id", "=", userId)
      .executeTakeFirst();

    if (!row) return undefined;

    return this.columnToValue(row);
  }

  async setSetting<K extends CoreSettingsKey, V>(
    userId: string,
    key: K,
    value: V
  ): Promise<void> {
    await SqliteDb.getDB()
      .updateTable("core_settings")
      .set({ [this.keyToColumn(key)]: value } as Partial<CoreSettingsTable>)
      .where("user_id", "=", userId)
      .execute();
  }

  async getAllSettings(): Promise<Record<string, unknown>> {
    const rows = await SqliteDb.getDB().selectFrom("core_settings").selectAll().execute();
    const result: Record<string, unknown> = {};

    for (const row of rows) {
      const userId = row.user_id;
      result[userId] = row;
    }

    return result;
  }

  async initializeDefaults(
    userId: string,
    deviceId: string
  ): Promise<void> {
    // check if defaults already exist
    const existing = await SqliteDb.getDB()
      .selectFrom("core_settings")
      .select("user_id")
      .where("user_id", "=", userId)
      .executeTakeFirst();

    if (existing) {
      return;
    }

    const defaults = DEFAULT_CORE_SETTINGS;
    const computedDefaults: CoreSettingsTable = {
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
    await SqliteDb.getDB().insertInto("core_settings").values(computedDefaults).execute();
  }

  private keyToColumn(key: CoreSettingsKey) {
    const map: Record<CoreSettingsKey, string> = {
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

  private columnToValue<K extends CoreSettingsKey>(
    row: Record<string, unknown>
  ): CoreSettings[K] {
    return Object.values(row)[0] as CoreSettings[K];
  }
}
