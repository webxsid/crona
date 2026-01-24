import type { SessionSegmentType } from "../domain";

export type TimerStatePayload =
  | {
    state: "idle";
  }
  | {
    state: "running";
    sessionId: string;
    issueId: string;
    segmentType: SessionSegmentType;
    elapsedSeconds: number;
  }
  | {
    state: "paused";
    issueId: string;
    elapsedSeconds: number;
  };
