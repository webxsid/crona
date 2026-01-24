import type { ActiveContext } from "../../domain";

export interface IActiveContextRepository {
  get(userId: string, deviceId: string): Promise<ActiveContext | null>;
  set(
    userId: string,
    deviceId: string,
    context: Partial<Omit<ActiveContext, "userId" | "updatedAt">>
  ): Promise<ActiveContext>;

  clear(userId: string, deviceId: string): Promise<void>;
  initializeDefaults(userId: string, deviceId: string): Promise<void>;
}
