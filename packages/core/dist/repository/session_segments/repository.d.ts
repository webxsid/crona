import type { SessionSegment, SessionSegmentType } from "../../domain";
import type { ISessionSegmentRepository } from "./interface";
export declare class SessionSegmentRepository implements ISessionSegmentRepository {
    getActive(userId: string, deviceId: string, sessionId: string): Promise<SessionSegment | null>;
    startSegment(userId: string, deviceId: string, sessionId: string, type: SessionSegmentType): Promise<SessionSegment>;
    endActiveSegment(userId: string, deviceId: string, sessionId: string): Promise<void>;
    listBySession(sessionId: string): Promise<SessionSegment[]>;
    countWorkSegments(sessionId: string): Promise<number>;
    applyElapsedOffset(sessionId: string, offsetSeconds: number): Promise<void>;
    private map;
}
