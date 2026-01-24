import type { CoreSettings } from "../../domain";
export type CoreSettingsKey = keyof Omit<CoreSettings, "userId" | "deviceId" | "createdAt" | "updatedAt">;
export interface ICoreSettingsRepository {
    getSetting<T extends CoreSettingsKey>(userId: string, key: T): Promise<CoreSettings[T] | undefined>;
    setSetting<T extends CoreSettingsKey, K>(userId: string, key: T, value: K): Promise<void>;
    getAllSettings(): Promise<Record<string, unknown>>;
    initializeDefaults(userId: string, deviceId: string): Promise<void>;
}
