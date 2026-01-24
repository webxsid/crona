export type SessionSegmentType = "work" | "short_break" | "long_break" | "rest";
export interface SessionSegment {
    id: string;
    userId: string;
    deviceId: string;
    sessionId: string;
    segmentType: SessionSegmentType;
    startTime: Date;
    endTime?: Date | undefined;
    elapsedOffsetSeconds?: number | undefined;
    createdAt: Date;
}
