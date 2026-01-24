import type { SessionSegmentType } from "./session_segment";

export interface Stash {
  id: string;
  userId: string;
  deviceId: string;

  repoId?: string | undefined;
  streamId?: string | undefined;
  issueId?: string | undefined;

  sessionId?: string | undefined;
  pausedSegmentType?: SessionSegmentType | undefined;
  elapsedSeconds?: number | undefined;

  note?: string | undefined;

  createdAt: Date;
  updatedAt: Date;
}
