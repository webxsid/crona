import type { ScratchPadMeta } from "../../domain";

export interface IScratchRepo {
  upsert(ScratchPadMeta: ScratchPadMeta, meta: {
    userId: string;
    deviceId: string;
  }): Promise<void>;
  list(
    userId: string,
    deviceId: string,
    options?: {
      pinnedOnly?: boolean;
    }
  ): Promise<ScratchPadMeta[]>;
  get(path: string, meta: {
    userId: string;
    deviceId: string;
  }): Promise<ScratchPadMeta | null>;
  remove(path: string, meta: {
    userId: string;
    deviceId: string;
  }): Promise<void>;
}
