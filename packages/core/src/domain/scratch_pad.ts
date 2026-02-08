export interface ScratchPadMeta {
  id: string;
  path: string;
  name: string;
  lastOpenedAt: Date;
  pinned: boolean;
}

export const SCRATCH_DIR_NAME = "scratch";
