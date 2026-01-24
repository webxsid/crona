import type { SessionSegment, SessionSegmentType } from "../../domain";

export interface ISessionSegmentRepository {
  getActive(
    userId: string,
    deviceId: string,
    sessionId: string
  ): Promise<SessionSegment | null>;

  startSegment(
    userId: string,
    deviceId: string,
    sessionId: string,
    type: SessionSegmentType
  ): Promise<SessionSegment>;

  endActiveSegment(
    userId: string,
    deviceId: string,
    sessionId: string
  ): Promise<void>;

  listBySession(
    sessionId: string
  ): Promise<SessionSegment[]>;

  applyElapsedOffset(
    sessionId: string,
    offsetSeconds: number
  ): Promise<void>;

  countWorkSegments(
    sessionId: string
  ): Promise<number>;
}
