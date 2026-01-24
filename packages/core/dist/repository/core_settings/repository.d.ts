import type { CoreSettings } from "../../domain";
import type { CoreSettingsKey, ICoreSettingsRepository } from "./interface";
export declare class CoreSettingsRepository implements ICoreSettingsRepository {
    constructor();
    getSetting<K extends CoreSettingsKey>(userId: string, key: K): Promise<CoreSettings[K] | undefined>;
    setSetting<K extends CoreSettingsKey, V>(userId: string, key: K, value: V): Promise<void>;
    getAllSettings(): Promise<Record<string, unknown>>;
    initializeDefaults(userId: string, deviceId: string): Promise<void>;
    private keyToColumn;
    private columnToValue;
}
